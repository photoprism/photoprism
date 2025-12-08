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
	"github.com/photoprism/photoprism/internal/photoprism/get"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/http/header"
)

const (
	metricsNamespace           = "photoprism"
	metricsUsageSubsystem      = "usage"
	metricsStatisticsSubsystem = "statistics"
	metricsClusterSubsystem    = "cluster"

	metricsLabelState   = "state"
	metricsLabelStat    = "stat"
	metricsLabelRole    = "role"
	metricsLabelUUID    = "uuid"
	metricsLabelCIDR    = "cidr"
	metricsLabelEdition = "edition"
	metricsLabelGoVer   = "goversion"
	metricsLabelVersion = "version"

	metricFilesBytes     = "files_bytes"
	metricFilesRatio     = "files_ratio"
	metricAccountsRatio  = "accounts_ratio"
	metricAccountsActive = "accounts_active"
	metricMediaCount     = "media_count"
	metricBuildInfo      = "build_info"
	metricClusterNodes   = "nodes"
	metricClusterInfo    = "info"

	metricsAccountsHelp      = "active user and guest accounts on this PhotoPrism instance"
	metricsFilesBytesHelp    = "filesystem usage in bytes for files indexed by this PhotoPrism instance"
	metricsFilesRatioHelp    = "filesystem usage for files indexed by this PhotoPrism instance"
	metricsAccountsRatioHelp = "account quota usage for this PhotoPrism instance"
	metricsMediaCountHelp    = "media statistics for this PhotoPrism instance"
	metricsBuildInfoHelp     = "information about the photoprism instance"
	metricsClusterNodesHelp  = "registered cluster nodes grouped by role"
	metricsClusterInfoHelp   = "cluster metadata for this PhotoPrism portal"
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
			registry := prometheus.NewRegistry()
			registry.MustRegister(collectors.NewGoCollector())

			factory := promauto.With(registry)

			registerCountMetrics(factory, counts)
			registerBuildInfoMetric(factory, conf.ClientPublic())
			registerUsageMetrics(factory, usage)
			registerClusterMetrics(factory, conf)

			var metrics []*dto.MetricFamily
			var err error

			metrics, err = registry.Gather()

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
			Namespace: metricsNamespace,
			Subsystem: metricsStatisticsSubsystem,
			Name:      metricMediaCount,
			Help:      metricsMediaCountHelp,
		}, []string{metricsLabelStat},
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
	}

	for _, stat := range stats {
		metric.With(prometheus.Labels{metricsLabelStat: stat.label}).Set(float64(stat.value))
	}
}

// registerBuildInfoMetric registers a metric that provides build information.
func registerBuildInfoMetric(factory promauto.Factory, conf *config.ClientConfig) {
	factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      metricBuildInfo,
			Help:      metricsBuildInfoHelp,
		}, []string{metricsLabelEdition, metricsLabelGoVer, metricsLabelVersion},
	).With(prometheus.Labels{
		metricsLabelEdition: conf.Edition,
		metricsLabelGoVer:   runtime.Version(),
		metricsLabelVersion: conf.Version,
	}).Set(1.0)
}

// registerUsageMetrics registers filesystem and account usage metrics derived from the active configuration.
// Ratios follow Prometheus best practices (0..1) instead of percentages.
func registerUsageMetrics(factory promauto.Factory, usage config.Usage) {
	filesBytes := factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsUsageSubsystem,
			Name:      metricFilesBytes,
			Help:      metricsFilesBytesHelp,
		}, []string{metricsLabelState},
	)

	filesBytes.With(prometheus.Labels{metricsLabelState: "used"}).Set(float64(usage.FilesUsed))
	filesBytes.With(prometheus.Labels{metricsLabelState: "free"}).Set(float64(usage.FilesFree))
	filesBytes.With(prometheus.Labels{metricsLabelState: "total"}).Set(float64(usage.FilesTotal))

	filesRatio := factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsUsageSubsystem,
			Name:      metricFilesRatio,
			Help:      metricsFilesRatioHelp,
		}, []string{metricsLabelState},
	)

	filesUsed := usage.FilesUsedRatio()
	filesRatio.With(prometheus.Labels{metricsLabelState: "used"}).Set(filesUsed)
	filesRatio.With(prometheus.Labels{metricsLabelState: "free"}).Set(1 - filesUsed)

	accountsActive := factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsUsageSubsystem,
			Name:      metricAccountsActive,
			Help:      metricsAccountsHelp,
		}, []string{metricsLabelState},
	)

	accountsActive.With(prometheus.Labels{metricsLabelState: "users"}).Set(float64(usage.UsersActive))
	accountsActive.With(prometheus.Labels{metricsLabelState: "guests"}).Set(float64(usage.GuestsActive))

	accountsRatio := factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsUsageSubsystem,
			Name:      metricAccountsRatio,
			Help:      metricsAccountsRatioHelp,
		}, []string{metricsLabelState},
	)

	accountsUsed := usage.UsersUsedRatio()
	accountsRatio.With(prometheus.Labels{metricsLabelState: "used"}).Set(accountsUsed)
	accountsRatio.With(prometheus.Labels{metricsLabelState: "free"}).Set(1 - accountsUsed)
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
			Namespace: metricsNamespace,
			Subsystem: metricsClusterSubsystem,
			Name:      metricClusterNodes,
			Help:      metricsClusterNodesHelp,
		}, []string{metricsLabelRole},
	)

	for role, value := range counts {
		nodeMetric.With(prometheus.Labels{metricsLabelRole: role}).Set(float64(value))
	}

	infoMetric := factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsClusterSubsystem,
			Name:      metricClusterInfo,
			Help:      metricsClusterInfoHelp,
		}, []string{metricsLabelUUID, metricsLabelCIDR},
	)

	infoMetric.With(prometheus.Labels{
		metricsLabelUUID: conf.ClusterUUID(),
		metricsLabelCIDR: conf.ClusterCIDR(),
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
