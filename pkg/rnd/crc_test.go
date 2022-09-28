package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrcToken(t *testing.T) {
	t.Run("size 4", func(t *testing.T) {
		token := CrcToken()
		t.Logf("CrcToken: %s", token)
		assert.NotEmpty(t, token)
	})
}

func TestValidateCrcToken(t *testing.T) {
	t.Run("Static", func(t *testing.T) {
		assert.True(t, ValidateCrcToken("8yva-2ni4-385e"))
	})
	t.Run("Generated", func(t *testing.T) {
		assert.True(t, ValidateCrcToken(CrcToken()))
	})
	t.Run("InvalidChecksum", func(t *testing.T) {
		assert.False(t, ValidateCrcToken("8yva-2ni4-185e"))
	})
	t.Run("InvalidToken", func(t *testing.T) {
		assert.False(t, ValidateCrcToken("8ava-2ni4-385e"))
	})
	t.Run("ValidToken", func(t *testing.T) {
		assert.True(t, ValidateCrcToken("hsqq-181q-13ed"))
	})
}
