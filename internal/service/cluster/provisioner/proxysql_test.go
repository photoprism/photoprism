package provisioner

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
)

func TestEnsureCredentials_ProxySQLIntegration(t *testing.T) {
	if os.Getenv("PHOTOPRISM_TEST_PROXYSQL") == "" {
		t.Skip("PHOTOPRISM_TEST_PROXYSQL not set; skipping ProxySQL integration test")
	}

	ctx := context.Background()

	proxyDSN := os.Getenv("PHOTOPRISM_TEST_PROXYSQL_DSN")
	if proxyDSN == "" {
		proxyDSN = "admin:admin@tcp(127.0.0.1:6032)/"
	}

	adminDsn, err := normalizeProxyDSN(proxyDSN)
	require.NoError(t, err)

	adminDB, err := sql.Open("mysql", adminDsn)
	if err != nil {
		t.Skipf("proxy DSN not openable: %v", err)
	}
	t.Cleanup(func() { _ = adminDB.Close() })

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	if err := adminDB.PingContext(pingCtx); err != nil {
		cancel()
		t.Skipf("proxy DSN not reachable: %v", err)
	}
	cancel()

	origDSN := ProvisionProxyDSN
	origOpts := ProvisionProxyOptions
	ProvisionProxyDSN = proxyDSN
	ProvisionProxyOptions = ProxyOptions{
		Hostgroup:      DefaultProxyHostgroup,
		Frontend:       DefaultProxyFrontend,
		Backend:        DefaultProxyBackend,
		MaxConnections: DefaultProxyMaxConnections,
		UseSSL:         DefaultProxyUseSSL,
		Comment:        DefaultProxyComment,
	}
	t.Cleanup(func() {
		ProvisionProxyDSN = origDSN
		ProvisionProxyOptions = origOpts
	})

	conf := config.NewConfig(config.CliTestContext())
	conf.Options().ClusterUUID = time.Now().UTC().Format("20060102-150405.000000000")

	nodeUUID := "11111111-1111-4111-8111-333333333333"
	nodeName := "pp-proxy-itest"

	creds, _, err := EnsureCredentials(ctx, conf, nodeUUID, nodeName, true)
	if err != nil {
		t.Fatalf("EnsureCredentials with ProxySQL error: %v", err)
	}

	t.Cleanup(func() {
		if creds.Name == "" || creds.User == "" {
			return
		}
		if dropErr := DropCredentials(ctx, creds.Name, creds.User); dropErr != nil {
			t.Logf("cleanup drop credentials: %v", dropErr)
		}
	})

	var defaultSchema, comment string
	var frontend, backend, maxConnections int

	err = adminDB.QueryRowContext(ctx, `
		SELECT default_schema, frontend, backend, max_connections, comment
		  FROM mysql_users WHERE username = ?
	`, creds.User).Scan(&defaultSchema, &frontend, &backend, &maxConnections, &comment)
	if err != nil {
		t.Fatalf("mysql_users lookup: %v", err)
	}

	if defaultSchema != creds.Name {
		t.Fatalf("expected default_schema %q, got %q", creds.Name, defaultSchema)
	}
	if frontend != ProvisionProxyOptions.Frontend {
		t.Fatalf("expected frontend=%d, got %d", ProvisionProxyOptions.Frontend, frontend)
	}
	if backend != ProvisionProxyOptions.Backend {
		t.Fatalf("expected backend=%d, got %d", ProvisionProxyOptions.Backend, backend)
	}
	if maxConnections != ProvisionProxyOptions.MaxConnections {
		t.Fatalf("expected max_connections=%d, got %d", ProvisionProxyOptions.MaxConnections, maxConnections)
	}
	if comment != ProvisionProxyOptions.Comment {
		t.Fatalf("expected comment %q, got %q", ProvisionProxyOptions.Comment, comment)
	}

	if _, _, err := EnsureCredentials(ctx, conf, nodeUUID, nodeName, false); err != nil {
		t.Fatalf("EnsureCredentials (rotate=false) error: %v", err)
	}

	var count int
	if err := adminDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM mysql_users WHERE username = ?", creds.User).Scan(&count); err != nil {
		t.Fatalf("count mysql_users: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected mysql_users count 1 after idempotent ensure, got %d", count)
	}

	nodeUser := creds.User
	nodeSchema := creds.Name

	if err := DropCredentials(ctx, nodeSchema, nodeUser); err != nil {
		t.Fatalf("DropCredentials error: %v", err)
	}
	creds.Name, creds.User = "", ""

	if err := adminDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM mysql_users WHERE username = ?", nodeUser).Scan(&count); err != nil {
		t.Fatalf("count mysql_users after drop: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected mysql_users count 0 after drop, got %d", count)
	}
}

