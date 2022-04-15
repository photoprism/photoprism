package rnd

// GeneratePasswd returns a random password with 8 characters as string.
func GeneratePasswd() string {
	return GenerateToken(8)
}
