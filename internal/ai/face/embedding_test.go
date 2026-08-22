package face

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbedding_Dist(t *testing.T) {
	t.Run("EqualLength", func(t *testing.T) {
		a := Embedding{0, 0, 0}
		b := Embedding{3, 4, 0}

		assert.InDelta(t, 5.0, a.Dist(b), 1e-9)
		assert.InDelta(t, 5.0, b.Dist(a), 1e-9)
	})
	t.Run("MismatchedLength", func(t *testing.T) {
		a := Embedding{0, 0}
		b := Embedding{1}

		assert.Equal(t, -1.0, a.Dist(b))
	})
}

func TestEmbedding_DistWithin(t *testing.T) {
	a := Embedding{0, 0, 0}
	b := Embedding{3, 4, 0}

	t.Run("Within", func(t *testing.T) {
		assert.InDelta(t, 5.0, a.DistWithin(b, 6), 1e-9)
	})
	t.Run("AtLimit", func(t *testing.T) {
		assert.InDelta(t, 5.0, a.DistWithin(b, 5), 1e-9)
	})
	t.Run("Beyond", func(t *testing.T) {
		assert.Equal(t, -1.0, a.DistWithin(b, 4.999))
	})
	t.Run("ZeroLimit", func(t *testing.T) {
		assert.InDelta(t, 0.0, a.DistWithin(a, 0), 1e-9)
		assert.Equal(t, -1.0, a.DistWithin(b, 0))
	})
	t.Run("NegativeLimit", func(t *testing.T) {
		assert.Equal(t, -1.0, a.DistWithin(b, -1))
	})
	t.Run("MismatchedLength", func(t *testing.T) {
		assert.Equal(t, -1.0, Embedding{0, 0}.DistWithin(Embedding{1}, 100))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, -1.0, a.DistWithin(Embedding{}, 100))
	})
	t.Run("NonFinite", func(t *testing.T) {
		// Dist reports NaN for these, which passes every threshold comparison it is fed to.
		// DistWithin reports "not comparable" instead, so a corrupt vector cannot win.
		nan := Embedding{math.NaN(), 0, 0}
		inf := Embedding{math.Inf(1), 0, 0}

		assert.Equal(t, -1.0, a.DistWithin(nan, 100))
		assert.Equal(t, -1.0, nan.DistWithin(a, 100))
		assert.Equal(t, -1.0, a.DistWithin(inf, 100))
	})
	t.Run("AbandonsMidVector", func(t *testing.T) {
		// A limit the running sum passes early is decided by the in-loop test rather than the
		// final one, which is the path the block stride exists for.
		x := make(Embedding, 512)
		y := make(Embedding, 512)
		for i := range x {
			x[i] = 1
		}

		assert.Equal(t, -1.0, x.DistWithin(y, 0.5))
		assert.InDelta(t, x.Dist(y), x.DistWithin(y, x.Dist(y)), 1e-9)
	})
	t.Run("AgreesWithDist", func(t *testing.T) {
		// Lengths that are not a multiple of the block matter because their tail is never
		// block-tested, so only the final verdict can reject them.
		for _, dims := range []int{1, 15, 16, 17, 128, 129, 512} {
			for i := range 32 {
				x := make(Embedding, dims)
				y := make(Embedding, dims)

				for j := range x {
					x[j] = float64((i*7+j*13)%97)/97 - 0.5
					y[j] = float64((i*11+j*5)%89)/89 - 0.5
				}

				full := x.Dist(y)

				assert.InDelta(t, full, x.DistWithin(y, full), 1e-9,
					"dims %d case %d must report the true distance at its own limit", dims, i)

				if full > 0 {
					assert.Equal(t, -1.0, x.DistWithin(y, math.Nextafter(full, 0)),
						"dims %d case %d must be refused just below the true distance", dims, i)
				}
			}
		}
	})
}

func TestNewEmbedding_Normalized(t *testing.T) {
	raw := []float32{3, 4}
	result := NewEmbedding(raw)

	var sum float64
	for _, v := range result {
		sum += v * v
	}

	assert.InDelta(t, 1.0, math.Sqrt(sum), 1e-9)
}

func TestNormalizeEmbedding(t *testing.T) {
	t.Run("Unnormalized", func(t *testing.T) {
		e := Embedding{3, 4}
		normalizeEmbedding(e)
		assert.InDelta(t, 0.6, e[0], 1e-6)
		assert.InDelta(t, 0.8, e[1], 1e-6)
	})
	t.Run("AlreadyNormalized", func(t *testing.T) {
		// Clustering unmarshals vectors that are already unit length and then hands the
		// same slices to EmbeddingsMidpoint, so the second pass has to leave them alone
		// rather than rewrite the caller's data.
		e := Embedding{0.6, 0.8}
		normalizeEmbedding(e)
		assert.Equal(t, Embedding{0.6, 0.8}, e)
	})
	t.Run("Zero", func(t *testing.T) {
		e := Embedding{0, 0}
		normalizeEmbedding(e)
		assert.Equal(t, Embedding{0, 0}, e)
	})
	t.Run("Empty", func(t *testing.T) {
		assert.NotPanics(t, func() { normalizeEmbedding(Embedding{}) })
	})
}

var benchmarkEmbeddingDist float64

// benchmarkEmbeddingPair returns deterministic embeddings for Dist benchmarks.
func benchmarkEmbeddingPair(dim int) (Embedding, Embedding) {
	a := make(Embedding, dim)
	other := make(Embedding, dim)

	scale := float64(dim)

	for i := 0; i < dim; i++ {
		a[i] = float64(i+1) / scale
		other[i] = float64(dim-i) / scale
	}

	return a, other
}

func BenchmarkEmbeddingDist(b *testing.B) {
	for _, dim := range []int{128, 512, 1024} {
		b.Run(fmt.Sprintf("Dim%d", dim), func(b *testing.B) {
			a, other := benchmarkEmbeddingPair(dim)

			b.SetBytes(int64(dim * 16))
			b.ReportAllocs()

			for b.Loop() {
				benchmarkEmbeddingDist = a.Dist(other)
			}
		})
	}
}

func TestUnmarshalEmbedding(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		emb, err := UnmarshalEmbedding("[-0.013,-0.031]")

		assert.NoError(t, err)
		assert.Len(t, emb, 2)

		var norm float64
		for _, v := range emb {
			norm += v * v
		}

		assert.InDelta(t, 1.0, math.Sqrt(norm), 1e-9)
		assert.Less(t, emb[0], 0.0)
		assert.Less(t, emb[1], 0.0)
	})
	t.Run("NoPrefix", func(t *testing.T) {
		emb, err := UnmarshalEmbedding("-0.013,-0.031]")

		assert.Error(t, err)
		assert.Nil(t, emb)
	})
	t.Run("InvalidValues", func(t *testing.T) {
		emb, err := UnmarshalEmbedding("[true, false]")

		assert.Error(t, err)
		assert.Equal(t, Embedding{0, 0}, emb)
	})
}
