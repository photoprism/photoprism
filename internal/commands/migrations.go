package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/pkg/txt/report"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MigrationsStatusCommand lists migration status.
var MigrationsStatusCommand = &cli.Command{
	Name:      "ls",
	Aliases:   []string{"status", "show"},
	Usage:     "Displays the status of schema migrations",
	ArgsUsage: "[migrations]...",
	Flags:     report.CliFlags,
	Action:    migrationsStatusAction,
}

// MigrationsRunCommand runs pending migrations.
var MigrationsRunCommand = &cli.Command{
	Name:      "run",
	Aliases:   []string{"execute", "migrate"},
	Usage:     "Executes database schema migrations",
	ArgsUsage: "[migrations]...",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "failed",
			Aliases: []string{"f"},
			Usage:   "runs previously failed migrations",
		},
		&cli.BoolFlag{
			Name:    "trace",
			Aliases: []string{"t"},
			Usage:   "shows trace logs for debugging",
		},
	},
	Action: migrationsRunAction,
}

var MigrationsTransferCommand = &cli.Command{
	Name:      "transfer",
	Aliases:   []string{"copy"},
	Usage:     "Executes database data transfers",
	ArgsUsage: "[migrations...]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "force",
			Aliases: []string{"f"},
			Usage:   "truncates target tables if populated",
		},
		&cli.BoolFlag{
			Name:    "trace",
			Aliases: []string{"t"},
			Usage:   "shows trace logs for debugging",
		},
		&cli.UintFlag{
			Name:    "batch",
			Aliases: []string{"b"},
			Usage:   "the number of records to transfer in each batch",
			Value:   100,
		},
	},
	Action: migrationsTransferAction,
}

// MigrationsCommands registers the "migrations" CLI command.
var MigrationsCommands = &cli.Command{
	Name:  "migrations",
	Usage: "Database schema migration subcommands",
	Subcommands: []*cli.Command{
		MigrationsStatusCommand,
		MigrationsRunCommand,
		MigrationsTransferCommand,
	},
}

// migrationsStatusAction lists the status of schema migration.
func migrationsStatusAction(ctx *cli.Context) error {
	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	defer conf.Shutdown()

	var ids []string

	// Check argument for specific migrations to be run.
	if migrations := strings.TrimSpace(ctx.Args().First()); migrations != "" {
		ids = strings.Fields(migrations)
	}

	db := conf.Db()

	status, err := migrate.Status(db, ids)

	if err != nil {
		return err
	}

	// Report columns.
	cols := []string{"ID", "Dialect", "Stage", "Started At", "Finished At", "Status"}

	// Report rows.
	rows := make([][]string, 0, len(status))

	for _, m := range status {
		var stage, started, finished, info string

		if m.Stage == "" {
			stage = "main"
		} else {
			stage = m.Stage
		}

		if m.StartedAt.IsZero() {
			started = "-"
		} else {
			started = m.StartedAt.Format("2006-01-02 15:04:05")
		}

		if m.Finished() {
			finished = m.FinishedAt.Format("2006-01-02 15:04:05")
		} else {
			finished = "-"
		}

		switch {
		case m.Error != "":
			info = m.Error
		case m.Finished():
			info = "OK"
		case m.StartedAt.IsZero():
			info = "-"
		case m.Repeat(false):
			info = "Repeat"
		default:
			info = "Running?"
		}

		rows = append(rows, []string{m.ID, m.Dialect, stage, started, finished, info})
	}

	// Display report.
	info, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

	if err != nil {
		return err
	}

	fmt.Println(info)

	return nil
}

// migrationsRunAction executes database schema migrations.
func migrationsRunAction(ctx *cli.Context) error {
	if ctx.Args().First() == "ls" {
		return fmt.Errorf("run '%s migrations ls' to display the status of schema migrations", filepath.Base(os.Args[0]))
	}

	start := time.Now()

	conf := config.NewConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	defer conf.Shutdown()

	if ctx.Bool("trace") {
		log.SetLevel(logrus.TraceLevel)
		log.Infoln("migrate: enabled trace mode")
	}

	runFailed := ctx.Bool("failed")

	if runFailed {
		log.Infoln("migrate: running previously failed migrations")
	}

	var ids []string

	// Check argument for specific migrations to be run.
	if migrations := strings.TrimSpace(ctx.Args().First()); migrations != "" {
		ids = strings.Fields(migrations)
	}

	log.Infoln("migrating database schema...")

	// Run migrations.
	conf.MigrateDb(runFailed, ids)

	elapsed := time.Since(start)

	log.Infof("completed in %s", elapsed)

	return nil
}

