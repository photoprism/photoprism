package entity

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/pkg/clean"
)

// TableMap holds the table name and the definition
type TableMap struct {
	TableName       string
	TableDefinition any
}

// Tables is the map to allow ordered table setup/teardown
type Tables map[int]TableMap

// Entities is the list of tables in the order that they must be processed for setup/teardown
var Entities = Tables{
	10:   {migrate.Migration{}.TableName(), &migrate.Migration{}},
	20:   {migrate.Version{}.TableName(), &migrate.Version{}},
	100:  {Error{}.TableName(), &Error{}},
	110:  {Password{}.TableName(), &Password{}},
	120:  {Passcode{}.TableName(), &Passcode{}},
	130:  {Duplicate{}.TableName(), &Duplicate{}},
	1000: {Subject{}.TableName(), &Subject{}},
	1010: {Service{}.TableName(), &Service{}},
	1020: {Lens{}.TableName(), &Lens{}},
	1030: {Camera{}.TableName(), &Camera{}},
	1040: {Place{}.TableName(), &Place{}},
	1050: {Folder{}.TableName(), &Folder{}},
	1060: {Label{}.TableName(), &Label{}},
	1070: {User{}.TableName(), &User{}},
	1080: {Keyword{}.TableName(), &Keyword{}},
	2000: {Face{}.TableName(), &Face{}},
	2010: {Cell{}.TableName(), &Cell{}},
	2020: {Category{}.TableName(), &Category{}},
	2030: {UserSettings{}.TableName(), &UserSettings{}},
	2040: {PhotoUser{}.TableName(), &PhotoUser{}},
	2050: {Client{}.TableName(), &Client{}},
	2060: {UserShare{}.TableName(), &UserShare{}},
	2070: {Reaction{}.TableName(), &Reaction{}},
	3000: {Country{}.TableName(), &Country{}},
	3010: {Photo{}.TableName(), &Photo{}},
	3020: {Country{}.TableName(), &Country{}}, // Deliberate as there is a fk to and from photo.
	3030: {Session{}.TableName(), &Session{}},
	3040: {UserDetails{}.TableName(), &UserDetails{}},
	4000: {PhotoKeyword{}.TableName(), &PhotoKeyword{}},
	4010: {File{}.TableName(), &File{}},
	4020: {Details{}.TableName(), &Details{}},
	4030: {Album{}.TableName(), &Album{}},
	4040: {PhotoLabel{}.TableName(), &PhotoLabel{}},
	5000: {Marker{}.TableName(), &Marker{}},
	5010: {FileSync{}.TableName(), &FileSync{}},
	5020: {FileShare{}.TableName(), &FileShare{}},
	5030: {PhotoAlbum{}.TableName(), &PhotoAlbum{}},
	5040: {AlbumUser{}.TableName(), &AlbumUser{}},
	6000: {Link{}.TableName(), &Link{}},
}

// WaitForMigration waits for the database migration to be successful and returns an error otherwise.
func (list Tables) WaitForMigration(db *gorm.DB) error {
	type RowCount struct {
		Count int
	}

	const attempts = 100
	for _, tables := range list {
		for i := 0; i <= attempts; i++ {
			count := RowCount{}
			if err := db.Raw(fmt.Sprintf("SELECT COUNT(*) AS count FROM %s", tables.TableName)).Scan(&count).Error; err != nil {
				log.Tracef("migrate: waiting for %s migration (%s)", clean.Log(tables.TableName), err.Error())
				time.Sleep(100 * time.Millisecond)
			} else {
				log.Tracef("migrate: %s migrated", clean.Log(tables.TableName))
				break
			}

			if i == attempts {
				return errors.New("some database tables are missing")
			}
		}
	}

	return nil
}

