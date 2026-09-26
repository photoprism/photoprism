package photoprism

import (
	"math"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// consensusTestFileUID is a fixture file, so a run does not remove the markers as orphans.
const consensusTestFileUID = "fs6sg6bw45bnlqdw"

// consensusTestSubject stores a person the consensus tests can name.
func consensusTestSubject(t *testing.T, name string) *entity.Subject {
	t.Helper()

	s := entity.NewSubject(name, entity.SubjPerson, entity.SrcManual)
	require.NotNil(t, s)
	require.NoError(t, s.Create())

	return s
}

// consensusTestFace stores an anonymous automatic cluster whose vector leans on its own axis.
func consensusTestFace(t *testing.T, axis int) *entity.Face {
	t.Helper()

	dims := face.ExpectedDims()
	require.Greater(t, dims, axis)

	v := make(face.Embedding, dims)
	v[0] = 1
	v[axis] = 0.5

	for i := range v {
		v[i] /= math.Sqrt(1.25)
	}

	f := entity.NewFace("", entity.SrcAuto, face.Embeddings{v}, face.EmbeddingModelName())
	require.NotNil(t, f)
	require.NoError(t, f.Create())

	return f
}

// consensusTestMarkers stores n matched face markers on a cluster and returns their UIDs.
func consensusTestMarkers(t *testing.T, f *entity.Face, n int, subjUID, subjSrc string, invalid bool) []string {
	t.Helper()

	uids := make([]string, 0, n)

	for range n {
		m := entity.Marker{
			MarkerUID:     rnd.GenerateUID('m'),
			FileUID:       consensusTestFileUID,
			MarkerType:    entity.MarkerFace,
			MarkerInvalid: invalid,
			MarkerReview:  true,
			SubjUID:       subjUID,
			SubjSrc:       subjSrc,
			FaceID:        f.ID,
			FaceDist:      0.1,
			EmbedModel:    f.EmbedModel,
			MatchedAt:     entity.TimeStamp(),
			W:             0.1,
			H:             0.1,
		}

		require.NoError(t, entity.UnscopedDb().Create(&m).Error)

		uids = append(uids, m.MarkerUID)
	}

	return uids
}

// consensusTestSubjects returns the subject of each marker, re-read from the database.
func consensusTestSubjects(t *testing.T, uids []string) []string {
	t.Helper()

	result := make([]string, len(uids))

	for i, uid := range uids {
		m := entity.FindMarker(uid)
		require.NotNil(t, m, uid)
		result[i] = m.SubjUID
	}

	return result
}

// consensusTestRepeat returns a slice holding n copies of s.
func consensusTestRepeat(s string, n int) []string {
	result := make([]string, n)

	for i := range result {
		result[i] = s
	}

	return result
}

func TestFaces_NameByConsensus(t *testing.T) {
	w := isolatedTestFaces(t, "facesconsensus")
	core := w.conf.FaceClusterCore()

	alice := consensusTestSubject(t, "Consensus Alice")
	bob := consensusTestSubject(t, "Consensus Bob")

	qualify := consensusTestFace(t, 1)
	qualifyAuto := consensusTestMarkers(t, qualify, core, alice.SubjUID, entity.SrcAuto, false)
	qualifyUnnamed := consensusTestMarkers(t, qualify, 2, "", entity.SrcAuto, false)
	qualifyInvalid := consensusTestMarkers(t, qualify, 1, "", entity.SrcAuto, true)

	belowCore := consensusTestFace(t, 2)
	consensusTestMarkers(t, belowCore, core-1, alice.SubjUID, entity.SrcAuto, false)
	belowCoreUnnamed := consensusTestMarkers(t, belowCore, 1, "", entity.SrcAuto, false)

	twoSubjects := consensusTestFace(t, 3)
	consensusTestMarkers(t, twoSubjects, core, alice.SubjUID, entity.SrcAuto, false)
	twoSubjectsBob := consensusTestMarkers(t, twoSubjects, 1, bob.SubjUID, entity.SrcAuto, false)
	twoSubjectsUnnamed := consensusTestMarkers(t, twoSubjects, 1, "", entity.SrcAuto, false)

	manual := consensusTestFace(t, 4)
	consensusTestMarkers(t, manual, core, alice.SubjUID, entity.SrcAuto, false)
	consensusTestMarkers(t, manual, 1, alice.SubjUID, entity.SrcManual, false)
	manualUnnamed := consensusTestMarkers(t, manual, 1, "", entity.SrcAuto, false)

	// Bob is confirmed by nobody, so an XMP name for him neither votes nor blocks.
	xmp := consensusTestFace(t, 5)
	consensusTestMarkers(t, xmp, core, alice.SubjUID, entity.SrcAuto, false)
	xmpBob := consensusTestMarkers(t, xmp, 1, bob.SubjUID, entity.SrcXmp, false)
	xmpUnnamed := consensusTestMarkers(t, xmp, 1, "", entity.SrcAuto, false)

	// Carol is verified, so XMP names for her vote, alone or against another person.
	carol := consensusTestSubject(t, "Consensus Carol")
	require.NoError(t, entity.UnscopedDb().Model(carol).UpdateColumn("verified", true).Error)
	xmpVote := consensusTestFace(t, 8)
	consensusTestMarkers(t, xmpVote, core-1, carol.SubjUID, entity.SrcAuto, false)
	xmpVoters := consensusTestMarkers(t, xmpVote, 1, carol.SubjUID, entity.SrcXmp, false)
	xmpOther := consensusTestFace(t, 9)
	consensusTestMarkers(t, xmpOther, core, alice.SubjUID, entity.SrcAuto, false)
	xmpOtherVoter := consensusTestMarkers(t, xmpOther, 1, carol.SubjUID, entity.SrcXmp, false)

	minority := consensusTestFace(t, 6)
	consensusTestMarkers(t, minority, core, alice.SubjUID, entity.SrcAuto, false)
	minorityUnnamed := consensusTestMarkers(t, minority, core+1, "", entity.SrcAuto, false)

	result, err := w.NameByConsensus()
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, FacesConsensusResult{Named: 3, Updated: 3}, result)

		f := entity.FindFace(qualify.ID)
		require.NotNil(t, f)
		assert.Equal(t, alice.SubjUID, f.SubjUID)
		assert.Equal(t, entity.SrcAuto, f.FaceSrc, "the name stays automatic")

		assert.Equal(t, consensusTestRepeat(alice.SubjUID, core), consensusTestSubjects(t, qualifyAuto))
		assert.Equal(t, consensusTestRepeat(alice.SubjUID, 2), consensusTestSubjects(t, qualifyUnnamed))
		assert.Equal(t, []string{""}, consensusTestSubjects(t, qualifyInvalid), "an invalid marker is left alone")

		for _, uid := range qualifyUnnamed {
			m := entity.FindMarker(uid)
			require.NotNil(t, m)
			assert.Equal(t, entity.SrcAuto, m.SubjSrc)
			assert.False(t, m.MarkerReview)
		}
	})
	t.Run("XmpNeutral", func(t *testing.T) {
		assert.Equal(t, alice.SubjUID, entity.FindFace(xmp.ID).SubjUID)
		assert.Equal(t, []string{alice.SubjUID}, consensusTestSubjects(t, xmpUnnamed))

		m := entity.FindMarker(xmpBob[0])
		require.NotNil(t, m)
		assert.Equal(t, bob.SubjUID, m.SubjUID, "the XMP marker keeps its own link")
		assert.Equal(t, entity.SrcXmp, m.SubjSrc)
	})
	t.Run("XmpVote", func(t *testing.T) {
		assert.Equal(t, carol.SubjUID, entity.FindFace(xmpVote.ID).SubjUID)

		m := entity.FindMarker(xmpVoters[0])
		require.NotNil(t, m)
		assert.Equal(t, entity.SrcXmp, m.SubjSrc, "the XMP marker keeps its source")
	})
	t.Run("Unchanged", func(t *testing.T) {
		for name, tc := range map[string]struct {
			face    *entity.Face
			markers []string
			want    []string
		}{
			"BelowCore":   {belowCore, belowCoreUnnamed, []string{""}},
			"TwoSubjects": {twoSubjects, append(twoSubjectsBob, twoSubjectsUnnamed...), []string{bob.SubjUID, ""}},
			"Manual":      {manual, manualUnnamed, []string{""}},
			"XmpOther":    {xmpOther, xmpOtherVoter, []string{carol.SubjUID}},
			"Minority":    {minority, minorityUnnamed, consensusTestRepeat("", core+1)},
		} {
			f := entity.FindFace(tc.face.ID)
			require.NotNil(t, f, name)
			assert.Empty(t, f.SubjUID, name)
			assert.Equal(t, tc.want, consensusTestSubjects(t, tc.markers), name)
		}
	})
	t.Run("Idempotent", func(t *testing.T) {
		again, err := w.NameByConsensus()
		require.NoError(t, err)
		assert.Equal(t, FacesConsensusResult{}, again)
	})
}

