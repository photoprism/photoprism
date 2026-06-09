package video

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sunfish-shogi/bufseekio"
)

// sampleEntryHeader returns a valid 16-byte ISO BMFF visual sample entry header
// (box size, coding name, six reserved zero bytes, data_reference_index = 1) for
// use in SampleEntryOffset tests.
func sampleEntryHeader(code Chunk) []byte {
	b := make([]byte, sampleEntryHeaderLen)
	binary.BigEndian.PutUint32(b[0:4], minVisualSampleEntrySize)
	copy(b[4:8], code.Bytes())
	binary.BigEndian.PutUint16(b[14:16], 1)
	return b
}

// placeSampleEntry writes a valid visual sample entry into buf so that the
// coding name begins at codingNameOffset, mirroring the on-disk layout where
// the four-byte box size precedes the coding name.
func placeSampleEntry(buf []byte, codingNameOffset int, code Chunk) {
	copy(buf[codingNameOffset-4:], sampleEntryHeader(code))
}

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
	t.Run("Mp4vAvc1Mp4", func(t *testing.T) {
		index, err := ChunkFTYP.FileOffset("testdata/mp4v-avc1.mp4")
		require.NoError(t, err)
		assert.Equal(t, 4, index)
	})
	t.Run("IsomAvc1Mp4", func(t *testing.T) {
		index, err := ChunkFTYP.FileOffset("testdata/isom-avc1.mp4")
		require.NoError(t, err)
		assert.Equal(t, 4, index)
	})
	t.Run("ImageIsomAvc1Jpg", func(t *testing.T) {
		index, err := ChunkFTYP.FileOffset("testdata/image-isom-avc1.jpg")
		require.NoError(t, err)
		assert.Equal(t, 23213, index)
	})
	t.Run("MotionPhotoHeif", func(t *testing.T) {
		index, err := ChunkFTYP.FileOffset("testdata/motion-photo.heif")
		require.NoError(t, err)
		assert.Equal(t, 4, index)
		index, err = ChunkHEIC.FileOffset("testdata/motion-photo.heif")
		require.NoError(t, err)
		assert.Equal(t, 8, index)
		index, err = ChunkHVC1.FileOffset("testdata/motion-photo.heif")
		require.NoError(t, err)
		assert.Equal(t, 976016, index)
	})
}

