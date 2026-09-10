package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixtureDb(t *testing.T) {
	ValidateFixtures(t)
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
	ValidateFixtures(t)
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
