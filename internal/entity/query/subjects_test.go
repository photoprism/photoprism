package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPeople(t *testing.T) {
	if results, err := People(); err != nil {
		t.Fatal(err)
	} else {
		assert.LessOrEqual(t, 3, len(results))
		t.Logf("people: %#v", results)
	}
}

func TestPeopleCount(t *testing.T) {
	if result, err := PeopleCount(); err != nil {
		t.Fatal(err)
	} else {
		assert.LessOrEqual(t, 3, result)
		t.Logf("there are %d people", result)
	}
}

func TestSubjects(t *testing.T) {
	results, err := Subjects(3, 0)

	if err != nil {
		t.Fatal(err)
	}

	assert.GreaterOrEqual(t, len(results), 1)

	for _, val := range results {
		assert.IsType(t, entity.Subject{}, val)
	}
}

func TestSubjectMap(t *testing.T) {
	results, err := SubjectMap()

	if err != nil {
		t.Fatal(err)
	}

	assert.GreaterOrEqual(t, len(results), 1)

	for _, val := range results {
		assert.IsType(t, entity.Subject{}, val)
	}
}

func TestRemoveOrphanSubjects(t *testing.T) {
	affected, err := RemoveOrphanSubjects()

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(1), affected)
}

func TestCreateMarkerSubjects(t *testing.T) {
	affected, err := CreateMarkerSubjects()

	assert.NoError(t, err)
	assert.LessOrEqual(t, int64(0), affected)
}

// TestRemoveOrphanSubjects_Verified covers the flag that survives a face reset.
func TestRemoveOrphanSubjects_Verified(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Cleanup(entity.ResetTestFixtures)

	plain := entity.NewSubject("Reset Plain", entity.SubjPerson, entity.SrcMarker)
	require.NotNil(t, plain)
	require.NoError(t, plain.Create())

	vouched := entity.NewSubject("Reset Vouched", entity.SubjPerson, entity.SrcMarker)
	require.NotNil(t, vouched)
	vouched.Verified = true
	require.NoError(t, vouched.Create())

	t.Cleanup(func() {
		entity.UnscopedDb().Delete(&entity.Subject{}, "subj_uid IN (?)", []string{plain.SubjUID, vouched.SubjUID})
	})

	_, err := RemoveOrphanSubjects()
	require.NoError(t, err)

	assert.Nil(t, entity.FindSubject(plain.SubjUID), "an unreferenced marker subject is removed")

	kept := entity.FindSubject(vouched.SubjUID)
	require.NotNil(t, kept, "a verified person survives, so the name stays comparable across runs")
	assert.Equal(t, "Reset Vouched", kept.SubjName)
}
