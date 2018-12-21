package test

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	c := NewConfig()

	assert.IsType(t, new(Config), c)

	assert.Equal(t, AssetsPath, c.AssetsPath())
	assert.False(t, c.Debug())
}

func TestConfig_ConnectToDatabase(t *testing.T) {
	c := NewConfig()

	db := c.Db()

	assert.IsType(t, &gorm.DB{}, db)
}