func TestChunks(t *testing.T) {
	t.Run("Mp4vAvc1Mp4", func(t *testing.T) {
		f := openTestFile(t, "testdata/mp4v-avc1.mp4")
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
	t.Run("IsomAvc1Mp4", func(t *testing.T) {
		f := openTestFile(t, "testdata/isom-avc1.mp4")

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
	t.Run("ImageIsomAvc1Jpg", func(t *testing.T) {
		f := openTestFile(t, "testdata/image-isom-avc1.jpg")

		b := make([]byte, 12)

		// Read first 12 bytes from video file.
		if n, err := f.Read(b); err != nil {
			t.Fatal(err)
		} else if n != 12 {
			t.Fatalf("expected to read 12 bytes instead of %d", n)
		}

		assert.NotEqual(t, ChunkFTYP, [4]byte(b[4:8]))
		assert.NotEqual(t, ChunkISOM, [4]byte(b[8:12]))
	})
}

func TestChunks_DataOffset(t *testing.T) {
	t.Run("FirstMatchWins", func(t *testing.T) {
		f := openTestFile(t, "testdata/motion-photo.heif")
		// ChunkHVC1 lives at 976016; ChunkHEIC lives at 8. With both as needles
		// the earlier one (HEIC) must win the single-pass scan.
		pos, hit, err := Chunks{ChunkHVC1, ChunkHEIC}.DataOffset(f, 0, -1)
		require.NoError(t, err)
		assert.Equal(t, 8, pos)
		assert.Equal(t, ChunkHEIC, hit)
	})
	t.Run("SingleChunkSamePosition", func(t *testing.T) {
		f := openTestFile(t, "testdata/motion-photo.heif")
		pos, hit, err := Chunks{ChunkHVC1}.DataOffset(f, 0, -1)
		require.NoError(t, err)
		assert.Equal(t, 976016, pos)
		assert.Equal(t, ChunkHVC1, hit)
	})
	t.Run("MaxOffsetCapsScan", func(t *testing.T) {
		f := openTestFile(t, "testdata/motion-photo.heif")
		// HVC1 sits at 976016; a cap below that must short-circuit before reading it.
		pos, hit, err := Chunks{ChunkHVC1}.DataOffset(f, 0, 512*1024)
		require.NoError(t, err)
		assert.Equal(t, -1, pos)
		assert.Equal(t, Chunk{}, hit)
	})
	t.Run("NotFound", func(t *testing.T) {
		f := openTestFile(t, "testdata/mp4v-avc1.mp4")
		pos, hit, err := Chunks{ChunkHVC1, ChunkHEV1}.DataOffset(f, 0, -1)
		require.NoError(t, err)
		assert.Equal(t, -1, pos)
		assert.Equal(t, Chunk{}, hit)
	})
	t.Run("Empty", func(t *testing.T) {
		f := openTestFile(t, "testdata/mp4v-avc1.mp4")
		pos, hit, err := Chunks{}.DataOffset(f, 0, -1)
		require.NoError(t, err)
		assert.Equal(t, -1, pos)
		assert.Equal(t, Chunk{}, hit)
	})
	t.Run("NilFile", func(t *testing.T) {
		pos, hit, err := Chunks{ChunkHVC1}.DataOffset(nil, 0, -1)
		require.Error(t, err)
		assert.Equal(t, -1, pos)
		assert.Equal(t, Chunk{}, hit)
	})
}

func TestIsVisualSampleEntry(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.True(t, isVisualSampleEntry(sampleEntryHeader(ChunkM8RG)))
	})
	t.Run("TooShort", func(t *testing.T) {
		assert.False(t, isVisualSampleEntry(sampleEntryHeader(ChunkM8RG)[:15]))
	})
	t.Run("SizeTooSmall", func(t *testing.T) {
		b := sampleEntryHeader(ChunkM8RG)
		binary.BigEndian.PutUint32(b[0:4], minVisualSampleEntrySize-1)
		assert.False(t, isVisualSampleEntry(b))
	})
	t.Run("SizeTooLarge", func(t *testing.T) {
		b := sampleEntryHeader(ChunkM8RG)
		binary.BigEndian.PutUint32(b[0:4], maxVisualSampleEntrySize+1)
		assert.False(t, isVisualSampleEntry(b))
	})
	t.Run("ReservedNotZero", func(t *testing.T) {
		b := sampleEntryHeader(ChunkM8RG)
		b[10] = 0x01 // One of the six reserved bytes is nonzero.
		assert.False(t, isVisualSampleEntry(b))
	})
	t.Run("ZeroDataReferenceIndex", func(t *testing.T) {
		b := sampleEntryHeader(ChunkM8RG)
		binary.BigEndian.PutUint16(b[14:16], 0)
		assert.False(t, isVisualSampleEntry(b))
	})
}

func TestChunks_SampleEntryOffset(t *testing.T) {
	t.Run("RealMagicYuvFile", func(t *testing.T) {
		f := openTestFile(t, "testdata/magicyuv.mov")
		pos, hit, err := MagicYuvChunks.SampleEntryOffset(f, HeadScanLimit)
		require.NoError(t, err)
		assert.Equal(t, 3537, pos)
		assert.Equal(t, ChunkM8RG, hit)
	})
	t.Run("RealHevcFile", func(t *testing.T) {
		f := openTestFile(t, "testdata/quicktime-hvc1.mov")
		pos, hit, err := HevcChunks.SampleEntryOffset(f, HeadScanLimit)
		require.NoError(t, err)
		assert.Greater(t, pos, 0)
		assert.Equal(t, ChunkHVC1, hit)
	})
	t.Run("RejectsStrayCollision", func(t *testing.T) {
		// A four-byte MagicYUV code embedded in raw payload bytes, not framed as
		// a sample entry, must not be reported as a codec (issue #5617).
		buf := bytes.Repeat([]byte{0xAA}, 4096)
		copy(buf[1000:], ChunkM8Y4.Bytes())
		pos, hit, err := MagicYuvChunks.SampleEntryOffset(bytes.NewReader(buf), HeadScanLimit)
		require.NoError(t, err)
		assert.Equal(t, -1, pos)
		assert.Equal(t, Chunk{}, hit)
	})
	t.Run("AcceptsFramedEntry", func(t *testing.T) {
		buf := bytes.Repeat([]byte{0xAA}, 4096)
		placeSampleEntry(buf, 2000, ChunkM8Y2)
		pos, hit, err := MagicYuvChunks.SampleEntryOffset(bytes.NewReader(buf), HeadScanLimit)
		require.NoError(t, err)
		assert.Equal(t, 2000, pos)
		assert.Equal(t, ChunkM8Y2, hit)
	})
	t.Run("StrayBeforeRealEntry", func(t *testing.T) {
		// An earlier stray collision must not shadow a later valid sample entry;
		// the scan continues past invalid candidates.
		buf := bytes.Repeat([]byte{0xAA}, 4096)
		copy(buf[500:], ChunkM8Y4.Bytes())
		placeSampleEntry(buf, 2000, ChunkM8RG)
		pos, hit, err := MagicYuvChunks.SampleEntryOffset(bytes.NewReader(buf), HeadScanLimit)
		require.NoError(t, err)
		assert.Equal(t, 2000, pos)
		assert.Equal(t, ChunkM8RG, hit)
	})
	t.Run("BoundarySpanning", func(t *testing.T) {
		// A valid entry whose coding name straddles the internal 128 KiB block
		// boundary must still be found via the carry-over between reads.
		const codingNameOffset = 128*1024 - 1
		buf := bytes.Repeat([]byte{0xAA}, 200000)
		placeSampleEntry(buf, codingNameOffset, ChunkM8YA)
		pos, hit, err := MagicYuvChunks.SampleEntryOffset(bytes.NewReader(buf), HeadScanLimit)
		require.NoError(t, err)
		assert.Equal(t, codingNameOffset, pos)
		assert.Equal(t, ChunkM8YA, hit)
	})
	t.Run("MaxOffsetCapsScan", func(t *testing.T) {
		buf := bytes.Repeat([]byte{0xAA}, 200000)
		placeSampleEntry(buf, 150000, ChunkM8RG)
		pos, hit, err := MagicYuvChunks.SampleEntryOffset(bytes.NewReader(buf), 64*1024)
		require.NoError(t, err)
		assert.Equal(t, -1, pos)
		assert.Equal(t, Chunk{}, hit)
	})
	t.Run("NotFound", func(t *testing.T) {
		f := openTestFile(t, "testdata/mp4v-avc1.mp4")
		pos, hit, err := HevcChunks.SampleEntryOffset(f, HeadScanLimit)
		require.NoError(t, err)
		assert.Equal(t, -1, pos)
		assert.Equal(t, Chunk{}, hit)
	})
	t.Run("Empty", func(t *testing.T) {
		f := openTestFile(t, "testdata/mp4v-avc1.mp4")
		pos, hit, err := Chunks{}.SampleEntryOffset(f, HeadScanLimit)
		require.NoError(t, err)
		assert.Equal(t, -1, pos)
		assert.Equal(t, Chunk{}, hit)
	})
	t.Run("NilFile", func(t *testing.T) {
		pos, hit, err := MagicYuvChunks.SampleEntryOffset(nil, HeadScanLimit)
		require.Error(t, err)
		assert.Equal(t, -1, pos)
		assert.Equal(t, Chunk{}, hit)
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
