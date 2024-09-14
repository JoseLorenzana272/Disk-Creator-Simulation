package Commands

import (
	structures "archivos_pro1/Structures"
	"archivos_pro1/global"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type RMGRP struct {
	Name string
}

func ParserRmgrp(tokens []string) (string, error) {
	cmd := &RMGRP{}

	// Unir los tokens en un solo string
	args := strings.Join(tokens, " ")

	// Expresión regular para encontrar los parámetros
	re := regexp.MustCompile(`-name="[^"]+"|-name=[^\s]+`)
	matches := re.FindAllString(args, -1)

	// Recorrer los parámetros encontrados
	for _, match := range matches {
		// Separar la clave del valor
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("format of parameter is invalid: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		// Eliminar las comillas del valor si las tiene
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Evaluar la clave
		switch key {
		case "-name":
			if value == "" {
				return "", errors.New("the name of the group cannot be empty")
			}
			cmd.Name = value
		default:
			return "", fmt.Errorf("unknown parameter found: %s", key)
		}
	}

	// Verificar que el nombre del grupo no esté vacío
	if cmd.Name == "" {
		return "", errors.New("the name of the group cannot be empty")
	}

	// Ejecutar el comando RMGRP
	err := commandRmgrp(cmd)
	if err != nil {
		return "", err
	}

	return "RMGRP: Group: " + cmd.Name + " deleted succesfully", nil
}

func commandRmgrp(cmd *RMGRP) error {
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

	// Verificar si el grupo existe
	if _, exists := global.GroupsUser[IdPartitionGlobal][cmd.Name]; !exists {
		return errors.New("el grupo no existe")
	}

	// Actualizar el archivo users.txt para marcar el grupo como eliminado
	err = sb.UpdateUsersFile(path, cmd.Name, fullString)

	var verificationGroup = global.GroupsUser[IdPartitionGlobal][cmd.Name]

	for _, user := range verificationGroup {
		fullString, err = sb.PrintUsersFileContent(path)
		if err != nil {
			return err
		}
		fmt.Println("Usuario: ", user)
		sb.UpdateUsersFileUsers(path, user, fullString)
	}

	if err != nil {
		return err
	}

	// Eliminar el grupo del map global
	delete(global.GroupsUser[IdPartitionGlobal], cmd.Name)
	fmt.Println("Nombre eliminado: ", cmd.Name)

	fmt.Println("Grupo eliminado correctamente")

	return nil
}
