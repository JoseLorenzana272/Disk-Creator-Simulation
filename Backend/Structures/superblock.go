package structures

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Ahora mejor crear un mapa dentro de otro mapa, donde el primer mapa sea del id de la particion y el segundo mapa sea el de los grupos

var Groups = make(map[string]string)

type SuperBlock struct {
	S_filesystem_type   int32
	S_inodes_count      int32
	S_blocks_count      int32
	S_free_inodes_count int32
	S_free_blocks_count int32
	S_mtime             float32
	S_umtime            float32
	S_mnt_count         int32
	S_magic             int32
	S_inode_size        int32
	S_block_size        int32
	S_first_ino         int32
	S_first_blo         int32
	S_bm_inode_start    int32
	S_bm_block_start    int32
	S_inode_start       int32
	S_block_start       int32
	// Total: 68 bytes
}

// Serialize escribe la estructura SuperBlock en un archivo binario en la posición especificada
func (sb *SuperBlock) Serialize(path string, offset int64) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Serializar la estructura SuperBlock directamente en el archivo
	err = binary.Write(file, binary.LittleEndian, sb)
	if err != nil {
		return err
	}

	return nil
}

// Deserialize lee la estructura SuperBlock desde un archivo binario en la posición especificada
func (sb *SuperBlock) Deserialize(path string, offset int64) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Obtener el tamaño de la estructura SuperBlock
	sbSize := binary.Size(sb)
	if sbSize <= 0 {
		return fmt.Errorf("invalid SuperBlock size: %d", sbSize)
	}

	// Leer solo la cantidad de bytes que corresponden al tamaño de la estructura SuperBlock
	buffer := make([]byte, sbSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserializar los bytes leídos en la estructura SuperBlock
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, sb)
	if err != nil {
		return err
	}

	return nil
}

// Crear users.txt
func (sb *SuperBlock) CreateUsersFile(path string) error {
	// ----------- Creamos / -----------
	// Creamos el inodo raíz
	rootInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  0,
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'0'},
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Serializar el inodo raíz
	err := rootInode.Serialize(path, int64(sb.S_first_ino))
	if err != nil {
		return err
	}

	// Actualizar el bitmap de inodos
	err = sb.UpdateBitmapInode(path)
	if err != nil {
		return err
	}

	// Actualizar el superbloque
	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size

	// Creamos el bloque del Inodo Raíz
	rootBlock := &FolderBlock{
		B_content: [4]FolderContent{
			{B_name: [12]byte{'.'}, B_inodo: 0},
			{B_name: [12]byte{'.', '.'}, B_inodo: 0},
			{B_name: [12]byte{'-'}, B_inodo: -1},
			{B_name: [12]byte{'-'}, B_inodo: -1},
		},
	}

	// Actualizar el bitmap de bloques
	err = sb.UpdateBitmapBlock(path)
	if err != nil {
		return err
	}

	// Serializar el bloque de carpeta raíz
	err = rootBlock.Serialize(path, int64(sb.S_first_blo))
	if err != nil {
		return err
	}

	// Actualizar el superbloque
	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += sb.S_block_size

	// Verificar el inodo raíz
	fmt.Println("\nInodo Raíz:")
	rootInode.Print()

	// Verificar el bloque de carpeta raíz
	fmt.Println("\nBloque de Carpeta Raíz:")
	rootBlock.Print()

	// ----------- Creamos /users.txt -----------
	usersText := "1,G,root\n1,U,root,root,123\n"

	// Deserializar el inodo raíz
	err = rootInode.Deserialize(path, int64(sb.S_inode_start+0)) // 0 porque es el inodo raíz
	if err != nil {
		return err
	}

	// Actualizamos el inodo raíz
	rootInode.I_atime = float32(time.Now().Unix())

	// Serializar el inodo raíz
	err = rootInode.Serialize(path, int64(sb.S_inode_start+0)) // 0 porque es el inodo raíz
	if err != nil {
		return err
	}

	// Deserializar el bloque de carpeta raíz
	err = rootBlock.Deserialize(path, int64(sb.S_block_start+0)) // 0 porque es el bloque de carpeta raíz
	if err != nil {
		return err
	}

	// Actualizamos el bloque de carpeta raíz
	rootBlock.B_content[2] = FolderContent{B_name: [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'}, B_inodo: sb.S_inodes_count}

	// Serializar el bloque de carpeta raíz
	err = rootBlock.Serialize(path, int64(sb.S_block_start+0)) // 0 porque es el bloque de carpeta raíz
	if err != nil {
		return err
	}

	// Creamos el inodo users.txt
	usersInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  int32(len(usersText)),
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'1'},
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Actualizar el bitmap de inodos
	err = sb.UpdateBitmapInode(path)
	if err != nil {
		return err
	}

	// Serializar el inodo users.txt
	err = usersInode.Serialize(path, int64(sb.S_first_ino))
	if err != nil {
		return err
	}

	// Actualizamos el superbloque
	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size

	// Creamos el bloque de users.txt
	usersBlock := &FileBlock{
		B_content: [64]byte{},
	}
	// Copiamos el texto de usuarios en el bloque
	copy(usersBlock.B_content[:], usersText)

	// Serializar el bloque de users.txt
	err = usersBlock.Serialize(path, int64(sb.S_first_blo))
	if err != nil {
		return err
	}

	// Actualizar el bitmap de bloques
	err = sb.UpdateBitmapBlock(path)
	if err != nil {
		return err
	}

	// Actualizamos el superbloque
	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += sb.S_block_size

	// Verificar el inodo raíz
	fmt.Println("\nInodo Raíz Actualizado:")
	rootInode.Print()

	// Verificar el bloque de carpeta raíz
	fmt.Println("\nBloque de Carpeta Raíz Actualizado:")
	rootBlock.Print()

	// Verificar el inodo users.txt
	fmt.Println("\nInodo users.txt:")
	usersInode.Print()

	// Verificar el bloque de users.txt
	fmt.Println("\nBloque de users.txt:")
	usersBlock.Print()

	return nil
}

