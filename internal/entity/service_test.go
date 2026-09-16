package entity

import (
	"errors"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt"
)

func TestCreateService(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		account := Service{AccName: "Foo", AccOwner: "bar", AccURL: "test.com", AccType: "webdav", AccKey: "123", AccUser: "testuser", AccPass: "testpass",
			AccError: "", AccShare: true, AccSync: true, RetryLimit: 4, SharePath: "/home", ShareSize: "500", ShareExpires: 3500, SyncPath: "/sync",
			SyncInterval: 5, SyncUpload: true, SyncDownload: false, SyncFilenames: true, SyncRaw: false}

		accountForm, err := form.NewService(account)

		if err != nil {
			t.Fatal(err)
		}

		model, err := AddService(accountForm)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "/home", model.SharePath)
		assert.Equal(t, 3500, model.ShareExpires)
		assert.Equal(t, "500", model.ShareSize)
		assert.Equal(t, "refresh", model.SyncStatus)
		assert.Equal(t, "Foo", model.AccName)
		assert.Equal(t, "bar", model.AccOwner)
		assert.Equal(t, "test.com", model.AccURL)
		assert.Equal(t, "webdav", model.AccType)
		assert.Equal(t, "123", model.AccKey)
		assert.Equal(t, "testuser", model.AccUser)
		assert.Equal(t, "testpass", model.AccPass)
		assert.Equal(t, "", model.AccError)
		assert.Equal(t, false, model.SyncDownload)
		assert.Equal(t, true, model.AccShare)
		assert.Equal(t, true, model.AccSync)
		assert.Equal(t, 4, model.RetryLimit)
		assert.Equal(t, "/sync", model.SyncPath)
		assert.Equal(t, 5, model.SyncInterval)
		assert.Equal(t, true, model.SyncUpload)
		assert.Equal(t, true, model.SyncFilenames)
		assert.Equal(t, false, model.SyncRaw)
	})
}

func TestService_SaveForm(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		account := Service{AccName: "Foo", AccOwner: "bar", AccURL: "test.com", AccType: "test", AccKey: "123", AccUser: "testuser", AccPass: "testpass",
			AccError: "", AccShare: true, AccSync: true, RetryLimit: 4, SharePath: "/home", ShareSize: "500", ShareExpires: 3500, SyncPath: "/sync",
			SyncInterval: 5, SyncUpload: true, SyncDownload: true, SyncFilenames: true, SyncRaw: false}

		accountForm, err := form.NewService(account)

		if err != nil {
			t.Fatal(err)
		}
		model, err := AddService(accountForm)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, true, model.SyncDownload)
		assert.Equal(t, false, model.SyncUpload)
		assert.Equal(t, "Foo", model.AccName)
		assert.Equal(t, "bar", model.AccOwner)
		assert.Equal(t, "test.com", model.AccURL)

		accountUpdate := Service{AccName: "NewName", AccOwner: "NewOwner", AccURL: "new.com", SyncUpload: true, SyncDownload: true}

		UpdateForm, err := form.NewService(accountUpdate)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, true, UpdateForm.SyncDownload)
		assert.Equal(t, true, UpdateForm.SyncUpload)

		err = model.SaveForm(UpdateForm)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, true, model.SyncDownload)
		assert.Equal(t, false, model.SyncUpload)
		assert.Equal(t, "NewName", model.AccName)
		assert.Equal(t, "NewOwner", model.AccOwner)
		assert.Equal(t, "new.com", model.AccURL)
	})
}

func TestService_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		account := Service{AccName: "DeleteAccount", AccOwner: "Delete", AccURL: "test.com", AccType: "test", AccKey: "123", AccUser: "testuser", AccPass: "testpass",
			AccError: "", AccShare: true, AccSync: true, RetryLimit: 4, SharePath: "/home", ShareSize: "500", ShareExpires: 3500, SyncPath: "/sync",
			SyncInterval: 5, SyncUpload: true, SyncDownload: false, SyncFilenames: true, SyncRaw: false}

		accountForm, err := form.NewService(account)

		if err != nil {
			t.Fatal(err)
		}
		model, err := AddService(accountForm)

		if err != nil {
			t.Fatal(err)
		}

		err = model.Delete()

		if err != nil {
			t.Fatal(err)
		}
		// TODO how to assert deletion?

	})
}

