package level

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevelJsonEncoding(t *testing.T) {
	type X struct {
		Level Severity
	}

	var x X
	x.Level = Warning
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	assert.NoError(t, enc.Encode(x))
	dec := json.NewDecoder(&buf)
	var y X
	assert.NoError(t, dec.Decode(&y))
}

func TestLevelUnmarshalText(t *testing.T) {
	var u Severity
	for _, level := range Levels {
		t.Run(level.String(), func(t *testing.T) {
			assert.NoError(t, u.UnmarshalText([]byte(level.String())))
			assert.Equal(t, level, u)
		})
	}
	t.Run("invalid", func(t *testing.T) {
		assert.Error(t, u.UnmarshalText([]byte("invalid")))
	})
}

func TestLevelMarshalText(t *testing.T) {
	levelStrings := []string{
		"emergency",
		"alert",
		"critical",
		"error",
		"warning",
		"notice",
		"info",
		"debug",
	}
	for idx, val := range Levels {
		level := val
		t.Run(level.String(), func(t *testing.T) {
			var cmp Severity
			b, err := level.MarshalText()
			assert.NoError(t, err)
			assert.Equal(t, levelStrings[idx], string(b))
			err = cmp.UnmarshalText(b)
			assert.NoError(t, err)
			assert.Equal(t, level, cmp)
		})
	}
}
