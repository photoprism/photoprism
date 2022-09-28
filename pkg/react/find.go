package react

// Find finds a reaction by name and emoji.
func Find(reaction string) Emoji {
	if reaction == "" {
		return Unknown
	}

	if found := Reactions[reaction]; found != "" {
		return found
	} else if found = Emoji(reaction); found.Unknown() {
		return Unknown
	} else {
		return found
	}
}

// Known checks if the emoji represents a known reaction.
func Known(reaction string) bool {
	return !Find(reaction).Unknown()
}
