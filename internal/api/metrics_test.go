package api

import (
	"net/http"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestGetMetrics(t *testing.T) {
	t.Run("ExposeCountStatistics", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetMetrics(router)

		resp := PerformRequestWithStream(app, "GET", "/api/v1/metrics")

		if resp.Code != http.StatusOK {
			t.Fatal(resp.Body.String())
		}

		body := resp.Body.String()
		floatPattern := `[-+]?\d+(?:\.\d+)?(?:e[-+]?\d+)?`
		stats := []string{
			"all",
			"photos",
			"media",
			"animated",
			"live",
			"videos",
			"audio",
			"documents",
			"albums",
			"private_albums",
			"folders",
			"private_folders",
			"files",
			"hidden",
			"favorites",
			"private",
			"people",
			"labels",
			"label_max_photos",
		}

		for _, stat := range stats {
			assert.Regexp(t, regexp.MustCompile(`photoprism_statistics_media_count{stat="`+stat+`"} `+floatPattern), body)
		}
	})
	t.Run("ExposeBuildInformation", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetMetrics(router)

		resp := PerformRequestWithStream(app, "GET", "/api/v1/metrics")

		if resp.Code != http.StatusOK {
			t.Fatal(resp.Body.String())
		}

		body := resp.Body.String()

		assert.Regexp(t, regexp.MustCompile(`photoprism_build_info{edition=".+",goversion=".+",version=".+"} 1`), body)
	})
	t.Run("ExposeUsageMetrics", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetMetrics(router)

		resp := PerformRequestWithStream(app, "GET", "/api/v1/metrics")

		if resp.Code != http.StatusOK {
			t.Fatal(resp.Body.String())
		}

		body := resp.Body.String()
		floatPattern := `[-+]?\d+(?:\.\d+)?(?:e[-+]?\d+)?`

		assert.Regexp(t, regexp.MustCompile(`photoprism_usage_files_bytes{state="used"} `+floatPattern), body)
		assert.Regexp(t, regexp.MustCompile(`photoprism_usage_files_bytes{state="free"} `+floatPattern), body)
		assert.Regexp(t, regexp.MustCompile(`photoprism_usage_files_bytes{state="total"} `+floatPattern), body)
		assert.Regexp(t, regexp.MustCompile(`photoprism_usage_files_ratio{state="used"} `+floatPattern), body)
		assert.Regexp(t, regexp.MustCompile(`photoprism_usage_files_ratio{state="free"} `+floatPattern), body)
		assert.Regexp(t, regexp.MustCompile(`photoprism_usage_accounts_active{state="users"} `+floatPattern), body)
		assert.Regexp(t, regexp.MustCompile(`photoprism_usage_accounts_active{state="guests"} `+floatPattern), body)
		assert.Regexp(t, regexp.MustCompile(`photoprism_usage_accounts_ratio{state="used"} `+floatPattern), body)
		assert.Regexp(t, regexp.MustCompile(`photoprism_usage_accounts_ratio{state="free"} `+floatPattern), body)

		t.Run("RatiosSumToOne", func(t *testing.T) {
			parseValue := func(pattern string) float64 {
				m := regexp.MustCompile(pattern).FindStringSubmatch(body)
				if len(m) != 2 {
					t.Fatalf("metric not found: %s", pattern)
				}
				v, err := strconv.ParseFloat(m[1], 64)
				if err != nil {
					t.Fatalf("parse metric: %v", err)
				}
				return v
			}

			filesUsed := parseValue(`photoprism_usage_files_ratio{state="used"} (` + floatPattern + `)`)
			filesFree := parseValue(`photoprism_usage_files_ratio{state="free"} (` + floatPattern + `)`)
			assert.InEpsilon(t, 1.0, filesUsed+filesFree, 0.01)

			accountsUsed := parseValue(`photoprism_usage_accounts_ratio{state="used"} (` + floatPattern + `)`)
			accountsFree := parseValue(`photoprism_usage_accounts_ratio{state="free"} (` + floatPattern + `)`)
			assert.InEpsilon(t, 1.0, accountsUsed+accountsFree, 0.01)
		})
	})
	t.Run("ExposeClusterMetricsForPortal", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)

		GetMetrics(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)

		nodeDefs := []struct {
			name string
			role string
		}{
			{"metrics-app-1", string(cluster.RoleInstance)},
			{"metrics-service-1", string(cluster.RoleService)},
		}

		var cleanupIDs []string
		for _, def := range nodeDefs {
			n := &reg.Node{Node: cluster.Node{Name: def.name, Role: def.role, UUID: rnd.UUIDv7()}}
			assert.NoError(t, regy.Put(n))
			cleanupIDs = append(cleanupIDs, n.UUID)
		}

		t.Cleanup(func() {
			for _, uuid := range cleanupIDs {
				_ = regy.Delete(uuid)
			}
		})

		resp := PerformRequestWithStream(app, "GET", "/api/v1/metrics")

		if resp.Code != http.StatusOK {
			t.Fatal(resp.Body.String())
		}

		body := resp.Body.String()
		floatPattern := `[-+]?\d+(?:\.\d+)?(?:e[-+]?\d+)?`
		assert.Regexp(t, regexp.MustCompile(`photoprism_cluster_nodes{role="total"} `+floatPattern), body)
		assert.Regexp(t, regexp.MustCompile(`photoprism_cluster_nodes{role="instance"} `+floatPattern), body)
		assert.Regexp(t, regexp.MustCompile(`photoprism_cluster_nodes{role="service"} `+floatPattern), body)
		infoPattern := `photoprism_cluster_info\{(?:cidr="[^"]*",[^}]*uuid="[^"]*"|uuid="[^"]*",[^}]*cidr="[^"]*")\} 1`
		assert.Regexp(t, regexp.MustCompile(infoPattern), body)
	})
	t.Run("HasPrometheusExpositionFormatAsContentType", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetMetrics(router)

		resp := PerformRequestWithStream(app, "GET", "/api/v1/metrics")
		if resp.Code != http.StatusOK {
			t.Fatal(resp.Body.String())
		}
		assert.Equal(t, header.ContentTypePrometheus, resp.Result().Header.Get("Content-Type"))
	})
}
