package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/service/cluster"
)

func TestOptions_Report(t *testing.T) {
	m := Options{}
	r, _ := m.Report()
	assert.GreaterOrEqual(t, len(r), 1)
}

func TestOptions_ReportSkipsInlineAndNonFlags(t *testing.T) {
	rows, _ := Options{}.Report()

	for _, row := range rows {
		if len(row) < 3 {
			t.Fatalf("expected report row with 3 columns, got %v", row)
		}

		assert.NotEqual(t, ",inline,omitempty", row[0])
		assert.NotEqual(t, "---", row[2])
	}
}

func TestOptions_ReportFrontendUriVisibility(t *testing.T) {
	hasFrontendUri := func(rows [][]string) bool {
		for _, row := range rows {
			if len(row) > 0 && row[0] == "FrontendUri" {
				return true
			}
		}

		return false
	}

	originalFeatures := Features
	t.Cleanup(func() { Features = originalFeatures })

	t.Run("CommunityInstance", func(t *testing.T) {
		Features = Community
		m := Options{NodeRole: cluster.RoleInstance}
		rows, _ := m.Report()
		assert.False(t, hasFrontendUri(rows))
	})
	t.Run("ProInstance", func(t *testing.T) {
		Features = Pro
		m := Options{NodeRole: cluster.RoleInstance}
		rows, _ := m.Report()
		assert.True(t, hasFrontendUri(rows))
	})
	t.Run("ProPortalNode", func(t *testing.T) {
		Features = Pro
		m := Options{NodeRole: cluster.RolePortal}
		rows, _ := m.Report()
		assert.True(t, hasFrontendUri(rows))
	})
	t.Run("CommunityPortalNode", func(t *testing.T) {
		Features = Community
		m := Options{NodeRole: cluster.RolePortal}
		rows, _ := m.Report()
		assert.True(t, hasFrontendUri(rows))
	})
}