func TestFaces_nameByConsensus(t *testing.T) {
	w := isolatedTestFaces(t, "facesconsensusgate")
	core := w.conf.FaceClusterCore()

	alice := consensusTestSubject(t, "Consensus Gate Alice")
	f := consensusTestFace(t, 1)
	consensusTestMarkers(t, f, core, alice.SubjUID, entity.SrcAuto, false)
	unnamed := consensusTestMarkers(t, f, 1, "", entity.SrcAuto, false)

	t.Run("MatchingFailed", func(t *testing.T) {
		assert.Zero(t, w.nameByConsensus(false))

		found := entity.FindFace(f.ID)
		require.NotNil(t, found)
		assert.Empty(t, found.SubjUID)
		assert.Equal(t, []string{""}, consensusTestSubjects(t, unnamed))
	})
	t.Run("Matched", func(t *testing.T) {
		assert.Equal(t, 1, w.nameByConsensus(true))

		found := entity.FindFace(f.ID)
		require.NotNil(t, found)
		assert.Equal(t, alice.SubjUID, found.SubjUID)
		assert.Equal(t, []string{alice.SubjUID}, consensusTestSubjects(t, unnamed))
	})
}

func TestClaimConsensusFace(t *testing.T) {
	_ = isolatedTestFaces(t, "facesconsensusclaim")

	alice := consensusTestSubject(t, "Claim Alice")
	bob := consensusTestSubject(t, "Claim Bob")

	t.Run("Success", func(t *testing.T) {
		f := consensusTestFace(t, 1)

		claimed, err := claimConsensusFace(f.ID, alice.SubjUID)
		require.NoError(t, err)
		assert.True(t, claimed)
		assert.Equal(t, alice.SubjUID, entity.FindFace(f.ID).SubjUID)
	})
	t.Run("AlreadyNamed", func(t *testing.T) {
		f := consensusTestFace(t, 2)
		require.NoError(t, f.Update("SubjUID", bob.SubjUID))

		claimed, err := claimConsensusFace(f.ID, alice.SubjUID)
		require.NoError(t, err)
		assert.False(t, claimed)
		assert.Equal(t, bob.SubjUID, entity.FindFace(f.ID).SubjUID)
	})
	t.Run("Changed", func(t *testing.T) {
		hidden := consensusTestFace(t, 3)
		require.NoError(t, hidden.Hide())
		ambiguous := consensusTestFace(t, 4)
		require.NoError(t, ambiguous.Update("FaceKind", int(face.AmbiguousFace)))
		manual := consensusTestFace(t, 5)
		require.NoError(t, manual.Update("FaceSrc", entity.SrcManual))

		for name, f := range map[string]*entity.Face{"Hidden": hidden, "Ambiguous": ambiguous, "ManualSrc": manual} {
			claimed, err := claimConsensusFace(f.ID, alice.SubjUID)
			require.NoError(t, err, name)
			assert.False(t, claimed, name)
			assert.Empty(t, entity.FindFace(f.ID).SubjUID, name)
		}
	})
	t.Run("PersonDeleted", func(t *testing.T) {
		gone := consensusTestSubject(t, "Claim Gone")
		require.NoError(t, entity.UnscopedDb().Model(gone).UpdateColumn("deleted_at", entity.Now()).Error)
		f := consensusTestFace(t, 6)

		claimed, err := claimConsensusFace(f.ID, gone.SubjUID)
		require.NoError(t, err)
		assert.False(t, claimed)
		assert.Empty(t, entity.FindFace(f.ID).SubjUID)
	})
	t.Run("NotAPerson", func(t *testing.T) {
		pet := consensusTestSubject(t, "Claim Pet")
		require.NoError(t, entity.UnscopedDb().Model(pet).UpdateColumn("subj_type", "pet").Error)
		f := consensusTestFace(t, 7)

		claimed, err := claimConsensusFace(f.ID, pet.SubjUID)
		require.NoError(t, err)
		assert.False(t, claimed)
		assert.Empty(t, entity.FindFace(f.ID).SubjUID)
	})
	t.Run("NotFound", func(t *testing.T) {
		claimed, err := claimConsensusFace("NOTACLUSTER", alice.SubjUID)
		require.NoError(t, err)
		assert.False(t, claimed)
	})
}

