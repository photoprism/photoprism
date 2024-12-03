package performancetest

import (
	"os"
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
)

func TestMigrateDatabase(t *testing.T) {

	t.Run("Ok", func(t *testing.T) {

		driver := os.Getenv("PHOTOPRISM_TEST_DRIVER")
		dsn := os.Getenv("PHOTOPRISM_TEST_DSN")

		db := &entity.DbConn{
			Driver: driver,
			Dsn:    dsn,
		}
		defer db.Close()

		entity.SetDbProvider(db)

		beforeTimestamp := time.Now().UTC()

		entity.InitDb(migrate.Opt(true, false, nil))

		afterTimestamp := time.Now().UTC()

		log.Infof("Migration Started At %v Ended At %v", beforeTimestamp, afterTimestamp)
		log.Infof("Migration Duration %s", time.Since(beforeTimestamp))

	})

}
