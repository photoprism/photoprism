package entity

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestUpdateLabelCounts_Retry checks database recovery and the successful-refresh timestamp.
func TestUpdateLabelCounts_Retry(t *testing.T) {
	if !IsDialect(dsn.DriverMySQL) {
		t.Skip("requires MySQL trigger support")
	}
	WaitForAsyncJobs()
	t.Cleanup(WaitForAsyncJobs)
	// A dedicated pool keeps the trigger's session counter stable without changing the shared pool.
	original := dbConn
	conn := &DbConn{Driver: dsn.DriverMySQL, Dsn: TestDbDSN(dsn.DriverMySQL, os.Getenv("PHOTOPRISM_TEST_DSN"))}
	conn.Db().DB().SetMaxOpenConns(1)
	SetDbProvider(conn)
	t.Cleanup(func() { WaitForAsyncJobs(); SetDbProvider(original); conn.Close() })
	previousStamp := updateLabelCountsLastUpdated.Load()
	t.Cleanup(func() { updateLabelCountsLastUpdated.Store(previousStamp) })
	for _, mode := range []string{"Recover", "Exhausted", "OtherError"} {
		t.Run(mode, func(t *testing.T) {
			trigger := "test_label_counts_" + rnd.GenerateUID('l')
			number := 1213
			if mode == "OtherError" {
				number = 1644
			}
			// The session counter survives statement rollback and distinguishes retry attempts.
			condition := "@label_count_attempts = 1"
			if mode == "Exhausted" {
				condition = "1 = 1"
			}
			create := "CREATE TRIGGER " + trigger + " BEFORE UPDATE ON labels FOR EACH ROW BEGIN SET @label_count_attempts = COALESCE(@label_count_attempts, 0) + 1; IF " + condition + " THEN SIGNAL SQLSTATE '40001' SET MYSQL_ERRNO = " + fmt.Sprint(number) + ", MESSAGE_TEXT = 'count write control'; END IF; END"
			require.NoError(t, UnscopedDb().Exec(create).Error)
			t.Cleanup(func() { require.NoError(t, UnscopedDb().Exec("DROP TRIGGER "+trigger).Error) })
			require.NoError(t, UnscopedDb().Exec("SET @label_count_attempts = 0").Error)
			updateLabelCountsLastUpdated.Store(123)
			err := UpdateLabelCounts()
			var attempts int
			require.NoError(t, UnscopedDb().Raw("SELECT @label_count_attempts").Row().Scan(&attempts))
			switch mode {
			case "Recover":
				require.NoError(t, err)
				assert.Greater(t, attempts, 1)
				assert.Greater(t, updateLabelCountsLastUpdated.Load(), int64(123))
			case "Exhausted":
				require.Error(t, err)
				assert.Equal(t, 3, attempts)
				assert.Equal(t, int64(123), updateLabelCountsLastUpdated.Load())
			case "OtherError":
				require.Error(t, err)
				assert.Equal(t, 1, attempts)
				assert.Equal(t, int64(123), updateLabelCountsLastUpdated.Load())
			}
		})
	}
}
