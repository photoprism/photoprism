package oidc

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestClientConfig(t *testing.T) {
	c := config.NewConfig(config.CliTestContext())
	result := ClientConfig(c, config.ClientPublic)
	assert.IsType(t, config.Map{}, result)
}
