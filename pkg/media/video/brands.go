package video

import (
	"errors"
	"os"

	"github.com/photoprism/photoprism/pkg/fs"
)

// ChunkFTYP specifies the start chunk of the ISO base media
// format, it must be followed by a valid subtype chunk.
// https://en.wikipedia.org/wiki/ISO_base_media_file_format
var (
	ChunkFTYP = Chunk{'f', 't', 'y', 'p'}
	ChunkQT   = Chunk{'q', 't', ' ', ' '}
	ChunkISOM = Chunk{'i', 's', 'o', 'm'}
	ChunkISO2 = Chunk{'i', 's', 'o', '2'}
	ChunkISO3 = Chunk{'i', 's', 'o', '3'}
	ChunkISO4 = Chunk{'i', 's', 'o', '4'}
	ChunkISO5 = Chunk{'i', 's', 'o', '5'}
	ChunkISO6 = Chunk{'i', 's', 'o', '6'}
	ChunkISO7 = Chunk{'i', 's', 'o', '7'}
	ChunkISO8 = Chunk{'i', 's', 'o', '8'}
	ChunkISO9 = Chunk{'i', 's', 'o', '9'}
	ChunkAVC1 = Chunk{'a', 'v', 'c', '1'}
	ChunkHEV1 = Chunk{'h', 'e', 'v', '1'}
	ChunkHVC1 = Chunk{'h', 'v', 'c', '1'}
	ChunkAV01 = Chunk{'a', 'v', '0', '1'}
	ChunkAV1C = Chunk{'a', 'v', '1', 'C'}
	ChunkMMP4 = Chunk{'m', 'm', 'p', '4'}
	ChunkMP4V = Chunk{'m', 'p', '4', 'v'}
	ChunkMP41 = Chunk{'m', 'p', '4', '1'}
	ChunkMP42 = Chunk{'m', 'p', '4', '2'}
	ChunkMP71 = Chunk{'m', 'p', '7', '1'}
	ChunkHEIC = Chunk{'h', 'e', 'i', 'c'}
)

// CompatibleBrands contains compatible subtypes chunks.
var CompatibleBrands = Chunks{
	ChunkQT,
	ChunkISOM,
	ChunkISO2,
	ChunkISO3,
	ChunkISO4,
	ChunkISO5,
	ChunkISO6,
	ChunkISO7,
	ChunkISO8,
	ChunkISO9,
	ChunkAVC1,
	ChunkHEV1,
	ChunkHVC1,
	ChunkAV01,
	ChunkAV1C,
	ChunkMMP4,
	ChunkMP4V,
	ChunkMP41,
	ChunkMP42,
	ChunkMP71,
}

// FileTypeOffset returns the file type start offset, or -1 if it was not found.
func FileTypeOffset(fileName string, brands Chunks) (int, error) {
	if !fs.FileExists(fileName) {
		return -1, errors.New("file not found")
	}

	file, err := os.Open(fileName)

	if err != nil {
		return -1, err
	}

	defer file.Close()

	index, err := brands.FileTypeOffset(file)

	return index, err
}
