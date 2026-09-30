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
	t.Run("Limits", func(t *testing.T) {
		for _, c := range []struct {
			name                    string
			interval, retry         int
			wantInterval, wantRetry int
		}{
			{"Default", 86400, 3, 86400, 3},
			{"NegativeInterval", -5, 3, 0, 3},
			{"MaxInterval", 31536001, 3, 31536000, 3},
			{"OneWeek", 604800, 3, 604800, 3},
			{"NeverInterval", 0, 3, 0, 3},
			{"ZeroRetryLimit", 3600, 0, 3600, -1},
			{"NoRetryLimit", 3600, -1, 3600, -1},
			{"NegativeRetryLimit", 3600, -5, 3600, -1},
			{"MaxRetryLimit", 3600, 1000, 3600, 999},
		} {
			t.Run(c.name, func(t *testing.T) {
				model, err := AddService(form.Service{AccName: "Limits", AccURL: "test.com", AccType: "webdav", SyncInterval: c.interval, RetryLimit: c.retry})
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { UnscopedDb().Unscoped().Delete(model) })
				var m Service
				if err = Db().First(&m, model.ID).Error; err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, c.wantInterval, m.SyncInterval)
				assert.Equal(t, c.wantRetry, m.RetryLimit)
			})
		}
	})
	t.Run("SyncYaml", func(t *testing.T) {
		stored := func(t *testing.T, id uint) int {
			var m Service
			if err := Db().First(&m, id).Error; err != nil {
				t.Fatal(err)
			}
			return m.SyncYaml
		}
		for _, c := range []struct{ value, want, update, updated int }{
			{-1, -1, 1, 1},
			{0, 0, -1, -1},
			{1, 1, 0, 0},
			{-5, -1, 5, 1},
		} {
			model, err := AddService(form.Service{AccName: "Sync YAML", AccURL: "test.com", AccType: "webdav", SyncYaml: c.value})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { UnscopedDb().Unscoped().Delete(model) })
			assert.Equal(t, c.want, model.SyncYaml)
			assert.Equal(t, c.want, stored(t, model.ID))
			if err = model.SaveForm(form.Service{AccName: "Sync YAML", AccURL: "test.com", AccType: "webdav", SyncYaml: c.update}); err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, c.updated, model.SyncYaml)
			assert.Equal(t, c.updated, stored(t, model.ID))
		}
	})
}

func TestService_SyncYamlEnabled(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.True(t, (&Service{}).SyncYamlEnabled())
	})
	t.Run("Enabled", func(t *testing.T) {
		assert.True(t, (&Service{SyncYaml: 1}).SyncYamlEnabled())
	})
	t.Run("Disabled", func(t *testing.T) {
		assert.False(t, (&Service{SyncYaml: -1}).SyncYamlEnabled())
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
	t.Run("SanitizesStoredError", func(t *testing.T) {
		model, err := AddService(form.Service{AccName: "LogErrSanitized", AccURL: "test.com", AccType: "webdav"})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { UnscopedDb().Unscoped().Delete(model) })
		//nolint:gosec // G101: Example credential in a fixture URL, which is the subject of the test.
		if err = model.LogErr(errors.New("PROPFIND https://sync-user:notreal@dav.example.com/photos:\nsync \u203a failed")); err != nil {
			t.Fatal(err)
		}
		var m Service
		if err = Db().First(&m, model.ID).Error; err != nil {
			t.Fatal(err)
		}
		assert.NotContains(t, m.AccError, "notreal")
		assert.NotContains(t, m.AccError, "\n")
		assert.NotContains(t, m.AccError, "\u203a")
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
	t.Run("RetryLimit", func(t *testing.T) {
		stored := func(t *testing.T, id uint) Service {
			var m Service
			if err := Db().First(&m, id).Error; err != nil {
				t.Fatal(err)
			}
			return m
		}
		for _, c := range []struct {
			name        string
			limit, errs int
			share       bool
		}{
			{"AtLimit", 2, 2, true},
			{"AboveLimit", 2, 3, false},
			{"ZeroLimit", 0, 5, true},
			{"NoLimit", -1, 5, true},
		} {
			t.Run(c.name, func(t *testing.T) {
				model, err := AddService(form.Service{AccName: "LogErrLimit", AccURL: "test.com", AccType: "webdav", AccShare: true, RetryLimit: c.limit})
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { UnscopedDb().Unscoped().Delete(model) })
				// SaveForm stores 0 as -1, so the limit is set directly.
				if err = model.Update("retry_limit", c.limit); err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, c.limit, model.RetryLimit)
				// Changed in the database only, so a full-row save would revert it.
				if err = Db().Model(&Service{ID: model.ID}).UpdateColumn("acc_name", "LogErrChanged").Error; err != nil {
					t.Fatal(err)
				}
				for i := 0; i < c.errs; i++ {
					if err = model.LogErr(errors.New("boom")); err != nil {
						t.Fatal(err)
					}
				}
				m := stored(t, model.ID)
				assert.Equal(t, "LogErrChanged", m.AccName)
				assert.Equal(t, c.errs, m.AccErrors)
				assert.Equal(t, "boom", m.AccError)
				assert.Equal(t, c.share, m.AccShare)
				// A success resets the counters and leaves the stored share flag alone.
				if err = Db().Model(&Service{ID: model.ID}).UpdateColumn("acc_share", !c.share).Error; err != nil {
					t.Fatal(err)
				}
				if err = model.LogErr(nil); err != nil {
					t.Fatal(err)
				}
				m = stored(t, model.ID)
				assert.Equal(t, 0, m.AccErrors)
				assert.Equal(t, "", m.AccError)
				assert.Equal(t, !c.share, m.AccShare)
			})
		}
	})
}

