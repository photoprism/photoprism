package txt

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/photoprism/photoprism/pkg/fs"
)

var FileTitleRegexp = regexp.MustCompile("[\\p{L}\\-,':]{2,}")

var SpecialWords = map[string]string{
	"nyc":                "NYC",
	"ny":                 "NY",
	"uae":                "UAE",
	"usa":                "USA",
	"amd":                "AMD",
	"tiff":               "TIFF",
	"ibm":                "IBM",
	"usd":                "USD",
	"gbp":                "GBP",
	"chf":                "CHF",
	"ceo":                "CEO",
	"cto":                "CTO",
	"cfo":                "CFO",
	"cia":                "CIA ",
	"fbi":                "FBI",
	"bnd":                "BND",
	"fsb":                "FSB",
	"nsa":                "NSA",
	"lax":                "LAX",
	"sfx":                "SFX",
	"ber":                "BER",
	"sfo":                "SFO",
	"lh":                 "LH",
	"lhr":                "LHR",
	"afl":                "AFL",
	"nrl":                "NRL",
	"nsw":                "NSW",
	"qld":                "QLD",
	"vic":                "VIC",
	"iphone":             "iPhone",
	"imac":               "iMac",
	"ipad":               "iPad",
	"ipod":               "iPod",
	"macbook":            "MacBook",
	"airplay":            "AirPlay",
	"airpods":            "AirPods",
	"youtube":            "YouTube",
	"photoprism":         "PhotoPrism",
	"macgyver":           "MacGyver",
	"o'brien":            "O'Brien",
	"mcgregor":           "McGregor",
	"mcdonald":           "McDonald",
	"mcdonalds":          "McDonald's",
	"mcdonald's":         "McDonald's",
	"macalister":         "MacAlister",
	"mcalister":          "McAlister",
	"mcallister":         "McAllister",
	"macauley":           "MacAuley",
	"mccauley":           "McCauley",
	"mcawley":            "McAwley",
	"macauliffe":         "MacAuliffe",
	"macbride":           "MacBride",
	"mcbride":            "McBride",
	"maccabe":            "MacCabe",
	"mccabe":             "McCabe",
	"maccann":            "MacCann",
	"mccann":             "McCann",
	"maccarthy":          "MacCarthy",
	"mccarthy":           "McCarthy",
	"maccormack":         "MacCormack",
	"mccormick":          "McCormick",
	"maccullagh":         "MacCullagh",
	"macnully":           "MacNully",
	"mackenna":           "MacKenna",
	"macnamara":          "MacNamara",
	"mcnamara":           "McNamara",
	"gelaende":           "Gelände",
	"schwaebisch":        "Schwäbisch",
	"schwaebische":       "Schwäbische",
	"aegypten":           "Ägypten",
	"muenchen":           "München",
	"wuerttemberg":       "Württemberg",
	"baden-wuerttemberg": "Baden-Württemberg",
	"nuernberg":          "Nürnberg",
	"wuerzburg":          "Würzburg",
	"tubingen":           "Tübingen",
	"tuebingen":          "Tübingen",
	"koeln":              "Köln",
	"oesterreich":        "Österreich",
	"woerthersee":        "Wörthersee",
	"oeland":             "Öland",
	"schoenefeld":        "Schönefeld",
	"duesseldorf":        "Düsseldorf",
	"dusseldorf":         "Düsseldorf",
	"saarbrucken":        "Saarbrücken",
	"saarbruecken":       "Saarbrücken",
	"zuerich":            "Zürich",
}

var SmallWords = map[string]bool{
	"a":    true,
	"an":   true,
	"as":   true,
	"at":   true,
	"by":   true,
	"in":   true,
	"of":   true,
	"on":   true,
	"or":   true,
	"up":   true,
	"to":   true,
	"and":  true,
	"but":  true,
	"for":  true,
	"nor":  true,
	"the":  true,
	"from": true,
	"with": true,
}

// isSeparator reports whether the rune could mark a word boundary.
func isSeparator(r rune) bool {
	// ASCII alphanumerics and underscore are not separators
	if r <= 0x7F {
		switch {
		case '0' <= r && r <= '9':
			return false
		case 'a' <= r && r <= 'z':
			return false
		case 'A' <= r && r <= 'Z':
			return false
		case r == '_', r == '\'':
			return false
		}
		return true
	}
	// Letters and digits are not separators
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return false
	}
	// Otherwise, all we can do for now is treat spaces as separators.
	return unicode.IsSpace(r)
}

// UcFirst returns the string with the first character converted to uppercase.
func UcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

// Title returns the string with the first characters of each word converted to uppercase.
func Title(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.Trim(s, "/ -")
	blocks := strings.Split(s, "/")
	result := make([]string, 0, len(blocks))

	for _, block := range blocks {
		words := strings.Fields(block)

		if len(words) == 0 {
			continue
		}

		for i, w := range words {
			search := strings.ToLower(strings.Trim(w, ":.,;!?"))

			if match, ok := SpecialWords[search]; ok {
				words[i] = strings.Replace(strings.ToLower(w), search, match, 1)
			} else if i > 0 && SmallWords[search] {
				words[i] = strings.ToLower(w)
			} else {
				prev := ' '
				words[i] = strings.Map(
					func(r rune) rune {
						if isSeparator(prev) {
							prev = r
							return unicode.ToTitle(r)
						}
						prev = r
						return r
					},
					w)
			}
		}

		result = append(result, strings.Join(words, " "))
	}

	return strings.Join(result, " / ")
}

// TitleFromFileName returns the string with the first characters of each word converted to uppercase.
func TitleFromFileName(s string) string {
	s = fs.Base(s, true)

	if len(s) < 3 {
		return ""
	}

	words := FileTitleRegexp.FindAllString(s, -1)
	var result []string

	found := 0

	for _, w := range words {
		w = strings.ToLower(w)

		if len(w) < 3 && found == 0 {
			continue
		}

		if _, ok := Stopwords[w]; ok && found == 0 {
			continue
		}

		result = append(result, w)

		found++

		if found > 10 {
			break
		}
	}

	if found == 0 {
		return ""
	}

	title := strings.Join(result, " ")

	title = strings.ReplaceAll(title, "--", " / ")
	title = strings.ReplaceAll(title, "-", " ")
	title = strings.ReplaceAll(title, "  ", " ")

	if len(title) < 3 {
		return ""
	}

	return Title(title)
}
