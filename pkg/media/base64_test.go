package media

import (
	"io"
	"strings"
	"testing"

	"github.com/gabriel-vasile/mimetype"
	"github.com/stretchr/testify/assert"
)

func TestBase64(t *testing.T) {
	t.Run("DecodeString", func(t *testing.T) {
		data, err := DecodeBase64String(gopher)
		assert.NoError(t, err)

		if mime := mimetype.Detect(data); mime == nil {
			t.Fatal("mimetype image/png expected")
		} else {
			assert.Equal(t, "image/png", mime.String())
		}
	})
	t.Run("DecodeString", func(t *testing.T) {
		data, err := DecodeBase64String(gopher)
		assert.NoError(t, err)
		assert.Equal(t, gopher, EncodeBase64String(data))
	})
	t.Run("Read", func(t *testing.T) {
		reader := ReadBase64(strings.NewReader(gopher))

		if data, err := io.ReadAll(reader); err != nil {
			t.Fatal(err)
		} else if decodeData, decodeErr := DecodeBase64String(gopher); decodeErr != nil {
			t.Fatal(decodeErr)
		} else {
			assert.Equal(t, data, decodeData)
			assert.Equal(t, EncodeBase64String(data), gopher)
		}
	})
	t.Run("DecodeBytes", func(t *testing.T) {
		encoded := []byte(gopher)
		encodedLen := len(encoded)
		decodedLen := DecodedLenBase64(encodedLen)
		binary := make([]byte, decodedLen)

		if n, err := DecodeBase64Bytes(binary, encoded); err != nil {
			t.Fatal(err)
		} else {
			assert.GreaterOrEqual(t, decodedLen, n)
		}
	})
	t.Run("EncodeBytes", func(t *testing.T) {
		encoded := []byte(gopher)
		encodedLen := len(encoded)
		decodedLen := DecodedLenBase64(encodedLen)
		binary := make([]byte, decodedLen)

		if n, err := DecodeBase64Bytes(binary, encoded); err != nil {
			t.Fatal(err)
		} else {
			binary = binary[:n]
			assert.GreaterOrEqual(t, decodedLen, n)
		}

		binaryEncodedLen := EncodedLenBase64(len(binary))
		binaryEncoded := make([]byte, binaryEncodedLen)

		EncodeBase64Bytes(binaryEncoded, binary)
		assert.Equal(t, encoded, binaryEncoded)
		assert.Equal(t, gopher, string(binaryEncoded))

		data, err := DecodeBase64String(string(binaryEncoded))
		assert.NoError(t, err)
		assert.Equal(t, gopher, EncodeBase64String(data))
	})
}
