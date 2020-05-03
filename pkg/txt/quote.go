package txt

import (
	"fmt"
	"strings"
)

// Quote adds quotation marks to a string if needed.
func Quote(text string) string {
	if strings.ContainsAny(text, " \n'\"") {
		return fmt.Sprintf("“%s”", text)
	}

	return text
}
