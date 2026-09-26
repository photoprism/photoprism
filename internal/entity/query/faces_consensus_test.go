package query

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// consensusTestCore is the minimum number of votes the tests below pass.
const consensusTestCore = 5

// consensusTestFace stores an anonymous automatic cluster whose vector leans on its own axis, so
// every case hashes to a distinct id.
func consensusTestFace(t *testing.T, axis int) *entity.Face {
	t.Helper()

	dims := face.ExpectedDims()
	require.Greater(t, dims, axis)

	v := make(face.Embedding, dims)
	v[0] = 1
	v[axis] = 0.5

	norm := math.Sqrt(1.25)

	for i := range v {
		v[i] /= norm
	}

	f := entity.NewFace("", entity.SrcAuto, face.Embeddings{v}, face.EmbeddingModelName())
	require.NotNil(t, f)
	require.NotEmpty(t, f.ID)
	require.NoError(t, f.Create())

	t.Cleanup(func() {
		UnscopedDb().Delete(entity.Face{}, "id = ?", f.ID)
	})

	return f
}

// consensusTestMarkers stores n face markers on a cluster with the given subject and source.
func consensusTestMarkers(t *testing.T, f *entity.Face, n int, subjUID, subjSrc string, invalid bool, model string) {
	t.Helper()

	for range n {
		m := entity.Marker{
			MarkerUID:     rnd.GenerateUID('m'),
			FileUID:       rnd.GenerateUID(entity.FileUID),
			MarkerType:    entity.MarkerFace,
			MarkerInvalid: invalid,
			SubjUID:       subjUID,
			SubjSrc:       subjSrc,
			FaceID:        f.ID,
			FaceDist:      0.1,
			EmbedModel:    model,
			W:             0.1,
			H:             0.1,
		}

		require.NoError(t, UnscopedDb().Create(&m).Error)

		t.Cleanup(func() {
			UnscopedDb().Delete(entity.Marker{}, "marker_uid = ?", m.MarkerUID)
		})
	}
}

// foreignEmbeddingModel returns a model whose vectors the configured one cannot compare.
func foreignEmbeddingModel(t *testing.T) string {
	t.Helper()

	for _, name := range []string{face.ModelArcFaceR50, face.ModelSFace} {
		if !face.SameEmbeddingSpace(name, face.EmbeddingModelName()) {
			return name
		}
	}

	t.Fatal("no foreign embedding model")

	return ""
}

// findConsensus returns the counts reported for a cluster, or nil.
func findConsensus(counts []FaceConsensus, faceID string) *FaceConsensus {
	for i := range counts {
		if counts[i].FaceID == faceID {
			return &counts[i]
		}
	}

	return nil
}

func TestFaceConsensus_Agrees(t *testing.T) {
	agrees := FaceConsensus{FaceID: "f", SubjUID: "s", Votes: 5, Matched: 1, Unnamed: 5, Valid: 10}

	t.Run("Success", func(t *testing.T) {
		assert.True(t, agrees.Agrees(5))
	})
	t.Run("BelowCore", func(t *testing.T) {
		assert.False(t, agrees.Agrees(6))
	})
	t.Run("CoreBelowOne", func(t *testing.T) {
		assert.False(t, FaceConsensus{SubjUID: "s"}.Agrees(0))
		assert.True(t, FaceConsensus{SubjUID: "s", Votes: 1, Matched: 1, Valid: 1}.Agrees(0))
	})
	t.Run("Split", func(t *testing.T) {
		c := agrees
		c.Split = true
		assert.False(t, c.Agrees(5))
	})
	t.Run("Asserted", func(t *testing.T) {
		c := agrees
		c.Asserted = 1
		assert.False(t, c.Agrees(5))
	})
	t.Run("Minority", func(t *testing.T) {
		c := agrees
		c.Valid = 11
		assert.False(t, c.Agrees(5))
	})
	t.Run("XmpVotesAlone", func(t *testing.T) {
		c := agrees
		c.Matched = 0
		assert.True(t, c.Agrees(5), "XMP votes alone suffice")
	})
	t.Run("NoSubject", func(t *testing.T) {
		c := agrees
		c.SubjUID = ""
		assert.False(t, c.Agrees(5))
	})
}

