package Commands

import (
	"bufio"
	"errors" // Paquete para manejar errores y crear nuevos errores con mensajes personalizados
	"fmt"    // Paquete para formatear cadenas y realizar operaciones de entrada/salida
	"os"
	"regexp" // Paquete para trabajar con expresiones regulares, útil para encontrar y manipular patrones en cadenas

	// Paquete para convertir cadenas a otros tipos de datos, como enteros
	"strings" // Paquete para manipular cadenas, como unir, dividir, y modificar contenido de cadenas
)

// RMDISK estructura que representa el comando rmdisk con su parámetro
type RMDISK struct {
	path string // Ruta del archivo del disco
}

/*
	rmdisk -path=/home/user/Disco1.mia
	rmdisk -path="/home/mis discos/Disco4.mia"
*/

// CommandRmdisk parsea el comando rmdisk y devuelve una instancia de RMDISK
func ParserRmdisk(tokens []string) (string, error) {
	cmd := &RMDISK{} // Crea una nueva instancia de RMDISK

	// Unir tokens en una sola cadena y luego dividir por espacios, respetando las comillas
	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar el parámetro del comando rmdisk
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+`)
	// Encuentra todas las coincidencias de la expresión regular en la cadena de argumentos
	matches := re.FindAllString(args, -1)

	// Itera sobre cada coincidencia encontrada
	for _, match := range matches {
		// Divide cada parte en clave y valor usando "=" como delimitador
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		// Remove quotes from value if present
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Switch para manejar el parámetro -path
		switch key {
		case "-path":
			// Verifica que el path no esté vacío
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que el parámetro -path haya sido proporcionado
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	// Ejecuta el comando rmdisk
	err := commandRmdisk(cmd)
	if err != nil {
		return "", err
	}

	return "RMDISK: Disk removed successfully", nil
}

func commandRmdisk(rmdisk *RMDISK) error {
	// Verifica si el archivo existe
	if _, err := os.Stat(rmdisk.path); os.IsNotExist(err) {
		return errors.New("the file does not exist")
	}

	// Solicita confirmación al usuario
	fmt.Printf("Are you sure you want to delete the file '%s'? (yes/no): ", rmdisk.path)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	confirmation := strings.TrimSpace(scanner.Text())

	if strings.ToLower(confirmation) != "yes" {
		return errors.New("deletion canceled by user")
	}

	// Elimina el archivo del disco
	err := os.Remove(rmdisk.path)
	if err != nil {
		return err
	}

	return nil
}
