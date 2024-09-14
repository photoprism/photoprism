package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountFileHashes(t *testing.T) {
	count := CountFileHashes()

	t.Logf("FILE HASH COUNT: %d", count)

	assert.LessOrEqual(t, 30, count)
}

func TestFileHashMap(t *testing.T) {
	result, err := FileHashMap()

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%d FILE HASHES: %#v", len(result), result)

	if len(result) < 3 {
		t.Fatalf("at least 3 file hashes expected")
	}
}

func TestThumbHashMap(t *testing.T) {
	result, err := ThumbHashMap()

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("THUMB HASHES: %#v", result)

	if len(result) < 1 {
		t.Fatalf("at least one thumb hashe expected")
	}
}
