package meta

import (
	"testing"
)

// BenchmarkXMPRichSidecar measures per-file parse time on a fully
// populated Adobe Bridge sidecar (descriptive metadata, GPS, full
// camera/lens/exposure cluster, IDs). The proposal's performance
// budget is ≤ 2× the previous hand-rolled reader; running this with
// `go test -bench .` documents the new reader's absolute time so the
// budget can be reasoned about without re-introducing the old code.
func BenchmarkXMPRichSidecar(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var data Data
		if err := data.XMP("testdata/xmp/adobe/bridge.xmp"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkXMPMinimalSidecar measures the cheap path: a single-attribute
// fixture (F-Stop favorite). Sets a lower bound for how fast the loader
// + accessor pipeline can return when there is almost nothing to read.
func BenchmarkXMPMinimalSidecar(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var data Data
		if err := data.XMP("testdata/fstop-favorite.xmp"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkXMPMultiDescription measures the worst-case shape: four
// sibling rdf:Description blocks each declaring a different namespace
// binding. Catches regressions where the XPath walk degrades when
// properties scatter across blocks.
func BenchmarkXMPMultiDescription(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var data Data
		if err := data.XMP("testdata/xmp/synthetic/multi-rdf-description.xmp"); err != nil {
			b.Fatal(err)
		}
	}
}
