package fs

import (
	"path/filepath"
	"regexp"
	"strings"
)

// Insta360VideoPattern matches the lens and proxy files of an Insta360 separate-lens video capture.
var Insta360VideoPattern = regexp.MustCompile(`^((?i:VID|LRV))_(\d{8})_(\d{6})_(00|10|11)_(\d{3})\.[Ii][Nn][Ss][Vv]$`)

// Insta360PhotoPattern matches the lens files of an Insta360 separate-lens photo capture.
var Insta360PhotoPattern = regexp.MustCompile(`^((?i:IMG))_(\d{8})_(\d{6})_(00|10)_(\d{3})\.[Ii][Nn][Ss][Pp]$`)

// stackRule maps the files of a capture that is only viewable when stacked to one shared name.
type stackRule struct {
	pattern *regexp.Regexp
	ext     string
	name    func(match []string) string
}

// stackRules lists the name mappings applied by StackPrefix.
var stackRules = []stackRule{
	{pattern: Insta360VideoPattern, ext: ExtInsv, name: insta360StackName},
	{pattern: Insta360PhotoPattern, ext: ExtInsp, name: insta360StackName},
}

// StackPrefix returns the name under which a file is stacked with the other files of a photo.
// Files of a separate-lens capture and their sidecars share the name of the left lens file,
// regardless of stripSequence. For all other files, it returns the same as BasePrefix.
func StackPrefix(fileName string, stripSequence bool) string {
	prefix := BasePrefix(fileName, false)

	if name := stackRuleName(filepath.Base(fileName), prefix); name != "" {
		prefix = name
	}

	if !stripSequence {
		return prefix
	}

	return StripSequence(prefix)
}

// StackGroup returns the shared stack name of a lens or proxy original of a multi-file capture,
// including the file the others are stacked under, or an empty string for any other file.
func StackGroup(fileName string) string {
	baseName := filepath.Base(fileName)

	for _, rule := range stackRules {
		if match := rule.pattern.FindStringSubmatch(baseName); match != nil {
			return rule.name(match)
		}
	}

	return ""
}

// KeepStacked reports whether a file is an original stacked under the name of another file of its
// capture, such as a right lens or proxy, so it must not be separated from it. Sidecars are not.
func KeepStacked(fileName string) bool {
	name := StackGroup(fileName)
	return name != "" && name != BasePrefix(fileName, false)
}

// stackRuleName returns the shared stack name if the base name, cut after the extension that
// follows the prefix, matches a stack rule, or an empty string otherwise.
func stackRuleName(baseName, prefix string) string {
	for _, rule := range stackRules {
		n := len(prefix) + len(rule.ext)

		if len(baseName) < n || len(baseName) > n && baseName[n] != '.' {
			continue
		}

		if match := rule.pattern.FindStringSubmatch(baseName[:n]); match != nil {
			return rule.name(match)
		}
	}

	return ""
}

// insta360StackName returns the left lens name for the lens and proxy files of an Insta360
// capture, keeping the case of each letter of the original prefix.
func insta360StackName(match []string) string {
	if len(match) != 6 {
		return ""
	}

	prefix, lens := match[1], match[4]

	switch strings.ToUpper(prefix) + "_" + lens {
	case "VID_00", "VID_10", "IMG_00", "IMG_10":
	case "LRV_11":
		prefix = matchCase("VID", prefix)
	default:
		return ""
	}

	return prefix + "_" + match[2] + "_" + match[3] + "_00_" + match[5]
}

// matchCase lowercases each ASCII letter of s where ref has a lowercase letter at the same position.
func matchCase(s, ref string) string {
	b := []byte(s)

	for i := range b {
		if i < len(ref) && ref[i] >= 'a' && ref[i] <= 'z' && b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}

	return string(b)
}
