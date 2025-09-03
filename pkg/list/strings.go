package list

import "github.com/photoprism/photoprism/pkg/constants"

// StringLengthLimit specifies the maximum length of string return values.
var StringLengthLimit = 767

// Bool specifies boolean string values so they can be normalized.
var Bool = map[string]string{
	"true":    constants.True,
	"yes":     constants.True,
	"on":      constants.True,
	"enable":  constants.True,
	"false":   constants.False,
	"no":      constants.False,
	"off":     constants.False,
	"disable": constants.False,
}
