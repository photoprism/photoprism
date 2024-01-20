package checksum

// Digit returns a Crc32-based checksum number between 0 and 9.
func Digit(data []byte) int {
	return int(Crc32(data) % 10)
}
