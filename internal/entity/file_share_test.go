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

func TestFileShare_FirstOrCreate(t *testing.T) {
	fileShare := &FileShare{FileID: 123}
	r := fileShare.FirstOrCreate()
	assert.Equal(t, uint(0x7b), r.FileID)
}
