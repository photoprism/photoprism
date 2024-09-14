package pluscode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	t.Run("germany", func(t *testing.T) {
		plusCode := Encode(48.56344833333333, 8.996878333333333)

		expected := "8FWCHX7W+"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("lat_overflow", func(t *testing.T) {
		plusCode := Encode(548.56344833333333, 8.996878333333333)

		assert.Equal(t, "", plusCode)
	})

	t.Run("lng_overflow", func(t *testing.T) {
		plusCode := Encode(48.56344833333333, 258.996878333333333)

		assert.Equal(t, "", plusCode)
	})
}

func TestEncodeLength(t *testing.T) {
	t.Run("germany_9", func(t *testing.T) {
		plusCode, err := EncodeLength(48.56344833333333, 8.996878333333333, 9)
		if err != nil {
			t.Fatal(err)
		}

		expected := "8FWCHX7W+9Q"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("germany_8", func(t *testing.T) {
		plusCode, err := EncodeLength(48.56344833333333, 8.996878333333333, 8)
		if err != nil {
			t.Fatal(err)
		}

		expected := "8FWCHX7W+"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("germany_7", func(t *testing.T) {
		plusCode, err := EncodeLength(48.56344833333333, 8.996878333333333, 7)
		if err != nil {
			t.Fatal(err)
		}

		expected := "8FWCHX7W+"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("germany_6", func(t *testing.T) {
		plusCode, err := EncodeLength(48.56344833333333, 8.996878333333333, 6)
		if err != nil {
			t.Fatal(err)
		}

		expected := "8FWCHX00+"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("lat_overflow", func(t *testing.T) {
		plusCode, err := EncodeLength(548.56344833333333, 8.996878333333333, 7)
		if err == nil {
			t.Fatal("encode should return error")
		}
		assert.Equal(t, "", plusCode)
	})

	t.Run("lng_overflow", func(t *testing.T) {
		plusCode, err := EncodeLength(48.56344833333333, 258.996878333333333, 7)
		if err == nil {
			t.Fatal("encode should return error")
		}
		assert.Equal(t, "", plusCode)
	})
}

func TestS2(t *testing.T) {
	t.Run("germany", func(t *testing.T) {
		token := S2("8FWCHX7W+")

		assert.Equal(t, "4799e3772d14", token)
	})
	t.Run("empty code", func(t *testing.T) {
		token := S2("")

		assert.Equal(t, "", token)
	})
	t.Run("invalid code", func(t *testing.T) {
		token := S2("xxx")

		assert.Equal(t, "", token)
	})
}
