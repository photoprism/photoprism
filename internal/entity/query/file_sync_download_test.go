package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetDownloadFileID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		err := SetDownloadFileID("exampleFileName.jpg", 1000000)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("filename empty", func(t *testing.T) {
		err := SetDownloadFileID("", 1000000)
		if err == nil {
			t.Fatal()
		}
		assert.Equal(t, "sync: cannot update, filename empty", err.Error())
	})
}
