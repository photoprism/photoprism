package entity

import (
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/schema"

	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestFileSync_TableName(t *testing.T) {
	fileSync := &FileSync{}
	assert.Equal(t, "files_sync", fileSync.TableName())
}

func TestNewFileSync(t *testing.T) {
	r := NewFileSync(123, "test")
	assert.IsType(t, &FileSync{}, r)
	assert.Equal(t, uint(0x7b), r.ServiceID)
	assert.Equal(t, "test", r.RemoteName)
	assert.Equal(t, "new", r.Status)
}

func TestFirstOrCreateFileSync(t *testing.T) {
	t.Run("NotYetExisting", func(t *testing.T) {
		fileSync := &FileSync{ServiceID: 123, FileID: 888, RemoteName: "test123"}
		result := FirstOrCreateFileSync(fileSync)

		if result == nil {
			t.Fatal("result must not be nil")
		}

		if result.FileID != fileSync.FileID {
			t.Errorf("FileID should be the same: %d %d", result.FileID, fileSync.FileID)
		}

		if result.ServiceID != fileSync.ServiceID {
			t.Errorf("ServiceID should be the same: %d %d", result.ServiceID, fileSync.ServiceID)
		}
	})
	t.Run("Existing", func(t *testing.T) {
		fileSync := NewFileSync(778, "NameForRemote")
		result := FirstOrCreateFileSync(fileSync)

		if result == nil {
			t.Fatal("result share must not be nil")
		}

		if result.FileID != fileSync.FileID {
			t.Errorf("FileID should be the same: %d %d", result.FileID, fileSync.FileID)
		}

		if result.ServiceID != fileSync.ServiceID {
			t.Errorf("ServiceID should be the same: %d %d", result.ServiceID, fileSync.ServiceID)
		}
	})
	t.Run("LongName", func(t *testing.T) {
		name := "/" + strings.Repeat("n", RemoteNameMaxLen-5) + ".jpg"
		require.Len(t, name, RemoteNameMaxLen)
		t.Cleanup(func() {
			assert.NoError(t, UnscopedDb().Delete(&FileSync{}, "service_id = ? AND remote_name = ?", 779, name).Error)
		})

		created := FirstOrCreateFileSync(NewFileSync(779, name))
		require.NotNil(t, created)

		found := FirstOrCreateFileSync(NewFileSync(779, name))
		require.NotNil(t, found)
		assert.Equal(t, name, found.RemoteName)

		var count int64
		require.NoError(t, Db().Model(&FileSync{}).Where("service_id = ?", 779).Count(&count).Error)
		assert.EqualValues(t, 1, count)
	})
}

// fileSyncWidthTable is the table that TestFileSync_RemoteNameMigration creates and drops.
const fileSyncWidthTable = "files_sync_width_test"

// fileSyncWidthLegacy has the remote name column of earlier releases.
type fileSyncWidthLegacy struct {
	RemoteName string `gorm:"primaryKey;autoIncrement:false;type:bytes;size:255"`
	ServiceID  uint   `gorm:"primaryKey;autoIncrement:false;size:32"`
	Status     string `gorm:"type:bytes;size:16;"`
}

// TableName returns the name of the test table.
func (fileSyncWidthLegacy) TableName() string { return fileSyncWidthTable }

// fileSyncWidthCurrent has the remote name column of FileSync and FileShare, without their indexes.
type fileSyncWidthCurrent struct {
	RemoteName string `gorm:"primaryKey;autoIncrement:false;type:bytes;size:1024"`
	ServiceID  uint   `gorm:"primaryKey;autoIncrement:false;size:32"`
	Status     string `gorm:"type:bytes;size:16;"`
}

// TableName returns the name of the test table.
func (fileSyncWidthCurrent) TableName() string { return fileSyncWidthTable }

func TestFileSync_RemoteNameMigration(t *testing.T) {
	// The test tables mirror the remote name column, so it must match the entities.
	for _, model := range []any{&FileSync{}, &FileShare{}, &fileSyncWidthCurrent{}} {
		s, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{})
		require.NoError(t, err)
		field := s.LookUpField("RemoteName")
		require.NotNil(t, field)
		assert.Equal(t, RemoteNameMaxLen, field.Size, s.Name)
		assert.True(t, field.PrimaryKey, s.Name)
	}

	db := Db()
	m := db.Migrator()

	require.NoError(t, m.DropTable(fileSyncWidthTable))
	require.NoError(t, m.CreateTable(&fileSyncWidthLegacy{}))
	t.Cleanup(func() { assert.NoError(t, db.Migrator().DropTable(fileSyncWidthTable)) })

	short := fileSyncWidthLegacy{RemoteName: "/legacy/short.jpg", ServiceID: 7, Status: FileSyncNew}
	require.NoError(t, db.Create(&short).Error)

	// Auto migration widens the column in place.
	require.NoError(t, db.AutoMigrate(&fileSyncWidthCurrent{}))

	if !IsDialect(dsn.DialectSQLite) {
		columns, err := m.ColumnTypes(&fileSyncWidthCurrent{})
		require.NoError(t, err)

		var length int64

		for _, c := range columns {
			if strings.EqualFold(c.Name(), "remote_name") {
				length, _ = c.Length()
			}
		}

		assert.EqualValues(t, RemoteNameMaxLen, length)
	}

	// Existing rows are still matched, and the primary key still refuses a duplicate.
	var found fileSyncWidthCurrent
	require.NoError(t, db.Where("service_id = ? AND remote_name = ?", 7, short.RemoteName).First(&found).Error)
	assert.Equal(t, FileSyncNew, found.Status)
	assert.Error(t, db.Create(&fileSyncWidthCurrent{RemoteName: short.RemoteName, ServiceID: 7}).Error)

	long := fileSyncWidthCurrent{RemoteName: "/" + strings.Repeat("l", RemoteNameMaxLen-1), ServiceID: 7, Status: FileSyncNew}
	require.NoError(t, db.Create(&long).Error)
	var foundLong fileSyncWidthCurrent
	require.NoError(t, db.Where("service_id = ? AND remote_name = ?", 7, long.RemoteName).First(&foundLong).Error)

	var count int64
	require.NoError(t, db.Model(&fileSyncWidthCurrent{}).Count(&count).Error)
	assert.EqualValues(t, 2, count)
}

