package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestS2Encode(t *testing.T) {
	t.Run("Wildgehege", func(t *testing.T) {
		plusCode := S2Encode(48.56344833333333, 8.996878333333333)
		expected := uint64(0x799e370c0000000)

		assert.Equal(t, expected, plusCode)
	})

	t.Run("LatOverflow", func(t *testing.T) {
		plusCode := S2Encode(548.56344833333333, 8.996878333333333)
		expected := uint64(0)

		assert.Equal(t, expected, plusCode)
	})

	t.Run("LongOverflow", func(t *testing.T) {
		plusCode := S2Encode(48.56344833333333, 258.996878333333333)
		expected := uint64(0)

		assert.Equal(t, expected, plusCode)
	})
}

func TestS2EncodeLevel(t *testing.T) {
	t.Run("Wildgehege30", func(t *testing.T) {
		plusCode := S2EncodeLevel(48.56344833333333, 8.996878333333333, 30)
		expected := uint64(0x799e370ca54c8b9)

		assert.Equal(t, expected, plusCode)
	})

	t.Run("Wildgehege18", func(t *testing.T) {
		plusCode := S2EncodeLevel(48.56344833333333, 8.996878333333333, 18)
		expected := uint64(0x799e370cb000000)

		assert.Equal(t, expected, plusCode)
	})

	t.Run("Wildgehege10", func(t *testing.T) {
		plusCode := S2EncodeLevel(48.56344833333333, 8.996878333333333, 10)
		expected := uint64(0x799e30000000000)

		assert.Equal(t, expected, plusCode)
	})

	t.Run("LatOverflow", func(t *testing.T) {
		plusCode := S2EncodeLevel(548.56344833333333, 8.996878333333333, 10)
		expected := uint64(0)

		assert.Equal(t, expected, plusCode)
	})

	t.Run("LongOverflow", func(t *testing.T) {
		plusCode := S2EncodeLevel(48.56344833333333, 258.996878333333333, 15)
		expected := uint64(0)

		assert.Equal(t, expected, plusCode)
	})
}

func TestS2Token(t *testing.T) {
	t.Run("Wildgehege", func(t *testing.T) {
		plusCode := S2Token(48.56344833333333, 8.996878333333333)
		expected := "4799e370c"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("LatOverflow", func(t *testing.T) {
		plusCode := S2Token(548.56344833333333, 8.996878333333333)
		expected := ""

		assert.Equal(t, expected, plusCode)
	})

	t.Run("LongOverflow", func(t *testing.T) {
		plusCode := S2Token(48.56344833333333, 258.996878333333333)
		expected := ""

		assert.Equal(t, expected, plusCode)
	})
}

func TestS2TokenLevel(t *testing.T) {
	t.Run("Wildgehege30", func(t *testing.T) {
		plusCode := S2TokenLevel(48.56344833333333, 8.996878333333333, 30)
		expected := "4799e370ca54c8b9"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("NearWildgehege30", func(t *testing.T) {
		plusCode := S2TokenLevel(48.56344839999999, 8.996878339999999, 30)
		expected := "4799e370ca54c8b7"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("Wildgehege18", func(t *testing.T) {
		plusCode := S2TokenLevel(48.56344833333333, 8.996878333333333, 18)
		expected := "4799e370cb"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("NearWildgehege18", func(t *testing.T) {
		plusCode := S2TokenLevel(48.56344839999999, 8.996878339999999, 18)
		expected := "4799e370cb"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("NearWildgehege15", func(t *testing.T) {
		plusCode := S2TokenLevel(48.56344833333333, 8.996878333333333, 15)
		expected := "4799e370c"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("Wildgehege10", func(t *testing.T) {
		plusCode := S2TokenLevel(48.56344833333333, 8.996878333333333, 10)
		expected := "4799e3"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("LatOverflow", func(t *testing.T) {
		plusCode := S2TokenLevel(548.56344833333333, 8.996878333333333, 30)
		expected := ""

		assert.Equal(t, expected, plusCode)
	})

	t.Run("LongOverflow", func(t *testing.T) {
		plusCode := S2TokenLevel(48.56344833333333, 258.996878333333333, 30)
		expected := ""

		assert.Equal(t, expected, plusCode)
	})
}
