package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeUnicode(t *testing.T) {
	t.Run("Ascii", func(t *testing.T) {
		assert.Equal(t, "IMG_0599", SanitizeUnicode("IMG_0599"))
	})

	t.Run("Unicode", func(t *testing.T) {
		assert.Equal(t, "Naïve bonds and futures surge as inflation eases 🚀🚀🚀", SanitizeUnicode("  Naïve bonds and futures surge as inflation eases 🚀🚀🚀 "))
	})

	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", SanitizeUnicode(""))
	})
}

func TestSanitizeTitle(t *testing.T) {
	t.Run("IMG_0599", func(t *testing.T) {
		result := SanitizeTitle("IMG_0599")

		if result != "" {
			t.Fatal("result should be empty")
		}
	})

	t.Run("IMG_0599.JPG", func(t *testing.T) {
		result := SanitizeTitle("IMG_0599.JPG")

		if result != "" {
			t.Fatal("result should be empty")
		}
	})

	t.Run("IMG_0599 ABC", func(t *testing.T) {
		result := SanitizeTitle("IMG_0599 ABC")

		if result != "IMG_0599 ABC" {
			t.Fatal("result should be IMG_0599 ABC")
		}
	})

	t.Run("DSC10599", func(t *testing.T) {
		result := SanitizeTitle("DSC10599")

		if result != "" {
			t.Fatal("result should be empty")
		}
	})

	t.Run("titanic_cloud_computing.jpg", func(t *testing.T) {
		result := SanitizeTitle("titanic_cloud_computing.jpg")

		assert.Equal(t, "Titanic Cloud Computing", result)
	})

	t.Run("naomi-watts--ewan-mcgregor--the-impossible--tiff-2012_7999540939_o.jpg", func(t *testing.T) {
		result := SanitizeTitle("naomi-watts--ewan-mcgregor--the-impossible--tiff-2012_7999540939_o.jpg")

		assert.Equal(t, "Naomi Watts / Ewan McGregor / The Impossible / TIFF", result)
	})

	t.Run("Bei den Landungsbrücken.png", func(t *testing.T) {
		result := SanitizeTitle("Bei den Landungsbrücken.png")

		assert.Equal(t, "Bei den Landungsbrücken", result)
	})

	t.Run("Bei den Landungsbrücken.foo", func(t *testing.T) {
		result := SanitizeTitle("Bei den Landungsbrücken.foo")

		assert.Equal(t, "Bei den Landungsbrücken.foo", result)
	})

	t.Run("let_it_snow", func(t *testing.T) {
		result := SanitizeTitle("let_it_snow")

		assert.Equal(t, "let_it_snow", result)
	})

	t.Run("let_it_snow.jpg", func(t *testing.T) {
		result := SanitizeTitle("let_it_snow.jpg")

		assert.Equal(t, "Let It Snow", result)
	})

	t.Run("Niklaus_Wirth.jpg", func(t *testing.T) {
		result := SanitizeTitle("Niklaus_Wirth.jpg")

		assert.Equal(t, "Niklaus Wirth", result)
	})

	t.Run("Niklaus_Wirth", func(t *testing.T) {
		result := SanitizeTitle("Niklaus_Wirth")

		assert.Equal(t, "Niklaus_Wirth", result)
	})

	t.Run("string with binary data", func(t *testing.T) {
		result := SanitizeTitle("string with binary data blablabla")

		assert.Equal(t, "", result)
	})
}

func TestSanitizeDescription(t *testing.T) {
	t.Run("IMG_0599", func(t *testing.T) {
		result := SanitizeDescription("IMG_0599")

		if result == "" {
			t.Fatal("result should not be empty")
		}
	})

	t.Run("OLYMPUS DIGITAL CAMERA", func(t *testing.T) {
		result := SanitizeDescription("OLYMPUS DIGITAL CAMERA")

		if result != "" {
			t.Fatal("result should be empty")
		}
	})

	t.Run("GoPro", func(t *testing.T) {
		result := SanitizeDescription("DCIM\\108GOPRO\\GOPR2137.JPG")

		if result != "" {
			t.Fatal("result should be empty")
		}
	})

	t.Run("hdrpl", func(t *testing.T) {
		result := SanitizeDescription("hdrpl")

		if result != "" {
			t.Fatal("result should be empty")
		}
	})

	t.Run("btf", func(t *testing.T) {
		result := SanitizeDescription("btf")

		if result != "" {
			t.Fatal("result should be empty")
		}
	})

	t.Run("wtf", func(t *testing.T) {
		result := SanitizeDescription("wtf")

		if result != "wtf" {
			t.Fatal("result should be 'wtf'")
		}
	})
}

func TestSanitizeUID(t *testing.T) {
	t.Run("77d9a719ede3f95915abd081d7b7cb2c", func(t *testing.T) {
		result := SanitizeUID("77d9a719ede3f95915abd081d7b7CB2c")
		assert.Equal(t, "77d9a719ede3f95915abd081d7b7cb2c", result)
	})
	t.Run("77d", func(t *testing.T) {
		result := SanitizeUID("77d")
		assert.Equal(t, "", result)
	})
	t.Run(":77d9a719ede3f95915abd081d7b7cb2c", func(t *testing.T) {
		result := SanitizeUID(":77d9a719ede3f95915abd081d7b7CB2c")
		assert.Equal(t, "77d9a719ede3f95915abd081d7b7cb2c", result)
	})

}
