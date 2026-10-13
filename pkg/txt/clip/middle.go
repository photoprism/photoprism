package clip

import "unicode/utf8"

// Middle shortens s to at most maxBytes without splitting a character, replacing a part in the middle with
// an ellipsis so that the end stays visible. It keeps at most tail bytes of the end, or a third of the
// remaining budget if tail is not positive.
func Middle(s string, maxBytes, tail int) string {
	return MiddleFunc(s, maxBytes, tail, runeLen)
}

// MiddleFunc shortens s like Middle, cutting only between the units whose length unit returns for the text
// that starts with them, so that a caller can keep a longer sequence than a character in one piece. A tail
// that does not fit the budget keeps a third of it as well.
func MiddleFunc(s string, maxBytes, tail int, unit func(s string) int) string {
	if len(s) <= maxBytes {
		return s
	}

	budget := maxBytes - len(Ellipsis)

	if budget <= 0 {
		return ""
	}

	if tail <= 0 || tail > budget {
		tail = budget / 3
	}

	// Only the ends are kept, so the rest is not walked.
	s = Ends(s, 4*maxBytes)

	head, end := 0, len(s)

	// The end starts at the first unit boundary that leaves at most tail bytes after it.
	for i := 0; i < len(s); {
		if len(s)-i <= tail {
			end = i
			break
		}

		i += max(unit(s[i:]), 1)
	}

	// The start takes what the end leaves of the budget.
	for i := 0; i < end; {
		n := max(unit(s[i:]), 1)

		if i+n > budget-(len(s)-end) {
			break
		}

		i += n
		head = i
	}

	return s[:head] + Ellipsis + s[end:]
}

// Ends returns s if it has at most maxBytes, and otherwise its first and last maxBytes/2 bytes, cut on
// character boundaries. It bounds the work of a caller that keeps only both ends of a long text.
func Ends(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}

	head, tail := maxBytes/2, len(s)-maxBytes/2

	for head > 0 && !utf8.RuneStart(s[head]) {
		head--
	}

	for tail < len(s) && !utf8.RuneStart(s[tail]) {
		tail++
	}

	return s[:head] + s[tail:]
}

// runeLen returns the length of the character s starts with.
func runeLen(s string) int {
	_, n := utf8.DecodeRuneInString(s)
	return n
}
