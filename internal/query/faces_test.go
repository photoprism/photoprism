package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/stretchr/testify/assert"
)

func TestFaces(t *testing.T) {
	results, err := Faces(true, "")

	if err != nil {
		t.Fatal(err)
	}

	assert.GreaterOrEqual(t, len(results), 1)

	for _, val := range results {
		assert.IsType(t, entity.Face{}, val)
	}
}

func TestMatchFaceMarkers(t *testing.T) {
	const faceFixtureId = uint(6)

	if m, err := MarkerByID(faceFixtureId); err != nil {
		t.Fatal(err)
	} else {
		assert.Empty(t, m.SubjectUID)
	}

	// Reset subject_uid.
	if err := Db().Model(&entity.Marker{}).
		Where("subject_src = ?", entity.SrcAuto).
		Where("subject_uid = ?", "jqu0xs11qekk9jx8").
		Updates(entity.Values{"SubjectUID": ""}).Error; err != nil {
		t.Fatal(err)
	}

	affected, err := MatchFaceMarkers()

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(1), affected)

	if m, err := MarkerByID(faceFixtureId); err != nil {
		t.Fatal(err)
	} else {
		assert.Equal(t, "jqu0xs11qekk9jx8", m.SubjectUID)
	}
}

func TestRemoveAnonymousFaceClusters(t *testing.T) {
	removed, err := RemoveAnonymousFaceClusters()

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(1), removed)
}

func TestCountNewFaceMarkers(t *testing.T) {
	n := CountNewFaceMarkers()
	assert.GreaterOrEqual(t, n, 1)
}

func TestCountUnmatchedFaceMarkers(t *testing.T) {
	n := CountUnmatchedFaceMarkers()
	assert.GreaterOrEqual(t, n, 0)
}