// TestFaces_startNamesByConsensus pins that a worker run reaches the pass and reports it.
func TestFaces_startNamesByConsensus(t *testing.T) {
	w := isolatedTestFaces(t, "facesconsensusstart")
	core := w.conf.FaceClusterCore()

	// A first run settles what the fixtures leave for it, such as orphan markers.
	_, err := w.start(FacesOptions{Threshold: 1000000})
	require.NoError(t, err)

	alice := consensusTestSubject(t, "Consensus Start Alice")
	f := consensusTestFace(t, 1)
	consensusTestMarkers(t, f, core, alice.SubjUID, entity.SrcAuto, false)
	unnamed := consensusTestMarkers(t, f, 1, "", entity.SrcAuto, false)
	require.NoError(t, entity.UnscopedDb().Model(alice).UpdateColumns(entity.Values{"file_count": 0, "photo_count": 0}).Error)

	// A threshold no library reaches, so clustering stays out of the way.
	result, err := w.start(FacesOptions{Threshold: 1000000})
	require.NoError(t, err)

	assert.Equal(t, 1, result.Named)
	assert.True(t, result.Moved())

	// Nothing else in this run moved a marker, so only the naming can have refreshed the counts.
	subj := entity.FindSubject(alice.SubjUID)
	require.NotNil(t, subj)
	assert.Equal(t, 1, subj.FileCount)

	found := entity.FindFace(f.ID)
	require.NotNil(t, found)
	assert.Equal(t, alice.SubjUID, found.SubjUID)
	assert.Equal(t, []string{alice.SubjUID}, consensusTestSubjects(t, unnamed))
}

