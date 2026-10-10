package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestZeroKeyUpdates checks that update helpers refuse a model without a primary key instead of
// writing to every row, and that the same helpers still update a model with one.
func TestZeroKeyUpdates(t *testing.T) {
	t.Run("NoKey", func(t *testing.T) {
		for name, update := range map[string]func() error{
			"FileUpdate":       func() error { return (&File{}).Update("file_missing", true) },
			"FileUpdates":      func() error { return (&File{}).Updates(Values{"file_missing": true}) },
			"MarkerUpdate":     func() error { return (&Marker{}).Update("marker_review", true) },
			"MarkerUpdates":    func() error { return (&Marker{}).Updates(Values{"marker_review": true}) },
			"MarkerMatched":    func() error { return (&Marker{}).Matched() },
			"MarkerUnmatched":  func() error { return (&Marker{}).Unmatched() },
			"FaceMatched":      func() error { return (&Face{}).Matched() },
			"SubjectUpdate":    func() error { return (&Subject{}).Update("subj_notes", "zero key") },
			"SubjectUpdates":   func() error { return (&Subject{}).Updates(Values{"subj_notes": "zero key"}) },
			"SubjectRestore":   func() error { return (&Subject{DeletedAt: DeletedAtTime(Now())}).Restore() },
			"LabelRestore":     func() error { return (&Label{DeletedAt: DeletedAtTime(Now())}).Restore() },
			"FolderUpdates":    func() error { return (&Folder{}).Updates(Values{"folder_title": "zero key"}) },
			"ServiceUpdate":    func() error { return (&Service{}).Update("acc_error", "zero key") },
			"ServiceUpdates":   func() error { return (&Service{}).Updates(Values{"acc_error": "zero key"}) },
			"FileSyncUpdate":   func() error { return (&FileSync{}).Update("error", "zero key") },
			"FileSyncUpdates":  func() error { return (&FileSync{}).Updates(Values{"error": "zero key"}) },
			"FileShareUpdate":  func() error { return (&FileShare{}).Update("error", "zero key") },
			"FileShareUpdates": func() error { return (&FileShare{}).Updates(Values{"error": "zero key"}) },
			"SessionUpdates":   func() error { return (&Session{}).Updates(Values{"user_agent": "zero key"}) },
			"ClientUpdates":    func() error { return (&Client{}).Updates(Values{"client_name": "zero key"}) },
			"UserUpdates":      func() error { return (&User{}).Updates(Values{"display_name": "zero key"}) },
			"UserShareUpdates": func() error { return (&UserShare{}).Updates(Values{"comment": "zero key"}) },
			"UserSettingsUpdates": func() error {
				return (&UserSettings{}).Updates(Values{"ui_theme": "zero key"})
			},
			"UserDetailsUpdates": func() error {
				return (&UserDetails{}).Updates(Values{"given_name": "zero key"})
			},
			"PasscodeUpdates": func() error { return (&Passcode{}).Updates(Values{"key_url": "zero key"}) },
		} {
			t.Run(name, func(t *testing.T) {
				assert.ErrorIs(t, update(), gorm.ErrMissingWhereClause)
			})
		}
	})
	t.Run("LinkRedeem", func(t *testing.T) {
		var before, after int64

		require.NoError(t, UnscopedDb().Model(&Link{}).Select("COALESCE(SUM(link_views), 0)").Row().Scan(&before))
		(&Link{}).Redeem()
		require.NoError(t, UnscopedDb().Model(&Link{}).Select("COALESCE(SUM(link_views), 0)").Row().Scan(&after))

		assert.Equal(t, before, after)
	})
	t.Run("Keyed", func(t *testing.T) {
		m := MarkerFixtures.Pointer("1000003-4")
		review := m.MarkerReview
		t.Cleanup(func() { _ = m.Update("marker_review", review) })

		require.NoError(t, m.Update("marker_review", !review))

		found := Marker{}
		require.NoError(t, UnscopedDb().Where("marker_uid = ?", m.MarkerUID).First(&found).Error)
		assert.Equal(t, !review, found.MarkerReview)
	})
}
