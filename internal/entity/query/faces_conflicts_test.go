package query

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
)

// conflictTestSubject stores a person the conflict tests can name, and removes it again so the
// shared fixtures other tests count on are unchanged.
func conflictTestSubject(t *testing.T, name string) *entity.Subject {
	t.Helper()

	s := entity.NewSubject(name, entity.SubjPerson, entity.SrcManual)
	require.NotNil(t, s)
	require.NoError(t, s.Create())

	t.Cleanup(func() {
		UnscopedDb().Delete(entity.Subject{}, "subj_uid = ?", s.SubjUID)
	})

	return s
}

// conflictTestFace stores a cluster whose vector leans on one axis, so two of them match while
// hashing to distinct ids. Built rather than hard-coded, because a literal vector belongs to one
// embedding model and would be ineligible under any other, testing only the early exit.
func conflictTestFace(t *testing.T, subjUID string, axis int, tilt float64) *entity.Face {
	t.Helper()

	dims := face.ExpectedDims()
	require.Greater(t, dims, axis)

	v := make(face.Embedding, dims)
	v[0] = 1
	v[axis] = tilt

	sum := 0.0

	for _, x := range v {
		sum += x * x
	}

	for i := range v {
		v[i] /= math.Sqrt(sum)
	}

	f := entity.NewFace(subjUID, entity.SrcManual, face.Embeddings{v}, face.EmbeddingModelName())
	require.NotNil(t, f)
	require.NotEmpty(t, f.ID)
	require.NoError(t, f.Create())

	t.Cleanup(func() {
		UnscopedDb().Delete(entity.Face{}, "id = ?", f.ID)
	})

	return f
}

// findConflict returns the reported pair for two clusters, in whichever order the walk found it.
func findConflict(conflicts []FaceConflict, a, b string) *FaceConflict {
	for i := range conflicts {
		if conflicts[i].ID == a && conflicts[i].OtherID == b {
			return &conflicts[i]
		} else if conflicts[i].ID == b && conflicts[i].OtherID == a {
			return &conflicts[i]
		}
	}

	return nil
}