func TestFaces_nameConsensusFaces(t *testing.T) {
	w := isolatedTestFaces(t, "facesconsensusname")
	core := w.conf.FaceClusterCore()

	alice := consensusTestSubject(t, "Consensus Name Alice")
	bob := consensusTestSubject(t, "Consensus Name Bob")

	t.Run("NamedMeanwhile", func(t *testing.T) {
		// A person names the cluster after the counts were read, which the claim has to see.
		f := consensusTestFace(t, 1)
		consensusTestMarkers(t, f, core, alice.SubjUID, entity.SrcAuto, false)
		unnamed := consensusTestMarkers(t, f, 1, "", entity.SrcAuto, false)
		candidates := []query.FaceConsensus{{FaceID: f.ID, SubjUID: alice.SubjUID, Votes: core, Unnamed: 1, Valid: core + 1}}
		require.NoError(t, f.Update("SubjUID", bob.SubjUID))

		result, err := w.nameConsensusFaces(candidates)
		require.NoError(t, err)
		assert.Equal(t, FacesConsensusResult{}, result)
		assert.Equal(t, bob.SubjUID, entity.FindFace(f.ID).SubjUID)
		assert.Equal(t, []string{""}, consensusTestSubjects(t, unnamed))
	})
	t.Run("Missing", func(t *testing.T) {
		result, err := w.nameConsensusFaces([]query.FaceConsensus{{FaceID: "NOTACLUSTER", SubjUID: alice.SubjUID}})
		require.NoError(t, err)
		assert.Equal(t, FacesConsensusResult{}, result)
	})
}

