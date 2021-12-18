package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirstOrCreateAddress(t *testing.T) {
	t.Run("existing address", func(t *testing.T) {
		address := Address{ID: 1234567, AddressLine1: "Line 1", AddressCountry: "DE"}

		result := FirstOrCreateAddress(&address)

		if result == nil {
			t.Fatal("result should not be nil")
		}
		t.Log(result)

		assert.Equal(t, 1234567, result.ID)
	})
	t.Run("not existing address", func(t *testing.T) {
		address := &Address{}

		result := FirstOrCreateAddress(address)

		if result == nil {
			t.Fatal("result should not be nil")
		}

		t.Log(result)

		assert.Equal(t, 1234568, result.ID)
	})
}

func TestAddress_String(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		address := Address{ID: 1234567, AddressLine1: "Line 1", AddressLine2: "Line 2", AddressCity: "Berlin", AddressCountry: "DE"}
		addressString := address.String()
		assert.Equal(t, "'Line 1,  Berlin, DE'", addressString)
	})
}

func TestAddress_Unknown(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		address := Address{ID: 1234567, AddressLine1: "Line 1", AddressLine2: "Line 2", AddressCity: "Berlin", AddressCountry: "DE"}
		assert.False(t, address.Unknown())
	})
}
