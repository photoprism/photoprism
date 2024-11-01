package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPassword(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "passwd", false)
		assert.Len(t, p.Hash, 60)
	})
	t.Run("empty password", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "", false)
		assert.Equal(t, "", p.Hash)
	})
}

func TestPassword_SetPassword(t *testing.T) {
	t.Run("Text", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "passwd", false)
		assert.Len(t, p.Hash, 60)
		assert.True(t, p.Valid("passwd"))
		assert.False(t, p.Valid("other"))

		if err := p.SetPassword("abcd", false); err != nil {
			t.Fatal(err)
		}

		assert.Len(t, p.Hash, 60)
		assert.True(t, p.Valid("abcd"))
		assert.False(t, p.Valid("other"))
	})
	t.Run("Too long", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "hgfttrgkncgdhfkbvuvygvbekdjbrtugbnljbtruhogtgbotuhblenbhoyuhntyyhngytohrpnehotyihniy", false)

		err := p.SetPassword("hgfttrgkncgdhfkbvuvygvbekdjbrtugbnljbtruhogtgbotuhblenbhoyuhntyyhngytohrpnehotyihniy", false)

		assert.Error(t, err)
	})
	t.Run("Too short", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "", false)

		err := p.SetPassword("", false)

		assert.Error(t, err)
	})
	t.Run("Hash", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "$2a$14$qCcNjxupSJV1gjhgdYxz8e9l0e0fTZosX0s0qhMK54IkI9YOyWLt2", true)
		assert.Len(t, p.Hash, 60)
		assert.True(t, p.Valid("photoprism"))
		assert.False(t, p.Valid("$2a$14$qCcNjxupSJV1gjhgdYxz8e9l0e0fTZosX0s0qhMK54IkI9YOyWLt2"))
		assert.False(t, p.Valid("other"))
	})
}

func TestPassword_Valid(t *testing.T) {
	t.Run("EmptyHash", func(t *testing.T) {
		p := Password{Hash: ""}
		assert.True(t, p.Empty())
		assert.False(t, p.Valid(""))
	})
	t.Run("EmptyPassword", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "", false)
		assert.True(t, p.Empty())
		assert.False(t, p.Valid(""))
	})
	t.Run("ShortPassword", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "passwd", false)
		assert.True(t, p.Valid("passwd"))
		assert.False(t, p.Valid("$2a$14$p3HKuLvrTuePG/pjXLJQseUnSeAVeVO2cy4b0.34KXsLPK8lkI92G"))
	})
	t.Run("LongPassword", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "photoprism", false)
		assert.True(t, p.Valid("photoprism"))
		assert.False(t, p.Valid("$2a$14$p3HKuLvrTuePG/pjXLJQseUnSeAVeVO2cy4b0.34KXsLPK8lkI92G"))
	})
}

func TestPassword_Invalid(t *testing.T) {
	t.Run("EmptyHash", func(t *testing.T) {
		p := Password{Hash: ""}
		assert.True(t, p.Empty())
		assert.True(t, p.Invalid(""))
	})
	t.Run("EmptyPassword", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "", false)
		assert.True(t, p.Empty())
		assert.True(t, p.Invalid(""))
	})
	t.Run("ShortPassword", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "passwd", false)
		assert.True(t, p.Invalid("$2a$14$p3HKuLvrTuePG/pjXLJQseUnSeAVeVO2cy4b0.34KXsLPK8lkI92G"))
		assert.False(t, p.Invalid("passwd"))
	})
	t.Run("LongPassword", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "photoprism", false)
		assert.True(t, p.Invalid("$2a$14$p3HKuLvrTuePG/pjXLJQseUnSeAVeVO2cy4b0.34KXsLPK8lkI92G"))
		assert.False(t, p.Invalid("photoprism"))
	})
}

func TestPassword_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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
		p := NewPassword("urrwaxd19ldtz68x", "passwd", false)
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
			assert.False(t, p.Invalid("Alice123!"))
		}
	})
	t.Run("Bob", func(t *testing.T) {
		if p := FindPassword("uqxc08w3d0ej2283"); p == nil {
			t.Fatal("password not found")
		} else {
			assert.False(t, p.Invalid("Bobbob123!"))
		}
	})
}

func TestPassword_Cost(t *testing.T) {
	t.Run("DefaultPasswordCost", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "photoprism", false)
		if cost, err := p.Cost(); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, DefaultPasswordCost, cost)
		}
	})
	t.Run("PasswordCost14", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "$2a$14$qCcNjxupSJV1gjhgdYxz8e9l0e0fTZosX0s0qhMK54IkI9YOyWLt2", true)
		if cost, err := p.Cost(); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 14, cost)
		}
	})
	t.Run("EmptyPassword", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "", false)
		_, err := p.Cost()
		assert.Error(t, err)
	})
}

func TestPassword_String(t *testing.T) {
	t.Run("BCrypt", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "lkjhgtyu", false)
		assert.Len(t, p.String(), 60)
	})
}

func TestPassword_IsEmpty(t *testing.T) {
	t.Run("False", func(t *testing.T) {
		p := NewPassword("urrwaxd19ldtz68x", "lkjhgtyu", false)
		assert.False(t, p.Empty())
	})
	t.Run("True", func(t *testing.T) {
		p := Password{}
		assert.True(t, p.Empty())
	})
}