// consensusTestModel configures FaceNet for the duration of a test, so the embedding model gate
// is exercised rather than skipped.
func consensusTestModel(t *testing.T) {
	t.Helper()

	restore := face.ConfiguredModel()
	require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{
		Name:  face.ModelFaceNet,
		Model: face.FindEmbeddingModel(face.ModelFaceNet),
	}))
	t.Cleanup(func() {
		_ = face.ConfigureEmbedder(face.EmbedderSettings{Name: restore, Model: face.FindEmbeddingModel(restore)})
	})
	require.Equal(t, face.ModelFaceNet, face.EmbeddingModelName())
}

// consensusFixtures holds one cluster per eligibility branch.
type consensusFixtures struct {
	alice, bob                                                   *entity.Subject
	qualify, multiRow, belowCore, twoSubjects, manual, xmp       *entity.Face
	invalidManual, minority, hidden, deleted, nonPerson, named   *entity.Face
	foreignFace, foreignVotes, foreignSplit, manualSrc           *entity.Face
	ambiguous, unclassified, foreignManual                       *entity.Face
	xmpUnconfirmed, xmpUnlinked, xmpOther, xmpMinority, xmpAlone *entity.Face
	xmpDeleted, xmpPet, xmpCorroborated, xmpForeignMatch         *entity.Face
}

