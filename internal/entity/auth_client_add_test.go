package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/form"
)

func Test_AddClient(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := form.Client{
			ClientName:   "test",
			AuthProvider: "client_credentials",
			AuthMethod:   "oauth2",
			AuthScope:    "all",
		}

		c, err := AddClient(m)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "test", c.ClientName)
	})
	t.Run("ClientNameEmpty", func(t *testing.T) {
		m := form.Client{
			ClientName:   "",
			AuthProvider: "client_credentials",
			AuthMethod:   "oauth2",
			AuthScope:    "all",
		}

		c, err := AddClient(m)

		if err == nil {
			t.Fatal("error expected")
		}

		assert.Equal(t, "", c.ClientName)
	})
	t.Run("ExistingClient", func(t *testing.T) {
		m := form.Client{
			ClientID: "cs5cpu17n6gj2qo5",
		}

		c, err := AddClient(m)

		if err == nil {
			t.Fatal("error expected")
		}

		assert.Equal(t, "Monitoring", c.ClientName)
	})
}

func Test_AddClient_WithRole(t *testing.T) {
	t.Run("AdminRole", func(t *testing.T) {
		frm := form.Client{
			ClientID:     "cs5cpu17n6gj9r10",
			ClientName:   "Role Admin",
			ClientRole:   "admin",
			AuthProvider: "client_credentials",
			AuthMethod:   "oauth2",
			AuthScope:    "all",
		}

		c, err := AddClient(frm)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "admin", c.ClientRole)
		assert.True(t, c.HasRole(acl.RoleAdmin))

		// Verify persisted role via lookup.
		persisted := FindClientByUID("cs5cpu17n6gj9r10")
		if persisted == nil {
			t.Fatal("persisted client not found")
		}
		assert.Equal(t, "admin", persisted.ClientRole)
	})
	t.Run("InvalidRoleDefaultsToClient", func(t *testing.T) {
		frm := form.Client{
			ClientID:     "cs5cpu17n6gj9r11",
			ClientName:   "Role Invalid",
			ClientRole:   "superuser",
			AuthProvider: "client_credentials",
			AuthMethod:   "oauth2",
			AuthScope:    "all",
		}

		c, err := AddClient(frm)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "client", c.ClientRole)
		assert.True(t, c.HasRole(acl.RoleClient))
	})
}
