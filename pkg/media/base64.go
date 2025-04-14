package media

import (
	"encoding/base64"
	"io"
)

// EncodeBase64String returns the base64 encoding of bin.
func EncodeBase64String(bin []byte) string {
	return base64.StdEncoding.EncodeToString(bin)
}

// EncodedLenBase64 returns the length in bytes of the base64 encoding of an input buffer of length n.
func EncodedLenBase64(decodedBytes int) int {
	return base64.StdEncoding.EncodedLen(decodedBytes)
}

// DecodeBase64String returns the bytes represented by the base64 string s.
// If the input is malformed, it returns the partially decoded data and
// [CorruptInputError]. Newline characters (\r and \n) are ignored.
func DecodeBase64String(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// ReadBase64 returns a new reader that decodes base64 and returns binary data.
func ReadBase64(stream io.Reader) io.Reader {
	return base64.NewDecoder(base64.StdEncoding, stream)
}

// EncodeBase64Bytes encodes src, writing EncodedLenBase64 bytes to dst.
//
// The encoding pads the output to a multiple of 4 bytes,
// so Encode is not appropriate for use on individual blocks
// of a large data stream.
func EncodeBase64Bytes(dst, src []byte) {
	base64.StdEncoding.Encode(dst, src)
}

// DecodedLenBase64 returns the maximum length in bytes of the decoded data
// corresponding to n bytes of base64-encoded data.
func DecodedLenBase64(encodedBytes int) int {
	return base64.StdEncoding.DecodedLen(encodedBytes)
}

// DecodeBase64Bytes decodes src, writing at most DecodedLenBase64 bytes to dst.
// If src contains invalid base64 data, it returns the number of bytes successfully
// written. New line characters (\r and \n) are ignored.
func DecodeBase64Bytes(dst, src []byte) (n int, err error) {
	return base64.StdEncoding.Decode(dst, src)
}
