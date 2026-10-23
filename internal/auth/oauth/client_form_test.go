package oauth

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestValidateClientForm(t *testing.T) {
	app := form.Client{AuthMethod: authn.MethodPKCE.String(), RedirectURIs: []string{"https://claude.ai/cb"}}

	t.Run("Valid", func(t *testing.T) {
		assert.NoError(t, ValidateClientForm(app))
		assert.NoError(t, ValidateClientForm(form.Client{AuthMethod: authn.MethodOAuth2.String()}))
	})
	t.Run("NoRedirectURI", func(t *testing.T) {
		f := app
		f.RedirectURIs = nil
		assert.EqualError(t, ValidateClientForm(f), "app clients require at least one redirect URI")
	})
	t.Run("User", func(t *testing.T) {
		f := app
		f.UserName = "alice"
		assert.EqualError(t, ValidateClientForm(f), "app clients cannot belong to a user")
	})
	t.Run("Secret", func(t *testing.T) {
		f := app
		f.ClientSecret = "xcCbOrw6I0vcoXzhnOmXhjpVSyFq9i8u"
		assert.EqualError(t, ValidateClientForm(f), "app clients do not use a client secret")
	})
	t.Run("Role", func(t *testing.T) {
		f := app
		f.ClientRole = "instance"
		assert.EqualError(t, ValidateClientForm(f), "app clients must have the client role")
		f.ClientRole = "client"
		assert.NoError(t, ValidateClientForm(f))
	})
	t.Run("RedirectURIWithoutApp", func(t *testing.T) {
		assert.EqualError(t, ValidateClientForm(form.Client{AuthMethod: authn.MethodOAuth2.String(), RedirectURIs: []string{"https://claude.ai/cb"}}), "redirect URIs can only be set for app clients")
	})
	t.Run("InvalidRedirectURI", func(t *testing.T) {
		f := app
		f.RedirectURIs = []string{"http://example.com/cb"}
		assert.EqualError(t, ValidateClientForm(f), "invalid redirect URI 'http://example.com/cb'")
	})
}

func TestValidateClientChange(t *testing.T) {
	app := &entity.Client{AuthMethod: authn.MethodPKCE.String()}
	api := &entity.Client{AuthMethod: authn.MethodOAuth2.String()}

	t.Run("Valid", func(t *testing.T) {
		assert.NoError(t, ValidateClientChange(app, form.Client{ClientName: "App"}, ClientChange{}))
		assert.NoError(t, ValidateClientChange(app, form.Client{RedirectURIs: []string{"https://claude.ai/cb"}}, ClientChange{}))
		assert.NoError(t, ValidateClientChange(app, form.Client{ClientRole: "client"}, ClientChange{Role: true}))
		assert.NoError(t, ValidateClientChange(app, form.Client{AuthMethod: authn.MethodPKCE.String()}, ClientChange{Method: true}))
		assert.NoError(t, ValidateClientChange(api, form.Client{ClientRole: "admin"}, ClientChange{Role: true, NewSecret: true}))
	})
	t.Run("NoClient", func(t *testing.T) {
		assert.EqualError(t, ValidateClientChange(nil, form.Client{}, ClientChange{}), "client not found")
	})
	t.Run("MethodSwitch", func(t *testing.T) {
		assert.EqualError(t, ValidateClientChange(app, form.Client{AuthMethod: authn.MethodOAuth2.String()}, ClientChange{Method: true}), "the authentication method of app clients cannot be changed")
		assert.EqualError(t, ValidateClientChange(api, form.Client{AuthMethod: authn.MethodPKCE.String()}, ClientChange{Method: true}), "the authentication method of app clients cannot be changed")
	})
	t.Run("MethodValue", func(t *testing.T) {
		assert.EqualError(t, ValidateClientChange(api, form.Client{AuthMethod: authn.MethodPKCE.String()}, ClientChange{}), "the authentication method of app clients cannot be changed")
		assert.EqualError(t, ValidateClientChange(app, form.Client{AuthMethod: "jwt"}, ClientChange{}), "the authentication method of app clients cannot be changed")
		assert.NoError(t, ValidateClientChange(api, form.Client{AuthMethod: authn.MethodOAuth2.String()}, ClientChange{}))
	})
	t.Run("RedirectURIWithoutApp", func(t *testing.T) {
		assert.EqualError(t, ValidateClientChange(api, form.Client{RedirectURIs: []string{"https://claude.ai/cb"}}, ClientChange{}), "redirect URIs can only be set for app clients")
		assert.EqualError(t, ValidateClientChange(api, form.Client{RedirectURIs: []string{}}, ClientChange{}), "redirect URIs can only be set for app clients")
	})
	t.Run("Secret", func(t *testing.T) {
		assert.EqualError(t, ValidateClientChange(app, form.Client{ClientSecret: "xcCbOrw6I0vcoXzhnOmXhjpVSyFq9i8u"}, ClientChange{}), "app clients do not use a client secret")
		assert.EqualError(t, ValidateClientChange(app, form.Client{}, ClientChange{NewSecret: true}), "app clients do not use a client secret")
	})
	t.Run("Role", func(t *testing.T) {
		assert.EqualError(t, ValidateClientChange(app, form.Client{ClientRole: "admin"}, ClientChange{Role: true}), "app clients must have the client role")
		assert.EqualError(t, ValidateClientChange(app, form.Client{ClientRole: ""}, ClientChange{Role: true}), "app clients must have the client role")
		assert.NoError(t, ValidateClientChange(app, form.Client{ClientRole: "admin"}, ClientChange{}))
	})
	t.Run("InvalidRedirectURI", func(t *testing.T) {
		assert.EqualError(t, ValidateClientChange(app, form.Client{RedirectURIs: []string{"javascript:alert(1)"}}, ClientChange{}), "invalid redirect URI 'javascript:alert(1)'")
	})
}

func TestValidateRedirectURIs(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.NoError(t, ValidateRedirectURIs([]string{"https://claude.ai/cb", "http://localhost:6274/oauth/callback", "app.photoprism.mobile:/oauth"}))
		assert.NoError(t, ValidateRedirectURIs(nil))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Error(t, ValidateRedirectURIs([]string{"https://claude.ai/cb", "http://example.com/cb"}))
		assert.Error(t, ValidateRedirectURIs([]string{"javascript:alert(1)"}))
	})
}
