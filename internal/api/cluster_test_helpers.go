package api

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service/cluster"
)

// enablePortalAPIs configures test settings so cluster/portal API routes are enabled.
func enablePortalAPIs(t testing.TB, conf *config.Config) {
	t.Helper()

	prevEdition := conf.Options().Edition
	prevRole := conf.Options().NodeRole
	prevClusterCIDR := conf.Options().ClusterCIDR

	t.Cleanup(func() {
		conf.Options().Edition = prevEdition
		conf.Options().NodeRole = prevRole
		conf.Options().ClusterCIDR = prevClusterCIDR
	})

	conf.Options().Edition = config.Portal
	conf.Options().NodeRole = cluster.RolePortal
	conf.Options().ClusterCIDR = ""
}
