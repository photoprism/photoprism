package entity

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/webdav"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/fs"
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

func TestService_SaveFormAccError(t *testing.T) {
	t.Run("NewAccount", func(t *testing.T) {
		model, err := AddService(form.Service{AccName: "Error New", AccURL: "test.com", AccType: "webdav", AccError: "client value"})
		require.NoError(t, err)
		t.Cleanup(func() { UnscopedDb().Unscoped().Delete(model) })
		assert.Empty(t, storedService(t, model.ID).AccError)
	})
	t.Run("Kept", func(t *testing.T) {
		model, err := AddService(form.Service{AccName: "Error Kept", AccURL: "test.com", AccType: "webdav"})
		require.NoError(t, err)
		t.Cleanup(func() { UnscopedDb().Unscoped().Delete(model) })
		require.NoError(t, model.Updates(Values{"acc_error": "worker error"}))

		stored := storedService(t, model.ID)
		frm, err := form.NewService(stored)
		require.NoError(t, err)
		frm.AccError = "client value"
		require.NoError(t, stored.SaveForm(frm))
		assert.Equal(t, "worker error", storedService(t, model.ID).AccError)
	})
	t.Run("ResetOnSave", func(t *testing.T) {
		// Saving an account with sync on still resets its errors, including the message.
		model, err := AddService(form.Service{AccName: "Error Reset", AccURL: "test.com", AccType: "webdav", AccSync: true})
		require.NoError(t, err)
		t.Cleanup(func() { UnscopedDb().Unscoped().Delete(model) })
		require.NoError(t, model.Updates(Values{"acc_error": "worker error", "acc_errors": 2}))

		stored := storedService(t, model.ID)
		frm, err := form.NewService(stored)
		require.NoError(t, err)
		require.NoError(t, stored.SaveForm(frm))
		assert.Empty(t, storedService(t, model.ID).AccError)
		assert.Zero(t, storedService(t, model.ID).AccErrors)
	})
	t.Run("Cleared", func(t *testing.T) {
		model, err := AddService(form.Service{AccName: "Error Cleared", AccURL: "test.com", AccType: "webdav"})
		require.NoError(t, err)
		t.Cleanup(func() { UnscopedDb().Unscoped().Delete(model) })
		require.NoError(t, model.Updates(Values{"acc_error": "worker error"}))

		stored := storedService(t, model.ID)
		frm, err := form.NewService(stored)
		require.NoError(t, err)
		frm.AccError = ""
		require.NoError(t, stored.SaveForm(frm))
		assert.Empty(t, storedService(t, model.ID).AccError)
	})
}

// storedService returns a service as currently stored in the database.
func storedService(t *testing.T, id uint) *Service {
	t.Helper()

	m := &Service{}
	require.NoError(t, Db().First(m, id).Error)

	return m
}

func TestService_SaveFormRefresh(t *testing.T) {
	// A save that changes the folder or a direction, or turns on raw and video files, refreshes the account in any
	// stage; other changes keep it.
	for _, c := range []struct {
		name    string
		status  string
		update  func(f *form.Service)
		refresh bool
	}{
		{"SyncedInterval", SyncStatusSynced, func(f *form.Service) { f.SyncInterval = 60 }, true},
		{"DownloadPath", SyncStatusDownload, func(f *form.Service) { f.SyncPath = "/new" }, true},
		{"DownloadOff", SyncStatusDownload, func(f *form.Service) { f.SyncDownload = false }, true},
		{"UploadOn", SyncStatusDownload, func(f *form.Service) { f.SyncDownload, f.SyncUpload = false, true }, true},
		{"DownloadInterval", SyncStatusDownload, func(f *form.Service) { f.SyncInterval = 60 }, false},
		{"UploadPassword", SyncStatusUpload, func(f *form.Service) { f.AccPass = "changed" }, false},
		{"RawOn", SyncStatusDownload, func(f *form.Service) { f.SyncRaw = true }, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			model, err := AddService(form.Service{AccName: "Refresh " + c.name, AccURL: "test.com", AccType: "webdav",
				AccSync: true, SyncDownload: true, SyncPath: "/old"})
			require.NoError(t, err)
			t.Cleanup(func() { UnscopedDb().Unscoped().Delete(model) })
			require.NoError(t, model.Updates(Values{"sync_status": c.status}))
			stored := storedService(t, model.ID)
			frm, err := form.NewService(stored)
			require.NoError(t, err)
			c.update(&frm)
			require.NoError(t, stored.SaveForm(frm))

			if c.refresh {
				assert.Equal(t, SyncStatusRefresh, storedService(t, model.ID).SyncStatus)
			} else {
				assert.Equal(t, c.status, storedService(t, model.ID).SyncStatus)
			}
		})
	}
}

