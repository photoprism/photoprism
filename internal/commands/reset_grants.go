package commands

import (
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity"
)

// ResetAppGrants drops and recreates the tables with app grants and pending authorization codes.
func ResetAppGrants(db *gorm.DB) error {
	for _, table := range []any{&entity.Grant{}, &entity.AuthCode{}} {
		if err := db.Migrator().DropTable(table); err != nil {
			return err
		} else if err = entity.CreateTable(db, table); err != nil {
			return err
		}
	}

	return nil
}
