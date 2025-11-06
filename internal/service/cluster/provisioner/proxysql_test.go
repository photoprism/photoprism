package provisioner

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

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

	adminDB, err := sql.Open("mysql", normalizeProxyDSN(proxyDSN))
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
