package provisioner

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestEnsureCredentials_MariaDB exercises the direct mysql driver path using the
// ProvisionDSN. It skips if MariaDB is not reachable or when not explicitly enabled
// via environment (PHOTOPRISM_TEST_DRIVER=mysql).
func TestEnsureCredentials_MariaDB(t *testing.T) {
	ctx := context.Background()

	// Quick liveness probe for AdminDsn; skip fast if not reachable.
	if db, err := sql.Open("mysql", ProvisionDSN); err != nil {
		t.Skipf("admin DSN not openable: %v", err)
	} else {
		c, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		if err := db.PingContext(c); err != nil {
			_ = db.Close()
			t.Skipf("admin DSN not reachable: %v", err)
		}
		_ = db.Close()
	}

	// A random ClusterUUID gives each run its own database and user names.
	c := config.NewConfig(config.CliTestContext())
	c.Options().ClusterUUID = rnd.UUID()

	nodeName := "pp-itest-node"

	// 1st call: rotate=true so we receive a password + DSN.
	creds, created, err := EnsureCredentials(ctx, c, "11111111-1111-4111-8111-111111111111", nodeName, true)
	if err != nil {
		t.Fatalf("EnsureCredentials(rotate=true) error: %v", err)
	}

	// Ensure we always drop the temporary DB/user created during this test.
	t.Cleanup(func() {
		if creds.Name == "" || creds.User == "" {
			return
		}
		if dropErr := DropCredentials(ctx, creds.Name, creds.User); dropErr != nil {
			t.Logf("cleanup: %v", dropErr)
		}
	})

	if creds.Name == "" || creds.User == "" {
		t.Fatalf("missing db name/user in creds: %+v", creds)
	}
	if creds.Password == "" || creds.DSN == "" {
		t.Fatalf("expected password and DSN on rotate/create; got: %+v (created=%v)", creds, created)
	}

	// DSN should be usable by the node user (at least ping).
	udb, err := sql.Open("mysql", creds.DSN)
	if err != nil {
		t.Fatalf("open node DSN: %v", err)
	}
	c2, cancel := context.WithTimeout(ctx, 5*time.Second)
	if err := udb.PingContext(c2); err != nil {
		cancel()
		_ = udb.Close()
		t.Fatalf("ping node DSN: %v", err)
	}
	cancel()
	_ = udb.Close()

	// 2nd call: rotate=false should not return a password (idempotent ensure).
	creds2, _, err := EnsureCredentials(ctx, c, "11111111-1111-4111-8111-111111111111", nodeName, false)
	if err != nil {
		t.Fatalf("EnsureCredentials(rotate=false) error: %v", err)
	}
	if creds2.Password != "" || creds2.DSN != "" {
		t.Fatalf("expected no password/DSN without rotation; got: %+v", creds2)
	}
}

// limitedAdminDSN creates an admin account that may create databases and users but lacks DROP,
// and returns a DSN for it. The account is removed when the test ends.
func limitedAdminDSN(t *testing.T, ctx context.Context) string {
	t.Helper()

	db, err := GetDB(ctx)
	if err != nil {
		t.Skipf("provisioning database is unavailable: %v", err)
	}

	const user = "pp_provision_limited"
	pass := rnd.Base62(24)

	for _, stmt := range []string{
		"DROP USER IF EXISTS '" + user + "'@'%'",
		"CREATE USER '" + user + "'@'%' IDENTIFIED BY '" + pass + "'",
		"GRANT CREATE, CREATE USER ON *.* TO '" + user + "'@'%' WITH GRANT OPTION",
	} {
		_, execErr := db.ExecContext(ctx, stmt)
		require.NoError(t, execErr, stmt)
	}

	t.Cleanup(func() {
		if root, rootErr := GetDB(context.Background()); rootErr == nil {
			_, _ = root.ExecContext(context.Background(), "DROP USER IF EXISTS '"+user+"'@'%'")
		}
	})

	cfg, err := mysql.ParseDSN(ProvisionDSN)
	require.NoError(t, err)
	cfg.User, cfg.Passwd = user, pass

	return cfg.FormatDSN()
}

