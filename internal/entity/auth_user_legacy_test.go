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

	t.Logf("Legacy Admin: %#v", notFound)

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

	t.Logf("Legacy Admin: %#v", found)

	if err := Db().DropTable(legacy.User{}).Error; err != nil {
		log.Errorf("TestFindLegacyUser: failed dropping legacy.User")
		t.Error(err)
	}
}
