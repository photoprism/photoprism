package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "passwd")
		assert.Len(t, p.Hash, 60)
	})
	t.Run("empty password", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "")
		assert.Equal(t, "", p.Hash)
	})
}

func TestPassword_SetPassword(t *testing.T) {
	t.Run("Text", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "passwd")
		assert.Len(t, p.Hash, 60)
		assert.True(t, p.IsValid("passwd"))
		assert.False(t, p.IsValid("other"))

		if err := p.SetPassword("abcd"); err != nil {
			t.Fatal(err)
		}

		assert.Len(t, p.Hash, 60)
		assert.True(t, p.IsValid("abcd"))
		assert.False(t, p.IsValid("other"))
	})
	t.Run("Hash", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "$2a$14$qCcNjxupSJV1gjhgdYxz8e9l0e0fTZosX0s0qhMK54IkI9YOyWLt2")
		assert.Len(t, p.Hash, 60)
		assert.True(t, p.IsValid("photoprism"))
		assert.False(t, p.IsValid("$2a$14$qCcNjxupSJV1gjhgdYxz8e9l0e0fTZosX0s0qhMK54IkI9YOyWLt2"))
		assert.False(t, p.IsValid("other"))
	})
}

func TestPassword_IsValid(t *testing.T) {
	t.Run("EmptyHash", func(t *testing.T) {
		p := Password{Hash: ""}
		assert.True(t, p.IsEmpty())
		assert.False(t, p.IsValid(""))
	})
	t.Run("EmptyPassword", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "")
		assert.True(t, p.IsEmpty())
		assert.False(t, p.IsValid(""))
	})
	t.Run("ShortPassword", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "passwd")
		assert.True(t, p.IsValid("passwd"))
		assert.False(t, p.IsValid("$2a$14$p3HKuLvrTuePG/pjXLJQseUnSeAVeVO2cy4b0.34KXsLPK8lkI92G"))
	})
	t.Run("LongPassword", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "photoprism")
		assert.True(t, p.IsValid("photoprism"))
		assert.False(t, p.IsValid("$2a$14$p3HKuLvrTuePG/pjXLJQseUnSeAVeVO2cy4b0.34KXsLPK8lkI92G"))
	})
}

func TestPassword_IsWrong(t *testing.T) {
	t.Run("EmptyHash", func(t *testing.T) {
		p := Password{Hash: ""}
		assert.True(t, p.IsEmpty())
		assert.True(t, p.IsWrong(""))
	})
	t.Run("EmptyPassword", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "")
		assert.True(t, p.IsEmpty())
		assert.True(t, p.IsWrong(""))
	})
	t.Run("ShortPassword", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "passwd")
		assert.True(t, p.IsWrong("$2a$14$p3HKuLvrTuePG/pjXLJQseUnSeAVeVO2cy4b0.34KXsLPK8lkI92G"))
		assert.False(t, p.IsWrong("passwd"))
	})
	t.Run("LongPassword", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "photoprism")
		assert.True(t, p.IsWrong("$2a$14$p3HKuLvrTuePG/pjXLJQseUnSeAVeVO2cy4b0.34KXsLPK8lkI92G"))
		assert.False(t, p.IsWrong("photoprism"))
	})
}

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
	t.Run("NotFound", func(t *testing.T) {
		r := FindPassword("xxx")
		assert.Nil(t, r)
	})
	t.Run("Exists", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "passwd")
		if err := p.Save(); err != nil {
			t.Fatal(err)
		}
		r := FindPassword("urrwaxd19ldtz68x")
		assert.NotEmpty(t, r)
	})
	t.Run("Alice", func(t *testing.T) {
		if p := FindPassword("uqxetse3cy5eo9z2"); p == nil {
			t.Fatal("password not found")
		} else {
			assert.False(t, p.IsWrong("Alice123!"))
		}
	})
	t.Run("Bob", func(t *testing.T) {
		if p := FindPassword("uqxc08w3d0ej2283"); p == nil {
			t.Fatal("password not found")
		} else {
			assert.False(t, p.IsWrong("Bobbob123!"))
		}
	})
}

func TestPassword_Cost(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "photoprism")
		if cost, err := p.Cost(); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, PasswordCost, cost)
		}
	})
	t.Run("14", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "$2a$14$qCcNjxupSJV1gjhgdYxz8e9l0e0fTZosX0s0qhMK54IkI9YOyWLt2")
		if cost, err := p.Cost(); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 14, cost)
		}
	})
}

func TestPassword_String(t *testing.T) {
	t.Run("return string", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "lkjhgtyu")
		assert.Len(t, p.String(), 60)
	})
}

func TestPassword_IsEmpty(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "lkjhgtyu")
		assert.False(t, p.IsEmpty())
	})
	t.Run("true", func(t *testing.T) {
		p := Password{}
		assert.True(t, p.IsEmpty())
	})
}
