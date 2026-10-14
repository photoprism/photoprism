package migrate

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// repairPhoto is a photo model with the columns that the conversion repairs.
type repairPhoto struct {
	ID        uint   `gorm:"primaryKey"`
	PhotoPath string `gorm:"type:bytes;size:1024"`
	PhotoName string `gorm:"type:bytes;size:255"`
}

// TableName returns the entity table name.
func (repairPhoto) TableName() string {
	return "photos"
}

// repairFile is a file model with the columns the photo and folder repairs read.
type repairFile struct {
	ID          uint   `gorm:"primaryKey"`
	PhotoID     uint   `gorm:"size:32"`
	FileRoot    string `gorm:"type:bytes;size:16"`
	FileName    string `gorm:"type:bytes;size:1024"`
	FilePrimary bool
}

// TableName returns the entity table name.
func (repairFile) TableName() string {
	return "files"
}

// repairFolder is a folder model with the columns that the conversion repairs.
type repairFolder struct {
	Path      string `gorm:"type:bytes;size:1024;uniqueIndex:idx_folders_path_root"`
	Root      string `gorm:"type:bytes;size:16;uniqueIndex:idx_folders_path_root"`
	FolderUID string `gorm:"type:bytes;size:42;primaryKey"`
}

// TableName returns the entity table name.
func (repairFolder) TableName() string {
	return "folders"
}

// repairAlbum is an album model with the columns that the conversion repairs.
type repairAlbum struct {
	ID         uint   `gorm:"primaryKey"`
	AlbumUID   string `gorm:"type:bytes;size:42"`
	AlbumTitle string `gorm:"size:160"`
	AlbumSlug  string `gorm:"type:bytes;size:160"`
	AlbumPath  string `gorm:"type:bytes;size:1024"`
	AlbumType  string `gorm:"type:bytes;size:8"`
}

// TableName returns the entity table name.
func (repairAlbum) TableName() string {
	return "albums"
}