func TestFileSync_Updates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		fileSync := NewFileSync(123, "NameBeforeUpdate")

		assert.Equal(t, "NameBeforeUpdate", fileSync.RemoteName)
		assert.Equal(t, uint(0x7b), fileSync.ServiceID)

		err := fileSync.Updates(FileSync{RemoteName: "NameAfterUpdate", ServiceID: 999})

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "NameAfterUpdate", fileSync.RemoteName)
		assert.Equal(t, uint(0x3e7), fileSync.ServiceID)
	})
}

func TestFileSync_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		fileSync := NewFileSync(123, "NameBeforeUpdate2")
		assert.Equal(t, "NameBeforeUpdate2", fileSync.RemoteName)
		assert.Equal(t, uint(0x7b), fileSync.ServiceID)

		err := fileSync.Update("RemoteName", "new-name")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "new-name", fileSync.RemoteName)
		assert.Equal(t, uint(0x7b), fileSync.ServiceID)
	})
}

func TestFileSync_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		fileSync := NewFileSync(123, "Nameavc")

		initialDate := fileSync.UpdatedAt

		err := fileSync.Save()

		if err != nil {
			t.Fatal(err)
		}
		afterDate := fileSync.UpdatedAt

		assert.True(t, afterDate.After(initialDate))
	})
}

func TestRemoteNameTooLong(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, RemoteNameTooLong(""))
	})
	t.Run("AtLimit", func(t *testing.T) {
		name := strings.Repeat("/"+strings.Repeat("p", 199), 5) + "/" + strings.Repeat("q", 23)
		require.Len(t, name, RemoteNameMaxLen)
		assert.False(t, RemoteNameTooLong(name))
		assert.True(t, RemoteNameTooLong(name+"q"))
	})
	t.Run("LongFileName", func(t *testing.T) {
		// Local file and folder name limits are not part of this check.
		assert.False(t, RemoteNameTooLong("/"+strings.Repeat("s", 300)))
	})
}

func TestFileSyncLongNamePrefix(t *testing.T) {
	prefix := FileSyncLongNamePrefix("/photos")
	assert.True(t, strings.HasPrefix(prefix, ".photoprism/sync/long/"), prefix)
	assert.True(t, strings.HasSuffix(prefix, "/"), prefix)
	assert.Len(t, prefix, len(".photoprism/sync/long/")+9)
	assert.Equal(t, prefix, FileSyncLongNamePrefix("/photos"))
	assert.NotEqual(t, prefix, FileSyncLongNamePrefix("/photos/2026"))
	assert.NotEqual(t, prefix, FileSyncLongNamePrefix("/Photos"))
}

