package entity

import (
	"sync/atomic"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
)

// raceTestFile is the fixture file the markers of these tests are stored on.
const raceTestFile = "fs6sg6bw45bn0001"

// raceTestPerson stores a person for a test.
func raceTestPerson(t *testing.T, name string) *Subject {
	t.Helper()

	s := NewSubject(name, SubjPerson, SrcManual)
	require.NoError(t, s.Create())
	t.Cleanup(func() { UnscopedDb().Delete(Subject{}, "subj_uid = ?", s.SubjUID) })

	return s
}

// raceTestFace stores a cluster of subjUID for a test.
func raceTestFace(t *testing.T, subjUID string, seed uint64) *Face {
	t.Helper()

	f := NewFace(subjUID, SrcAuto, face.Embeddings{face.FixtureEmbedding(seed)}, face.EmbeddingModelName())
	require.NoError(t, f.Create())
	t.Cleanup(func() { UnscopedDb().Delete(Face{}, "id = ?", f.ID) })

	return f
}

// raceBeforeMarkerUpdate runs fn once, before the first markers update issued after the returned flag is set.
func raceBeforeMarkerUpdate(t *testing.T, name string, fn func()) *atomic.Bool {
	t.Helper()

	armed := &atomic.Bool{}

	Db().Callback().Update().Before("gorm:begin_transaction").Register(name, func(scope *gorm.Scope) {
		if scope.TableName() == (Marker{}).TableName() && armed.CompareAndSwap(true, false) {
			fn()
		}
	})
	t.Cleanup(func() { Db().Callback().Update().Remove(name) })

	return armed
}

// raceTestRename names cluster f after subjUID and relinks its automatic markers, as a faces run does.
func raceTestRename(t *testing.T, f *Face, subjUID string) {
	t.Helper()

	require.NoError(t, UnscopedDb().Exec("UPDATE faces SET subj_uid = ? WHERE id = ?", subjUID, f.ID).Error)
	require.NoError(t, UnscopedDb().Exec("UPDATE markers SET subj_uid = ? WHERE face_id = ? AND subj_src = ?", subjUID, f.ID, SrcAuto).Error)
}

// TestMarker_SyncSubject_ConcurrentRename pins that the automatic markers of a cluster renamed between
// the check and the update keep the cluster's person.
func TestMarker_SyncSubject_ConcurrentRename(t *testing.T) {
	ada := raceTestPerson(t, "Race Rename Ada")
	bea := raceTestPerson(t, "Race Rename Bea")
	f := raceTestFace(t, ada.SubjUID, 7921)

	sibling := syncTestMarkers(t, f, 1, "", raceTestFile)
	named := syncTestMarkers(t, f, 1, ada.SubjUID, raceTestFile)
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", named[0]).
		UpdateColumns(Values{"subj_src": SrcManual, "marker_name": ada.SubjName}).Error)

	m := FindMarker(named[0])
	require.NotNil(t, m)
	m.face = FindFace(f.ID)
	require.Equal(t, ada.SubjUID, m.face.SubjUID)

	armed := raceBeforeMarkerUpdate(t, "race:sync-rename", func() { raceTestRename(t, f, bea.SubjUID) })
	armed.Store(true)
	require.NoError(t, m.SyncSubject(true))
	require.False(t, armed.Load(), "the rename ran")

	assert.Equal(t, bea.SubjUID, FindFace(f.ID).SubjUID)
	assert.Equal(t, bea.SubjUID, FindMarker(sibling[0]).SubjUID, "siblings keep the cluster's person")
}

