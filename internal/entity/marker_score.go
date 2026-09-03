package entity

import (
	"strings"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/pkg/clean"
)

// ClusterScoreCond returns the detection score restriction automatic clustering applies, as an SQL
// fragment and its arguments. Columns are qualified with the given table alias, so the fragment can
// be nested where another markers row is already in scope; pass an empty alias to leave them bare.
// The alias is interpolated rather than bound, so it is filtered to a bare identifier first.
func ClusterScoreCond(alias string, floor int) (string, []any) {
	score, detector := "score", "detect_model"

	if alias = clean.SqlAlias(alias); alias != "" {
		score, detector = alias+".score", alias+".detect_model"
	}

	// FACE_CLUSTER_SCORE outranks the per-detector bars when an operator set one, and removes it
	// when negative. Applying one value to every marker is safe here in a way that taking the
	// active detector's bar is not: it is a choice rather than a calibration a marker was never
	// scored against. Without this the option configured nothing at all.
	if floor < 0 && face.ClusterScoreThreshold != 0 {
		floor = max(face.ClusterScoreThreshold, 0)
	}

	switch {
	case floor > 0:
		return score + " >= ?", []any{floor}
	case floor == 0:
		// No score filter at all, which is what a caller counting every marker asks for.
		return "1 = 1", nil
	}

	// One bar per detector, as a CASE over the column rather than a disjunction that would have to
	// name every detector twice - once to claim its rows and once to exclude them from the default.
	// A detector the registry does not name falls to ELSE, and so does a row with no recorded one,
	// which is every row written before the provenance column existed.
	var bars strings.Builder

	args := make([]any, 0, 2*len(face.Detectors)+1)

	for _, d := range face.Detectors {
		if d.ClusterScore <= 0 {
			continue
		}

		// Postgres treats the type of the result as text without an explicit cast.
		// Which then fails the subsequent >= test with a type conversion issue.
		bars.WriteString(" WHEN ? THEN CAST(? AS int)")
		args = append(args, d.Name, d.ClusterScore)
	}

	// CASE needs at least one WHEN, and without a registered bar there is nothing to distinguish.
	if bars.Len() == 0 {
		return score + " >= ?", []any{face.ClusterScoreThresholdDefault}
	}

	args = append(args, face.ClusterScoreThresholdDefault)

	return score + " >= CASE " + detector + bars.String() + " ELSE ? END", args
}

// FaceMemberCond returns the restriction identifying the markers a cluster is measured over: valid
// face markers that hold a cluster id and carry a vector.
//
// The vector test asserts rather than repairs: every writer of a cluster id goes through a distance
// comparison, so a marker without one cannot hold it today. It is stated here because the columns
// derived from this set would otherwise disagree if that ever stopped being true.
func FaceMemberCond() (string, []any) {
	return "marker_type = ? AND marker_invalid = FALSE AND face_id <> '' AND LENGTH(embeddings_json) > 0",
		[]any{MarkerFace}
}

// EmbeddingModelCond returns the restriction matching vectors that can be compared with the given
// model, as an SQL fragment and its arguments, or an empty fragment when no model is configured.
// A row with no recorded model is FaceNet's, which is why the blank case is not simply excluded.
func EmbeddingModelCond(model string) (string, []any) {
	if model == "" {
		return "", nil
	}

	return "embed_model = ? OR (embed_model = '' AND ? = ?)", []any{model, model, face.ModelFaceNet}
}
