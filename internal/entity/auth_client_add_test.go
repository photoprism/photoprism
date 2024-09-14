package entity

import (
	"testing"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
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
