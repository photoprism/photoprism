package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteOrphanPeople(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		if count, err := DeleteOrphanPeople(); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("deleted %d faces", count)
		}
	})
}

// TestOrphanPeople_Verified covers the flag that keeps a person whose markers a reset removed.
//
// Re-clustering leaves a verified person unreferenced by design, and the row is what makes the same
// name comparable across runs instead of retyped after each one.
func TestOrphanPeople_Verified(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Cleanup(ResetTestFixtures)

	plain := NewSubject("Orphan Plain", SubjPerson, SrcMarker)
	require.NotNil(t, plain)
	require.NoError(t, plain.Create())

	vouched := NewSubject("Orphan Vouched", SubjPerson, SrcMarker)
	require.NotNil(t, vouched)
	vouched.Verified = true
	require.NoError(t, vouched.Create())

	t.Cleanup(func() {
		UnscopedDb().Delete(&Subject{}, "subj_uid IN (?)", []string{plain.SubjUID, vouched.SubjUID})
	})

	orphans, err := OrphanPeople()
	require.NoError(t, err)

	uids := make(map[string]bool, len(orphans))
	for _, o := range orphans {
		uids[o.SubjUID] = true
	}

	assert.True(t, uids[plain.SubjUID], "an unreferenced person is an orphan")
	assert.False(t, uids[vouched.SubjUID], "unless somebody vouched for the name")
}
