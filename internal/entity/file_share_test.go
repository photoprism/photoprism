package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFileShare_TableName(t *testing.T) {
	fileShare := &FileShare{}
	assert.Equal(t, "files_share", fileShare.TableName())
}

func TestNewFileShare(t *testing.T) {
	r := NewFileShare(123, 123, "test")
	assert.IsType(t, &FileShare{}, r)
	assert.Equal(t, uint(0x7b), r.FileID)
	assert.Equal(t, uint(0x7b), r.AccountID)
	assert.Equal(t, "test", r.RemoteName)
	assert.Equal(t, "new", r.Status)
}

func TestFirstOrCreateFileShare(t *testing.T) {
	t.Run("not yet existing", func(t *testing.T) {
		fileShare := &FileShare{FileID: 123, AccountID: 888, RemoteName: "test888"}
		result := FirstOrCreateFileShare(fileShare)

		if result == nil {
			t.Fatal("result share should not be nil")
		}

		if result.FileID != fileShare.FileID {
			t.Errorf("FileID should be the same: %d %d", result.FileID, fileShare.FileID)
		}

		if result.AccountID != fileShare.AccountID {
			t.Errorf("AccountID should be the same: %d %d", result.AccountID, fileShare.AccountID)
		}
	})

	t.Run("existing", func(t *testing.T) {
		fileShare := NewFileShare(778, 999, "NameForRemote")
		result := FirstOrCreateFileShare(fileShare)

		if result == nil {
			t.Fatal("result share should not be nil")
		}

		if result.FileID != fileShare.FileID {
			t.Errorf("FileID should be the same: %d %d", result.FileID, fileShare.FileID)
		}

		if result.AccountID != fileShare.AccountID {
			t.Errorf("AccountID should be the same: %d %d", result.AccountID, fileShare.AccountID)
		}
	})
}

func TestFileShare_Updates(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fileShare := NewFileShare(123, 123, "NameBeforeUpdate")

		assert.Equal(t, "NameBeforeUpdate", fileShare.RemoteName)
		assert.Equal(t, uint(0x7b), fileShare.AccountID)

		err := fileShare.Updates(FileShare{RemoteName: "NameAfterUpdate", AccountID: 999})

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "NameAfterUpdate", fileShare.RemoteName)
		assert.Equal(t, uint(0x3e7), fileShare.AccountID)
	})
}

func TestFileShare_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fileShare := NewFileShare(123, 123, "NameBeforeUpdate2")
		assert.Equal(t, "NameBeforeUpdate2", fileShare.RemoteName)
		assert.Equal(t, uint(0x7b), fileShare.AccountID)

		err := fileShare.Update("RemoteName", "new-name")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "new-name", fileShare.RemoteName)
		assert.Equal(t, uint(0x7b), fileShare.AccountID)
	})
}

func TestFileShare_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fileShare := NewFileShare(123, 123, "Nameavc")

		initialDate := fileShare.UpdatedAt

		err := fileShare.Save()

		if err != nil {
			t.Fatal(err)
		}
		afterDate := fileShare.UpdatedAt

		assert.True(t, afterDate.After(initialDate))
	})
}