// PrintSuperBlock imprime los valores de la estructura SuperBlock
func (sb *SuperBlock) Print() {
	// Convertir el tiempo de montaje a una fecha
	mountTime := time.Unix(int64(sb.S_mtime), 0)
	// Convertir el tiempo de desmontaje a una fecha
	unmountTime := time.Unix(int64(sb.S_umtime), 0)

	fmt.Printf("Filesystem Type: %d\n", sb.S_filesystem_type)
	fmt.Printf("Inodes Count: %d\n", sb.S_inodes_count)
	fmt.Printf("Blocks Count: %d\n", sb.S_blocks_count)
	fmt.Printf("Free Inodes Count: %d\n", sb.S_free_inodes_count)
	fmt.Printf("Free Blocks Count: %d\n", sb.S_free_blocks_count)
	fmt.Printf("Mount Time: %s\n", mountTime.Format(time.RFC3339))
	fmt.Printf("Unmount Time: %s\n", unmountTime.Format(time.RFC3339))
	fmt.Printf("Mount Count: %d\n", sb.S_mnt_count)
	fmt.Printf("Magic: %d\n", sb.S_magic)
	fmt.Printf("Inode Size: %d\n", sb.S_inode_size)
	fmt.Printf("Block Size: %d\n", sb.S_block_size)
	fmt.Printf("First Inode: %d\n", sb.S_first_ino)
	fmt.Printf("First Block: %d\n", sb.S_first_blo)
	fmt.Printf("Bitmap Inode Start: %d\n", sb.S_bm_inode_start)
	fmt.Printf("Bitmap Block Start: %d\n", sb.S_bm_block_start)
	fmt.Printf("Inode Start: %d\n", sb.S_inode_start)
	fmt.Printf("Block Start: %d\n", sb.S_block_start)

	//Generar el archivo .dot
	sb.GenerateSBDotFile()
}

func (sb *SuperBlock) GenerateSBDotFile() {
	// Convertir el tiempo de montaje a una fecha
	mountTime := time.Unix(int64(sb.S_mtime), 0)
	// Convertir el tiempo de desmontaje a una fecha
	unmountTime := time.Unix(int64(sb.S_umtime), 0)

	// Usamos strings.Builder para construir el archivo .dot
	var dotContent strings.Builder
	dotContent.WriteString("digraph SuperBlock {\n")
	dotContent.WriteString("  node [shape=none, fontname=\"Helvetica,Arial,sans-serif\"];\n")
	dotContent.WriteString("  rankdir=TB;\n\n")

	// Crear el nodo del SuperBlock
	dotContent.WriteString("  sb [label=<\n")
	dotContent.WriteString("    <table border='0' cellborder='1' cellspacing='0' cellpadding='10' style='rounded' bgcolor='#F5F5F5'>\n")
	dotContent.WriteString("      <tr>\n")
	dotContent.WriteString("        <td colspan='2' bgcolor='#333333'><font color='white'>SuperBlock Report</font></td>\n")
	dotContent.WriteString("      </tr>\n")

	// Agregar los atributos del SuperBlock
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Filesystem Type:</td><td>%d</td></tr>\n", sb.S_filesystem_type))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Inodes Count:</td><td>%d</td></tr>\n", sb.S_inodes_count))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Blocks Count:</td><td>%d</td></tr>\n", sb.S_blocks_count))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Free Inodes Count:</td><td>%d</td></tr>\n", sb.S_free_inodes_count))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Free Blocks Count:</td><td>%d</td></tr>\n", sb.S_free_blocks_count))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Mount Time:</td><td>%s</td></tr>\n", mountTime.Format(time.RFC3339)))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Unmount Time:</td><td>%s</td></tr>\n", unmountTime.Format(time.RFC3339)))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Mount Count:</td><td>%d</td></tr>\n", sb.S_mnt_count))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Magic:</td><td>%d</td></tr>\n", sb.S_magic))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Inode Size:</td><td>%d</td></tr>\n", sb.S_inode_size))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Block Size:</td><td>%d</td></tr>\n", sb.S_block_size))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>First Inode:</td><td>%d</td></tr>\n", sb.S_first_ino))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>First Block:</td><td>%d</td></tr>\n", sb.S_first_blo))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Bitmap Inode Start:</td><td>%d</td></tr>\n", sb.S_bm_inode_start))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Bitmap Block Start:</td><td>%d</td></tr>\n", sb.S_bm_block_start))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Inode Start:</td><td>%d</td></tr>\n", sb.S_inode_start))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Block Start:</td><td>%d</td></tr>\n", sb.S_block_start))

	// Cerrar tabla y nodo
	dotContent.WriteString("    </table>\n")
	dotContent.WriteString("  >];\n")
	dotContent.WriteString("}\n")

	// Crear el archivo .dot
	file, err := os.Create("SuperBlock.dot")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Escribir el contenido en el archivo
	_, err = file.WriteString(dotContent.String())
	if err != nil {
		fmt.Println("Error escribiendo en el archivo .dot:", err)
		return
	}

	fmt.Println("SuperBlock.dot file created")
}