func TestResetFileSyncLongNames(t *testing.T) {
	const serviceID, otherID = 4711, 4712
	t.Cleanup(func() {
		assert.NoError(t, UnscopedDb().Delete(&FileSync{}, "service_id IN (?)", []uint{serviceID, otherID}).Error)
	})

	current, previous := FileSyncLongNamePrefix("/current"), FileSyncLongNamePrefix("/previous")
	kept := []string{"/current/file.jpg", ".photoprism/sync/12", current + "13"}

	for _, name := range append([]string{previous + "14", previous + "15"}, kept...) {
		require.NoError(t, NewFileSync(serviceID, name).Create())
	}

	require.NoError(t, NewFileSync(otherID, previous+"16").Create())

	// names returns the remote names stored for an account.
	names := func(t *testing.T, id uint) (result []string) {
		require.NoError(t, Db().Model(&FileSync{}).Where("service_id = ?", id).Pluck("remote_name", &result).Error)
		return result
	}

	require.NoError(t, ResetFileSyncLongNames(serviceID, "/current"))
	assert.ElementsMatch(t, kept, names(t, serviceID))
	assert.ElementsMatch(t, []string{previous + "16"}, names(t, otherID))

	// A second reset finds nothing to remove.
	require.NoError(t, ResetFileSyncLongNames(serviceID, "/current"))
	assert.ElementsMatch(t, kept, names(t, serviceID))
}

func TestResetFileSyncUnportable(t *testing.T) {
	const serviceID, otherID = 4731, 4732

	var files []*File

	for _, name := range []string{"2019/renamed.jpg", "2019/a:b.jpg", "dir*/x.jpg"} {
		f := &File{FileUID: rnd.GenerateUID(FileUID), FileRoot: RootOriginals, FileName: name, FileHash: rnd.GenerateUID('h')}
		require.NoError(t, UnscopedDb().Create(f).Error)
		files = append(files, f)
	}

	t.Cleanup(func() {
		assert.NoError(t, UnscopedDb().Delete(&FileSync{}, "service_id IN (?)", []uint{serviceID, otherID}).Error)

		for _, f := range files {
			assert.NoError(t, UnscopedDb().Unscoped().Delete(f).Error)
		}
	})

	// placeholder records a file as skipped for its name.
	placeholder := func(t *testing.T, id uint, f *File) string {
		ignored := NewFileSync(id, FileSyncUnportablePrefix+strconv.FormatUint(uint64(f.ID), 10))
		ignored.FileID, ignored.Status = f.ID, FileSyncIgnore
		require.NoError(t, ignored.Create())
		return ignored.RemoteName
	}

	renamed := placeholder(t, serviceID, files[0])
	kept := []string{placeholder(t, serviceID, files[1]), placeholder(t, serviceID, files[2])}

	// Other placeholders of a file with a portable name are not removed.
	excluded := NewFileSync(serviceID, ".photoprism/sync/"+strconv.FormatUint(uint64(files[0].ID), 10))
	excluded.FileID, excluded.Status = files[0].ID, FileSyncIgnore
	require.NoError(t, excluded.Create())
	kept = append(kept, excluded.RemoteName)
	other := placeholder(t, otherID, files[0])

	// names returns the remote names stored for an account.
	names := func(t *testing.T, id uint) (result []string) {
		require.NoError(t, Db().Model(&FileSync{}).Where("service_id = ?", id).Pluck("remote_name", &result).Error)
		return result
	}

	require.NoError(t, ResetFileSyncUnportable(serviceID))
	assert.NotContains(t, names(t, serviceID), renamed)
	assert.ElementsMatch(t, kept, names(t, serviceID))
	assert.ElementsMatch(t, []string{other}, names(t, otherID))

	// A second reset finds nothing to remove.
	require.NoError(t, ResetFileSyncUnportable(serviceID))
	assert.ElementsMatch(t, kept, names(t, serviceID))
}

func TestPortablePath(t *testing.T) {
	t.Run("Portable", func(t *testing.T) {
		assert.True(t, portablePath("2019/03 - Março/#1 & Friends.jpg"))
		assert.True(t, portablePath("a.jpg"))
	})
	t.Run("NotPortable", func(t *testing.T) {
		assert.False(t, portablePath("2019/a:b.jpg"))
		assert.False(t, portablePath("dir*/x.jpg"))
		assert.False(t, portablePath("caf\xe9/x.jpg"))
	})
}

