package migrate

import (
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

// ConvertSQLiteDataTypes applies the data type conversions needed for SQLite and Gorm2
// This is a hacky attempt to prevent GORM from create temp, insert, drop, renaming for each column that has changed.
// It will still do it to create the foreign keys.
// If someone has a big sqlite database, this is going to take time.
func ConvertSQLiteDataTypes(db *gorm.DB) (err error) {
	results := db.Raw("select tbl_name from sqlite_master where type = 'table';")
	if results.Error != nil {
		log.Error("migrate: unable to get list of tables")
		return results.Error
	}
	type ResultTables struct {
		TblName string
	}
	type ResultSQL struct {
		SQL string
	}

	var tables []ResultTables
	if err := results.Scan(&tables).Error; err != nil {
		log.Errorf("migrate: unable to scan query %v", err)
	}

	reVarchar := regexp.MustCompile(`(?i)varchar\([0-9]+\)`)
	reVarbinary := regexp.MustCompile(`(?i)varbinary\([0-9]+\)|mediumblob`)
	reBigint := regexp.MustCompile(`(?i) bigint`)
	reBool := regexp.MustCompile(`(?i) bool`)
	reFloat := regexp.MustCompile(`(?i) float`)
	reCreate := regexp.MustCompile("(CREATE TABLE `[a-z_]+)(` )")
	reDblQuote := regexp.MustCompile(`"([a-z_]+)"`)
	reDEFAULTString := regexp.MustCompile(`DEFAULT '([a-z\/]*)'`)
	reTrailingSpaces := regexp.MustCompile(`([ ]+\))`)

	for _, table := range tables {
		log.Debugf("migrate: evaluating table %s", table.TblName)
		var createstatement ResultSQL
		db.Raw("SELECT sql FROM sqlite_master WHERE type = 'table' AND tbl_name = ? AND name = ?;", table.TblName, table.TblName).Scan(&createstatement)
		if strings.Contains(strings.ToLower(createstatement.SQL), "varchar") || strings.Contains(strings.ToLower(createstatement.SQL), "varbinary") || strings.Contains(strings.ToLower(createstatement.SQL), "bigint") {
			log.Debugf("migrate: working on table %s", table.TblName)
			tempStatement := reDblQuote.ReplaceAll([]byte(createstatement.SQL), []byte("`${1}`"))
			tempStatement = reDEFAULTString.ReplaceAll(tempStatement, []byte(`DEFAULT "${1}"`))
			tempStatement = reVarchar.ReplaceAll(tempStatement, []byte("text"))
			tempStatement = reVarbinary.ReplaceAll(tempStatement, []byte("blob"))
			tempStatement = reBool.ReplaceAll(tempStatement, []byte(" numeric"))
			tempStatement = reBigint.ReplaceAll(tempStatement, []byte(" integer"))
			tempStatement = reFloat.ReplaceAll(tempStatement, []byte(" real"))
			tempStatement = reCreate.ReplaceAll(tempStatement, []byte("${1}__temp${2} "))
			tempStatement = reTrailingSpaces.ReplaceAll(tempStatement, []byte(")"))
			createTempStatement := string(tempStatement)
			insertTempStatement := fmt.Sprintf("INSERT INTO %s__temp SELECT * FROM %s;", table.TblName, table.TblName)
			dropTempStatement := fmt.Sprintf("DROP TABLE %s;", table.TblName)
			alterTempStatement := fmt.Sprintf("ALTER TABLE %s__temp RENAME TO %s;", table.TblName, table.TblName)

			// Start a transaction
			tx := db.Begin()
			if tx.Error != nil {
				log.Errorf("migrate: unable to start transaction with %v", err)
				return fmt.Errorf("migrate: error creating transaction %w", tx.Error)
			}

			if err := tx.Exec(createTempStatement).Error; err != nil {
				if txErr := tx.Rollback().Error; txErr != nil {
					log.Errorf("migrate: rollback failure: %w", txErr)
				} else {
					log.Errorf("migrate: rolled back successfully")
				}
				log.Errorf("migrate: unable to execute %s with %v", createTempStatement, err)
				return err
			}

			if err := tx.Exec(insertTempStatement).Error; err != nil {
				if txErr := tx.Rollback().Error; txErr != nil {
					log.Errorf("migrate: rollback failure: %w", txErr)
				} else {
					log.Errorf("migrate: rolled back successfully")
				}
				log.Errorf("migrate: unable to execute %s with %v", insertTempStatement, err)
				return err
			}

			if err := tx.Exec(dropTempStatement).Error; err != nil {
				if txErr := tx.Rollback().Error; txErr != nil {
					log.Errorf("migrate: rollback failure: %w", txErr)
				} else {
					log.Errorf("migrate: rolled back successfully")
				}
				log.Errorf("migrate: unable to execute %s with %v", dropTempStatement, err)
				return err
			}

			if err := tx.Exec(alterTempStatement).Error; err != nil {
				if txErr := tx.Rollback().Error; txErr != nil {
					log.Errorf("migrate: rollback failure: %w", txErr)
				} else {
					log.Errorf("migrate: rolled back successfully")
				}
				log.Errorf("migrate: unable to execute %s with %v", alterTempStatement, err)
				return err
			}

			if txErr := tx.Commit().Error; txErr != nil {
				log.Errorf("migrate: commit failure for Convert SQLite Data Types: %w", txErr)
				return txErr
			}
			log.Debugf("migrate: committed changes to %s", table.TblName)
		}
	}

	return nil
}
