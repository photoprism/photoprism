package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileSync_TableName(t *testing.T) {
	fileSync := &FileSync{}
	assert.Equal(t, "files_sync", fileSync.TableName())
}

func TestNewFileSync(t *testing.T) {
	newService := &Service{ID: 123}
	Db().Create(newService)

	r := NewFileSync(123, "test")
	assert.IsType(t, &FileSync{}, r)
	assert.Equal(t, uint(0x7b), r.ServiceID)
	assert.Equal(t, "test", r.RemoteName)
	assert.Equal(t, "new", r.Status)
	UnscopedDb().Delete(newService)
}

func TestFirstOrCreateFileSync(t *testing.T) {
	t.Run("not yet existing", func(t *testing.T) {
		newFile := &File{ID: 888, PhotoID: 1000041} // Can't add share if the file and service aren't in the database.
		Db().Create(newFile)
		newService := &Service{ID: 123}
		Db().Create(newService)

		fileSync := &FileSync{ServiceID: 123, FileID: &newFile.ID, RemoteName: "test123"}
		result := FirstOrCreateFileSync(fileSync)

		if result == nil {
			t.Fatal("result should not be nil")
		}

		if result.FileID != fileSync.FileID {
			t.Errorf("FileID should be the same: %d %d", result.FileID, fileSync.FileID)
		}

		if result.ServiceID != fileSync.ServiceID {
			t.Errorf("ServiceID should be the same: %d %d", result.ServiceID, fileSync.ServiceID)
		}
		UnscopedDb().Delete(fileSync)
		UnscopedDb().Delete(newFile)
		UnscopedDb().Delete(newService)
	})

	t.Run("existing", func(t *testing.T) {
		fileSync := NewFileSync(1000000, "NameForRemote")                 // Point at a real service
		fileSync.FileID = &FileFixtures.Pointer("exampleFileName.jpg").ID // Add a pointer to a real file
		result := FirstOrCreateFileSync(fileSync)
		if result == nil {
			t.Fatal("result share should not be nil")
		}

		if result.FileID != fileSync.FileID {
			t.Errorf("FileID should be the same: %d %d", result.FileID, fileSync.FileID)
		}

		if result.ServiceID != fileSync.ServiceID {
			t.Errorf("ServiceID should be the same: %d %d", result.ServiceID, fileSync.ServiceID)
		}
	})
}

func TestFileSync_Updates(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		newService := &Service{ID: 123}
		Db().Create(newService)
		fileSync := NewFileSync(123, "NameBeforeUpdate")

		assert.Equal(t, "NameBeforeUpdate", fileSync.RemoteName)
		assert.Equal(t, uint(0x7b), fileSync.ServiceID)

		err := fileSync.Updates(FileSync{RemoteName: "NameAfterUpdate", ServiceID: 999})

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "NameAfterUpdate", fileSync.RemoteName)
		assert.Equal(t, uint(0x3e7), fileSync.ServiceID)
		UnscopedDb().Delete(newService)
	})
}

func TestFileSync_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		newService := &Service{ID: 123}
		Db().Create(newService)

		fileSync := NewFileSync(123, "NameBeforeUpdate2")
		assert.Equal(t, "NameBeforeUpdate2", fileSync.RemoteName)
		assert.Equal(t, uint(0x7b), fileSync.ServiceID)

		err := fileSync.Update("RemoteName", "new-name")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "new-name", fileSync.RemoteName)
		assert.Equal(t, uint(0x7b), fileSync.ServiceID)
		UnscopedDb().Delete(newService)
	})
}

func TestFileSync_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		newService := &Service{ID: 123}
		Db().Create(newService)
		newFile := &File{ID: 888, PhotoID: 1000041, FileName: "Nameavc"} // Can't add share if the file and service aren't in the database.
		Db().Create(newFile)

		fileSync := NewFileSync(123, "Nameavc")

		initialDate := fileSync.UpdatedAt

		err := fileSync.Save()

		if err != nil {
			t.Fatal(err)
		}
		afterDate := fileSync.UpdatedAt

		assert.True(t, afterDate.After(initialDate))
		UnscopedDb().Delete(fileSync)
		UnscopedDb().Delete(newFile)
		UnscopedDb().Delete(newService)
	})
}
