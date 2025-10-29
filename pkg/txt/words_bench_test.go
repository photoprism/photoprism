package txt

import (
	"strings"
	"testing"
)

func makeLargeText(distinct, repeats int) string {
	// Seed a pool of mixed tokens: ASCII, unicode, hyphenated, apostrophes.
	base := []string{
		"alpha", "beta", "Gamma", "delta", "epsilon", "zeta", "eta", "theta",
		"iota", "kappa", "lambda", "mu", "nu", "xi", "omicron", "pi",
		"rho", "sigma", "tau", "upsilon", "phi", "chi", "psi", "omega",
		"New-York", "ma'am", "réseau", "Schäferhund", "Île", "Réunion",
		"Cote-d'Azur", "San_Francisco", "île-de-france", "vacation", "mountain",
		"beach", "sunset", "family", "holiday", "猫", "桥船", "ландшафт", "árvore",
		"52nd", "80s", "IMG20240101", "VID_2023-12-31", "IMG-20201231-WA1234",
	}
	if distinct > len(base) {
		distinct = len(base)
	}
	base = base[:distinct]

	var sb strings.Builder
	// Rough preallocation: average word ~6 chars + space
	sb.Grow(distinct * repeats * 8)
	for r := 0; r < repeats; r++ {
		for i, w := range base {
			if i%17 == 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(w)
			if i%13 == 0 {
				sb.WriteString(", ")
			} else {
				sb.WriteByte(' ')
			}
		}
		if r%10 == 0 {
			sb.WriteString(" and or WITH in AT ")
		}
	}
	return sb.String()
}

func BenchmarkWords_Large(b *testing.B) {
	s := makeLargeText(200, 200) // ~40k tokens mixed
	b.ReportAllocs()
	for b.Loop() {
		_ = Words(s)
	}
}

func BenchmarkUniqueKeywords_Large(b *testing.B) {
	s := makeLargeText(200, 200)
	b.ReportAllocs()
	for b.Loop() {
		_ = UniqueKeywords(s)
	}
}

func BenchmarkUniqueKeywords_ManyDup(b *testing.B) {
	s := makeLargeText(20, 2000) // many repeats, few distinct
	b.ReportAllocs()
	for b.Loop() {
		_ = UniqueKeywords(s)
	}
}
