package reports

import (
	structures "archivos_pro1/Structures"
	"archivos_pro1/utils"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// ReportInode genera un reporte visual de los inodos y lo guarda en la ruta especificada
func ReportInode(superblock *structures.SuperBlock, diskPath string, path string) error {
	// Crear directorios padres si no existen
	if err := utils.CreateParentDirs(path); err != nil {
		return fmt.Errorf("error creando directorios padre: %v", err)
	}

	// Obtener nombres de archivo base y de salida
	dotFileName, outputImage := utils.GetFileNames(path)

	// Iniciar el contenido DOT con una estructura profesional
	dotContent := `digraph InodeReport {
		rankdir=TB;
		node [shape=none, fontname="Helvetica, Arial, sans-serif"];
		graph [splines=true, nodesep=0.5, ranksep=0.4];
		edge [color=black, arrowhead=normal];
	`

	// Iterar sobre los inodos en el superblock
	for i := int32(0); i < superblock.S_inodes_count; i++ {
		inode := &structures.Inode{}
		inodeStart := int64(superblock.S_inode_start + (i * superblock.S_inode_size))

		// Deserializar el inodo desde el disco
		if err := inode.Deserialize(diskPath, inodeStart); err != nil {
			return fmt.Errorf("error al deserializar inodo %d: %v", i, err)
		}

		// Formatear las fechas como cadenas
		atime := time.Unix(int64(inode.I_atime), 0).Format(time.RFC3339)
		ctime := time.Unix(int64(inode.I_ctime), 0).Format(time.RFC3339)
		mtime := time.Unix(int64(inode.I_mtime), 0).Format(time.RFC3339)

		// Generar el contenido DOT para el inodo actual
		dotContent += fmt.Sprintf(`inode%d [label=<
			<table border="0" cellborder="1" cellspacing="0" cellpadding="4">
				<tr><td colspan="2" bgcolor="#CCCCCC"><b>INODE REPORT %d</b></td></tr>
				<tr><td><b>UID</b></td><td>%d</td></tr>
				<tr><td><b>GID</b></td><td>%d</td></tr>
				<tr><td><b>Size</b></td><td>%d</td></tr>
				<tr><td><b>Access Time</b></td><td>%s</td></tr>
				<tr><td><b>Creation Time</b></td><td>%s</td></tr>
				<tr><td><b>Modification Time</b></td><td>%s</td></tr>
				<tr><td><b>Type</b></td><td>%c</td></tr>
				<tr><td><b>Permissions</b></td><td>%s</td></tr>
				<tr><td colspan="2" bgcolor="#E0E0E0"><b>Direct Blocks</b></td></tr>
		`, i, i, inode.I_uid, inode.I_gid, inode.I_size, atime, ctime, mtime, rune(inode.I_type[0]), string(inode.I_perm[:]))

		// Incluir los bloques directos
		for j, block := range inode.I_block[:12] {
			dotContent += fmt.Sprintf(`<tr><td>Block %d</td><td>%d</td></tr>`, j+1, block)
		}

		// Incluir los bloques indirectos
		dotContent += fmt.Sprintf(`
			<tr><td colspan="2" bgcolor="#E0E0E0"><b>Indirect Blocks</b></td></tr>
			<tr><td>Single Indirect</td><td>%d</td></tr>
			<tr><td>Double Indirect</td><td>%d</td></tr>
			<tr><td>Triple Indirect</td><td>%d</td></tr>
			</table>>];
		`, inode.I_block[12], inode.I_block[13], inode.I_block[14])

		// Conectar inodos secuenciales
		if i < superblock.S_inodes_count-1 {
			dotContent += fmt.Sprintf("inode%d -> inode%d;\n", i, i+1)
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

	fmt.Printf("Reporte de inodos generado: %s\n", outputImage)
	return nil
}
