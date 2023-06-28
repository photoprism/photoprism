package form

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/authn"

	"github.com/stretchr/testify/assert"
)

func TestUser(t *testing.T) {
	form := &User{UserName: "foobar", UserEmail: "test@test.com", Password: "passwd"}

	assert.Equal(t, "foobar", form.UserName)
	assert.Equal(t, "test@test.com", form.UserEmail)
	assert.Equal(t, "passwd", form.Password)
}

func TestUser_Username(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		form := &User{UserName: "", UserEmail: "test@test.com", Password: "passwd"}
		assert.Equal(t, "", form.Username())
	})
	t.Run("Valid", func(t *testing.T) {
		form := &User{UserName: "foobar", UserEmail: "test@test.com", Password: "passwd"}
		assert.Equal(t, "foobar", form.Username())
	})
	t.Run("Invalid", func(t *testing.T) {
		form := &User{UserName: " Foo Bar4w45 !", UserEmail: "test@test.com", Password: "passwd"}
		assert.Equal(t, "foo bar4w45 !", form.Username())
	})
}

func TestUser_Provider(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		form := &User{UserName: "", UserEmail: "test@test.com", Password: "passwd", AuthProvider: ""}
		assert.Equal(t, authn.ProviderDefault, form.Provider())
	})
	t.Run("Valid", func(t *testing.T) {
		form := &User{UserName: "John", UserEmail: "test@test.com", Password: "passwd", AuthProvider: "local"}
		assert.Equal(t, authn.ProviderLocal, form.Provider())
	})
}

func TestUser_Email(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		form := &User{UserName: "foobar", UserEmail: "", Password: "passwd"}
		assert.Equal(t, "", form.Email())
	})
	t.Run("Valid", func(t *testing.T) {
		form := &User{UserName: "foobar", UserEmail: "test@test.com", Password: "passwd"}
		assert.Equal(t, "test@test.com", form.Email())
	})
	t.Run("Invalid", func(t *testing.T) {
		form := &User{UserName: " Foo Bar4w45 !", UserEmail: "  testtest.srg awrygcom  ", Password: "passwd"}
		assert.Equal(t, "", form.Email())
	})
}

func TestUser_Role(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		form := &User{UserName: "", UserEmail: "test@test.com", Password: "passwd", UserRole: ""}
		assert.Equal(t, "", form.Role())
	})
	t.Run("Valid", func(t *testing.T) {
		form := &User{UserName: "John", UserEmail: "test@test.com", Password: "passwd", UserRole: "admin"}
		assert.Equal(t, "admin", form.Role())
	})
	t.Run("Invalid", func(t *testing.T) {
		form := &User{UserName: "John", UserEmail: "test@test.com", Password: "passwd", UserRole: "ad&*min"}
		assert.Equal(t, "admin", form.Role())
	})
}

func TestUser_Attr(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		form := &User{UserName: "", UserEmail: "test@test.com", Password: "passwd", UserAttr: ""}
		assert.Equal(t, "", form.Attr())
	})
}
