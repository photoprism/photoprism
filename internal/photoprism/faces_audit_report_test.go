package photoprism

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
)

// auditConflictMarker persists an anonymous cluster and a marker that names a subject the cluster
// does not carry, which is the row MarkersWithSubjectConflict returns.
func auditConflictMarker(t *testing.T, seed uint64, subjSrc string) (*entity.Face, *entity.Subject, *entity.Marker) {
	t.Helper()

	subj := entity.NewSubject("Audit Conflict Person", entity.SubjPerson, entity.SrcManual)
	require.NotNil(t, subj)
	require.NoError(t, subj.Create())

	embeddings := make(face.Embeddings, face.ClusterCore)
	base := face.FixtureEmbedding(seed)

	for i := range embeddings {
		embeddings[i] = base
	}

	f := entity.NewFace("", entity.SrcAuto, embeddings, face.EmbeddingModelName())
	require.NotNil(t, f)
	require.NoError(t, f.Create())

	m := &entity.Marker{
		FileUID:    "fs6sg6bw45bnlqdw",
		MarkerType: entity.MarkerFace,
		MarkerSrc:  entity.SrcImage,
		FaceID:     f.ID,
		SubjUID:    subj.SubjUID,
		SubjSrc:    subjSrc,
		Size:       face.SizeThreshold,
		Score:      50,
		X:          0.2,
		Y:          0.2,
		W:          0.1,
		H:          0.1,
	}

	m.SetEmbeddings(face.Embeddings{face.FixtureEmbeddingAt(base, 0.05, seed+1)}, f.EmbedModel, face.DetectorYuNet)
	require.NoError(t, entity.Db().Create(m).Error)

	return f, subj, m
}

// auditSubject persists a person the audit can report on.
func auditSubject(t *testing.T, name string) *entity.Subject {
	t.Helper()

	subj := entity.NewSubject(name, entity.SubjPerson, entity.SrcManual)
	require.NotNil(t, subj)
	require.NoError(t, subj.Create())

	return subj
}

// auditNamedCluster persists a cluster that already carries a subject.
func auditNamedCluster(t *testing.T, subjUID string, seed uint64) *entity.Face {
	t.Helper()

	embeddings := make(face.Embeddings, face.ClusterCore)
	for i := range embeddings {
		embeddings[i] = face.FixtureEmbedding(seed)
	}

	f := entity.NewFace(subjUID, entity.SrcManual, embeddings, face.EmbeddingModelName())
	require.NotNil(t, f)
	require.NoError(t, f.Create())

	return f
}

// auditMarker persists a face marker pointing at a cluster, with the given subject and source.
func auditMarker(t *testing.T, faceID, subjUID, subjSrc string, seed uint64) *entity.Marker {
	t.Helper()

	m := &entity.Marker{
		FileUID:    "fs6sg6bw45bnlqdw",
		MarkerType: entity.MarkerFace,
		MarkerSrc:  entity.SrcImage,
		FaceID:     faceID,
		SubjUID:    subjUID,
		SubjSrc:    subjSrc,
		Size:       face.SizeThreshold,
		Score:      50,
		X:          0.2,
		Y:          0.2,
		W:          0.1,
		H:          0.1,
	}

	m.SetEmbeddings(face.Embeddings{face.FixtureEmbedding(seed)}, face.EmbeddingModelName(), face.DetectorYuNet)
	require.NoError(t, entity.Db().Create(m).Error)

	return m
}

