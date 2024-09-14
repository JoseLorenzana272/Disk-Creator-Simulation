package Commands

import (
	structures "archivos_pro1/Structures"
	"archivos_pro1/global"
	utils "archivos_pro1/utils"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type MKFILE struct {
	path string // Ruta del archivo
	r    bool   // Opción recursiva
	size int    // Tamaño del archivo
	cont string // Contenido del archivo
}

func ParserMkfile(tokens []string) (string, error) {
	cmd := &MKFILE{} // Crea una nueva instancia de MKFILE

	// Unir tokens en una sola cadena y luego dividir por espacios, respetando las comillas
	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar los parámetros del comando mkfile
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-r|-size=\d+|-cont="[^"]+"|-cont=[^\s]+`)
	// Encuentra todas las coincidencias de la expresión regular en la cadena de argumentos
	matches := re.FindAllString(args, -1)

	// Verificar que todos los tokens fueron reconocidos por la expresión regular
	if len(matches) != len(tokens) {
		// Identificar el parámetro inválido
		for _, token := range tokens {
			if !re.MatchString(token) {
				return "", fmt.Errorf("parámetro inválido: %s", token)
			}
		}
	}

	// Itera sobre cada coincidencia encontrada
	for _, match := range matches {
		// Divide cada parte en clave y valor usando "=" como delimitador
		kv := strings.SplitN(match, "=", 2)
		key := strings.ToLower(kv[0])
		var value string
		if len(kv) == 2 {
			value = kv[1]
		}

		// Remove quotes from value if present
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Switch para manejar diferentes parámetros
		switch key {
		case "-path":
			// Verifica que el path no esté vacío
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-r":
			// Establece el valor de r a true
			cmd.r = true
		case "-size":
			// Convierte el valor del tamaño a un entero
			size, err := strconv.Atoi(value)
			if err != nil || size < 0 {
				return "", errors.New("el tamaño debe ser un número entero no negativo")
			}
			cmd.size = size
		case "-cont":
			// Verifica que el contenido no esté vacío
			if value == "" {
				return "", errors.New("el contenido no puede estar vacío")
			}
			cmd.cont = value
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que el parámetro -path haya sido proporcionado
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	// Si no se proporcionó el tamaño, se establece por defecto a 0
	if cmd.size == 0 {
		cmd.size = 0
	}

	// Si no se proporcionó el contenido, se establece por defecto a ""
	if cmd.cont == "" {
		cmd.cont = ""
	}

	// Crear el archivo con los parámetros proporcionados
	err := commandMkfile(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("MKFILE: Archivo %s creado correctamente.", cmd.path), nil // Devuelve el comando MKFILE creado
}

func commandMkfile(mkfile *MKFILE) error {
	// Obtener la partición montada
	partitionSuperblock, mountedPartition, partitionPath, err := global.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Generar el contenido del archivo si no se proporcionó
	if mkfile.cont == "" {
		mkfile.cont = generateContent(mkfile.size)
	}

	// Crear el archivo
	err = createFile(mkfile.path, mkfile.size, mkfile.cont, partitionSuperblock, partitionPath, mountedPartition, mkfile.r)
	if err != nil {
		err = fmt.Errorf("error al crear el archivo: %w", err)
	}

	return err
}

// generateContent genera una cadena de números del 0 al 9 hasta cumplir el tamaño ingresado
func generateContent(size int) string {
	content := ""
	for len(content) < size {
		content += "0123456789"
	}
	return content[:size] // Recorta la cadena al tamaño exacto
}

// Funcion para crear un archivo
func createFile(filePath string, size int, content string, sb *structures.SuperBlock, partitionPath string, mountedPartition *structures.Partition, allowParents bool) error {
	fmt.Println("\nCreando archivo:", filePath)

	parentDirs, destDir := utils.GetParentDirectories(filePath)
	fmt.Println("\nDirectorios padres:", parentDirs)
	fmt.Println("Directorio destino:", destDir)

	// Validar si los directorios padres existen
	if !allowParents {
		for _, parent := range parentDirs {
			if !sb.DirectoryExists(partitionPath, parent) {
				fmt.Println("Directorio padre no existe:", parent)
				return fmt.Errorf("error: el directorio padre '%s' no existe y no se especificó la opción -r", parent)
			}
		}
	} else {
		// Si la opción -r fue especificada, crea los directorios padres si no existen
		err := utils.CreateParentDirs(filePath)
		if err != nil {
			return fmt.Errorf("error al crear las carpetas padre: %w", err)
		}
	}

	// Obtener contenido por chunks
	chunks := utils.SplitStringIntoChunks(content)
	fmt.Println("\nChunks del contenido:", chunks)

	// Crear el archivo
	err := sb.CreateFile(partitionPath, parentDirs, destDir, size, chunks)
	if err != nil {
		return fmt.Errorf("error al crear el archivo: %w", err)
	}

	// Imprimir inodos y bloques
	sb.PrintInodes(partitionPath)
	sb.PrintBlocks(partitionPath)

	// Serializar el superbloque
	err = sb.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}
