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
}
