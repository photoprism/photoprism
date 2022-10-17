package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser(t *testing.T) {
	form := &User{UserName: "foobar", UserEmail: "test@test.com", Password: "passwd"}

	assert.Equal(t, "foobar", form.UserName)
	assert.Equal(t, "test@test.com", form.UserEmail)
	assert.Equal(t, "passwd", form.Password)
}

func TestNewUser(t *testing.T) {
	val := &User{UserName: "foobar", UserEmail: "test@test.com", Password: "passwd"}
	form, err := NewUser(val)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "foobar", form.UserName)
	assert.Equal(t, "test@test.com", form.UserEmail)
	assert.Equal(t, "passwd", form.Password)
}

func TestUser_Name(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		form := &User{UserName: "", UserEmail: "test@test.com", Password: "passwd"}
		assert.Equal(t, "", form.Name())
	})
	t.Run("Valid", func(t *testing.T) {
		form := &User{UserName: "foobar", UserEmail: "test@test.com", Password: "passwd"}
		assert.Equal(t, "foobar", form.Name())
	})
	t.Run("Invalid", func(t *testing.T) {
		form := &User{UserName: " Foo Bar4w45 !", UserEmail: "test@test.com", Password: "passwd"}
		assert.Equal(t, "foobar4w45", form.Name())
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
