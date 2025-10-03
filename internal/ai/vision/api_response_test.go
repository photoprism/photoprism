package vision

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestLabelResultToClassifyTopicality(t *testing.T) {
	r := LabelResult{Name: "tree", Confidence: 0.75, Topicality: 0.62}
	label := r.ToClassify(entity.SrcAuto)

	if label.Topicality != 62 {
		t.Fatalf("expected topicality 62, got %d", label.Topicality)
	}

	if label.Uncertainty >= 30 {
		t.Fatalf("expected uncertainty less than 30, got %d", label.Uncertainty)
	}
}
