package entity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarker_MarshalJSON(t *testing.T) {
	if m := MarkerFixtures.Pointer("actor-a-2"); m == nil {
		t.Fatal("must not be nil")
	} else if j, err := m.MarshalJSON(); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("json: %s", j)
	}
}

// TestMarker_MarshalJSON_Unlinked pins that serializing a marker whose name is not linked to a person
// creates, restores and links none, while the name is still shown.
func TestMarker_MarshalJSON_Unlinked(t *testing.T) {
	gone := NewSubject("Json Unlinked Gone", SubjPerson, SrcManual)
	require.NotNil(t, gone)
	require.NoError(t, gone.Create())
	require.NoError(t, UnscopedDb().Model(gone).UpdateColumn("deleted_at", Now()).Error)
	known := NewSubject("Json Unlinked Known", SubjPerson, SrcManual)
	require.NotNil(t, known)
	require.NoError(t, known.Create())
	hidden := NewSubject("Json Unlinked Hidden", SubjPerson, SrcManual)
	require.NotNil(t, hidden)
	require.NoError(t, hidden.Create())
	require.NoError(t, hidden.Update("SubjHidden", true))
	renamed := NewSubject("Json Unlinked Former", SubjPerson, SrcManual)
	require.NotNil(t, renamed)
	require.NoError(t, renamed.Create())
	_, err := renamed.UpdateName("Json Unlinked Current")
	require.NoError(t, err)

	t.Cleanup(func() {
		UnscopedDb().Delete(&Subject{}, "subj_name IN (?)", []string{"Json Unlinked Gone", "Json Unlinked Known", "Json Unlinked Missing", "Json Unlinked Hidden", "Json Unlinked Current"})
	})

	decode := func(t *testing.T, m *Marker) (name, subjUID string) {
		t.Helper()

		data, err := json.Marshal(m)
		require.NoError(t, err)

		var out struct {
			Name    string
			SubjUID string
		}

		require.NoError(t, json.Unmarshal(data, &out))

		return out.Name, out.SubjUID
	}

	t.Run("Linked", func(t *testing.T) {
		m := &Marker{MarkerType: MarkerFace, SubjSrc: SrcAuto, SubjUID: known.SubjUID, MarkerName: "Stored Spelling"}
		name, subjUID := decode(t, m)
		assert.Equal(t, known.SubjName, name, "a linked marker shows its person's name")
		assert.Equal(t, known.SubjUID, subjUID)
	})

	for _, src := range []string{SrcXmp, SrcManual} {
		t.Run(src, func(t *testing.T) {
			t.Run("Missing", func(t *testing.T) {
				m := &Marker{MarkerType: MarkerFace, SubjSrc: src, MarkerName: "Json Unlinked Missing"}
				name, subjUID := decode(t, m)
				assert.Equal(t, "Json Unlinked Missing", name)
				assert.Empty(t, subjUID)
				assert.Empty(t, m.SubjUID)
				assert.Nil(t, FindSubjectByName("Json Unlinked Missing", false), "a read creates no person")
			})
			t.Run("Deleted", func(t *testing.T) {
				m := &Marker{MarkerType: MarkerFace, SubjSrc: src, MarkerName: "json unlinked gone"}
				name, subjUID := decode(t, m)
				assert.Equal(t, "json unlinked gone", name, "a deleted person's name is not shown")
				assert.Empty(t, subjUID)

				var s Subject
				require.NoError(t, UnscopedDb().Where("subj_uid = ?", gone.SubjUID).First(&s).Error)
				assert.True(t, s.Deleted(), "a read restores no person")
			})
			t.Run("Renamed", func(t *testing.T) {
				m := &Marker{MarkerType: MarkerFace, SubjSrc: src, MarkerName: "Json Unlinked Former"}
				name, subjUID := decode(t, m)
				assert.Equal(t, "Json Unlinked Former", name, "a former name shows as stored")
				assert.Empty(t, subjUID)
			})
			t.Run("Withheld", func(t *testing.T) {
				m := &Marker{MarkerType: MarkerFace, SubjSrc: src, MarkerName: "json unlinked hidden"}
				name, _ := decode(t, m)
				assert.Equal(t, "json unlinked hidden", name, "a withheld person's spelling is not shown")
			})
			t.Run("Existing", func(t *testing.T) {
				m := &Marker{MarkerType: MarkerFace, SubjSrc: src, MarkerName: "json unlinked known"}
				name, subjUID := decode(t, m)
				assert.Equal(t, known.SubjName, name, "the person's name is shown")
				assert.Empty(t, subjUID, "the stored link is reported as stored")
				assert.Empty(t, m.SubjUID)
			})
		})
	}
}