// Reset the ID increment to Current Max
func resetIDToMax(tableName string, db *gorm.DB) (err error) {
	sqlCommand := ""

	var maxid int64
	// The following lists all tables that have an ID column that needs to be reset.
	switch tableName {
	case Album{}.TableName():
		fallthrough
	case User{}.TableName():
		fallthrough
	case Camera{}.TableName():
		fallthrough
	case Error{}.TableName():
		fallthrough
	case File{}.TableName():
		fallthrough
	case Keyword{}.TableName():
		fallthrough
	case Label{}.TableName():
		fallthrough
	case Lens{}.TableName():
		fallthrough
	case Photo{}.TableName():
		fallthrough
	case Service{}.TableName():
		fallthrough
	case migrate.Version{}.TableName():
		switch DbDialect() {
		case MySQL:
			sqlCommand = fmt.Sprintf("ALTER TABLE `%v` AUTO_INCREMENT = 1", tableName)
		case Postgres:
			if err := db.Table(tableName).Select("COALESCE(MAX(id), 0) as MaxID").Row().Scan(&maxid); err != nil {
				return fmt.Errorf("resetIDToMax: unable to get max id from %s with %+v", tableName, err)
			}
			sqlCommand = fmt.Sprintf("ALTER SEQUENCE %v_id_seq RESTART WITH %d", tableName, maxid+1)
		case SQLite3:
			if err := db.Table(tableName).Select("COALESCE(MAX(id), 0) as MaxID").Row().Scan(&maxid); err != nil {
				return fmt.Errorf("resetIDToMax: unable to get max id from %s with %+v", tableName, err)
			}
			sqlCommand = fmt.Sprintf("UPDATE SQLITE_SEQUENCE SET SEQ=%d WHERE NAME='%v'", maxid, tableName)
		default:
			log.Warnf("resetIDToMax: Unsupported database dialect %s", DbDialect())
			return nil
		}
		if res := db.Unscoped().Exec(sqlCommand); res.Error != nil {
			return fmt.Errorf("resetIDToMax: failed with %+v", res.Error)
		}
	default:
		return nil
	}

	return nil
}

// ResetSequences resets the auto increment sequence for ID columns to the next valid number
func (list Tables) ResetSequences(db *gorm.DB) (err error) {
	for _, entity := range list {
		if err = resetIDToMax(entity.TableName, db); err != nil {
			return err
		}
	}
	return nil
}