// TestMarker_SyncSubject_StaleOtherPerson pins that a cached cluster showing another person reports
// nothing when the stored cluster carries the marker's person, and its automatic markers follow.
func TestMarker_SyncSubject_StaleOtherPerson(t *testing.T) {
	cal := raceTestPerson(t, "Race Stale Cal")
	dee := raceTestPerson(t, "Race Stale Dee")
	f := raceTestFace(t, dee.SubjUID, 7922)

	sibling := syncTestMarkers(t, f, 1, "", raceTestFile)
	named := syncTestMarkers(t, f, 1, cal.SubjUID, raceTestFile)
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", named[0]).
		UpdateColumns(Values{"subj_src": SrcManual, "marker_name": cal.SubjName}).Error)

	m := FindMarker(named[0])
	require.NotNil(t, m)
	m.face = FindFace(f.ID)
	require.Equal(t, dee.SubjUID, m.face.SubjUID)

	// Someone names the cluster after the marker's person once the copy was loaded.
	raceTestRename(t, f, cal.SubjUID)

	require.NoError(t, m.SyncSubject(true))

	stored := FindFace(f.ID)
	require.NotNil(t, stored)
	assert.Zero(t, stored.Collisions, "no collision reported")
	assert.Equal(t, f.ID, FindMarker(named[0]).FaceID, "the marker stays in its cluster")
	assert.Equal(t, cal.SubjUID, FindMarker(sibling[0]).SubjUID, "siblings follow")
}

// TestFace_ResolveCollision_StaleNarrowing pins that a copy loaded before a tighter narrowing does not
// widen the stored cluster.
func TestFace_ResolveCollision_StaleNarrowing(t *testing.T) {
	eve := raceTestPerson(t, "Race Narrow Eve")
	f := raceTestFace(t, eve.SubjUID, 7923)
	syncTestMarkers(t, f, 2, eve.SubjUID, raceTestFile)

	stale := FindFace(f.ID)
	require.NotNil(t, stale)

	accept := f.AcceptDist()
	tight := 0.4 * accept
	require.Greater(t, tight, face.CollisionDist, "the stored narrowing is one that narrows")
	require.NoError(t, UnscopedDb().Model(&Face{}).Where("id = ?", f.ID).UpdateColumn("collision_radius", tight).Error)

	// A collision the stale copy still matches, further out than the stored narrowing.
	e := face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), 0.8*accept, 31)}
	match, dist := stale.Match(e, stale.EmbedModel)
	require.True(t, match)
	require.Greater(t, dist-face.Epsilon, tight)

	resolved, err := stale.ResolveCollision(e, stale.EmbedModel)
	require.NoError(t, err)
	assert.False(t, resolved)

	stored := FindFace(f.ID)
	require.NotNil(t, stored)
	assert.InDelta(t, tight, stored.CollisionRadius, 1e-9, "stored radius unchanged")
	assert.Zero(t, stored.Collisions)
}

// TestFace_ResolveCollision_Narrows pins that a collision further in than the stored narrowing still
// narrows the cluster and releases what it no longer holds.
func TestFace_ResolveCollision_Narrows(t *testing.T) {
	fay := raceTestPerson(t, "Race Narrow Fay")
	f := raceTestFace(t, fay.SubjUID, 7924)
	accept := f.AcceptDist()

	outer := syncTestMarkers(t, f, 1, fay.SubjUID, raceTestFile)
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", outer[0]).UpdateColumn("embeddings_json",
		face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), 0.7*accept, 41)}.JSON()).Error)
	require.NoError(t, UnscopedDb().Model(&Face{}).Where("id = ?", f.ID).UpdateColumn("collision_radius", 0.9*accept).Error)

	f = FindFace(f.ID)
	require.NotNil(t, f)

	resolved, err := f.ResolveCollision(face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), 0.5*accept, 42)}, f.EmbedModel)
	require.NoError(t, err)
	assert.True(t, resolved)

	stored := FindFace(f.ID)
	require.NotNil(t, stored)
	assert.Less(t, stored.CollisionRadius, 0.9*accept, "narrowed")
	assert.Equal(t, 1, stored.Collisions)
	assert.Nil(t, stored.MatchedAt, "reopened")
	assert.Empty(t, FindMarker(outer[0]).FaceID, "released")
}

