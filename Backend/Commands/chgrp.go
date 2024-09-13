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

func ParserChgrp(tokens []string) (*CHGRP, error) {
	cmd := &CHGRP{}

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`-usuario="[^"]+"|-usuario=[^\s]+|-grp="[^"]+"|-grp=[^\s]+`)
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
		case "-usuario":
			if value == "" {
				return nil, errors.New("el usuario no puede estar vacío")
			}
			cmd.Usuario = value
		case "-grp":
			if value == "" {
				return nil, errors.New("el grupo no puede estar vacío")
			}
			cmd.Grp = value
		default:
			return nil, fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.Usuario == "" {
		return nil, errors.New("el usuario no puede estar vacío")
	}

	if cmd.Grp == "" {
		return nil, errors.New("el grupo no puede estar vacío")
	}

	err := commandChgrp(cmd)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func commandChgrp(chgrp *CHGRP) error {
	if !IsLogged {
		return errors.New("necesitas iniciar sesión para ejecutar este comando")
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