func TestFaces_NameByConsensusCanceled(t *testing.T) {
	w := isolatedTestFaces(t, "facesconsensuscanceled")
	core := w.conf.FaceClusterCore()

	alice := consensusTestSubject(t, "Consensus Canceled Alice")
	f := consensusTestFace(t, 1)
	consensusTestMarkers(t, f, core, alice.SubjUID, entity.SrcAuto, false)
	candidates := []query.FaceConsensus{{FaceID: f.ID, SubjUID: alice.SubjUID, Votes: core, Valid: core}}

	require.NoError(t, mutex.FacesWorker.Start())
	t.Cleanup(mutex.FacesWorker.Stop)
	mutex.FacesWorker.Cancel()

	t.Run("BeforeCounting", func(t *testing.T) {
		result, err := w.NameByConsensus()
		require.NoError(t, err)
		assert.Equal(t, FacesConsensusResult{}, result)
	})
	t.Run("BeforeNaming", func(t *testing.T) {
		result, err := w.nameConsensusFaces(candidates)
		require.NoError(t, err)
		assert.Equal(t, FacesConsensusResult{}, result)
	})

	assert.Empty(t, entity.FindFace(f.ID).SubjUID)
}

func TestFaces_NameByConsensusError(t *testing.T) {
	w := isolatedTestFaces(t, "facesconsensuserror")

	// The counting query reads the faces table, so a missing one fails it.
	require.NoError(t, entity.Db().Exec("ALTER TABLE faces RENAME TO faces_consensus_error").Error)
	t.Cleanup(func() {
		require.NoError(t, entity.Db().Exec("ALTER TABLE faces_consensus_error RENAME TO faces").Error)
	})

	t.Run("NameByConsensus", func(t *testing.T) {
		_, err := w.NameByConsensus()
		assert.Error(t, err)
	})
	t.Run("nameByConsensus", func(t *testing.T) {
		assert.Zero(t, w.nameByConsensus(true))
	})
}

