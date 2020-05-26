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
}