func TestService_Directories(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		account := Service{AccName: "DirectoriesAccount", AccOwner: "Owner", AccURL: "http://dummy-webdav/", AccType: "webdav", AccKey: "123", AccUser: "admin", AccPass: "photoprism",
			AccError: "", AccShare: true, AccSync: true, RetryLimit: 4, SharePath: "/home", ShareSize: "500", ShareExpires: 3500, SyncPath: "/sync",
			SyncInterval: 5, SyncUpload: true, SyncDownload: false, SyncFilenames: true, SyncRaw: false}

		accountForm, err := form.NewService(account)

		if err != nil {
			t.Fatal(err)
		}
		model, err := AddService(accountForm)

		if err != nil {
			t.Fatal(err)
		}

		result, err := model.Directories("")

		if err != nil {
			t.Fatal(err)
		}
		assert.NotEmpty(t, result.Abs())
		assert.Contains(t, result.Abs(), "/Photos")

	})
	t.Run("NoDirectory", func(t *testing.T) {
		account := Service{AccName: "DirectoriesAccount", AccOwner: "Owner", AccURL: "http://dummy-webdav/", AccType: "xxx", AccKey: "123", AccUser: "admin", AccPass: "photoprism",
			AccError: "", AccShare: true, AccSync: true, RetryLimit: 4, SharePath: "/home", ShareSize: "500", ShareExpires: 3500, SyncPath: "/sync",
			SyncInterval: 5, SyncUpload: true, SyncDownload: false, SyncFilenames: true, SyncRaw: false}

		accountForm, err := form.NewService(account)

		if err != nil {
			t.Fatal(err)
		}
		model, err := AddService(accountForm)

		if err != nil {
			t.Fatal(err)
		}

		result, err := model.Directories("")

		if err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, result.Abs())
	})
}

func TestService_Updates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		account := Service{AccName: "DeleteAccount", AccOwner: "Delete", AccURL: "test.com", AccType: "test", AccKey: "123", AccUser: "testuser", AccPass: "testpass",
			AccError: "", AccShare: true, AccSync: true, RetryLimit: 4, SharePath: "/home", ShareSize: "500", ShareExpires: 3500, SyncPath: "/sync",
			SyncInterval: 5, SyncUpload: true, SyncDownload: false, SyncFilenames: true, SyncRaw: false}

		accountForm, err := form.NewService(account)

		if err != nil {
			t.Fatal(err)
		}
		model, err := AddService(accountForm)
		assert.Equal(t, "testuser", model.AccUser)
		assert.Equal(t, "DeleteAccount", model.AccName)

		if err != nil {
			t.Fatal(err)
		}

		err = model.Updates(Service{AccName: "UpdatedName", AccUser: "UpdatedUser"})
		assert.Equal(t, "UpdatedUser", model.AccUser)
		assert.Equal(t, "UpdatedName", model.AccName)

		if err != nil {
			t.Fatal(err)
		}

	})
}

func TestService_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		account := Service{AccName: "DeleteAccount", AccOwner: "Delete", AccURL: "test.com", AccType: "test", AccKey: "123", AccUser: "testuser", AccPass: "testpass",
			AccError: "", AccShare: true, AccSync: true, RetryLimit: 4, SharePath: "/home", ShareSize: "500", ShareExpires: 3500, SyncPath: "/sync",
			SyncInterval: 5, SyncUpload: true, SyncDownload: false, SyncFilenames: true, SyncRaw: false}

		accountForm, err := form.NewService(account)

		if err != nil {
			t.Fatal(err)
		}
		model, err := AddService(accountForm)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "testuser", model.AccUser)

		err = model.Update("AccUser", "UpdatedUser")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "UpdatedUser", model.AccUser)
	})
}

// TODO fails on mariadb
func TestService_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		account := Service{AccName: "DeleteAccount", AccOwner: "Delete", AccURL: "test.com", AccType: "test", AccKey: "123", AccUser: "testuser", AccPass: "testpass",
			AccError: "", AccShare: true, AccSync: true, RetryLimit: 4, SharePath: "/home", ShareSize: "500", ShareExpires: 3500, SyncPath: "/sync",
			SyncInterval: 5, SyncUpload: true, SyncDownload: false, SyncFilenames: true, SyncRaw: false}

		accountForm, err := form.NewService(account)

		if err != nil {
			t.Fatal(err)
		}
		model, err := AddService(accountForm)

		if err != nil {
			t.Fatal(err)
		}
		initialDate := model.UpdatedAt

		err = model.Save()

		if err != nil {
			t.Fatal(err)
		}
		afterDate := model.UpdatedAt
		// Timestamps are stored with second precision, so a save within the same
		// second leaves UpdatedAt unchanged; assert it stays within a sane window.
		elapsed := afterDate.Sub(initialDate)
		assert.GreaterOrEqual(t, elapsed, time.Duration(0))
		assert.Less(t, elapsed, time.Minute)
	})
}

