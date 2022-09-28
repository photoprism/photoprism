package list

// StringLengthLimit specifies the maximum length of string return values.
var StringLengthLimit = 767

// True and False specify boolean string representations.
var (
	True  = "true"
	False = "false"
)

// Bool specifies boolean string values so they can be normalized.
var Bool = map[string]string{
	"true":    True,
	"yes":     True,
	"on":      True,
	"enable":  True,
	"false":   False,
	"no":      False,
	"off":     False,
	"disable": False,
}
