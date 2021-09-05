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
func JoinNames(names []string) (result string) {
	l := len(names)

	if l == 0 {
		return ""
	} else if l == 1 {
		return names[0]
	}

	var familySuffix string

	if i := strings.LastIndex(names[0], " "); i > 1 && len(names[0][i:]) > 2 {
		familySuffix = names[0][i:]

		for i := 1; i < l; i++ {
			if !strings.HasSuffix(names[i], familySuffix) {
				familySuffix = ""
				break
			}
		}
	}

	if l == 2 {
		result = strings.Join(names, " & ")
	} else {
		result = fmt.Sprintf("%s & %s", strings.Join(names[:l-1], ", "), names[l-1])
	}

	if familySuffix != "" {
		result = strings.Replace(result, familySuffix, "", l-1)
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
