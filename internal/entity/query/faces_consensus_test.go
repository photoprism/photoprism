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

// consensusTestCore is the minimum number of automatic names the tests below pass.
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
	agrees := FaceConsensus{FaceID: "f", SubjUID: "s", Auto: 5, Unnamed: 5, Valid: 10}

	t.Run("Success", func(t *testing.T) {
		assert.True(t, agrees.Agrees(5))
	})
	t.Run("BelowCore", func(t *testing.T) {
		assert.False(t, agrees.Agrees(6))
	})
	t.Run("CoreBelowOne", func(t *testing.T) {
		assert.False(t, FaceConsensus{SubjUID: "s"}.Agrees(0))
		assert.True(t, FaceConsensus{SubjUID: "s", Auto: 1, Valid: 1}.Agrees(0))
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
	alice, bob                                                 *entity.Subject
	qualify, multiRow, belowCore, twoSubjects, manual, xmp     *entity.Face
	invalidManual, minority, hidden, deleted, nonPerson, named *entity.Face
	foreignFace, foreignVotes, foreignSplit, manualSrc         *entity.Face
	ambiguous, unclassified, foreignManual                     *entity.Face
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
		assert.Equal(t, FaceConsensus{FaceID: fx.qualify.ID, SubjUID: alice, Auto: core, Unnamed: 2, Valid: core + 2}, *c)
	})
	t.Run("MultiRow", func(t *testing.T) {
		c := findConsensus(counts, fx.multiRow.ID)
		require.NotNil(t, c)
		assert.Equal(t, FaceConsensus{FaceID: fx.multiRow.ID, SubjUID: alice, Auto: core, Unnamed: 1, Valid: core + 1}, *c)
	})
	t.Run("BelowCore", func(t *testing.T) {
		c := findConsensus(counts, fx.belowCore.ID)
		require.NotNil(t, c)
		assert.Equal(t, core-1, c.Auto)
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
	t.Run("Xmp", func(t *testing.T) {
		c := findConsensus(counts, fx.xmp.ID)
		require.NotNil(t, c)
		assert.Equal(t, core, c.Auto, "an XMP name is not a vote")
		assert.Equal(t, 1, c.Asserted)
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
		assert.Equal(t, core-1, c.Auto)
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

		assert.NotNil(t, findConsensus(result, fx.qualify.ID))
		assert.NotNil(t, findConsensus(result, fx.multiRow.ID))

		for name, f := range map[string]*entity.Face{
			"BelowCore":     fx.belowCore,
			"TwoSubjects":   fx.twoSubjects,
			"Manual":        fx.manual,
			"Xmp":           fx.xmp,
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
