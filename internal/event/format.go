package event

import (
	"fmt"
	"strings"
)

// MessageSep separates event topic segments when rendered as text.
var MessageSep = " › "

// Format formats an audit log event.
func Format(ev []string, args ...any) string {
	return fmt.Sprintf(strings.Join(ev, MessageSep), args...)
}
