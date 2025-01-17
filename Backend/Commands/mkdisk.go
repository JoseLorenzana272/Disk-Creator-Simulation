package Commands

import (
	structures "archivos_pro1/Structures"
	utils "archivos_pro1/utils"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type MKDISK struct {
	size int
	unit string
	fit  string
	path string
}

func ParserMkdisk(tokens []string) (string, error) {
	cmd := &MKDISK{} // Crea una nueva instancia de MKDISK

	// Unir tokens en una sola cadena y luego dividir por espacios, respetando las comillas
	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar los parámetros del comando mkdisk
	re := regexp.MustCompile(`-size=\d+|-unit=[kKmM]|-fit=[bBfFwW]{2}|-path="[^"]+"|-path=[^\s]+`)
	// Encuentra todas las coincidencias de la expresión regular en la cadena de argumentos
	matches := re.FindAllString(args, -1)

	// Itera sobre cada coincidencia encontrada
	for _, match := range matches {
		// Divide cada parte en clave y valor usando "=" como delimitador
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("format of parameter is invalid: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		// Remove quotes from value if present
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Switch para manejar diferentes parámetros
		switch key {
		case "-size":
			// Convierte el valor del tamaño a un entero
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return "", errors.New("the size must be a positive integer")
			}
			cmd.size = size
		case "-unit":
			// Verifica que la unidad sea "K" o "M"
			if value != "K" && value != "M" {
				return "", errors.New("the unit must be K or M")
			}
			cmd.unit = strings.ToUpper(value)
		case "-fit":
			// Verifica que el ajuste sea "BF", "FF" o "WF"
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return "", errors.New("the fit must be BF, FF or WF")
			}
			cmd.fit = value
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

	// Verifica que los parámetros -size y -path hayan sido proporcionados
	if cmd.size == 0 {
		return "", errors.New("not enough parameters: -size")
	}
	if cmd.path == "" {
		return "", errors.New("not enough parameters: -path")
	}

	// Si no se proporcionó la unidad, se establece por defecto a "M"
	if cmd.unit == "" {
		cmd.unit = "M"
	}

	// Si no se proporcionó el ajuste, se establece por defecto a "FF"
	if cmd.fit == "" {
		cmd.fit = "FF"
	}

	// Crear el disco con los parámetros proporcionados
	err := commandMkdisk(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return "MKDISK: Disk created succesfully.", nil // Devuelve el comando MKDISK creado
}

func commandMkdisk(mkdisk *MKDISK) error {
	// Convertir el tamaño a bytes
	sizeBytes, err := utils.ConvertToBytes(mkdisk.size, mkdisk.unit)
	if err != nil {
		fmt.Println("Error converting size:", err)
		return err
	}

	// Crear el disco con el tamaño proporcionado
	err = createDisk(mkdisk, sizeBytes)
	if err != nil {
		fmt.Println("Error creating disk:", err)
		return err
	}

	// Crear el MBR con el tamaño proporcionado
	err = createMBR(mkdisk, sizeBytes)
	if err != nil {
		fmt.Println("Error creating MBR:", err)
		return err
	}

	return nil
}

func createDisk(mkdisk *MKDISK, sizeBytes int) error {
	// Crear las carpetas necesarias
	err := os.MkdirAll(filepath.Dir(mkdisk.path), os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directories:", err)
		return err
	}

	// Crear el archivo binario
	file, err := os.Create(mkdisk.path)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	// Escribir en el archivo usando un buffer de 1 MB
	buffer := make([]byte, 1024*1024) // Crea un buffer de 1 MB
	for sizeBytes > 0 {
		writeSize := len(buffer)
		if sizeBytes < writeSize {
			writeSize = sizeBytes // Ajusta el tamaño de escritura si es menor que el buffer
		}
		if _, err := file.Write(buffer[:writeSize]); err != nil {
			return err // Devuelve un error si la escritura falla
		}
		sizeBytes -= writeSize // Resta el tamaño escrito del tamaño total
	}
	return nil
}

func createMBR(mkdisk *MKDISK, sizeBytes int) error {
	// Crear el MBR con los valores proporcionados
	mbr := &structures.MBR{
		Mbr_size:           int32(sizeBytes),
		Mbr_creation_date:  float32(time.Now().Unix()),
		Mbr_disk_signature: rand.Int31(),
		Mbr_disk_fit:       [1]byte{mkdisk.fit[0]},
		Mbr_partitions: [4]structures.Partition{
			{Part_status: [1]byte{'9'}, Part_type: [1]byte{'0'}, Part_fit: [1]byte{'0'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'0'}, Part_correlative: -1, Part_id: [4]byte{'0'}},
			{Part_status: [1]byte{'9'}, Part_type: [1]byte{'0'}, Part_fit: [1]byte{'0'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'0'}, Part_correlative: -1, Part_id: [4]byte{'0'}},
			{Part_status: [1]byte{'9'}, Part_type: [1]byte{'0'}, Part_fit: [1]byte{'0'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'0'}, Part_correlative: -1, Part_id: [4]byte{'0'}},
			{Part_status: [1]byte{'9'}, Part_type: [1]byte{'0'}, Part_fit: [1]byte{'0'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'0'}, Part_correlative: -1, Part_id: [4]byte{'0'}},
		},
	}

	// Serializar el MBR en el archivo
	err := mbr.Serialize(mkdisk.path)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return nil
}
