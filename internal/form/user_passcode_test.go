package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserPasscode_HasPassword(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		form := &UserPasscode{Passcode: "123456", Password: ""}
		assert.Equal(t, false, form.HasPassword())
	})
	t.Run("true", func(t *testing.T) {
		form := &UserPasscode{Passcode: "123456", Password: "passwd1234"}
		assert.Equal(t, true, form.HasPassword())
	})
}
