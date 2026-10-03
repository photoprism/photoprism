package photoprism

import (
	"errors"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestFaces_startKeepsPartialSubjects pins that the markers linked to people before an error still
// count as a change, so counts and covers are refreshed.
func TestFaces_startKeepsPartialSubjects(t *testing.T) {
	w := isolatedTestFaces(t, "facessubjectspartial")

	// A first run settles what the fixtures leave for it.
	_, err := w.start(FacesOptions{Threshold: 1000000})
	require.NoError(t, err)

	newMarker := func(name string) string {
		m := entity.Marker{
			MarkerUID:  rnd.GenerateUID('m'),
			FileUID:    consensusTestFileUID,
			MarkerType: entity.MarkerFace,
			MarkerName: name,
			SubjSrc:    entity.SrcManual,
			W:          0.1,
			H:          0.1,
		}
		require.NoError(t, entity.UnscopedDb().Create(&m).Error)
		t.Cleanup(func() { entity.UnscopedDb().Delete(entity.Marker{}, "marker_uid = ?", m.MarkerUID) })
		return m.MarkerUID
	}

	t.Cleanup(func() {
		entity.UnscopedDb().Delete(entity.Subject{}, "subj_name IN (?)", []string{"Partial Subjects Ada", "Partial Subjects Bea"})
	})

	// Both people exist and keep a linked marker, so the run removes no orphan person that would
	// refresh counts on its own.
	ada := consensusTestSubject(t, "Partial Subjects Ada")
	bea := consensusTestSubject(t, "Partial Subjects Bea")

	for _, s := range []*entity.Subject{ada, bea} {
		kept := newMarker(s.SubjName)
		require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).Where("marker_uid = ?", kept).UpdateColumn("subj_uid", s.SubjUID).Error)
	}

	// A settled run, so nothing else in the next one changes what counts are computed from.
	_, err = w.start(FacesOptions{Threshold: 1000000})
	require.NoError(t, err)
	require.NoError(t, entity.UnscopedDb().Model(ada).UpdateColumn("file_count", 0).Error)

	first := newMarker("Partial Subjects Ada")
	second := newMarker("Partial Subjects Bea")

	entity.Db().Callback().Update().After("gorm:update").Register("test:subjects-partial", func(scope *gorm.Scope) {
		if m, ok := scope.Value.(*entity.Marker); ok && m.MarkerUID == second {
			_ = scope.Err(errors.New("update refused"))
		}
	})
	t.Cleanup(func() { entity.Db().Callback().Update().Remove("test:subjects-partial") })

	result, err := w.start(FacesOptions{Threshold: 1000000})
	require.NoError(t, err)

	linked := entity.FindMarker(first)
	require.NotNil(t, linked)
	require.NotEmpty(t, linked.SubjUID, "the first marker is linked")
	assert.Equal(t, 1, result.Subjects, "only the name whose marker was linked counts")

	require.Equal(t, ada.SubjUID, linked.SubjUID)
	assert.Positive(t, entity.FindSubject(ada.SubjUID).FileCount, "the counts are refreshed")
}
