package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
)

// TestMarker_SetFace_StoredAcceptance checks assignment against the stored cluster.
func TestMarker_SetFace_StoredAcceptance(t *testing.T) {
	for i, tc := range []struct {
		name     string
		radius   float64
		rename   bool
		deleted  bool
		accepted bool
	}{
		{name: "Narrowed", radius: 0.25},
		{name: "Renamed", rename: true},
		{name: "Deleted", deleted: true},
		{name: "Inside", radius: 0.75, accepted: true},
		{name: "Unbounded", accepted: true},
		{name: "InertBand", radius: face.CollisionDist / 2, accepted: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			person := raceTestPerson(t, "Stored Acceptance "+tc.name)
			f := raceTestFace(t, "", uint64(8200+i))
			uids := syncTestMarkers(t, f, 1, "", raceTestFile)
			require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", uids[0]).UpdateColumns(Values{"face_id": "", "face_dist": -1, "matched_at": nil}).Error)
			m := FindMarker(uids[0])
			require.NotNil(t, m)
			before := *m
			if tc.deleted {
				require.NoError(t, UnscopedDb().Delete(Face{}, "id = ?", f.ID).Error)
			} else {
				values := Values{"collision_radius": tc.radius}
				if tc.rename {
					values["subj_uid"] = person.SubjUID
				}
				require.NoError(t, UnscopedDb().Model(&Face{}).Where("id = ?", f.ID).UpdateColumns(values).Error)
			}
			statements := countStatements(t, func() {
				updated, err := m.SetFace(f, 0.5)
				require.NoError(t, err)
				assert.Equal(t, tc.accepted, updated)
			})
			stored := FindMarker(m.MarkerUID)
			require.NotNil(t, stored)
			if tc.accepted {
				assert.Equal(t, f.ID, stored.FaceID)
				assert.Equal(t, 0.5, stored.FaceDist)
				assert.NotNil(t, stored.MatchedAt)
				assert.Len(t, statements, 2, "assignment and photo refresh: %v", statements)
			} else {
				assert.Len(t, statements, 2, "guarded update and cluster read only: %v", statements)
				assert.Equal(t, before.FaceID, stored.FaceID)
				assert.Equal(t, before.FaceDist, stored.FaceDist)
				assert.Equal(t, before.SubjUID, stored.SubjUID)
				assert.Nil(t, stored.MatchedAt)
				assert.Equal(t, before.FaceID, m.FaceID)
				assert.Equal(t, before.FaceDist, m.FaceDist)
				assert.Equal(t, before.SubjUID, m.SubjUID)
				assert.Nil(t, m.MatchedAt)
				if !tc.deleted {
					assert.Equal(t, tc.radius, f.CollisionRadius)
				}
				if tc.rename {
					assert.Equal(t, person.SubjUID, f.SubjUID)
				}
			}
		})
	}
}

// TestMarker_SetFace_StoredAcceptanceNaming checks the final assignment of a named marker.
func TestMarker_SetFace_StoredAcceptanceNaming(t *testing.T) {
	person := raceTestPerson(t, "Stored Acceptance Naming")
	f := raceTestFace(t, "", 8210)
	syncTestMarkers(t, f, 1, "", raceTestFile)
	named := syncTestMarkers(t, f, 1, person.SubjUID, raceTestFile)
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", named[0]).UpdateColumns(Values{
		"face_id": "", "face_dist": -1, "matched_at": nil, "subj_src": SrcManual, "marker_name": person.SubjName,
	}).Error)
	require.NoError(t, UnscopedDb().Model(&Face{}).Where("id = ?", f.ID).UpdateColumn("collision_radius", 0.25).Error)
	m := FindMarker(named[0])
	require.NotNil(t, m)
	updated, err := m.SetFace(f, 0.5)
	require.NoError(t, err)
	assert.False(t, updated)
	assert.Empty(t, FindMarker(named[0]).FaceID)
}

