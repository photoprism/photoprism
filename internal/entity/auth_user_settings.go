package entity

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// UserSettings represents user preferences.
type UserSettings struct {
	UserUID              string    `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;" json:"-" yaml:"UserUID"`
	UITheme              string    `gorm:"type:VARBINARY(32);column:ui_theme;" json:"UITheme,omitempty" yaml:"UITheme,omitempty"`
	UILanguage           string    `gorm:"type:VARBINARY(32);column:ui_language;" json:"UILanguage,omitempty" yaml:"UILanguage,omitempty"`
	UITimeZone           string    `gorm:"type:VARBINARY(64);column:ui_time_zone;" json:"UITimeZone,omitempty" yaml:"UITimeZone,omitempty"`
	MapsStyle            string    `gorm:"type:VARBINARY(32);" json:"MapsStyle,omitempty" yaml:"MapsStyle,omitempty"`
	MapsAnimate          int       `gorm:"default:0;" json:"MapsAnimate,omitempty" yaml:"MapsAnimate,omitempty"`
	IndexPath            string    `gorm:"type:VARBINARY(1024);" json:"IndexPath,omitempty" yaml:"IndexPath,omitempty"`
	IndexRescan          int       `gorm:"default:0;" json:"IndexRescan,omitempty" yaml:"IndexRescan,omitempty"`
	ImportPath           string    `gorm:"type:VARBINARY(1024);" json:"ImportPath,omitempty" yaml:"ImportPath,omitempty"`
	ImportMove           int       `gorm:"default:0;" json:"ImportMove,omitempty" yaml:"ImportMove,omitempty"`
	DownloadOriginals    int       `gorm:"default:0;" json:"DownloadOriginals,omitempty" yaml:"DownloadOriginals,omitempty"`
	DownloadMediaRaw     int       `gorm:"default:0;" json:"DownloadMediaRaw,omitempty" yaml:"DownloadMediaRaw,omitempty"`
	DownloadMediaSidecar int       `gorm:"default:0;" json:"DownloadMediaSidecar,omitempty" yaml:"DownloadMediaSidecar,omitempty"`
	UploadPath           string    `gorm:"type:VARBINARY(1024);" json:"UploadPath,omitempty" yaml:"UploadPath,omitempty"`
	DefaultPage          string    `gorm:"type:VARBINARY(128);" json:"DefaultPage,omitempty" yaml:"DefaultPage,omitempty"`
	CreatedAt            time.Time `json:"CreatedAt" yaml:"-"`
	UpdatedAt            time.Time `json:"UpdatedAt" yaml:"-"`
}

// TableName returns the entity table name.
func (UserSettings) TableName() string {
	return "auth_users_settings"
}

// NewUserSettings creates new user preferences.
func NewUserSettings(uid string) *UserSettings {
	return &UserSettings{UserUID: uid}
}

// CreateUserSettings creates new user settings or returns nil on error.
func CreateUserSettings(user *User) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	if user.GetUID() == "" {
		return fmt.Errorf("empty user uid")
	}

	user.UserSettings = &UserSettings{}

	if err := Db().Where("user_uid = ?", user.GetUID()).First(user.UserSettings).Error; err == nil {
		return nil
	}

	return user.UserSettings.Create()
}

// HasID tests if the entity has a valid uid.
func (m *UserSettings) HasID() bool {
	return rnd.IsUID(m.UserUID, UserUID)
}

// Create new entity in the database.
func (m *UserSettings) Create() error {
	return Db().Create(m).Error
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *UserSettings) Save() error {
	return Db().Save(m).Error
}

// Updates multiple properties in the database.
func (m *UserSettings) Updates(values interface{}) error {
	return UnscopedDb().Model(m).Updates(values).Error
}

// Apply applies the settings provided to the user preferences and keeps current values if they are not specified.
func (m *UserSettings) Apply(s *customize.Settings) *UserSettings {
	// UI preferences.
	if s.UI.Theme != "" {
		m.UITheme = s.UI.Theme
	}

	if s.UI.Language != "" {
		m.UILanguage = s.UI.Language
	}

	if s.UI.TimeZone != "" {
		m.UITimeZone = s.UI.TimeZone
	}

	// Maps preferences.
	if s.Maps.Style != "" {
		m.MapsStyle = s.Maps.Style

		if s.Maps.Animate > 0 {
			m.MapsAnimate = s.Maps.Animate
		} else {
			m.MapsAnimate = -1
		}
	}

	// Index preferences.
	if s.Index.Path != "" {
		m.IndexPath = s.Index.Path

		if s.Index.Rescan {
			m.IndexRescan = 1
		} else {
			m.IndexRescan = -1
		}
	}

	// Import preferences.
	if s.Import.Path != "" {
		m.ImportPath = s.Import.Path

		if s.Import.Move {
			m.ImportMove = 1
		} else {
			m.ImportMove = -1
		}
	}

	// Download preferences.
	if s.Download.Name != "" {
		if s.Download.Originals {
			m.DownloadOriginals = 1
		} else {
			m.DownloadOriginals = -1
		}
		if s.Download.MediaRaw {
			m.DownloadMediaRaw = 1
		} else {
			m.DownloadMediaRaw = -1
		}
		if s.Download.MediaSidecar {
			m.DownloadMediaSidecar = 1
		} else {
			m.DownloadMediaSidecar = -1
		}
	}

	return m
}

// ApplyTo applies the user preferences to the client settings and keeps the default settings if they are not specified.
func (m *UserSettings) ApplyTo(s *customize.Settings) *customize.Settings {
	if m.UITheme != "" {
		s.UI.Theme = m.UITheme
	}

	if m.UILanguage != "" {
		s.UI.Language = m.UILanguage
	}

	if m.UITimeZone != "" {
		s.UI.TimeZone = m.UITimeZone
	}

	if m.MapsStyle != "" {
		s.Maps.Style = m.MapsStyle
	}

	if m.MapsAnimate > 0 {
		s.Maps.Animate = m.MapsAnimate
	} else if m.MapsAnimate < 0 {
		s.Maps.Animate = 0
	}

	if m.IndexPath != "" {
		s.Index.Path = m.IndexPath
	}

	if m.IndexRescan > 0 {
		s.Index.Rescan = true
	} else if m.IndexRescan < 0 {
		s.Index.Rescan = false
	}

	if m.ImportPath != "" {
		s.Import.Path = m.ImportPath
	}

	if m.ImportMove > 0 {
		s.Import.Move = true
	} else if m.ImportMove < 0 {
		s.Import.Move = false
	}

	if m.DownloadOriginals > 0 {
		s.Download.Originals = true
	} else if m.DownloadOriginals < 0 {
		s.Download.Originals = false
	}

	if m.DownloadMediaRaw > 0 {
		s.Download.MediaRaw = true
	} else if m.DownloadMediaRaw < 0 {
		s.Download.MediaRaw = false
	}

	if m.DownloadMediaSidecar > 0 {
		s.Download.MediaSidecar = true
	} else if m.DownloadMediaSidecar < 0 {
		s.Download.MediaSidecar = false
	}

	return s
}
