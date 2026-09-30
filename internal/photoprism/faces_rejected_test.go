package photoprism

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestFacesMatchKeepsRejectedMatch pins that matching never names a marker whose name a person
// removed, while a marker nobody rejected is still named by the same cluster.
func TestFacesMatchKeepsRejectedMatch(t *testing.T) {
	c := config.TestConfig()
	w := NewFaces(c)

	jane := entity.SubjectFixtures.Get("jane-doe").SubjUID
	emb := face.Embeddings{face.FixtureEmbedding(6001)}
	named := entity.NewFace(jane, entity.SrcManual, emb, face.EmbeddingModelName())
	require.NotNil(t, named)
	require.NoError(t, named.Create())
	t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Face{}, "id = ?", named.ID) })

	newMarker := func(t *testing.T, subjSrc, name string) *entity.Marker {
		t.Helper()

		m := &entity.Marker{
			MarkerUID:      rnd.GenerateUID('m'),
			MarkerType:     entity.MarkerFace,
			MarkerSrc:      entity.SrcImage,
			MarkerName:     name,
			Size:           face.ClusterSizeThreshold,
			Score:          face.ClusterScore("") + 10,
			EmbeddingsJSON: emb.JSON(),
			EmbedModel:     face.EmbeddingModelName(),
			FaceDist:       -1,
			SubjSrc:        subjSrc,
		}

		require.NoError(t, entity.Db().Create(m).Error)
		t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Marker{}, "marker_uid = ?", m.MarkerUID) })

		return m
	}

	reload := func(t *testing.T, uid string) entity.Marker {
		t.Helper()

		var m entity.Marker
		require.NoError(t, entity.Db().Where("marker_uid = ?", uid).Take(&m).Error)

		return m
	}

	t.Run("Forced", func(t *testing.T) {
		m := newMarker(t, entity.SrcManual, "")
		require.True(t, m.RejectedMatch())

		_, err := w.MatchFaces(entity.Faces{*named}, true, nil, nil)
		require.NoError(t, err)

		got := reload(t, m.MarkerUID)
		assert.Empty(t, got.SubjUID, "a rejected match stays unnamed")
		assert.Equal(t, entity.SrcManual, got.SubjSrc)
		assert.Empty(t, got.MarkerName)
		assert.Equal(t, named.ID, got.FaceID, "membership is not a name")
		assert.NotNil(t, got.MatchedAt)
	})
	t.Run("UnmatchedPass", func(t *testing.T) {
		m := newMarker(t, entity.SrcManual, "")

		_, err := w.MatchFaces(entity.Faces{*named}, false, entity.TimeStamp(), nil)
		require.NoError(t, err)

		got := reload(t, m.MarkerUID)
		assert.Empty(t, got.SubjUID, "a rejected match stays unnamed")
		assert.Equal(t, entity.SrcManual, got.SubjSrc)
		assert.Equal(t, named.ID, got.FaceID, "membership is not a name")
		assert.NotNil(t, got.MatchedAt)
	})
	t.Run("AutomaticMarkerIsNamed", func(t *testing.T) {
		m := newMarker(t, entity.SrcAuto, "")
		require.False(t, m.RejectedMatch())

		_, err := w.MatchFaces(entity.Faces{*named}, true, nil, nil)
		require.NoError(t, err)

		got := reload(t, m.MarkerUID)
		assert.Equal(t, named.ID, got.FaceID)
		assert.Equal(t, jane, got.SubjUID)
	})
}

// rejectTestMarker stores a marker on a cluster whose name a person removed.
func rejectTestMarker(t *testing.T, f *entity.Face) string {
	t.Helper()

	uids := consensusTestMarkers(t, f, 1, "", entity.SrcManual, false)
	require.True(t, entity.FindMarker(uids[0]).RejectedMatch())

	return uids[0]
}

