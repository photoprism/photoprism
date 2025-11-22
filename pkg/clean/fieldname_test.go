package clean

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "Path", in: "/go/src/github.com/photoprism/photoprism", want: "gosrcgithubcomphotoprismphotoprism"},
		{name: "DotsAndUpper", in: "filename.TXT", want: "filenameTXT"},
		{name: "SpacesAndPunctuation", in: "The quick brown fox.", want: "Thequickbrownfox"},
		{name: "QuestionAndDot", in: "file?name.jpg", want: "filenamejpg"},
		{name: "ControlCharacter", in: "filename." + string(rune(127)), want: "filename"},
		{name: "Empty", in: "", want: ""},
		{name: "TooLong", in: strings.Repeat("a", 256), want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FieldName(tt.in))
		})
	}
}

func TestFieldNameLower(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "LowerCase", in: "file?name.JPG", want: "filenamejpg"},
		{name: "UpperOnly", in: "ABC", want: "abc"},
		{name: "MixedSeparators", in: "Album-Photos_123", want: "albumphotos123"},
		{name: "TooLong", in: strings.Repeat("B", 300), want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FieldNameLower(tt.in))
		})
	}
}