// migrationsTransferAction executes database data migrations.
func migrationsTransferAction(ctx *cli.Context) error {
	if ctx.Args().First() == "ls" {
		return fmt.Errorf("run '%s migrations ls' to display the status of schema migrations", filepath.Base(os.Args[0]))
	}

	batchSize := int(ctx.Uint("batch"))
	log.Infof("migrate: transfer batch size set to %d", batchSize)

	start := time.Now()

	conf := config.NewConfig(ctx)
	tfrConf := config.NewConfig(ctx)

	//log = event.Log

	if err := tfrConf.SwapDBAndTransfer(); err != nil {
		return err
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	if err := tfrConf.Init(); err != nil {
		return err
	}

	defer conf.Shutdown()
	defer tfrConf.Shutdown()

	if ctx.Bool("trace") {
		log.SetLevel(logrus.TraceLevel)
		log.Infoln("migrate: enabled trace mode")
	}

	var ids []string

	runForced := ctx.Bool("force")
	log.Infoln("migrate: ensure target is empty...")
	var entitiesToCheck []any
	entitiesToCheck = append(entitiesToCheck, &entity.Photo{})
	entitiesToCheck = append(entitiesToCheck, &entity.Album{})
	entitiesToCheck = append(entitiesToCheck, &entity.File{})
	entitiesToCheck = append(entitiesToCheck, &entity.Face{})
	entitiesToCheck = append(entitiesToCheck, &entity.Folder{})

	recordsFound := false
	totalRecords := int64(0)
	for _, toCheck := range entitiesToCheck {
		if tfrConf.Db().Unscoped().Migrator().HasTable(toCheck) {
			var toCheckCount int64
			if err := tfrConf.Db().Unscoped().Model(toCheck).Count(&toCheckCount).Error; err != nil {
				log.Errorf("migrate: count of model has failed with %s", err.Error())
				return err
			}
			if toCheckCount > 0 {
				recordsFound = true
				totalRecords += toCheckCount
			}
		}
	}

	if !recordsFound {
		runForced = false
	} else {
		if !runForced {
			errmsg := fmt.Sprintf("migrate: transfer target database is not empty, %d records found across %d tables", totalRecords, len(entitiesToCheck))
			log.Error(errmsg)
			return fmt.Errorf("%s", errmsg)
		} else {
			errmsg := fmt.Sprintf("migrate: transfer target database is not empty, %d records found across %d tables but force enabled", totalRecords, len(entitiesToCheck))
			log.Warning(errmsg)
		}
	}

	log.Infoln("migrate: migrating database schema...")

	// Run migrations.
	log.Infoln("migrate: migrating against target")
	if !tfrConf.Db().Unscoped().Migrator().HasTable(&migrate.Version{}) {
		err := tfrConf.Db().Unscoped().AutoMigrate(&migrate.Version{})
		if err != nil {
			return fmt.Errorf("migrate: %s (create versions table)", err)
		}
	}
	if !tfrConf.Db().Unscoped().Migrator().HasTable(&migrate.Migration{}) {
		err := tfrConf.Db().Unscoped().AutoMigrate(&migrate.Migration{})
		if err != nil {
			return fmt.Errorf("migrate: %s (create migrations table)", err)
		}
	}
	entity.SetDbProvider(tfrConf)
	tfrConf.MigrateDb(false, ids)
	if runForced {
		entity.Entities.Truncate(tfrConf.Db())
		entity.CreateDefaultFixtures()
	}
	log.Infoln("migrate: migrating against source")
	entity.SetDbProvider(conf)
	conf.MigrateDb(false, ids)

	// Copy tables
	type MaxResult struct {
		Id uint
	}
	var currentid MaxResult

	// Replace the admin user with the source one.
	var userRecord entity.User
	if err := tfrConf.Db().Unscoped().
		Where("id = 1").
		Find(&userRecord).Error; err != nil {
		log.Errorf("migrate: error in admin user preselect %s", err.Error())
		return err
	} else {
		if err = tfrConf.Db().Unscoped().Delete(&entity.UserDetails{UserUID: userRecord.UserUID}).Error; err != nil {
			log.Errorf("migrate: error in admin user cleanup user details %s", err.Error())
			return err
		}
		if err = tfrConf.Db().Unscoped().Delete(&entity.UserSettings{UserUID: userRecord.UserUID}).Error; err != nil {
			log.Errorf("migrate: error in admin user cleanup user settings %s", err.Error())
			return err
		}
		if err = tfrConf.Db().Unscoped().Delete(&entity.UserShare{UserUID: userRecord.UserUID}).Error; err != nil {
			log.Errorf("migrate: error in admin user cleanup user share %s", err.Error())
			return err
		}
		if err = tfrConf.Db().Unscoped().Delete(&entity.Password{UID: userRecord.UserUID}).Error; err != nil {
			log.Errorf("migrate: error in admin user cleanup password %s", err.Error())
			return err
		}
		if err = tfrConf.Db().Unscoped().Delete(&entity.Passcode{UID: userRecord.UserUID}).Error; err != nil {
			log.Errorf("migrate: error in admin user cleanup passcode %s", err.Error())
			return err
		}
		if err := conf.Db().Unscoped().
			Where("id = 1").
			Find(&userRecord).Error; err != nil {
			log.Errorf("migrate: error in admin user preselect %s", err.Error())
			return err
		} else {
			if err = tfrConf.Db().Save(&userRecord).Error; err != nil {
				log.Errorf("migrate: error in admin user update %s", err.Error())
				return err
			}
		}
	}

	// Bring the rest of the users across
	var users entity.Users
	result := conf.Db().Unscoped().
		//Clauses(clause.OnConflict{UpdateAll: true}).  // <-- This is not working against sqlite.  Not generating the ON CONFLICT statements.
		Where("id > 1").
		FindInBatches(&users, batchSize, func(tx *gorm.DB, batch int) error {
			var newUsers []*entity.User
			for _, user := range users {
				newUsers = append(newUsers, &user)
			}
			if result := tfrConf.Db().Create(newUsers); result.Error != nil {
				log.Errorf("migrate: error in batch user create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of users transfered %v", result.RowsAffected)
		if result.RowsAffected > 0 {
			if result = tfrConf.Db().Unscoped().
				Model(&entity.User{}).Select("MAX(id) as id").Scan(&currentid); result.Error != nil {
				log.Errorf("migrate: error in getting max id %v", result.Error)
				return result.Error
			} else {
				resetIDToValue(tfrConf.Db(), entity.User{}.TableName(), int(currentid.Id))
			}
		}
	}

	var albums entity.Albums
	result = conf.Db().Unscoped().
		FindInBatches(&albums, batchSize, func(tx *gorm.DB, batch int) error {
			var newAlbums []*entity.Album
			for _, album := range albums {
				newAlbums = append(newAlbums, &album)
			}
			if result := tfrConf.Db().Create(newAlbums); result.Error != nil {
				log.Errorf("migrate: error in batch album create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of albums transfered %v", result.RowsAffected)
		if result.RowsAffected > 0 {
			if result = tfrConf.Db().Unscoped().
				Model(&entity.Album{}).Select("MAX(id) as id").Scan(&currentid); result.Error != nil {
				log.Errorf("migrate: error in getting max id %v", result.Error)
				return result.Error
			} else {
				resetIDToValue(tfrConf.Db(), entity.Album{}.TableName(), int(currentid.Id))
			}
		}
	}

	var cameras []entity.Camera
	result = conf.Db().Unscoped().
		Where("id > 1").
		FindInBatches(&cameras, batchSize, func(tx *gorm.DB, batch int) error {
			var newCameras []*entity.Camera
			for _, camera := range cameras {
				newCameras = append(newCameras, &camera)
			}
			if result := tfrConf.Db().Create(newCameras); result.Error != nil {
				log.Errorf("migrate: error in batch camera create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of cameras transfered %v", result.RowsAffected)
		if result.RowsAffected > 0 {
			if result = tfrConf.Db().Unscoped().
				Model(&entity.Camera{}).Select("MAX(id) as id").Scan(&currentid); result.Error != nil {
				log.Errorf("migrate: error in getting max id %v", result.Error)
				return result.Error
			} else {
				resetIDToValue(tfrConf.Db(), entity.Camera{}.TableName(), int(currentid.Id))
			}
		}
	}

	var lenses entity.Lenses
	result = conf.Db().Unscoped().
		Where("id > 1").
		FindInBatches(&lenses, batchSize, func(tx *gorm.DB, batch int) error {
			var newLenses []*entity.Lens
			for _, lens := range lenses {
				newLenses = append(newLenses, &lens)
			}
			if result := tfrConf.Db().Create(newLenses); result.Error != nil {
				log.Errorf("migrate: error in batch lens create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of lenses transfered %v", result.RowsAffected)
		if result.RowsAffected > 0 {
			if result = tfrConf.Db().Unscoped().
				Model(&entity.Lens{}).Select("MAX(id) as id").Scan(&currentid); result.Error != nil {
				log.Errorf("migrate: error in getting max id %v", result.Error)
				return result.Error
			} else {
				resetIDToValue(tfrConf.Db(), entity.Lens{}.TableName(), int(currentid.Id))
			}
		}
	}

	var places []entity.Place
	result = conf.Db().Unscoped().
		Where("id <> 'zz'").
		FindInBatches(&places, batchSize, func(tx *gorm.DB, batch int) error {
			var newPlaces []*entity.Place
			for _, place := range places {
				newPlaces = append(newPlaces, &place)
			}
			if result := tfrConf.Db().Create(newPlaces); result.Error != nil {
				log.Errorf("migrate: error in batch place create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of places transfered %v", result.RowsAffected)
	}

	var cells []entity.Cell
	result = conf.Db().Unscoped().
		Where("id <> 'zz'").
		FindInBatches(&cells, batchSize, func(tx *gorm.DB, batch int) error {
			var newCells []*entity.Cell
			for _, cell := range cells {
				newCells = append(newCells, &cell)
			}
			if result := tfrConf.Db().Create(newCells); result.Error != nil {
				log.Errorf("migrate: error in batch cell create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of cells transfered %v", result.RowsAffected)
	}

	var countries entity.Countries
	result = conf.Db().Unscoped().
		Where("id <> 'zz'").
		FindInBatches(&countries, batchSize, func(tx *gorm.DB, batch int) error {
			var newCountries []*entity.Country
			for _, country := range countries {
				newCountries = append(newCountries, &country)
			}
			if result := tfrConf.Db().Create(newCountries); result.Error != nil {
				log.Errorf("migrate: error in batch country create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of countries transfered %v", result.RowsAffected)
	}

	var keywords []entity.Keyword
	result = conf.Db().Unscoped().
		FindInBatches(&keywords, batchSize, func(tx *gorm.DB, batch int) error {
			var newKeywords []*entity.Keyword
			for _, keyword := range keywords {
				newKeywords = append(newKeywords, &keyword)
			}
			if result := tfrConf.Db().Create(newKeywords); result.Error != nil {
				log.Errorf("migrate: error in batch keyword create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of keywords transfered %v", result.RowsAffected)
		if result.RowsAffected > 0 {
			if result = tfrConf.Db().Unscoped().
				Model(&entity.Keyword{}).Select("MAX(id) as id").Scan(&currentid); result.Error != nil {
				log.Errorf("migrate: error in getting max id %v", result.Error)
				return result.Error
			} else {
				resetIDToValue(tfrConf.Db(), entity.Keyword{}.TableName(), int(currentid.Id))
			}
		}
	}

	var labels []entity.Label
	result = conf.Db().Unscoped().
		FindInBatches(&labels, batchSize, func(tx *gorm.DB, batch int) error {
			var newLabels []*entity.Label
			for _, label := range labels {
				newLabels = append(newLabels, &label)
			}
			if result := tfrConf.Db().Create(newLabels); result.Error != nil {
				log.Errorf("migrate: error in batch label create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of labels transfered %v", result.RowsAffected)
		if result.RowsAffected > 0 {
			if result = tfrConf.Db().Unscoped().
				Model(&entity.Label{}).Select("MAX(id) as id").Scan(&currentid); result.Error != nil {
				log.Errorf("migrate: error in getting max id %v", result.Error)
				return result.Error
			} else {
				resetIDToValue(tfrConf.Db(), entity.Label{}.TableName(), int(currentid.Id))
			}
		}
	}

	var photos []entity.Photo
	result = conf.Db().Unscoped().
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos_labels.uncertainty ASC, photos_labels.label_id DESC")
		}).
		Preload("Labels.Label").
		Preload("Camera").
		Preload("Lens").
		Preload("Details").
		Preload("Place").
		Preload("Cell").
		Preload("Cell.Place").
		Preload("Albums").
		Preload("Keywords").
		Preload("Labels").
		FindInBatches(&photos, batchSize, func(tx *gorm.DB, batch int) error {
			var newPhotos []*entity.Photo
			for _, photo := range photos {
				newPhotos = append(newPhotos, &photo)
			}
			if result := tfrConf.Db().Create(newPhotos); result.Error != nil {
				log.Errorf("migrate: error in batch photo create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of photos transfered %v", result.RowsAffected)
		if result.RowsAffected > 0 {
			if result = tfrConf.Db().Unscoped().
				Model(&entity.Photo{}).Select("MAX(id) as id").Scan(&currentid); result.Error != nil {
				log.Errorf("migrate: error in getting max id %v", result.Error)
				return result.Error
			} else {
				resetIDToValue(tfrConf.Db(), entity.Photo{}.TableName(), int(currentid.Id))
			}
		}
	}

	var files entity.Files
	result = conf.Db().Unscoped().FindInBatches(&files, batchSize, func(tx *gorm.DB, batch int) error {
		var newFiles []*entity.File
		for _, file := range files {
			newFiles = append(newFiles, &file)
		}
		if result := tfrConf.Db().Create(newFiles); result.Error != nil {
			log.Errorf("migrate: error in batch file create %s", result.Error)
			return result.Error
		}
		return nil
	})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of files transfered %v", result.RowsAffected)
		if result.RowsAffected > 0 {
			if result = tfrConf.Db().Unscoped().
				Model(&entity.File{}).Select("MAX(id) as id").Scan(&currentid); result.Error != nil {
				log.Errorf("migrate: error in getting max id %v", result.Error)
				return result.Error
			} else {
				resetIDToValue(tfrConf.Db(), entity.File{}.TableName(), int(currentid.Id))
			}
		}
	}

	var albumUsers []entity.AlbumUser
	records := batchSize
	currentOffset := 0
	for records == batchSize {
		if result = conf.Db().Unscoped().
			Model(&entity.AlbumUser{}).
			Limit(batchSize).Offset(currentOffset).
			Order("uid, user_uid").
			Find(&albumUsers); result.Error != nil {
			log.Errorf("migrate: error in albumuser find %s", result.Error)
			return result.Error
		}
		records = int(result.RowsAffected)
		currentOffset += records
		if records > 0 {
			if result := tfrConf.Db().Create(&albumUsers); result.Error != nil {
				log.Errorf("migrate: error in albumuser create %s", result.Error)
				return result.Error
			}
		}
	}
	log.Infof("migrate: number of albumusers transfered %v", currentOffset)

	// Gorm bug with composite foreign keys prevents the following from working.
	// It can be used once pull https://github.com/go-gorm/gorm/pull/7453 (or equivalent) has been made to gorm
	// result = conf.Db().Unscoped().
	// 	FindInBatches(&albumUsers, batchSize, func(tx *gorm.DB, batch int) error {
	// 		var newAlbumUsers []*entity.AlbumUser
	// 		for _, albumUser := range albumUsers {
	// 			newAlbumUsers = append(newAlbumUsers, &albumUser)
	// 		}
	// 		if result := tfrConf.Db().Create(newAlbumUsers); result.Error != nil {
	// 			log.Errorf("migrate: error in batch albumuser create %s", result.Error)
	// 			return result.Error
	// 		}
	// 		return nil
	// 	})
	// if result.Error != nil {
	// 	return result.Error
	// } else {
	// 	log.Infof("migrate: number of albumusers transfered %v", result.RowsAffected)
	// }

	var clients entity.Clients
	result = conf.Db().Unscoped().
		FindInBatches(&clients, batchSize, func(tx *gorm.DB, batch int) error {
			var newClients []*entity.Client
			for _, client := range clients {
				newClients = append(newClients, &client)
			}
			if result := tfrConf.Db().Create(newClients); result.Error != nil {
				log.Errorf("migrate: error in batch client create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of clients transfered %v", result.RowsAffected)
	}

	var sessions entity.Sessions
	result = conf.Db().Unscoped().
		FindInBatches(&sessions, batchSize, func(tx *gorm.DB, batch int) error {
			var newSessions []*entity.Session
			for _, session := range sessions {
				newSessions = append(newSessions, &session)
			}
			if result := tfrConf.Db().Create(newSessions); result.Error != nil {
				log.Errorf("migrate: error in batch session create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of sessions transfered %v", result.RowsAffected)
	}

	var userdetails []entity.UserDetails
	result = conf.Db().Unscoped().
		Where("user_uid <> ''").
		FindInBatches(&userdetails, batchSize, func(tx *gorm.DB, batch int) error {
			var newUserDetails []*entity.UserDetails
			for _, userdetail := range userdetails {
				newUserDetails = append(newUserDetails, &userdetail)
			}
			if result := tfrConf.Db().
				Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "user_uid"}},
					UpdateAll: true,
				}).Create(newUserDetails); result.Error != nil {
				log.Errorf("migrate: error in batch userdetail create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of userdetails transfered %v", result.RowsAffected)
	}

	var usersettings []entity.UserSettings
	result = conf.Db().Unscoped().
		Where("user_uid <> ''").
		FindInBatches(&usersettings, batchSize, func(tx *gorm.DB, batch int) error {
			var newUserSettings []*entity.UserSettings
			for _, usersetting := range usersettings {
				newUserSettings = append(newUserSettings, &usersetting)
			}
			if result := tfrConf.Db().
				Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "user_uid"}},
					UpdateAll: true,
				}).
				Create(newUserSettings); result.Error != nil {
				log.Errorf("migrate: error in batch usersetting create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of usersettings transfered %v", result.RowsAffected)
	}

	var usershares []entity.UserShare
	records = batchSize
	currentOffset = 0
	for records == batchSize {
		if result = conf.Db().Unscoped().
			Model(&entity.UserShare{}).
			Limit(batchSize).Offset(currentOffset).
			Order("user_uid, share_uid").
			Find(&usershares); result.Error != nil {
			log.Errorf("migrate: error in usershares find %s", result.Error)
			return result.Error
		}
		records = int(result.RowsAffected)
		currentOffset += records
		if records > 0 {
			if result := tfrConf.Db().Create(&usershares); result.Error != nil {
				log.Errorf("migrate: error in usershares create %s", result.Error)
				return result.Error
			}
		}
	}
	log.Infof("migrate: number of usershares transfered %v", currentOffset)

	// Gorm bug with composite foreign keys prevents the following from working.
	// It can be used once pull https://github.com/go-gorm/gorm/pull/7453 (or equivalent) has been made to gorm
	// var usershares entity.UserShares
	// result = conf.Db().Unscoped().
	// 	Where("user_uid <> ''").
	// 	FindInBatches(&usershares, batchSize, func(tx *gorm.DB, batch int) error {
	// 		var newUserShares []*entity.UserShare
	// 		for _, usershare := range usershares {
	// 			newUserShares = append(newUserShares, &usershare)
	// 		}
	// 		if result := tfrConf.Db().Create(newUserShares); result.Error != nil {
	// 			log.Errorf("migrate: error in batch usershare create %s", result.Error)
	// 			return result.Error
	// 		}
	// 		return nil
	// 	})
	// if result.Error != nil {
	// 	return result.Error
	// } else {
	// 	log.Infof("migrate: number of usershares transfered %v", result.RowsAffected)
	// }

	var categories []entity.Category
	records = batchSize
	currentOffset = 0
	for records == batchSize {
		if result = conf.Db().Unscoped().
			Model(&entity.Category{}).
			Limit(batchSize).Offset(currentOffset).
			Order("label_id, category_id").
			Find(&categories); result.Error != nil {
			log.Errorf("migrate: error in categories find %s", result.Error)
			return result.Error
		}
		records = int(result.RowsAffected)
		currentOffset += records
		if records > 0 {
			if result := tfrConf.Db().Create(&categories); result.Error != nil {
				log.Errorf("migrate: error in categories create %s", result.Error)
				return result.Error
			}
		}
	}
	log.Infof("migrate: number of categories transfered %v", currentOffset)

	// Gorm bug with composite foreign keys prevents the following from working.
	// It can be used once pull https://github.com/go-gorm/gorm/pull/7453 (or equivalent) has been made to gorm
	// var categorys []entity.Category
	// result = conf.Db().Unscoped().
	// 	FindInBatches(&categorys, batchSize, func(tx *gorm.DB, batch int) error {
	// 		var newCategorys []*entity.Category
	// 		for _, category := range categorys {
	// 			newCategorys = append(newCategorys, &category)
	// 		}
	// 		if result := tfrConf.Db().Create(newCategorys); result.Error != nil {
	// 			log.Errorf("migrate: error in batch category create %s", result.Error)
	// 			return result.Error
	// 		}
	// 		return nil
	// 	})
	// if result.Error != nil {
	// 	log.Errorf("migrate: error in batch category findinbatches %s", result.Error)
	// 	return result.Error
	// } else {
	// 	log.Infof("migrate: number of categories transfered %v", result.RowsAffected)
	// }

	var duplicates []entity.Duplicate
	records = batchSize
	currentOffset = 0
	for records == batchSize {
		if result = conf.Db().Unscoped().
			Model(&entity.Duplicate{}).
			Limit(batchSize).Offset(currentOffset).
			Order("file_name, file_root").
			Find(&duplicates); result.Error != nil {
			log.Errorf("migrate: error in duplicates find %s", result.Error)
			return result.Error
		}
		records = int(result.RowsAffected)
		currentOffset += records
		if records > 0 {
			if result := tfrConf.Db().Create(&duplicates); result.Error != nil {
				log.Errorf("migrate: error in duplicates create %s", result.Error)
				return result.Error
			}
		}
	}
	log.Infof("migrate: number of duplicates transfered %v", currentOffset)

	// Gorm bug with composite foreign keys prevents the following from working.
	// It can be used once pull https://github.com/go-gorm/gorm/pull/7453 (or equivalent) has been made to gorm
	// var duplicates entity.Duplicates
	// result = conf.Db().Unscoped().
	// 	FindInBatches(&duplicates, batchSize, func(tx *gorm.DB, batch int) error {
	// 		var newDuplicates []*entity.Duplicate
	// 		for _, duplicate := range duplicates {
	// 			newDuplicates = append(newDuplicates, &duplicate)
	// 		}
	// 		if result := tfrConf.Db().Create(newDuplicates); result.Error != nil {
	// 			log.Errorf("migrate: error in batch duplicate create %s", result.Error)
	// 			return result.Error
	// 		}
	// 		return nil
	// 	})
	// if result.Error != nil {
	// 	return result.Error
	// } else {
	// 	log.Infof("migrate: number of duplicates transfered %v", result.RowsAffected)
	// }

	var errors entity.Errors
	result = conf.Db().Unscoped().
		FindInBatches(&errors, batchSize, func(tx *gorm.DB, batch int) error {
			var newErrors []*entity.Error
			for _, error := range errors {
				newErrors = append(newErrors, &error)
			}
			if result := tfrConf.Db().
				Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "id"}},
					UpdateAll: true,
				}).
				Create(newErrors); result.Error != nil {
				log.Errorf("migrate: error in batch error create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of errors transfered %v", result.RowsAffected)
		if result.RowsAffected > 0 {
			if result = tfrConf.Db().Unscoped().
				Model(&entity.Error{}).Select("MAX(id) as id").Scan(&currentid); result.Error != nil {
				log.Errorf("migrate: error in getting max id %v", result.Error)
				return result.Error
			} else {
				resetIDToValue(tfrConf.Db(), entity.Error{}.TableName(), int(currentid.Id))
			}
		}
	}

	var faces entity.Faces
	result = conf.Db().Unscoped().
		FindInBatches(&faces, batchSize, func(tx *gorm.DB, batch int) error {
			var newFaces []*entity.Face
			for _, face := range faces {
				newFaces = append(newFaces, &face)
			}
			if result := tfrConf.Db().Create(newFaces); result.Error != nil {
				log.Errorf("migrate: error in batch face create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of faces transfered %v", result.RowsAffected)
	}

	var services entity.Services
	result = conf.Db().Unscoped().
		FindInBatches(&services, batchSize, func(tx *gorm.DB, batch int) error {
			var newServices []*entity.Service
			for _, service := range services {
				newServices = append(newServices, &service)
			}
			if result := tfrConf.Db().Create(newServices); result.Error != nil {
				log.Errorf("migrate: error in batch service create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of services transfered %v", result.RowsAffected)
		if result.RowsAffected > 0 {
			if result = tfrConf.Db().Unscoped().
				Model(&entity.Service{}).Select("MAX(id) as id").Scan(&currentid); result.Error != nil {
				log.Errorf("migrate: error in getting max id %v", result.Error)
				return result.Error
			} else {
				resetIDToValue(tfrConf.Db(), entity.Service{}.TableName(), int(currentid.Id))
			}
		}
	}

	var fileshares []entity.FileShare
	records = batchSize
	currentOffset = 0
	for records == batchSize {
		if result = conf.Db().Unscoped().
			Model(&entity.FileShare{}).
			Limit(batchSize).Offset(currentOffset).
			Order("file_id, service_id").
			Find(&fileshares); result.Error != nil {
			log.Errorf("migrate: error in fileshares find %s", result.Error)
			return result.Error
		}
		records = int(result.RowsAffected)
		currentOffset += records
		if records > 0 {
			if result := tfrConf.Db().Create(&fileshares); result.Error != nil {
				log.Errorf("migrate: error in fileshares create %s", result.Error)
				return result.Error
			}
		}
	}
	log.Infof("migrate: number of fileshares transfered %v", currentOffset)

	// Gorm bug with composite foreign keys prevents the following from working.
	// It can be used once pull https://github.com/go-gorm/gorm/pull/7453 (or equivalent) has been made to gorm
	// var fileshares []entity.FileShare
	// result = conf.Db().Unscoped().
	// 	FindInBatches(&fileshares, batchSize, func(tx *gorm.DB, batch int) error {
	// 		var newFileShares []*entity.FileShare
	// 		for _, fileshare := range fileshares {
	// 			newFileShares = append(newFileShares, &fileshare)
	// 		}
	// 		if result := tfrConf.Db().Create(newFileShares); result.Error != nil {
	// 			log.Errorf("migrate: error in batch fileshare create %s", result.Error)
	// 			return result.Error
	// 		}
	// 		return nil
	// 	})
	// if result.Error != nil {
	// 	return result.Error
	// } else {
	// 	log.Infof("migrate: number of fileshares transfered %v", result.RowsAffected)
	// }

	var filesyncs []entity.FileSync
	records = batchSize
	currentOffset = 0
	for records == batchSize {
		if result = conf.Db().Unscoped().
			Model(&entity.FileSync{}).
			Limit(batchSize).Offset(currentOffset).
			Order("remote_name, service_id").
			Find(&filesyncs); result.Error != nil {
			log.Errorf("migrate: error in filesyncs find %s", result.Error)
			return result.Error
		}
		records = int(result.RowsAffected)
		currentOffset += records
		if records > 0 {
			if result := tfrConf.Db().Create(&filesyncs); result.Error != nil {
				log.Errorf("migrate: error in filesyncs create %s", result.Error)
				return result.Error
			}
		}
	}
	log.Infof("migrate: number of filesyncs transfered %v", currentOffset)

	// Gorm bug with composite foreign keys prevents the following from working.
	// It can be used once pull https://github.com/go-gorm/gorm/pull/7453 (or equivalent) has been made to gorm
	// var filesyncs []entity.FileSync
	// result = conf.Db().Unscoped().
	// 	FindInBatches(&filesyncs, batchSize, func(tx *gorm.DB, batch int) error {
	// 		var newFileSyncs []*entity.FileSync
	// 		for _, filesync := range filesyncs {
	// 			newFileSyncs = append(newFileSyncs, &filesync)
	// 		}
	// 		if result := tfrConf.Db().Create(newFileSyncs); result.Error != nil {
	// 			log.Errorf("migrate: error in batch filesync create %s", result.Error)
	// 			return result.Error
	// 		}
	// 		return nil
	// 	})
	// if result.Error != nil {
	// 	return result.Error
	// } else {
	// 	log.Infof("migrate: number of filesyncs transfered %v", result.RowsAffected)
	// }

	var folders entity.Folders
	result = conf.Db().Unscoped().
		FindInBatches(&folders, batchSize, func(tx *gorm.DB, batch int) error {
			var newFolders []*entity.Folder
			for _, folder := range folders {
				newFolders = append(newFolders, &folder)
			}
			if result := tfrConf.Db().Create(newFolders); result.Error != nil {
				log.Errorf("migrate: error in batch folder create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of folders transfered %v", result.RowsAffected)
	}

	var links entity.Links
	result = conf.Db().Unscoped().
		FindInBatches(&links, batchSize, func(tx *gorm.DB, batch int) error {
			var newLinks []*entity.Link
			for _, link := range links {
				newLinks = append(newLinks, &link)
			}
			if result := tfrConf.Db().Create(newLinks); result.Error != nil {
				log.Errorf("migrate: error in batch link create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of links transfered %v", result.RowsAffected)
	}

	var markers entity.Markers
	result = conf.Db().Unscoped().
		FindInBatches(&markers, batchSize, func(tx *gorm.DB, batch int) error {
			var newMarkers []*entity.Marker
			for _, marker := range markers {
				newMarkers = append(newMarkers, &marker)
			}
			if result := tfrConf.Db().Create(newMarkers); result.Error != nil {
				log.Errorf("migrate: error in batch marker create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of markers transfered %v", result.RowsAffected)
	}

	var passcodes []entity.Passcode
	records = batchSize
	currentOffset = 0
	for records == batchSize {
		if result = conf.Db().Unscoped().
			Model(&entity.Passcode{}).
			Limit(batchSize).Offset(currentOffset).
			Order("uid, key_type").
			Find(&passcodes); result.Error != nil {
			log.Errorf("migrate: error in passcodes find %s", result.Error)
			return result.Error
		}
		records = int(result.RowsAffected)
		currentOffset += records
		if records > 0 {
			if result := tfrConf.Db().Create(&passcodes); result.Error != nil {
				log.Errorf("migrate: error in passcodes create %s", result.Error)
				return result.Error
			}
		}
	}
	log.Infof("migrate: number of passcodes transfered %v", currentOffset)

	// Gorm bug with composite foreign keys prevents the following from working.
	// It can be used once pull https://github.com/go-gorm/gorm/pull/7453 (or equivalent) has been made to gorm
	// var passcodes []entity.Passcode
	// result = conf.Db().Unscoped().
	// 	FindInBatches(&passcodes, batchSize, func(tx *gorm.DB, batch int) error {
	// 		var newPasscodes []*entity.Passcode
	// 		for _, passcode := range passcodes {
	// 			newPasscodes = append(newPasscodes, &passcode)
	// 		}
	// 		if result := tfrConf.Db().Create(newPasscodes); result.Error != nil {
	// 			log.Errorf("migrate: error in batch passcode create %s", result.Error)
	// 			return result.Error
	// 		}
	// 		return nil
	// 	})
	// if result.Error != nil {
	// 	return result.Error
	// } else {
	// 	log.Infof("migrate: number of passcodes transfered %v", result.RowsAffected)
	// }

	var passwords []entity.Password
	result = conf.Db().Unscoped().
		FindInBatches(&passwords, batchSize, func(tx *gorm.DB, batch int) error {
			var newPasswords []*entity.Password
			for _, password := range passwords {
				newPasswords = append(newPasswords, &password)
			}
			if result := tfrConf.Db().Create(newPasswords); result.Error != nil {
				log.Errorf("migrate: error in batch password create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of passwords transfered %v", result.RowsAffected)
	}

	var photousers []entity.PhotoUser
	records = batchSize
	currentOffset = 0
	for records == batchSize {
		if result = conf.Db().Unscoped().
			Model(&entity.PhotoUser{}).
			Limit(batchSize).Offset(currentOffset).
			Order("uid, user_uid").
			Find(&photousers); result.Error != nil {
			log.Errorf("migrate: error in photousers find %s", result.Error)
			return result.Error
		}
		records = int(result.RowsAffected)
		currentOffset += records
		if records > 0 {
			if result := tfrConf.Db().Create(&photousers); result.Error != nil {
				log.Errorf("migrate: error in photousers create %s", result.Error)
				return result.Error
			}
		}
	}
	log.Infof("migrate: number of photousers transfered %v", currentOffset)

	// Gorm bug with composite foreign keys prevents the following from working.
	// It can be used once pull https://github.com/go-gorm/gorm/pull/7453 (or equivalent) has been made to gorm
	// var photousers []entity.PhotoUser
	// result = conf.Db().Unscoped().
	// 	FindInBatches(&photousers, batchSize, func(tx *gorm.DB, batch int) error {
	// 		var newPhotoUsers []*entity.PhotoUser
	// 		for _, photouser := range photousers {
	// 			newPhotoUsers = append(newPhotoUsers, &photouser)
	// 		}
	// 		if result := tfrConf.Db().Create(newPhotoUsers); result.Error != nil {
	// 			log.Errorf("migrate: error in batch photouser create %s", result.Error)
	// 			return result.Error
	// 		}
	// 		return nil
	// 	})
	// if result.Error != nil {
	// 	return result.Error
	// } else {
	// 	log.Infof("migrate: number of photousers transfered %v", result.RowsAffected)
	// }

	var reactions []entity.Reaction
	records = batchSize
	currentOffset = 0
	for records == batchSize {
		if result = conf.Db().Unscoped().
			Model(&entity.Reaction{}).
			Limit(batchSize).Offset(currentOffset).
			Order("uid, user_uid, reaction").
			Find(&reactions); result.Error != nil {
			log.Errorf("migrate: error in reactions find %s", result.Error)
			return result.Error
		}
		records = int(result.RowsAffected)
		currentOffset += records
		if records > 0 {
			if result := tfrConf.Db().Create(&reactions); result.Error != nil {
				log.Errorf("migrate: error in reactions create %s", result.Error)
				return result.Error
			}
		}
	}
	log.Infof("migrate: number of reactions transfered %v", currentOffset)

	// Gorm bug with composite foreign keys prevents the following from working.
	// It can be used once pull https://github.com/go-gorm/gorm/pull/7453 (or equivalent) has been made to gorm
	// var reactions []entity.Reaction
	// result = conf.Db().Unscoped().
	// 	FindInBatches(&reactions, batchSize, func(tx *gorm.DB, batch int) error {
	// 		var newReactions []*entity.Reaction
	// 		for _, reaction := range reactions {
	// 			newReactions = append(newReactions, &reaction)
	// 		}
	// 		if result := tfrConf.Db().Create(newReactions); result.Error != nil {
	// 			log.Errorf("migrate: error in batch reaction create %s", result.Error)
	// 			return result.Error
	// 		}
	// 		return nil
	// 	})
	// if result.Error != nil {
	// 	return result.Error
	// } else {
	// 	log.Infof("migrate: number of reactions transfered %v", result.RowsAffected)
	// }

	var subjects entity.Subjects
	result = conf.Db().Unscoped().
		FindInBatches(&subjects, batchSize, func(tx *gorm.DB, batch int) error {
			var newSubjects []*entity.Subject
			for _, subject := range subjects {
				newSubjects = append(newSubjects, &subject)
			}
			if result := tfrConf.Db().Create(newSubjects); result.Error != nil {
				log.Errorf("migrate: error in batch subject create %s", result.Error)
				return result.Error
			}
			return nil
		})
	if result.Error != nil {
		return result.Error
	} else {
		log.Infof("migrate: number of subjects transfered %v", result.RowsAffected)
	}

	elapsed := time.Since(start)

	log.Infof("completed in %s", elapsed)

	return nil
}

// Reset the ID increment to passed in value
func resetIDToValue(db *gorm.DB, tableName string, value int) error {
	sqlCommand := ""
	switch db.Dialector.Name() {
	case entity.MySQL:
		sqlCommand = fmt.Sprintf("ALTER TABLE `%v` AUTO_INCREMENT = %d", tableName, value+1)
	case entity.Postgres:
		sqlCommand = fmt.Sprintf("ALTER SEQUENCE %v_id_seq RESTART WITH %d", tableName, value+1)
	case entity.SQLite3:
		sqlCommand = fmt.Sprintf("UPDATE SQLITE_SEQUENCE SET SEQ=%d WHERE NAME='%v'", value, tableName)
	default:
		return fmt.Errorf("Unsupported Dialector %s", db.Dialector.Name())
	}
	if res := db.Exec(sqlCommand); res.Error != nil {
		return fmt.Errorf("Reset Auto Increment failed with %v", res.Error)
	}
	return nil
}
