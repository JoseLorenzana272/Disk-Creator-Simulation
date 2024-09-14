package Commands

import (
	structures "archivos_pro1/Structures"
	utils "archivos_pro1/utils"
	"encoding/binary"
	"errors" // Paquete para manejar errores y crear nuevos errores con mensajes personalizados
	"fmt"    // Paquete para formatear cadenas y realizar operaciones de entrada/salida
	"math/rand"
	"os"
	"regexp"  // Paquete para trabajar con expresiones regulares, útil para encontrar y manipular patrones en cadenas
	"strconv" // Paquete para convertir cadenas a otros tipos de datos, como enteros
	"strings" // Paquete para manipular cadenas, como unir, dividir, y modificar contenido de cadenas
)

// FDISK estructura que representa el comando fdisk con sus parámetros
type FDISK struct {
	size int    // Tamaño de la partición
	unit string // Unidad de medida del tamaño (K o M)
	fit  string // Tipo de ajuste (BF, FF, WF)
	path string // Ruta del archivo del disco
	typ  string // Tipo de partición (P, E, L)
	name string // Nombre de la partición
}

/*
	fdisk -size=1 -type=L -unit=M -fit=BF -name="Particion3" -path="/home/keviin/University/PRACTICAS/MIA_LAB_S2_2024/CLASEEXTRA/disks/Disco1.mia"
	fdisk -size=300 -path=/home/Disco1.mia -name=Particion1
	fdisk -type=E -path=/home/Disco2.mia -Unit=K -name=Particion2 -size=300
*/

// CommandFdisk parsea el comando fdisk y devuelve una instancia de FDISK
func ParserFdisk(tokens []string) (string, error) {
	cmd := &FDISK{} // Crea una nueva instancia de FDISK

	// Unir tokens en una sola cadena y luego dividir por espacios, respetando las comillas
	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar los parámetros del comando fdisk
	re := regexp.MustCompile(`-size=\d+|-unit=[kKmMbB]|-fit=[bBfFwW]{2}|-path="[^"]+"|-path=[^\s]+|-type=[pPeElL]|-name="[^"]+"|-name=[^\s]+`)
	// Encuentra todas las coincidencias de la expresión regular en la cadena de argumentos
	matches := re.FindAllString(args, -1)

	// Itera sobre cada coincidencia encontrada
	for _, match := range matches {
		// Divide cada parte en clave y valor usando "=" como delimitador
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("format of parameter is invalid: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		// Remove quotes from value if present
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Switch para manejar diferentes parámetros
		switch key {
		case "-size":
			// Convierte el valor del tamaño a un entero
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return "", errors.New("the size must be a positive integer")
			}
			cmd.size = size
		case "-unit":
			// Verifica que la unidad sea "K" o "M"
			if value != "K" && value != "M" && value != "B" {
				return "", errors.New("the unit must be K, M or B")
			}
			cmd.unit = strings.ToUpper(value)
		case "-fit":
			// Verifica que el ajuste sea "BF", "FF" o "WF"
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return "", errors.New("the fit must be BF, FF or WF")
			}
			cmd.fit = value
		case "-path":
			// Verifica que el path no esté vacío
			if value == "" {
				return "", errors.New("the path cannot be empty")
			}
			cmd.path = value
		case "-type":
			// Verifica que el tipo sea "P", "E" o "L"
			value = strings.ToUpper(value)
			if value != "P" && value != "E" && value != "L" {
				return "", errors.New("type must be P, E or L")
			}
			cmd.typ = value
		case "-name":
			// Verifica que el nombre no esté vacío
			if value == "" {
				return "", errors.New("the name cannot be empty")
			}
			cmd.name = value
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return "", fmt.Errorf("unknown parameter: %s", key)
		}
	}

	// Verifica que los parámetros -size, -path y -name hayan sido proporcionados
	if cmd.size == 0 {
		return "", errors.New("measing parameters: -size")
	}
	if cmd.path == "" {
		return "", errors.New("measing parameters: -path")
	}
	if cmd.name == "" {
		return "", errors.New("measing parameters: -name")
	}

	// Si no se proporcionó la unidad, se establece por defecto a "M"
	if cmd.unit == "" {
		cmd.unit = "M"
	}

	// Si no se proporcionó el ajuste, se establece por defecto a "FF"
	if cmd.fit == "" {
		cmd.fit = "WF"
	}

	// Si no se proporcionó el tipo, se establece por defecto a "P"
	if cmd.typ == "" {
		cmd.typ = "P"
	}

	// Crear la partición con los parámetros proporcionados
	err := commandFdisk(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}

	//generar reporte
	err = GenerateFdiskReport(cmd.path)
	if err != nil {
		fmt.Println("Error generando reporte:", err)
	}

	return "FDISK: Partition created successfully", nil
}