// Imprimir inodos
func (sb *SuperBlock) PrintInodes(path string) error {
	// Imprimir inodos
	fmt.Println("\nInodos\n----------------")
	// Iterar sobre cada inodo
	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := &Inode{}
		// Deserializar el inodo
		err := inode.Deserialize(path, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil {
			return err
		}
		// Imprimir el inodo
		fmt.Printf("\nInodo %d:\n", i)
		inode.Print()
	}

	return nil
}

// Impriir bloques
func (sb *SuperBlock) PrintBlocks(path string) error {
	// Imprimir bloques
	fmt.Println("\nBloques\n----------------")
	// Iterar sobre cada inodo
	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := &Inode{}
		// Deserializar el inodo
		err := inode.Deserialize(path, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil {
			return err
		}
		// Iterar sobre cada bloque del inodo (apuntadores)
		for _, blockIndex := range inode.I_block {
			// Si el bloque no existe, salir
			if blockIndex == -1 {
				break
			}
			// Si el inodo es de tipo carpeta
			if inode.I_type[0] == '0' {
				block := &FolderBlock{}
				// Deserializar el bloque
				err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
				if err != nil {
					return err
				}
				// Imprimir el bloque
				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()
				continue

				// Si el inodo es de tipo archivo
			} else if inode.I_type[0] == '1' {
				block := &FileBlock{}
				// Deserializar el bloque
				err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
				if err != nil {
					return err
				}
				// Imprimir el bloque
				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()
				continue
			}

		}
	}

	return nil
}

// PrintUsersFileContent imprime el contenido de users.txt
func (sb *SuperBlock) PrintUsersFileContent(path string) (string, error) {
	rootBlock := &FolderBlock{}
	// Deserializar el bloque de carpeta raíz
	err := rootBlock.Deserialize(path, int64(sb.S_block_start+0))
	if err != nil {
		return "", fmt.Errorf("error al leer el bloque de carpeta raíz: %v", err)
	}

	// Buscar el inodo de users.txt
	var usersInodeIndex int32 = -1
	for _, content := range rootBlock.B_content {
		fileName := strings.TrimRight(string(content.B_name[:]), "\x00")
		if fileName == "users.txt" {
			usersInodeIndex = content.B_inodo
			break
		}
	}

	if usersInodeIndex == -1 {
		return "", fmt.Errorf("no se encontró el inodo de users.txt")
	}

	// Leer el inodo de users.txt
	usersInode := &Inode{}
	err = usersInode.Deserialize(path, int64(sb.S_inode_start+(usersInodeIndex*sb.S_inode_size)))
	if err != nil {
		return "", fmt.Errorf("error al leer el inodo de users.txt: %v", err)
	}

	// Leer el contenido del archivo users.txt
	var usersContent string
	for i := 0; i < len(usersInode.I_block); i++ {

		if usersInode.I_block[i] == -1 {
			fmt.Println("NOSEEEEEEEEEEE: ", usersInode.I_block[i])
			break
		}

		// Calcular el offset del bloque actual

		blockOffset := sb.S_block_start + (usersInode.I_block[i] * sb.S_block_size)

		// Leer el bloque de datos
		usersBlock := &FileBlock{}
		err = usersBlock.Deserialize(path, int64(blockOffset))
		if err != nil {
			return "", fmt.Errorf("error al leer el bloque %d de users.txt: %v", i, err)
		}

		// Convertir el contenido del bloque a string y agregarlo a usersText
		usersContent += string(usersBlock.B_content[:])
	}

	if usersInode.I_block[12] != -1 {
		fmt.Println("APUNTADOR GOKU: ", usersInode.I_block[12])
		blockIndirect := &PointerBlock{}
		err = blockIndirect.Deserialize(path, int64(sb.S_block_start+(usersInode.I_block[12]*sb.S_block_size)))
		if err != nil {
			fmt.Println("ERROR AL LEER EL BLOQUE INDIRECTO: ", err)
			return "", fmt.Errorf("error al leer el bloque indirecto de users.txt: %v", err)
		}
		fmt.Println("BLOQUE INDIRECTO: ", blockIndirect.P_pointers)
		for _, blockIndex := range blockIndirect.P_pointers {
			if blockIndex == -1 {
				break
			}

			block := &FileBlock{}
			err = block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {
				return "", fmt.Errorf("error al leer el bloque %d de users.txt: %v", blockIndex, err)
			}
			fmt.Println("BLOQUE: ", blockIndex, " CONTENIDO: ", string(block.B_content[:]))
			usersContent += string(block.B_content[:])
		}

	}

	// Agregar los grupos al mapa
	lines := strings.Split(usersContent, "\n")
	for _, line := range lines {
		//fmt.Println("Procesando línea:", line)
		parts := strings.Split(line, ",")

		// Verificar que la línea tenga al menos 3 partes (ya que los grupos tienen solo 3 partes)
		if len(parts) < 3 {
			//fmt.Println("Línea ignorada, no tiene suficientes partes:", parts)
			continue
		}

		// Verificar si la línea es un grupo (segunda parte es "G")
		if parts[1] == "G" {
			Groups[parts[2]] = "" // Almacenar el nombre del grupo en el mapa

			//fmt.Printf("Grupo añadido: %s\n", parts[2])
		} else {
			//fmt.Println("Línea ignorada, no es un grupo:", parts)
		}
	}

	fmt.Println("GRUPOS: ", Groups)

	// Devolver el contenido completo de users.txt
	return usersContent, nil
}