// newConsensusFixtures stores the clusters of consensusFixtures and removes them after the test.
func newConsensusFixtures(t *testing.T) (fx consensusFixtures) {
	t.Helper()

	consensusTestModel(t)

	fx.alice = conflictTestSubject(t, "Consensus Alice")
	fx.bob = conflictTestSubject(t, "Consensus Bob")
	gone := conflictTestSubject(t, "Consensus Gone")
	require.NoError(t, UnscopedDb().Model(gone).UpdateColumn("deleted_at", entity.Now()).Error)
	pet := conflictTestSubject(t, "Consensus Pet")
	require.NoError(t, UnscopedDb().Model(pet).UpdateColumn("subj_type", "pet").Error)

	model := face.EmbeddingModelName()
	foreign := foreignEmbeddingModel(t)
	core := consensusTestCore
	alice, bob := fx.alice.SubjUID, fx.bob.SubjUID

	fx.qualify = consensusTestFace(t, 1)
	consensusTestMarkers(t, fx.qualify, core, alice, entity.SrcAuto, false, model)
	consensusTestMarkers(t, fx.qualify, 2, "", entity.SrcAuto, false, model)
	consensusTestMarkers(t, fx.qualify, 3, bob, entity.SrcAuto, true, model)

	// Rows are ordered by marker model, so the foreign row without a name comes last and would
	// clear the subject if a row could overwrite it.
	fx.multiRow = consensusTestFace(t, 2)
	consensusTestMarkers(t, fx.multiRow, core, alice, entity.SrcAuto, false, "")
	consensusTestMarkers(t, fx.multiRow, 1, "", entity.SrcAuto, false, foreign)

	fx.belowCore = consensusTestFace(t, 3)
	consensusTestMarkers(t, fx.belowCore, core-1, alice, entity.SrcAuto, false, model)

	fx.twoSubjects = consensusTestFace(t, 4)
	consensusTestMarkers(t, fx.twoSubjects, core, alice, entity.SrcAuto, false, model)
	consensusTestMarkers(t, fx.twoSubjects, 1, bob, entity.SrcAuto, false, model)

	fx.manual = consensusTestFace(t, 5)
	consensusTestMarkers(t, fx.manual, core, alice, entity.SrcAuto, false, model)
	consensusTestMarkers(t, fx.manual, 1, alice, entity.SrcManual, false, model)

	fx.xmp = consensusTestFace(t, 6)
	consensusTestMarkers(t, fx.xmp, core, alice, entity.SrcAuto, false, model)
	consensusTestMarkers(t, fx.xmp, 1, alice, entity.SrcXmp, false, model)
	consensusTestMarkers(t, fx.xmp, 1, alice, entity.SrcXmp, true, model)

	fx.invalidManual = consensusTestFace(t, 7)
	consensusTestMarkers(t, fx.invalidManual, core, alice, entity.SrcAuto, false, model)
	consensusTestMarkers(t, fx.invalidManual, 1, bob, entity.SrcManual, true, model)

	fx.minority = consensusTestFace(t, 8)
	consensusTestMarkers(t, fx.minority, core, alice, entity.SrcAuto, false, model)
	consensusTestMarkers(t, fx.minority, core+1, "", entity.SrcAuto, false, model)

	fx.hidden = consensusTestFace(t, 9)
	consensusTestMarkers(t, fx.hidden, core, alice, entity.SrcAuto, false, model)
	require.NoError(t, fx.hidden.Hide())

	fx.deleted = consensusTestFace(t, 10)
	consensusTestMarkers(t, fx.deleted, core, gone.SubjUID, entity.SrcAuto, false, model)

	fx.nonPerson = consensusTestFace(t, 11)
	consensusTestMarkers(t, fx.nonPerson, core, pet.SubjUID, entity.SrcAuto, false, model)

	fx.named = consensusTestFace(t, 12)
	consensusTestMarkers(t, fx.named, core, bob, entity.SrcAuto, false, model)
	require.NoError(t, fx.named.Update("SubjUID", alice))

	fx.foreignFace = consensusTestFace(t, 13)
	consensusTestMarkers(t, fx.foreignFace, core, alice, entity.SrcAuto, false, model)
	require.NoError(t, fx.foreignFace.Update("EmbedModel", foreign))

	fx.foreignVotes = consensusTestFace(t, 14)
	consensusTestMarkers(t, fx.foreignVotes, core-1, alice, entity.SrcAuto, false, model)
	consensusTestMarkers(t, fx.foreignVotes, 2, alice, entity.SrcAuto, false, foreign)

	fx.foreignSplit = consensusTestFace(t, 15)
	consensusTestMarkers(t, fx.foreignSplit, core, alice, entity.SrcAuto, false, model)
	consensusTestMarkers(t, fx.foreignSplit, 1, bob, entity.SrcAuto, false, foreign)

	// People an XMP name may link to: one nobody confirmed, one a person verified without naming a
	// face, and one who is deleted although a person named a face of theirs.
	micha := conflictTestSubject(t, "Consensus Micha")
	carol := conflictTestSubject(t, "Consensus Carol")
	require.NoError(t, UnscopedDb().Model(carol).UpdateColumn("verified", true).Error)
	require.NoError(t, UnscopedDb().Model(fx.bob).UpdateColumn("verified", true).Error)
	erased := conflictTestSubject(t, "Consensus Erased")
	consensusTestMarkers(t, consensusTestFace(t, 36), 1, erased.SubjUID, entity.SrcManual, false, model)
	require.NoError(t, UnscopedDb().Model(erased).UpdateColumn("deleted_at", entity.Now()).Error)

	fx.xmpUnconfirmed = consensusTestFace(t, 30)
	consensusTestMarkers(t, fx.xmpUnconfirmed, core, alice, entity.SrcAuto, false, model)
	consensusTestMarkers(t, fx.xmpUnconfirmed, 1, micha.SubjUID, entity.SrcXmp, false, model)
	// The matcher assigned Micha elsewhere, which confirms nothing.
	consensusTestMarkers(t, consensusTestFace(t, 38), 1, micha.SubjUID, entity.SrcAuto, false, model)

	fx.xmpUnlinked = consensusTestFace(t, 31)
	consensusTestMarkers(t, fx.xmpUnlinked, core, alice, entity.SrcAuto, false, model)
	consensusTestMarkers(t, fx.xmpUnlinked, 1, "", entity.SrcXmp, false, model)

	fx.xmpOther = consensusTestFace(t, 32)
	consensusTestMarkers(t, fx.xmpOther, core, alice, entity.SrcAuto, false, model)
	consensusTestMarkers(t, fx.xmpOther, 1, bob, entity.SrcXmp, false, model)

	fx.xmpMinority = consensusTestFace(t, 33)
	consensusTestMarkers(t, fx.xmpMinority, core, alice, entity.SrcAuto, false, model)
	consensusTestMarkers(t, fx.xmpMinority, core+1, micha.SubjUID, entity.SrcXmp, false, model)

	fx.xmpAlone = consensusTestFace(t, 34)
	consensusTestMarkers(t, fx.xmpAlone, core, carol.SubjUID, entity.SrcXmp, false, model)

	// The only matcher vote comes from another embedding space, so it is not counted as one.
	fx.xmpForeignMatch = consensusTestFace(t, 40)
	consensusTestMarkers(t, fx.xmpForeignMatch, core, carol.SubjUID, entity.SrcXmp, false, model)
	consensusTestMarkers(t, fx.xmpForeignMatch, 1, carol.SubjUID, entity.SrcAuto, false, foreign)

	fx.xmpCorroborated = consensusTestFace(t, 39)
	consensusTestMarkers(t, fx.xmpCorroborated, core-1, carol.SubjUID, entity.SrcXmp, false, model)
	consensusTestMarkers(t, fx.xmpCorroborated, 1, carol.SubjUID, entity.SrcAuto, false, model)

	fx.xmpDeleted = consensusTestFace(t, 35)
	consensusTestMarkers(t, fx.xmpDeleted, core, alice, entity.SrcAuto, false, model)
	consensusTestMarkers(t, fx.xmpDeleted, 1, erased.SubjUID, entity.SrcXmp, false, model)

	// A verified subject that is not a person gets no vote either.
	require.NoError(t, UnscopedDb().Model(pet).UpdateColumn("verified", true).Error)
	fx.xmpPet = consensusTestFace(t, 37)
	consensusTestMarkers(t, fx.xmpPet, core, alice, entity.SrcAuto, false, model)
	consensusTestMarkers(t, fx.xmpPet, 1, pet.SubjUID, entity.SrcXmp, false, model)

	fx.foreignManual = consensusTestFace(t, 19)
	consensusTestMarkers(t, fx.foreignManual, core, alice, entity.SrcAuto, false, model)
	consensusTestMarkers(t, fx.foreignManual, 1, alice, entity.SrcManual, false, foreign)

	fx.manualSrc = consensusTestFace(t, 16)
	consensusTestMarkers(t, fx.manualSrc, core, alice, entity.SrcAuto, false, model)
	require.NoError(t, fx.manualSrc.Update("FaceSrc", entity.SrcManual))

	fx.ambiguous = consensusTestFace(t, 17)
	consensusTestMarkers(t, fx.ambiguous, core, alice, entity.SrcAuto, false, model)
	require.NoError(t, fx.ambiguous.Update("FaceKind", int(face.AmbiguousFace)))

	fx.unclassified = consensusTestFace(t, 18)
	consensusTestMarkers(t, fx.unclassified, core, alice, entity.SrcAuto, false, model)
	require.NoError(t, fx.unclassified.Update("FaceKind", int(face.UnclassifiedFace)))

	return fx
}