// TestNormalizeProxyDSN checks the parameters added to, dropped from, and required of a ProxySQL admin DSN.
func TestNormalizeProxyDSN(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for in, want := range map[string]string{
			"admin:admin@tcp(127.0.0.1:6032)/":                                                                           "admin:admin@tcp(127.0.0.1:6032)/?charset=utf8mb4&interpolateParams=true",
			"admin:admin@tcp(127.0.0.1:6032)/?charset=utf8mb4":                                                           "admin:admin@tcp(127.0.0.1:6032)/?charset=utf8mb4&interpolateParams=true",
			"admin:admin@tcp(127.0.0.1:6032)/?interpolateParams=false":                                                   "admin:admin@tcp(127.0.0.1:6032)/?interpolateParams=false&charset=utf8mb4",
			"admin:admin@tcp(127.0.0.1:6032)/?collation=utf8mb4_unicode_ci&":                                             "admin:admin@tcp(127.0.0.1:6032)/?collation=utf8mb4_unicode_ci&interpolateParams=true",
			"admin:p?w@tcp(127.0.0.1:6032)/":                                                                             "admin:p?w@tcp(127.0.0.1:6032)/?charset=utf8mb4&interpolateParams=true",
			"admin:interpolateParams=x@tcp(127.0.0.1:6032)/?charset=utf8mb4":                                             "admin:interpolateParams=x@tcp(127.0.0.1:6032)/?charset=utf8mb4&interpolateParams=true",
			"admin:admin@tcp(127.0.0.1:6032)/?":                                                                          "admin:admin@tcp(127.0.0.1:6032)/?charset=utf8mb4&interpolateParams=true",
			"admin:admin@tcp(127.0.0.1:6032)/?parseTime=true&timeout=15s":                                                "admin:admin@tcp(127.0.0.1:6032)/?parseTime=true&timeout=15s&charset=utf8mb4&interpolateParams=true",
			"admin:admin@tcp(127.0.0.1:6032)/main?timeout=1s":                                                            "admin:admin@tcp(127.0.0.1:6032)/main?timeout=1s&charset=utf8mb4&interpolateParams=true",
			"admin:admin@tcp(127.0.0.1:6032)/main":                                                                       "admin:admin@tcp(127.0.0.1:6032)/main?charset=utf8mb4&interpolateParams=true",
			"admin:admin@tcp(127.0.0.1:6032)/?tls=skip-verify&maxAllowedPacket=4194304&readTimeout=30s&writeTimeout=30s": "admin:admin@tcp(127.0.0.1:6032)/?tls=skip-verify&maxAllowedPacket=4194304&readTimeout=30s&writeTimeout=30s&charset=utf8mb4&interpolateParams=true",
			"": "",
		} {
			got, err := normalizeProxyDSN(in)
			require.NoError(t, err, in)
			assert.Equal(t, want, got, in)
		}
	})
	t.Run("Dropped", func(t *testing.T) {
		// Parameters without a rule are dropped.
		for in, want := range map[string]string{
			"admin:admin@tcp(127.0.0.1:6032)/?wait_timeout=28800,sql_mode=ANSI":                           "admin:admin@tcp(127.0.0.1:6032)/?charset=utf8mb4&interpolateParams=true",
			"admin:admin@tcp(127.0.0.1:6032)/?wait_timeout=28800%2Csql_mode%3DANSI&charset=utf8mb4":       "admin:admin@tcp(127.0.0.1:6032)/?charset=utf8mb4&interpolateParams=true",
			"admin:admin@tcp(127.0.0.1:6032)/?sql_mode=ANSI,time_zone=UTC&collation=utf8mb4_bin":          "admin:admin@tcp(127.0.0.1:6032)/?collation=utf8mb4_bin&interpolateParams=true",
			"admin:admin@tcp(127.0.0.1:6032)/?option%20x,other=1":                                         "admin:admin@tcp(127.0.0.1:6032)/?charset=utf8mb4&interpolateParams=true",
			"admin:admin@tcp(127.0.0.1:6032)/?time_zone=UTC&allowAllFiles=true&stray&interpolateParams=1": "admin:admin@tcp(127.0.0.1:6032)/?interpolateParams=1&charset=utf8mb4",
		} {
			got, err := normalizeProxyDSN(in)
			require.NoError(t, err, in)
			assert.Equal(t, want, got, in)
		}
	})
	t.Run("DroppedLogged", func(t *testing.T) {
		logger, hook := logtest.NewNullLogger()
		prev := log
		log = logger
		t.Cleanup(func() { log = prev })
		_, err := normalizeProxyDSN("admin:admin@tcp(127.0.0.1:6032)/?wait_timeout=28800,sql_mode=ANSI&x%20y=secret&stray")
		require.NoError(t, err)
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, "proxysql: ignored 3 unsupported admin dsn parameters wait_timeout", hook.LastEntry().Message)
	})
	t.Run("InvalidDSN", func(t *testing.T) {
		logger, hook := logtest.NewNullLogger()
		prev := log
		log = logger
		t.Cleanup(func() { log = prev })
		for _, in := range []string{"admin:pa?ss=word@tcp(127.0.0.1:6032)", "admin:p/a?ss=word@tcp(127.0.0.1:6032)",
			"admin:admin@tcp(127.0.0.1:6032)/?tls=custom", "admin:admin@tcp(127.0.0.1:6032)/?strict=1"} {
			_, err := normalizeProxyDSN(in)
			assert.EqualError(t, err, "proxysql: invalid admin dsn", in)
		}
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("InvalidParam", func(t *testing.T) {
		for in, want := range map[string]string{
			"admin:admin@tcp(127.0.0.1:6032)/?charset=latin1":                                     "proxysql: invalid dsn parameter charset",
			"admin:admin@tcp(127.0.0.1:6032)/?collation=latin1_swedish_ci&interpolateParams=true": "proxysql: invalid dsn parameter collation",
			"admin:admin@tcp(127.0.0.1:6032)/?charset=utf8mb4&charset=latin1":                     "proxysql: duplicate dsn parameter charset",
		} {
			_, err := normalizeProxyDSN(in)
			assert.EqualError(t, err, want, in)
		}
	})
}

// TestValidProxyDSN checks which DSNs the MySQL driver can parse.
func TestValidProxyDSN(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.True(t, validProxyDSN("admin:admin@tcp(127.0.0.1:6032)/?charset=utf8mb4"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.False(t, validProxyDSN("admin:pa?ss=word@tcp(127.0.0.1:6032)"))
		assert.False(t, validProxyDSN("admin:admin@tcp(127.0.0.1:6032)/?strict=1"))
	})
}