func TestFaceConflicts(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		alice := conflictTestSubject(t, "Conflict Alice")
		bob := conflictTestSubject(t, "Conflict Bob")
		f1 := conflictTestFace(t, alice.SubjUID, 1, 0.05)
		f2 := conflictTestFace(t, bob.SubjUID, 2, 0.05)
		conflicts, scan, err := FaceConflicts(alice.SubjUID, 1000, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, scan.Clusters)
		assert.Positive(t, scan.Compared)
		found := findConflict(conflicts, f1.ID, f2.ID)
		require.NotNil(t, found, "the two clusters must be reported as a conflict")
		assert.Equal(t, "Conflict Alice", found.SubjName)
		assert.Equal(t, "Conflict Bob", found.OtherSubjName)
		assert.Positive(t, found.Dist)
		assert.Positive(t, found.Accept)
		assert.Positive(t, found.OtherAccept)
	})
	t.Run("ReportedFromTheSideThatAccepts", func(t *testing.T) {
		// Match gates on the receiver's own collision radius, so a cluster an earlier resolution
		// narrowed refuses a pair its counterparty still accepts. Testing one direction dropped it.
		a := conflictTestSubject(t, "Conflict Refuses")
		b := conflictTestSubject(t, "Conflict Accepts")
		f1 := conflictTestFace(t, a.SubjUID, 1, 0.05)
		f2 := conflictTestFace(t, b.SubjUID, 2, 0.05)
		// Above CollisionDist so it is enforced, and below the distance between the two.
		require.NoError(t, UnscopedDb().Model(&entity.Face{}).
			Where("id = ?", f1.ID).UpdateColumn("collision_radius", face.CollisionDist+0.01).Error)
		conflicts, _, err := FaceConflicts(a.SubjUID, 1000, 0)
		require.NoError(t, err)
		found := findConflict(conflicts, f1.ID, f2.ID)
		require.NotNil(t, found, "the pair must be reported from the cluster that accepts it")
		assert.Equal(t, f2.ID, found.ID, "reported from the accepting side")
		assert.Equal(t, "Conflict Accepts", found.SubjName)
	})
	t.Run("FoundWithoutAPersonFilter", func(t *testing.T) {
		// The unscoped walk is the path a person filter cannot cover, so a positive case has to
		// exist for it: every other one forces the named cluster to be compared first.
		a := conflictTestSubject(t, "Conflict Unscoped A")
		b := conflictTestSubject(t, "Conflict Unscoped B")
		f1 := conflictTestFace(t, a.SubjUID, 1, 0.05)
		f2 := conflictTestFace(t, b.SubjUID, 2, 0.05)
		conflicts, scan, err := FaceConflicts("", 1000, 0)
		require.NoError(t, err)
		assert.Positive(t, scan.Clusters)
		assert.NotNil(t, findConflict(conflicts, f1.ID, f2.ID))
	})
	t.Run("SamePersonIsNotAConflict", func(t *testing.T) {
		alice := conflictTestSubject(t, "Conflict Same")
		f1 := conflictTestFace(t, alice.SubjUID, 1, 0.05)
		f2 := conflictTestFace(t, alice.SubjUID, 2, 0.05)
		conflicts, _, err := FaceConflicts(alice.SubjUID, 1000, 0)
		require.NoError(t, err)
		assert.Nil(t, findConflict(conflicts, f1.ID, f2.ID), "one person's own clusters must not conflict")
	})
	t.Run("NameMatchesNobody", func(t *testing.T) {
		conflicts, scan, err := FaceConflicts("Nobody Is Named This", 1000, 0)
		require.NoError(t, err)
		assert.Empty(t, conflicts)
		assert.Zero(t, scan.Clusters)
		assert.Zero(t, scan.Compared)
	})
	t.Run("FoundByName", func(t *testing.T) {
		alice := conflictTestSubject(t, "Conflict Named")
		bob := conflictTestSubject(t, "Conflict Other")
		f1 := conflictTestFace(t, alice.SubjUID, 1, 0.05)
		f2 := conflictTestFace(t, bob.SubjUID, 2, 0.05)
		conflicts, scan, err := FaceConflicts("Conflict Named", 1000, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, scan.Clusters)
		assert.NotNil(t, findConflict(conflicts, f1.ID, f2.ID))
	})
	t.Run("OffsetPastTheEnd", func(t *testing.T) {
		alice := conflictTestSubject(t, "Conflict Paged")
		bob := conflictTestSubject(t, "Conflict Paged Other")
		conflictTestFace(t, alice.SubjUID, 1, 0.05)
		conflictTestFace(t, bob.SubjUID, 2, 0.05)
		conflicts, scan, err := FaceConflicts(alice.SubjUID, 1000, 1000)
		require.NoError(t, err)
		assert.Empty(t, conflicts)
		// The scan still ran; only the page is empty, which is what tells the two apart.
		assert.Equal(t, 1, scan.Clusters)
		assert.Positive(t, scan.Compared)
	})
}

func TestFaceConflict_Ambiguous(t *testing.T) {
	t.Run("BelowAmbiguityDist", func(t *testing.T) {
		assert.True(t, FaceConflict{Dist: face.AmbiguityDist() / 2}.Ambiguous())
	})
	t.Run("AtAmbiguityDist", func(t *testing.T) {
		assert.False(t, FaceConflict{Dist: face.AmbiguityDist()}.Ambiguous())
	})
	t.Run("Above", func(t *testing.T) {
		assert.False(t, FaceConflict{Dist: face.AmbiguityDist() * 10}.Ambiguous())
	})
	t.Run("Unmeasured", func(t *testing.T) {
		// Embeddings.Dist reports -1 when nothing is comparable, which must not read as the
		// closest possible pair.
		assert.False(t, FaceConflict{Dist: -1}.Ambiguous())
	})
}

func TestFaceConflict_Narrows(t *testing.T) {
	t.Run("AboveTheEnforcedFloor", func(t *testing.T) {
		assert.True(t, FaceConflict{Dist: face.CollisionDist + face.Epsilon + 0.001}.Narrows())
	})
	t.Run("AtTheFloorRecordsAnInertRadius", func(t *testing.T) {
		// dist - Epsilon lands exactly on CollisionDist, which Face.Match does not enforce.
		assert.False(t, FaceConflict{Dist: face.CollisionDist + face.Epsilon}.Narrows())
	})
	t.Run("BetweenAmbiguousAndTheFloor", func(t *testing.T) {
		assert.False(t, FaceConflict{Dist: face.AmbiguityDist() + 0.001}.Narrows())
	})
	t.Run("FollowsTheConfiguredFloor", func(t *testing.T) {
		// CollisionDist is set from PHOTOPRISM_FACE_COLLISION_DIST at startup, so the cutoff has
		// to be read per call rather than captured.
		restore := face.CollisionDist
		t.Cleanup(func() { face.CollisionDist = restore })
		c := FaceConflict{Dist: restore + face.Epsilon + 0.001}
		assert.True(t, c.Narrows())
		face.CollisionDist = restore + 0.1
		assert.False(t, c.Narrows())
	})
}

