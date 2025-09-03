package list

import "github.com/photoprism/photoprism/pkg/enum"

// StringLengthLimit specifies the maximum length of string return values.
var StringLengthLimit = 767

// Bool specifies boolean string values so they can be normalized.
var Bool = map[string]string{
	"true":    enum.True,
	"yes":     enum.True,
	"on":      enum.True,
	"enable":  enum.True,
	"false":   enum.False,
	"no":      enum.False,
	"off":     enum.False,
	"disable": enum.False,
}
