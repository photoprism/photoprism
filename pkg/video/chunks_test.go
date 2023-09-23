package video

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sunfish-shogi/bufseekio"
)

func TestChunk_TypeCast(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		assert.Equal(t, "ftyp", ChunkFTYP.String())
	})
	t.Run("Hex", func(t *testing.T) {
		assert.Equal(t, "0x66747970", ChunkFTYP.Hex())
	})
	t.Run("Uint32", func(t *testing.T) {
		assert.Equal(t, uint32(0x66747970), ChunkFTYP.Uint32())
	})
}

func TestChunk_FileOffset(t *testing.T) {
	t.Run("mp4v-avc1.mp4", func(t *testing.T) {
		index, err := ChunkFTYP.FileOffset("testdata/mp4v-avc1.mp4")
		require.NoError(t, err)
		assert.Equal(t, 4, index)
	})
	t.Run("isom-avc1.mp4", func(t *testing.T) {
		index, err := ChunkFTYP.FileOffset("testdata/isom-avc1.mp4")
		require.NoError(t, err)
		assert.Equal(t, 4, index)
	})
	t.Run("image-isom-avc1.jpg", func(t *testing.T) {
		index, err := ChunkFTYP.FileOffset("testdata/image-isom-avc1.jpg")
		require.NoError(t, err)
		assert.Equal(t, 23213, index)
	})
}

func TestChunks(t *testing.T) {
	t.Run("mp4v-avc1.mp4", func(t *testing.T) {
		f, fileErr := os.Open("testdata/mp4v-avc1.mp4")
		require.NoError(t, fileErr)
		defer f.Close()
		r := bufseekio.NewReadSeeker(f, 8, 4)

		var startChunk = make([]byte, 4)
		var subType = make([]byte, 4)

		if _, err := r.Seek(4, io.SeekStart); err != nil {
			t.Fatal(err)
		}

		// Read first 4-byte chunk.
		if n, err := r.Read(startChunk); err != nil {
			t.Fatal(err)
		} else if n != 4 {
			t.Fatal("expected to read 4 bytes")
		}

		// Read second 4-byte chunk.
		if n, err := r.Read(subType); err != nil {
			t.Fatal(err)
		} else if n != 4 {
			t.Fatal("expected to read 4 bytes")
		}

		assert.Equal(t, ChunkFTYP.Bytes(), startChunk[:4])
		assert.Equal(t, ChunkMP4V.Bytes(), subType[:4])
	})
	t.Run("isom-avc1.mp4", func(t *testing.T) {
		f, fileErr := os.Open("testdata/isom-avc1.mp4")
		require.NoError(t, fileErr)
		defer f.Close()

		b := make([]byte, 12)

		// Read first 12 bytes from video file.
		if n, err := f.Read(b); err != nil {
			t.Fatal(err)
		} else if n != 12 {
			t.Fatalf("expected to read 12 bytes instead of %d", n)
		}

		assert.Equal(t, ChunkFTYP[:], b[4:8])
		assert.Equal(t, ChunkISOM[:], b[8:12])
	})
	t.Run("image-isom-avc1.jpg", func(t *testing.T) {
		f, fileErr := os.Open("testdata/image-isom-avc1.jpg")
		require.NoError(t, fileErr)
		defer f.Close()

		b := make([]byte, 12)

		// Read first 12 bytes from video file.
		if n, err := f.Read(b); err != nil {
			t.Fatal(err)
		} else if n != 12 {
			t.Fatalf("expected to read 12 bytes instead of %d", n)
		}

		assert.NotEqual(t, ChunkFTYP, *(*[4]byte)(b[4:8]))
		assert.NotEqual(t, ChunkISOM, *(*[4]byte)(b[8:12]))
	})
}

func TestChunks_Contains(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		assert.True(t, CompatibleBrands.Contains(ChunkMP41))
	})
	t.Run("NotFound", func(t *testing.T) {
		assert.False(t, CompatibleBrands.Contains(ChunkFTYP))
	})
}

func TestChunks_ContainsAny(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		chunks := [][4]byte{ChunkMP41, ChunkMP42}
		assert.True(t, CompatibleBrands.ContainsAny(chunks))
	})
	t.Run("NotFound", func(t *testing.T) {
		chunks := [][4]byte{ChunkFTYP}
		assert.False(t, CompatibleBrands.ContainsAny(chunks))
	})
}
