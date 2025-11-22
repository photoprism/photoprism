package list

import (
	"fmt"
	"testing"
)

func makeStrings(prefix string, n int) []string {
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = fmt.Sprintf("%s_%06d", prefix, i)
	}
	return out
}

func shuffleEveryK(a []string, k int) []string {
	out := make([]string, len(a))
	copy(out, a)
	if k <= 1 {
		return out
	}
	for i := 0; i < len(out)-k; i += k {
		out[i], out[i+k-1] = out[i+k-1], out[i]
	}
	return out
}

func BenchmarkContainsAny_LargeOverlap(b *testing.B) {
	a := makeStrings("a", 5000)
	bList := makeStrings("b", 5000)
	// Introduce overlap: copy 20% of a into bList
	for i := 0; i < 1000; i++ {
		bList[i] = a[i*4]
	}
	b.ReportAllocs()
	for b.Loop() {
		if !ContainsAny(a, bList) {
			b.Fatalf("expected overlap")
		}
	}
}

func BenchmarkContainsAny_Disjoint(b *testing.B) {
	a := makeStrings("a", 5000)
	bList := makeStrings("b", 5000)
	b.ReportAllocs()
	for b.Loop() {
		if ContainsAny(a, bList) {
			b.Fatalf("expected disjoint")
		}
	}
}

func BenchmarkJoin_Large(b *testing.B) {
	a := makeStrings("x", 5000)
	j := append(makeStrings("y", 5000), a[:1000]...) // 1000 duplicates
	j = shuffleEveryK(j, 7)
	b.ReportAllocs()
	for b.Loop() {
		out := Join(a, j)
		if len(out) != 10000 {
			b.Fatalf("unexpected length: %d", len(out))
		}
	}
}
