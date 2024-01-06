package entity

import (
	"testing"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func Test_AddClient(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := form.Client{ClientName: "test", AuthMethod: "basic", AuthScope: "all"}

		c, err := AddClient(m)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "test", c.ClientName)
	})
	t.Run("ClientNameEmpty", func(t *testing.T) {
		m := form.Client{ClientName: "", AuthMethod: "basic", AuthScope: "all"}

		c, err := AddClient(m)

		if err == nil {
			t.Fatal("error expected")
		}

		assert.Equal(t, "", c.ClientName)
	})
}