// TestMarker_SetFace_ConcurrentNaming pins that a faces run does not overwrite a person named on the
// cluster after the run loaded it.
func TestMarker_SetFace_ConcurrentNaming(t *testing.T) {
	gus := raceTestPerson(t, "Race Naming Gus")
	hal := raceTestPerson(t, "Race Naming Hal")
	f := raceTestFace(t, "", 7925)

	sibling := syncTestMarkers(t, f, 1, "", raceTestFile)
	named := syncTestMarkers(t, f, 1, gus.SubjUID, raceTestFile)
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", named[0]).
		UpdateColumns(Values{"face_id": "", "subj_src": SrcManual, "marker_name": gus.SubjName}).Error)

	m := FindMarker(named[0])
	require.NotNil(t, m)

	stale := FindFace(f.ID)
	require.NotNil(t, stale)
	require.Empty(t, stale.SubjUID)

	// A person names the cluster during the run.
	raceTestRename(t, f, hal.SubjUID)

	updated, err := m.SetFace(stale, 0.5*stale.AcceptDist())
	require.NoError(t, err)
	assert.False(t, updated)

	assert.Equal(t, hal.SubjUID, FindFace(f.ID).SubjUID, "the run does not overwrite it")
	assert.Equal(t, hal.SubjUID, FindMarker(sibling[0]).SubjUID)
	assert.Empty(t, FindMarker(named[0]).FaceID, "the marker is left for the next run")
}

// TestMarker_SetFace_NamesUnnamedCluster pins that a faces run still names an unnamed cluster after a
// marker's person, and its automatic markers follow.
func TestMarker_SetFace_NamesUnnamedCluster(t *testing.T) {
	ivo := raceTestPerson(t, "Race Naming Ivo")
	f := raceTestFace(t, "", 7926)

	sibling := syncTestMarkers(t, f, 1, "", raceTestFile)
	named := syncTestMarkers(t, f, 1, ivo.SubjUID, raceTestFile)
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", named[0]).
		UpdateColumns(Values{"face_id": "", "subj_src": SrcManual, "marker_name": ivo.SubjName}).Error)

	m := FindMarker(named[0])
	require.NotNil(t, m)

	updated, err := m.SetFace(f, 0.5*f.AcceptDist())
	require.NoError(t, err)
	assert.True(t, updated)

	assert.Equal(t, ivo.SubjUID, f.SubjUID)
	assert.Equal(t, ivo.SubjUID, FindFace(f.ID).SubjUID)
	assert.Equal(t, ivo.SubjUID, FindMarker(sibling[0]).SubjUID)
	assert.Equal(t, f.ID, FindMarker(named[0]).FaceID)
}

// TestFace_ClaimSubject pins when a cluster is named after a person and its automatic markers follow.
func TestFace_ClaimSubject(t *testing.T) {
	jan := raceTestPerson(t, "Race Claim Jan")
	kit := raceTestPerson(t, "Race Claim Kit")

	t.Run("Unnamed", func(t *testing.T) {
		f := raceTestFace(t, "", 7927)
		sibling := syncTestMarkers(t, f, 1, "", raceTestFile)

		carries, err := f.ClaimSubject(jan.SubjUID)
		require.NoError(t, err)
		assert.True(t, carries)
		assert.Equal(t, jan.SubjUID, f.SubjUID)
		assert.Equal(t, jan.SubjUID, FindFace(f.ID).SubjUID)
		assert.Equal(t, jan.SubjUID, FindMarker(sibling[0]).SubjUID)
	})
	t.Run("SamePerson", func(t *testing.T) {
		f := raceTestFace(t, jan.SubjUID, 7928)
		sibling := syncTestMarkers(t, f, 1, "", raceTestFile)
		stale := &Face{ID: f.ID}

		carries, err := stale.ClaimSubject(jan.SubjUID)
		require.NoError(t, err)
		assert.True(t, carries)
		assert.Equal(t, jan.SubjUID, FindMarker(sibling[0]).SubjUID)
	})
	t.Run("OtherPerson", func(t *testing.T) {
		f := raceTestFace(t, kit.SubjUID, 7929)
		sibling := syncTestMarkers(t, f, 1, kit.SubjUID, raceTestFile)
		stale := &Face{ID: f.ID}

		carries, err := stale.ClaimSubject(jan.SubjUID)
		require.NoError(t, err)
		assert.False(t, carries)
		assert.Equal(t, kit.SubjUID, stale.SubjUID, "the copy takes over the stored person")
		assert.Equal(t, kit.SubjUID, FindFace(f.ID).SubjUID)
		assert.Equal(t, kit.SubjUID, FindMarker(sibling[0]).SubjUID)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := (&Face{}).ClaimSubject(jan.SubjUID)
		assert.Error(t, err)
		_, err = (&Face{ID: "X"}).ClaimSubject("")
		assert.Error(t, err)
	})
}

