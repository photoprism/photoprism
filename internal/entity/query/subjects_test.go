package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPeople(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		results, err := People()
		if assert.Nil(t, err) {
			assert.LessOrEqual(t, 3, len(results))
			t.Logf("people: %#v", results)
		}
	})
	t.Run("NotNil", func(t *testing.T) {
		t.Cleanup(func() {
			entity.Entities.Truncate(entity.Db())
			entity.CreateDefaultFixtures()
			entity.CreateTestFixtures()
			entity.File{}.RegenerateIndex()
		})
		// Clean the database as if it's brand new
		entity.Entities.Truncate(entity.Db())
		entity.CreateDefaultFixtures()
		entity.File{}.RegenerateIndex()

		results, err := People()

		if assert.Nil(t, err) {
			assert.NotNil(t, results)
			assert.Len(t, results, 0)
		}
	})
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
	subjects, linked, err := CreateMarkerSubjects()

	assert.NoError(t, err)
	assert.LessOrEqual(t, int64(0), subjects)
	assert.LessOrEqual(t, int64(0), linked)
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

// TestRemoveOrphanSubjects_Tombstones pins that the sweep still collects soft-deleted rows.
//
// This is the garbage collection for what MergeWith leaves: it moves every marker and face to the
// survivor and soft-deletes the source. Guarding the verified flag without excepting deleted rows
// stopped collecting any of them, and the earlier test exercised live rows only.
func TestRemoveOrphanSubjects_Tombstones(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Cleanup(entity.ResetTestFixtures)

	collected := func(t *testing.T, verified bool) bool {
		t.Helper()

		s := entity.NewSubject("ZZ Tombstone Sweep", entity.SubjPerson, entity.SrcMarker)
		require.NotNil(t, s)

		s.Verified = verified

		require.NoError(t, s.Create())
		require.NoError(t, s.Delete())

		t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Subject{}, "subj_uid = ?", s.SubjUID) })

		_, err := RemoveOrphanSubjects()
		require.NoError(t, err)

		var n int64
		require.NoError(t, entity.UnscopedDb().Model(&entity.Subject{}).
			Where("subj_uid = ?", s.SubjUID).Count(&n).Error)

		return n == 0
	}

	t.Run("Unverified", func(t *testing.T) {
		assert.True(t, collected(t, false))
	})
	t.Run("Verified", func(t *testing.T) {
		// The flag protects a person, not a row nothing can reach: a deleted one is collected too.
		assert.True(t, collected(t, true))
	})
}

// TestCreateMarkerSubjects_Sources pins that an XMP name is only linked to a person who exists and
// never names a cluster, while a manual name still finds or creates its person and names one.
func TestCreateMarkerSubjects_Sources(t *testing.T) {
	existing := conflictTestSubject(t, "Sources Existing Bob")
	deleted := conflictTestSubject(t, "Sources Deleted Eve")
	require.NoError(t, UnscopedDb().Model(deleted).UpdateColumn("deleted_at", entity.Now()).Error)
	pet := conflictTestSubject(t, "Sources Pet")
	require.NoError(t, UnscopedDb().Model(pet).UpdateColumn("subj_type", "pet").Error)

	t.Cleanup(func() {
		UnscopedDb().Delete(entity.Subject{}, "subj_name IN (?)", []string{"Sources Missing Carl", "Sources Manual Dana", "Sources Shared Finn"})
	})

	// Settles what the fixtures leave, so the count below is this test's alone.
	_, _, err := CreateMarkerSubjects()
	require.NoError(t, err)

	newMarker := func(t *testing.T, f *entity.Face, name, src string) string {
		t.Helper()

		m := entity.Marker{
			MarkerUID:    rnd.GenerateUID('m'),
			FileUID:      rnd.GenerateUID(entity.FileUID),
			MarkerType:   entity.MarkerFace,
			MarkerName:   name,
			MarkerReview: true,
			SubjSrc:      src,
			FaceID:       f.ID,
			W:            0.1,
			H:            0.1,
		}

		require.NoError(t, UnscopedDb().Create(&m).Error)
		t.Cleanup(func() { UnscopedDb().Delete(entity.Marker{}, "marker_uid = ?", m.MarkerUID) })

		return m.MarkerUID
	}

	xmpExisting := consensusTestFace(t, 21)
	xmpExistingMarker := newMarker(t, xmpExisting, "sources existing bob", entity.SrcXmp)
	xmpMissing := consensusTestFace(t, 22)
	xmpMissingMarker := newMarker(t, xmpMissing, "Sources Missing Carl", entity.SrcXmp)
	xmpDeleted := consensusTestFace(t, 23)
	xmpDeletedMarker := newMarker(t, xmpDeleted, deleted.SubjName, entity.SrcXmp)
	xmpPet := consensusTestFace(t, 25)
	xmpPetMarker := newMarker(t, xmpPet, pet.SubjName, entity.SrcXmp)
	manual := consensusTestFace(t, 24)
	manualMarker := newMarker(t, manual, "Sources Manual Dana", entity.SrcManual)
	sharedXmp := consensusTestFace(t, 26)
	sharedXmpMarker := newMarker(t, sharedXmp, "Sources Shared Finn", entity.SrcXmp)
	sharedManual := consensusTestFace(t, 27)
	newMarker(t, sharedManual, "Sources Shared Finn", entity.SrcManual)

	subjects, linked, err := CreateMarkerSubjects()
	require.NoError(t, err)

	t.Run("Counts", func(t *testing.T) {
		assert.Equal(t, int64(2), subjects, "two names resolved for manual markers")
		assert.Equal(t, int64(2), linked, "two XMP markers linked to existing people")
	})
	t.Run("XmpExistingPerson", func(t *testing.T) {
		m := entity.FindMarker(xmpExistingMarker)
		require.NotNil(t, m)
		assert.Equal(t, existing.SubjUID, m.SubjUID)
		assert.Equal(t, existing.SubjName, m.MarkerName, "the person's name is adopted")
		assert.Equal(t, entity.SrcXmp, m.SubjSrc)
		assert.False(t, m.MarkerReview)
		assert.Empty(t, entity.FindFace(xmpExisting.ID).SubjUID, "an XMP name never names a cluster")
	})
	t.Run("XmpMissingPerson", func(t *testing.T) {
		m := entity.FindMarker(xmpMissingMarker)
		require.NotNil(t, m)
		assert.Empty(t, m.SubjUID)
		assert.Nil(t, entity.FindSubjectByName("Sources Missing Carl", false), "an XMP name creates no person")
		assert.Empty(t, entity.FindFace(xmpMissing.ID).SubjUID)
	})
	t.Run("XmpDeletedPerson", func(t *testing.T) {
		m := entity.FindMarker(xmpDeletedMarker)
		require.NotNil(t, m)
		assert.Empty(t, m.SubjUID)
		assert.Empty(t, entity.FindFace(xmpDeleted.ID).SubjUID)

		var s entity.Subject
		require.NoError(t, UnscopedDb().Where("subj_uid = ?", deleted.SubjUID).First(&s).Error)
		assert.True(t, s.Deleted(), "and restores none")
	})
	t.Run("XmpNonPerson", func(t *testing.T) {
		m := entity.FindMarker(xmpPetMarker)
		require.NotNil(t, m)
		assert.Empty(t, m.SubjUID)
	})
	t.Run("ManualName", func(t *testing.T) {
		m := entity.FindMarker(manualMarker)
		require.NotNil(t, m)
		require.NotEmpty(t, m.SubjUID)
		assert.Equal(t, m.SubjUID, entity.FindFace(manual.ID).SubjUID, "a manual name names its unnamed cluster")
	})
	t.Run("SharedName", func(t *testing.T) {
		// The manual marker creates the person first, so the XMP marker links to it in the same run.
		m := entity.FindMarker(sharedXmpMarker)
		require.NotNil(t, m)
		require.NotEmpty(t, m.SubjUID)
		assert.Equal(t, m.SubjUID, entity.FindFace(sharedManual.ID).SubjUID)
		assert.Empty(t, entity.FindFace(sharedXmp.ID).SubjUID, "the XMP marker still names no cluster")
	})
}