// TestEnsureCredentials_Incomplete checks that a failed grant reports the created database, and that
// a later call with sufficient privileges completes it.
func TestEnsureCredentials_Incomplete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	rootDSN := ProvisionDSN
	t.Cleanup(func() { ProvisionDSN = rootDSN })
	limitedDSN := limitedAdminDSN(t, ctx)

	c := config.NewConfig(config.CliTestContext())
	c.Options().ClusterUUID = rnd.UUID()
	nodeUUID, nodeName := "11111111-1111-4111-8111-555555555555", "pp-incomplete"
	dbName, dbUser, _ := GenerateCredentials(c, nodeUUID, nodeName)

	t.Cleanup(func() {
		ProvisionDSN = rootDSN
		if dropErr := DropCredentials(context.Background(), dbName, dbUser); dropErr != nil {
			t.Errorf("cleanup: %v", dropErr)
		}
	})

	ProvisionDSN = limitedDSN
	creds, created, err := EnsureCredentials(ctx, c, nodeUUID, nodeName, true)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrIncomplete)
	assert.True(t, created)
	assert.Equal(t, dbName, creds.Name)
	assert.Equal(t, dbUser, creds.User)
	assert.Empty(t, creds.Password)
	assert.Empty(t, creds.DSN)
	assert.Empty(t, creds.RotatedAt)

	ProvisionDSN = rootDSN
	creds, created, err = EnsureCredentials(ctx, c, nodeUUID, nodeName, true)

	require.NoError(t, err)
	assert.False(t, created)
	assert.NotEmpty(t, creds.Password)

	udb, err := sql.Open("mysql", creds.DSN)
	require.NoError(t, err)
	defer udb.Close()
	assert.NoError(t, udb.PingContext(ctx))
}

// TestEnsureCredentials_DriverNormalization verifies driver normalization and rejections.
func TestEnsureCredentials_DriverNormalization(t *testing.T) {
	orig := DatabaseDriver
	t.Cleanup(func() { DatabaseDriver = orig })

	c := config.NewConfig(config.CliTestContext())
	ctx := context.Background()

	// Postgres in weird case should hit the explicit rejection path.
	DatabaseDriver = "PostGreS"
	_, _, err := EnsureCredentials(ctx, c, "11111111-1111-4111-8111-111111111111", "pp-node", false)
	assert.ErrorIs(t, err, ErrUnsupportedDriver)
	assert.Equal(t, "PostGreS", DatabaseDriver)

	// Unknown driver should return the unsupported error including normalized name.
	DatabaseDriver = "TiDB"
	_, _, err = EnsureCredentials(ctx, c, "11111111-1111-4111-8111-111111111111", "pp-node", false)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "unsupported auto-provisioning database driver: tidb")
		assert.ErrorIs(t, err, ErrUnsupportedDriver)
	}
	assert.Equal(t, "TiDB", DatabaseDriver)
}

// TestDropCredentials_Repeated checks repeated cleanup of absent database credentials.
func TestDropCredentials_Repeated(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	db, err := GetDB(ctx)
	if err != nil {
		t.Skip("provisioning database is unavailable")
	}

	dbName := "test_d" + rnd.GenerateUID('c')
	dbUser := "test_u" + rnd.GenerateUID('c')

	// Check absence on the same provisioning connection used by cleanup.
	var count int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = ?", dbName).Scan(&count))
	require.Zero(t, count, "test database name must be unused")
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mysql.user WHERE User = ?", dbUser).Scan(&count))
	require.Zero(t, count, "test account name must be unused")

	assert.NoError(t, DropCredentials(ctx, dbName, dbUser), "absent credentials must drop cleanly")
	assert.NoError(t, DropCredentials(ctx, dbName, dbUser), "a repeated drop must stay clean")
	assert.NoError(t, DropCredentials(ctx, "", dbUser), "an absent user alone must drop cleanly")
	assert.NoError(t, DropCredentials(ctx, dbName, ""), "an absent database alone must drop cleanly")
}

