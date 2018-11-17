package test

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	c := NewConfig()

	assert.IsType(t, new(Config), c)

	assert.Equal(t, AssetsPath, c.GetAssetsPath())
	assert.False(t, c.IsDebug())
}

func TestConfig_ConnectToDatabase(t *testing.T) {
	c := NewConfig()

	db := c.GetDb()

	assert.IsType(t, &gorm.DB{}, db)
}