// TestMarker_SetFace_StaleSubjectNarrowing pins that a cluster reassigned to the marker's own person
// after the run loaded it is not narrowed against that person.
func TestMarker_SetFace_StaleSubjectNarrowing(t *testing.T) {
	ada := raceTestPerson(t, "Race Stale Ada")
	bea := raceTestPerson(t, "Race Stale Bea")
	f := raceTestFace(t, ada.SubjUID, 7941)

	named := syncTestMarkers(t, f, 1, bea.SubjUID, raceTestFile)
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", named[0]).
		UpdateColumns(Values{"face_id": "", "subj_src": SrcManual, "marker_name": bea.SubjName}).Error)

	stale := FindFace(f.ID)
	require.NotNil(t, stale)
	m := FindMarker(named[0])
	require.NotNil(t, m)
	match, dist := stale.Match(m.Embeddings(), m.EmbedModel)
	require.True(t, match)
	require.Greater(t, dist-face.Epsilon, face.CollisionDist, "narrowing branch")

	// A person reassigns the cluster to the marker's person while the run holds the copy.
	raceTestRename(t, f, bea.SubjUID)

	_, err := m.SetFace(stale, dist)
	require.NoError(t, err)

	stored := FindFace(f.ID)
	require.NotNil(t, stored)
	assert.Zero(t, stored.Collisions, "no collision of Bea with Bea's own cluster")
	assert.Zero(t, stored.CollisionRadius)
	assert.Equal(t, bea.SubjUID, stale.SubjUID, "the run's copy takes over the stored person")
}

// TestMarker_SyncSubject_StoredUnnamed pins that a hand-given name names a cluster that was unnamed
// after the marker's copy was loaded, and its automatic markers follow.
func TestMarker_SyncSubject_StoredUnnamed(t *testing.T) {
	cal := raceTestPerson(t, "Race Unnamed Cal")
	dee := raceTestPerson(t, "Race Unnamed Dee")
	f := raceTestFace(t, dee.SubjUID, 7942)

	sibling := syncTestMarkers(t, f, 1, "", raceTestFile)
	named := syncTestMarkers(t, f, 1, cal.SubjUID, raceTestFile)
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", named[0]).
		UpdateColumns(Values{"subj_src": SrcManual, "marker_name": cal.SubjName}).Error)

	m := FindMarker(named[0])
	require.NotNil(t, m)
	m.face = FindFace(f.ID)
	require.Equal(t, dee.SubjUID, m.face.SubjUID)

	// Dee is deleted, which unnames the cluster (Subject.Delete clears faces.subj_uid).
	require.NoError(t, UnscopedDb().Exec("UPDATE faces SET subj_uid = '' WHERE id = ?", f.ID).Error)

	require.NoError(t, m.SyncSubject(true))

	assert.Equal(t, f.ID, FindMarker(named[0]).FaceID, "the marker stays")
	assert.Equal(t, cal.SubjUID, FindFace(f.ID).SubjUID, "the hand-given name names the unnamed cluster")
	assert.Equal(t, cal.SubjUID, FindMarker(sibling[0]).SubjUID, "siblings follow")
}

