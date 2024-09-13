package Commands

import (
	structures "archivos_pro1/Structures"
	"archivos_pro1/global"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type RMUSR struct {
	User string
}

func ParserRmusr(tokens []string) (*RMUSR, error) {
	cmd := &RMUSR{}

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`-user="[^"]+"|-user=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-user":
			if value == "" {
				return nil, errors.New("el usuario no puede estar vacío")
			}
			cmd.User = value
		default:
			return nil, fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.User == "" {
		return nil, errors.New("el usuario no puede estar vacío")
	}

	err := commandRmusr(cmd)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func commandRmusr(cmd *RMUSR) error {
	if !IsLogged {
		return errors.New("necesitas iniciar sesión para ejecutar este comando")
	}

	mountedPartition, path, err := global.GetMountedPartition(IdPartitionGlobal)
	if err != nil {
		return err
	}

	// Deserializar el superbloque
	sb := &structures.SuperBlock{}
	err = sb.Deserialize(path, int64(mountedPartition.Part_start))
	if err != nil {
		return err
	}
	var fullString string
	fullString, err = sb.PrintUsersFileContent(path)
	if err != nil {
		return err
	}

	// Actualizar el archivo users.txt para marcar el grupo como eliminado

	err = sb.UpdateUsersFileUsers(path, cmd.User, fullString)
	fmt.Println("Path: ", path)
	if err != nil {
		return err
	}

	// Eliminar el grupo del map global
	delete(global.GroupsUser[IdPartitionGlobal], cmd.User)
	fmt.Println("Nombre eliminado: ", cmd.User)

	fmt.Println("Grupo eliminado correctamente")

	return nil
}
