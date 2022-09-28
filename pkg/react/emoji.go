package react

// Emoji represents user feedback expressed by an emoji:
// https://www.unicode.org/Public/emoji/14.0/emoji-sequences.txt
type Emoji string

// Unknown checks if the reaction is unknown.
func (emo Emoji) Unknown() bool {
	if l := len(emo); l < 2 || len(emo) > 64 {
		return true
	}

	return Names[emo] == ""
}

// Name returns the ASCII name of the reaction.
func (emo Emoji) Name() string {
	return Names[emo]
}

// String returns the reaction as string.
func (emo Emoji) String() string {
	return string(emo)
}

// Bytes returns the reaction emoji as a slice with a maximum size of 64 bytes.
func (emo Emoji) Bytes() (b []byte) {
	if b = []byte(emo); len(b) <= 64 {
		return b
	}

	return b[0:64]
}