// TestFace_ClaimSubject_ConcurrentRename pins that automatic markers keep the person of a cluster
// renamed between the claim and the relink.
func TestFace_ClaimSubject_ConcurrentRename(t *testing.T) {
	jan := raceTestPerson(t, "Race Claim2 Jan")
	kit := raceTestPerson(t, "Race Claim2 Kit")
	f := raceTestFace(t, "", 7943)
	sibling := syncTestMarkers(t, f, 1, "", raceTestFile)

	armed := raceBeforeMarkerUpdate(t, "race:claim-rename", func() { raceTestRename(t, f, kit.SubjUID) })
	armed.Store(true)

	carries, err := f.ClaimSubject(jan.SubjUID)
	require.NoError(t, err)
	require.False(t, armed.Load(), "the rename ran")
	assert.True(t, carries)
	assert.Equal(t, kit.SubjUID, FindFace(f.ID).SubjUID)
	assert.Equal(t, kit.SubjUID, FindMarker(sibling[0]).SubjUID, "siblings keep the cluster's person")
}

// TestMarker_SetFace_ConcurrentNamingByName pins that a marker whose person is resolved from its name
// does not rename a cluster someone named during the run, and is not matched against it again.
func TestMarker_SetFace_ConcurrentNamingByName(t *testing.T) {
	ivy := raceTestPerson(t, "Race Naming2 Ivy")
	hal := raceTestPerson(t, "Race Naming2 Hal")
	f := raceTestFace(t, "", 7944)

	sibling := syncTestMarkers(t, f, 1, "", raceTestFile)
	named := syncTestMarkers(t, f, 1, "", raceTestFile)
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", named[0]).
		UpdateColumns(Values{"face_id": "", "subj_src": SrcManual, "marker_name": ivy.SubjName}).Error)

	m := FindMarker(named[0])
	require.NotNil(t, m)
	stale := FindFace(f.ID)
	require.NotNil(t, stale)

	raceTestRename(t, f, hal.SubjUID)

	updated, err := m.SetFace(stale, 0.5*stale.AcceptDist())
	require.NoError(t, err)
	assert.False(t, updated)
	assert.Equal(t, hal.SubjUID, FindFace(f.ID).SubjUID)
	assert.Equal(t, hal.SubjUID, FindMarker(sibling[0]).SubjUID)
	assert.Empty(t, FindMarker(named[0]).FaceID, "Ivy's marker does not join Hal's cluster")
	assert.Empty(t, m.FaceID, "the refused copy keeps no face")
	assert.Empty(t, m.SubjUID, "and no person")
	assert.Equal(t, ivy.SubjName, m.MarkerName)
	assert.NotNil(t, FindMarker(named[0]).MatchedAt, "and waits for a forced run")
	assert.Equal(t, hal.SubjUID, stale.SubjUID, "the run's copy takes over the stored person")
}

// TestFace_refreshCollision pins that a copy takes over the stored person and collision fields.
func TestFace_refreshCollision(t *testing.T) {
	lou := raceTestPerson(t, "Race Refresh Lou")
	f := raceTestFace(t, lou.SubjUID, 7945)

	t.Run("Success", func(t *testing.T) {
		require.NoError(t, UnscopedDb().Model(&Face{}).Where("id = ?", f.ID).
			UpdateColumns(Values{"subj_uid": "", "collisions": 3, "collision_radius": 0.25, "face_kind": int(face.AmbiguousFace)}).Error)

		copied := *f
		copied.refreshCollision()
		assert.Empty(t, copied.SubjUID)
		assert.Equal(t, 3, copied.Collisions)
		assert.InDelta(t, 0.25, copied.CollisionRadius, 1e-9)
		assert.Equal(t, int(face.AmbiguousFace), copied.FaceKind)
	})
	t.Run("NotFound", func(t *testing.T) {
		missing := Face{ID: "RACEREFRESHMISSING", SubjUID: lou.SubjUID, Collisions: 1}
		missing.refreshCollision()
		assert.Equal(t, lou.SubjUID, missing.SubjUID, "a copy without a stored row keeps its values")
		assert.Equal(t, 1, missing.Collisions)
	})
}