func TestAnonymousFaceConsensus(t *testing.T) {
	fx := newConsensusFixtures(t)
	core := consensusTestCore
	alice := fx.alice.SubjUID

	counts, err := AnonymousFaceConsensus()
	require.NoError(t, err)

	t.Run("Qualify", func(t *testing.T) {
		c := findConsensus(counts, fx.qualify.ID)
		require.NotNil(t, c)
		assert.Equal(t, FaceConsensus{FaceID: fx.qualify.ID, SubjUID: alice, Votes: core, Matched: core, Unnamed: 2, Valid: core + 2}, *c)
	})
	t.Run("MultiRow", func(t *testing.T) {
		c := findConsensus(counts, fx.multiRow.ID)
		require.NotNil(t, c)
		assert.Equal(t, FaceConsensus{FaceID: fx.multiRow.ID, SubjUID: alice, Votes: core, Matched: core, Unnamed: 1, Valid: core + 1}, *c)
	})
	t.Run("BelowCore", func(t *testing.T) {
		c := findConsensus(counts, fx.belowCore.ID)
		require.NotNil(t, c)
		assert.Equal(t, core-1, c.Votes)
	})
	t.Run("TwoSubjects", func(t *testing.T) {
		c := findConsensus(counts, fx.twoSubjects.ID)
		require.NotNil(t, c)
		assert.True(t, c.Split)
	})
	t.Run("Manual", func(t *testing.T) {
		c := findConsensus(counts, fx.manual.ID)
		require.NotNil(t, c)
		assert.Equal(t, 1, c.Asserted)
	})
	t.Run("XmpConfirmedVote", func(t *testing.T) {
		// Alice is confirmed by a name a person gave one of her faces, so her XMP name votes.
		c := findConsensus(counts, fx.xmp.ID)
		require.NotNil(t, c)
		assert.Equal(t, core+1, c.Votes)
		assert.Zero(t, c.Asserted, "an XMP name is not a veto")
		assert.False(t, c.Split)
	})
	t.Run("XmpNeutral", func(t *testing.T) {
		for name, f := range map[string]*entity.Face{"Unconfirmed": fx.xmpUnconfirmed, "Unlinked": fx.xmpUnlinked, "Deleted": fx.xmpDeleted, "Pet": fx.xmpPet} {
			c := findConsensus(counts, f.ID)
			require.NotNil(t, c, name)
			assert.Equal(t, FaceConsensus{FaceID: f.ID, SubjUID: fx.alice.SubjUID, Votes: core, Matched: core, Valid: core + 1}, *c, name)
		}
	})
	t.Run("XmpOtherPerson", func(t *testing.T) {
		c := findConsensus(counts, fx.xmpOther.ID)
		require.NotNil(t, c)
		assert.True(t, c.Split)
	})
	t.Run("XmpAlone", func(t *testing.T) {
		// Carol is confirmed by the verified flag alone, and no vote comes from the matcher.
		c := findConsensus(counts, fx.xmpAlone.ID)
		require.NotNil(t, c)
		assert.Equal(t, core, c.Votes)
		assert.Zero(t, c.Matched)
		assert.True(t, c.Agrees(core))
	})
	t.Run("XmpForeignMatch", func(t *testing.T) {
		c := findConsensus(counts, fx.xmpForeignMatch.ID)
		require.NotNil(t, c)
		assert.Equal(t, core, c.Votes)
		assert.Zero(t, c.Matched)
	})
	t.Run("XmpCorroborated", func(t *testing.T) {
		c := findConsensus(counts, fx.xmpCorroborated.ID)
		require.NotNil(t, c)
		assert.Equal(t, core, c.Votes)
		assert.Equal(t, 1, c.Matched)
		assert.True(t, c.Agrees(core))
	})
	t.Run("InvalidManual", func(t *testing.T) {
		c := findConsensus(counts, fx.invalidManual.ID)
		require.NotNil(t, c)
		assert.Equal(t, 1, c.Asserted, "a name a person gave counts on an invalid marker too")
		assert.Equal(t, core, c.Valid)
	})
	t.Run("ForeignManual", func(t *testing.T) {
		c := findConsensus(counts, fx.foreignManual.ID)
		require.NotNil(t, c)
		assert.Equal(t, 1, c.Asserted, "a name a person gave counts from any embedding space")
	})
	t.Run("Minority", func(t *testing.T) {
		c := findConsensus(counts, fx.minority.ID)
		require.NotNil(t, c)
		assert.Equal(t, 2*core+1, c.Valid)
	})
	t.Run("ForeignModelVotes", func(t *testing.T) {
		c := findConsensus(counts, fx.foreignVotes.ID)
		require.NotNil(t, c)
		assert.Equal(t, core-1, c.Votes)
		assert.False(t, c.Split)
		assert.Equal(t, core+1, c.Valid)
	})
	t.Run("ForeignModelSplit", func(t *testing.T) {
		c := findConsensus(counts, fx.foreignSplit.ID)
		require.NotNil(t, c)
		assert.True(t, c.Split, "a foreign name is still one SetSubjectUID would overwrite")
	})
	t.Run("Excluded", func(t *testing.T) {
		for name, f := range map[string]*entity.Face{
			"Hidden":       fx.hidden,
			"Named":        fx.named,
			"ForeignModel": fx.foreignFace,
			"ManualSrc":    fx.manualSrc,
			"Ambiguous":    fx.ambiguous,
			"Unclassified": fx.unclassified,
		} {
			assert.Nil(t, findConsensus(counts, f.ID), name)
		}
	})
}