// legacyRepairDb returns a SQLite database with legacy tables whose path and name columns hold numbers.
func legacyRepairDb(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := openTestDb("sqlite3", filepath.Join(t.TempDir(), "repair.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = closeTestDb(db) })

	for _, stmt := range []string{
		"CREATE TABLE photos (id integer primary key autoincrement, photo_path VARBINARY(1024), photo_name VARBINARY(255))",
		"CREATE TABLE files (id integer primary key autoincrement, photo_id integer, file_root VARBINARY(16), file_name VARBINARY(1024), file_primary bool)",
		"CREATE TABLE folders (path VARBINARY(1024), root VARBINARY(16), folder_uid VARBINARY(42), PRIMARY KEY (folder_uid))",
		"CREATE TABLE albums (id integer primary key autoincrement, album_uid VARBINARY(42), album_title VARCHAR(160), " +
			"album_slug VARBINARY(160), album_path VARBINARY(1024), album_type VARBINARY(8))",
		"INSERT INTO photos (id, photo_path, photo_name) VALUES (1, '', '0001'), (2, '01', '0002'), (3, 'scans', '1e5'), " +
			"(4, 'scans', 'abc'), (5, 'scans', '0009'), (6, '2020', '0006')",
		"INSERT INTO files (id, photo_id, file_root, file_name, file_primary) VALUES " +
			"(1, 1, '/', '1.dng', 0), (2, 1, '/', '0001.jpg', 1), (3, 2, '/', '01/0002.jpg', 1), (4, 3, '/', 'scans/1e5.jpg', 1), " +
			"(5, 4, '/', 'scans/abc.jpg', 1), (6, 6, '/', '2020/0006 (2).jpg', 1), " +
			"(7, 0, 'sidecar', '02/a.jpg', 0), (8, 0, 'sidecar', '2/b.jpg', 0)",
		"INSERT INTO folders (path, root, folder_uid) VALUES ('01', '/', 'd1'), ('01/sub', '/', 'd2'), ('2020', '/', 'd3'), " +
			"('007', '/', 'd4'), ('02', 'sidecar', 'd5'), ('', '/', 'd6')",
		"INSERT INTO albums (id, album_uid, album_title, album_slug, album_path, album_type) VALUES " +
			"(1, 'a1', '01', '01', '01', 'folder'), (2, 'a2', '007', '007', '05', 'album'), (3, 'a3', '2020', '2020', '2020', 'folder')",
	} {
		require.NoError(t, db.Exec(stmt).Error, stmt)
	}

	return db
}

// repairModels are the models of the tables in legacyRepairDb.
var repairModels = map[string]any{
	"photos":  &repairPhoto{},
	"files":   &repairFile{},
	"folders": &repairFolder{},
	"albums":  &repairAlbum{},
}

func TestConvertSQLiteSchema_Repair(t *testing.T) {
	// Tables are converted in a fixed order per run, so that the result is shown not to depend on it.
	orders := [][]string{
		{"photos", "files", "folders", "albums"},
		{"files", "folders", "albums", "photos"},
		{"albums", "folders", "photos", "files"},
		{"folders", "albums", "files", "photos"},
	}

	for _, order := range orders {
		db := legacyRepairDb(t)

		for _, name := range order {
			require.NoError(t, ConvertSQLiteSchema(db, map[string]any{name: repairModels[name]}), name)
		}

		var photos []repairPhoto
		require.NoError(t, db.Order("id").Find(&photos).Error)
		require.Len(t, photos, 6)
		assert.Equal(t, repairPhoto{ID: 1, PhotoPath: "", PhotoName: "0001"}, photos[0], "primary file")
		assert.Equal(t, repairPhoto{ID: 2, PhotoPath: "01", PhotoName: "0002"}, photos[1])
		assert.Equal(t, repairPhoto{ID: 3, PhotoPath: "scans", PhotoName: "1e5"}, photos[2])
		assert.Equal(t, repairPhoto{ID: 4, PhotoPath: "scans", PhotoName: "abc"}, photos[3])
		assert.Equal(t, repairPhoto{ID: 5, PhotoPath: "scans", PhotoName: "9"}, photos[4], "no files")
		assert.Equal(t, repairPhoto{ID: 6, PhotoPath: "2020", PhotoName: "0006"}, photos[5], "stack name")

		var folders []repairFolder
		require.NoError(t, db.Order("folder_uid").Find(&folders).Error)
		require.Len(t, folders, 6)
		assert.Equal(t, "01", folders[0].Path)
		assert.Equal(t, "01/sub", folders[1].Path)
		assert.Equal(t, "2020", folders[2].Path)
		assert.Equal(t, "7", folders[3].Path, "no files or subfolders")
		assert.Equal(t, "2", folders[4].Path, "ambiguous")
		assert.Equal(t, "", folders[5].Path)

		var albums []repairAlbum
		require.NoError(t, db.Order("id").Find(&albums).Error)
		require.Len(t, albums, 3)
		assert.Equal(t, "01", albums[0].AlbumPath)
		assert.Equal(t, "5", albums[1].AlbumPath, "not a folder album")
		assert.Equal(t, "2020", albums[2].AlbumPath)

		var files []repairFile
		require.NoError(t, db.Order("id").Find(&files).Error)
		require.Len(t, files, 8)
		assert.Equal(t, "0001.jpg", files[1].FileName)
	}
}

// sqliteInt returns the number SQLite stores for an integer value.
func sqliteInt(i int64) sqliteNumber {
	return sqliteNumber{Value: float64(i), Int: &i, Text: strconv.FormatInt(i, 10)}
}

func TestSqliteMatch(t *testing.T) {
	t.Run("Integer", func(t *testing.T) {
		match, ok := sqliteMatch(sqliteInt(1), []string{"abc", "0001"})
		assert.True(t, ok)
		assert.Equal(t, "0001", match)
	})
	t.Run("Exponent", func(t *testing.T) {
		match, ok := sqliteMatch(sqliteInt(100000), []string{"1e5"})
		assert.True(t, ok)
		assert.Equal(t, "1e5", match)
	})
	t.Run("Real", func(t *testing.T) {
		match, ok := sqliteMatch(sqliteNumber{Value: 1.5, Text: "1.5"}, []string{"01.50"})
		assert.True(t, ok)
		assert.Equal(t, "01.50", match)
	})
	t.Run("Repeated", func(t *testing.T) {
		match, ok := sqliteMatch(sqliteInt(1), []string{"01", "01"})
		assert.True(t, ok)
		assert.Equal(t, "01", match)
	})
	t.Run("Ambiguous", func(t *testing.T) {
		_, ok := sqliteMatch(sqliteInt(1), []string{"01", "1"})
		assert.False(t, ok)
	})
	t.Run("NoMatch", func(t *testing.T) {
		_, ok := sqliteMatch(sqliteInt(2), []string{"01", "abc", ""})
		assert.False(t, ok)
	})
}

func TestSqliteNumber_Equals(t *testing.T) {
	t.Run("LargeInteger", func(t *testing.T) {
		n := sqliteInt(9007199254740992)
		assert.True(t, n.Equals("9007199254740992"))
		assert.False(t, n.Equals("9007199254740993"))
		assert.True(t, n.Equals("9.007199254740993e15"), "SQLite stores this real as the integer it rounds to")
	})
	t.Run("Spaces", func(t *testing.T) {
		assert.True(t, sqliteInt(12).Equals(" 12 "))
		assert.True(t, sqliteInt(12).Equals("\v12\t"))
	})
	t.Run("LargeIntegerFromReal", func(t *testing.T) {
		assert.True(t, sqliteInt(100000000000000000).Equals("1e17"))
		assert.False(t, sqliteInt(100000000000000001).Equals("1e17"))
	})
	t.Run("NotNumeric", func(t *testing.T) {
		assert.False(t, sqliteNumber{Value: 16, Text: "16.0"}.Equals("0x1p4"))
		assert.False(t, sqliteNumber{Value: 0, Text: "0"}.Equals("nan"))
		assert.False(t, sqliteNumber{Value: 0, Text: "0"}.Equals("inf"))
		assert.False(t, sqliteInt(1).Equals("1_0"))
	})
	t.Run("Sign", func(t *testing.T) {
		assert.True(t, sqliteInt(1).Equals("+1"))
		assert.True(t, sqliteInt(-1).Equals("-01"))
	})
}

func TestRepairSQLitePhotos(t *testing.T) {
	t.Run("NoFilesTable", func(t *testing.T) {
		db := legacyRepairDb(t)
		require.NoError(t, db.Exec("DROP TABLE files").Error)
		assert.Error(t, repairSQLitePhotos(db, "photos"))
	})
	t.Run("ConversionContinues", func(t *testing.T) {
		db := legacyRepairDb(t)
		require.NoError(t, db.Exec("DROP TABLE files").Error)
		require.NoError(t, ConvertSQLiteSchema(db, map[string]any{"photos": &repairPhoto{}}))
		var name string
		require.NoError(t, db.Raw("SELECT photo_name FROM photos WHERE id = 1").Row().Scan(&name))
		assert.Equal(t, "1", name)
	})
	t.Run("NumericFileName", func(t *testing.T) {
		// A file name that is a number itself is no candidate, as it may have been converted too.
		db := legacyRepairDb(t)
		require.NoError(t, ConvertSQLiteSchema(db, map[string]any{"files": &repairFile{}}))
		require.NoError(t, db.Exec("INSERT INTO files (id, photo_id, file_root, file_name, file_primary) VALUES "+
			"(9, 5, '/', '9', 1), (10, 5, '/', 'scans/0009.jpg', 0)").Error)
		require.NoError(t, ConvertSQLiteSchema(db, map[string]any{"photos": &repairPhoto{}}))
		var name string
		require.NoError(t, db.Raw("SELECT photo_name FROM photos WHERE id = 5").Row().Scan(&name))
		assert.Equal(t, "0009", name)
	})
}

func TestRepairSQLiteFolders(t *testing.T) {
	t.Run("NoTable", func(t *testing.T) {
		db := legacyRepairDb(t)
		assert.Error(t, repairSQLiteFolders(db, "missing"))
	})
}

func TestRepairSQLiteAlbums(t *testing.T) {
	t.Run("NoFoldersTable", func(t *testing.T) {
		db := legacyRepairDb(t)
		require.NoError(t, db.Exec("DROP TABLE folders").Error)
		assert.Error(t, repairSQLiteAlbums(db, "albums"))
	})
}

func TestSqliteNumbers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db := legacyRepairDb(t)
		numbers, err := sqliteNumbers(db, "folders", "folder_uid", "root", "", "path")
		require.NoError(t, err)
		require.Len(t, numbers, 4)
		byKey := make(map[string]sqliteNumber)
		for _, n := range numbers {
			byKey[n.Key] = n
		}
		require.NotNil(t, byKey["d1"].Int)
		assert.Equal(t, int64(1), *byKey["d1"].Int)
		assert.Equal(t, "1", byKey["d1"].Text)
		assert.Equal(t, "/", byKey["d1"].Root)
		assert.Equal(t, "path", byKey["d1"].Column)
		assert.Equal(t, "sidecar", byKey["d5"].Root)
	})
	t.Run("Where", func(t *testing.T) {
		db := legacyRepairDb(t)
		numbers, err := sqliteNumbers(db, "albums", "album_uid", "''", " AND album_type = 'folder'", "album_path")
		require.NoError(t, err)
		require.Len(t, numbers, 2)
	})
	t.Run("NoTable", func(t *testing.T) {
		db := legacyRepairDb(t)
		_, err := sqliteNumbers(db, "missing", "id", "''", "", "path")
		assert.Error(t, err)
	})
}

