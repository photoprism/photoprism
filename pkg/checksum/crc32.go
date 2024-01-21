package checksum

import (
	"fmt"
	"hash/crc32"
)

var Crc32Castagnoli = crc32.MakeTable(crc32.Castagnoli)

// Crc32 returns the CRC-32 checksum of data using the crc32.IEEE polynomial.
func Crc32(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

// Serial returns the CRC-32 checksum as a hexadecimal encoded string using the Castagnoli polynomial,
// which has better error detection properties than crc32.IEEE.
func Serial(data []byte) string {
	return fmt.Sprintf("%08x", crc32.Checksum(data, Crc32Castagnoli))
}