func AddUserGroups(superblock *SuperBlock, path string, user string) error {
	user = strings.ReplaceAll(user, "\x00", "")
	// Deserializar el bloque de carpeta raíz para encontrar el inodo de users.txt
	rootBlock := &FolderBlock{}
	// Deserializar el bloque de carpeta raíz
	err := rootBlock.Deserialize(path, int64(superblock.S_block_start+0))
	if err != nil {
		return fmt.Errorf("error al leer el bloque de carpeta raíz: %v", err)
	}

	// Buscar el inodo asociado a users.txt
	// Buscar el inodo de users.txt
	var usersInodeIndex int32 = -1
	for _, content := range rootBlock.B_content {
		fileName := strings.TrimRight(string(content.B_name[:]), "\x00")
		//fmt.Println("AAAAA FILENAME: ", fileName)
		if fileName == "users.txt" {
			usersInodeIndex = content.B_inodo
			break
		}
	}

	if usersInodeIndex == -1 {
		return errors.New("users.txt no encontrado en la carpeta raíz")
	}

	// Leer el inodo de users.txt

	usersInode := &Inode{}
	InodoStart := superblock.S_inode_start + usersInodeIndex*superblock.S_inode_size
	err = usersInode.Deserialize(path, int64(InodoStart))
	if err != nil {
		return fmt.Errorf("error al leer el inodo de users.txt: %v", err)
	}
	//characters := []rune(user)
	truthyString := strings.ReplaceAll(user, "\x00", "")
	fmt.Println("TRUTHY STRING HOY: ", truthyString)
	// Rellenar los bloques
	var j int = 0
	for _, indexBlock := range usersInode.I_block {
		if indexBlock == -1 {
			break
		}
		block := &FileBlock{}
		err = block.Deserialize(path, int64(superblock.S_block_start+(indexBlock*superblock.S_block_size)))
		if err != nil {
			return fmt.Errorf("error al leer el bloque de users.txt: %v", err)
		}
		// Actualizar el bloque con el nuevo contenido
		if len(truthyString) > 64 {
			subString := truthyString[:64]
			copy(block.B_content[:], subString)
			truthyString = truthyString[64:]
		} else {
			for i := 0; i < len(block.B_content); i++ {
				block.B_content[i] = '\x00'
			}

			copy(block.B_content[:], truthyString)
			truthyString = ""
		}
		err = block.Serialize(path, int64(superblock.S_block_start+(indexBlock*superblock.S_block_size)))
		if err != nil {
			return fmt.Errorf("error al actualizar el bloque de users.txt: %v", err)
		}
		j++
	}
	var i int = 0
	fmt.Println("TRUTHY STRING: ", truthyString)
	for _, indexBlock := range usersInode.I_block {
		if indexBlock == -1 && len(truthyString) > 0 {
			if len(truthyString) > 64 {
				block := &FileBlock{}
				block.B_content = [64]byte{}
				subString := string(truthyString[:64])
				copy(block.B_content[:], subString)
				usersInode.I_block[i] = superblock.S_blocks_count
				err = block.Serialize(path, int64(superblock.S_block_start+(superblock.S_blocks_count*superblock.S_block_size)))
				if err != nil {
					return fmt.Errorf("error al actualizar el bloque de users.txt: %v", err)
				}
				truthyString = string(truthyString[64:])
				superblock.S_blocks_count++
				superblock.S_free_blocks_count--
				superblock.S_first_blo += superblock.S_block_size
			} else {
				fmt.Println("HOLAAAAAAAAAAA")
				block := &FileBlock{}
				block.B_content = [64]byte{}
				copy(block.B_content[:], truthyString)
				fmt.Println("CONTENNNNT: ", string(block.B_content[:]))
				usersInode.I_block[i] = superblock.S_blocks_count
				fmt.Println("INDEX BLOCK: ", usersInode.I_block[i])
				err = block.Serialize(path, int64(superblock.S_block_start+(superblock.S_blocks_count*superblock.S_block_size)))
				if err != nil {
					return fmt.Errorf("error al actualizar el bloque de users.txt: %v", err)
				}
				superblock.S_blocks_count++
				superblock.S_free_blocks_count--
				superblock.S_first_blo += superblock.S_block_size
				truthyString = ""
			}
		}
		i++
	}
	// AQUI VA LO DE LOS INDIRECTOS
	/*if len(truthyString) > 0 {
		var k int = 0
		for k = 12; k < 15; k++ {
			fmt.Println("ESTE ES EL TRUTHY STRING: ", truthyString)
			_, err := superblock.indirectBlock(path, truthyString, usersInode, k)
			if err != nil {
				return err
			}
			break

		}
	}*/

	// Actualizar el tamaño del archivo en el inodo
	usersInode.I_size = int32(len(user))
	usersInode.I_mtime = float32(time.Now().Unix())

	// Serializar el inodo actualizado
	err = usersInode.Serialize(path, int64(superblock.S_inode_start+(usersInodeIndex*superblock.S_inode_size)))
	if err != nil {
		return fmt.Errorf("error al actualizar el inodo de users.txt: %v", err)
	}

	fullString, err := superblock.PrintUsersFileContent(path)
	usersInode.Print()
	if err != nil {
		return err
	}
	fmt.Println("FULL STRING AHHHH: ", fullString)

	return nil
}

