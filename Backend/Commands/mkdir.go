package Commands

import (
	structures "archivos_pro1/Structures"
	"archivos_pro1/global"
	utils "archivos_pro1/utils"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type MKDIR struct {
	path string
	p    bool
}

/*
   mkdir -p -path=/home/user/docs/usac
   mkdir -path="/home/mis documentos/archivos clases"
*/

func ParserMkdir(tokens []string) (string, error) {
	cmd := &MKDIR{} // Crea una nueva instancia de MKDIR

	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar los parámetros del comando mkdir
	re := regexp.MustCompile(`-path=[^\s]+|-p`)
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

		// Switch para manejar diferentes parámetros
		switch key {
		case "-path":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := kv[1]
			// Remove quotes from value if present
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
			cmd.path = value
		case "-p":
			cmd.p = true
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que el parámetro -path haya sido proporcionado
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	// Aquí se puede agregar la lógica para ejecutar el comando mkdir con los parámetros proporcionados
	err := commandMkdir(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("MKDIR: Directory %s created successfully", cmd.path), nil
}

var idPartition string

func commandMkdir(mkdir *MKDIR) error {
	idPartition = IdPartitionGlobal
	// Obtener la partición montada
	fmt.Println("ID Partition Global:", IdPartitionGlobal)
	fmt.Println("ID Partition:", idPartition)
	partitionSuperblock, mountedPartition, partitionPath, err := global.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Crear el directorio
	err = createDirectory(mkdir.path, partitionSuperblock, partitionPath, mountedPartition, mkdir.p)
	if err != nil {
		err = fmt.Errorf("error al crear el directorio: %w", err)
	}

	return err
}

func createDirectory(dirPath string, sb *structures.SuperBlock, partitionPath string, mountedPartition *structures.Partition, allowCreateParents bool) error {
	fmt.Println("\nCreando directorio:", dirPath)

	parentDirs, destDir := utils.GetParentDirectories(dirPath)
	fmt.Println("\nDirectorios padres:", parentDirs)
	fmt.Println("Directorio destino:", destDir)

	// Validar si los directorios padres existen
	if !allowCreateParents {
		for _, parent := range parentDirs {
			if !sb.DirectoryExists(partitionPath, parent) {
				return fmt.Errorf("error: el directorio padre '%s' no existe y no se especificó la opción -p", parent)
			}
		}
	} else {
		// Si la opción -p fue especificada, crea los directorios padres si no existen
		err := utils.CreateParentDirs(dirPath)
		if err != nil {
			return fmt.Errorf("error al crear las carpetas padre: %w", err)
		}
	}

	// Crear el directorio destino
	err := sb.CreateFolder(partitionPath, parentDirs, destDir)
	if err != nil {
		return fmt.Errorf("error al crear el directorio: %w", err)
	}

	// Actualizar el mapa de directorios creados
	if global.DirectoriesCreated[partitionPath] == nil {
		global.DirectoriesCreated[partitionPath] = make(map[string][]string)
	}

	// Agregar los directorios padres creados
	for _, parent := range parentDirs {
		if _, exists := global.DirectoriesCreated[partitionPath][parent]; !exists {
			global.DirectoriesCreated[partitionPath][parent] = append(global.DirectoriesCreated[partitionPath][parent], parent)
		}
	}

	// Agregar el directorio destino creado
	global.DirectoriesCreated[partitionPath][destDir] = append(global.DirectoriesCreated[partitionPath][destDir], destDir)

	// Imprimir inodos y bloques
	sb.PrintInodes(partitionPath)
	sb.PrintBlocks(partitionPath)

	// Serializar el superbloque
	err = sb.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	//Imprimir estructura de directorios
	fmt.Println("Directorios creados:", global.DirectoriesCreated)

	return nil
}