// TestMarker_SetFace_StoredAcceptanceRejected checks matching after a named member is rejected.
func TestMarker_SetFace_StoredAcceptanceRejected(t *testing.T) {
	for i, name := range []string{"NamedCluster", "RenamedCluster", "NarrowedCluster"} {
		renamed := name == "RenamedCluster"
		narrowed := name == "NarrowedCluster"
		t.Run(name, func(t *testing.T) {
			person := raceTestPerson(t, "Acceptance Rejected "+name)
			f := raceTestFace(t, person.SubjUID, uint64(8211+i))
			siblings := syncTestMarkers(t, f, 1, person.SubjUID, raceTestFile)
			require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", siblings[0]).UpdateColumns(Values{
				"face_dist": 0, "embeddings_json": face.Embeddings{f.Embedding()}.JSON(),
			}).Error)
			uids := syncTestMarkers(t, f, 1, person.SubjUID, raceTestFile)
			dist := face.CollisionDist / 2
			if narrowed {
				dist = 0.5 * f.AcceptDist()
			}
			require.Greater(t, dist, face.AmbiguityDist())
			require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", uids[0]).UpdateColumn("embeddings_json",
				face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), dist, 8220)}.JSON()).Error)
			m := FindMarker(uids[0])
			require.NotNil(t, m)
			require.Equal(t, f.ID, m.FaceID)
			require.Equal(t, person.SubjUID, m.SubjUID)
			require.NoError(t, m.ClearSubject(SrcManual))
			m = FindMarker(uids[0])
			require.NotNil(t, m)
			require.True(t, m.RejectedMatch())
			require.Empty(t, m.FaceID)
			before := *m
			f = FindFace(f.ID)
			require.NotNil(t, f)
			require.Equal(t, person.SubjUID, f.SubjUID)
			require.Positive(t, f.CollisionRadius)
			if narrowed {
				require.Greater(t, f.CollisionRadius, face.CollisionDist)
				require.Less(t, f.CollisionRadius, dist)
			} else {
				require.LessOrEqual(t, f.CollisionRadius, face.CollisionDist)
			}
			if renamed {
				other := raceTestPerson(t, "Acceptance Rejected Other")
				require.NoError(t, UnscopedDb().Model(&Face{}).Where("id = ?", f.ID).UpdateColumn("subj_uid", other.SubjUID).Error)
			}
			updated, err := m.SetFace(f, dist)
			require.NoError(t, err)
			assert.Equal(t, !renamed && !narrowed, updated)
			stored := FindMarker(m.MarkerUID)
			require.NotNil(t, stored)
			assert.True(t, stored.RejectedMatch())
			assert.Empty(t, stored.SubjUID)
			if renamed || narrowed {
				assert.Empty(t, stored.FaceID)
				assert.Equal(t, before.MatchedAt, stored.MatchedAt)
			} else {
				assert.Equal(t, f.ID, stored.FaceID)
				assert.NotNil(t, stored.MatchedAt)
			}
		})
	}
}

// TestMarker_SetFace_AcceptanceBounds checks the stored radius, its inert floor, and computed distances.
func TestMarker_SetFace_AcceptanceBounds(t *testing.T) {
	for i, name := range []string{"Boundary", "Floor", "Widened", "Named", "Computed", "Repeated"} {
		t.Run(name, func(t *testing.T) {
			f := raceTestFace(t, "", uint64(8240+i))
			if name == "Named" {
				person := raceTestPerson(t, "Acceptance Bounds Named")
				require.NoError(t, f.Update("subj_uid", person.SubjUID))
				f.SubjUID = person.SubjUID
			}
			uids := syncTestMarkers(t, f, 1, "", raceTestFile)
			require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", uids[0]).UpdateColumns(Values{"face_id": "", "face_dist": -1, "matched_at": nil}).Error)
			m := FindMarker(uids[0])
			require.NotNil(t, m)
			dist := m.Embeddings().Dist(f.Embedding())
			require.Greater(t, dist, face.CollisionDist)
			input := dist
			radius := dist
			switch name {
			case "Floor":
				radius = face.CollisionDist
			case "Widened":
				f.CollisionRadius = dist / 2
				radius = 0
			case "Computed":
				input = -1
				radius = dist / 2
			}
			require.NoError(t, UnscopedDb().Model(&Face{}).Where("id = ?", f.ID).UpdateColumn("collision_radius", radius).Error)
			before := *m
			statements := countStatements(t, func() {
				updated, err := m.SetFace(f, input)
				require.NoError(t, err)
				assert.Equal(t, name != "Computed", updated)
			})
			require.Len(t, statements, 2, "%v", statements)
			if name == "Computed" {
				assert.Equal(t, before, *m)
				assert.Empty(t, FindMarker(m.MarkerUID).FaceID)
				return
			}
			stored := FindMarker(m.MarkerUID)
			assert.Equal(t, f.ID, stored.FaceID)
			assert.Equal(t, f.SubjUID, stored.SubjUID)
			assert.InDelta(t, dist, stored.FaceDist, 1e-6)
			if name == "Repeated" {
				updated, err := m.SetFace(f, input)
				require.NoError(t, err)
				assert.False(t, updated)
				assert.Equal(t, f.ID, m.FaceID)
				assert.Equal(t, FindMarker(m.MarkerUID).MatchedAt, m.MatchedAt)
				assert.False(t, m.MatchedAt.Before(*stored.MatchedAt))
			}
		})
	}
}