func TestSqliteTopFolders(t *testing.T) {
	t.Run("Originals", func(t *testing.T) {
		db := legacyRepairDb(t)
		tops, err := sqliteTopFolders(db, "folders", "/")
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"", "01", "2020", "scans"}, tops)
	})
	t.Run("Sidecar", func(t *testing.T) {
		db := legacyRepairDb(t)
		tops, err := sqliteTopFolders(db, "folders", "sidecar")
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"02", "2"}, tops)
	})
}

func TestSqliteRestore(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Converted columns have blob affinity, so the restored string stays text.
		db := legacyRepairDb(t)
		require.NoError(t, db.Exec("CREATE TABLE restore_tests (uid blob(0), path blob(0))").Error)
		require.NoError(t, db.Exec("INSERT INTO restore_tests (uid, path) VALUES ('d1', '1')").Error)
		assert.True(t, sqliteRestore(db, "restore_tests", "uid", sqliteNumber{Key: "d1", Column: "path"}, "01"))
		var path string
		require.NoError(t, db.Raw("SELECT path FROM restore_tests WHERE uid = 'd1'").Row().Scan(&path))
		assert.Equal(t, "01", path)
	})
	t.Run("Unchanged", func(t *testing.T) {
		// The value equals the converted text, so the row is not written.
		db := legacyRepairDb(t)
		require.NoError(t, db.Exec("CREATE TABLE restore_tests (uid blob(0), path blob(0))").Error)
		require.NoError(t, db.Exec("INSERT INTO restore_tests (uid, path) VALUES ('d1', 'converted')").Error)
		assert.False(t, sqliteRestore(db, "restore_tests", "uid", sqliteNumber{Key: "d1", Column: "path", Text: "2020"}, "2020"))
		var path string
		require.NoError(t, db.Raw("SELECT path FROM restore_tests WHERE uid = 'd1'").Row().Scan(&path))
		assert.Equal(t, "converted", path)
	})
	t.Run("Error", func(t *testing.T) {
		db := legacyRepairDb(t)
		assert.False(t, sqliteRestore(db, "folders", "folder_uid", sqliteNumber{Key: "d1", Column: "missing"}, "01"))
	})
}