func (sb *SuperBlock) UpdateUsersFile(path string, target string, words string) error {
	rootBlock := &FolderBlock{}
	// Deserializar el bloque de carpeta raíz
	err := rootBlock.Deserialize(path, int64(sb.S_block_start+0))
	if err != nil {
		return fmt.Errorf("error al leer el bloque de carpeta raíz: %v", err)
	}

	// Buscar el inodo de users.txt
	var usersInodeIndex int32 = -1
	for _, content := range rootBlock.B_content {
		fileName := strings.TrimRight(string(content.B_name[:]), "\x00")
		if fileName == "users.txt" {
			usersInodeIndex = content.B_inodo
			break
		}
	}

	if usersInodeIndex == -1 {
		return fmt.Errorf("no se encontró el inodo de users.txt")
	}

	// Leer el inodo de users.txt
	usersInode := &Inode{}
	err = usersInode.Deserialize(path, int64(sb.S_inode_start+(usersInodeIndex*sb.S_inode_size)))
	if err != nil {
		return fmt.Errorf("error al leer el inodo de users.txt: %v", err)
	}

	// Leer y actualizar el contenido del archivo users.txt
	var updatedContent strings.Builder
	found := false

	// Actualizar el contenido del bloque
	blockContent := strings.Split(string(words), "\n")
	fmt.Println("PALABRAAAAS: ", words)
	for _, line := range blockContent {
		parts := strings.Split(line, ",")
		fmt.Println("PARTS LENGTH: ", len(parts))
		if len(parts) > 1 && len(parts) < 4 {
			fmt.Println("PARTS: ", parts)
			if parts[2] == target {
				found = true
				parts[0] = "0"
				updatedContent.WriteString(strings.Join(parts, ",") + "\n")
			}

		} else {
			// Agregar la línea al contenido actualizado
			updatedContent.WriteString(line + "\n")
			fmt.Println("LINEA: ", line)
		}
	}

	fmt.Println("Updated Content Bytes: ", len(updatedContent.String()))
	fmt.Println("Updated Content: ", updatedContent.String())
	characters := []rune(updatedContent.String())
	truthyString := updatedContent.String()
	// Rellenar los bloques
	for _, indexBlock := range usersInode.I_block {
		if indexBlock == -1 {
			break
		}
		block := &FileBlock{}
		err = block.Deserialize(path, int64(sb.S_block_start+(indexBlock*sb.S_block_size)))
		if err != nil {
			return fmt.Errorf("error al leer el bloque de users.txt: %v", err)
		}
		// Actualizar el bloque con el nuevo contenido
		if len(truthyString) > 64 {
			subString := string(characters[:64])
			copy(block.B_content[:], subString)
			truthyString = string(characters[64:])
		} else {
			copy(block.B_content[:], truthyString)

		}
		err = block.Serialize(path, int64(sb.S_block_start+(indexBlock*sb.S_block_size)))
		if err != nil {
			return fmt.Errorf("error al actualizar el bloque de users.txt: %v", err)
		}
	}

	if !found {
		return fmt.Errorf("no se encontró el usuario/grupo con el id %s", target)
	}

	// Actualizar la última fecha de modificación del inodo
	usersInode.I_mtime = float32(time.Now().Unix())
	err = usersInode.Serialize(path, int64(sb.S_inode_start+(usersInodeIndex*sb.S_inode_size)))
	if err != nil {
		return fmt.Errorf("error al actualizar el inodo de users.txt: %v", err)
	}

	return nil
}

