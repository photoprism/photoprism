package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileShare_TableName(t *testing.T) {
	fileShare := &FileShare{}
	assert.Equal(t, "files_share", fileShare.TableName())
}

func TestNewFileShare(t *testing.T) {
	r := NewFileShare(123, 123, "test")
	assert.IsType(t, &FileShare{}, r)
	assert.Equal(t, uint(0x7b), r.FileID)
	assert.Equal(t, uint(0x7b), r.ServiceID)
	assert.Equal(t, "test", r.RemoteName)
	assert.Equal(t, "new", r.Status)
}

func TestFirstOrCreateFileShare(t *testing.T) {
	t.Run("not yet existing", func(t *testing.T) {
		newFile := &File{ID: 123, PhotoID: 1000041} // Can't add share if the file and service aren't in the database.
		Db().Create(newFile)
		newService := &Service{ID: 888}
		Db().Create(newService)

		fileShare := &FileShare{FileID: 123, ServiceID: 888, RemoteName: "test888"}
		result := FirstOrCreateFileShare(fileShare)

		if result == nil {
			t.Fatal("result share should not be nil")
		}

		if result.FileID != fileShare.FileID {
			t.Errorf("FileID should be the same: %d %d", result.FileID, fileShare.FileID)
		}

		if result.ServiceID != fileShare.ServiceID {
			t.Errorf("ServiceID should be the same: %d %d", result.ServiceID, fileShare.ServiceID)
		}

		UnscopedDb().Delete(fileShare)
		UnscopedDb().Delete(newFile)
		UnscopedDb().Delete(newService)
	})

	t.Run("existing", func(t *testing.T) {
		newFile := &File{ID: 778, PhotoID: 1000041} // Can't add share if the file and service aren't in the database.
		Db().Create(newFile)
		newService := &Service{ID: 999}
		Db().Create(newService)
		fileShare := NewFileShare(778, 999, "NameForRemote")

		result := FirstOrCreateFileShare(fileShare)

		if result == nil {
			t.Fatal("result share should not be nil")
		}

		if result.FileID != fileShare.FileID {
			t.Errorf("FileID should be the same: %d %d", result.FileID, fileShare.FileID)
		}

		if result.ServiceID != fileShare.ServiceID {
			t.Errorf("ServiceID should be the same: %d %d", result.ServiceID, fileShare.ServiceID)
		}
		UnscopedDb().Delete(fileShare)
		UnscopedDb().Delete(newFile)
		UnscopedDb().Delete(newService)
	})
}

func TestFileShare_Updates(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		newFile := &File{ID: 123, PhotoID: 1000041} // Can't add share if the file and service aren't in the database.
		Db().Create(newFile)
		newService123 := &Service{ID: 123}
		Db().Create(newService123)
		newService999 := &Service{ID: 123}
		Db().Create(newService999)

		fileShare := NewFileShare(123, 123, "NameBeforeUpdate")

		assert.Equal(t, "NameBeforeUpdate", fileShare.RemoteName)
		assert.Equal(t, uint(0x7b), fileShare.ServiceID)

		err := fileShare.Updates(FileShare{RemoteName: "NameAfterUpdate", ServiceID: 999})

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "NameAfterUpdate", fileShare.RemoteName)
		assert.Equal(t, uint(0x3e7), fileShare.ServiceID)
		UnscopedDb().Delete(fileShare)
		UnscopedDb().Delete(newFile)
		UnscopedDb().Delete(newService123)
		UnscopedDb().Delete(newService999)
	})
}

func TestFileShare_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		newFile := &File{ID: 123, PhotoID: 1000041} // Can't add share if the file and service aren't in the database.
		Db().Create(newFile)
		newService123 := &Service{ID: 123}
		Db().Create(newService123)
		fileShare := NewFileShare(123, 123, "NameBeforeUpdate2")
		assert.Equal(t, "NameBeforeUpdate2", fileShare.RemoteName)
		assert.Equal(t, uint(0x7b), fileShare.ServiceID)

		err := fileShare.Update("RemoteName", "new-name")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "new-name", fileShare.RemoteName)
		assert.Equal(t, uint(0x7b), fileShare.ServiceID)

		UnscopedDb().Delete(fileShare)
		UnscopedDb().Delete(newFile)
		UnscopedDb().Delete(newService123)
	})
}

func TestFileShare_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		newFile := &File{ID: 123, PhotoID: 1000041} // Can't add share if the file and service aren't in the database.
		Db().Create(newFile)
		newService123 := &Service{ID: 123}
		Db().Create(newService123)

		fileShare := NewFileShare(123, 123, "Nameavc")

		initialDate := fileShare.UpdatedAt

		err := fileShare.Save()

		if err != nil {
			t.Fatal(err)
		}
		afterDate := fileShare.UpdatedAt

		assert.True(t, afterDate.After(initialDate))

		UnscopedDb().Delete(fileShare)
		UnscopedDb().Delete(newFile)
		UnscopedDb().Delete(newService123)
	})
}