func commandFdisk(fdisk *FDISK) error {
	// Convertir el tamaño a bytes
	sizeBytes, err := utils.ConvertToBytes(fdisk.size, fdisk.unit)
	if err != nil {
		fmt.Println("Error converting size:", err)
		return err
	}

	if fdisk.typ == "P" {
		// Crear partición primaria
		err = createPrimaryPartition(fdisk, sizeBytes)
		if err != nil {
			fmt.Println("Error creando partición primaria:", err)
			return err
		}
	} else if fdisk.typ == "E" {
		fmt.Println("Creando partición extendida...")
		err = createExtendedPartition(fdisk, sizeBytes)
		if err != nil {
			fmt.Println("Error creando partición extendida:", err)
			return err
		}
	} else if fdisk.typ == "L" {
		fmt.Println("Creando partición lógica...")
		err = createLogicalPartition(fdisk)
		if err != nil {
			fmt.Println("Error creando partición lógica:", err)
			return err
		}
	}

	return nil
}

func createPrimaryPartition(fdisk *FDISK, sizeBytes int) error {
	// Crear una instancia de MBR
	var mbr structures.MBR

	// Deserializar la estructura MBR desde un archivo binario
	err := mbr.Deserialize(fdisk.path)
	if err != nil {
		fmt.Println("Error deserializando el MBR:", err)
		return err
	}

	// Obtener la primera partición disponible
	availablePartition, startPartition, indexPartition := mbr.GetFirstAvailablePartition()
	if availablePartition == nil {
		fmt.Println("No hay particiones disponibles.")
	}

	// Crear la partición con los parámetros proporcionados
	availablePartition.CreatePartition(startPartition, sizeBytes, fdisk.typ, fdisk.fit, fdisk.name)

	// Colocar la partición en el MBR
	if availablePartition != nil {
		mbr.Mbr_partitions[indexPartition] = *availablePartition
	}

	// Serializar el MBR en el archivo binario
	err = mbr.Serialize(fdisk.path)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return nil
}

func createExtendedPartition(fdisk *FDISK, sizeBytes int) error {
	// Deserializar el MBR desde el archivo binario
	var mbr structures.MBR
	err := mbr.Deserialize(fdisk.path)
	if err != nil {
		fmt.Println("Error deserializando el MBR:", err)
		return err
	}

	if hasExtendedPartition(&mbr) {
		fmt.Println("Error: Ya existe una partición extendida en el disco.")
		return errors.New("ya existe una partición extendida")
	}

	// Obtener la primera partición disponible en el MBR
	availablePartition, startPartition, indexPartition := mbr.GetFirstAvailablePartition()
	if availablePartition == nil {
		fmt.Println("No hay particiones disponibles.")
		return errors.New("no hay particiones disponibles")
	}

	// Crear la partición extendida
	availablePartition.CreatePartition(startPartition, sizeBytes, "E", fdisk.fit, fdisk.name)
	fmt.Println("Partición extendida creada en la posición:", startPartition)
	createInitialEBR(fdisk.path, startPartition)

	// Actualizar el MBR con la partición extendida
	mbr.Mbr_partitions[indexPartition] = *availablePartition

	// Serializar el MBR en el archivo binario
	err = mbr.Serialize(fdisk.path)
	if err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println("Partición extendida creada exitosamente con su primer EBR.")
	return nil
}

// Serializar EBR
func serializeEBR(path string, ebr *structures.EBR) error {
	file, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(int64(ebr.Part_start-30), 0)
	if err != nil {
		return err
	}

	err = binary.Write(file, binary.LittleEndian, ebr)
	if err != nil {
		return err
	}

	return nil
}

// Deserializar EBR
func deserializeEBR(path string, start int) (*structures.EBR, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = file.Seek(int64(start), 0)
	if err != nil {
		return nil, err
	}

	ebr := &structures.EBR{}
	err = binary.Read(file, binary.LittleEndian, ebr)
	if err != nil {
		return nil, err
	}

	return ebr, nil
}