// TestFace_ResolveCollision_StaleAmbiguous pins that a copy showing another person than the stored
// cluster does not mark that cluster ambiguous.
func TestFace_ResolveCollision_StaleAmbiguous(t *testing.T) {
	mia := raceTestPerson(t, "Race Ambiguous Mia")
	ned := raceTestPerson(t, "Race Ambiguous Ned")
	f := raceTestFace(t, mia.SubjUID, 7946)
	syncTestMarkers(t, f, 2, mia.SubjUID, raceTestFile)

	stale := FindFace(f.ID)
	require.NotNil(t, stale)
	raceTestRename(t, f, ned.SubjUID)

	e := face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), face.AmbiguityDist()/2, 51)}
	_, dist := stale.Match(e, stale.EmbedModel)
	require.Less(t, dist, face.AmbiguityDist(), "ambiguous branch")

	resolved, err := stale.ResolveCollision(e, stale.EmbedModel)
	require.NoError(t, err)
	assert.False(t, resolved)

	stored := FindFace(f.ID)
	require.NotNil(t, stored)
	assert.Equal(t, int(face.RegularFace), stored.FaceKind, "not marked ambiguous")
	assert.Zero(t, stored.Collisions)
	assert.Equal(t, ned.SubjUID, stale.SubjUID, "the copy takes over the stored person")
}

// TestMarker_SetFace_StaleSubjectInBand pins that a marker the stale copy would refuse as a noted
// collision joins the cluster once the stored cluster carries the marker's person.
func TestMarker_SetFace_StaleSubjectInBand(t *testing.T) {
	ora := raceTestPerson(t, "Race Band Ora")
	pia := raceTestPerson(t, "Race Band Pia")
	f := raceTestFace(t, ora.SubjUID, 7947)

	dist := (face.AmbiguityDist() + face.CollisionDist) / 2
	radius := dist - face.Epsilon
	require.Greater(t, radius, 0.0)
	require.NoError(t, UnscopedDb().Model(&Face{}).Where("id = ?", f.ID).UpdateColumn("collision_radius", radius/2).Error)

	named := syncTestMarkers(t, f, 1, pia.SubjUID, raceTestFile)
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", named[0]).UpdateColumns(Values{
		"face_id": "", "subj_src": SrcManual, "marker_name": pia.SubjName, "matched_at": nil,
		"embeddings_json": face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), dist, 61)}.JSON(),
	}).Error)

	stale := FindFace(f.ID)
	require.NotNil(t, stale)
	m := FindMarker(named[0])
	require.NotNil(t, m)
	_, got := stale.Match(m.Embeddings(), m.EmbedModel)
	require.True(t, stale.CollisionNoted(got), "the stale copy would refuse it as noted")

	raceTestRename(t, f, pia.SubjUID)

	updated, err := m.SetFace(stale, got)
	require.NoError(t, err)
	assert.True(t, updated)
	assert.Equal(t, f.ID, FindMarker(named[0]).FaceID, "the marker joins its person's cluster")
}

// TestFace_ResolveCollision_StaleSubjectInBand pins that a copy showing another person than the stored
// cluster records no collision that cannot narrow it.
func TestFace_ResolveCollision_StaleSubjectInBand(t *testing.T) {
	quin := raceTestPerson(t, "Race Band Quin")
	rae := raceTestPerson(t, "Race Band Rae")
	f := raceTestFace(t, quin.SubjUID, 7948)
	syncTestMarkers(t, f, 2, quin.SubjUID, raceTestFile)

	stale := FindFace(f.ID)
	require.NotNil(t, stale)
	raceTestRename(t, f, rae.SubjUID)

	dist := (face.AmbiguityDist() + face.CollisionDist) / 2
	e := face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), dist, 71)}
	_, got := stale.Match(e, stale.EmbedModel)
	require.False(t, stale.CollisionNoted(got))
	require.LessOrEqual(t, got-face.Epsilon, face.CollisionDist, "band branch")

	resolved, err := stale.ResolveCollision(e, stale.EmbedModel)
	require.NoError(t, err)
	assert.False(t, resolved)

	stored := FindFace(f.ID)
	require.NotNil(t, stored)
	assert.Zero(t, stored.Collisions)
	assert.Zero(t, stored.CollisionRadius)
}

