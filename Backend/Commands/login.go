package Commands

import (
	structures "archivos_pro1/Structures"
	"archivos_pro1/global"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var IsLogged bool
var IdPartitionGlobal string

type LOGIN struct {
	User string
	Pass string
	Id   string
}

func ParserLogin(tokens []string) (string, error) {
	cmd := &LOGIN{}

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`-user="[^"]+"|-user=[^\s]+|-pass="[^"]+"|-pass=[^\s]+ |-id="[^"]+"|-id=[^\s]+`)
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
		case "-id":
			if value == "" {
				return "", errors.New("the id cannot be empty")
			}
			cmd.Id = value
		default:
			return "", fmt.Errorf("unknown parameter: %s", key)
		}
	}

	if cmd.User == "" {
		return "", errors.New("there is a missing required parameter: -user")
	}

	if cmd.Pass == "" {
		return "", errors.New("there is a missing required parameter: -pass")
	}

	if cmd.Id == "" {
		return "", errors.New("there is a missing required parameter: -id")
	}

	err := commandLogin(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return "LOGIN: You are logged in with user " + cmd.User, nil
}

func commandLogin(login *LOGIN) error {
	if IsLogged {
		return errors.New("there's already a user logged in")
	}
	//Traer el path del disco
	idPartition, path, err := global.GetMountedPartition(login.Id)
	IdPartitionGlobal = login.Id
	if err != nil {
		return err
	}

	startOfPartition := idPartition.Part_start

	//Se extrae el texto
	sb := structures.SuperBlock{}
	sb.Deserialize(path, int64(startOfPartition))
	//Obtener contenido del archivo
	content, err := sb.PrintUsersFileContent(path)
	if err != nil {
		return err
	}

	fmt.Println("CONTENT: ", content)

	//Comparar si el usuario y contraseña existen
	lines := strings.Split(content, "\n")

	// Iterar sobre cada línea para buscar al usuario
	for _, line := range lines {
		// Separar la línea por comas
		parts := strings.Split(line, ",")
		if len(parts) < 4 {
			// Si la línea no tiene al menos 4 partes, no es una línea válida para usuario
			continue
		}

		// Verificar si es una línea de usuario ('U')
		if parts[1] == "U" {
			user := parts[3]
			password := parts[4]
			fmt.Println("User:", user)
			fmt.Println("Password:", password)

			// Comparar usuario y contraseña
			if user == strings.TrimSpace(login.User) && password == strings.TrimSpace(login.Pass) {
				fmt.Println("Login exitoso!")
				IsLogged = true
				return nil
			}
			fmt.Println("Usuario premium: ", login.User)
			fmt.Println("Contraseña premium: ", login.Pass)
		}
	}

	// Si no se encontró al usuario, devolver un error
	return fmt.Errorf("usuario o contraseña incorrectos")

}
