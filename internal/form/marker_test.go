package form

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestNewMarker(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var m = struct {
			SubjSrc       string
			MarkerName    string
			MarkerReview  bool
			MarkerInvalid bool
		}{
			SubjSrc:       "manual",
			MarkerName:    "Foo",
			MarkerReview:  true,
			MarkerInvalid: true,
		}

		f, err := NewMarker(m)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "manual", f.SubjSrc)
		assert.Equal(t, "Foo", f.MarkerName)
		assert.Equal(t, true, f.MarkerReview)
		assert.Equal(t, true, f.MarkerInvalid)
	})
}

func TestMarker_Validate(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		frm := Marker{}
		assert.Error(t, frm.Validate())
	})
	t.Run("False", func(t *testing.T) {
		frm := Marker{
			FileUID:       "frygcme3hc9re8nc",
			MarkerType:    "face",
			X:             0.303519,
			Y:             0.260742,
			W:             0.548387,
			H:             0.365234,
			SubjSrc:       "manual",
			MarkerName:    "Jens Mander",
			MarkerReview:  false,
			MarkerInvalid: false,
		}
		assert.Nil(t, frm.Validate())
	})
	t.Run("FileUID", func(t *testing.T) {
		frm := Marker{
			FileUID:       "rygcme3hc9re8nc",
			MarkerType:    "face",
			X:             0.303519,
			Y:             0.260742,
			W:             0.548387,
			H:             0.365234,
			SubjSrc:       "manual",
			MarkerName:    "Jens Mander",
			MarkerReview:  false,
			MarkerInvalid: false,
		}
		assert.Error(t, frm.Validate())
	})
	t.Run("Area", func(t *testing.T) {
		frm := Marker{
			FileUID:       "frygcme3hc9re8nc",
			MarkerType:    "face",
			X:             0.303519,
			Y:             1.260742,
			W:             0.548387,
			H:             0.365234,
			SubjSrc:       "manual",
			MarkerName:    "Jens Mander",
			MarkerReview:  false,
			MarkerInvalid: false,
		}
		assert.Error(t, frm.Validate())
	})
	t.Run("Name", func(t *testing.T) {
		frm := Marker{
			FileUID:       "frygcme3hc9re8nc",
			MarkerType:    "face",
			X:             0.303519,
			Y:             0.260742,
			W:             0.548387,
			H:             0.365234,
			SubjSrc:       "manual",
			MarkerName:    "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer...",
			MarkerReview:  false,
			MarkerInvalid: false,
		}
		assert.Error(t, frm.Validate())
	})
}

// TestMarker_UnmarshalJSON covers review aliases and partial updates.
func TestMarker_UnmarshalJSON(t *testing.T) {
	for _, tc := range []struct {
		name, body             string
		review, invalid, fails bool
	}{
		{"Approve", `{"Review":false,"Invalid":false}`, false, false, false},
		{"Reject", `{"Review":false,"Invalid":true}`, false, true, false},
		{"Alias", `{"MarkerReview":false}`, false, true, false},
		{"Omitted", `{}`, true, true, false},
		{"Null", `{"Review":null,"MarkerReview":null}`, true, true, false},
		{"NullCanonical", `{"Review":null,"MarkerReview":false}`, false, true, false},
		{"CanonicalTrue", `{"Review":true,"MarkerReview":false}`, true, true, false},
		{"Precedence", `{"Review":false,"MarkerReview":true}`, false, true, false},
		{"PrecedenceReversed", `{"MarkerReview":true,"Review":false}`, false, true, false},
		{"InvalidReview", `{"Review":"false"}`, true, true, true},
		{"InvalidAlias", `{"MarkerReview":0}`, true, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			frm := Marker{FileUID: "frygcme3hc9re8nc", MarkerReview: true, MarkerInvalid: true}
			err := json.Unmarshal([]byte(tc.body), &frm)
			if tc.fails {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.review, frm.MarkerReview)
			assert.Equal(t, tc.invalid, frm.MarkerInvalid)
			assert.Equal(t, "frygcme3hc9re8nc", frm.FileUID)
		})
	}
	t.Run("RoundTrip", func(t *testing.T) {
		encoded, err := json.Marshal(Marker{MarkerReview: true})
		require.NoError(t, err)
		assert.Contains(t, string(encoded), `"Review":true`)
		assert.NotContains(t, string(encoded), `"MarkerReview"`)
		var frm Marker
		require.NoError(t, json.Unmarshal(encoded, &frm))
		assert.True(t, frm.MarkerReview)
	})
}
