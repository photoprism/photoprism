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

func TestLabelResultToClassifyNSFW(t *testing.T) {
	r := LabelResult{Name: "lingerie", Confidence: 0.9, Topicality: 0.8, NSFW: true, NSFWConfidence: 0.65}
	label := r.ToClassify(entity.SrcAuto)

	if !label.NSFW {
		t.Fatalf("expected NSFW true")
	}

	if label.NSFWConfidence != 65 {
		t.Fatalf("expected NSFW confidence 65, got %d", label.NSFWConfidence)
	}
}
