package rnd

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	petname "github.com/dustinkirkland/golang-petname"
)

// Name returns a pronounceable name consisting of a pet name and an adverb or adjective.
func Name() string {
	return NameN(2)
}

// NameN returns a pronounceable name consisting of a random combination of adverbs, an adjective, and a pet name.
func NameN(n int) string {
	return cases.Title(language.English, cases.Compact).String(petname.Generate(n, " "))
}