func TestLogRestored(t *testing.T) {
	logger, ok := log.(*logrus.Logger)
	require.True(t, ok)

	t.Run("Found", func(t *testing.T) {
		hook := logtest.NewLocal(logger)
		t.Cleanup(hook.Reset)
		logRestored("photos", 1, 2)
		require.NotNil(t, hook.LastEntry())
		assert.Contains(t, hook.LastEntry().Message, "restored 1 of 2 names stored as numbers in photos")
	})
	t.Run("None", func(t *testing.T) {
		hook := logtest.NewLocal(logger)
		t.Cleanup(hook.Reset)
		logRestored("photos", 0, 0)
		assert.Empty(t, hook.AllEntries())
	})
}

// repairLabel is a label model with the slug columns that the conversion repairs.
type repairLabel struct {
	ID         uint   `gorm:"primaryKey"`
	LabelUID   string `gorm:"type:bytes;size:42"`
	LabelName  string `gorm:"size:160"`
	LabelSlug  string `gorm:"type:bytes;size:160"`
	CustomSlug string `gorm:"type:bytes;size:160"`
}

// TableName returns the entity table name.
func (repairLabel) TableName() string {
	return "labels"
}

// repairCamera is a camera model with the slug column that the conversion repairs.
type repairCamera struct {
	ID         uint   `gorm:"primaryKey"`
	CameraName string `gorm:"size:160"`
	CameraSlug string `gorm:"type:bytes;size:160"`
}

// TableName returns the entity table name.
func (repairCamera) TableName() string {
	return "cameras"
}

// repairSubject is a subject model with the slug column that the conversion repairs.
type repairSubject struct {
	SubjUID  string `gorm:"type:bytes;size:42;primaryKey"`
	SubjName string `gorm:"size:160"`
	SubjSlug string `gorm:"type:bytes;size:160"`
}

// TableName returns the entity table name.
func (repairSubject) TableName() string {
	return "subjects"
}

