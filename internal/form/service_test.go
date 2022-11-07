package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		service := Service{AccName: "Foo", AccOwner: "bar", AccURL: "test.com", AccType: "test", AccKey: "123", AccUser: "testuser", AccPass: "testpass",
			AccError: "", AccShare: true, AccSync: true, RetryLimit: 4, SharePath: "/home", ShareSize: "500", ShareExpires: 3500, SyncPath: "/sync",
			SyncInterval: 5, SyncUpload: true, SyncDownload: false, SyncFilenames: true, SyncRaw: false}

		r, err := NewService(service)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Foo", r.AccName)
		assert.Equal(t, "bar", r.AccOwner)
		assert.Equal(t, "test.com", r.AccURL)
		assert.Equal(t, "test", r.AccType)
		assert.Equal(t, "123", r.AccKey)
		assert.Equal(t, "testuser", r.AccUser)
		assert.Equal(t, "testpass", r.AccPass)
		assert.Equal(t, "", r.AccError)
		assert.Equal(t, false, r.SyncDownload)
		assert.Equal(t, true, r.AccShare)
		assert.Equal(t, true, r.AccSync)
		assert.Equal(t, 4, r.RetryLimit)
		assert.Equal(t, "/home", r.SharePath)
		assert.Equal(t, "500", r.ShareSize)
		assert.Equal(t, 3500, r.ShareExpires)
		assert.Equal(t, "/sync", r.SyncPath)
		assert.Equal(t, 5, r.SyncInterval)
		assert.Equal(t, true, r.SyncUpload)
		assert.Equal(t, true, r.SyncFilenames)
		assert.Equal(t, false, r.SyncRaw)
	})
}

func TestService_Discovery(t *testing.T) {
	t.Run("error = nil", func(t *testing.T) {
		service := Service{AccName: "Foo", AccOwner: "bar", AccURL: "photoprism.app", AccType: "test", SyncDownload: false, AccShare: true}

		err := service.Discovery()
		assert.Equal(t, nil, err)
	})
	t.Run("error != nil", func(t *testing.T) {
		service := Service{AccName: "XXX", AccOwner: "bar"}

		err := service.Discovery()
		assert.Equal(t, "service URL is empty", err.Error())
	})
}
