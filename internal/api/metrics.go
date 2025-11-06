package api

import (
	"io"
	"runtime"

	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// GetMetrics provides a Prometheus-compatible metrics endpoint for monitoring the instance, including usage details and portal cluster metrics.
//
//	@Summary	a prometheus-compatible metrics endpoint for monitoring this instance
//	@Id			GetMetrics
//	@Tags		Metrics
//	@Produce	plain
//	@Success	200		{object}	[]dto.MetricFamily
//	@Failure	401,403	{object}	i18n.Response
//	@Router		/api/v1/metrics [get]
func GetMetrics(router *gin.RouterGroup) {
	router.GET("/metrics", func(c *gin.Context) {
		s := Auth(c, acl.ResourceMetrics, acl.AccessAll)

		// Abort if permission is not granted.
		if s.Abort(c) {
			return
		}

		conf := get.Config()
		counts := conf.ClientUser(false).Count
		usage := conf.Usage()

		c.Header(header.ContentType, header.ContentTypePrometheus)

		c.Stream(func(w io.Writer) bool {
			reg := prometheus.NewRegistry()
			reg.MustRegister(collectors.NewGoCollector())

			factory := promauto.With(reg)

			registerCountMetrics(factory, counts)
			registerBuildInfoMetric(factory, conf.ClientPublic())
			registerUsageMetrics(factory, usage)
			registerClusterMetrics(factory, conf)

			var metrics []*dto.MetricFamily
			var err error

			metrics, err = reg.Gather()

			if err != nil {
				logErr("metrics", err)
				return false
			}

			for _, metric := range metrics {
				if _, err = expfmt.MetricFamilyToText(w, metric); err != nil {
					logErr("metrics", err)
					return false
				}
			}

			return false
		})
	})
}

// registerCountMetrics registers media count metrics exposed via /api/v1/metrics.
func registerCountMetrics(factory promauto.Factory, counts config.ClientCounts) {
	metric := factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "photoprism",
			Subsystem: "statistics",
			Name:      "media_count",
			Help:      "media statistics for this PhotoPrism instance",
		}, []string{"stat"},
	)

	stats := []struct {
		label string
		value int
	}{
		{"all", counts.All},
		{"photos", counts.Photos},
		{"media", counts.Media},
		{"animated", counts.Animated},
		{"live", counts.Live},
		{"audio", counts.Audio},
		{"videos", counts.Videos},
		{"documents", counts.Documents},
		{"cameras", counts.Cameras},
		{"lenses", counts.Lenses},
		{"countries", counts.Countries},
		{"hidden", counts.Hidden},
		{"archived", counts.Archived},
		{"favorites", counts.Favorites},
		{"review", counts.Review},
		{"stories", counts.Stories},
		{"private", counts.Private},
		{"albums", counts.Albums},
		{"private_albums", counts.PrivateAlbums},
		{"moments", counts.Moments},
		{"private_moments", counts.PrivateMoments},
		{"months", counts.Months},
		{"private_months", counts.PrivateMonths},
		{"states", counts.States},
		{"private_states", counts.PrivateStates},
		{"folders", counts.Folders},
		{"private_folders", counts.PrivateFolders},
		{"files", counts.Files},
		{"people", counts.People},
		{"places", counts.Places},
		{"labels", counts.Labels},
		{"label_max_photos", counts.LabelMaxPhotos},
		{"users", query.CountUsers(true, true, nil, []string{"guest"})},
		{"guests", query.CountUsers(true, true, []string{"guest"}, nil)},
	}

	for _, stat := range stats {
		metric.With(prometheus.Labels{"stat": stat.label}).Set(float64(stat.value))
	}
}

// registerBuildInfoMetric registers a metric that provides build information.
func registerBuildInfoMetric(factory promauto.Factory, conf *config.ClientConfig) {
	factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "photoprism",
			Name:      "build_info",
			Help:      "information about the photoprism instance",
		}, []string{"edition", "goversion", "version"},
	).With(prometheus.Labels{
		"edition":   conf.Edition,
		"goversion": runtime.Version(),
		"version":   conf.Version,
	}).Set(1.0)
}

// registerUsageMetrics registers filesystem and account usage metrics derived from the active configuration.
func registerUsageMetrics(factory promauto.Factory, usage config.Usage) {
	filesBytes := factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "photoprism",
			Subsystem: "usage",
			Name:      "files_bytes",
			Help:      "filesystem usage in bytes for files indexed by this PhotoPrism instance",
		}, []string{"state"},
	)

	filesBytes.With(prometheus.Labels{"state": "used"}).Set(float64(usage.FilesUsed))
	filesBytes.With(prometheus.Labels{"state": "free"}).Set(float64(usage.FilesFree))
	filesBytes.With(prometheus.Labels{"state": "total"}).Set(float64(usage.FilesTotal))

	filesPercent := factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "photoprism",
			Subsystem: "usage",
			Name:      "files_percent",
			Help:      "filesystem usage in percent for files indexed by this PhotoPrism instance",
		}, []string{"state"},
	)

	filesPercent.With(prometheus.Labels{"state": "used"}).Set(float64(usage.FilesUsedPct))
	filesPercent.With(prometheus.Labels{"state": "free"}).Set(float64(usage.FilesFreePct))

	accountsPercent := factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "photoprism",
			Subsystem: "usage",
			Name:      "accounts_percent",
			Help:      "account quota usage in percent for this PhotoPrism instance",
		}, []string{"state"},
	)

	accountsPercent.With(prometheus.Labels{"state": "used"}).Set(float64(usage.UsersUsedPct))
	accountsPercent.With(prometheus.Labels{"state": "free"}).Set(float64(usage.UsersFreePct))
}

// registerClusterMetrics exports cluster-specific metrics when running as a portal instance.
func registerClusterMetrics(factory promauto.Factory, conf *config.Config) {
	if !conf.Portal() {
		return
	}

	counts, err := clusterNodeCounts(conf)
	if err != nil {
		logErr("metrics", err)
		return
	}

	nodeMetric := factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "photoprism",
			Subsystem: "cluster",
			Name:      "nodes",
			Help:      "registered cluster nodes grouped by role",
		}, []string{"role"},
	)

	for role, value := range counts {
		nodeMetric.With(prometheus.Labels{"role": role}).Set(float64(value))
	}

	infoMetric := factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "photoprism",
			Subsystem: "cluster",
			Name:      "info",
			Help:      "cluster metadata for this PhotoPrism portal",
		}, []string{"uuid", "cidr"},
	)

	infoMetric.With(prometheus.Labels{
		"uuid": conf.ClusterUUID(),
		"cidr": conf.ClusterCIDR(),
	}).Set(1.0)
}

// clusterNodeCounts returns cluster node counts keyed by role plus a total entry.
func clusterNodeCounts(conf *config.Config) (map[string]int, error) {
	regy, err := reg.NewClientRegistryWithConfig(conf)
	if err != nil {
		return nil, err
	}

	nodes, err := regy.List()
	if err != nil {
		return nil, err
	}

	counts := map[string]int{"total": len(nodes)}
	for _, node := range nodes {
		role := node.Role
		if role == "" {
			role = "unknown"
		}
		counts[role]++
	}

	return counts, nil
}
