package dsn

import (
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// identPattern matches a database name or user that needs no quoting in a DSN or client command.
var identPattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_-]{0,63}$`)

// hostLabelPattern matches a host name label in a database server address.
var hostLabelPattern = regexp.MustCompile(`^[A-Za-z0-9_]([A-Za-z0-9_-]{0,61}[A-Za-z0-9_])?$`)

// ValidIdent reports whether s is a database name or user of up to 64 letters, digits, "_", and "-",
// not beginning with "-".
func ValidIdent(s string) bool {
	return identPattern.MatchString(s)
}

// ValidServer reports whether s is a host name or IP address with an optional port, or a port alone
// in the form ":3306".
func ValidServer(s string) bool {
	host, port := s, ""

	if strings.HasPrefix(s, "[") || strings.Count(s, ":") == 1 {
		h, p, err := net.SplitHostPort(s)

		if err != nil {
			return false
		}

		host, port = h, p
	}

	if port != "" {
		if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 {
			return false
		}
	}

	switch {
	case host == "":
		return port != ""
	case net.ParseIP(host) != nil:
		return true
	case len(host) > 253:
		return false
	}

	for label := range strings.SplitSeq(strings.TrimSuffix(host, "."), ".") {
		if !hostLabelPattern.MatchString(label) {
			return false
		}
	}

	return true
}

// ValidBool reports whether v is a boolean DSN parameter value the MySQL driver accepts.
func ValidBool(v string) bool {
	return IsTrue(v) || IsFalse(v)
}

// IsTrue reports whether v is a DSN parameter value the MySQL driver reads as true.
func IsTrue(v string) bool {
	switch v {
	case "1", "true", "TRUE", "True":
		return true
	}

	return false
}

// IsFalse reports whether v is a DSN parameter value the MySQL driver reads as false.
func IsFalse(v string) bool {
	switch v {
	case "0", "false", "FALSE", "False":
		return true
	}

	return false
}

// ValidDuration reports whether v is a positive duration DSN parameter value such as 15s.
func ValidDuration(v string) bool {
	d, err := time.ParseDuration(v)
	return err == nil && d > 0 && !strings.ContainsAny(v, "+-")
}
