package event

import (
	"fmt"
	"strings"
)

var MessageSep = " › "

// Format formats an audit log event.
func Format(ev []string, args ...interface{}) string {
	return fmt.Sprintf(strings.Join(ev, MessageSep), args...)
}
