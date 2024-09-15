package reports

import (
	structures "archivos_pro1/Structures"
	"archivos_pro1/utils"
	"fmt"
	"os"
	"os/exec"
	"time"
)

/*func (sb *SuperBlock) GenerateSBDotFile() {
	// Convertir el tiempo de montaje a una fecha
	mountTime := time.Unix(int64(sb.S_mtime), 0)
	// Convertir el tiempo de desmontaje a una fecha
	unmountTime := time.Unix(int64(sb.S_umtime), 0)

	// Usamos strings.Builder para construir el archivo .dot
	var dotContent strings.Builder
	dotContent.WriteString("digraph SuperBlock {\n")
	dotContent.WriteString("  node [shape=none, fontname=\"Helvetica,Arial,sans-serif\"];\n")
	dotContent.WriteString("  rankdir=TB;\n\n")

	// Crear el nodo del SuperBlock
	dotContent.WriteString("  sb [label=<\n")
	dotContent.WriteString("    <table border='0' cellborder='1' cellspacing='0' cellpadding='10' style='rounded' bgcolor='#F5F5F5'>\n")
	dotContent.WriteString("      <tr>\n")
	dotContent.WriteString("        <td colspan='2' bgcolor='#333333'><font color='white'>SuperBlock Report</font></td>\n")
	dotContent.WriteString("      </tr>\n")

	// Agregar los atributos del SuperBlock
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Filesystem Type:</td><td>%d</td></tr>\n", sb.S_filesystem_type))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Inodes Count:</td><td>%d</td></tr>\n", sb.S_inodes_count))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Blocks Count:</td><td>%d</td></tr>\n", sb.S_blocks_count))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Free Inodes Count:</td><td>%d</td></tr>\n", sb.S_free_inodes_count))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Free Blocks Count:</td><td>%d</td></tr>\n", sb.S_free_blocks_count))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Mount Time:</td><td>%s</td></tr>\n", mountTime.Format(time.RFC3339)))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Unmount Time:</td><td>%s</td></tr>\n", unmountTime.Format(time.RFC3339)))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Mount Count:</td><td>%d</td></tr>\n", sb.S_mnt_count))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Magic:</td><td>%d</td></tr>\n", sb.S_magic))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Inode Size:</td><td>%d</td></tr>\n", sb.S_inode_size))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Block Size:</td><td>%d</td></tr>\n", sb.S_block_size))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>First Inode:</td><td>%d</td></tr>\n", sb.S_first_ino))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>First Block:</td><td>%d</td></tr>\n", sb.S_first_blo))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Bitmap Inode Start:</td><td>%d</td></tr>\n", sb.S_bm_inode_start))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Bitmap Block Start:</td><td>%d</td></tr>\n", sb.S_bm_block_start))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Inode Start:</td><td>%d</td></tr>\n", sb.S_inode_start))
	dotContent.WriteString(fmt.Sprintf("      <tr><td>Block Start:</td><td>%d</td></tr>\n", sb.S_block_start))

	// Cerrar tabla y nodo
	dotContent.WriteString("    </table>\n")
	dotContent.WriteString("  >];\n")
	dotContent.WriteString("}\n")

	// Crear el archivo .dot
	file, err := os.Create("SuperBlock.dot")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Escribir el contenido en el archivo
	_, err = file.WriteString(dotContent.String())
	if err != nil {
		fmt.Println("Error escribiendo en el archivo .dot:", err)
		return
	}

	fmt.Println("SuperBlock.dot file created")
}*/

func ReportSuperblock(sb *structures.SuperBlock, path string) error {
	// Crear las carpetas padre si no existen
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	dotFileName, outputImage := utils.GetFileNames(path)

	// Definir el contenido DOT con una tabla estilizada
	dotContent := fmt.Sprintf(`digraph G {
		node [shape=plaintext, fontname="Helvetica, Arial, sans-serif"]
		tabla [label=<
			<table border="0" cellborder="1" cellspacing="0" cellpadding="10" bgcolor="#f7f7f7" style="rounded">
				<tr><td colspan="2" bgcolor="#4CAF50" align="center" cellpadding="4" cellspacing="0"><b><font color="white">SUPERBLOCK REPORT</font></b></td></tr>
				<tr bgcolor="#eeeeee"><td><b>Filesystem Type</b></td><td>%d</td></tr>
				<tr><td><b>Inodes Count</b></td><td>%d</td></tr>
				<tr bgcolor="#eeeeee"><td><b>Blocks Count</b></td><td>%d</td></tr>
				<tr><td><b>Free Inodes Count</b></td><td>%d</td></tr>
				<tr bgcolor="#eeeeee"><td><b>Free Blocks Count</b></td><td>%d</td></tr>
				<tr><td><b>Mount Time</b></td><td>%s</td></tr>
				<tr bgcolor="#eeeeee"><td><b>Unmount Time</b></td><td>%s</td></tr>
				<tr><td><b>Mount Count</b></td><td>%d</td></tr>
				<tr bgcolor="#eeeeee"><td><b>Magic</b></td><td>%d</td></tr>
				<tr><td><b>Inode Size</b></td><td>%d</td></tr>
				<tr bgcolor="#eeeeee"><td><b>Block Size</b></td><td>%d</td></tr>
				<tr><td><b>First Inode</b></td><td>%d</td></tr>
				<tr bgcolor="#eeeeee"><td><b>First Block</b></td><td>%d</td></tr>
				<tr><td><b>Bitmap Inode Start</b></td><td>%d</td></tr>
				<tr bgcolor="#eeeeee"><td><b>Bitmap Block Start</b></td><td>%d</td></tr>
				<tr><td><b>Inode Start</b></td><td>%d</td></tr>
				<tr bgcolor="#eeeeee"><td><b>Block Start</b></td><td>%d</td></tr>
			</table>
		> ]}
	`, sb.S_filesystem_type, sb.S_inodes_count, sb.S_blocks_count, sb.S_free_inodes_count, sb.S_free_blocks_count, time.Unix(int64(sb.S_mtime), 0).Format("2006-01-02 15:04:05"), time.Unix(int64(sb.S_umtime), 0).Format("2006-01-02 15:04:05"), sb.S_mnt_count, sb.S_magic, sb.S_inode_size, sb.S_block_size, sb.S_first_ino, sb.S_first_blo, sb.S_bm_inode_start, sb.S_bm_block_start, sb.S_inode_start, sb.S_block_start)

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

	//Ejecutar el comando dot para generar la imagen
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar el comando Graphviz: %v", err)
	}

	fmt.Println("SuperBlock report created successfully")
	return nil

}