func TestConsensusFaces(t *testing.T) {
	fx := newConsensusFixtures(t)

	t.Run("Success", func(t *testing.T) {
		result, err := ConsensusFaces(consensusTestCore)
		require.NoError(t, err)

		for name, f := range map[string]*entity.Face{
			"Qualify":         fx.qualify,
			"MultiRow":        fx.multiRow,
			"Xmp":             fx.xmp,
			"XmpUnconfirmed":  fx.xmpUnconfirmed,
			"XmpUnlinked":     fx.xmpUnlinked,
			"XmpDeleted":      fx.xmpDeleted,
			"XmpCorroborated": fx.xmpCorroborated,
			"XmpAlone":        fx.xmpAlone,
		} {
			assert.NotNil(t, findConsensus(result, f.ID), name)
		}

		for name, f := range map[string]*entity.Face{
			"BelowCore":     fx.belowCore,
			"TwoSubjects":   fx.twoSubjects,
			"Manual":        fx.manual,
			"XmpOther":      fx.xmpOther,
			"XmpMinority":   fx.xmpMinority,
			"InvalidManual": fx.invalidManual,
			"Minority":      fx.minority,
			"Hidden":        fx.hidden,
			"Deleted":       fx.deleted,
			"NonPerson":     fx.nonPerson,
			"Named":         fx.named,
			"ForeignModel":  fx.foreignFace,
			"ForeignVotes":  fx.foreignVotes,
			"ForeignSplit":  fx.foreignSplit,
			"ForeignManual": fx.foreignManual,
			"ManualSrc":     fx.manualSrc,
			"Ambiguous":     fx.ambiguous,
			"Unclassified":  fx.unclassified,
		} {
			assert.Nil(t, findConsensus(result, f.ID), name)
		}
	})
	t.Run("SubjectChecked", func(t *testing.T) {
		counts, err := AnonymousFaceConsensus()
		require.NoError(t, err)

		for name, f := range map[string]*entity.Face{"Deleted": fx.deleted, "NonPerson": fx.nonPerson} {
			c := findConsensus(counts, f.ID)
			require.NotNil(t, c, name)
			assert.True(t, c.Agrees(consensusTestCore), name)
		}
	})
	t.Run("NoneAgree", func(t *testing.T) {
		result, err := ConsensusFaces(100)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestExistingPeople(t *testing.T) {
	alice := conflictTestSubject(t, "Existing Alice")
	bob := conflictTestSubject(t, "Existing Bob")
	gone := conflictTestSubject(t, "Existing Gone")
	require.NoError(t, UnscopedDb().Model(gone).UpdateColumn("deleted_at", entity.Now()).Error)
	pet := conflictTestSubject(t, "Existing Pet")
	require.NoError(t, UnscopedDb().Model(pet).UpdateColumn("subj_type", "pet").Error)
	uids := []string{alice.SubjUID, gone.SubjUID, pet.SubjUID, bob.SubjUID, rnd.GenerateUID('j')}

	t.Run("Success", func(t *testing.T) {
		result, err := existingPeople(uids, 1)
		require.NoError(t, err)
		assert.Equal(t, map[string]bool{alice.SubjUID: true, bob.SubjUID: true}, result)
	})
	t.Run("OneBatch", func(t *testing.T) {
		result, err := existingPeople(uids, 100)
		require.NoError(t, err)
		assert.Equal(t, map[string]bool{alice.SubjUID: true, bob.SubjUID: true}, result)
	})
	t.Run("InvalidSize", func(t *testing.T) {
		result, err := existingPeople(uids, 0)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})
	t.Run("Empty", func(t *testing.T) {
		result, err := existingPeople(nil, 1)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}