func TestFileSyncUploadName(t *testing.T) {
	assert.Equal(t, ".photoprism/sync/upload/42", FileSyncUploadName(42))
	assert.True(t, strings.HasPrefix(FileSyncUploadName(0), FileSyncUploadPrefix))
}

func TestFailFileSyncUpload(t *testing.T) {
	const serviceID = 4723
	t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(&FileSync{}, "service_id = ?", serviceID).Error) })

	// The first failure creates the placeholder, and later ones count the errors on it.
	require.NoError(t, FailFileSyncUpload(serviceID, 7, "first"))
	require.NoError(t, Db().Model(&FileSync{}).Where("service_id = ? AND remote_name = ?", serviceID,
		FileSyncUploadName(7)).UpdateColumn("status", FileSyncRetry).Error)
	require.NoError(t, FailFileSyncUpload(serviceID, 7, "second"))

	var m FileSync
	require.NoError(t, Db().Where("service_id = ? AND remote_name = ?", serviceID, FileSyncUploadName(7)).First(&m).Error)
	assert.Equal(t, uint(7), m.FileID)
	assert.Equal(t, FileSyncFailed, m.Status)
	assert.Equal(t, "second", m.Error)
	assert.Equal(t, 2, m.Errors)

	count, err := CountFailedFileSyncUploads(serviceID)
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)

	require.NoError(t, ClearFileSyncUpload(serviceID, 7))
	count, err = CountFailedFileSyncUploads(serviceID)
	require.NoError(t, err)
	assert.Zero(t, count)
	require.NoError(t, ClearFileSyncUpload(serviceID, 7))
}

func TestClearFileSyncUpload(t *testing.T) {
	const serviceID = 4724
	t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(&FileSync{}, "service_id = ?", serviceID).Error) })

	require.NoError(t, FailFileSyncUpload(serviceID, 8, "failed"))
	require.NoError(t, FailFileSyncUpload(serviceID, 9, "failed"))
	require.NoError(t, ClearFileSyncUpload(serviceID, 8))

	var names []string
	require.NoError(t, Db().Model(&FileSync{}).Where("service_id = ?", serviceID).Pluck("remote_name", &names).Error)
	assert.Equal(t, []string{FileSyncUploadName(9)}, names)
}

func TestCountFailedFileSyncUploads(t *testing.T) {
	const serviceID = 4725
	t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(&FileSync{}, "service_id = ?", serviceID).Error) })

	count, err := CountFailedFileSyncUploads(serviceID)
	require.NoError(t, err)
	assert.Zero(t, count)

	// Placeholders marked for retry and other records are not counted.
	require.NoError(t, FailFileSyncUpload(serviceID, 10, "failed"))
	require.NoError(t, FailFileSyncUpload(serviceID, 11, "failed"))
	require.NoError(t, NewFileSync(serviceID, "/photos/failed.jpg").Create())
	require.NoError(t, Db().Model(&FileSync{}).Where("service_id = ? AND remote_name = ?", serviceID,
		FileSyncUploadName(11)).UpdateColumn("status", FileSyncRetry).Error)
	require.NoError(t, Db().Model(&FileSync{}).Where("service_id = ? AND remote_name = ?", serviceID,
		"/photos/failed.jpg").UpdateColumn("status", FileSyncFailed).Error)

	count, err = CountFailedFileSyncUploads(serviceID)
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)

	// Files that are missing or deleted are not counted.
	missing, deleted := FileFixtures.Pointer("missing.jpg"), FileFixtures.Pointer("Photo18.jpg")
	require.True(t, missing.FileMissing)
	require.NoError(t, FailFileSyncUpload(serviceID, missing.ID, "failed"))
	require.NoError(t, FailFileSyncUpload(serviceID, deleted.ID, "failed"))
	count, err = CountFailedFileSyncUploads(serviceID)
	require.NoError(t, err)
	assert.EqualValues(t, 2, count)

	require.NoError(t, UnscopedDb().Model(&File{}).Where("id = ?", deleted.ID).UpdateColumn("deleted_at", Now()).Error)
	t.Cleanup(func() {
		assert.NoError(t, UnscopedDb().Model(&File{}).Where("id = ?", deleted.ID).UpdateColumn("deleted_at", nil).Error)
	})
	count, err = CountFailedFileSyncUploads(serviceID)
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)
}

