package constants

// Maps to represent Yes or No in the following languages Czech, Danish, Dutch, English, French, German, Indonesian, Italian, Polish, Portuguese, Russian, Ukrainian.
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
	"tak":      {},
	"sim":      {},
	"Да":       {},
	"ya":       {},
	"так":      {},
}

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
