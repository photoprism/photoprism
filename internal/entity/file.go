package entity

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/internal/face"

	"github.com/photoprism/photoprism/pkg/colors"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

type DownloadName string

const (
	DownloadNameFile     DownloadName = "file"
	DownloadNameOriginal DownloadName = "original"
	DownloadNameShare    DownloadName = "share"
	DownloadNameDefault               = DownloadNameFile
)

// Files represents a file result set.
type Files []File

// Index updates should not run simultaneously.
var fileIndexMutex = sync.Mutex{}
var filePrimaryMutex = sync.Mutex{}

// File represents an image or sidecar file that belongs to a photo.
type File struct {
	ID               uint          `gorm:"primary_key" json:"-" yaml:"-"`
	Photo            *Photo        `json:"-" yaml:"-"`
	PhotoID          uint          `gorm:"index:idx_files_photo_id;" json:"-" yaml:"-"`
	PhotoUID         string        `gorm:"type:VARBINARY(42);index;" json:"PhotoUID" yaml:"PhotoUID"`
	PhotoTakenAt     time.Time     `gorm:"type:DATETIME;index;" json:"TakenAt" yaml:"TakenAt"`
	TimeIndex        *string       `gorm:"type:VARBINARY(48);" json:"TimeIndex" yaml:"TimeIndex"`
	MediaID          *string       `gorm:"type:VARBINARY(32);" json:"MediaID" yaml:"MediaID"`
	InstanceID       string        `gorm:"type:VARBINARY(42);index;" json:"InstanceID,omitempty" yaml:"InstanceID,omitempty"`
	FileUID          string        `gorm:"type:VARBINARY(42);unique_index;" json:"UID" yaml:"UID"`
	FileName         string        `gorm:"type:VARBINARY(755);unique_index:idx_files_name_root;" json:"Name" yaml:"Name"`
	FileRoot         string        `gorm:"type:VARBINARY(16);default:'/';unique_index:idx_files_name_root;" json:"Root" yaml:"Root,omitempty"`
	OriginalName     string        `gorm:"type:VARBINARY(755);" json:"OriginalName" yaml:"OriginalName,omitempty"`
	FileHash         string        `gorm:"type:VARBINARY(128);index" json:"Hash" yaml:"Hash,omitempty"`
	FileSize         int64         `json:"Size" yaml:"Size,omitempty"`
	FileCodec        string        `gorm:"type:VARBINARY(32)" json:"Codec" yaml:"Codec,omitempty"`
	FileType         string        `gorm:"type:VARBINARY(32)" json:"Type" yaml:"Type,omitempty"`
	FileMime         string        `gorm:"type:VARBINARY(64)" json:"Mime" yaml:"Mime,omitempty"`
	FilePrimary      bool          `gorm:"index:idx_files_photo_id;" json:"Primary" yaml:"Primary,omitempty"`
	FileSidecar      bool          `json:"Sidecar" yaml:"Sidecar,omitempty"`
	FileMissing      bool          `json:"Missing" yaml:"Missing,omitempty"`
	FilePortrait     bool          `json:"Portrait" yaml:"Portrait,omitempty"`
	FileVideo        bool          `json:"Video" yaml:"Video,omitempty"`
	FileDuration     time.Duration `json:"Duration" yaml:"Duration,omitempty"`
	FileWidth        int           `json:"Width" yaml:"Width,omitempty"`
	FileHeight       int           `json:"Height" yaml:"Height,omitempty"`
	FileOrientation  int           `json:"Orientation" yaml:"Orientation,omitempty"`
	FileProjection   string        `gorm:"type:VARBINARY(64);" json:"Projection,omitempty" yaml:"Projection,omitempty"`
	FileAspectRatio  float32       `gorm:"type:FLOAT;" json:"AspectRatio" yaml:"AspectRatio,omitempty"`
	FileHDR          bool          `gorm:"column:file_hdr;"  json:"IsHDR" yaml:"IsHDR,omitempty"`
	FileColorProfile string        `gorm:"type:VARBINARY(64);" json:"ColorProfile,omitempty" yaml:"ColorProfile,omitempty"`
	FileMainColor    string        `gorm:"type:VARBINARY(16);index;" json:"MainColor" yaml:"MainColor,omitempty"`
	FileColors       string        `gorm:"type:VARBINARY(9);" json:"Colors" yaml:"Colors,omitempty"`
	FileLuminance    string        `gorm:"type:VARBINARY(9);" json:"Luminance" yaml:"Luminance,omitempty"`
	FileDiff         uint32        `json:"Diff" yaml:"Diff,omitempty"`
	FileChroma       uint8         `json:"Chroma" yaml:"Chroma,omitempty"`
	FileError        string        `gorm:"type:VARBINARY(512)" json:"Error" yaml:"Error,omitempty"`
	ModTime          int64         `json:"ModTime" yaml:"-"`
	CreatedAt        time.Time     `json:"CreatedAt" yaml:"-"`
	CreatedIn        int64         `json:"CreatedIn" yaml:"-"`
	UpdatedAt        time.Time     `json:"UpdatedAt" yaml:"-"`
	UpdatedIn        int64         `json:"UpdatedIn" yaml:"-"`
	DeletedAt        *time.Time    `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
	Share            []FileShare   `json:"-" yaml:"-"`
	Sync             []FileSync    `json:"-" yaml:"-"`
	markers          *Markers
}

// TableName returns the entity database table name.
func (File) TableName() string {
	return "files"
}

// RegenerateIndex updates the search index columns.
func (m File) RegenerateIndex() {
	fileIndexMutex.Lock()
	defer fileIndexMutex.Unlock()

	start := time.Now()
	filesTable := File{}.TableName()
	photosTable := Photo{}.TableName()

	var updateWhere *gorm.SqlExpr
	var scope string

	if m.PhotoID > 0 {
		updateWhere = gorm.Expr("photo_id = ?", m.PhotoID)
		scope = "partial"
	} else if m.PhotoUID != "" {
		updateWhere = gorm.Expr("photo_uid = ?", m.PhotoUID)
		scope = "partial"
	} else if m.ID > 0 {
		updateWhere = gorm.Expr("id = ?", m.ID)
		scope = "file"
	} else {
		updateWhere = gorm.Expr("photo_id IS NOT NULL")
		scope = "full"
	}

	switch DbDialect() {
	case MySQL:
		Log("files", "regenerate photo_taken_at",
			Db().Exec("UPDATE ? f JOIN ? p ON p.id = f.photo_id SET f.photo_taken_at = p.taken_at_local WHERE ?",
				gorm.Expr(filesTable), gorm.Expr(photosTable), updateWhere).Error)

		Log("files", "regenerate media_id",
			Db().Exec("UPDATE ? SET media_id = CASE WHEN file_missing = 0 AND deleted_at IS NULL THEN CONCAT(HEX(100000000000 - photo_id), '-', 1 + file_sidecar - file_primary, '-', file_uid) ELSE NULL END WHERE ?",
				gorm.Expr(filesTable), updateWhere).Error)

		Log("files", "regenerate time_index",
			Db().Exec("UPDATE ? SET time_index = CASE WHEN media_id IS NOT NULL AND photo_taken_at IS NOT NULL THEN CONCAT(100000000000000 - CAST(photo_taken_at AS UNSIGNED), '-', media_id) ELSE NULL END WHERE ?",
				gorm.Expr(filesTable), updateWhere).Error)
	case SQLite3:
		Log("files", "regenerate photo_taken_at",
			Db().Exec("UPDATE ? SET photo_taken_at = (SELECT p.taken_at_local FROM ? p WHERE p.id = photo_id) WHERE ?",
				gorm.Expr(filesTable), gorm.Expr(photosTable), updateWhere).Error)

		Log("files", "regenerate media_id",
			Db().Exec("UPDATE ? SET media_id = CASE WHEN file_missing = 0 AND deleted_at IS NULL THEN (HEX(100000000000 - photo_id) || '-' || (1 + file_sidecar - file_primary) || '-' || file_uid) ELSE NULL END WHERE ?",
				gorm.Expr(filesTable), updateWhere).Error)

		Log("files", "regenerate time_index",
			Db().Exec("UPDATE ? SET time_index = CASE WHEN media_id IS NOT NULL AND photo_taken_at IS NOT NULL THEN ((100000000000000 - CAST(photo_taken_at AS UNSIGNED)) || '-' || media_id) ELSE NULL END WHERE ?",
				gorm.Expr(filesTable), updateWhere).Error)
	default:
		log.Warnf("sql: unsupported dialect %s", DbDialect())
	}

	log.Debugf("files: updated %s search index [%s]", scope, time.Since(start))
}

type FileInfos struct {
	FileWidth       int
	FileHeight      int
	FileOrientation int
	FileAspectRatio float32
	FileMainColor   string
	FileColors      string
	FileLuminance   string
	FileDiff        uint32
	FileChroma      uint8
}

// FirstFileByHash gets a file in db from its hash
func FirstFileByHash(fileHash string) (File, error) {
	var file File

	res := Db().Unscoped().First(&file, "file_hash = ?", fileHash)

	return file, res.Error
}

// PrimaryFile returns the primary file for a photo uid.
func PrimaryFile(photoUID string) (*File, error) {
	file := File{}

	res := Db().Unscoped().First(&file, "file_primary = 1 AND photo_uid = ?", photoUID)

	return &file, res.Error
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *File) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUID(m.FileUID, 'f') {
		return nil
	}

	return scope.SetColumn("FileUID", rnd.PPID('f'))
}

// DownloadName returns the download file name.
func (m *File) DownloadName(n DownloadName, seq int) string {
	switch n {
	case DownloadNameFile:
		return m.Base(seq)
	case DownloadNameOriginal:
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

	name := strings.Title(slug.MakeLang(photo.PhotoTitle, "en"))
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
	if m.ModTime == modTime.Unix() {
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
		return fmt.Errorf("invalid file id %d / uid %s", m.ID, sanitize.Log(m.FileUID))
	}

	if err := UnscopedDb().Delete(Marker{}, "file_uid = ?", m.FileUID).Error; err != nil {
		log.Errorf("file %s: %s while removing markers", sanitize.Log(m.FileUID), err)
	}

	if err := UnscopedDb().Delete(FileShare{}, "file_id = ?", m.ID).Error; err != nil {
		log.Errorf("file %s: %s while removing share info", sanitize.Log(m.FileUID), err)
	}

	if err := UnscopedDb().Delete(FileSync{}, "file_id = ?", m.ID).Error; err != nil {
		log.Errorf("file %s: %s while removing remote sync info", sanitize.Log(m.FileUID), err)
	}

	if err := m.ReplaceHash(""); err != nil {
		log.Errorf("file %s: %s while removing covers", sanitize.Log(m.FileUID), err)
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
		log.Tracef("file %s: removing hash %s", sanitize.Log(m.FileUID), sanitize.Log(m.FileHash))
	} else if m.FileHash != "" && newHash != "" {
		log.Tracef("file %s: hash %s changed to %s", sanitize.Log(m.FileUID), sanitize.Log(m.FileHash), sanitize.Log(newHash))
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

		if res := UnscopedDb().Model(entity).Where("thumb = ?", oldHash).Update("thumb", newHash); res.Error != nil {
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
		return fmt.Errorf("invalid file id %d / uid %s", m.ID, sanitize.Log(m.FileUID))
	}

	if permanently {
		return m.DeletePermanently()
	}

	return Db().Delete(m).Error
}

// Purge removes a file from the index by marking it as missing.
func (m *File) Purge() error {
	deletedAt := TimeStamp()
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
		log.Errorf("file %s: %s while saving markers", sanitize.Log(m.FileUID), err)
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

// Save stores the file in the database.
func (m *File) Save() error {
	if m.PhotoID == 0 {
		return fmt.Errorf("file %s: cannot save file with empty photo id", m.FileUID)
	}

	if err := UnscopedDb().Save(m).Error; err != nil {
		log.Errorf("file %s: %s while saving", sanitize.Log(m.FileUID), err)
		return err
	}

	if _, err := m.SaveMarkers(); err != nil {
		log.Errorf("file %s: %s while saving markers", sanitize.Log(m.FileUID), err)
		return err
	}

	return m.ResolvePrimary()
}

// UpdateVideoInfos updates related video infos based on this file.
func (m *File) UpdateVideoInfos() error {
	values := FileInfos{}

	if err := deepcopier.Copy(&values).From(m); err != nil {
		return err
	}

	return Db().Model(File{}).Where("photo_id = ? AND file_video = 1", m.PhotoID).Updates(values).Error
}

// Update updates a column in the database.
func (m *File) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).Update(attr, value).Error
}

// Updates multiple columns in the database.
func (m *File) Updates(values interface{}) error {
	return UnscopedDb().Model(m).Updates(values).Error
}

// Rename updates the name and path of this file.
func (m *File) Rename(fileName, rootName, filePath, fileBase string) error {
	log.Debugf("file %s: renaming %s to %s", sanitize.Log(m.FileUID), sanitize.Log(m.FileName), sanitize.Log(fileName))

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

	log.Debugf("file %s: removed missing flag from %s", sanitize.Log(m.FileUID), sanitize.Log(m.FileName))

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

// NoJPEG returns true if the file is not a JPEG image file.
func (m *File) NoJPEG() bool {
	return m.FileType != string(fs.FormatJpeg)
}

// Links returns all share links for this entity.
func (m *File) Links() Links {
	return FindLinks("", m.FileUID)
}

// Panorama tests if the file seems to be a panorama image.
func (m *File) Panorama() bool {
	if m.FileSidecar || m.FileWidth <= 1000 || m.FileHeight <= 500 {
		return false
	}

	return m.Projection() != ProjDefault || (m.FileWidth/m.FileHeight) >= 2
}

// Projection returns the panorama projection name if any.
func (m *File) Projection() string {
	return SanitizeStringTypeLower(m.FileProjection)
}

// SetProjection sets the panorama projection name.
func (m *File) SetProjection(name string) {
	m.FileProjection = SanitizeStringTypeLower(name)
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
func (m *File) AddFace(f face.Face, subjUID string) {
	// Only add faces with exactly one embedding so that they can be compared and clustered.
	if !f.Embeddings.One() {
		return
	}

	// Create new marker from face.
	marker := NewFaceMarker(f, *m, subjUID)

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
	// Primary file of an existing photo?
	if !m.FilePrimary || m.PhotoID == 0 {
		return 0, nil
	}

	c = m.ValidFaceCount()

	err = UnscopedDb().Model(Photo{}).
		Where("id = ?", m.PhotoID).
		Update("photo_faces", c).Error

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
		log.Warnf("file %s: %s while loading markers", sanitize.Log(m.FileUID), err)
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
