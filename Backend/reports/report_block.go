package reports

import (
	structures "archivos_pro1/Structures"
	"archivos_pro1/utils"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ReportBlock genera un reporte visual de los bloques (pointer, folder, file) y lo guarda en la ruta especificada
func ReportBlock(superblock *structures.SuperBlock, diskPath string, path string) error {
	// Crear directorios padres si no existen
	if err := utils.CreateParentDirs(path); err != nil {
		return fmt.Errorf("error creando directorios padre: %v", err)
	}

	// Obtener nombres de archivo base y de salida
	dotFileName, outputImage := utils.GetFileNames(path)

	// Iniciar el contenido DOT con una estructura profesional
	dotContent := `digraph BlockReport {
		rankdir=TB;
		node [shape=none, fontname="Helvetica, Arial, sans-serif"];
		graph [splines=true, nodesep=0.5, ranksep=0.4];
		edge [color=black, arrowhead=normal];
	`

	// Iterar sobre los bloques
	for i := int32(0); i < superblock.S_blocks_count; i++ {
		blockType := determineBlockType(i) // Función para determinar el tipo de bloque (Pointer, Folder, File)
		blockStart := int64(superblock.S_block_start + (i * superblock.S_block_size))

		switch blockType {
		case "PointerBlock":
			pointerBlock := &structures.PointerBlock{}
			if err := pointerBlock.Deserialize(diskPath, blockStart); err != nil {
				return fmt.Errorf("error deserializando PointerBlock %d: %v", i, err)
			}
			dotContent += formatPointerBlock(i, pointerBlock)

		case "FolderBlock":
			folderBlock := &structures.FolderBlock{}
			if err := folderBlock.Deserialize(diskPath, blockStart); err != nil {
				return fmt.Errorf("error deserializando FolderBlock %d: %v", i, err)
			}
			dotContent += formatFolderBlock(i, folderBlock)

		case "FileBlock":
			fileBlock := &structures.FileBlock{}
			if err := fileBlock.Deserialize(diskPath, blockStart); err != nil {
				return fmt.Errorf("error deserializando FileBlock %d: %v", i, err)
			}
			dotContent += formatFileBlock(i, fileBlock)
		}
	}

	// Finalizar el contenido DOT
	dotContent += "}"

	// Guardar el contenido DOT en un archivo
	if file, err := os.Create(dotFileName); err == nil {
		defer file.Close()
		if _, err := file.WriteString(dotContent); err != nil {
			return fmt.Errorf("error escribiendo contenido DOT: %v", err)
		}
	} else {
		return fmt.Errorf("error creando archivo DOT: %v", err)
	}

	// Ejecutar Graphviz para generar la imagen
	if err := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage).Run(); err != nil {
		return fmt.Errorf("error ejecutando Graphviz: %v", err)
	}

	fmt.Printf("Reporte de bloques generado: %s\n", outputImage)
	return nil
}

// determineBlockType retorna el tipo de bloque (Pointer, Folder, File) basándose en su índice o metadata
func determineBlockType(index int32) string {
	if index%3 == 0 {
		return "PointerBlock"
	} else if index%3 == 1 {
		return "FolderBlock"
	} else {
		return "FileBlock"
	}
}

// formatPointerBlock genera el contenido DOT para un PointerBlock
func formatPointerBlock(index int32, block *structures.PointerBlock) string {
	content := fmt.Sprintf(`pointerBlock%d [label=<
		<table border="0" cellborder="1" cellspacing="0" cellpadding="4">
			<tr><td colspan="2" bgcolor="#CCCCCC"><b>POINTER BLOCK %d</b></td></tr>
	`, index, index)

	for i, pointer := range block.P_pointers {
		content += fmt.Sprintf(`<tr><td>Pointer %d</td><td>%d</td></tr>`, i+1, pointer)
	}

	content += "</table>>];\n"
	return content
}

// formatFolderBlock genera el contenido DOT para un FolderBlock
func formatFolderBlock(index int32, block *structures.FolderBlock) string {
	content := fmt.Sprintf(`folderBlock%d [label=<
		<table border="0" cellborder="1" cellspacing="0" cellpadding="4">
			<tr><td colspan="2" bgcolor="#CCCCCC"><b>FOLDER BLOCK %d</b></td></tr>
	`, index, index)

	for i, contentItem := range block.B_content {
		name := string(contentItem.B_name[:])
		content += fmt.Sprintf(`<tr><td>Name %d</td><td>%s</td></tr>`, i+1, name)
		content += fmt.Sprintf(`<tr><td>Inode</td><td>%d</td></tr>`, contentItem.B_inodo)
	}

	content += "</table>>];\n"
	return content
}

// formatFileBlock genera el contenido DOT para un FileBlock
func formatFileBlock(index int32, block *structures.FileBlock) string {
	// Convertir el contenido del bloque a string y eliminar los caracteres nulos
	fileContent := string(block.B_content[:])
	fileContent = cleanString(fileContent)

	content := fmt.Sprintf(`fileBlock%d [label=<
		<table border="0" cellborder="1" cellspacing="0" cellpadding="4">
			<tr><td colspan="2" bgcolor="#CCCCCC"><b>FILE BLOCK %d</b></td></tr>
	`, index, index)

	content += fmt.Sprintf(`<tr><td colspan="2">%s</td></tr>`, fileContent)

	content += "</table>>];\n"
	return content
}

// cleanString elimina caracteres nulos y otros caracteres no deseados
func cleanString(input string) string {
	// Remover caracteres nulos (\x00)
	return strings.ReplaceAll(input, "\x00", "")
}