func (sb *SuperBlock) UpdateUsersFileUsers(path string, target string, words string) error {
	rootBlock := &FolderBlock{}
	// Deserializar el bloque de carpeta raíz
	err := rootBlock.Deserialize(path, int64(sb.S_block_start+0))
	if err != nil {
		return fmt.Errorf("error al leer el bloque de carpeta raíz: %v", err)
	}

	// Buscar el inodo de users.txt
	var usersInodeIndex int32 = -1
	for _, content := range rootBlock.B_content {
		fileName := strings.TrimRight(string(content.B_name[:]), "\x00")
		if fileName == "users.txt" {
			usersInodeIndex = content.B_inodo
			break
		}
	}

	if usersInodeIndex == -1 {
		return fmt.Errorf("no se encontró el inodo de users.txt")
	}

	// Leer el inodo de users.txt
	usersInode := &Inode{}
	err = usersInode.Deserialize(path, int64(sb.S_inode_start+(usersInodeIndex*sb.S_inode_size)))
	if err != nil {
		return fmt.Errorf("error al leer el inodo de users.txt: %v", err)
	}

	// Leer y actualizar el contenido del archivo users.txt
	var updatedContent strings.Builder
	found := false

	// Actualizar el contenido del bloque
	blockContent := strings.Split(string(words), "\n")
	fmt.Println("PALABRAAAAS: ", words)
	for _, line := range blockContent {
		parts := strings.Split(line, ",")
		fmt.Println("PARTS LENGTH: ", len(parts))
		if len(parts) > 4 {
			fmt.Println("PARTS: ", parts)
			if parts[3] == target {
				found = true
				parts[0] = "0" // Marcando como eliminado
				updatedContent.WriteString(strings.Join(parts, ",") + "\n")
			} else {
				updatedContent.WriteString(line + "\n")
			}
		} else {
			// Agregar la línea al contenido actualizado
			updatedContent.WriteString(line + "\n")
		}
	}

	if !found {
		return fmt.Errorf("no se encontró el usuario/grupo con el id %s", target)
	}

	// Escribir el contenido actualizado en bloques
	remainingContent := updatedContent.String()
	for _, indexBlock := range usersInode.I_block {
		if indexBlock == -1 || len(remainingContent) == 0 {
			break
		}
		block := &FileBlock{}
		err = block.Deserialize(path, int64(sb.S_block_start+(indexBlock*sb.S_block_size)))
		if err != nil {
			return fmt.Errorf("error al leer el bloque de users.txt: %v", err)
		}

		blockSize := len(block.B_content)
		if len(remainingContent) > blockSize {
			copy(block.B_content[:], remainingContent[:blockSize])
			remainingContent = remainingContent[blockSize:]
		} else {
			copy(block.B_content[:], remainingContent)
			remainingContent = ""
		}

		err = block.Serialize(path, int64(sb.S_block_start+(indexBlock*sb.S_block_size)))
		if err != nil {
			return fmt.Errorf("error al actualizar el bloque de users.txt: %v", err)
		}
	}

	// Actualizar la última fecha de modificación del inodo
	usersInode.I_mtime = float32(time.Now().Unix())
	err = usersInode.Serialize(path, int64(sb.S_inode_start+(usersInodeIndex*sb.S_inode_size)))
	if err != nil {
		return fmt.Errorf("error al actualizar el inodo de users.txt: %v", err)
	}

	return nil
}

