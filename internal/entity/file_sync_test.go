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

func TestFileSync_FirstOrCreate(t *testing.T) {
	fileSync := &FileSync{AccountID: 123}
	r := fileSync.FirstOrCreate()
	assert.Equal(t, uint(0x7b), r.AccountID)
}
