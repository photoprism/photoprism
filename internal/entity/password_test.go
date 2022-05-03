package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := NewPassword("abc567", "passwd")
		assert.Len(t, p.Hash, 60)
	})
	t.Run("empty password", func(t *testing.T) {
		p := NewPassword("abc567", "")
		assert.Equal(t, "", p.Hash)
	})
}

func TestPassword_SetPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := NewPassword("abc567", "passwd")
		assert.Len(t, p.Hash, 60)
		if err := p.SetPassword("abcd"); err != nil {
			t.Fatal(err)
		}
		assert.Len(t, p.Hash, 60)
	})
}

func TestPassword_InvalidPasswordPassword(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		p := Password{Hash: ""}
		assert.False(t, p.InvalidPassword(""))
	})
	t.Run("false", func(t *testing.T) {
		p := NewPassword("abc567", "")
		assert.False(t, p.InvalidPassword(""))
	})
	t.Run("true", func(t *testing.T) {
		p := NewPassword("abc567", "passwd")
		assert.True(t, p.InvalidPassword("$2a$14$p3HKuLvrTuePG/pjXLJQseUnSeAVeVO2cy4b0.34KXsLPK8lkI92G"))
	})
}

//TODO fails on mariadb
func TestPassword_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := Password{}

		err := p.Create()

		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestFindPassword(t *testing.T) {
	t.Run("not existing", func(t *testing.T) {
		r := FindPassword("xxx")
		assert.Nil(t, r)
	})
	t.Run("existing", func(t *testing.T) {
		p := NewPassword("abc567", "passwd")
		if err := p.Save(); err != nil {
			t.Fatal(err)
		}
		r := FindPassword("abc567")
		assert.NotEmpty(t, r)
	})
	t.Run("alice", func(t *testing.T) {
		if p := FindPassword("uqxetse3cy5eo9z2"); p == nil {
			t.Fatal("password not found")
		} else {
			assert.False(t, p.InvalidPassword("Alice123!"))
		}
	})
	t.Run("bob", func(t *testing.T) {
		if p := FindPassword("uqxc08w3d0ej2283"); p == nil {
			t.Fatal("password not found")
		} else {
			assert.False(t, p.InvalidPassword("Bobbob123!"))
		}
	})
}

func TestPassword_String(t *testing.T) {
	t.Run("return string", func(t *testing.T) {
		p := NewPassword("abc567", "lkjhgtyu")
		assert.Len(t, p.String(), 60)
	})
}

func TestPassword_Unknown(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		p := NewPassword("abc567", "lkjhgtyu")
		assert.False(t, p.Unknown())
	})
	t.Run("true", func(t *testing.T) {
		p := Password{}
		assert.True(t, p.Unknown())
	})
}