// TestFaces_RejectedMatchStaysUnnamed pins that no pass which names a cluster's markers names one
// whose name a person removed.
func TestFaces_RejectedMatchStaysUnnamed(t *testing.T) {
	w := isolatedTestFaces(t, "facesrejected")
	core := w.conf.FaceClusterCore()

	alice := consensusTestSubject(t, "Rejected Alice")
	bob := consensusTestSubject(t, "Rejected Bob")

	t.Run("SetSubjectUID", func(t *testing.T) {
		f := consensusTestFace(t, 1)
		rejected := rejectTestMarker(t, f)
		auto := consensusTestMarkers(t, f, 1, "", entity.SrcAuto, false)

		require.NoError(t, f.SetSubjectUID(alice.SubjUID))
		assert.Equal(t, []string{""}, consensusTestSubjects(t, []string{rejected}))
		assert.Equal(t, []string{alice.SubjUID}, consensusTestSubjects(t, auto), "positive control")
	})
	t.Run("ConsensusNaming", func(t *testing.T) {
		f := consensusTestFace(t, 2)
		consensusTestMarkers(t, f, core, bob.SubjUID, entity.SrcAuto, false)
		rejected := rejectTestMarker(t, f)

		result, err := w.NameByConsensus()
		require.NoError(t, err)
		require.Equal(t, 1, result.Named)
		assert.Equal(t, bob.SubjUID, entity.FindFace(f.ID).SubjUID)
		assert.Equal(t, []string{""}, consensusTestSubjects(t, []string{rejected}))
	})
	t.Run("MatchFaceMarkers", func(t *testing.T) {
		// Two samples, since a cluster built from one is never matchable.
		f := consensusTestFace(t, 3)
		require.NoError(t, f.Updates(entity.Values{"subj_uid": alice.SubjUID, "samples": 2}))
		rejected := rejectTestMarker(t, f)
		auto := consensusTestMarkers(t, f, 1, "", entity.SrcAuto, false)

		_, err := query.MatchFaceMarkers()
		require.NoError(t, err)
		assert.Equal(t, []string{""}, consensusTestSubjects(t, []string{rejected}))
		assert.Equal(t, []string{alice.SubjUID}, consensusTestSubjects(t, auto), "positive control")
	})
	t.Run("ManualNameNamesCluster", func(t *testing.T) {
		// Positive control: a name a person gave still names the unnamed cluster it joins.
		f := consensusTestFace(t, 6)
		m := entity.FindMarker(consensusTestMarkers(t, consensusTestFace(t, 7), 1, alice.SubjUID, entity.SrcManual, false)[0])
		require.NotNil(t, m)

		_, err := m.SetFace(f, -1)
		require.NoError(t, err)
		assert.Equal(t, alice.SubjUID, entity.FindFace(f.ID).SubjUID)

		got := entity.FindMarker(m.MarkerUID)
		require.NotNil(t, got)
		assert.Equal(t, entity.SrcManual, got.SubjSrc)
		assert.Equal(t, alice.SubjUID, got.SubjUID)
	})
	t.Run("RelatedMarkers", func(t *testing.T) {
		f := consensusTestFace(t, 4)
		rejected := rejectTestMarker(t, f)
		named := entity.FindMarker(consensusTestMarkers(t, f, 1, "", entity.SrcAuto, false)[0])
		require.NotNil(t, named)

		_, err := named.SetName(alice.SubjName, entity.SrcManual)
		require.NoError(t, err)
		assert.Equal(t, []string{""}, consensusTestSubjects(t, []string{rejected}))
	})
}

// TestFaces_AuditRejectedMatch pins that the audit reports a rejected match in a named cluster as
// expected state, and that --fix leaves it in its cluster.
func TestFaces_AuditRejectedMatch(t *testing.T) {
	w := isolatedTestFaces(t, "facesrejectedaudit")

	alice := consensusTestSubject(t, "Rejected Audit Alice")
	bob := consensusTestSubject(t, "Rejected Audit Bob")
	f := consensusTestFace(t, 1)
	require.NoError(t, f.Update("SubjUID", alice.SubjUID))
	rejected := rejectTestMarker(t, f)

	t.Run("OtherSubject", func(t *testing.T) {
		hook := captureLog(t)

		require.NoError(t, w.Audit(false, bob.SubjUID))
		assert.NotContains(t, strings.Join(loggedMessages(hook, logrus.InfoLevel), "\n"), "whose name a person removed")
	})
	t.Run("Subject", func(t *testing.T) {
		hook := captureLog(t)

		require.NoError(t, w.Audit(false, alice.SubjUID))
		assert.Contains(t, loggedMessages(hook, logrus.InfoLevel), "faces: found 1 marker whose name a person removed, left unnamed in a named cluster")
	})
	t.Run("Fix", func(t *testing.T) {
		hook := captureLog(t)

		require.NoError(t, w.Audit(true, ""))
		assert.Contains(t, loggedMessages(hook, logrus.InfoLevel), "faces: found 1 marker whose name a person removed, left unnamed in a named cluster")
		assert.NotContains(t, strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n"), rejected)

		m := entity.FindMarker(rejected)
		require.NotNil(t, m)
		assert.Equal(t, f.ID, m.FaceID, "--fix keeps the cluster it belongs to")
		assert.Empty(t, m.SubjUID)
		assert.NotNil(t, m.MatchedAt)
	})
}

// TestFaces_AuditXmpName pins that the audit leaves an XMP name in a cluster named after someone else
// as it is, since an XMP name labels only its own marker.
func TestFaces_AuditXmpName(t *testing.T) {
	w := isolatedTestFaces(t, "facesauditxmp")

	alice := consensusTestSubject(t, "Audit Xmp Alice")
	micha := consensusTestSubject(t, "Audit Xmp Micha")
	f := consensusTestFace(t, 1)
	require.NoError(t, f.Update("SubjUID", alice.SubjUID))
	xmp := consensusTestMarkers(t, f, 1, micha.SubjUID, entity.SrcXmp, false)[0]
	hook := captureLog(t)

	require.NoError(t, w.Audit(true, ""))
	assert.Contains(t, loggedMessages(hook, logrus.InfoLevel), "faces: found 1 marker with an XMP name in a cluster named after someone else")
	assert.NotContains(t, strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n"), xmp)

	m := entity.FindMarker(xmp)
	require.NotNil(t, m)
	assert.Equal(t, micha.SubjUID, m.SubjUID, "--fix keeps the XMP name")
	assert.Equal(t, entity.SrcXmp, m.SubjSrc)
	assert.Equal(t, f.ID, m.FaceID)
}
