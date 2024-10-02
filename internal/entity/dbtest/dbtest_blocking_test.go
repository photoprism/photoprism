package entity

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Test Table to be blocked
type Blocker struct {
	ID        int
	PhotoUID  string
	OtherData string
	CreatedAt time.Time
	UpdatedAt time.Time `gorm:"autoUpdateTime:true"`
}

// Migrate the Blocker table into the database
func migrateTestBlocker(db *gorm.DB) {
	if err := db.AutoMigrate(&Blocker{}); err != nil {
		panic(err)
	}
}

// Function to hold a lock on the Blocker table for 15 seconds.  Hopefully long enough.
func lock_blocker_for_test(t *testing.T, m interface{}, keyNames ...string) {

	// Extract interface slice with all values including zero.
	values, keys, err := entity.ModelValues(m, keyNames...)

	// Has keys and values?
	if err != nil {
		t.Logf("lock_blocker_for_test ModelValues Error = %s", err.Error())
		return
	} else if len(keys) != len(keyNames) {
		t.Logf("lock_blocker_for_test keys issue")
		return
	}

	db_test := entity.UnscopedDb()
	db_test.Transaction(func(tx *gorm.DB) error {
		if tx.Error != nil {
			t.Logf("lock_blocker_for_test Begin = %s", tx.Error.Error())
			return tx.Error
		}

		result := tx.Model(m).Updates(values)

		if result.Error != nil {
			t.Logf("lock_blocker_for_test Model Updates Error = %s", result.Error.Error())
			return result.Error
		}

		time.Sleep(30 * time.Second)
		return nil
	})

	t.Logf("lock_blocker_for_test Rollback Done")
}

// This test locks a record in a newly created database table and tests that an error is
// captured by entity.Update.
func TestEntity_UpdateDBErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	dbtestMutex.Lock()
	defer dbtestMutex.Unlock()

	migrateTestBlocker(entity.UnscopedDb())

	if entity.DbDialect() == entity.MySQL {
		entity.Db().Exec("SET GLOBAL lock_wait_timeout=5;")
		entity.Db().Exec("SET GLOBAL innodb_lock_wait_timeout=5;")
	}

	startTime := time.Now()

	t.Run("LockedDB", func(t *testing.T) {
		m := &Blocker{ID: 100, PhotoUID: "ATestString", OtherData: "Another Test String", CreatedAt: time.Now()}
		updatedAt := m.UpdatedAt
		entity.Db().Create(m)

		// Should be updated without any issues.
		if err := entity.Update(m, "ID", "PhotoUID"); err != nil {
			assert.Greater(t, m.UpdatedAt.UTC(), updatedAt.UTC())
			t.Fatal(err)
			return
		} else {
			assert.Greater(t, m.UpdatedAt.UTC(), updatedAt.UTC())
			t.Logf("(1) UpdatedAt: %s -> %s", updatedAt.UTC(), m.UpdatedAt.UTC())
			t.Logf("(1) Successfully updated values")
		}

		go lock_blocker_for_test(t, m, "ID", "PhotoUID")
		// Wait a bit for the other thread to start waiting.
		time.Sleep(time.Second * 2)

		// Should return an error with lock timeout.
		if err := entity.Update(m, "ID", "PhotoUID"); err != nil {
			assert.Greater(t, m.UpdatedAt.UTC(), updatedAt.UTC())
			if entity.DbDialect() == entity.SQLite3 {
				assert.ErrorContains(t, err, "no such table: blockers") // Sql Lite doesn't have wait locking it just throws a missing table message.
			} else {
				assert.ErrorContains(t, err, "timeout")
			}
			t.Logf("(2) Error was %s", err.Error())
			return
		} else {
			t.Logf("(2) Error not found")
			t.Fail()
		}
	})

	if entity.DbDialect() == entity.MySQL {
		entity.Db().Exec("SET GLOBAL lock_wait_timeout=DEFAULT;")
		entity.Db().Exec("SET GLOBAL innodb_lock_wait_timeout=DEFAULT;")
	}
	// Need to sleep here waiting for the child process to end.
	timeLeft := time.Duration(time.Second*32) - time.Now().Sub(startTime)
	if timeLeft > 0 {
		time.Sleep(timeLeft)
	}
}

// This test locks a record in a newly created database table and tests that an error is
// captured by entity.Save.
func TestEntity_SaveDBErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	dbtestMutex.Lock()
	defer dbtestMutex.Unlock()

	migrateTestBlocker(entity.UnscopedDb())

	if entity.DbDialect() == entity.MySQL {
		entity.Db().Exec("SET GLOBAL lock_wait_timeout=5;")
		entity.Db().Exec("SET GLOBAL innodb_lock_wait_timeout=5;")
	}

	startTime := time.Now()

	t.Run("LockedDB", func(t *testing.T) {
		m := &Blocker{ID: 100, PhotoUID: "ATestString", OtherData: "Another Test String", CreatedAt: time.Now()}
		updatedAt := m.UpdatedAt
		entity.Db().Create(m)

		// Should be updated without any issues.
		if err := entity.Save(m, "ID", "PhotoUID"); err != nil {
			assert.Greater(t, m.UpdatedAt.UTC(), updatedAt.UTC())
			t.Fatal(err)
			return
		} else {
			assert.Greater(t, m.UpdatedAt.UTC(), updatedAt.UTC())
			t.Logf("(1) UpdatedAt: %s -> %s", updatedAt.UTC(), m.UpdatedAt.UTC())
			t.Logf("(1) Successfully updated values")
		}

		go lock_blocker_for_test(t, m, "ID", "PhotoUID")
		// Wait a bit for the other thread to start waiting.
		time.Sleep(time.Second * 2)

		// Should return an error with lock timeout.
		if err := entity.Update(m, "ID", "PhotoUID"); err != nil {
			assert.Greater(t, m.UpdatedAt.UTC(), updatedAt.UTC())
			if entity.DbDialect() == entity.SQLite3 {
				assert.ErrorContains(t, err, "no such table: blockers") // Sql Lite doesn't have wait locking it just throws a missing table message.
			} else {
				assert.ErrorContains(t, err, "timeout")
			}
			t.Logf("(2) Error was %s", err.Error())
			return
		} else {
			t.Logf("(2) Error not found")
			t.Fail()
		}
	})

	if entity.DbDialect() == entity.MySQL {
		entity.Db().Exec("SET GLOBAL lock_wait_timeout=DEFAULT;")
		entity.Db().Exec("SET GLOBAL innodb_lock_wait_timeout=DEFAULT;")
	}
	// Need to sleep here waiting for the child process to end.
	timeLeft := time.Duration(time.Second*32) - time.Now().Sub(startTime)
	if timeLeft > 0 {
		time.Sleep(timeLeft)
	}
}
