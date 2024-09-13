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

func ParserLogin(tokens []string) (*LOGIN, error) {
	cmd := &LOGIN{}

	args := strings.Join(tokens, " ")

	re := regexp.MustCompile(`-user="[^"]+"|-user=[^\s]+|-pass="[^"]+"|-pass=[^\s]+ |-id="[^"]+"|-id=[^\s]+`)
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
		case "-pass":
			if value == "" {
				return nil, errors.New("la contraseña no puede estar vacía")
			}
			cmd.Pass = value
		case "-id":
			if value == "" {
				return nil, errors.New("el id no puede estar vacío")
			}
			cmd.Id = value
		default:
			return nil, fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.User == "" {
		return nil, errors.New("faltan parámetros requeridos: -user")
	}

	if cmd.Pass == "" {
		return nil, errors.New("faltan parámetros requeridos: -pass")
	}

	if cmd.Id == "" {
		return nil, errors.New("faltan parámetros requeridos: -id")
	}

	err := commandLogin(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return cmd, nil
}

func commandLogin(login *LOGIN) error {
	if IsLogged {
		return errors.New("ya hay una sesión iniciada")
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
