package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity/legacy"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestFindLegacyUser(t *testing.T) {
	// Handle case where we are using a real database, rather than sqlite in memory db
	if Db().Migrator().HasTable("users") {
		if err := Db().Migrator().DropTable(legacy.User{}); err != nil {
			log.Errorf("TestFindLegacyUser: failed dropping legacy.User")
			t.Error(err)
		}
	}
	notFound := FindLegacyUser(Admin)
	assert.Nil(t, notFound)

	// t.Logf("Legacy Admin: %#v", notFound)

	if err := Db().AutoMigrate(&legacy.User{}); err != nil {
		log.Debugf("TestFindLegacyUser: %s (waiting 1s)", err.Error())

		time.Sleep(time.Second)

		if err = Db().AutoMigrate(&legacy.User{}); err != nil {
			log.Errorf("TestFindLegacyUser: failed migrating legacy.User")
			t.Error(err)
		}
	}

	Db().Save(&legacy.Admin)

	found := FindLegacyUser(Admin)
	assert.NotNil(t, found)

	// t.Logf("Legacy Admin: %#v", found)

	if err := Db().Migrator().DropTable(legacy.User{}); err != nil {
		log.Errorf("TestFindLegacyUser: failed dropping legacy.User")
		t.Error(err)
	}
}

func TestFindLegacyUsers(t *testing.T) {
	// Handle case where we are using a real database, rather than sqlite in memory db
	if Db().Migrator().HasTable("users") {
		if err := Db().Migrator().DropTable(legacy.User{}); err != nil {
			log.Errorf("TestFindLegacyUser: failed dropping legacy.User")
			t.Error(err)
		}
	}
	notFound := FindLegacyUsers("all")
	assert.Len(t, notFound, 0)

	// t.Logf("Legacy Users: %#v", notFound)

	if err := Db().AutoMigrate(&legacy.User{}); err != nil {
		log.Debugf("TestFindLegacyUser: %s (waiting 1s)", err.Error())

		time.Sleep(time.Second)

		if err = Db().AutoMigrate(&legacy.User{}); err != nil {
			log.Errorf("TestFindLegacyUser: failed migrating legacy.User")
			t.Error(err)
		}
	}

	Db().Save(&legacy.Admin)

	found := FindLegacyUsers("all")

	assert.NotNil(t, found)
	assert.Len(t, found, 1)

	// t.Logf("Legacy Users: %#v", found)

	if err := Db().Migrator().DropTable(legacy.User{}); err != nil {
		log.Errorf("TestFindLegacyUser: failed dropping legacy.User")
		t.Error(err)
	}
}

func TestFindLegacyUsers_Literal(t *testing.T) {
	require.NoError(t, Db().Migrator().AutoMigrate(legacy.User{}))
	t.Cleanup(func() { _ = Db().Migrator().DropTable(legacy.User{}) })

	base := "zzl" + rnd.Base36(5)

	for _, name := range []string{base + "_a", base + "Xa"} {
		require.NoError(t, Db().Create(&legacy.User{UserUID: rnd.GenerateUID(UserUID), UserName: name}).Error)
	}

	found := FindLegacyUsers(base + "_a")

	if assert.Len(t, found, 1) {
		assert.Equal(t, base+"_a", found[0].UserName)
	}
}