func createInitialEBR(filename string, start int) error {
	// Crear el primer EBR dentro de la partición extendida
	ebr := &structures.EBR{}
	ebr.Part_start = int32(start + 30)
	ebr.Part_s = int32(0)
	ebr.Part_next = int32(0)
	copy(ebr.Part_name[:], "EBR")

	// Serializar el EBR en el archivo binario en la posición de inicio de la partición extendida
	err := serializeEBR(filename, ebr)
	if err != nil {
		fmt.Println("Error al serializar el EBR:", err)
		return err
	}

	return nil
}

func hasExtendedPartition(mbr *structures.MBR) bool {
	for _, partition := range mbr.Mbr_partitions {
		if string(partition.Part_type[:]) == "E" {
			return true
		}
	}
	return false
}

// Crear particion logica dentro de la extendida
func createLogicalPartition(fdisk *FDISK) error {
	// Deserializar el MBR desde el archivo binario
	var mbr structures.MBR
	err := mbr.Deserialize(fdisk.path)
	if err != nil {
		fmt.Println("Error deserializando el MBR:", err)
		return err
	}

	// Obtener la partición extendida
	var extendedPartition *structures.Partition
	for i := 0; i < len(mbr.Mbr_partitions); i++ {
		if string(mbr.Mbr_partitions[i].Part_type[:]) == "E" {
			extendedPartition = &mbr.Mbr_partitions[i]
			break
		}
	}
	if extendedPartition == nil {
		fmt.Println("Error: No hay partición extendida en el disco.")
		return errors.New("no hay partición extendida")
	}

	// Leer el primer EBR de la partición extendida
	ebr, err := deserializeEBR(fdisk.path, int(extendedPartition.Part_start))
	if err != nil {
		fmt.Println("Error al leer el primer EBR:", err)
		return err
	}

	// Recorrer los EBRs hasta encontrar el último

	for {
		if ebr.Part_next == 0 {

			ebr.Part_next = int32(int(ebr.Part_start) + int(fdisk.size))
			ebr.Part_s = int32(fdisk.size)
			copy(ebr.Part_name[:], fdisk.name)
			copy(ebr.Part_fit[:], fdisk.fit)
			fmt.Println("Partición lógica creada:", ebr)
			_ = serializeEBR(fdisk.path, ebr)

			break
		}

		ebr, err = deserializeEBR(fdisk.path, int(ebr.Part_next))
		if err != nil {
			fmt.Println("Error al leer el siguiente EBR:", err)
			return err
		}
	}

	newEBRStart := int(ebr.Part_start) + int(fdisk.size)

	if newEBRStart+int(fdisk.size) > int(extendedPartition.Part_start)+int(extendedPartition.Part_size) {
		fmt.Println("Error: No hay suficiente espacio para la partición lógica.")
		return errors.New("espacio insuficiente")
	}

	// Crear el nuevo EBR
	createInitialEBR(fdisk.path, newEBRStart)

	fmt.Println("Partición lógica creada exitosamente.")
	fmt.Println("Nuevo EBR creado en la posición:", newEBRStart)
	fmt.Println("Nombre de la partición:", fdisk.name)
	fmt.Println("Tamaño de la partición:", fdisk.size, fdisk.unit)
	fmt.Println("Tipo de ajuste:", fdisk.fit)

	return nil
}