// Truncate removes all data from tables without dropping them.
func (list Tables) Truncate(db *gorm.DB) {
	var name string

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("migrate: %s in %s (truncate)", r, name)
		}
	}()

	log.Info("migrate: truncate all data tables")

	// Delete tables based on referential integrity.
	log.Info("migrate: truncate Details")
	if res := db.Unscoped().Where("1=1").Delete(&Details{}); res.Error != nil {
		log.Errorf("Delete of Details failed with %v", res.Error)
	}
	log.Info("migrate: truncate FileShare")
	if res := db.Unscoped().Where("1=1").Delete(&FileShare{}); res.Error != nil {
		log.Errorf("Delete of FileShare failed with %v", res.Error)
	}
	log.Info("migrate: truncate FileSync")
	if res := db.Unscoped().Where("1=1").Delete(&FileSync{}); res.Error != nil {
		log.Errorf("Delete of FileSync failed with %v", res.Error)
	}
	log.Info("migrate: truncate File")
	if res := db.Unscoped().Where("1=1").Delete(&File{}); res.Error != nil {
		log.Errorf("Delete of Files failed with %v", res.Error)
	}
	log.Info("migrate: truncate PhotoKeyword")
	if res := db.Unscoped().Where("1=1").Delete(&PhotoKeyword{}); res.Error != nil {
		log.Errorf("Delete of PhotoKeyword failed with %v", res.Error)
	}
	log.Info("migrate: truncate PhotoLabel")
	if res := db.Unscoped().Where("1=1").Delete(&PhotoLabel{}); res.Error != nil {
		log.Errorf("Delete of PhotoLabel failed with %v", res.Error)
	}
	log.Info("migrate: truncate PhotoAlbum")
	if res := db.Unscoped().Where("1=1").Delete(&PhotoAlbum{}); res.Error != nil {
		log.Errorf("Delete of PhotoAlbum failed with %v", res.Error)
	}
	log.Info("migrate: truncate Photo")
	if res := db.Unscoped().Where("1=1").Delete(&Photo{}); res.Error != nil {
		log.Errorf("Delete of Photo failed with %v", res.Error)
	}
	log.Info("migrate: truncate Error")
	if res := db.Unscoped().Where("1=1").Delete(&Error{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Password")
	if res := db.Unscoped().Where("1=1").Delete(&Password{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Passcode")
	if res := db.Unscoped().Where("1=1").Delete(&Passcode{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate UserDetails")
	if res := db.Unscoped().Where("1=1").Delete(&UserDetails{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate UserSettings")
	if res := db.Unscoped().Where("1=1").Delete(&UserSettings{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate UserShare")
	if res := db.Unscoped().Where("1=1").Delete(&UserShare{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate User")
	if res := db.Unscoped().Where("1=1").Delete(&User{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Session")
	if res := db.Unscoped().Where("1=1").Delete(&Session{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Client")
	if res := db.Unscoped().Where("1=1").Delete(&Client{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Service")
	if res := db.Unscoped().Where("1=1").Delete(&Service{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Folder")
	if res := db.Unscoped().Where("1=1").Delete(&Folder{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Duplicate")
	if res := db.Unscoped().Where("1=1").Delete(&Duplicate{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate PhotoUser")
	if res := db.Unscoped().Where("1=1").Delete(&PhotoUser{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Cell")
	if res := db.Unscoped().Where("1=1").Delete(&Cell{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Place")
	if res := db.Unscoped().Where("1=1").Delete(&Place{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Camera")
	if res := db.Unscoped().Where("1=1").Delete(&Camera{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Lens")
	if res := db.Unscoped().Where("1=1").Delete(&Lens{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Country")
	if res := db.Unscoped().Where("1=1").Delete(&Country{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate AlbumUser")
	if res := db.Unscoped().Where("1=1").Delete(&AlbumUser{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Albums")
	if res := db.Unscoped().Where("1=1").Delete(&Albums{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Category")
	if res := db.Unscoped().Where("1=1").Delete(&Category{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Label")
	if res := db.Unscoped().Where("1=1").Delete(&Labels{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Keyword")
	if res := db.Unscoped().Where("1=1").Delete(&Keyword{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Link")
	if res := db.Unscoped().Where("1=1").Delete(&Link{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Subjects")
	if res := db.Unscoped().Where("1=1").Delete(&Subjects{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Face")
	if res := db.Unscoped().Where("1=1").Delete(&Face{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Marker")
	if res := db.Unscoped().Where("1=1").Delete(&Marker{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	log.Info("migrate: truncate Reaction")
	if res := db.Unscoped().Where("1=1").Delete(&Reaction{}); res.Error != nil {
		log.Errorf("Delete failed with %v", res.Error)
	}
	list.ResetSequences(db)
}

// Migrate migrates all database tables of registered entities.
func (list Tables) Migrate(db *gorm.DB, opt migrate.Options) {
	var name string

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("migrate: %s in %s (panic)", r, name)
		}
	}()

	log.Debugf("migrate: running database migrations")

	// Check if this is an empty database
	tablesFound := 0
	for _, tableMap := range list {
		if db.Migrator().HasTable(tableMap.TableDefinition) {
			tablesFound++
		}
	}

	if tablesFound < 3 { // migrations and versions may be found.
		opt.NewDatabase = true
	}

	// Run pre migrations, if any.
	if err := migrate.Run(db, opt.Pre()); err != nil {
		log.Error(err)
	}

	// Run ORM auto migrations.
	if opt.AutoMigrate {
		// Check if the DBMS AuthID fix has been applied?
		version := migrate.FirstOrCreateVersion(db, migrate.NewVersion("DBMS AuthID Fix", "Any Editions"))
		if version.NeedsMigration() {
			if err := migrate.ConvertDBMSAuthIDDataTypes(db); err != nil {
				log.Errorf("migrate: could not apply dbms auth_id fix : %v", err)
				version.Error = err.Error()
				if saveErr := version.Save(db); saveErr != nil {
					log.Errorf("migrate: could not save dbms auth_id fix status: %v", saveErr)
				}
			} else {
				if migratedErr := version.Migrated(db); migratedErr != nil {
					log.Errorf("migrate: could not persist dbms auth_id fix status: %v", migratedErr)
				}
				log.Debug("migrate: DBMS AuthID fix migrated")
			}
		} else {
			log.Debug("migrate: DBMS AuthID fix skipped")
		}

		// Check if the GORMv2 sqlite conversion has been done?
		if db.Dialector.Name() == SQLite3 {
			version := migrate.FirstOrCreateVersion(db, migrate.NewVersion("Gorm For SQLite", "V2 Upgrade"))
			if version.NeedsMigration() {
				if err := migrate.ConvertSQLiteDataTypes(db); err != nil {
					log.Error("migrate: could not convert sqlite datatypes : ", err)
				} else {
					version.Migrated(db)
				}
			}
		}

		// Setup required explicit join tables
		err := db.SetupJoinTable(&Photo{}, "Albums", &PhotoAlbum{})
		if err != nil {
			log.Error("migrate: could not setup join table for Photo - Albums: ", err)
		}
		err = db.SetupJoinTable(&Photo{}, "Keywords", &PhotoKeyword{})
		if err != nil {
			log.Error("migrate: could not setup join table for Photo - Keywords: ", err)
		}
		err = db.SetupJoinTable(&Label{}, "LabelCategories", &Category{})
		if err != nil {
			log.Error("migrate: could not setup join table for Label - Categories: ", err)
		}

		var entity any
		orderedList := make([]int, len(list))
		i := 0
		for id := range list {
			orderedList[i] = id
			i++
		}
		slices.Sort(orderedList)

		var tableMap TableMap

		for _, id := range orderedList {
			tableMap = list[id]
			name = tableMap.TableName
			entity = tableMap.TableDefinition
			log.Debugf("migrate: auto-migrating %s", name)
			if err := db.AutoMigrate(entity); err != nil {
				log.Debugf("migrate: %s (waiting 1s)", err.Error())

				time.Sleep(time.Second)

				if err = db.AutoMigrate(entity); err != nil {
					log.Errorf("migrate: failed migrating %s", clean.Log(name))
					panic(err)
				}
			}
			counter := int64(0)
			if res := db.Model(entity).Count(&counter); res.Error != nil {
				log.Warningf("migrate: automigrate failure on %v, retrying in 1s", name)
				time.Sleep(time.Second)
				if err = db.AutoMigrate(entity); err != nil {
					log.Errorf("migrate: failed migrating %s", clean.Log(name))
					panic(err)
				}
			}
		}
	}

	// Run main migrations, if any.
	if err := migrate.Run(db, opt); err != nil {
		log.Error(err)
	}

	// Run post-migrations, if any.
	if err := migrate.Run(db, opt.Post()); err != nil {
		log.Error(err)
	}

}

// Drop drops all database tables of registered entities.
func (list Tables) Drop(db *gorm.DB) {
	for _, entity := range list {
		if err := db.Migrator().DropTable(entity.TableName); err != nil {
			// Migrator().DropTable(table) is doing a DROP TABLE IF EXISTS under the covers.
			// Testing will show if we need a panic here or a log.Error(err)
			panic(err)
		}
	}
	// Put Migrations and Versions back as they are handled by Once and it's already run.
	if err := db.Migrator().AutoMigrate(&migrate.Migration{}); err != nil {
		panic(err)
	}
	if err := db.Migrator().AutoMigrate(&migrate.Version{}); err != nil {
		panic(err)
	}
}
