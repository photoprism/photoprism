package rnd

const (
	ClientSecretLength = 32
)

// ClientSecret generates a random client secret containing 32 upper and lower case letters as well as numbers.
func ClientSecret() string {
	return Base62(ClientSecretLength)
}

// IsClientSecret checks if the string represents a valid client secret.
func IsClientSecret(s string) bool {
	if l := len(s); l == ClientSecretLength {
		return IsAlnum(s)
	}

	return false
}
