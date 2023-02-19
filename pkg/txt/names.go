package txt

import (
	"fmt"
	"strings"
)

// UniqueNames removes exact duplicates from a list of strings without changing their order.
func UniqueNames(names []string) (result []string) {
	if len(names) < 1 {
		return []string{}
	}

	k := make(map[string]bool)

	for _, n := range names {
		if _, value := k[n]; !value {
			k[n] = true
			result = append(result, n)
		}
	}

	return result
}

// AppendName appends a name to an existing name.
func AppendName(s, n string) string {
	s = strings.TrimSpace(s)
	n = strings.TrimSpace(n)

	if s == "" {
		return n
	} else if s == n {
		return s
	}

	return fmt.Sprintf("%s %s", s, n)
}

// JoinNames joins a list of names to be used in titles and descriptions.
func JoinNames(names []string, shorten bool) (result string) {
	l := len(names)

	if l == 0 {
		return ""
	} else if l == 1 {
		return names[0]
	}

	var familyName string

	// Common family name?
	if i := strings.LastIndex(names[0], " "); i > 1 && len(names[0][i:]) > 2 {
		familyName = names[0][i:]

		for i := 1; i < l; i++ {
			if !strings.HasSuffix(names[i], familyName) {
				familyName = ""
				break
			}
		}
	}

	// Shorten names?
	if shorten {
		shortNames := make([]string, l)

		for i := 0; i < l; i++ {
			shortNames[i] = strings.SplitN(names[i], Space, 2)[0]

			if i > 0 && shortNames[i] == strings.SplitN(names[i-1], Space, 2)[0] {
				shortNames[i] = names[i]
				shortNames[i-1] = names[i-1]
			}
		}

		names = shortNames
	}

	if l == 2 {
		result = strings.Join(names, " & ")
	} else {
		result = fmt.Sprintf("%s & %s", strings.Join(names[:l-1], ", "), names[l-1])
	}

	// Keep family name at the end.
	if familyName != "" {
		if shorten {
			result = result + familyName
		} else {
			result = strings.Replace(result, familyName, "", l-1)
		}
	}

	return result
}

// NameKeywords returns a list of unique, lowercase keywords based on a person's names and aliases.
func NameKeywords(names, aliases string) (results []string) {
	if names == "" && aliases == "" {
		return []string{}
	}

	names = strings.ToLower(names)
	aliases = strings.ToLower(aliases)

	return UniqueNames(append(Words(names), Words(aliases)...))
}
