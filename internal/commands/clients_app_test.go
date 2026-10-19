package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestValidateAppClientForm(t *testing.T) {
	app := form.Client{AuthMethod: authn.MethodPKCE.String(), RedirectURIs: []string{"https://claude.ai/cb"}}

	t.Run("Valid", func(t *testing.T) {
		assert.NoError(t, validateAppClientForm(app))
		assert.NoError(t, validateAppClientForm(form.Client{AuthMethod: authn.MethodOAuth2.String()}))
	})
	t.Run("NoCallback", func(t *testing.T) {
		f := app
		f.RedirectURIs = nil
		assert.Error(t, validateAppClientForm(f))
	})
	t.Run("User", func(t *testing.T) {
		f := app
		f.UserName = "alice"
		assert.Error(t, validateAppClientForm(f))
	})
	t.Run("Secret", func(t *testing.T) {
		f := app
		f.ClientSecret = "xcCbOrw6I0vcoXzhnOmXhjpVSyFq9i8u"
		assert.Error(t, validateAppClientForm(f))
	})
	t.Run("Role", func(t *testing.T) {
		f := app
		f.ClientRole = "instance"
		assert.Error(t, validateAppClientForm(f))
		f.ClientRole = "client"
		assert.NoError(t, validateAppClientForm(f))
	})
	t.Run("CallbackWithoutApp", func(t *testing.T) {
		assert.Error(t, validateAppClientForm(form.Client{AuthMethod: authn.MethodOAuth2.String(), RedirectURIs: []string{"https://claude.ai/cb"}}))
	})
}

func TestValidateCallbackURLs(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.NoError(t, validateCallbackURLs([]string{"https://claude.ai/cb", "http://localhost:6274/oauth/callback", "app.photoprism.mobile:/oauth"}))
		assert.NoError(t, validateCallbackURLs(nil))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Error(t, validateCallbackURLs([]string{"https://claude.ai/cb", "http://example.com/cb"}))
		assert.Error(t, validateCallbackURLs([]string{"javascript:alert(1)"}))
	})
}
