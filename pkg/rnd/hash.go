package rnd

// IsMD5 checks if the string appears to be an MD5 hash.
// Example: 79054025255fb1a26e4bc422aef54eb4
func IsMD5(s string) bool {
	if len(s) != 32 {
		return false
	}

	return IsHex(s)
}

// IsSHA1 checks if the string appears to be a SHA1 hash.
// Example: de9f2c7fd25e1b3afad3e85a0bd17d9b100db4b3
func IsSHA1(s string) bool {
	if len(s) != 40 {
		return false
	}

	return IsHex(s)
}

// IsSHA224 checks if the string appears to be a SHA224 hash.
// Example: d14a028c2a3a2bc9476102bb288234c415a2b01f828ea62ac5b3e42f
func IsSHA224(s string) bool {
	if len(s) != 56 {
		return false
	}

	return IsHex(s)
}

// IsSHA256 checks if the string appears to be a SHA256 hash.
// Example: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
func IsSHA256(s string) bool {
	if len(s) != 64 {
		return false
	}

	return IsHex(s)
}

// IsSHA384 checks if the string appears to be a SHA384 hash.
// Example: 38b060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b
func IsSHA384(s string) bool {
	if len(s) != 96 {
		return false
	}

	return IsHex(s)

}

// IsSHA512 checks if the string appears to be a SHA512 hash.
// Example: cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e
func IsSHA512(s string) bool {
	if len(s) != 128 {
		return false
	}

	return IsHex(s)
}