// TestMarker_joins pins when a marker may join a cluster whose stored row changed during a run.
func TestMarker_joins(t *testing.T) {
	sam := raceTestPerson(t, "Race Joins Sam")
	tia := raceTestPerson(t, "Race Joins Tia")
	f := raceTestFace(t, sam.SubjUID, 7949)
	near := face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), 0.3*f.AcceptDist(), 81)}
	m := &Marker{SubjUID: sam.SubjUID, EmbeddingsJSON: near.JSON(), EmbedModel: f.EmbedModel}

	t.Run("SamePerson", func(t *testing.T) {
		assert.True(t, m.joins(f))
	})
	t.Run("Unnamed", func(t *testing.T) {
		unnamed := *f
		unnamed.SubjUID = ""
		assert.True(t, m.joins(&unnamed))
	})
	t.Run("OtherPerson", func(t *testing.T) {
		other := *f
		other.SubjUID = tia.SubjUID
		assert.False(t, m.joins(&other))
	})
	t.Run("Ambiguous", func(t *testing.T) {
		ambiguous := *f
		ambiguous.FaceKind = int(face.AmbiguousFace)
		assert.False(t, m.joins(&ambiguous))
	})
	t.Run("Narrowed", func(t *testing.T) {
		narrowed := *f
		narrowed.CollisionRadius = 0.2 * f.AcceptDist()
		require.Greater(t, narrowed.CollisionRadius, face.CollisionDist)
		assert.False(t, m.joins(&narrowed))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.False(t, m.joins(nil))
	})
}

// TestMarker_SetFace_StoredUnnamedInBand pins that a marker the stale copy would refuse as a noted
// collision names the cluster once the stored cluster is unnamed.
func TestMarker_SetFace_StoredUnnamedInBand(t *testing.T) {
	uma := raceTestPerson(t, "Race Band Uma")
	val := raceTestPerson(t, "Race Band Val")
	f := raceTestFace(t, uma.SubjUID, 7950)

	dist := (face.AmbiguityDist() + face.CollisionDist) / 2
	radius := dist - face.Epsilon
	require.NoError(t, UnscopedDb().Model(&Face{}).Where("id = ?", f.ID).UpdateColumn("collision_radius", radius/2).Error)

	named := syncTestMarkers(t, f, 1, val.SubjUID, raceTestFile)
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", named[0]).UpdateColumns(Values{
		"face_id": "", "subj_src": SrcManual, "marker_name": val.SubjName, "matched_at": nil,
		"embeddings_json": face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), dist, 91)}.JSON(),
	}).Error)

	stale := FindFace(f.ID)
	require.NotNil(t, stale)
	m := FindMarker(named[0])
	require.NotNil(t, m)
	_, got := stale.Match(m.Embeddings(), m.EmbedModel)
	require.True(t, stale.CollisionNoted(got))

	// The person is deleted, which unnames the cluster.
	require.NoError(t, UnscopedDb().Exec("UPDATE faces SET subj_uid = '' WHERE id = ?", f.ID).Error)

	updated, err := m.SetFace(stale, got)
	require.NoError(t, err)
	assert.True(t, updated)
	assert.Equal(t, val.SubjUID, FindFace(f.ID).SubjUID, "the marker names the unnamed cluster")
	assert.Equal(t, f.ID, FindMarker(named[0]).FaceID)
}
