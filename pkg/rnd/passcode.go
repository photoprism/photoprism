package rnd

// GeneratePasscode returns a random 16-digit passcode that can, for example, be used as an app password.
// It is separated by 3 dashes for better readability, resulting in a total length of 19 characters.
func GeneratePasscode() string {
	code := make([]byte, 0, 19)
	code = append(code, Base62(4)...)

	for n := 0; n < 3; n++ {
		code = append(code, "-"+Base62(4)...)
	}

	return string(code)
}
