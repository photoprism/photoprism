package header

import (
	"fmt"
	"regexp"
	"strings"
)

// Bots contains common search engine bot names.
var Bots = []string{
	"Googlebot",
	"PetalBot",
	"Bingbot",
	"DuckDuckBot",
	"Yahoo! Slurp",
	"Baiduspider",
	"Bytespider",
	"YandexBot",
	"Sogou",
	"Exabot",
	"Applebot",
	"facebot",
	"github-camo",
}

// BotsRegexp matches common search engine bot names.
var BotsRegexp = regexp.MustCompile(fmt.Sprintf("(%s)", strings.Join(Bots, "|")))

// IsBot checks whether the specified user agent string probably belongs to a search engine bot.
func IsBot(userAgent string) bool {
	return BotsRegexp.FindString(userAgent) != ""
}
