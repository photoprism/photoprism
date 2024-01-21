package checksum

const (
	CharsetBase36 = "abcdefghijklmnopqrstuvwxyz0123456789"
	CharsetBase62 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// Char returns a simple checksum byte based on Crc32 and the 62 characters specified in CharsetBase62.
func Char(data []byte) byte {
	return CharsetBase62[Crc32(data)%62]
}

// Base36 returns a simple checksum byte based on Crc32 and the 36 lower-case characters specified in CharsetBase36.
func Base36(data []byte) byte {
	return CharsetBase36[Crc32(data)%36]
}
