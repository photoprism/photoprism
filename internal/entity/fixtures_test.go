package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixtureDb(t *testing.T) {
	t.Run("NoTransaction", func(t *testing.T) {
		assert.Nil(t, fixtureTx)
		assert.Equal(t, Db(), fixtureDb())
	})
	t.Run("InTransaction", func(t *testing.T) {
		done := beginFixtureTx()
		assert.NotNil(t, fixtureTx)
		assert.NotEqual(t, Db(), fixtureDb(), "fixture inserts must not run on the shared connection")
		done()
		assert.Nil(t, fixtureTx)
		assert.Equal(t, Db(), fixtureDb(), "the shared connection must be restored after commit")
	})
}

func TestBeginFixtureTx(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		done := beginFixtureTx()
		require.NotNil(t, done)
		require.NotNil(t, fixtureTx)
		done()
		assert.Nil(t, fixtureTx)
	})
	t.Run("SharedProviderIsNeverSwapped", func(t *testing.T) {
		// The background workers read Db() concurrently, so it must keep pointing at the
		// pooled connection while a fixture transaction is open.
		before := Db()
		done := beginFixtureTx()
		assert.Equal(t, before, Db())
		done()
		assert.Equal(t, before, Db())
	})
	t.Run("Nested", func(t *testing.T) {
		outer := beginFixtureTx()
		tx := fixtureTx
		require.NotNil(t, tx)
		// The inner call cannot start a transaction on an open one, so it is a no-op and
		// leaves the outer transaction in place for the inner inserts.
		inner := beginFixtureTx()
		assert.Equal(t, tx, fixtureTx)
		inner()
		assert.Equal(t, tx, fixtureTx)
		outer()
		assert.Nil(t, fixtureTx)
	})
	t.Run("Panic", func(t *testing.T) {
		assert.Panics(t, func() {
			defer beginFixtureTx()()
			panic("fixture blew up")
		})
		assert.Nil(t, fixtureTx, "a panic must clear the fixture transaction")
		assert.Equal(t, Db(), fixtureDb())
	})
	t.Run("NoProvider", func(t *testing.T) {
		prev := dbConn
		SetDbProvider(nil)
		defer SetDbProvider(prev)
		done := beginFixtureTx()
		require.NotNil(t, done)
		assert.Nil(t, fixtureTx, "no provider means no transaction")
		done()
	})
}

func TestCreateFixture(t *testing.T) {
	t.Run("Duplicate", func(t *testing.T) {
		existing := CameraFixtures.Get("canon-eos-6d")
		duplicate := existing
		duplicate.CameraModel = "Duplicate Fixture"

		// A fixture whose key exists is skipped without an error, also inside the fixture transaction.
		done := beginFixtureTx()
		res := createFixture(&duplicate)
		done()
		require.NoError(t, res.Error)

		var stored Camera
		require.NoError(t, Db().First(&stored, existing.ID).Error)
		assert.Equal(t, existing.CameraModel, stored.CameraModel)
	})
	t.Run("New", func(t *testing.T) {
		camera := &Camera{CameraSlug: "create-fixture-test", CameraName: "Create Fixture Test", CameraMake: "Test", CameraModel: "Fixture"}
		require.NoError(t, createFixture(camera).Error)
		t.Cleanup(func() { _ = UnscopedDb().Delete(camera).Error })
		assert.NotZero(t, camera.ID)
	})
}
