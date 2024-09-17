package Commands

import (
	global "archivos_pro1/global"
	"archivos_pro1/reports"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type REP struct {
	id   string // ID del disco
	path string // Ruta del archivo del disco
	name string // Nombre del reporte
	ruta string // Ruta del archivo ls (opcional)
}

// ParserRep parsea el comando rep y devuelve una instancia de REP
func ParserRep(tokens []string) (string, error) {
	cmd := &REP{} // Crea una nueva instancia de REP

	// Unir tokens en una sola cadena y luego dividir por espacios, respetando las comillas
	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar los parámetros del comando rep
	re := regexp.MustCompile(`-id=[^\s]+|-path="[^"]+"|-path=[^\s]+|-name=[^\s]+|-ruta="[^"]+"|-ruta=[^\s]+`)
	// Encuentra todas las coincidencias de la expresión regular en la cadena de argumentos
	matches := re.FindAllString(args, -1)

	// Itera sobre cada coincidencia encontrada
	for _, match := range matches {
		// Divide cada parte en clave y valor usando "=" como delimitador
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		// Remove quotes from value if present
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Switch para manejar diferentes parámetros
		switch key {
		case "-id":
			// Verifica que el id no esté vacío
			if value == "" {
				return "", errors.New("the id cannot be empty")
			}
			cmd.id = value
		case "-path":
			// Verifica que el path no esté vacío
			if value == "" {
				return "", errors.New("the path cannot be empty")
			}
			cmd.path = value
		case "-name":
			// Verifica que el nombre sea uno de los valores permitidos
			validNames := []string{"mbr", "disk", "inode", "block", "bm_inode", "bm_block", "sb", "file", "ls"}
			if !contains(validNames, value) {
				return "", errors.New("name must be one of: mbr, disk, inode, block, bm_inode, bm_block, sb, file, ls")
			}
			cmd.name = value
		case "-ruta":
			cmd.ruta = value
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return "", fmt.Errorf("unknown parameter: %s", key)
		}
	}

	// Verifica que los parámetros obligatorios hayan sido proporcionados
	if cmd.id == "" || cmd.path == "" || cmd.name == "" {
		return "", errors.New("there are missing required parameters")
	}

	// Aquí se puede agregar la lógica para ejecutar el comando rep con los parámetros proporcionados
	err := commandRep(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return "REP: Report for " + cmd.name + " created successfully", nil
}

// Función auxiliar para verificar si un valor está en una lista
func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

// Ejemplo de función commandRep (debe ser implementada)
func commandRep(rep *REP) error {
	// Obtener la partición montada
	mountedMbr, mountedSb, mountedDiskPath, err := global.GetMountedPartitionRep(rep.id)
	if err != nil {
		return err
	}

	// Switch para manejar diferentes tipos de reportes
	switch rep.name {
	case "mbr":
		err = reports.ReportMBR(mountedMbr, rep.path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "inode":
		err = reports.ReportInode(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "bm_inode":
		err = reports.ReportBMInode(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	case "disk":

		dotPath := strings.TrimSuffix(mountedDiskPath, ".mia") + "_fdisk_report.dot"
		imgPath := strings.TrimSuffix(mountedDiskPath, ".mia") + "_fdisk_report.png"

		// Comando para convertir el archivo DOT en una imagen usando Graphviz
		cmd := exec.Command("dot", "-Tpng", dotPath, "-o", imgPath)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("error generating disk image: %v", err)
		}

		// Mensaje de éxito
		fmt.Printf("Imagen del reporte de disco creada en: %s\n", imgPath)

	case "sb":
		err = reports.ReportSuperblock(mountedSb, rep.path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	case "block":
		err = reports.ReportBlock(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	case "bm_block":
		err = reports.ReportBMBlock(mountedSb, mountedDiskPath, rep.path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	return nil
}
