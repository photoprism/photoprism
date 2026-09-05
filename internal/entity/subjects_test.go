package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteOrphanPeople(t *testing.T) {
	ValidateFixtures(t)
	t.Run("Ok", func(t *testing.T) {
		if count, err := DeleteOrphanPeople(); err != nil {
			t.Fatal(err)
		} else {
			t.Cleanup(func() {
				for _, fs := range SubjectFixtures {
					assert.NoError(t, UnscopedDb().Model(&Subject{}).Where("subj_uid = ?", fs.SubjUID).Updates(Values{"photo_count": fs.PhotoCount, "file_count": fs.FileCount, "deleted_at": fs.DeletedAt}).Error)
				}
				for _, ff := range FaceFixtures {
					assert.NoError(t, UnscopedDb().Model(&Face{}).Where("id = ?", ff.ID).Updates(Values{"subj_uid": ff.SubjUID}).Error)
				}
			})
			t.Logf("deleted %d faces", count)
		}
	})
}
