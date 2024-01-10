package rnd

import (
	petname "github.com/dustinkirkland/golang-petname"
)

// Name returns a pronounceable name consisting of a pet name and an adverb or adjective.
func Name() string {
	return NameN(2)
}

// NameN returns a pronounceable name consisting of a random combination of adverbs, an adjective, and a pet name.
func NameN(n int) string {
	return petname.Generate(n, "-")
}