func TestConflictScope(t *testing.T) {
	faces := FaceMap{
		"A": entity.Face{ID: "A", SubjUID: "js6sg6b1qekk9jx8"},
		"B": entity.Face{ID: "B", SubjUID: "js6sg6b1h1njaaab"},
		"C": entity.Face{ID: "C"},
	}
	ids := IDs{"A", "B", "C"}
	t.Run("NoPersonKeepsEverything", func(t *testing.T) {
		scope, err := conflictScope("", faces, ids)
		require.NoError(t, err)
		assert.Equal(t, ids, scope)
	})
	t.Run("SubjectUID", func(t *testing.T) {
		scope, err := conflictScope("js6sg6b1qekk9jx8", faces, ids)
		require.NoError(t, err)
		assert.Equal(t, IDs{"A"}, scope)
	})
	t.Run("UnknownSubjectSelectsNothing", func(t *testing.T) {
		scope, err := conflictScope("js6sg6b1h1njaaac", faces, ids)
		require.NoError(t, err)
		assert.Empty(t, scope)
	})
	t.Run("Name", func(t *testing.T) {
		s := conflictTestSubject(t, "Scope Test Person")
		scoped := FaceMap{"D": entity.Face{ID: "D", SubjUID: s.SubjUID}}
		scope, err := conflictScope("Scope Test Person", scoped, IDs{"D"})
		require.NoError(t, err)
		assert.Equal(t, IDs{"D"}, scope)
	})
}

func TestSortFaceConflicts(t *testing.T) {
	t.Run("ClosestFirst", func(t *testing.T) {
		conflicts := []FaceConflict{{ID: "B", Dist: 0.9}, {ID: "A", Dist: 0.1}}
		sortFaceConflicts(conflicts)
		assert.Equal(t, "A", conflicts[0].ID)
	})
	t.Run("TiesBreakOnID", func(t *testing.T) {
		conflicts := []FaceConflict{{ID: "B", OtherID: "X", Dist: 0.5}, {ID: "A", OtherID: "Y", Dist: 0.5}}
		sortFaceConflicts(conflicts)
		assert.Equal(t, "A", conflicts[0].ID)
	})
	t.Run("TiesBreakOnOtherID", func(t *testing.T) {
		conflicts := []FaceConflict{{ID: "A", OtherID: "Y", Dist: 0.5}, {ID: "A", OtherID: "X", Dist: 0.5}}
		sortFaceConflicts(conflicts)
		assert.Equal(t, "X", conflicts[0].OtherID)
	})
	t.Run("Empty", func(t *testing.T) {
		assert.NotPanics(t, func() { sortFaceConflicts(nil) })
	})
}

func TestPageFaceConflicts(t *testing.T) {
	conflicts := []FaceConflict{{ID: "A"}, {ID: "B"}, {ID: "C"}}
	t.Run("FirstPage", func(t *testing.T) {
		assert.Len(t, pageFaceConflicts(conflicts, 2, 0), 2)
	})
	t.Run("SecondPage", func(t *testing.T) {
		page := pageFaceConflicts(conflicts, 2, 2)
		require.Len(t, page, 1)
		assert.Equal(t, "C", page[0].ID)
	})
	t.Run("OffsetPastTheEnd", func(t *testing.T) {
		assert.Empty(t, pageFaceConflicts(conflicts, 2, 99))
	})
	t.Run("NegativeOffset", func(t *testing.T) {
		assert.Len(t, pageFaceConflicts(conflicts, 3, -5), 3)
	})
	t.Run("InvalidCount", func(t *testing.T) {
		assert.Empty(t, pageFaceConflicts(conflicts, 0, 0))
	})
}

