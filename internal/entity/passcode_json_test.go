package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPasscode_MarshalJSON(t *testing.T) {
	m := &Passcode{
		UID:          "uqxc08w3d0ej2283",
		KeyURL:       "otpauth://totp/Example:alice",
		RecoveryCode: "123",
	}

	r, err := m.MarshalJSON()

	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, string(r), "uqxc08w3d0ej2283")
}
