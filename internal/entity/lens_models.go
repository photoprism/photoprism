package entity

import "regexp"

// LensModelIgnore is a regular expression that matches lens model substrings to be ignored.
var LensModelIgnore = regexp.MustCompile(`(?i)\sback.*camera\s`)