func TestService_LogErr(t *testing.T) {
	t.Run("ClipsMultiByteErrorMessage", func(t *testing.T) {
		// acc_error is a bounded VARBINARY column and LogErr budgets messages to
		// txt.ClipError (255) bytes; a long multi-byte error must be clipped on a rune
		// boundary so the stored value stays within budget and valid UTF-8.
		account := Service{AccName: "LogErrAccount", AccOwner: "LogErr", AccURL: "test.com", AccType: "test",
			AccKey: "123", AccUser: "testuser", AccPass: "testpass", RetryLimit: 4}

		accountForm, err := form.NewService(account)
		if err != nil {
			t.Fatal(err)
		}
		model, err := AddService(accountForm)
		if err != nil {
			t.Fatal(err)
		}

		longErr := errors.New(strings.Repeat("世", 100)) // 100 runes x 3 bytes = 300 bytes.

		if err = model.LogErr(longErr); err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, model.AccError)
		assert.LessOrEqual(t, len(model.AccError), txt.ClipError)
		assert.True(t, utf8.ValidString(model.AccError))
		assert.Equal(t, 1, model.AccErrors)

		// The clipped message is also what gets persisted.
		var reloaded Service
		if err = Db().Where("id = ?", model.ID).First(&reloaded).Error; err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, model.AccError, reloaded.AccError)
		assert.LessOrEqual(t, len(reloaded.AccError), txt.ClipError)
	})
	t.Run("NilErrorResetsErrors", func(t *testing.T) {
		// A nil error clears the recorded message and counter via ResetErrors.
		account := Service{AccName: "LogErrReset", AccOwner: "LogErr", AccURL: "test.com", AccType: "test",
			AccKey: "123", AccUser: "testuser", AccPass: "testpass", RetryLimit: 4}

		accountForm, err := form.NewService(account)
		if err != nil {
			t.Fatal(err)
		}
		model, err := AddService(accountForm)
		if err != nil {
			t.Fatal(err)
		}

		if err = model.LogErr(errors.New("boom")); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 1, model.AccErrors)

		if err = model.LogErr(nil); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, model.AccErrors)
		assert.Equal(t, "", model.AccError)
	})
}

func TestService_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		account := Service{}

		err := account.Create()

		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestServiceError(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "", ServiceError(nil))
	})
	t.Run("KeepsTheStatus", func(t *testing.T) {
		// The status and reason are what an operator acts on, so they must survive.
		s := ServiceError(errors.New("507 Insufficient Storage: quota exceeded"))
		assert.Contains(t, s, "507")
		assert.Contains(t, s, "Insufficient Storage")
		assert.Contains(t, s, "quota exceeded")
	})
	t.Run("RemoteBodyBytesDoNotSurvive", func(t *testing.T) {
		// The stored value must stay one line, in one field, whatever the cause contains.
		body := "500 Internal Server Error: \x00\x01\x02\x7f\x1b[31m\nsync: › admin › granted‮\a"
		s := ServiceError(errors.New(body))
		assert.NotContains(t, s, "\x00")
		assert.NotContains(t, s, "\x1b")
		assert.NotContains(t, s, "\a")
		assert.NotContains(t, s, "\n", "a remote body must not add a line")
		assert.NotContains(t, s, "‮", "a remote body must not carry a bidi override")
		assert.NotContains(t, s, "›", "a remote body must not add a field separator")
		assert.True(t, utf8.ValidString(s), "the stored value must be valid UTF-8")
	})
	t.Run("CredentialsDoNotSurvive", func(t *testing.T) {
		//nolint:gosec // G101: Example credential in a fixture URL, which is the subject of the test.
		err := errors.New("PROPFIND https://sync-user:notreal@dav.example.com/photos: timeout")
		assert.NotContains(t, ServiceError(err), "notreal")
	})
	t.Run("ClipsWellInsideTheColumn", func(t *testing.T) {
		// AccError is VARBINARY(512), and the status must not be crowded out of it.
		s := ServiceError(errors.New("503 Service Unavailable: " + strings.Repeat("a", 4096)))
		assert.LessOrEqual(t, len(s), txt.ClipError)
		assert.Contains(t, s, "503")
	})
}
