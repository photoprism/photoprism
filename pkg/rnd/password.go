package rnd

// Password returns a random password with 8 characters as string.
func Password() string {
	return Token(8)
}
