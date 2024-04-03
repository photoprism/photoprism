package api

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/pquerna/otp/totp"
	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
)

func TestCreateUserPasscode(t *testing.T) {
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateUserPasscode(router)
		r := PerformRequest(app, "POST", "/api/v1/users/uqxc08w3d0ej2283/passcode")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("Unauthorized", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		CreateUserPasscode(router)

		r := PerformRequest(app, "POST", "/api/v1/users/uqxc08w3d0ej2283/passcode")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("UsersDontMatch", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		CreateUserPasscode(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.UserPasscode{
			Passcode: "",
			Password: "Alice123!",
			Type:     "totp",
		}
		if pcStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "POST", "/api/v1/users/uqxc08w3d0ej2283/passcode", string(pcStr), sessId)
			assert.Equal(t, http.StatusForbidden, r.Code)
		}
	})
	t.Run("AliceUnsupportedType", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		CreateUserPasscode(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.UserPasscode{
			Passcode: "",
			Password: "abcdef",
			Type:     "xxx",
		}
		if pcStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "POST", "/api/v1/users/uqxetse3cy5eo9z2/passcode", string(pcStr), sessId)
			assert.Equal(t, http.StatusBadRequest, r.Code)
		}
	})
	t.Run("AliceInvalidPassword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		CreateUserPasscode(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.UserPasscode{
			Passcode: "",
			Password: "wrong",
			Type:     "totp",
		}
		if pcStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "POST", "/api/v1/users/uqxetse3cy5eo9z2/passcode", string(pcStr), sessId)
			assert.Equal(t, http.StatusForbidden, r.Code)
		}
	})
	t.Run("AliceSuccess", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		CreateUserPasscode(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.UserPasscode{
			Passcode: "",
			Password: "Alice123!",
			Type:     "totp",
		}
		if pcStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "POST", "/api/v1/users/uqxetse3cy5eo9z2/passcode", string(pcStr), sessId)
			assert.Equal(t, http.StatusOK, r.Code)
		}
	})
}

func TestConfirmUserPasscode(t *testing.T) {
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()
		ConfirmUserPasscode(router)
		r := PerformRequest(app, "POST", "/api/v1/users/uqxc08w3d0ej2283/passcode/confirm")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("AlicePasscodeTooShort", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		ConfirmUserPasscode(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.UserPasscode{
			Passcode: "123",
			Password: "Alice123!",
			Type:     "totp",
		}
		if pcStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "POST", "/api/v1/users/uqxetse3cy5eo9z2/passcode/confirm", string(pcStr), sessId)
			assert.Equal(t, http.StatusForbidden, r.Code)
		}
	})
	t.Run("AliceInvalidPassCode", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		ConfirmUserPasscode(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.UserPasscode{
			Passcode: "123456",
			Password: "Alice123!",
			Type:     "totp",
		}
		if pcStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "POST", "/api/v1/users/uqxetse3cy5eo9z2/passcode/confirm", string(pcStr), sessId)
			assert.Equal(t, http.StatusForbidden, r.Code)
		}
	})
}

func TestActivateUserPasscode(t *testing.T) {
	t.Run("UsersDontMatch", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		ActivateUserPasscode(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.UserPasscode{
			Passcode: "",
			Password: "Alice123!",
			Type:     "totp",
		}
		if pcStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "POST", "/api/v1/users/uqxc08w3d0ej2283/passcode/activate", string(pcStr), sessId)
			assert.Equal(t, http.StatusForbidden, r.Code)
		}
	})
	t.Run("InvalidPasscode", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		ActivateUserPasscode(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.UserPasscode{
			Passcode: "",
			Password: "Alice123!",
			Type:     "totp",
		}
		if pcStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "POST", "/api/v1/users/uqxetse3cy5eo9z2/passcode/activate", string(pcStr), sessId)
			assert.Equal(t, http.StatusForbidden, r.Code)
		}
	})
}

func TestDeactivateUserPasscode(t *testing.T) {
	t.Run("Unauthorized", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		DeactivateUserPasscode(router)

		r := PerformRequest(app, "POST", "/api/v1/users/uqxc08w3d0ej2283/passcode/deactivate")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("AliceInvalidPassword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		DeactivateUserPasscode(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.UserPasscode{
			Passcode: "",
			Password: "wrong",
			Type:     "totp",
		}
		if pcStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "POST", "/api/v1/users/uqxetse3cy5eo9z2/passcode/deactivate", string(pcStr), sessId)
			assert.Equal(t, http.StatusForbidden, r.Code)
		}
	})
}

func TestUserPasscode(t *testing.T) {
	//create
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)
	CreateUserPasscode(router)
	sessId := AuthenticateUser(app, router, "alice", "Alice123!")

	f0 := form.UserPasscode{
		Passcode: "",
		Password: "Alice123!",
		Type:     "totp",
	}

	pcStr, err := json.Marshal(f0)

	if err != nil {
		log.Fatal(err)
	}

	r := AuthenticatedRequestWithBody(app, "POST", "/api/v1/users/uqxetse3cy5eo9z2/passcode", string(pcStr), sessId)
	assert.Equal(t, http.StatusOK, r.Code)

	secret := gjson.Get(r.Body.String(), "Secret").String()
	activatedAt := gjson.Get(r.Body.String(), "ActivatedAt").String()
	verifiedAt := gjson.Get(r.Body.String(), "VerifiedAt").String()

	assert.Empty(t, activatedAt)
	assert.Empty(t, verifiedAt)

	code, err := totp.GenerateCode(secret, time.Now())

	//confirm
	ConfirmUserPasscode(router)

	if err != nil {
		t.Fatal(err)
	}

	f := form.UserPasscode{
		Passcode: code,
		Password: "Alice123!",
		Type:     "totp",
	}
	pcStr, err = json.Marshal(f)

	if err != nil {
		log.Fatal(err)
	}

	r = AuthenticatedRequestWithBody(app, "POST", "/api/v1/users/uqxetse3cy5eo9z2/passcode/confirm", string(pcStr), sessId)
	assert.Equal(t, http.StatusOK, r.Code)

	activatedAt = gjson.Get(r.Body.String(), "ActivatedAt").String()
	verifiedAt = gjson.Get(r.Body.String(), "VerifiedAt").String()

	assert.Empty(t, activatedAt)
	assert.NotEmpty(t, verifiedAt)

	//activate
	ActivateUserPasscode(router)

	r = AuthenticatedRequestWithBody(app, "POST", "/api/v1/users/uqxetse3cy5eo9z2/passcode/activate", string(pcStr), sessId)
	assert.Equal(t, http.StatusOK, r.Code)

	activatedAt = gjson.Get(r.Body.String(), "ActivatedAt").String()
	verifiedAt = gjson.Get(r.Body.String(), "VerifiedAt").String()

	assert.NotEmpty(t, activatedAt)
	assert.NotEmpty(t, verifiedAt)

	//deactivate
	DeactivateUserPasscode(router)

	r = AuthenticatedRequestWithBody(app, "POST", "/api/v1/users/uqxetse3cy5eo9z2/passcode/deactivate", string(pcStr), sessId)
	assert.Equal(t, http.StatusOK, r.Code)
}
