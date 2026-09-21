package provisioner

import (
	"context"
	"database/sql"
	"testing"
	"time"

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

	// Unique-ish ClusterUUID to avoid collisions across runs; format is not strictly validated.
	c := config.NewConfig(config.CliTestContext())
	c.Options().ClusterUUID = time.Now().UTC().Format("20060102-150405.000000000")

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

// TestEnsureCredentials_DriverNormalization verifies driver normalization and rejections.
func TestEnsureCredentials_DriverNormalization(t *testing.T) {
	orig := DatabaseDriver
	t.Cleanup(func() { DatabaseDriver = orig })

	c := config.NewConfig(config.CliTestContext())
	ctx := context.Background()

	// Postgres in weird case should hit the explicit rejection path.
	DatabaseDriver = "PostGreS"
	_, _, err := EnsureCredentials(ctx, c, "11111111-1111-4111-8111-111111111111", "pp-node", false)
	assert.Error(t, err)
	assert.Equal(t, "PostGreS", DatabaseDriver)

	// Unknown driver should return the unsupported error including normalized name.
	DatabaseDriver = "TiDB"
	_, _, err = EnsureCredentials(ctx, c, "11111111-1111-4111-8111-111111111111", "pp-node", false)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "unsupported auto-provisioning database driver: tidb")
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
