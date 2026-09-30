package entity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFile_MarshalJSON(t *testing.T) {
	if m := FileFixtures.Pointer("Video.mp4"); m == nil {
		t.Fatal("must not be nil")
	} else if j, err := m.MarshalJSON(); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("json: %s", j)
	}
}

// TestFile_MarshalJSON_KeepStacked verifies that capture lens files are flagged to stay stacked.
func TestFile_MarshalJSON_KeepStacked(t *testing.T) {
	t.Run("CaptureLens", func(t *testing.T) {
		data, err := json.Marshal(&File{FileUID: "fs6sg6bw45bnlqde", FileName: "2022/VID_20220625_140410_10_008.insv"})
		require.NoError(t, err)
		assert.Contains(t, string(data), `"KeepStacked":true`)
		assert.Contains(t, string(data), `"StackGroup":"VID_20220625_140410_00_008"`)
	})
	t.Run("Other", func(t *testing.T) {
		data, err := json.Marshal(&File{FileUID: "fs6sg6bw45bnlqdf", FileName: "2022/VID_20220625_140410_10_008.insv.jpg"})
		require.NoError(t, err)
		assert.NotContains(t, string(data), "KeepStacked")
		assert.NotContains(t, string(data), "StackGroup")
	})
}
