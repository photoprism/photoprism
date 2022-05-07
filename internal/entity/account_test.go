package entity

import (
	"testing"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestCreateAccount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		account := Account{AccName: "Foo", AccOwner: "bar", AccURL: "test.com", AccType: "webdav", AccKey: "123", AccUser: "testuser", AccPass: "testpass",
			AccError: "", AccShare: true, AccSync: true, RetryLimit: 4, SharePath: "/home", ShareSize: "500", ShareExpires: 3500, SyncPath: "/sync",
			SyncInterval: 5, SyncUpload: true, SyncDownload: false, SyncFilenames: true, SyncRaw: false}

		accountForm, err := form.NewAccount(account)

		if err != nil {
			t.Fatal(err)
		}

		model, err := CreateAccount(accountForm)

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

func TestAccount_SaveForm(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		account := Account{AccName: "Foo", AccOwner: "bar", AccURL: "test.com", AccType: "test", AccKey: "123", AccUser: "testuser", AccPass: "testpass",
			AccError: "", AccShare: true, AccSync: true, RetryLimit: 4, SharePath: "/home", ShareSize: "500", ShareExpires: 3500, SyncPath: "/sync",
			SyncInterval: 5, SyncUpload: true, SyncDownload: true, SyncFilenames: true, SyncRaw: false}

		accountForm, err := form.NewAccount(account)

		if err != nil {
			t.Fatal(err)
		}
		model, err := CreateAccount(accountForm)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, true, model.SyncDownload)
		assert.Equal(t, false, model.SyncUpload)
		assert.Equal(t, "Foo", model.AccName)
		assert.Equal(t, "bar", model.AccOwner)
		assert.Equal(t, "test.com", model.AccURL)

		accountUpdate := Account{AccName: "NewName", AccOwner: "NewOwner", AccURL: "new.com", SyncUpload: true, SyncDownload: true}

		UpdateForm, err := form.NewAccount(accountUpdate)

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

func TestAccount_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		account := Account{AccName: "DeleteAccount", AccOwner: "Delete", AccURL: "test.com", AccType: "test", AccKey: "123", AccUser: "testuser", AccPass: "testpass",
			AccError: "", AccShare: true, AccSync: true, RetryLimit: 4, SharePath: "/home", ShareSize: "500", ShareExpires: 3500, SyncPath: "/sync",
			SyncInterval: 5, SyncUpload: true, SyncDownload: false, SyncFilenames: true, SyncRaw: false}

		accountForm, err := form.NewAccount(account)

		if err != nil {
			t.Fatal(err)
		}
		model, err := CreateAccount(accountForm)

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

func TestAccount_Directories(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		account := Account{AccName: "DirectoriesAccount", AccOwner: "Owner", AccURL: "http://dummy-webdav/", AccType: "webdav", AccKey: "123", AccUser: "admin", AccPass: "photoprism",
			AccError: "", AccShare: true, AccSync: true, RetryLimit: 4, SharePath: "/home", ShareSize: "500", ShareExpires: 3500, SyncPath: "/sync",
			SyncInterval: 5, SyncUpload: true, SyncDownload: false, SyncFilenames: true, SyncRaw: false}

		accountForm, err := form.NewAccount(account)

		if err != nil {
			t.Fatal(err)
		}
		model, err := CreateAccount(accountForm)

		if err != nil {
			t.Fatal(err)
		}

		result, err := model.Directories()

		if err != nil {
			t.Fatal(err)
		}
		assert.NotEmpty(t, result.Abs())
		assert.Contains(t, result.Abs(), "/Photos")

	})
	t.Run("no directory", func(t *testing.T) {
		account := Account{AccName: "DirectoriesAccount", AccOwner: "Owner", AccURL: "http://dummy-webdav/", AccType: "xxx", AccKey: "123", AccUser: "admin", AccPass: "photoprism",
			AccError: "", AccShare: true, AccSync: true, RetryLimit: 4, SharePath: "/home", ShareSize: "500", ShareExpires: 3500, SyncPath: "/sync",
			SyncInterval: 5, SyncUpload: true, SyncDownload: false, SyncFilenames: true, SyncRaw: false}

		accountForm, err := form.NewAccount(account)

		if err != nil {
			t.Fatal(err)
		}
		model, err := CreateAccount(accountForm)

		if err != nil {
			t.Fatal(err)
		}

		result, err := model.Directories()

		if err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, result.Abs())
	})
}

func TestAccount_Updates(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		account := Account{AccName: "DeleteAccount", AccOwner: "Delete", AccURL: "test.com", AccType: "test", AccKey: "123", AccUser: "testuser", AccPass: "testpass",
			AccError: "", AccShare: true, AccSync: true, RetryLimit: 4, SharePath: "/home", ShareSize: "500", ShareExpires: 3500, SyncPath: "/sync",
			SyncInterval: 5, SyncUpload: true, SyncDownload: false, SyncFilenames: true, SyncRaw: false}

		accountForm, err := form.NewAccount(account)

		if err != nil {
			t.Fatal(err)
		}
		model, err := CreateAccount(accountForm)
		assert.Equal(t, "testuser", model.AccUser)
		assert.Equal(t, "DeleteAccount", model.AccName)

		if err != nil {
			t.Fatal(err)
		}

		err = model.Updates(Account{AccName: "UpdatedName", AccUser: "UpdatedUser"})
		assert.Equal(t, "UpdatedUser", model.AccUser)
		assert.Equal(t, "UpdatedName", model.AccName)

		if err != nil {
			t.Fatal(err)
		}

	})
}

func TestAccount_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		account := Account{AccName: "DeleteAccount", AccOwner: "Delete", AccURL: "test.com", AccType: "test", AccKey: "123", AccUser: "testuser", AccPass: "testpass",
			AccError: "", AccShare: true, AccSync: true, RetryLimit: 4, SharePath: "/home", ShareSize: "500", ShareExpires: 3500, SyncPath: "/sync",
			SyncInterval: 5, SyncUpload: true, SyncDownload: false, SyncFilenames: true, SyncRaw: false}

		accountForm, err := form.NewAccount(account)

		if err != nil {
			t.Fatal(err)
		}
		model, err := CreateAccount(accountForm)

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

//TODO fails on mariadb
func TestAccount_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		account := Account{AccName: "DeleteAccount", AccOwner: "Delete", AccURL: "test.com", AccType: "test", AccKey: "123", AccUser: "testuser", AccPass: "testpass",
			AccError: "", AccShare: true, AccSync: true, RetryLimit: 4, SharePath: "/home", ShareSize: "500", ShareExpires: 3500, SyncPath: "/sync",
			SyncInterval: 5, SyncUpload: true, SyncDownload: false, SyncFilenames: true, SyncRaw: false}

		accountForm, err := form.NewAccount(account)

		if err != nil {
			t.Fatal(err)
		}
		model, err := CreateAccount(accountForm)

		if err != nil {
			t.Fatal(err)
		}
		initialDate := model.UpdatedAt

		err = model.Save()

		if err != nil {
			t.Fatal(err)
		}
		afterDate := model.UpdatedAt
		assert.True(t, afterDate.After(initialDate))
	})
}

func TestAccount_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		account := Account{}

		err := account.Create()

		if err != nil {
			t.Fatal(err)
		}
	})
}
