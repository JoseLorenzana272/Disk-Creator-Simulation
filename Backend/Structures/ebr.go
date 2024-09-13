package structures

type EBR struct {
	Part_mount [1]byte
	Part_fit   [2]byte
	Part_start int32
	Part_s     int32
	Part_next  int32
	Part_name  [16]byte
}
