package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMomentsTime(t *testing.T) {
	t.Run("result found", func(t *testing.T) {
		result, err := GetMomentsTime()

		assert.Nil(t, err)
		assert.Equal(t, 2790, result[0].PhotoYear)
		assert.Equal(t, 2, result[0].Count)
	})
}
