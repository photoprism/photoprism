package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeUnscanned(t *testing.T) {
	baseline := map[string]int{
		"internal/api/a.go\tevent.AuditInfo\tvalue":    2,
		"pro/internal/api/b.go\tevent.AuditWarn\tname": 1,
		"portal/internal/c.go\tevent.AuditErr\tid":     3,
	}
	t.Run("EditionsMissing", func(t *testing.T) {
		keys := map[string]int{"internal/api/a.go\tevent.AuditInfo\tvalue": 1}

		merged := mergeUnscanned(keys, baseline, []string{"internal", "pkg"})

		assert.Equal(t, map[string]int{
			"internal/api/a.go\tevent.AuditInfo\tvalue":    1,
			"pro/internal/api/b.go\tevent.AuditWarn\tname": 1,
			"portal/internal/c.go\tevent.AuditErr\tid":     3,
		}, merged)
	})
	t.Run("AllScanned", func(t *testing.T) {
		keys := map[string]int{"internal/api/a.go\tevent.AuditInfo\tvalue": 1}

		merged := mergeUnscanned(keys, baseline, []string{"internal", "pkg", "pro/internal", "portal/internal"})

		assert.Equal(t, keys, merged)
	})
	t.Run("EmptyBaseline", func(t *testing.T) {
		keys := map[string]int{"internal/api/a.go\tevent.AuditInfo\tvalue": 1}

		assert.Equal(t, keys, mergeUnscanned(keys, map[string]int{}, []string{"internal"}))
	})
}

func TestUnderRoot(t *testing.T) {
	t.Run("Inside", func(t *testing.T) {
		assert.True(t, underRoot("internal/api/a.go", []string{"internal"}))
		assert.True(t, underRoot("pro/internal/api/b.go", []string{"internal", "pro/internal/"}))
		assert.True(t, underRoot("internal/api/a.go", []string{"./internal"}))
		assert.True(t, underRoot("internal/api/a.go", []string{"."}))
	})
	t.Run("Outside", func(t *testing.T) {
		assert.False(t, underRoot("pro/internal/api/b.go", []string{"internal", "pkg"}))
		assert.False(t, underRoot("internalx/a.go", []string{"internal"}))
		assert.False(t, underRoot("internal/api/a.go", nil))
	})
}
