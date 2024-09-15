package reports

import (
	structures "archivos_pro1/Structures"
	"archivos_pro1/utils"
	"fmt"
	"os"
	"strings"
)

// ReportBMBlock genera un reporte del bitmap de bloques y lo guarda en la ruta especificada
func ReportBMBlock(superblock *structures.SuperBlock, diskPath string, path string) error {
	// Crear las carpetas padre si no existen
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	// Abrir el archivo de disco
	file, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo de disco: %v", err)
	}
	defer file.Close()

	// Calcular el número total de bloques
	totalBlocks := superblock.S_blocks_count + superblock.S_free_blocks_count

	var bitmapContent strings.Builder

	for i := int32(0); i < totalBlocks; i++ {
		offset := int32(superblock.S_bm_block_start) + (i / 8)
		_, err := file.Seek(int64(offset), 0)
		if err != nil {
			return fmt.Errorf("error al establecer el puntero en el archivo: %v", err)
		}

		// Leer un byte
		char := make([]byte, 1)
		_, err = file.Read(char)
		if err != nil {
			return fmt.Errorf("error al leer el byte del archivo: %v", err)
		}

		bit := (char[0] >> (7 - (i % 8))) & 1

		if bit == 0 {
			bitmapContent.WriteByte('0')
		} else {
			bitmapContent.WriteByte('1')
		}

		if (i+1)%20 == 0 {
			bitmapContent.WriteString("\n")
		}
	}

	// Crear el archivo TXT
	txtFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error al crear el archivo TXT: %v", err)
	}
	defer txtFile.Close()

	// Escribir el contenido del bitmap en el archivo TXT
	_, err = txtFile.WriteString(bitmapContent.String())
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo TXT: %v", err)
	}

	fmt.Println("Archivo del bitmap de bloques generado:", path)
	return nil
}