func (sb *SuperBlock) UpdateGroupForUser(path string, target string, group string, words string) error {
	rootBlock := &FolderBlock{}
	// Deserializar el bloque de carpeta raíz
	err := rootBlock.Deserialize(path, int64(sb.S_block_start+0))
	if err != nil {
		return fmt.Errorf("error al leer el bloque de carpeta raíz: %v", err)
	}

	// Buscar el inodo de users.txt
	var usersInodeIndex int32 = -1
	for _, content := range rootBlock.B_content {
		fileName := strings.TrimRight(string(content.B_name[:]), "\x00")
		if fileName == "users.txt" {
			usersInodeIndex = content.B_inodo
			break
		}
	}

	if usersInodeIndex == -1 {
		return fmt.Errorf("no se encontró el inodo de users.txt")
	}

	// Leer el inodo de users.txt
	usersInode := &Inode{}
	err = usersInode.Deserialize(path, int64(sb.S_inode_start+(usersInodeIndex*sb.S_inode_size)))
	if err != nil {
		return fmt.Errorf("error al leer el inodo de users.txt: %v", err)
	}

	// Leer y actualizar el contenido del archivo users.txt
	var updatedContent strings.Builder
	found := false

	// Actualizar el contenido del bloque
	blockContent := strings.Split(string(words), "\n")
	fmt.Println("PALABRAAAAS: ", words)
	for _, line := range blockContent {
		parts := strings.Split(line, ",")
		fmt.Println("PARTS LENGTH: ", len(parts))
		if len(parts) > 4 {
			fmt.Println("PARTS: ", parts)
			if parts[3] == target {
				found = true
				parts[2] = group // Marcando como eliminado
				updatedContent.WriteString(strings.Join(parts, ",") + "\n")
			} else {
				updatedContent.WriteString(line + "\n")
			}
		} else {
			// Agregar la línea al contenido actualizado
			updatedContent.WriteString(line + "\n")
		}
	}

	if !found {
		return fmt.Errorf("no se encontró el usuario/grupo con el id %s", target)
	}

	// Escribir el contenido actualizado en bloques
	remainingContent := updatedContent.String()
	for _, indexBlock := range usersInode.I_block {
		if indexBlock == -1 || len(remainingContent) == 0 {
			break
		}
		block := &FileBlock{}
		err = block.Deserialize(path, int64(sb.S_block_start+(indexBlock*sb.S_block_size)))
		if err != nil {
			return fmt.Errorf("error al leer el bloque de users.txt: %v", err)
		}

		blockSize := len(block.B_content)
		if len(remainingContent) > blockSize {
			copy(block.B_content[:], remainingContent[:blockSize])
			remainingContent = remainingContent[blockSize:]
		} else {
			copy(block.B_content[:], remainingContent)
			remainingContent = ""
		}

		err = block.Serialize(path, int64(sb.S_block_start+(indexBlock*sb.S_block_size)))
		if err != nil {
			return fmt.Errorf("error al actualizar el bloque de users.txt: %v", err)
		}
	}

	// Actualizar la última fecha de modificación del inodo
	usersInode.I_mtime = float32(time.Now().Unix())
	err = usersInode.Serialize(path, int64(sb.S_inode_start+(usersInodeIndex*sb.S_inode_size)))
	if err != nil {
		return fmt.Errorf("error al actualizar el inodo de users.txt: %v", err)
	}

	return nil
}

/*func (superblock *SuperBlock) indirectBlock(path string, userText string, InodeUser *Inode, counter int) (string, error) {
	userText = strings.ReplaceAll(userText, "\x00", "")
	switch counter {
	case 12:
		var blockIndirect *PointerBlock
		if InodeUser.I_block[counter] == -1 {
			blockIndirect = &PointerBlock{[16]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}}
			InodeUser.I_block[counter] = superblock.S_blocks_count

			//Serializar el bloque indirecto
			err := blockIndirect.Serialize(path, int64(superblock.S_first_blo))
			if err != nil {
				return "", fmt.Errorf("error al serializar el bloque indirecto: %v", err)
			}

			superblock.S_blocks_count++
			superblock.S_free_blocks_count--
			superblock.S_first_blo += superblock.S_block_size

		}

		blockStarter := superblock.S_block_start + (InodeUser.I_block[counter] * superblock.S_block_size)
		blockIndirect = &PointerBlock{}
		err := blockIndirect.Deserialize(path, int64(blockStarter))
		fmt.Println("BLOQUE INDIRECTO PIOLIN: ", blockIndirect.P_pointers, "PASTEL: ", blockStarter)
		if err != nil {
			return "", fmt.Errorf("error al deserializar el bloque indirecto: %v", err)
		}

		for _, indexBlock := range blockIndirect.P_pointers {
			if indexBlock == -1 {
				break
			}
			fmt.Println("FUNCION INDIRECT: ", indexBlock)
			block := &FileBlock{}
			err = block.Deserialize(path, int64(superblock.S_block_start+(indexBlock*superblock.S_block_size)))
			if err != nil {
				return "", fmt.Errorf("error al leer el bloque de users.txt: %v", err)
			}
			// Actualizar el bloque con el nuevo contenido
			if len(userText) > 64 {
				subString := userText[:64]
				copy(block.B_content[:], subString)
				userText = userText[64:]
			} else {
				for i := 0; i < len(block.B_content); i++ {
					block.B_content[i] = '\x00'
				}

				copy(block.B_content[:], userText)
				userText = ""
			}
			err = block.Serialize(path, int64(superblock.S_block_start+(indexBlock*superblock.S_block_size)))
			if err != nil {
				return "", fmt.Errorf("error al actualizar el bloque de users.txt: %v", err)
			}
		}
		var i int = 0
		fmt.Println("TRUTHY STRING: ", userText)
		for _, indexBlock := range blockIndirect.P_pointers {
			if indexBlock == -1 && i < 12 && len(userText) > 0 {
				if len(userText) > 64 {
					block := &FileBlock{}
					block.B_content = [64]byte{}
					subString := string(userText[:64])
					copy(block.B_content[:], subString)
					blockIndirect.P_pointers[i] = superblock.S_blocks_count
					err = block.Serialize(path, int64(superblock.S_first_blo))
					if err != nil {
						return "", fmt.Errorf("error al actualizar el bloque de users.txt: %v", err)
					}
					userText = string(userText[64:])
					superblock.S_blocks_count++
					superblock.S_free_blocks_count--
					superblock.S_first_blo += superblock.S_block_size
				} else {
					fmt.Println("HOLAAAAAAAAAAA")
					block := &FileBlock{}
					block.B_content = [64]byte{}
					copy(block.B_content[:], userText)
					fmt.Println("CONTENNNNT: ", string(block.B_content[:]))
					blockIndirect.P_pointers[i] = superblock.S_blocks_count
					fmt.Println("INDEX BLOCK: ", blockIndirect.P_pointers[i])
					err = block.Serialize(path, int64(superblock.S_first_blo))
					if err != nil {
						return "", fmt.Errorf("error al actualizar el bloque de users.txt: %v", err)
					}
					superblock.S_blocks_count++
					superblock.S_free_blocks_count--
					superblock.S_first_blo += superblock.S_block_size
					userText = ""
				}
			}
			fmt.Println("COUNT COUNT: ", superblock.S_blocks_count)
			i++
		}
		fmt.Println("BLOQUE INDIRECTO PIOLIN: ", blockIndirect.P_pointers, userText, "AGUA: ", superblock.S_block_start+(InodeUser.I_block[counter]*superblock.S_block_size))
		err = blockIndirect.Serialize(path, int64(superblock.S_block_start+(InodeUser.I_block[counter]*superblock.S_block_size)))
		if err != nil {
			return "", fmt.Errorf("error al actualizar el bloque de users.txt: %v", err)
		}

	}

	return "", nil
}*/