// repairLink is a link model with the slug and token columns that the conversion checks.
type repairLink struct {
	LinkUID   string `gorm:"type:bytes;size:42;primaryKey"`
	ShareUID  string `gorm:"type:bytes;size:42"`
	ShareSlug string `gorm:"type:bytes;size:160"`
	LinkToken string `gorm:"type:bytes;size:160"`
}

// TableName returns the entity table name.
func (repairLink) TableName() string {
	return "links"
}

func TestConvertSQLiteSchema_RepairSlugs(t *testing.T) {
	db := legacyRepairDb(t)

	for _, stmt := range []string{
		"CREATE TABLE labels (id integer primary key autoincrement, label_uid VARBINARY(42), label_name VARCHAR(160), " +
			"label_slug VARBINARY(160), custom_slug VARBINARY(160))",
		"CREATE TABLE links (link_uid VARBINARY(42), share_uid VARBINARY(42), share_slug VARBINARY(160), " +
			"link_token VARBINARY(160), PRIMARY KEY (link_uid))",
		"INSERT INTO labels (id, label_uid, label_name, label_slug, custom_slug) VALUES (1, 'l1', '007', '007', '007'), (2, 'l2', 'Cat', 'cat', 'cat')",
		"CREATE TABLE cameras (id integer primary key autoincrement, camera_name VARCHAR(160), camera_slug VARBINARY(160))",
		"INSERT INTO cameras (id, camera_name, camera_slug) VALUES (5, '0815', '0815')",
		"CREATE TABLE subjects (subj_uid VARBINARY(42), subj_name VARCHAR(160), subj_slug VARBINARY(160), PRIMARY KEY (subj_uid))",
		"INSERT INTO subjects (subj_uid, subj_name, subj_slug) VALUES ('j1', '0042', '0042')",
		"INSERT INTO links (link_uid, share_uid, share_slug, link_token) VALUES ('s3', 'j1', '0042', 'def')",
		"INSERT INTO links (link_uid, share_uid, share_slug, link_token) VALUES ('s1', 'a2', '007', '012345'), ('s2', 'l1', '007', 'abc')",
	} {
		require.NoError(t, db.Exec(stmt).Error, stmt)
	}

	models := map[string]any{"labels": &repairLabel{}, "links": &repairLink{}, "albums": &repairAlbum{},
		"cameras": &repairCamera{}, "subjects": &repairSubject{}}
	require.NoError(t, ConvertSQLiteSchema(db, models))

	var labels []repairLabel
	require.NoError(t, db.Order("id").Find(&labels).Error)
	require.Len(t, labels, 2)
	assert.Equal(t, "007", labels[0].LabelSlug)
	assert.Equal(t, "007", labels[0].CustomSlug)
	assert.Equal(t, "cat", labels[1].LabelSlug)

	var albums []repairAlbum
	require.NoError(t, db.Order("id").Find(&albums).Error)
	require.Len(t, albums, 3)
	assert.Equal(t, "01", albums[0].AlbumSlug)
	assert.Equal(t, "007", albums[1].AlbumSlug)
	assert.Equal(t, "2020", albums[2].AlbumSlug)

	var links []repairLink
	require.NoError(t, db.Order("link_uid").Find(&links).Error)
	require.Len(t, links, 3)
	assert.Equal(t, "007", links[0].ShareSlug, "album title")
	assert.Equal(t, "007", links[1].ShareSlug, "label name")
	assert.Equal(t, "0042", links[2].ShareSlug, "subject name")

	var camera repairCamera
	require.NoError(t, db.First(&camera, 5).Error)
	assert.Equal(t, "0815", camera.CameraSlug)

	var subject repairSubject
	require.NoError(t, db.First(&subject, "subj_uid = ?", "j1").Error)
	assert.Equal(t, "0042", subject.SubjSlug)
	assert.Equal(t, "12345", links[0].LinkToken, "found by numeric value")
	assert.Equal(t, "abc", links[1].LinkToken)
}

func TestSqliteSlugRepair(t *testing.T) {
	t.Run("NoTable", func(t *testing.T) {
		db := legacyRepairDb(t)
		assert.Error(t, sqliteSlugRepair("labels", "label_uid", "label_name", "label_slug")(db, "missing"))
	})
}

func TestRepairSQLiteLinks(t *testing.T) {
	t.Run("NoTable", func(t *testing.T) {
		db := legacyRepairDb(t)
		assert.Error(t, repairSQLiteLinks(db, "missing"))
	})
}