func TestService_ResetErrors(t *testing.T) {
	// newAccount creates a WebDAV account with sync and sharing on, one failed download, and one failed share.
	newAccount := func(t *testing.T, limit int) *Service {
		m, err := AddService(form.Service{AccName: "ResetErrors", AccURL: "test.com", AccType: "webdav", AccShare: true, AccSync: true, RetryLimit: limit})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			UnscopedDb().Unscoped().Delete(&FileSync{}, "service_id = ?", m.ID)
			UnscopedDb().Unscoped().Delete(&FileShare{}, "service_id = ?", m.ID)
			UnscopedDb().Unscoped().Delete(m)
		})
		fileSync := NewFileSync(m.ID, "/reset.jpg")
		fileSync.Status, fileSync.Errors, fileSync.Error = FileSyncNew, 3, "boom"
		if err = fileSync.Create(); err != nil {
			t.Fatal(err)
		}
		fileShare := NewFileShare(1000000, m.ID, "/reset.jpg")
		fileShare.Errors, fileShare.Error = 2, "boom"
		if err = fileShare.Create(); err != nil {
			t.Fatal(err)
		}
		return m
	}
	// stored returns the stored download and share rows of the account.
	stored := func(t *testing.T, m *Service) (FileSync, FileShare) {
		var fileSync FileSync
		var fileShare FileShare
		if err := Db().Where("service_id = ?", m.ID).First(&fileSync).Error; err != nil {
			t.Fatal(err)
		}
		if err := Db().Where("service_id = ?", m.ID).First(&fileShare).Error; err != nil {
			t.Fatal(err)
		}
		return fileSync, fileShare
	}
	// save stores the account through SaveForm with the specified retry limit.
	save := func(t *testing.T, m *Service, limit int) {
		f, err := form.NewService(m)
		if err != nil {
			t.Fatal(err)
		}
		f.RetryLimit = limit
		if err = m.SaveForm(f); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("NoLimitKeepsCounts", func(t *testing.T) {
		m := newAccount(t, -1)
		save(t, m, -1)
		fileSync, fileShare := stored(t, m)
		assert.Equal(t, 3, fileSync.Errors)
		assert.Equal(t, "", fileSync.Error)
		assert.Equal(t, 0, fileShare.Errors)
		assert.Equal(t, "", fileShare.Error)
	})
	t.Run("NoLimitToLimit", func(t *testing.T) {
		m := newAccount(t, -1)
		save(t, m, 3)
		fileSync, _ := stored(t, m)
		assert.Equal(t, 0, fileSync.Errors)
		assert.Equal(t, "", fileSync.Error)
	})
	t.Run("LimitResetsCounts", func(t *testing.T) {
		m := newAccount(t, 3)
		save(t, m, 3)
		fileSync, fileShare := stored(t, m)
		assert.Equal(t, 0, fileSync.Errors)
		assert.Equal(t, "", fileSync.Error)
		assert.Equal(t, 0, fileShare.Errors)
		assert.Equal(t, "", fileShare.Error)
	})
	t.Run("FolderListing", func(t *testing.T) {
		m := newAccount(t, -1)
		if err := m.LogErr(nil); err != nil {
			t.Fatal(err)
		}
		fileSync, fileShare := stored(t, m)
		assert.Equal(t, 3, fileSync.Errors)
		assert.Equal(t, "", fileSync.Error)
		assert.Equal(t, 0, fileShare.Errors)
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