func (sb *SuperBlock) CreateFolder(path string, parentsDir []string, destDir string) error {
	// Si parentsDir está vacío, solo trabajar con el primer inodo que sería el raíz "/"
	if len(parentsDir) == 0 {
		return sb.createFolderInInode(path, 0, parentsDir, destDir)
	}

	// Iterar sobre cada inodo ya que se necesita buscar el inodo padre
	for i := int32(0); i < sb.S_inodes_count; i++ {
		err := sb.createFolderInInode(path, i, parentsDir, destDir)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateFile crea un archivo en el sistema de archivos
func (sb *SuperBlock) CreateFile(path string, parentsDir []string, destFile string, size int, cont []string) error {

	// Si parentsDir está vacío, solo trabajar con el primer inodo que sería el raíz "/"
	if len(parentsDir) == 0 {
		return sb.createFileInInode(path, 0, parentsDir, destFile, size, cont)
	}

	// Iterar sobre cada inodo ya que se necesita buscar el inodo padre
	for i := int32(0); i < sb.S_inodes_count; i++ {
		err := sb.createFileInInode(path, i, parentsDir, destFile, size, cont)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sb *SuperBlock) DirectoryExists(partitionPath string, dirPath string) bool {
	// Dividir la ruta en partes para verificar cada directorio
	dirs := strings.Split(dirPath, "/")
	fmt.Println("Buscando ruta:", dirs)

	// Leer el bloque de carpeta raíz
	currentBlock := &FolderBlock{}
	err := currentBlock.Deserialize(partitionPath, int64(sb.S_block_start))
	if err != nil {
		fmt.Println("Error al deserializar el bloque raíz:", err)
		return false
	}

	// Recorrer cada parte de la ruta
	for _, dir := range dirs {
		if dir == "" {
			continue // Si hay una cadena vacía (por ejemplo, al inicio), saltarla
		}

		fmt.Println("Buscando directorio:", dir)
		found := false
		for _, content := range currentBlock.B_content {
			// Convertir el nombre del contenido a string
			contentName := strings.TrimRight(string(content.B_name[:]), "\x00")

			fmt.Println("Comparando con:", contentName)
			// Si se encuentra el directorio, cargar su bloque y continuar con la búsqueda
			if contentName == dir {
				fmt.Println("Directorio encontrado:", dir)
				// Encontramos el directorio, cargar el siguiente bloque de carpeta
				nextBlock := &FolderBlock{}
				err := nextBlock.Deserialize(partitionPath, int64(content.B_inodo))
				if err != nil {
					fmt.Println("Error al deserializar el siguiente bloque:", err)
					return false
				}
				currentBlock = nextBlock
				found = true
				break
			}
		}

		// Si el directorio no fue encontrado en este nivel, retornar falso
		if !found {
			fmt.Println("Directorio no encontrado:", dir)
			return false
		}
	}

	// Si se recorrieron todos los directorios y se encontraron, retornar true
	fmt.Println("Ruta encontrada:", dirPath)
	return true
}
