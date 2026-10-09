package photoprism

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity"
)

// TestInsta360PhotoRemoved verifies that only deleted rows at quality -1 count as removed.
func TestInsta360PhotoRemoved(t *testing.T) {
	deletedAt := gorm.DeletedAt{Time: entity.Now(), Valid: true}
	assert.True(t, insta360PhotoRemoved(&entity.Photo{DeletedAt: deletedAt, PhotoQuality: -1}))
	assert.False(t, insta360PhotoRemoved(&entity.Photo{DeletedAt: deletedAt, PhotoQuality: 0}))
	assert.False(t, insta360PhotoRemoved(&entity.Photo{PhotoQuality: -1}))
	assert.False(t, insta360PhotoRemoved(&entity.Photo{}))
	assert.False(t, insta360PhotoRemoved(nil))
}

// TestInsta360CaptureState verifies which photo decides the archive state of a merged capture.
func TestInsta360CaptureState(t *testing.T) {
	removedAt := entity.Now()
	before, after := removedAt.Add(-time.Minute), removedAt.Add(time.Minute)
	photo := func(id uint, deletedAt *time.Time, quality int) *entity.Photo {
		realDeletedAt := gorm.DeletedAt{}
		if deletedAt != nil {
			realDeletedAt.Time = *deletedAt
			realDeletedAt.Valid = true
		}
		return &entity.Photo{ID: id, CreatedAt: before, DeletedAt: realDeletedAt, PhotoQuality: quality}
	}
	created := func(p *entity.Photo, createdAt time.Time) *entity.Photo {
		p.CreatedAt = createdAt
		return p
	}
	later := after.Add(time.Minute)

	cases := []struct {
		name   string
		photos entity.Photos
		expect uint
	}{
		{"Active", entity.Photos{photo(1, nil, 3), photo(2, &after, 3)}, 1},
		{"Archived", entity.Photos{photo(1, &removedAt, 3), photo(2, nil, 3)}, 1},
		{"RemovedActiveMember", entity.Photos{photo(1, &removedAt, -1), photo(2, nil, 3)}, 2},
		{"ActiveWinsOverArchived", entity.Photos{photo(1, &removedAt, -1), photo(2, &after, 3), photo(3, nil, 3)}, 3},
		{"ArchivedAfterRemoval", entity.Photos{photo(1, &removedAt, -1), photo(2, &before, 3), photo(3, &after, 4)}, 3},
		{"SkipsRemovedMembers", entity.Photos{photo(1, &removedAt, -1), photo(2, &after, -1), photo(3, &after, 4)}, 3},
		{"SameSecond", entity.Photos{photo(1, &removedAt, -1), photo(2, &removedAt, 4)}, 2},
		{"ArchivedBeforeRemoval", entity.Photos{photo(1, &removedAt, -1), photo(2, &before, 3)}, 1},
		{"AllRemoved", entity.Photos{photo(1, &removedAt, -1), photo(2, &after, -1)}, 1},
		{"FirstArchivedAfterRemoval", entity.Photos{photo(1, &removedAt, -1), photo(2, &after, 4), photo(3, &later, 5)}, 2},
		{"HiddenIsNotVisible", entity.Photos{photo(1, &removedAt, -1), photo(2, nil, -1), photo(3, &after, 4)}, 3},
		{"HiddenOnly", entity.Photos{photo(1, &removedAt, -1), photo(2, nil, -1)}, 1},
		{"VisibleCreatedLater", entity.Photos{photo(1, &removedAt, -1), photo(2, &after, 4), created(photo(3, nil, 3), later)}, 2},
		{"VisibleCreatedSameSecond", entity.Photos{photo(1, &removedAt, -1), photo(2, &after, 4), created(photo(3, nil, 3), after)}, 2},
		{"VisibleCreatedLaterOnly", entity.Photos{photo(1, &removedAt, -1), created(photo(2, nil, 3), later)}, 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := insta360CaptureState(tc.photos)
			require.NotNil(t, result)
			assert.Equal(t, tc.expect, result.ID)
		})
	}
	t.Run("Empty", func(t *testing.T) {
		assert.Nil(t, insta360CaptureState(nil))
		assert.Nil(t, insta360CaptureState(entity.Photos{nil}))
		assert.Equal(t, uint(1), insta360CaptureState(entity.Photos{photo(1, &removedAt, -1), nil}).ID)
	})
}
