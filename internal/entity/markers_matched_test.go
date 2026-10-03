package entity

import (
	"errors"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStampMarkerMatches checks chunk bounds, durable timestamps, and partial failure.
func TestStampMarkerMatches(t *testing.T) {
	f := raceTestFace(t, "", 8290)
	uids := syncTestMarkers(t, f, BatchSize()+2, "", raceTestFile)
	outside := syncTestMarkers(t, f, 1, "", raceTestFile)[0]
	oldStamp := Time("2000-01-01T00:00:00Z")
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", outside).UpdateColumn("matched_at", oldStamp).Error)
	for _, name := range []string{"Stored", "FirstChunkFails", "PriorStampFails"} {
		fail := name != "Stored"
		t.Run(name, func(t *testing.T) {
			var beforeStamp *time.Time
			if name == "PriorStampFails" {
				beforeStamp = Time("2001-01-01T00:00:00Z")
			}
			require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid IN (?)", uids).UpdateColumn("matched_at", beforeStamp).Error)
			var rows Markers
			require.NoError(t, UnscopedDb().Where("marker_uid IN (?)", uids).Order("marker_uid").Find(&rows).Error)
			markers := make([]*Marker, len(rows))
			for i := range rows {
				markers[i] = &rows[i]
			}
			updates := 0
			Db().Callback().Update().Before("gorm:begin_transaction").Register("test:stamp-matches", func(scope *gorm.Scope) {
				if scope.TableName() != (Marker{}).TableName() {
					return
				}
				updates++
				if fail && updates == 1 {
					_ = scope.Err(errors.New("test stamp failure"))
				}
			})
			t.Cleanup(func() { Db().Callback().Update().Remove("test:stamp-matches") })
			at := time.Date(2002, time.January, 2, 3, 4, 5, 123000000, time.FixedZone("UTC+2", 7200))
			statements := countStatements(t, func() {
				err := StampMarkerMatches(markers, &at)
				if fail {
					require.ErrorContains(t, err, "test stamp failure")
				} else {
					require.NoError(t, err)
				}
			})
			assert.Equal(t, oldStamp, FindMarker(outside).MatchedAt)
			assert.Equal(t, 2, updates)
			if !fail {
				assert.Len(t, statements, 2)
			}
			var stored Markers
			require.NoError(t, UnscopedDb().Where("marker_uid IN (?)", uids).Order("marker_uid").Find(&stored).Error)
			require.Len(t, stored, len(markers))
			for i, m := range markers {
				if fail && i < BatchSize() {
					assert.Equal(t, beforeStamp, m.MatchedAt)
					assert.Equal(t, beforeStamp, stored[i].MatchedAt)
				} else {
					require.NotNil(t, m.MatchedAt)
					assert.Equal(t, at.UTC().Truncate(time.Second), *m.MatchedAt)
					assert.Equal(t, m.MatchedAt, stored[i].MatchedAt)
				}
			}
		})
	}
	t.Run("Empty", func(t *testing.T) {
		at := TimeStamp()
		statements := countStatements(t, func() {
			require.NoError(t, StampMarkerMatches(nil, nil))
			require.NoError(t, StampMarkerMatches([]*Marker{nil, {}}, at))
			require.Error(t, StampMarkerMatches([]*Marker{{MarkerUID: uids[0]}}, nil))
		})
		assert.Empty(t, statements)
	})
}
