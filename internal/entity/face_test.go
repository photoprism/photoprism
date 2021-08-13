package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFace_TableName(t *testing.T) {
	m := &Face{}
	assert.Contains(t, m.TableName(), "faces")
}
