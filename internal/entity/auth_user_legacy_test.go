package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity/legacy"
)

func TestFindLegacyUser(t *testing.T) {
	notFound := FindLegacyUser(Admin)
	assert.Nil(t, notFound)

	// t.Logf("Legacy Admin: %#v", notFound)

	if err := Db().AutoMigrate(legacy.User{}).Error; err != nil {
		log.Debugf("TestFindLegacyUser: %s (waiting 1s)", err.Error())

		time.Sleep(time.Second)

		if err = Db().AutoMigrate(legacy.User{}).Error; err != nil {
			log.Errorf("TestFindLegacyUser: failed migrating legacy.User")
			t.Error(err)
		}
	}

	Db().Save(legacy.Admin)

	found := FindLegacyUser(Admin)
	assert.NotNil(t, found)

	// t.Logf("Legacy Admin: %#v", found)

	if err := Db().DropTable(legacy.User{}).Error; err != nil {
		log.Errorf("TestFindLegacyUser: failed dropping legacy.User")
		t.Error(err)
	}
}

func TestFindLegacyUsers(t *testing.T) {
	notFound := FindLegacyUsers("all")
	assert.Len(t, notFound, 0)

	// t.Logf("Legacy Users: %#v", notFound)

	if err := Db().AutoMigrate(legacy.User{}).Error; err != nil {
		log.Debugf("TestFindLegacyUser: %s (waiting 1s)", err.Error())

		time.Sleep(time.Second)

		if err = Db().AutoMigrate(legacy.User{}).Error; err != nil {
			log.Errorf("TestFindLegacyUser: failed migrating legacy.User")
			t.Error(err)
		}
	}

	Db().Save(legacy.Admin)

	found := FindLegacyUsers("all")

	assert.NotNil(t, found)
	assert.Len(t, found, 1)

	// t.Logf("Legacy Users: %#v", found)

	if err := Db().DropTable(legacy.User{}).Error; err != nil {
		log.Errorf("TestFindLegacyUser: failed dropping legacy.User")
		t.Error(err)
	}
}