func TestFaceConflictReportNotes(t *testing.T) {
	t.Run("CountsAHiddenCluster", func(t *testing.T) {
		s := conflictTestSubject(t, "Notes Hidden Person")
		f := conflictTestFace(t, s.SubjUID, 1, 0.05)
		before, err := FaceConflictReportNotes()
		require.NoError(t, err)
		require.NoError(t, UnscopedDb().Model(&entity.Face{}).
			Where("id = ?", f.ID).UpdateColumn("face_hidden", true).Error)
		after, err := FaceConflictReportNotes()
		require.NoError(t, err)
		assert.Equal(t, before.Hidden+1, after.Hidden)
	})
	t.Run("CountsAnInertRadius", func(t *testing.T) {
		s := conflictTestSubject(t, "Notes Inert Person")
		f := conflictTestFace(t, s.SubjUID, 1, 0.05)
		before, err := FaceConflictReportNotes()
		require.NoError(t, err)
		// At the floor exactly, which is the boundary the matcher still ignores.
		require.NoError(t, UnscopedDb().Model(&entity.Face{}).
			Where("id = ?", f.ID).UpdateColumn("collision_radius", face.CollisionDist).Error)
		after, err := FaceConflictReportNotes()
		require.NoError(t, err)
		assert.Equal(t, before.InertRadius+1, after.InertRadius)
		// An inert radius is not a narrowing, so it must not also be reported as one.
		assert.Equal(t, before.BelowOwnSpread, after.BelowOwnSpread)
	})
	t.Run("SkipsClustersTheWalkExcludes", func(t *testing.T) {
		// A hidden cluster is never compared, so counting its radius would describe a row the
		// same notes have just said was skipped.
		s := conflictTestSubject(t, "Notes Excluded Person")
		f := conflictTestFace(t, s.SubjUID, 1, 0.05)
		before, err := FaceConflictReportNotes()
		require.NoError(t, err)
		require.NoError(t, UnscopedDb().Model(&entity.Face{}).Where("id = ?", f.ID).
			UpdateColumns(entity.Values{"face_hidden": true, "collision_radius": face.CollisionDist / 2}).Error)
		after, err := FaceConflictReportNotes()
		require.NoError(t, err)
		assert.Equal(t, before.InertRadius, after.InertRadius)
		assert.Equal(t, before.BelowOwnSpread, after.BelowOwnSpread)
	})
	t.Run("CountsARetiredCluster", func(t *testing.T) {
		s := conflictTestSubject(t, "Notes Test Person")
		f := conflictTestFace(t, s.SubjUID, 1, 0.05)
		before, err := FaceConflictReportNotes()
		require.NoError(t, err)
		require.NoError(t, UnscopedDb().Model(&entity.Face{}).
			Where("id = ?", f.ID).
			UpdateColumn("face_kind", int(face.AmbiguousFace)).Error)
		after, err := FaceConflictReportNotes()
		require.NoError(t, err)
		assert.Equal(t, before.Ambiguous+1, after.Ambiguous)
	})
	t.Run("CountsARadiusInsideItsOwnSpread", func(t *testing.T) {
		s := conflictTestSubject(t, "Notes Spread Person")
		f := conflictTestFace(t, s.SubjUID, 1, 0.05)
		before, err := FaceConflictReportNotes()
		require.NoError(t, err)
		require.Positive(t, f.SampleRadius)
		require.NoError(t, UnscopedDb().Model(&entity.Face{}).
			Where("id = ?", f.ID).
			UpdateColumn("collision_radius", f.SampleRadius/2).Error)
		after, err := FaceConflictReportNotes()
		require.NoError(t, err)
		assert.Equal(t, before.BelowOwnSpread+1, after.BelowOwnSpread)
	})
}

func TestFaceConflictNames(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := conflictTestSubject(t, "Name Lookup Person")
		conflicts := []FaceConflict{{SubjUID: s.SubjUID, OtherSubjUID: s.SubjUID}}
		require.NoError(t, faceConflictNames(conflicts))
		assert.Equal(t, "Name Lookup Person", conflicts[0].SubjName)
		assert.Equal(t, "Name Lookup Person", conflicts[0].OtherSubjName)
	})
	t.Run("UnknownSubjectLeavesTheNameEmpty", func(t *testing.T) {
		conflicts := []FaceConflict{{SubjUID: "js6sg6b1h1njzzzz"}}
		require.NoError(t, faceConflictNames(conflicts))
		assert.Empty(t, conflicts[0].SubjName)
	})
	t.Run("NothingToLookUp", func(t *testing.T) {
		assert.NoError(t, faceConflictNames(nil))
	})
}