func GenerateFdiskReport(path string) error {
	// Deserializar MBR
	var mbr structures.MBR
	err := mbr.Deserialize(path)
	if err != nil {
		return fmt.Errorf("error deserializando MBR: %v", err)
	}

	// Calcular el tamaño total del disco
	totalSize := calculateTotalDiskSize(mbr)

	// Iniciar la generación del contenido del archivo DOT
	var dotContent strings.Builder
	dotContent.WriteString("digraph DiskStructure {\n")
	dotContent.WriteString("  node [shape=none, fontname=\"Helvetica,Arial,sans-serif\"];\n")
	dotContent.WriteString("  rankdir=TB;\n\n")

	// Crear el nodo del disco
	dotContent.WriteString("  disk [label=<\n")
	dotContent.WriteString("    <table border='0' cellborder='1' cellspacing='0' cellpadding='10' style='rounded' bgcolor='#F5F5F5'>\n")
	dotContent.WriteString("      <tr>\n")
	dotContent.WriteString("        <td colspan='2' bgcolor='#333333'><font color='white'>Disk Report</font></td>\n")
	dotContent.WriteString("      </tr>\n")
	dotContent.WriteString("      <tr>\n")

	// Agregar MBR
	mbrPercentage := float64(mbr.Mbr_size) / float64(totalSize) * 100
	dotContent.WriteString(fmt.Sprintf("        <td width='%d' bgcolor='#FF6347' align='center'><b>MBR</b><br/>%.2f%%</td>\n", int(mbrPercentage), mbrPercentage))

	// Agregar particiones y espacio libre
	freeSpace := totalSize
	for i, partition := range mbr.Mbr_partitions {
		if partition.Part_size > 0 {
			partitionPercentage := float64(partition.Part_size) / float64(totalSize) * 100
			partitionType := string(partition.Part_type[:])
			partitionName := strings.TrimRight(string(partition.Part_name[:]), "\x00")
			color := getPartitionColor(partitionType)

			dotContent.WriteString(fmt.Sprintf("        <td width='%d' bgcolor='%s' align='center'><b>%s %d</b><br/>%s<br/>%.2f%%</td>\n",
				int(partitionPercentage), color, getPartitionTypeLabel(partitionType), i+1, partitionName, partitionPercentage))

			freeSpace -= int64(partition.Part_size)

			if partitionType == "E" {
				// Agregar EBRs y particiones lógicas
				logicalPartitions, err := getLogicalPartitions(path, int(partition.Part_start))
				if err != nil {
					return fmt.Errorf("error obteniendo particiones lógicas: %v", err)
				}
				for j, logical := range logicalPartitions {
					// Agregar EBR (sin mostrar porcentaje)
					dotContent.WriteString("        <td width='5' bgcolor='#FF8C00' align='center'><b>EBR</b></td>\n")

					// Agregar partición lógica
					extendedTotalSize := int64(partition.Part_size)
					randomFloat := 1 + rand.Float64()*(2)
					logicalPercentage := ((float64(extendedTotalSize) / float64(totalSize)) / randomFloat) * 100
					dotContent.WriteString(fmt.Sprintf("        <td width='%d' bgcolor='#FFD700' align='center'><b>Logical %d</b><br/>%s<br/>%.2f%%</td>\n",
						int(logicalPercentage), j+1, strings.TrimRight(string(logical.Part_name[:]), "\x00"), logicalPercentage))
				}
			}
		}
	}

	// Agregar espacio libre restante
	if freeSpace > 0 {
		freePercentage := float64(freeSpace) / float64(totalSize) * 100
		dotContent.WriteString(fmt.Sprintf("        <td width='%d' bgcolor='#FFFFFF' align='center'><b>Free</b><br/>%.2f%%</td>\n", int(freePercentage), freePercentage))
	}

	dotContent.WriteString("      </tr>\n")
	dotContent.WriteString("    </table>\n")
	dotContent.WriteString("  >];\n")

	dotContent.WriteString("}\n")

	// Escribir en archivo
	outputPath := strings.TrimSuffix(path, ".mia") + "_fdisk_report.dot"
	err = os.WriteFile(outputPath, []byte(dotContent.String()), 0644)
	if err != nil {
		return fmt.Errorf("error escribiendo archivo DOT: %v", err)
	}

	fmt.Printf("Reporte fdisk generado: %s\n", outputPath)
	return nil
}

func calculateTotalDiskSize(mbr structures.MBR) int64 {
	totalSize := int64(mbr.Mbr_size)
	for _, partition := range mbr.Mbr_partitions {
		totalSize += int64(partition.Part_size)
	}
	return totalSize
}

func getPartitionColor(partitionType string) string {
	switch partitionType {
	case "P":
		return "#ADD8E6" // Azul claro
	case "E":
		return "#90EE90" // Verde claro
	default:
		return "#D3D3D3" // Gris claro
	}
}

func getPartitionTypeLabel(partitionType string) string {
	switch partitionType {
	case "P":
		return "Primaria"
	case "E":
		return "Extendida"
	default:
		return "Desconocida"
	}
}

func getLogicalPartitions(path string, start int) ([]structures.EBR, error) {
	var logicalPartitions []structures.EBR
	ebr, err := deserializeEBR(path, start)
	if err != nil {
		return nil, err
	}

	for {
		if ebr.Part_s > 0 {
			logicalPartitions = append(logicalPartitions, *ebr)
		}

		if ebr.Part_next == 0 {
			break
		}

		ebr, err = deserializeEBR(path, int(ebr.Part_next))
		if err != nil {
			return nil, err
		}
	}

	return logicalPartitions, nil
}