func TestService_SaveFormRefreshRawKept(t *testing.T) {
	// An account with raw and video files already on keeps its stage when other settings change.
	model, err := AddService(form.Service{AccName: "Refresh Raw Kept", AccURL: "test.com", AccType: "webdav",
		AccSync: true, SyncDownload: true, SyncRaw: true, SyncPath: "/old"})
	require.NoError(t, err)
	t.Cleanup(func() { UnscopedDb().Unscoped().Delete(model) })
	require.NoError(t, model.Updates(Values{"sync_status": SyncStatusDownload}))

	stored := storedService(t, model.ID)
	require.True(t, stored.SyncRaw)
	frm, err := form.NewService(stored)
	require.NoError(t, err)
	frm.SyncInterval = 60
	require.NoError(t, stored.SaveForm(frm))
	assert.Equal(t, SyncStatusDownload, storedService(t, model.ID).SyncStatus)
}

func TestService_SaveFormRefreshSyncOff(t *testing.T) {
	model, err := AddService(form.Service{AccName: "Refresh Sync Off", AccURL: "test.com", AccType: "webdav",
		AccSync: true, SyncPath: "/old"})
	require.NoError(t, err)
	t.Cleanup(func() { UnscopedDb().Unscoped().Delete(model) })
	require.NoError(t, model.Updates(Values{"sync_status": SyncStatusDownload}))

	// A folder changed while sync is off refreshes the account once sync is on again.
	stored := storedService(t, model.ID)
	frm, err := form.NewService(stored)
	require.NoError(t, err)
	frm.AccSync, frm.SyncPath = false, "/new"
	require.NoError(t, stored.SaveForm(frm))
	assert.Equal(t, SyncStatusRefresh, storedService(t, model.ID).SyncStatus)

	// Turning on uploads alone refreshes as well.
	require.NoError(t, model.Updates(Values{"sync_status": SyncStatusDownload}))
	stored = storedService(t, model.ID)
	frm, err = form.NewService(stored)
	require.NoError(t, err)
	frm.AccSync, frm.SyncUpload = true, true
	require.NoError(t, stored.SaveForm(frm))
	assert.Equal(t, SyncStatusRefresh, storedService(t, model.ID).SyncStatus)

	// Turning raw and video files on while sync is off refreshes as well, and turning them off does not.
	require.NoError(t, model.Updates(Values{"sync_status": SyncStatusDownload, "acc_sync": false}))
	stored = storedService(t, model.ID)
	frm, err = form.NewService(stored)
	require.NoError(t, err)
	frm.SyncRaw = true
	require.NoError(t, stored.SaveForm(frm))
	assert.Equal(t, SyncStatusRefresh, storedService(t, model.ID).SyncStatus)

	require.NoError(t, model.Updates(Values{"sync_status": SyncStatusDownload}))
	stored = storedService(t, model.ID)
	frm, err = form.NewService(stored)
	require.NoError(t, err)
	frm.SyncRaw = false
	require.NoError(t, stored.SaveForm(frm))
	assert.Equal(t, SyncStatusDownload, storedService(t, model.ID).SyncStatus)
}