func TestRetryFileSyncUploads(t *testing.T) {
	const serviceID, otherID = 4721, 4722
	t.Cleanup(func() {
		assert.NoError(t, UnscopedDb().Delete(&FileSync{}, "service_id IN (?)", []uint{serviceID, otherID}).Error)
	})

	status := func(t *testing.T, serviceID uint, name string) string {
		var m FileSync
		require.NoError(t, Db().Where("service_id = ? AND remote_name = ?", serviceID, name).First(&m).Error)
		return m.Status
	}

	// Other records keep their status, and placeholders of stalled uploads are removed.
	kept := []string{"/photos/file.jpg", ".photoprism/sync/12", FileSyncLongNamePrefix("/") + "13"}
	for _, name := range append([]string{fileSyncStalled + "14"}, kept...) {
		require.NoError(t, NewFileSync(serviceID, name).Create())
	}
	for id, errs := range map[uint]int{21: 1, 22: 3, 23: 4} {
		for range errs {
			require.NoError(t, FailFileSyncUpload(serviceID, id, "failed"))
		}
	}
	require.NoError(t, FailFileSyncUpload(otherID, 24, "failed"))

	require.NoError(t, RetryFileSyncUploads(serviceID, 3))

	var names []string
	require.NoError(t, Db().Model(&FileSync{}).Where("service_id = ? AND remote_name NOT LIKE ?", serviceID,
		FileSyncUploadPrefix+"%").Pluck("remote_name", &names).Error)
	assert.ElementsMatch(t, kept, names)
	assert.Equal(t, FileSyncNew, status(t, serviceID, kept[0]))
	assert.Equal(t, FileSyncRetry, status(t, serviceID, FileSyncUploadName(21)))
	assert.Equal(t, FileSyncRetry, status(t, serviceID, FileSyncUploadName(22)))
	assert.Equal(t, FileSyncFailed, status(t, serviceID, FileSyncUploadName(23)))
	assert.Equal(t, FileSyncFailed, status(t, otherID, FileSyncUploadName(24)))

	// Without a retry limit, every failed upload is tried again.
	require.NoError(t, RetryFileSyncUploads(serviceID, -1))
	assert.Equal(t, FileSyncRetry, status(t, serviceID, FileSyncUploadName(23)))
}

// TestRetryFileSyncUploads_MissingFile checks that a set-aside missing file is tried again once it is found.
func TestRetryFileSyncUploads_MissingFile(t *testing.T) {
	const serviceID = 4726
	missing := FileFixtures.Pointer("missing.jpg")
	require.True(t, missing.FileMissing)
	t.Cleanup(func() {
		assert.NoError(t, UnscopedDb().Delete(&FileSync{}, "service_id = ?", serviceID).Error)
		assert.NoError(t, UnscopedDb().Model(&File{}).Where("id = ?", missing.ID).UpdateColumn("file_missing", true).Error)
	})

	status := func() string {
		var m FileSync
		require.NoError(t, Db().Where("service_id = ? AND remote_name = ?", serviceID, FileSyncUploadName(missing.ID)).First(&m).Error)
		return m.Status
	}

	require.NoError(t, FailFileSyncUpload(serviceID, missing.ID, "failed"))
	require.NoError(t, RetryFileSyncUploads(serviceID, 3))
	assert.Equal(t, FileSyncFailed, status())

	require.NoError(t, UnscopedDb().Model(&File{}).Where("id = ?", missing.ID).UpdateColumn("file_missing", false).Error)
	require.NoError(t, RetryFileSyncUploads(serviceID, 3))
	assert.Equal(t, FileSyncRetry, status())
}

func TestFileSync_ServiceStatusIndex(t *testing.T) {
	// The queries of an account's queued files and shares filter by service and status. The index is checked in the
	// database of the test driver, so each driver needs its own run, e.g. make test-mariadb and make test-postgres.
	for _, c := range []struct {
		model any
		name  string
	}{{&FileSync{}, "idx_files_sync_service_status"}, {&FileShare{}, "idx_files_share_service_status"}} {
		s, err := schema.Parse(c.model, &sync.Map{}, schema.NamingStrategy{})
		require.NoError(t, err)
		idx := s.LookIndex(c.name)
		require.NotNil(t, idx, c.name)
		require.Len(t, idx.Fields, 2)
		assert.Equal(t, "service_id", idx.Fields[0].DBName)
		assert.Equal(t, "status", idx.Fields[1].DBName)
		assert.True(t, Db().Migrator().HasIndex(c.model, c.name), c.name)
	}
}