// TestMarker_SetFace_ConcurrentAcceptance checks eligibility at the marker write.
func TestMarker_SetFace_ConcurrentAcceptance(t *testing.T) {
	for i, name := range []string{"Radius", "Subject", "Deleted"} {
		t.Run(name, func(t *testing.T) {
			person := raceTestPerson(t, "Concurrent Acceptance "+name)
			f := raceTestFace(t, person.SubjUID, uint64(8250+i))
			uids := syncTestMarkers(t, f, 1, "", raceTestFile)
			require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", uids[0]).UpdateColumns(Values{"face_id": "", "face_dist": -1, "matched_at": nil}).Error)
			m := FindMarker(uids[0])
			require.NotNil(t, m)
			before := *m
			armed := raceBeforeMarkerUpdate(t, "acceptance:"+name, func() {
				switch name {
				case "Radius":
					require.NoError(t, UnscopedDb().Exec("UPDATE faces SET collision_radius = ? WHERE id = ?", 0.25, f.ID).Error)
				case "Subject":
					require.NoError(t, UnscopedDb().Exec("UPDATE faces SET subj_uid = '' WHERE id = ?", f.ID).Error)
				case "Deleted":
					require.NoError(t, UnscopedDb().Exec("DELETE FROM faces WHERE id = ?", f.ID).Error)
				}
			})
			armed.Store(true)
			updated, err := m.SetFace(f, 0.5)
			require.NoError(t, err)
			require.False(t, armed.Load())
			assert.False(t, updated)
			assert.Equal(t, before, *m)
			stored := FindMarker(m.MarkerUID)
			assert.Empty(t, stored.FaceID)
			assert.Empty(t, stored.SubjUID)
			assert.Nil(t, stored.MatchedAt)
		})
	}
}

// TestMarker_SetFace_StoredAcceptanceSameFace checks eligibility before stamping an existing membership.
func TestMarker_SetFace_StoredAcceptanceSameFace(t *testing.T) {
	for i, mode := range []string{"Narrowed", "Renamed", "Deleted", "Accepted"} {
		t.Run(mode, func(t *testing.T) {
			f := raceTestFace(t, "", uint64(8270+i))
			uids := syncTestMarkers(t, f, 1, "", raceTestFile)
			require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", uids[0]).UpdateColumn("matched_at", Time("2000-01-01T00:00:00Z")).Error)
			m := FindMarker(uids[0])
			require.NotNil(t, m)
			before := *m
			switch mode {
			case "Narrowed":
				require.NoError(t, UnscopedDb().Model(&Face{}).Where("id = ?", f.ID).UpdateColumn("collision_radius", m.FaceDist/2).Error)
			case "Renamed":
				person := raceTestPerson(t, "Same Face Renamed")
				require.NoError(t, UnscopedDb().Model(&Face{}).Where("id = ?", f.ID).UpdateColumn("subj_uid", person.SubjUID).Error)
			case "Deleted":
				require.NoError(t, UnscopedDb().Delete(f).Error)
			}
			updated, err := m.SetFace(f, m.FaceDist)
			require.NoError(t, err)
			assert.False(t, updated)
			if mode == "Accepted" {
				assert.True(t, m.MatchedAt.After(*before.MatchedAt))
				assert.Equal(t, FindMarker(m.MarkerUID).MatchedAt, m.MatchedAt)
			} else {
				assert.Equal(t, before, *m)
				assert.Equal(t, before.MatchedAt, FindMarker(m.MarkerUID).MatchedAt)
			}
		})
	}
}

// TestMarker_SetFace_PhotoRefresh checks maintenance state after accepted and refused assignments.
func TestMarker_SetFace_PhotoRefresh(t *testing.T) {
	flag := UpdateFaces.Load()
	t.Cleanup(func() { UpdateFaces.Store(flag) })
	for _, tc := range []struct {
		name string
		seed uint64
	}{{"Refused", 8290}, {"RefusedPending", 8291}, {"Accepted", 8292}} {
		t.Run(tc.name, func(t *testing.T) {
			name := tc.name
			f := raceTestFace(t, "", tc.seed)
			uids := syncTestMarkers(t, f, 1, "", raceTestFile)
			require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", uids[0]).UpdateColumns(Values{"face_id": "", "face_dist": -1, "matched_at": nil}).Error)
			radius := 0.25
			if name == "Accepted" {
				radius = 0.75
			}
			require.NoError(t, UnscopedDb().Model(&Face{}).Where("id = ?", f.ID).UpdateColumn("collision_radius", radius).Error)
			var file File
			require.NoError(t, UnscopedDb().Where("file_uid = ?", raceTestFile).First(&file).Error)
			var photo Photo
			require.NoError(t, UnscopedDb().First(&photo, file.PhotoID).Error)
			previous := photo.CheckedAt
			t.Cleanup(func() {
				require.NoError(t, UnscopedDb().Model(&Photo{}).Where("id = ?", photo.ID).UpdateColumn("checked_at", previous).Error)
			})
			checked := Time("2026-01-01T12:00:00Z")
			require.NoError(t, UnscopedDb().Model(&Photo{}).Where("id = ?", photo.ID).UpdateColumn("checked_at", checked).Error)
			UpdateFaces.Store(name == "RefusedPending")
			marker := FindMarker(uids[0])
			require.NotNil(t, marker)
			updated, err := marker.SetFace(f, 0.5)
			require.NoError(t, err)
			assert.Equal(t, name == "Accepted", updated)
			var stored Photo
			require.NoError(t, UnscopedDb().First(&stored, photo.ID).Error)
			if name == "Accepted" {
				assert.Nil(t, stored.CheckedAt)
				assert.True(t, UpdateFaces.Load())
			} else {
				assert.Equal(t, checked, stored.CheckedAt)
				assert.Equal(t, name == "RefusedPending", UpdateFaces.Load())
			}
		})
	}
}
