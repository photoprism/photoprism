package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToASCII(t *testing.T) {
	result := ToASCII("幸福 = Happiness.")
	assert.Equal(t, " = Happiness.", result)
}

func TestTrim(t *testing.T) {
	result := Trim(" 幸福 Hanzi are logograms developed for the writing of Chinese! ", 16)
	assert.Equal(t, "幸福 Hanzi are logograms developed for the writing of Chinese", result)
}

func TestSanitizeTypeString(t *testing.T) {
	result := SanitizeTypeString(" 幸福 Hanzi are logograms developed for the writing of Chinese! ")
	assert.Equal(t, "hanzi are logograms developed for the writing of chinese", result)
}
