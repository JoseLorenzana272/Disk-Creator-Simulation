package Commands

import (
	structures "archivos_pro1/Structures"
	"archivos_pro1/global"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type CHGRP struct {
	Usuario string
	Grp     string
}

func ParserChgrp(tokens []string) (string, error) {
	cmd := &CHGRP{}

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`-usuario="[^"]+"|-usuario=[^\s]+|-grp="[^"]+"|-grp=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("format of parameter is invalid: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-usuario":
			if value == "" {
				return "", errors.New("the user cannot be empty")
			}
			cmd.Usuario = value
		case "-grp":
			if value == "" {
				return "", errors.New("the group cannot be empty")
			}
			cmd.Grp = value
		default:
			return "", fmt.Errorf("unknown parameter: %s", key)
		}
	}

	if cmd.Usuario == "" {
		return "", errors.New("the user cannot be empty")
	}

	if cmd.Grp == "" {
		return "", errors.New("the group cannot be empty")
	}

	err := commandChgrp(cmd)
	if err != nil {
		return "", err
	}

	return "CHGRP: Group changed successfully", nil
}

func commandChgrp(chgrp *CHGRP) error {
	if !IsLogged {
		return errors.New("you must be logged to execute this command")
	}

	mountedPartition, path, err := global.GetMountedPartition(IdPartitionGlobal)
	if err != nil {
		return err
	}

	sb := &structures.SuperBlock{}

	errorD := sb.Deserialize(path, int64(mountedPartition.Part_start))
	if errorD != nil {
		return errorD
	}

	var fullString string
	fullString, err = sb.PrintUsersFileContent(path)
	if err != nil {
		return err
	}

	//Verificar si el grupo existe
	if _, exists := global.GroupsUser[IdPartitionGlobal][chgrp.Grp]; !exists {
		return errors.New("el grupo no existe")
	}

	foundUser := false

	for _, value := range global.GroupsUser[IdPartitionGlobal] {
		for _, user := range value {
			if user == chgrp.Usuario {
				foundUser = true
				break
			}
		}
	}

	if !foundUser {
		return errors.New("el usuario no existe")
	}

	//Actualizar el grupo del usuario
	err = sb.UpdateGroupForUser(path, chgrp.Usuario, chgrp.Grp, fullString)

	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	//Actualizar el map global
	global.GroupsUser[IdPartitionGlobal][chgrp.Grp] = append(global.GroupsUser[IdPartitionGlobal][chgrp.Grp], chgrp.Usuario)
	fmt.Println("Usuario agregado al grupo: ", chgrp.Grp, "Usuario: ", chgrp.Usuario)
	return nil

}
