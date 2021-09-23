package txt

import (
	"fmt"
	"strings"
)

// Quote adds quotation marks to a string if needed.
func Quote(text string) string {
	if text == "" || strings.ContainsAny(text, " \n'\"") {
		return fmt.Sprintf("“%s”", text)
	}

	return text
}

// QuoteLower converts a string to lowercase and adds quotation marks if needed.
func QuoteLower(text string) string {
	return Quote(strings.ToLower(text))
}
