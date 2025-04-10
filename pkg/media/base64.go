package media

import (
	"encoding/base64"
	"io"
)

// EncodeBase64 returns the base64 encoding of bin.
func EncodeBase64(bin []byte) string {
	return base64.StdEncoding.EncodeToString(bin)
}

// ReadBase64 returns a new reader that decodes base64 and returns binary data.
func ReadBase64(stream io.Reader) io.Reader {
	return base64.NewDecoder(base64.StdEncoding, stream)
}

// DecodeBase64 returns the bytes represented by the base64 string s.
// If the input is malformed, it returns the partially decoded data and
// [CorruptInputError]. Newline characters (\r and \n) are ignored.
func DecodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
