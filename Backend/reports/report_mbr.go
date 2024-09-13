package reports

import (
	structures "archivos_pro1/Structures"
	"archivos_pro1/utils"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func ReportMBR(mbr *structures.MBR, path string) error {
	// Crear las carpetas padre si no existen
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	// Obtener el nombre base del archivo sin la extensión
	dotFileName, outputImage := utils.GetFileNames(path)

	// Definir el contenido DOT con una tabla estilizada
	dotContent := fmt.Sprintf(`digraph G {
        node [shape=plaintext, fontname="Helvetica, Arial, sans-serif"]
        tabla [label=<
            <table border="0" cellborder="1" cellspacing="0" cellpadding="10" bgcolor="#f2f2f2">
                <tr><td colspan="2" bgcolor="#4CAF50" align="center"><b>REPORTE MBR</b></td></tr>
                <tr bgcolor="#dddddd"><td><b>mbr_tamano</b></td><td>%d</td></tr>
                <tr><td><b>mbr_fecha_creacion</b></td><td>%s</td></tr>
                <tr bgcolor="#dddddd"><td><b>mbr_disk_signature</b></td><td>%d</td></tr>
            `, mbr.Mbr_size, time.Unix(int64(mbr.Mbr_creation_date), 0).Format("2006-01-02 15:04:05"), mbr.Mbr_disk_signature)

	// Agregar las particiones a la tabla
	for i, part := range mbr.Mbr_partitions {
		// Convertir Part_name a string y eliminar los caracteres nulos
		partName := strings.TrimRight(string(part.Part_name[:]), "\x00")
		// Convertir Part_status, Part_type y Part_fit a char
		partStatus := rune(part.Part_status[0])
		partType := rune(part.Part_type[0])
		partFit := rune(part.Part_fit[0])

		// Definir un color de fondo alternado para las particiones
		bgColor := "#ffffff"
		if i%2 == 0 {
			bgColor = "#eeeeee"
		}

		// Agregar la partición a la tabla
		dotContent += fmt.Sprintf(`
				<tr><td colspan="2" bgcolor="#4CAF50" align="center"><b>PARTICIÓN %d</b></td></tr>
				<tr bgcolor="%s"><td><b>part_status</b></td><td>%c</td></tr>
				<tr bgcolor="#ffffff"><td><b>part_type</b></td><td>%c</td></tr>
				<tr bgcolor="%s"><td><b>part_fit</b></td><td>%c</td></tr>
				<tr bgcolor="#ffffff"><td><b>part_start</b></td><td>%d</td></tr>
				<tr bgcolor="%s"><td><b>part_size</b></td><td>%d</td></tr>
				<tr bgcolor="#ffffff"><td><b>part_name</b></td><td>%s</td></tr>
			`, i+1, bgColor, partStatus, partType, bgColor, partFit, part.Part_start, bgColor, part.Part_size, partName)
	}

	// Cerrar la tabla y el contenido DOT
	dotContent += "</table>>] }"

	// Guardar el contenido DOT en un archivo
	file, err := os.Create(dotFileName)
	if err != nil {
		return fmt.Errorf("error al crear el archivo: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(dotContent)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo: %v", err)
	}

	// Ejecutar el comando Graphviz para generar la imagen
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar el comando Graphviz: %v", err)
	}

	fmt.Println("Imagen de la tabla generada:", outputImage)
	return nil
}