// TestCreateMarkerSubjects_FormerName pins that an XMP name a person no longer carries is not linked
// to them after a rename.
func TestCreateMarkerSubjects_FormerName(t *testing.T) {
	suffix := rnd.Base36(6)
	former := "Former Xmp Anna " + suffix

	renamed := conflictTestSubject(t, former)
	_, err := renamed.UpdateName("Former Xmp Anne " + suffix)
	require.NoError(t, err)

	m := entity.Marker{
		MarkerUID:  rnd.GenerateUID('m'),
		FileUID:    rnd.GenerateUID(entity.FileUID),
		MarkerType: entity.MarkerFace,
		MarkerName: former,
		SubjSrc:    entity.SrcXmp,
		W:          0.1,
		H:          0.1,
	}

	require.NoError(t, UnscopedDb().Create(&m).Error)
	t.Cleanup(func() { UnscopedDb().Delete(entity.Marker{}, "marker_uid = ?", m.MarkerUID) })

	_, _, err = CreateMarkerSubjects()
	require.NoError(t, err)

	got := entity.FindMarker(m.MarkerUID)
	require.NotNil(t, got)
	assert.Empty(t, got.SubjUID)
	assert.Equal(t, former, got.MarkerName)
}

// TestCreateMarkerSubjects_CaseVariants pins that case variants of one name resolve one person, counted once.
func TestCreateMarkerSubjects_CaseVariants(t *testing.T) {
	t.Cleanup(func() { UnscopedDb().Delete(entity.Subject{}, "subj_name = ?", "Case Variant Gina") })

	// Settles what the fixtures leave, so the count below is this test's alone.
	_, _, err := CreateMarkerSubjects()
	require.NoError(t, err)

	uids := make([]string, 0, 2)

	for _, name := range []string{"Case Variant Gina", "case variant gina"} {
		m := entity.Marker{MarkerUID: rnd.GenerateUID('m'), FileUID: rnd.GenerateUID(entity.FileUID), MarkerType: entity.MarkerFace,
			MarkerName: name, SubjSrc: entity.SrcManual, W: 0.1, H: 0.1}
		require.NoError(t, UnscopedDb().Create(&m).Error)
		t.Cleanup(func() { UnscopedDb().Delete(entity.Marker{}, "marker_uid = ?", m.MarkerUID) })
		uids = append(uids, m.MarkerUID)
	}

	subjects, _, err := CreateMarkerSubjects()
	require.NoError(t, err)
	assert.Equal(t, int64(1), subjects)

	a, b := entity.FindMarker(uids[0]), entity.FindMarker(uids[1])
	require.NotNil(t, a)
	require.NotNil(t, b)
	assert.NotEmpty(t, a.SubjUID)
	assert.Equal(t, a.SubjUID, b.SubjUID)
}
