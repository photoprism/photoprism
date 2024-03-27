package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCredentials(t *testing.T) {
	r := Credentials("Name", "Value", "secretName", "secretValue")
	assert.NotEmpty(t, r)
	assert.Contains(t, r, "secretName")
}
