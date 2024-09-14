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

var userCount = 2

type MKUSER struct {
	User string
	Pass string
	Grp  string
}

func ParserMkuser(tokens []string) (string, error) {
	cmd := &MKUSER{}

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`-user="[^"]+"|-user=[^\s]+|-pass="[^"]+"|-pass=[^\s]+ |-grp="[^"]+"|-grp=[^\s]+`)
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
		case "-user":
			if value == "" {
				return "", errors.New("the user cannot be empty")
			}
			cmd.User = value
		case "-pass":
			if value == "" {
				return "", errors.New("the password cannot be empty")
			}
			cmd.Pass = value
		case "-grp":
			if value == "" {
				return "", errors.New("the group cannot be empty")
			}
			cmd.Grp = value
		default:
			return "", fmt.Errorf("unknown parameter: %s", key)
		}
	}

	if cmd.User == "" {
		return "", errors.New("there are missing required parameters: -user")
	}

	if cmd.Pass == "" {
		return "", errors.New("there are missing required parameters: -pass")
	}

	if cmd.Grp == "" {
		return "", errors.New("there are missing required parameters: -grp")
	}

	err := commandMkuser(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return "MKUSER: User: " + cmd.User + " created successfully", nil

}

func commandMkuser(mkuser *MKUSER) error {
	if !IsLogged {
		return errors.New("you must be logged in to execute this command")
	}

	mountedPartition, path, err := global.GetMountedPartition(IdPartitionGlobal)
	if err != nil {
		return err
	}

	//fmt.Println("Path: ", path)
	//fmt.Println("Name: ", mkuser.User)

	//Deserializar el superbloque
	sb := &structures.SuperBlock{}

	errorD := sb.Deserialize(path, int64(mountedPartition.Part_start))
	if errorD != nil {
		return errorD
	}

	userID := userCount
	userCount++

	userIDStr := strconv.Itoa(userID)

	fmt.Println("User ID: ", userIDStr)

	fullString, err := sb.PrintUsersFileContent(path)
	if err != nil {
		return err
	}

	newContent := fullString + userIDStr + "," + "U" + "," + strings.TrimSpace(mkuser.Grp) + "," + strings.TrimSpace(mkuser.User) + "," + strings.TrimSpace(mkuser.Pass) + "\n"

	newContent = strings.ReplaceAll(newContent, "\x00", "")

	//Verificar que el 3er parametro sea un grupo existente
	//imprimir grupos existentes
	//fmt.Println("Grupos existentes: ", structures.Groups)

	//Verificar que el grupo exista
	if global.GroupsUser[IdPartitionGlobal][mkuser.Grp] == nil {
		return errors.New("el grupo no existe")
	}

	for _, value := range global.GroupsUser[IdPartitionGlobal] {
		for _, user := range value {
			if user == mkuser.User {
				return errors.New("el usuario ya existe")
			}
		}
	}

	global.GroupsUser[IdPartitionGlobal][mkuser.Grp] = append(global.GroupsUser[IdPartitionGlobal][mkuser.Grp], mkuser.User)

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