// TestCredentials_ProxyError checks the error of a ProxySQL step as the provisioner reports it.
func TestCredentials_ProxyError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if _, err := GetDB(ctx); err != nil {
		t.Skip("provisioning database is unavailable")
	}

	c := config.NewConfig(config.CliTestContext())
	c.Options().ClusterUUID = rnd.UUID()
	nodeUUID, nodeName := "11111111-1111-4111-8111-444444444444", "pp-proxy-error"
	dbName, dbUser, _ := GenerateCredentials(c, nodeUUID, nodeName)

	// Cleanups run after the test context is canceled, so the drop uses its own.
	t.Cleanup(func() {
		dropCtx, dropCancel := context.WithTimeout(context.Background(), time.Minute)
		defer dropCancel()
		if dropErr := DropCredentials(dropCtx, dbName, dbUser); dropErr != nil {
			t.Errorf("cleanup: %v", dropErr)
		}
	})

	origDSN := ProvisionProxyDSN
	ProvisionProxyDSN = "admin:pa?ss=word@tcp(127.0.0.1:6032)"
	t.Cleanup(func() { ProvisionProxyDSN = origDSN })

	creds, _, err := EnsureCredentials(ctx, c, nodeUUID, nodeName, true)
	assert.EqualError(t, err, "database provisioning incomplete: proxysql: invalid admin dsn")
	assert.ErrorIs(t, err, ErrIncomplete)
	assert.Equal(t, dbName, creds.Name)
	assert.Equal(t, dbUser, creds.User)
	assert.Empty(t, creds.Password)
	assert.Empty(t, creds.DSN)
	assert.EqualError(t, DropCredentials(ctx, "", dbUser), "drop credentials: proxysql: invalid admin dsn")
}

func TestCheckConfig(t *testing.T) {
	driver, dsnValue := DatabaseDriver, ProvisionDSN
	t.Cleanup(func() { DatabaseDriver, ProvisionDSN = driver, dsnValue })

	t.Run("Ready", func(t *testing.T) {
		DatabaseDriver, ProvisionDSN = "MariaDB", "root:secret@tcp(127.0.0.1:1)/photoprism"
		assert.NoError(t, CheckConfig())
	})
	t.Run("UnsupportedDriver", func(t *testing.T) {
		DatabaseDriver, ProvisionDSN = "sqlite3", "root:secret@tcp(127.0.0.1:1)/photoprism"
		assert.ErrorIs(t, CheckConfig(), ErrUnsupportedDriver)
	})
	t.Run("NotConfigured", func(t *testing.T) {
		DatabaseDriver, ProvisionDSN = "mysql", ""
		assert.ErrorIs(t, CheckConfig(), ErrNotConfigured)
	})
}

func TestProvisionDriver(t *testing.T) {
	driver := DatabaseDriver
	t.Cleanup(func() { DatabaseDriver = driver })

	for value, expected := range map[string]string{"MySQL": "mysql", "mariadb": "mariadb"} {
		DatabaseDriver = value
		got, err := provisionDriver()
		assert.NoError(t, err, value)
		assert.Equal(t, expected, got)
	}

	for _, value := range []string{"postgres", "sqlite3", "oracle", ""} {
		DatabaseDriver = value
		_, err := provisionDriver()
		assert.ErrorIs(t, err, ErrUnsupportedDriver, value)
	}
}

func TestSchemaExists(t *testing.T) {
	db, err := GetDB(context.Background())
	if err != nil {
		t.Skipf("provisioning database is unavailable: %v", err)
	}

	assert.True(t, schemaExists(db, "information_schema"))
	assert.False(t, schemaExists(db, "cluster_d"+strings.ToLower(rnd.Base36(11))))
}
