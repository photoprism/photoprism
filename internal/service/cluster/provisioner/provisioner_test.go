package provisioner

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestMain executes runTestMain returning it's results.  It is done this way so that defer can be used to cleanup.
func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) (code int) {
	// Init test logger.
	log := logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	// Remove temporary SQLite files before running the tests.
	fs.PurgeTestDbFiles(".", false)
	// Remove temporary SQLite files after running the tests.
	defer fs.PurgeTestDbFiles(".", false)

	caller := "internal/service/cluster/provisioner/provisioner_test.go/TestMain"
	dbc, dbn, err := testextras.AcquireDBMutex(log, caller)
	if err != nil {
		log.Error("FAIL")
		return 1
	}
	defer testextras.UnlockDBMutex(dbc.Db())

	_, dsname := dsn.PhotoPrismTestToDriverDSN(dbn)
	dsn.SetDSNToEnv(dsname)

	// Run unit tests.
	code = m.Run()

	testextras.ReleaseDBMutex(dbc.Db(), log, caller, code)

	return code
}

func cleanupDB(t *testing.T, ctx context.Context, creds Credentials) {
	// Cleanup: drop user and database to keep the dev DB tidy.
	adb, err := GetDB(ctx)
	if err != nil {
		t.Fatalf("GetDB: %v", err)
	}
	qdb, err := quoteIdent(creds.Name)
	if err != nil {
		t.Fatalf("quoteIdent: %v", err)
	}
	acc, err := quoteAccount("%", creds.User)
	if err != nil {
		t.Fatalf("quoteAccount: %v", err)
	}
	// Best-effort cleanup; ignore individual errors to avoid masking earlier failures.
	_ = execTimeout(ctx, adb, 10*time.Second, "REVOKE ALL PRIVILEGES, GRANT OPTION FROM "+acc)
	_ = execTimeout(ctx, adb, 10*time.Second, "DROP USER IF EXISTS "+acc)
	_ = execTimeout(ctx, adb, 15*time.Second, "DROP DATABASE IF EXISTS "+qdb)
}
