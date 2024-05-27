package entity

import (
	"fmt"
	"image"
	"math"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/internal/customize"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/colors"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/projection"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	FileUID = byte('f')
)

// Files represents a file result set.
type Files []File

// Index updates should not run simultaneously.
var fileIndexMutex = sync.Mutex{}
var filePrimaryMutex = sync.Mutex{}

// File represents an image or sidecar file that belongs to a photo.
type File struct {
	ID                 uint          `gorm:"primary_key" json:"-" yaml:"-"`
	Photo              *Photo        `json:"-" yaml:"-"`
	PhotoID            uint          `gorm:"index:idx_files_photo_id;" json:"-" yaml:"-"`
	PhotoUID           string        `gorm:"type:VARBINARY(42);index;" json:"PhotoUID" yaml:"PhotoUID"`
	PhotoTakenAt       time.Time     `gorm:"type:DATETIME;index;" json:"TakenAt" yaml:"TakenAt"`
	TimeIndex          *string       `gorm:"type:VARBINARY(64);" json:"TimeIndex" yaml:"TimeIndex"`
	MediaID            *string       `gorm:"type:VARBINARY(32);" json:"MediaID" yaml:"MediaID"`
	MediaUTC           int64         `gorm:"column:media_utc;index;"  json:"MediaUTC" yaml:"MediaUTC,omitempty"`
	InstanceID         string        `gorm:"type:VARBINARY(64);index;" json:"InstanceID,omitempty" yaml:"InstanceID,omitempty"`
	FileUID            string        `gorm:"type:VARBINARY(42);unique_index;" json:"UID" yaml:"UID"`
	FileName           string        `gorm:"type:VARBINARY(1024);unique_index:idx_files_name_root;" json:"Name" yaml:"Name"`
	FileRoot           string        `gorm:"type:VARBINARY(16);default:'/';unique_index:idx_files_name_root;" json:"Root" yaml:"Root,omitempty"`
	OriginalName       string        `gorm:"type:VARBINARY(755);" json:"OriginalName" yaml:"OriginalName,omitempty"`
	FileHash           string        `gorm:"type:VARBINARY(128);index" json:"Hash" yaml:"Hash,omitempty"`
	FileSize           int64         `json:"Size" yaml:"Size,omitempty"`
	FileCodec          string        `gorm:"type:VARBINARY(32)" json:"Codec" yaml:"Codec,omitempty"`
	FileType           string        `gorm:"type:VARBINARY(16)" json:"FileType" yaml:"FileType,omitempty"`
	MediaType          string        `gorm:"type:VARBINARY(16)" json:"MediaType" yaml:"MediaType,omitempty"`
	FileMime           string        `gorm:"type:VARBINARY(64)" json:"Mime" yaml:"Mime,omitempty"`
	FilePrimary        bool          `gorm:"index:idx_files_photo_id;" json:"Primary" yaml:"Primary,omitempty"`
	FileSidecar        bool          `json:"Sidecar" yaml:"Sidecar,omitempty"`
	FileMissing        bool          `json:"Missing" yaml:"Missing,omitempty"`
	FilePortrait       bool          `json:"Portrait" yaml:"Portrait,omitempty"`
	FileVideo          bool          `json:"Video" yaml:"Video,omitempty"`
	FileDuration       time.Duration `json:"Duration" yaml:"Duration,omitempty"`
	FileFPS            float64       `gorm:"column:file_fps;" json:"FPS" yaml:"FPS,omitempty"`
	FileFrames         int           `gorm:"column:file_frames;" json:"Frames" yaml:"Frames,omitempty"`
	FileWidth          int           `gorm:"column:file_width;" json:"Width" yaml:"Width,omitempty"`
	FileHeight         int           `gorm:"column:file_height;" json:"Height" yaml:"Height,omitempty"`
	FileOrientation    int           `gorm:"column:file_orientation;" json:"Orientation" yaml:"Orientation,omitempty"`
	FileOrientationSrc string        `gorm:"column:file_orientation_src;type:VARBINARY(8);default:'';" json:"OrientationSrc" yaml:"OrientationSrc,omitempty"`
	FileProjection     string        `gorm:"column:file_projection;type:VARBINARY(64);" json:"Projection,omitempty" yaml:"Projection,omitempty"`
	FileAspectRatio    float32       `gorm:"column:file_aspect_ratio;type:FLOAT;" json:"AspectRatio" yaml:"AspectRatio,omitempty"`
	FileHDR            bool          `gorm:"column:file_hdr;"  json:"HDR" yaml:"HDR,omitempty"`
	FileWatermark      bool          `gorm:"column:file_watermark;"  json:"Watermark" yaml:"Watermark,omitempty"`
	FileColorProfile   string        `gorm:"type:VARBINARY(64);" json:"ColorProfile,omitempty" yaml:"ColorProfile,omitempty"`
	FileMainColor      string        `gorm:"type:VARBINARY(16);" json:"MainColor" yaml:"MainColor,omitempty"`
	FileColors         string        `gorm:"type:VARBINARY(18);" json:"Colors" yaml:"Colors,omitempty"`
	FileLuminance      string        `gorm:"type:VARBINARY(18);" json:"Luminance" yaml:"Luminance,omitempty"`
	FileDiff           int           `json:"Diff" yaml:"Diff,omitempty"`
	FileChroma         int16         `json:"Chroma" yaml:"Chroma,omitempty"`
	FileSoftware       string        `gorm:"type:VARCHAR(64)" json:"Software" yaml:"Software,omitempty"`
	FileError          string        `gorm:"type:VARBINARY(512);index;" json:"Error" yaml:"Error,omitempty"`
	ModTime            int64         `json:"ModTime" yaml:"-"`
	CreatedAt          time.Time     `json:"CreatedAt" yaml:"-"`
	CreatedIn          int64         `json:"CreatedIn" yaml:"-"`
	UpdatedAt          time.Time     `json:"UpdatedAt" yaml:"-"`
	UpdatedIn          int64         `json:"UpdatedIn" yaml:"-"`
	PublishedAt        *time.Time    `sql:"index" json:"PublishedAt,omitempty" yaml:"PublishedAt,omitempty"`
	DeletedAt          *time.Time    `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
	Share              []FileShare   `json:"-" yaml:"-"`
	Sync               []FileSync    `json:"-" yaml:"-"`
	markers            *Markers
}

// TableName returns the entity table name.
func (File) TableName() string {
	return "files"
}

// RegenerateIndex updates the search index columns.
func (m File) RegenerateIndex() {
	fileIndexMutex.Lock()
	defer fileIndexMutex.Unlock()

	start := time.Now()

	photosTable := Photo{}.TableName()

	var updateWhere *gorm.SqlExpr
	var scope string

	if m.PhotoID > 0 {
		updateWhere = gorm.Expr("files.photo_id = ?", m.PhotoID)
		scope = "index by photo id"
	} else if m.PhotoUID != "" {
		updateWhere = gorm.Expr("files.photo_uid = ?", m.PhotoUID)
		scope = "index by photo uid"
	} else if m.ID > 0 {
		updateWhere = gorm.Expr("files.id = ?", m.ID)
		scope = "index by file id"
	} else {
		updateWhere = gorm.Expr("files.photo_id IS NOT NULL")
		scope = "index"
	}

	switch DbDialect() {
	case MySQL:
		Log("files", "regenerate photo_taken_at",
			Db().Exec("UPDATE files JOIN ? p ON p.id = files.photo_id SET files.photo_taken_at = p.taken_at_local WHERE ?",
				gorm.Expr(photosTable), updateWhere).Error)

		Log("files", "regenerate media_id",
			Db().Exec("UPDATE files SET media_id = CASE WHEN file_missing = 0 AND deleted_at IS NULL THEN CONCAT((10000000000 - photo_id), '-', 1 + file_sidecar - file_primary, '-', file_uid) ELSE NULL END WHERE ?",
				updateWhere).Error)

		Log("files", "regenerate time_index",
			Db().Exec("UPDATE files SET time_index = CASE WHEN media_id IS NOT NULL AND photo_taken_at IS NOT NULL THEN CONCAT(100000000000000 - CAST(photo_taken_at AS UNSIGNED), '-', media_id) ELSE NULL END WHERE ?",
				updateWhere).Error)
	case SQLite3:
		Log("files", "regenerate photo_taken_at",
			Db().Exec("UPDATE files SET photo_taken_at = (SELECT p.taken_at_local FROM ? p WHERE p.id = photo_id) WHERE ?",
				gorm.Expr(photosTable), updateWhere).Error)

		Log("files", "regenerate media_id",
			Db().Exec("UPDATE files SET media_id = CASE WHEN file_missing = 0 AND deleted_at IS NULL THEN ((10000000000 - photo_id) || '-' || (1 + file_sidecar - file_primary) || '-' || file_uid) ELSE NULL END WHERE ?",
				updateWhere).Error)

		Log("files", "regenerate time_index",
			Db().Exec("UPDATE files SET time_index = CASE WHEN media_id IS NOT NULL AND photo_taken_at IS NOT NULL THEN ((100000000000000 - strftime('%Y%m%d%H%M%S', photo_taken_at)) || '-' || media_id) ELSE NULL END WHERE ?",
				updateWhere).Error)
	default:
		log.Warnf("sql: unsupported dialect %s", DbDialect())
	}

	log.Debugf("search: updated %s [%s]", scope, time.Since(start))
}

// FirstFileByHash gets a file in db from its hash
func FirstFileByHash(fileHash string) (File, error) {
	var file File

	res := Db().Unscoped().First(&file, "file_hash = ?", fileHash)

	return file, res.Error
}

// PrimaryFile returns the primary file for a photo uid.
func PrimaryFile(photoUid string) (*File, error) {
	file := File{}

	res := Db().Unscoped().First(&file, "file_primary = 1 AND photo_uid = ?", photoUid)

	return &file, res.Error
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *File) BeforeCreate(scope *gorm.Scope) error {
	// Set MediaType based on FileName if empty.
	if m.MediaType == "" && m.FileName != "" {
		m.MediaType = media.FromName(m.FileName).String()
	}

	// Set MediaUTC based on PhotoTakenAt if empty.
	if m.MediaUTC == 0 && !m.PhotoTakenAt.IsZero() {
		m.MediaUTC = m.PhotoTakenAt.UnixMilli()
	}

	// Return if uid exists.
	if rnd.IsUnique(m.FileUID, FileUID) {
		return nil
	}

	return scope.SetColumn("FileUID", rnd.GenerateUID(FileUID))
}

// DownloadName returns the download file name.
func (m *File) DownloadName(n customize.DownloadName, seq int) string {
	switch n {
	case customize.DownloadNameFile:
		return m.Base(seq)
	case customize.DownloadNameOriginal:
		return m.OriginalBase(seq)
	default:
		return m.ShareBase(seq)
	}
}

// Base returns the file name without path.
func (m *File) Base(seq int) string {
	if m.FileName == "" {
		return m.ShareBase(seq)
	}

	base := filepath.Base(m.FileName)

	if seq > 0 {
		return fmt.Sprintf("%s (%d)%s", fs.StripExt(base), seq, filepath.Ext(base))
	}

	return base
}

// OriginalBase returns the original file name without path.
func (m *File) OriginalBase(seq int) string {
	if m.OriginalName == "" {
		return m.Base(seq)
	}

	base := filepath.Base(m.OriginalName)

	if seq > 0 {
		return fmt.Sprintf("%s (%d)%s", fs.StripExt(base), seq, filepath.Ext(base))
	}

	return base
}

// ShareBase returns a meaningful file name for sharing.
func (m *File) ShareBase(seq int) string {
	photo := m.RelatedPhoto()

	if photo == nil {
		return fmt.Sprintf("%s.%s", m.FileHash, m.FileType)
	} else if len(m.FileHash) < 8 {
		return fmt.Sprintf("%s.%s", rnd.UUID(), m.FileType)
	} else if photo.TakenAtLocal.IsZero() || photo.PhotoTitle == "" {
		return fmt.Sprintf("%s.%s", m.FileHash, m.FileType)
	}

	name := txt.Title(slug.MakeLang(photo.PhotoTitle, "en"))
	taken := photo.TakenAtLocal.Format("20060102-150405")

	if seq > 0 {
		return fmt.Sprintf("%s-%s (%d).%s", taken, name, seq, m.FileType)
	}

	return fmt.Sprintf("%s-%s.%s", taken, name, m.FileType)
}

// Changed returns true if new and old file size or modified time are different.
func (m File) Changed(fileSize int64, modTime time.Time) bool {
	// File size has changed.
	if m.FileSize != fileSize {
		return true
	}

	// Modification time has changed.
	if m.ModTime == modTime.UTC().Truncate(time.Second).Unix() {
		return false
	}

	return true
}

// Missing returns true if this file is current missing or marked as deleted.
func (m File) Missing() bool {
	return m.FileMissing || m.DeletedAt != nil
}

// DeletePermanently permanently removes a file from the index.
func (m *File) DeletePermanently() error {
	if m.ID < 1 || m.FileUID == "" {
		return fmt.Errorf("invalid file id %d / uid %s", m.ID, clean.Log(m.FileUID))
	}

	if err := UnscopedDb().Delete(Marker{}, "file_uid = ?", m.FileUID).Error; err != nil {
		log.Errorf("file %s: %s while removing markers", clean.Log(m.FileUID), err)
	}

	if err := UnscopedDb().Delete(FileShare{}, "file_id = ?", m.ID).Error; err != nil {
		log.Errorf("file %s: %s while removing share info", clean.Log(m.FileUID), err)
	}

	if err := UnscopedDb().Delete(FileSync{}, "file_id = ?", m.ID).Error; err != nil {
		log.Errorf("file %s: %s while removing remote sync info", clean.Log(m.FileUID), err)
	}

	if err := m.ReplaceHash(""); err != nil {
		log.Errorf("file %s: %s while removing covers", clean.Log(m.FileUID), err)
	}

	return UnscopedDb().Delete(m).Error
}

// ReplaceHash updates file hash references.
func (m *File) ReplaceHash(newHash string) error {
	if m.FileHash == newHash {
		// Nothing to do.
		return nil
	}

	// Log values.
	if m.FileHash != "" && newHash == "" {
		log.Tracef("file %s: removing hash %s", clean.Log(m.FileUID), clean.Log(m.FileHash))
	} else if m.FileHash != "" && newHash != "" {
		log.Tracef("file %s: hash %s changed to %s", clean.Log(m.FileUID), clean.Log(m.FileHash), clean.Log(newHash))
		// Reset error when hash changes.
		m.FileError = ""
	}

	// Set file hash to new value.
	oldHash := m.FileHash
	m.FileHash = newHash

	// Ok to skip updating related tables?
	if m.NoJPEG() || m.FileHash == "" {
		return nil
	}

	entities := Tables{
		"albums": Album{},
		"labels": Label{},
	}

	// Search related tables for references and update them.
	for name, entity := range entities {
		start := time.Now()

		if res := UnscopedDb().Model(entity).Where("thumb = ?", oldHash).UpdateColumn("thumb", newHash); res.Error != nil {
			return res.Error
		} else if res.RowsAffected > 0 {
			log.Infof("%s: updated %s [%s]", name, english.Plural(int(res.RowsAffected), "cover", "covers"), time.Since(start))
		}
	}

	return nil
}

// Delete deletes the entity from the database.
func (m *File) Delete(permanently bool) error {
	if m.ID < 1 || m.FileUID == "" {
		return fmt.Errorf("invalid file id %d / uid %s", m.ID, clean.Log(m.FileUID))
	}

	if permanently {
		return m.DeletePermanently()
	}

	return Db().Delete(m).Error
}

// Purge removes a file from the index by marking it as missing.
func (m *File) Purge() error {
	deletedAt := Now()
	m.FileMissing = true
	m.FilePrimary = false
	m.DeletedAt = &deletedAt
	return UnscopedDb().Exec("UPDATE files SET file_missing = 1, file_primary = 0, deleted_at = ? WHERE id = ?", &deletedAt, m.ID).Error
}

// Found restores a previously purged file.
func (m *File) Found() error {
	m.FileMissing = false
	m.DeletedAt = nil
	return UnscopedDb().Exec("UPDATE files SET file_missing = 0, deleted_at = NULL WHERE id = ?", m.ID).Error
}

// AllFilesMissing returns true, if all files for the photo of this file are missing.
func (m *File) AllFilesMissing() bool {
	count := 0

	if err := Db().Model(&File{}).
		Where("photo_id = ? AND file_missing = 0", m.PhotoID).
		Count(&count).Error; err != nil {
		log.Errorf("file: %s", err.Error())
	}

	return count == 0
}

// Create inserts a new row to the database.
func (m *File) Create() error {
	if m.PhotoID == 0 {
		return fmt.Errorf("file: cannot create file with empty photo id")
	}

	if err := UnscopedDb().Create(m).Error; err != nil {
		log.Errorf("file: %s while saving", err)
		return err
	}

	if _, err := m.SaveMarkers(); err != nil {
		log.Errorf("file %s: %s while saving markers", clean.Log(m.FileUID), err)
		return err
	}

	return m.ResolvePrimary()
}

// ResolvePrimary ensures there is only one primary file for a photo.
func (m *File) ResolvePrimary() (err error) {
	filePrimaryMutex.Lock()
	defer filePrimaryMutex.Unlock()

	if !m.FilePrimary {
		return nil
	}

	err = UnscopedDb().
		Exec("UPDATE files SET file_primary = (id = ?) WHERE photo_id = ?", m.ID, m.PhotoID).Error

	if err == nil {
		m.RegenerateIndex()
	}

	return err
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *File) Save() error {
	if m.PhotoID == 0 {
		return fmt.Errorf("file %s: cannot save file with empty photo id", m.FileUID)
	}

	if err := UnscopedDb().Save(m).Error; err != nil {
		log.Errorf("file %s: %s while saving", clean.Log(m.FileUID), err)
		return err
	}

	if _, err := m.SaveMarkers(); err != nil {
		log.Errorf("file %s: %s while saving markers", clean.Log(m.FileUID), err)
		return err
	}

	return m.ResolvePrimary()
}

// UpdateVideoInfos updated related video files so they are properly grouped with the primary image in search results.
// see https://github.com/photoprism/photoprism/pull/3588#issuecomment-1683429455
func (m *File) UpdateVideoInfos() error {
	if m.PhotoID <= 0 {
		return fmt.Errorf("file has invalid photo id")
	}

	// Set the video dimensions from the primary image if it could not be determined from the video metadata.
	// see https://github.com/photoprism/photoprism/blob/develop/internal/photoprism/index_mediafile.go
	dimensions := FileDimensions{}

	if err := deepcopier.Copy(&dimensions).From(m); err != nil {
		return err
	} else if err = Db().Model(File{}).Where("photo_id = ? AND file_video = 1 AND file_width <= 0", m.PhotoID).Updates(dimensions).Error; err != nil {
		return err
	}

	// Set the video appearance from the primary file if it could not be detected e.g. from a JPEG sidecar file.
	// see https://github.com/photoprism/photoprism/blob/develop/internal/photoprism/index_mediafile.go
	appearance := FileAppearance{}

	if err := deepcopier.Copy(&appearance).From(m); err != nil {
		return err
	} else if err = Db().Model(File{}).Where("photo_id = ? AND file_video = 1", m.PhotoID).Updates(appearance).Error; err != nil {
		return err
	}

	return nil
}

// Update updates a column in the database.
func (m *File) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
}

// Updates multiple columns in the database.
func (m *File) Updates(values interface{}) error {
	return UnscopedDb().Model(m).UpdateColumns(values).Error
}

// Rename updates the name and path of this file.
func (m *File) Rename(fileName, rootName, filePath, fileBase string) error {
	log.Debugf("file %s: renaming %s to %s", clean.Log(m.FileUID), clean.Log(m.FileName), clean.Log(fileName))

	// Update database row.
	if err := m.Updates(map[string]interface{}{
		"FileName":    fileName,
		"FileRoot":    rootName,
		"FileMissing": false,
		"DeletedAt":   nil,
	}); err != nil {
		return err
	}

	m.FileName = fileName
	m.FileRoot = rootName
	m.FileMissing = false
	m.DeletedAt = nil

	// Update photo path and name if possible.
	if p := m.RelatedPhoto(); p != nil {
		return p.Updates(map[string]interface{}{
			"PhotoPath": filePath,
			"PhotoName": fileBase,
		})
	}

	return nil
}

// Undelete removes the missing flag from this file.
func (m *File) Undelete() error {
	if !m.Missing() {
		return nil
	}

	// Update database row.
	err := m.Updates(map[string]interface{}{
		"FileMissing": false,
		"DeletedAt":   nil,
	})

	if err != nil {
		return err
	}

	log.Debugf("file %s: removed missing flag from %s", clean.Log(m.FileUID), clean.Log(m.FileName))

	m.FileMissing = false
	m.DeletedAt = nil

	return nil
}

// RelatedPhoto returns the related photo entity.
func (m *File) RelatedPhoto() *Photo {
	if m.Photo != nil {
		return m.Photo
	}

	photo := Photo{}

	UnscopedDb().Model(m).Related(&photo)

	return &photo
}

// NoJPEG returns true if the file is not a JPEG image.
func (m *File) NoJPEG() bool {
	return fs.ImageJPEG.NotEqual(m.FileType)
}

// NoPNG returns true if the file is not a PNG image.
func (m *File) NoPNG() bool {
	return fs.ImagePNG.NotEqual(m.FileType)
}

// Type returns the file type.
func (m *File) Type() fs.Type {
	return fs.Type(m.FileType)
}

// Links returns all share links for this entity.
func (m *File) Links() Links {
	return FindLinks("", m.FileUID)
}

// Panorama checks if the file appears to be a panoramic image.
func (m *File) Panorama() bool {
	if m.FileSidecar || m.FileWidth <= 1000 || m.FileHeight <= 500 {
		// Too small.
		return false
	} else if m.Projection() != projection.Unknown {
		// Panoramic projection.
		return true
	}

	// Decide based on aspect ratio.
	return float64(m.FileWidth)/float64(m.FileHeight) > 1.9
}

// Bounds returns the file dimensions as image.Rectangle.
func (m *File) Bounds() image.Rectangle {
	return image.Rectangle{Min: image.Point{}, Max: image.Point{X: m.FileWidth, Y: m.FileHeight}}
}

// Projection returns the panorama projection name if any.
func (m *File) Projection() projection.Type {
	return projection.New(m.FileProjection)
}

// SetProjection sets the panorama projection name.
func (m *File) SetProjection(s string) {
	if s == "" {
		return
	} else if t := projection.New(s); !t.Unknown() {
		m.FileProjection = t.String()
	}
}

// IsHDR returns true if it is a high dynamic range file.
func (m *File) IsHDR() bool {
	return m.FileHDR
}

// SetHDR sets the high dynamic range flag.
func (m *File) SetHDR(isHdr bool) {
	if isHdr {
		m.FileHDR = true
	}
}

// ResetHDR removes the high dynamic range flag.
func (m *File) ResetHDR() {
	m.FileHDR = false
}

// HasWatermark returns true if the file has a watermark.
func (m *File) HasWatermark() bool {
	return m.FileWatermark
}

// IsAnimated returns true if the file has animated image frames.
func (m *File) IsAnimated() bool {
	return (m.FileFrames > 1 || m.FileDuration > 0) && media.Image.Equal(m.MediaType)
}

// ColorProfile returns the ICC color profile name if any.
func (m *File) ColorProfile() string {
	return SanitizeStringType(m.FileColorProfile)
}

// HasColorProfile tests if the file has a matching color profile.
func (m *File) HasColorProfile(profile colors.Profile) bool {
	return profile.Equal(m.FileColorProfile)
}

// SetColorProfile sets the ICC color profile name such as "Display P3".
func (m *File) SetColorProfile(name string) {
	if name = SanitizeStringType(name); name != "" {
		m.FileColorProfile = SanitizeStringType(name)
	}
}

// ResetColorProfile removes the ICC color profile name.
func (m *File) ResetColorProfile() {
	m.FileColorProfile = ""
}

// SetSoftware sets the software name.
func (m *File) SetSoftware(name string) {
	if name = SanitizeStringType(name); name != "" {
		m.FileSoftware = name
	}
}

// SetDuration sets the video/animation duration.
func (m *File) SetDuration(d time.Duration) {
	if d <= 0 {
		return
	}

	m.FileDuration = d.Round(10e6)

	// Update number of frames.
	if m.FileFrames == 0 && m.FileFPS > 1 {
		m.FileFrames = int(math.Round(m.FileFPS * m.FileDuration.Seconds()))
	}

	// Update number of frames per second.
	if m.FileFPS == 0 && m.FileFrames > 1 {
		m.FileFPS = float64(m.FileFrames) / m.FileDuration.Seconds()
	}
}

// Bitrate returns the average bitrate in MBit/s if the file has a duration.
func (m *File) Bitrate() float64 {
	// Return 0 if file size or video duration are unknown.
	if m.FileSize <= 0 || m.FileDuration <= 0 {
		return 0
	}

	// Divide number of bits through the duration in seconds.
	return ((float64(m.FileSize) * 8) / m.FileDuration.Seconds()) / 1e6
}

// SetFPS sets the average number of frames per second.
func (m *File) SetFPS(frameRate float64) {
	if frameRate <= 0 {
		return
	}

	m.FileFPS = frameRate

	// Update number of frames.
	if m.FileFrames == 0 && m.FileDuration > time.Second {
		m.FileFrames = int(math.Round(m.FileFPS * m.FileDuration.Seconds()))
	}
}

// SetFrames sets the number of video/animation frames.
func (m *File) SetFrames(n int) {
	if n <= 0 {
		return
	}

	m.FileFrames = n

	// Update FPS.
	if m.FileFPS <= 0 && m.FileDuration > 0 {
		m.FileFPS = float64(m.FileFrames) / m.FileDuration.Seconds()
	} else if m.FileFPS == 0 && m.FileDuration == 0 {
		m.FileFPS = 30.0 // Assume 30 frames per second.
		m.FileDuration = time.Duration(float64(m.FileFrames)/m.FileFPS) * time.Second
	}
}

// SetMediaUTC sets the media creation date from metadata as unix time in ms.
func (m *File) SetMediaUTC(taken time.Time) {
	if taken.IsZero() {
		return
	}

	m.MediaUTC = taken.UTC().UnixMilli()
}

// AddFaces adds face markers to the file.
func (m *File) AddFaces(faces face.Faces) {
	sort.Slice(faces, func(i, j int) bool {
		return faces[i].Size() > faces[j].Size()
	})

	for _, f := range faces {
		m.AddFace(f, "")
	}
}

// AddFace adds a face marker to the file.
func (m *File) AddFace(f face.Face, subjUid string) {
	// Only add faces with exactly one embedding so that they can be compared and clustered.
	if !f.Embeddings.One() {
		return
	}

	// Create new marker from face.
	marker := NewFaceMarker(f, *m, subjUid)

	// Failed creating new marker?
	if marker == nil {
		return
	}

	// Append marker if it doesn't conflict with existing marker.
	if markers := m.Markers(); !markers.Contains(*marker) {
		markers.AppendWithEmbedding(*marker)
	}
}

// ValidFaceCount returns the number of valid face markers.
func (m *File) ValidFaceCount() (c int) {
	return ValidFaceCount(m.FileUID)
}

// UpdatePhotoFaceCount updates the faces count in the index and returns it if the file is primary.
func (m *File) UpdatePhotoFaceCount() (c int, err error) {
	// Previews file of an existing photo?
	if !m.FilePrimary || m.PhotoID == 0 {
		return 0, nil
	}

	c = m.ValidFaceCount()

	err = UnscopedDb().Model(Photo{}).
		Where("id = ?", m.PhotoID).
		UpdateColumn("photo_faces", c).Error

	return c, err
}

// SaveMarkers updates markers in the index.
func (m *File) SaveMarkers() (count int, err error) {
	if m.markers == nil {
		return 0, nil
	}

	return m.markers.Save(m)
}

// Markers finds and returns existing file markers.
func (m *File) Markers() *Markers {
	if m.markers != nil {
		return m.markers
	} else if m.FileUID == "" {
		m.markers = &Markers{}
	} else if res, err := FindMarkers(m.FileUID); err != nil {
		log.Warnf("file %s: %s while loading markers", clean.Log(m.FileUID), err)
		m.markers = &Markers{}
	} else {
		m.markers = &res
	}

	return m.markers
}

// UnsavedMarkers tests if any marker hasn't been saved yet.
func (m *File) UnsavedMarkers() bool {
	if m.markers == nil {
		return false
	}

	return m.markers.Unsaved()
}

// SubjectNames returns all known subject names.
func (m *File) SubjectNames() []string {
	return m.Markers().SubjectNames()
}

// Orientation returns the file's Exif orientation value.
func (m *File) Orientation() int {
	return clean.Orientation(m.FileOrientation)
}

// SetOrientation sets the file's Exif orientation value.
func (m *File) SetOrientation(val int, src string) *File {
	// Ignore invalid values.
	val = clean.Orientation(val)
	if val == 0 {
		return m
	}

	// Only set values with a matching or higher priority.
	if SrcPriority[src] >= SrcPriority[m.FileOrientationSrc] {
		m.FileOrientation = val
		m.FileOrientationSrc = src
	}

	return m
}
