package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFileSync_TableName(t *testing.T) {
	fileSync := &FileSync{}
	assert.Equal(t, "files_sync", fileSync.TableName())
}

func TestNewFileSync(t *testing.T) {
	r := NewFileSync(123, "test")
	assert.IsType(t, &FileSync{}, r)
	assert.Equal(t, uint(0x7b), r.AccountID)
	assert.Equal(t, "test", r.RemoteName)
	assert.Equal(t, "new", r.Status)
}

func TestFirstOrCreateFileSync(t *testing.T) {
	t.Run("not yet existing", func(t *testing.T) {
		fileSync := &FileSync{AccountID: 123, FileID: 888, RemoteName: "test123"}
		result := FirstOrCreateFileSync(fileSync)

		if result == nil {
			t.Fatal("result should not be nil")
		}

		if result.FileID != fileSync.FileID {
			t.Errorf("FileID should be the same: %d %d", result.FileID, fileSync.FileID)
		}

		if result.AccountID != fileSync.AccountID {
			t.Errorf("AccountID should be the same: %d %d", result.AccountID, fileSync.AccountID)
		}
	})

	t.Run("existing", func(t *testing.T) {
		fileSync := NewFileSync(778, "NameForRemote")
		result := FirstOrCreateFileSync(fileSync)

		if result == nil {
			t.Fatal("result share should not be nil")
		}

		if result.FileID != fileSync.FileID {
			t.Errorf("FileID should be the same: %d %d", result.FileID, fileSync.FileID)
		}

		if result.AccountID != fileSync.AccountID {
			t.Errorf("AccountID should be the same: %d %d", result.AccountID, fileSync.AccountID)
		}
	})
}

func TestFileSync_Updates(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fileSync := NewFileSync(123, "NameBeforeUpdate")

		assert.Equal(t, "NameBeforeUpdate", fileSync.RemoteName)
		assert.Equal(t, uint(0x7b), fileSync.AccountID)

		err := fileSync.Updates(FileSync{RemoteName: "NameAfterUpdate", AccountID: 999})

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "NameAfterUpdate", fileSync.RemoteName)
		assert.Equal(t, uint(0x3e7), fileSync.AccountID)
	})
}

func TestFileSync_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fileSync := NewFileSync(123, "NameBeforeUpdate2")
		assert.Equal(t, "NameBeforeUpdate2", fileSync.RemoteName)
		assert.Equal(t, uint(0x7b), fileSync.AccountID)

		err := fileSync.Update("RemoteName", "new-name")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "new-name", fileSync.RemoteName)
		assert.Equal(t, uint(0x7b), fileSync.AccountID)
	})
}

func TestFileSync_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
