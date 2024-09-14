package Analyzer

import (
	commands "archivos_pro1/Commands"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Analyzer analiza el comando de entrada y ejecuta la acción correspondiente
func Analyzer(input string) (interface{}, error) {
	lines := strings.Split(input, "\n") // Divide la entrada en líneas
	var results []interface{}           // Guarda los resultados de cada línea

	for _, line := range lines {
		line = strings.TrimSpace(line) // Elimina espacios en blanco antes y después de cada línea
		if line == "" {
			continue // Ignora líneas vacías
		}

		tokens := strings.Fields(line)

		// Si no se proporcionó ningún comando, devuelve un error
		if len(tokens) == 0 {
			return nil, errors.New("no se proporcionó ningún comando")
		}

		// Switch para manejar diferentes comandos
		var result interface{}
		var err error

		switch tokens[0] {
		case "mkdisk":
			result, err = commands.ParserMkdisk(tokens[1:])
		case "rmdisk":
			result, err = commands.ParserRmdisk(tokens[1:])
		case "fdisk":
			result, err = commands.ParserFdisk(tokens[1:])
		case "mount":
			result, err = commands.ParserMount(tokens[1:])
		case "mkfs":
			result, err = commands.ParserMkfs(tokens[1:])
		case "rep":
			result, err = commands.ParserRep(tokens[1:])
		case "login":
			result, err = commands.ParserLogin(tokens[1:])
		case "mkgrp":
			result, err = commands.ParserMkgrp(tokens[1:])
		case "mkusr":
			result, err = commands.ParserMkuser(tokens[1:])
		case "logout":
			commands.IsLogged = false
			result = "Sesión cerrada"
		case "rmgrp":
			result, err = commands.ParserRmgrp(tokens[1:])
		case "rmusr":
			result, err = commands.ParserRmusr(tokens[1:])
		case "chgrp":
			result, err = commands.ParserChgrp(tokens[1:])
		case "mkdir":
			result, err = commands.ParserMkdir(tokens[1:])
		case "mkfile":
			result, err = commands.ParserMkfile(tokens[1:])
		case "clear":
			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			err = cmd.Run()
			if err != nil {
				err = errors.New("no se pudo limpiar la terminal")
			}
		default:
			err = fmt.Errorf("comando desconocido: %s", tokens[0])
		}

		if err != nil {
			return nil, err // Devuelve el error si hay uno
		}

		results = append(results, result) // Agrega el resultado al slice
	}

	return results, nil // Devuelve todos los resultados
}
