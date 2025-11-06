package migrate

import (
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

// Applies the data type conversions needed for SQLite and Gorm2
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
		Sql string
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
		log.Debugf("Evaluating table %s", table.TblName)
		var createstatement ResultSQL
		db.Raw("SELECT sql FROM sqlite_master WHERE type = 'table' AND tbl_name = ? AND name = ?;", table.TblName, table.TblName).Scan(&createstatement)
		//log.Debugf("%s", createstatement.Sql)
		if strings.Contains(strings.ToLower(createstatement.Sql), "varchar") || strings.Contains(strings.ToLower(createstatement.Sql), "varbinary") || strings.Contains(strings.ToLower(createstatement.Sql), "bigint") {
			log.Debugf("Working on table %s", table.TblName)
			tempStatement := reDblQuote.ReplaceAll([]byte(createstatement.Sql), []byte("`${1}`"))
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

			if err := db.Exec(createTempStatement).Error; err != nil {
				log.Errorf("migrate: unable to execute %s with %v", createTempStatement, err)
				return err
			}

			if err := db.Exec(insertTempStatement).Error; err != nil {
				log.Errorf("migrate: unable to execute %s with %v", insertTempStatement, err)
				return err
			}

			if err := db.Exec(dropTempStatement).Error; err != nil {
				log.Errorf("migrate: unable to execute %s with %v", dropTempStatement, err)
				return err
			}

			if err := db.Exec(alterTempStatement).Error; err != nil {
				log.Errorf("migrate: unable to execute %s with %v", alterTempStatement, err)
				return err
			}
		}
	}

	return nil
}