func TestService_SaveFormFolderTooLong(t *testing.T) {
	limit := strings.Repeat("/"+strings.Repeat("f", 200), 5) + "/" + strings.Repeat("g", 18)
	require.Len(t, limit, RemoteNameMaxLen)
	longName := "/photos/" + strings.Repeat("n", fs.NameMaxLen+1)

	for _, c := range []struct {
		name string
		form form.Service
		err  bool
	}{
		{"SyncPathAtLimit", form.Service{SyncPath: limit, SyncDownload: true}, true},
		{"SyncPathBelowLimit", form.Service{SyncPath: limit[:len(limit)-1], SyncDownload: true}, false},
		{"SharePathAtLimit", form.Service{SharePath: limit}, true},
		{"SharePathBelowLimit", form.Service{SharePath: limit[:len(limit)-1]}, false},
		{"DownloadFolderNameTooLong", form.Service{SyncPath: longName, SyncDownload: true}, true},
		// The remote decides which folder names it accepts for uploads.
		{"UploadFolderNameTooLong", form.Service{SyncPath: longName, SyncUpload: true}, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			c.form.AccName, c.form.AccURL, c.form.AccType = "Folder "+c.name, "test.com", "webdav"
			model, err := AddService(c.form)
			t.Cleanup(func() {
				if model != nil && model.ID > 0 {
					UnscopedDb().Unscoped().Delete(model)
				}
			})

			if c.err {
				assert.ErrorIs(t, err, ErrServiceFolderTooLong)
				var count int64
				require.NoError(t, Db().Model(&Service{}).Where("acc_name = ?", c.form.AccName).Count(&count).Error)
				assert.Zero(t, count)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// A stored folder that the rules refuse is checked only when a save changes it or turns on its use.
	t.Run("StoredFolder", func(t *testing.T) {
		newStored := func(t *testing.T) *Service {
			model, err := AddService(form.Service{AccName: "Folder Stored " + t.Name(), AccURL: "test.com", AccType: "webdav",
				AccShare: true, AccSync: true, SyncDownload: true})
			require.NoError(t, err)
			t.Cleanup(func() { UnscopedDb().Unscoped().Delete(model) })
			require.NoError(t, model.Updates(Values{"sync_path": longName, "share_path": limit}))
			return storedService(t, model.ID)
		}

		for _, c := range []struct {
			name   string
			update func(f *form.Service)
			err    bool
		}{
			{"SyncOff", func(f *form.Service) { f.AccSync = false }, false},
			{"ShareOff", func(f *form.Service) { f.AccShare = false }, false},
			{"Password", func(f *form.Service) { f.AccPass = "changed" }, false},
			{"Name", func(f *form.Service) { f.AccName += " Renamed" }, false},
			{"RetryLimit", func(f *form.Service) { f.RetryLimit = 5 }, false},
			{"SyncPathChanged", func(f *form.Service) { f.SyncPath = longName + "x" }, true},
			{"SharePathChanged", func(f *form.Service) { f.SharePath = limit + "x" }, true},
			{"SharePathShortened", func(f *form.Service) { f.SharePath = "/share" }, false},
		} {
			t.Run(c.name, func(t *testing.T) {
				stored := newStored(t)
				frm, err := form.NewService(stored)
				require.NoError(t, err)
				c.update(&frm)

				if err = stored.SaveForm(frm); c.err {
					assert.ErrorIs(t, err, ErrServiceFolderTooLong)
				} else {
					assert.NoError(t, err)
				}
			})
		}

		// Turning on what a stored folder is used for checks it.
		t.Run("DownloadOn", func(t *testing.T) {
			stored := newStored(t)
			require.NoError(t, stored.Updates(Values{"sync_download": false, "share_path": "/"}))
			stored = storedService(t, stored.ID)
			frm, err := form.NewService(stored)
			require.NoError(t, err)
			frm.SyncDownload = true
			assert.ErrorIs(t, stored.SaveForm(frm), ErrServiceFolderTooLong)
		})
		t.Run("SyncOn", func(t *testing.T) {
			stored := newStored(t)
			require.NoError(t, stored.Updates(Values{"acc_sync": false, "sync_download": false, "share_path": "/"}))
			stored = storedService(t, stored.ID)
			require.NoError(t, stored.Updates(Values{"sync_path": limit}))
			stored = storedService(t, stored.ID)
			frm, err := form.NewService(stored)
			require.NoError(t, err)
			frm.AccSync = true
			assert.ErrorIs(t, stored.SaveForm(frm), ErrServiceFolderTooLong)
		})
		t.Run("ControlCharacter", func(t *testing.T) {
			// A folder stored with a control character compares equal to its cleaned form.
			stored := newStored(t)
			require.NoError(t, stored.Updates(Values{"sync_path": longName + "\x01"}))
			stored = storedService(t, stored.ID)
			frm, err := form.NewService(stored)
			require.NoError(t, err)
			frm.AccSync = false
			assert.NoError(t, stored.SaveForm(frm))
		})
		t.Run("ShareOn", func(t *testing.T) {
			stored := newStored(t)
			require.NoError(t, stored.Updates(Values{"acc_share": false}))
			stored = storedService(t, stored.ID)
			frm, err := form.NewService(stored)
			require.NoError(t, err)
			frm.AccShare = true
			assert.ErrorIs(t, stored.SaveForm(frm), ErrServiceFolderTooLong)
		})
	})
}

func TestServiceFolder(t *testing.T) {
	assert.Equal(t, "/", serviceFolder(""))
	assert.Equal(t, "/", serviceFolder("\x01"))
	assert.Equal(t, "/photos", serviceFolder("/photos"))
	assert.Equal(t, "/photos", serviceFolder("/pho\x00tos"))
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
	// Listing the folders of an account changes neither its error count nor its sharing, nor the errors of its files.
	for _, c := range []struct {
		name string
		fail bool
	}{{"FailureKeepsAccount", true}, {"SuccessKeepsAccount", false}} {
		t.Run(c.name, func(t *testing.T) {
			remote := t.TempDir()
			require.NoError(t, os.MkdirAll(filepath.Join(remote, "Photos"), fs.ModeDir))
			handler := &webdav.Handler{FileSystem: webdav.Dir(remote), LockSystem: webdav.NewMemLS()}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if c.fail {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				handler.ServeHTTP(w, r)
			}))
			t.Cleanup(server.Close)

			m := &Service{AccName: "Directories " + c.name, AccURL: server.URL + "/", AccType: "webdav", AccTimeout: "low",
				AccShare: true, AccSync: true, RetryLimit: 1, AccErrors: 2, AccError: "previous error"}
			require.NoError(t, Db().Create(m).Error)
			fileSync, fileShare := NewFileSync(m.ID, "/listing.jpg"), NewFileShare(1, m.ID, "listing.jpg")
			fileSync.Errors, fileShare.Errors = 4, 5
			require.NoError(t, fileSync.Create())
			require.NoError(t, fileShare.Create())
			t.Cleanup(func() {
				UnscopedDb().Delete(&FileSync{}, "service_id = ?", m.ID)
				UnscopedDb().Delete(&FileShare{}, "service_id = ?", m.ID)
				UnscopedDb().Unscoped().Delete(m)
			})

			result, err := m.Directories("")
			if c.fail {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Contains(t, result.Abs(), "/Photos")
			}

			var stored Service
			require.NoError(t, Db().First(&stored, m.ID).Error)
			assert.Equal(t, 2, stored.AccErrors)
			assert.Equal(t, "previous error", stored.AccError)
			assert.True(t, stored.AccShare)

			var storedSync FileSync
			require.NoError(t, Db().Where("service_id = ?", m.ID).First(&storedSync).Error)
			assert.Equal(t, 4, storedSync.Errors)
			var storedShare FileShare
			require.NoError(t, Db().Where("service_id = ?", m.ID).First(&storedShare).Error)
			assert.Equal(t, 5, storedShare.Errors)
		})
	}
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
	t.Run("Direct", func(t *testing.T) {
		m := newAccount(t, -1)
		if err := m.ResetErrors(true, true); err != nil {
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
