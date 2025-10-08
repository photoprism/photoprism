package pluscode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	t.Run("Germany", func(t *testing.T) {
		plusCode := Encode(48.56344833333333, 8.996878333333333)

		expected := "8FWCHX7W+"

		assert.Equal(t, expected, plusCode)
	})
	t.Run("LatOverflow", func(t *testing.T) {
		plusCode := Encode(548.56344833333333, 8.996878333333333)

		assert.Equal(t, "", plusCode)
	})
	t.Run("LngOverflow", func(t *testing.T) {
		plusCode := Encode(48.56344833333333, 258.996878333333333)

		assert.Equal(t, "", plusCode)
	})
}

func TestEncodeLength(t *testing.T) {
	t.Run("GermanyNine", func(t *testing.T) {
		plusCode, err := EncodeLength(48.56344833333333, 8.996878333333333, 9)
		if err != nil {
			t.Fatal(err)
		}

		expected := "8FWCHX7W+9Q"

		assert.Equal(t, expected, plusCode)
	})
	t.Run("GermanyEight", func(t *testing.T) {
		plusCode, err := EncodeLength(48.56344833333333, 8.996878333333333, 8)
		if err != nil {
			t.Fatal(err)
		}

		expected := "8FWCHX7W+"

		assert.Equal(t, expected, plusCode)
	})
	t.Run("GermanySeven", func(t *testing.T) {
		plusCode, err := EncodeLength(48.56344833333333, 8.996878333333333, 7)
		if err != nil {
			t.Fatal(err)
		}

		expected := "8FWCHX7W+"

		assert.Equal(t, expected, plusCode)
	})
	t.Run("GermanySix", func(t *testing.T) {
		plusCode, err := EncodeLength(48.56344833333333, 8.996878333333333, 6)
		if err != nil {
			t.Fatal(err)
		}

		expected := "8FWCHX00+"

		assert.Equal(t, expected, plusCode)
	})
	t.Run("LatOverflow", func(t *testing.T) {
		plusCode, err := EncodeLength(548.56344833333333, 8.996878333333333, 7)
		if err == nil {
			t.Fatal("encode should return error")
		}
		assert.Equal(t, "", plusCode)
	})
	t.Run("LngOverflow", func(t *testing.T) {
		plusCode, err := EncodeLength(48.56344833333333, 258.996878333333333, 7)
		if err == nil {
			t.Fatal("encode should return error")
		}
		assert.Equal(t, "", plusCode)
	})
}

func TestS2(t *testing.T) {
	t.Run("Germany", func(t *testing.T) {
		token := S2("8FWCHX7W+")

		assert.Equal(t, "4799e3772d14", token)
	})
	t.Run("EmptyCode", func(t *testing.T) {
		token := S2("")

		assert.Equal(t, "", token)
	})
	t.Run("InvalidCode", func(t *testing.T) {
		token := S2("xxx")

		assert.Equal(t, "", token)
	})
}
