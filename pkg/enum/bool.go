package enum

// True and False specify boolean string representations.
const (
	True  = "true"
	False = "false"
)

// YesMap enumerates lower-case tokens we accept as an affirmative answer across
// supported languages (Czech, Danish, Dutch, English, French, German,
// Indonesian, Italian, Polish, Portuguese, Russian, Ukrainian). Callers should
// trim and lowercase input before performing lookups.
var YesMap = map[string]struct{}{
	"1":        {},
	"yes":      {},
	"include":  {},
	"true":     {},
	"positive": {},
	"please":   {},
	"ano":      {},
	"ja":       {},
	"oui":      {},
	"si":       {},
	"sí":       {},
	"tak":      {},
	"sim":      {},
	"да":       {},
	"ya":       {},
	"так":      {},
}

// NoMap enumerates lower-case tokens we accept as a negative answer across the
// same set of supported languages. Callers should trim and lowercase input
// before performing lookups.
var NoMap = map[string]struct{}{
	"0":        {},
	"no":       {},
	"none":     {},
	"exclude":  {},
	"false":    {},
	"negative": {},
	"unknown":  {},
	"žádný":    {},
	"ingen":    {},
	"nee":      {},
	"nein":     {},
	"non":      {},
	"nie":      {},
	"não":      {},
	"нет":      {},
	"tidak":    {},
	"ні":       {},
}
