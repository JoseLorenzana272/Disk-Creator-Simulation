package Commands

import (
	structures "archivos_pro1/Structures"
	global "archivos_pro1/global"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var groupCount = 1

type MKGRP struct {
	Name string
}

func ParserMkgrp(tokens []string) (string, error) {
	cmd := &MKGRP{}

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
				return "", errors.New("the group name cannot be empty")
			}
			cmd.Name = value
		default:
			return "", fmt.Errorf("unknown parameter: %s", key)
		}
	}

	// Verificar que el nombre del grupo no esté vacío
	if cmd.Name == "" {
		return "", errors.New("the group name cannot be empty")
	}

	// Ejecutar el comando MKGRP
	err := commandMkgrp(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return "MKGRP: Group: " + cmd.Name + " created successfully", nil
}

func commandMkgrp(mkgrp *MKGRP) error {
	if !IsLogged {
		return errors.New("necesitas iniciar sesión para ejecutar este comando")
	}

	mountedPartition, path, err := global.GetMountedPartition(IdPartitionGlobal)
	if err != nil {
		return err
	}

	//fmt.Println("Path: ", path)
	//fmt.Println("Name: ", mkgrp.Name)

	//Deserializar el superbloque
	sb := &structures.SuperBlock{}

	errorD := sb.Deserialize(path, int64(mountedPartition.Part_start))
	if errorD != nil {
		return errorD
	}

	groupID := groupCount
	groupCount++

	groupstr := strconv.Itoa(groupID)

	fmt.Println("Group ID: ", groupstr)

	//Verificar si el grupo ya existe
	if global.GroupsUser[IdPartitionGlobal][mkgrp.Name] != nil {
		return errors.New("el grupo ya existe")
	}

	//Agregar el grupo al map global
	structures.Groups[mkgrp.Name] = mkgrp.Name
	global.GroupsUser[IdPartitionGlobal][mkgrp.Name] = []string{}
	/*for key, value := range structures.Groups {
		fmt.Println("Key:", key, "Value:", value)
	}*/
	fullString, err := sb.PrintUsersFileContent(path)
	if err != nil {
		return err
	}

	newContent := fullString + groupstr + "," + "G" + "," + strings.TrimSpace(mkgrp.Name) + "\n"

	newContent = strings.ReplaceAll(newContent, "\x00", "")

	err = structures.AddUserGroups(sb, path, newContent)
	if err != nil {
		return err
	}

	//Serializar el superbloque
	err = sb.Serialize(path, int64(mountedPartition.Part_start))
	if err != nil {
		return err
	}

	return nil

}
