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

// JoinNames joins a list of names to be used in titles and descriptions.
func JoinNames(names []string) string {
	if l := len(names); l == 0 {
		return ""
	} else if l == 1 {
		return names[0]
	} else if l == 2 {
		return strings.Join(names, " & ")
	} else {
		return fmt.Sprintf("%s & %s", strings.Join(names[:l-1], ", "), names[l-1])
	}
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