// TestFaces_AuditReportsAnonymousClusters covers the split between the state matching produces and
// the one that is genuinely wrong. An automatic subject must not name its cluster, so the audit
// reporting every such marker as a conflict described a library working as designed as broken - and
// with --fix it unmatched them.
func TestFaces_AuditReportsAnonymousClusters(t *testing.T) {
	t.Run("AutomaticSubjectIsExpected", func(t *testing.T) {
		w := isolatedTestFaces(t, "faces-audit-anonymous-auto")
		_, _, m := auditConflictMarker(t, 9101, entity.SrcAuto)

		hook := captureLog(t)
		require.NoError(t, w.Audit(true, ""))

		warnings := strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n")
		assert.NotContains(t, warnings, m.MarkerUID, "a marker in the state SetFace produces is not a fault")

		reported := strings.Join(loggedMessages(hook, logrus.InfoLevel), "\n")
		assert.Contains(t, reported, "carry a subject the matcher assigned", "it is counted once instead")

		after := entity.FindMarker(m.MarkerUID)
		require.NotNil(t, after)
		assert.Equal(t, m.FaceID, after.FaceID, "and --fix must not unmatch it")
	})
	t.Run("ManualSubjectIsAFault", func(t *testing.T) {
		// SetFace names an anonymous cluster after a manual subject, so one that is still unnamed
		// here is the case the warning exists for - and the one the noise was burying.
		w := isolatedTestFaces(t, "faces-audit-anonymous-manual")
		_, _, m := auditConflictMarker(t, 9111, entity.SrcManual)

		hook := captureLog(t)
		require.NoError(t, w.Audit(false, ""))

		warnings := strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n")
		assert.Contains(t, warnings, m.MarkerUID)
		assert.Contains(t, warnings, "points to face")
	})
	t.Run("HiddenClusterIsNotMissing", func(t *testing.T) {
		// Isolates the lookup: the cluster exists but the matcher's own view excludes it, which
		// is not the same as a row that is gone. A manual subject, so the anonymous skip above
		// cannot decide this case instead.
		w := isolatedTestFaces(t, "faces-audit-hidden-cluster")
		f, _, m := auditConflictMarker(t, 9131, entity.SrcManual)
		require.NoError(t, entity.Db().Model(&entity.Face{}).Where("id = ?", f.ID).
			UpdateColumn("face_hidden", true).Error)

		hook := captureLog(t)
		require.NoError(t, w.Audit(false, ""))

		warnings := strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n")
		assert.NotContains(t, warnings, "references missing face")
		assert.Contains(t, warnings, "points to face", "it is reported as what it is")
		assert.Contains(t, warnings, m.MarkerUID)
	})
	t.Run("ScopedRunReportsBothSides", func(t *testing.T) {
		// Isolates the scope filter, in both directions: a conflict the scoped person is on either
		// side of is reported, and one they are on neither side of is left alone.
		w := isolatedTestFaces(t, "faces-audit-scoped-sides")

		scoped := auditSubject(t, "Audit Scope Scoped")
		named := auditSubject(t, "Audit Scope Named")
		other := auditSubject(t, "Audit Scope Other")

		onScoped := auditNamedCluster(t, scoped.SubjUID, 9141)
		mine := auditMarker(t, onScoped.ID, named.SubjUID, entity.SrcManual, 9142)

		onOther := auditNamedCluster(t, other.SubjUID, 9151)
		theirs := auditMarker(t, onOther.ID, named.SubjUID, entity.SrcManual, 9152)

		hook := captureLog(t)
		require.NoError(t, w.Audit(false, scoped.SubjUID))

		warnings := strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n")
		assert.Contains(t, warnings, mine.MarkerUID, "the scoped person is the cluster's side of it")
		assert.NotContains(t, warnings, theirs.MarkerUID, "and this conflict is not theirs")
	})
}

// TestAmbiguousPairMessage covers the single line that replaced the three the audit printed per
// conflicting pair, which was 51% of a migration's output.
func TestAmbiguousPairMessage(t *testing.T) {
	t.Run("BothNamed", func(t *testing.T) {
		// A measured extent on the first cluster only, so the two accept distances differ and the
		// line can be shown to report the side the pair was accepted from.
		f1 := entity.Face{ID: "f1", SubjUID: "js6sg6b1qekk9jx8", FaceSrc: entity.SrcAuto, Samples: 5, SampleRadius: 0.3, CollisionRadius: 0.4}
		f2 := entity.Face{ID: "f2", SubjUID: "js6sg6b1qekk9jx9", FaceSrc: entity.SrcManual}

		msg := ambiguousPairMessage(f1, f2, 0.82)

		assert.Contains(t, msg, "face f1")
		assert.Contains(t, msg, "face f2")
		assert.Contains(t, msg, "js6sg6b1qekk9jx8")
		assert.Contains(t, msg, "js6sg6b1qekk9jx9")
		assert.Contains(t, msg, "0.820000", "the distance that was measured")
		assert.Contains(t, msg, "5 samples", "and the evidence the bar rests on")
		// The bar is the first cluster's, since that is the side the pair was accepted from.
		assert.Contains(t, msg, fmt.Sprintf("%f", f1.AcceptDist()))
		assert.NotContains(t, msg, fmt.Sprintf("%f", f2.AcceptDist()))
	})
	t.Run("Anonymous", func(t *testing.T) {
		msg := ambiguousPairMessage(entity.Face{ID: "f1"}, entity.Face{ID: "f2", SubjUID: "js6sg6b1qekk9jx8"}, 0.5)

		assert.Contains(t, msg, "no subject")
		assert.Contains(t, msg, "js6sg6b1qekk9jx8")
	})
}

// TestFaceSubjectLabel covers both branches of the label the pair line is built from.
func TestFaceSubjectLabel(t *testing.T) {
	t.Run("Named", func(t *testing.T) {
		label := faceSubjectLabel(entity.Face{SubjUID: "js6sg6b1qekk9jx8", FaceSrc: entity.SrcManual})
		assert.Contains(t, label, "subject")
		assert.Contains(t, label, "js6sg6b1qekk9jx8")
	})
	t.Run("Unnamed", func(t *testing.T) {
		assert.Contains(t, faceSubjectLabel(entity.Face{FaceSrc: entity.SrcAuto}), "no subject")
	})
}

// TestAmbiguousPairSummary pins that the three counts are reported apart. They differ by an order
// of magnitude on a freshly migrated library, and only the people are actionable.
func TestAmbiguousPairSummary(t *testing.T) {
	t.Run("Many", func(t *testing.T) {
		summary := ambiguousPairSummary(553, 152, 8)

		assert.Contains(t, summary, "553 ambiguous pairs")
		assert.Contains(t, summary, "152 clusters")
		assert.Contains(t, summary, "8 subjects")
	})
	t.Run("One", func(t *testing.T) {
		assert.Equal(t, "1 ambiguous pair across 2 clusters and 2 subjects", ambiguousPairSummary(1, 2, 2))
	})
}