func TestFaces_auditConsensus(t *testing.T) {
	w := isolatedTestFaces(t, "facesconsensusaudit")
	core := w.conf.FaceClusterCore()

	alice := consensusTestSubject(t, "Consensus Audit Alice")
	bob := consensusTestSubject(t, "Consensus Audit Bob")
	f := consensusTestFace(t, 1)
	consensusTestMarkers(t, f, core, alice.SubjUID, entity.SrcAuto, false)
	unnamed := consensusTestMarkers(t, f, 1, "", entity.SrcAuto, false)
	split := consensusTestFace(t, 2)
	consensusTestMarkers(t, split, core, alice.SubjUID, entity.SrcAuto, false)
	consensusTestMarkers(t, split, 1, bob.SubjUID, entity.SrcAuto, false)
	belowCore := consensusTestFace(t, 3)
	consensusTestMarkers(t, belowCore, core-1, alice.SubjUID, entity.SrcAuto, false)
	gone := consensusTestSubject(t, "Consensus Audit Gone")
	consensusTestMarkers(t, consensusTestFace(t, 4), core, gone.SubjUID, entity.SrcAuto, false)
	require.NoError(t, entity.UnscopedDb().Model(gone).UpdateColumn("deleted_at", entity.Now()).Error)

	t.Run("Success", func(t *testing.T) {
		hook := captureLog(t)

		n, err := w.auditConsensus("")
		require.NoError(t, err)
		assert.Equal(t, 1, n)
		assert.Contains(t, loggedMessages(hook, logrus.InfoLevel),
			"faces: found 1 unnamed cluster whose recognized faces agree on one person, which a completed recognition run names")
	})
	t.Run("Subject", func(t *testing.T) {
		n, err := w.auditConsensus(alice.SubjUID)
		require.NoError(t, err)
		assert.Equal(t, 1, n)
	})
	t.Run("OtherSubject", func(t *testing.T) {
		hook := captureLog(t)

		n, err := w.auditConsensus(bob.SubjUID)
		require.NoError(t, err)
		assert.Zero(t, n)
		assert.Contains(t, loggedMessages(hook, logrus.InfoLevel), "faces: found no unnamed clusters whose recognized faces agree on "+entity.SubjNames.Log(bob.SubjUID))
	})
	t.Run("AuditFixDoesNotName", func(t *testing.T) {
		hook := captureLog(t)

		require.NoError(t, w.Audit(true, ""))
		assert.Contains(t, loggedMessages(hook, logrus.InfoLevel),
			"faces: found 1 unnamed cluster whose recognized faces agree on one person, which a completed recognition run names")
		assert.Empty(t, entity.FindFace(f.ID).SubjUID)
		assert.Equal(t, []string{""}, consensusTestSubjects(t, unnamed))
	})
}

func TestFaces_auditConsensusError(t *testing.T) {
	w := isolatedTestFaces(t, "facesconsensusauditerror")

	require.NoError(t, entity.Db().Exec("ALTER TABLE faces RENAME TO faces_consensus_error").Error)
	t.Cleanup(func() {
		require.NoError(t, entity.Db().Exec("ALTER TABLE faces_consensus_error RENAME TO faces").Error)
	})

	n, err := w.auditConsensus("")
	assert.Error(t, err)
	assert.Zero(t, n)
}

// TestFaces_ResetClearsConsensusName pins that a default reset removes what the pass named, since
// the name keeps the automatic source.
func TestFaces_ResetClearsConsensusName(t *testing.T) {
	w := isolatedTestFaces(t, "facesconsensusreset")
	core := w.conf.FaceClusterCore()

	alice := consensusTestSubject(t, "Consensus Reset Alice")
	f := consensusTestFace(t, 1)
	named := consensusTestMarkers(t, f, core, alice.SubjUID, entity.SrcAuto, false)
	unnamed := consensusTestMarkers(t, f, 1, "", entity.SrcAuto, false)
	manual := consensusTestMarkers(t, consensusTestFace(t, 2), 1, alice.SubjUID, entity.SrcManual, false)

	result, err := w.NameByConsensus()
	require.NoError(t, err)
	require.Equal(t, 1, result.Named)
	require.Equal(t, []string{alice.SubjUID}, consensusTestSubjects(t, unnamed))

	require.NoError(t, w.Reset())

	assert.Nil(t, entity.FindFace(f.ID), "the cluster is removed")
	assert.Equal(t, consensusTestRepeat("", core+1), consensusTestSubjects(t, append(named, unnamed...)))

	for _, uid := range append(named, unnamed...) {
		m := entity.FindMarker(uid)
		require.NotNil(t, m)
		assert.Empty(t, m.FaceID)
	}

	assert.Equal(t, []string{alice.SubjUID}, consensusTestSubjects(t, manual), "a name a person gave is kept")
}
