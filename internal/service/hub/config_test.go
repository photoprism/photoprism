package hub

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_MapKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := NewConfig("test", "testdata/new.yml", "zqkunt22r0bewti9", "test", "PhotoPrism/Test", "test")
		assert.Equal(t, "", c.MapKey())
	})
}

func TestConfig_Sponsor(t *testing.T) {
	t.Run("Status", func(t *testing.T) {
		c := NewConfig("test", "testdata/new.yml", "zr58wrg19i8jfjam", "test", "PhotoPrism/Test", "test")
		c.Key = "0e159b773c6fb779c3bf6c8ba6e322abf559dbaf"
		c.Secret = "23f0024975bd65ade06edcc8191f7fcc"
		assert.False(t, c.Sponsor())
		c.Status = "sponsor"
		assert.False(t, c.Sponsor())
		c.Session = "bde6d0cf514e5456591de5ae09d981056eb88dccf71ba268974bf2cc7b028545e7641c1dfbaa716e4d13f8b0e0d1863e64c16e1f0ac551fc85e1171a87cbda6608cbe330de9e0d5f5b0e14ff61f2ff08fee369"
		assert.True(t, c.Sponsor())
		c.Status = ""
		assert.False(t, c.Sponsor())
		c.Session = ""
	})
}
