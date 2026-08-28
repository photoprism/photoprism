package performancetest

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

var characterNumberRune = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randomSHA1() string {
	result := make([]rune, 32)
	for i := range result {
		result[i] = characterNumberRune[rand.IntN(len(characterNumberRune))] //nolint:gosec // test data generation crypto rand not required
	}
	return string(result)
}

func generateDatabase(numberOfPhotos int, driver string, dsname string, dropdb bool, databasescript bool) error {
	generateDSN := dsn.Parse(dsname)
	dbname := generateDSN.Name
	_, admindsn := dsn.PhotoPrismDriverToDriverDSN(driver)
	if dropdb {
		switch driver {
		case dsn.DriverMySQL:
			log.Infof("Connecting to %v", admindsn)
			database, err := gorm.Open(mysql.Open(admindsn), &gorm.Config{})
			if err != nil {
				log.Errorf("Unable to connect to MariaDB %v", err)
				return err
			}
			log.Infof("Dropping database %v if it exists", dbname)
			if res := database.Exec("DROP DATABASE IF EXISTS " + dbname + ";"); res.Error != nil {
				log.Errorf("Unable to drop database %v", res.Error)
				return res.Error
			}
			log.Infof("Creating database %v if it doesnt exist", dbname)
			if res := database.Exec("CREATE DATABASE IF NOT EXISTS " + dbname + " CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"); res.Error != nil {
				log.Errorf("Unable to create database %v", res.Error)
				return res.Error
			}
			log.Infof("Granting permissions to database %v", dbname)
			str := fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%%';", dbname, generateDSN.User)
			if res := database.Exec(str); res.Error != nil {
				log.Errorf("Unable to create database %v", res.Error)
				return res.Error
			}
		case dsn.DriverPostgres:
			log.Infof("Connecting to %v", admindsn)
			database, err := gorm.Open(postgres.Open(admindsn), &gorm.Config{})
			if err != nil {
				log.Errorf("Unable to connect to Postgres %v", err)
				return err
			}
			log.Infof("Dropping database %v if it exists", dbname)
			if res := database.Exec("DROP DATABASE IF EXISTS " + dbname + ";"); res.Error != nil {
				log.Errorf("Unable to drop database %v", res.Error)
				return res.Error
			}
			log.Infof("Creating database %v if it doesnt exist", dbname)
			if res := database.Exec("CREATE DATABASE " + dbname + " OWNER " + generateDSN.User + ";"); res.Error != nil {
				log.Errorf("Unable to create database %v", res.Error)
				return res.Error
			}
		case dsn.DriverSQLite3:
			fallthrough
		default:
			driver = dsn.DriverSQLite3
			filename := dsname
			if strings.Index(dsname, "?") > 0 {
				if strings.Index(dsname, ":") > 0 {
					filename = dsname[strings.Index(dsname, ":")+1 : strings.Index(dsname, "?")]
				} else {
					filename = dsname[0:strings.Index(dsname, "?")]
				}
			}
			log.Infof("Removing file %v", filename)
			_ = os.Remove(filename)
		}
	}

	log.Infof("Connecting to driver %v with dsn %v", driver, dsname)
	// Create gorm.DB connection provider.
	db := &DbConn{
		Driver: driver,
		Dsn:    dsname,
	}
	defer db.Close()

	SetDbProvider(db)

	// Disable journal to speed up.
	if driver == dsn.DriverSQLite3 {
		Db().Exec("PRAGMA journal_mode=OFF")
	}

	start := time.Now()

	log.Info("Create PhotoPrism tables if they don't exist")
	// Run migration if the photos table doesn't exist.
	// Otherwise assume that we have a valid structured database.
	photoCounter := int64(0)
	log.Info("Accept Table/Relation doesn't exist error that may follow this.")
	if err := Db().Model(&entity.Photo{}).Count(&photoCounter).Error; err != nil {
		// Handle SQLite differently as it does table recreates on initial migrate, so we need to be able to simulate that.
		switch {
		case driver == dsn.DriverSQLite3 && databasescript:
			filename := dsname
			if strings.Index(dsname, "?") > 0 {
				if strings.Index(dsname, ":") > 0 {
					filename = dsname[strings.Index(dsname, ":")+1 : strings.Index(dsname, "?")]
				} else {
					filename = dsname[0:strings.Index(dsname, "?")]
				}
			}

			var cmd *exec.Cmd

			bashCmd := fmt.Sprintf("cat ./testdata/sqlite3.sql | sqlite3 %s", filename)

			cmd = exec.Command("bash", "-c", bashCmd) //nolint:gosec // test generated input

			// Write to stdout or file.
			var f *os.File
			log.Infof("restore: creating database tables from script")
			f = os.Stdout
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			cmd.Stdout = f

			// Log exact command for debugging in trace mode.
			log.Debug(cmd.String())

			// Run restore command.
			if cmdErr := cmd.Run(); cmdErr != nil {
				if errStr := strings.TrimSpace(stderr.String()); errStr != "" {
					log.Error(errStr)
					return fmt.Errorf("%s", errStr)
				}
			}
		case driver == dsn.DriverMySQL && databasescript:
			// Prepare migrate mariadb db.
			if dumpName, err := filepath.Abs("./testdata/mariadb.sql"); err != nil {
				log.Error(err)
				return err
			} else if err = exec.Command("mariadb", "-u", "migrate", "-pmigrate", dbname, //nolint:gosec // generated command string
				"-e", "source "+dumpName).Run(); err != nil {
				log.Error(err)
				return err
			}
		case driver == dsn.DriverPostgres && databasescript:
			// Prepare migrate postgres db
			if dumpName, err := filepath.Abs("./testdata/postgres.sql"); err != nil {
				log.Error(err)
				return err
			} else if err = exec.Command("psql", generateDSN.ForPSQL(), "-f", dumpName).Run(); err != nil { //nolint:gosec // G204 generated command string
				log.Error(err)
				return err
			}

		default:
			entity.Entities.Migrate(Db(), migrate.Opt(true, false, nil))
			if err := entity.Entities.WaitForMigration(Db()); err != nil {
				log.Errorf("migrate: %s [%s]", err, time.Since(start))
			}
		}
	} else {
		log.Errorf("The photos table already exists in driver %v dsn %v.\nAborting...", driver, dsname)
		return fmt.Errorf("the photos table already exists in driver %v dsn %v", driver, dsname)
	}

	entity.SetDbProvider(db)

	log.Info("Create default fixtures")

	entity.CreateDefaultFixtures()

	// Load the database with data.

	// Create all the labels and keywords that have specific handling in internal/ai/classify/rules.go
	log.Info("Create labels and keywords")
	keywords := make(map[string]uint)
	labels := make(map[string]uint)
	keywordRandoms := make(map[int]uint)
	labelRandoms := make(map[int]uint)
	keywordPos, labelPos := 0, 0
	for label, rule := range classify.Rules {
		keyword := entity.Keyword{
			Keyword: label,
			Skip:    false,
		}
		if err := Db().Create(&keyword).Error; err != nil {
			return err
		}
		keywords[label] = keyword.ID
		keywordRandoms[keywordPos] = keyword.ID
		keywordPos++
		if rule.Label != "" {
			if _, found := keywords[rule.Label]; !found {
				keyword = entity.Keyword{
					Keyword: rule.Label,
					Skip:    false,
				}
				if err := Db().Create(&keyword).Error; err != nil {
					return err
				}
				keywords[rule.Label] = keyword.ID
				keywordRandoms[keywordPos] = keyword.ID
				keywordPos++
			}
			for _, category := range rule.Categories {
				if _, found := labels[category]; !found {
					labelDb := entity.Label{
						LabelSlug:        strings.ToLower(category),
						CustomSlug:       strings.ToLower(category),
						LabelName:        strings.ToLower(category),
						LabelPriority:    0,
						LabelFavorite:    false,
						LabelDescription: "",
						LabelNotes:       "",
						PhotoCount:       0,
						LabelCategories:  []*entity.Label{},
						CreatedAt:        time.Now().UTC(),
						UpdatedAt:        time.Now().UTC(),
						DeletedAt:        gorm.DeletedAt{},
						New:              false,
					}
					if err := Db().Create(&labelDb).Error; err != nil {
						return err
					}
					labels[category] = labelDb.ID
					labelRandoms[labelPos] = labelDb.ID
					labelPos++
				}
			}
			if _, found := labels[rule.Label]; !found {
				labelDb := entity.Label{
					LabelSlug:        strings.ToLower(rule.Label),
					CustomSlug:       strings.ToLower(rule.Label),
					LabelName:        strings.ToLower(rule.Label),
					LabelPriority:    0,
					LabelFavorite:    false,
					LabelDescription: "",
					LabelNotes:       "",
					PhotoCount:       0,
					LabelCategories:  []*entity.Label{},
					CreatedAt:        time.Now().UTC(),
					UpdatedAt:        time.Now().UTC(),
					DeletedAt:        gorm.DeletedAt{},
					New:              false,
				}
				if err := Db().Create(&labelDb).Error; err != nil {
					return err
				}
				labels[rule.Label] = labelDb.ID
				labelRandoms[labelPos] = labelDb.ID
				labelPos++
				for _, category := range rule.Categories {
					categoryDb := entity.Category{
						LabelID:    labelDb.ID,
						CategoryID: labels[category],
					}
					if err := Db().Create(&categoryDb).Error; err != nil {
						return err
					}
				}
			}
		}
	}

	// Create every possible camera and some lenses.  Yeah the data is garbage but it's test data anyway.
	log.Info("Create cameras and lenses")
	lensList := [6]string{"Wide Angle", "Fisheye", "Ultra Wide Angle", "Macro", "Super Zoom", "F80"}
	cameras := make(map[string]uint)
	lenses := make(map[string]uint)
	cameraRandoms := make(map[int]uint)
	lensRandoms := make(map[int]uint)
	cameraPos, lensPos := 0, 0

	for _, make := range entity.CameraMakes {
		for _, model := range entity.CameraModels {
			camera := entity.NewCamera(make, model)
			if _, found := cameras[camera.CameraSlug]; !found {
				if err := Db().Create(camera).Error; err != nil {
					log.Errorf("generatedatabase: camera create of %v failed with %v", camera, err)
					return err
				}
				cameras[camera.CameraSlug] = camera.ID
				cameraRandoms[cameraPos] = camera.ID
				cameraPos++
			}
		}
		for _, model := range lensList {
			lens := entity.NewLens(make, model)
			if _, found := lenses[lens.LensSlug]; !found {
				if err := Db().Create(lens).Error; err != nil {
					return err
				}
				lenses[lens.LensSlug] = lens.ID
				lensRandoms[lensPos] = lens.ID
				lensPos++
			}
		}
	}

	// Load up Countries and Places.
	log.Info("Create countries and places")
	countries := make(map[int]string)
	countryPos := 0
	places := make(map[int]string)
	placePos := 0

	PlaceUID := byte('P')

	file, _ := os.Open("../../pkg/txt/resources/countries.txt")
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")

		if len(parts) < 2 {
			continue
		}

		country := entity.NewCountry(strings.ToLower(parts[0]), strings.ToLower(parts[1]))
		counter := int64(0)
		Db().Model(&entity.Country{}).Where("id = ?", country.ID).Count(&counter)
		if counter == 0 {
			if err := Db().Create(country).Error; err != nil {
				return err
			}
			countries[countryPos] = strings.ToLower(parts[0])
			countryPos++
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	for word := range txt.StopWords {
		placeUID := rnd.GenerateUID(PlaceUID)
		country := countries[rand.IntN(len(countries))] //nolint:gosec // test data generation crypto rand not required
		place := entity.Place{
			ID:            placeUID,
			PlaceLabel:    word,
			PlaceDistrict: word,
			PlaceCity:     word,
			PlaceState:    word,
			PlaceCountry:  country,
			PlaceKeywords: "",
			PlaceFavorite: false,
			PhotoCount:    0,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}
		if err := Db().Create(&place).Error; err != nil {
			return err
		}
		places[placePos] = placeUID
		placePos++
	}

	// Create some Subjects
	log.Info("Create subjects")
	subjects := make(map[int]entity.Subject)
	subjectPos := 0

	for i := 1; i <= 100; i++ {
		subject := entity.Subject{
			SubjUID:      rnd.GenerateUID('j'),
			SubjType:     entity.SubjPerson,
			SubjSrc:      entity.SrcImage,
			SubjSlug:     fmt.Sprintf("person-%03d", i),
			SubjName:     fmt.Sprintf("Person %03d", i),
			SubjFavorite: false,
			SubjPrivate:  false,
			SubjExcluded: false,
			FileCount:    0,
			PhotoCount:   0,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
			DeletedAt:    gorm.DeletedAt{},
		}
		if err := Db().Create(&subject).Error; err != nil {
			return err
		}
		subjects[subjectPos] = subject
		subjectPos++
	}

	numberOfFaces := int(0.75 * float32(numberOfPhotos))
	sourceEmbeddings := make(face.Embeddings, numberOfFaces)
	jsonembed := make([]float32, 512)
	embeddings := make(face.Embeddings, 1)

	for i := range numberOfFaces {
		for k := range 512 {
			if rand.IntN(2) == 0 { //nolint:gosec // test data generation crypto rand not required
				jsonembed[k] = rand.Float32() //nolint:gosec // test data generation crypto rand not required
			} else {
				jsonembed[k] = rand.Float32() * -1.0 //nolint:gosec // test data generation crypto rand not required
			}
		}
		sourceEmbeddings[i] = face.NewEmbedding(jsonembed)
	}

	log.Info("Start creating photos")
	for i := 1; i <= numberOfPhotos; i++ {
		if _, frac := math.Modf(float64(i) / 100.0); frac == 0 {
			log.Infof("Generating photo number %v", i)
		}
		month := rand.IntN(11) + 1                                                                                                 //nolint:gosec // test data generation crypto rand not required
		day := rand.IntN(28) + 1                                                                                                   //nolint:gosec // test data generation crypto rand not required
		year := rand.IntN(45) + 1980                                                                                               //nolint:gosec // test data generation crypto rand not required
		takenAt := time.Date(year, time.Month(month), day, rand.IntN(24), rand.IntN(60), rand.IntN(60), rand.IntN(1000), time.UTC) //nolint:gosec // test data generation crypto rand not required
		labelCount := rand.IntN(5)                                                                                                 //nolint:gosec // test data generation crypto rand not required

		// Create the cell for the Photo's location
		placeId := places[rand.IntN(len(places))] //nolint:gosec // test data generation crypto rand not required
		lat := (rand.Float64() * 180.0) - 90.0    //nolint:gosec // test data generation crypto rand not required
		lng := (rand.Float64() * 360.0) - 180.0   //nolint:gosec // test data generation crypto rand not required
		cell := entity.NewCell(lat, lng)
		cell.PlaceID = placeId
		if err := Db().FirstOrCreate(cell).Error; err != nil {
			return err
		}

		folder := entity.Folder{}
		if res := Db().Model(&entity.Folder{}).Where("path = ?", fmt.Sprintf("%04d", year)).First(&folder); res.RowsAffected == 0 {
			folder = entity.NewFolder("/", fmt.Sprintf("%04d", year), time.Now().UTC())
			if err := folder.Create(); err != nil {
				return err
			}
		}
		folder = entity.Folder{}
		if res := Db().Model(&entity.Folder{}).Where("path = ?", fmt.Sprintf("%04d/%02d", year, month)).First(&folder); res.RowsAffected == 0 {
			folder = entity.NewFolder("/", fmt.Sprintf("%04d/%02d", year, month), time.Now().UTC())
			if err := folder.Create(); err != nil {
				return err
			}
		}

		photo := entity.Photo{
			//	ID
			//
			// UUID
			TakenAt:          takenAt,
			TakenAtLocal:     takenAt,
			TakenSrc:         entity.SrcMeta,
			PhotoUID:         rnd.GenerateUID(entity.PhotoUID),
			PhotoType:        "image",
			TypeSrc:          entity.SrcAuto,
			PhotoTitle:       "Performance Test Load",
			TitleSrc:         entity.SrcImage,
			PhotoDescription: "",
			DescriptionSrc:   entity.SrcAuto,
			PhotoPath:        fmt.Sprintf("%04d/%02d", year, month),
			PhotoName:        fmt.Sprintf("PIC%08d", i),
			OriginalName:     fmt.Sprintf("PIC%08d", i),
			PhotoStack:       0,
			PhotoFavorite:    false,
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         "America/Mexico_City",
			PlaceID:          placeId,
			PlaceSrc:         entity.SrcMeta,
			CellID:           cell.ID,
			CellAccuracy:     0,
			PhotoAltitude:    5,
			PhotoLat:         lat,
			PhotoLng:         lng,
			PhotoCountry:     countries[rand.IntN(len(countries))], //nolint:gosec // test data generation crypto rand not required
			PhotoYear:        year,
			PhotoMonth:       month,
			PhotoDay:         day,
			PhotoIso:         400,
			PhotoExposure:    "1/60",
			PhotoFNumber:     8,
			PhotoFocalLength: 2,
			PhotoQuality:     3,
			PhotoFaces:       0,
			PhotoResolution:  0,
			// PhotoDuration    : 0,
			PhotoColor:   12,
			CameraID:     cameraRandoms[rand.IntN(len(cameraRandoms))], //nolint:gosec // test data generation crypto rand not required
			CameraSerial: "",
			CameraSrc:    "",
			LensID:       lensRandoms[rand.IntN(len(lensRandoms))], //nolint:gosec // test data generation crypto rand not required
			// Details          :,
			// Camera
			// Lens
			// Cell
			// Place
			Keywords: []entity.Keyword{},
			Albums:   []entity.Album{},
			Files:    []entity.File{},
			Labels:   []entity.PhotoLabel{},
			// CreatedBy
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			EditedAt:    nil,
			PublishedAt: nil,
			CheckedAt:   nil,
			EstimatedAt: nil,
			DeletedAt:   gorm.DeletedAt{},
		}
		if err := Db().Create(&photo).Error; err != nil {
			return err
		}
		// Allocate the labels for this photo
		for range labelCount {
			photoLabel := entity.NewPhotoLabel(photo.ID, labelRandoms[rand.IntN(len(labelRandoms))], 0, entity.SrcMeta) //nolint:gosec // test data generation crypto rand not required
			if err := Db().FirstOrCreate(photoLabel).Error; err != nil {
				return err
			}
		}
		// Allocate the keywords for this photo
		keywordCount := rand.IntN(5) //nolint:gosec // test data generation crypto rand not required
		keywordStr := ""
		for range keywordCount {
			photoKeyword := entity.PhotoKeyword{PhotoID: photo.ID, KeywordID: keywordRandoms[rand.IntN(len(keywordRandoms))]} //nolint:gosec // test data generation crypto rand not required
			keyword := entity.Keyword{}
			Db().Model(&entity.Keyword{}).Where("id = ?", photoKeyword.KeywordID).First(&keyword)
			if err := Db().FirstOrCreate(&photoKeyword).Error; err != nil {
				return err
			}
			if len(keywordStr) > 0 {
				keywordStr = fmt.Sprintf("%s,%s", keywordStr, keyword.Keyword)
			} else {
				keywordStr = keyword.Keyword
			}
		}

		// Create File
		file := entity.File{
			//	ID
			// Photo
			PhotoID:      photo.ID,
			PhotoUID:     photo.PhotoUID,
			PhotoTakenAt: photo.TakenAt,
			// TimeIndex
			// MediaID
			// MediaUTC
			InstanceID:   "",
			FileUID:      rnd.GenerateUID(entity.FileUID),
			FileName:     fmt.Sprintf("%04d/%02d/PIC%08d.jpg", year, month, i),
			FileRoot:     entity.RootSidecar,
			OriginalName: "",
			FileHash:     rnd.GenerateUID(entity.FileUID),
			FileSize:     rand.Int64N(1000000), //nolint:gosec // test data generation crypto rand not required
			FileCodec:    "",
			FileType:     string(fs.ImageJpeg),
			MediaType:    string(media.Image),
			FileMime:     "image/jpg",
			FilePrimary:  true,
			FileSidecar:  false,
			FileMissing:  false,
			FilePortrait: true,
			FileVideo:    false,
			FileDuration: 0,
			// FileFPS
			// FileFrames
			FileWidth:          1200,
			FileHeight:         1600,
			FileOrientation:    6,
			FileOrientationSrc: entity.SrcMeta,
			FileProjection:     "",
			FileAspectRatio:    0.75,
			// FileHDR            : false,
			// FileWatermark
			// FileColorProfile
			FileMainColor: "magenta",
			FileColors:    "226611CC1",
			FileLuminance: "ABCDEF123",
			FileDiff:      456,
			FileChroma:    15,
			// FileSoftware
			// FileError
			ModTime:   time.Now().Unix(),
			CreatedAt: time.Now().UTC(),
			CreatedIn: 935962,
			UpdatedAt: time.Now().UTC(),
			UpdatedIn: 935962,
			// PublishedAt
			DeletedAt: gorm.DeletedAt{},
			Share:     []entity.FileShare{},
			Sync:      []entity.FileSync{},
			// markers
		}
		if err := Db().Create(&file).Error; err != nil {
			return err
		}

		// Add Markers
		markersToCreate := rand.IntN(5) //nolint:gosec // test data generation crypto rand not required

		for range markersToCreate {
			subject := subjects[rand.IntN(len(subjects))] //nolint:gosec // test data generation crypto rand not required

			embeddings[0] = sourceEmbeddings[rand.IntN(numberOfFaces)] //nolint:gosec // test data generation crypto rand not required
			marker := entity.Marker{
				MarkerUID:  rnd.GenerateUID('m'),
				FileUID:    file.FileUID,
				MarkerType: entity.MarkerFace,
				MarkerSrc:  entity.SrcImage,
				// MarkerName:    subject.SubjName,
				MarkerReview:  false,
				MarkerInvalid: false,
				// SubjUID:       subject.SubjUID,
				SubjSrc:        entity.SrcAuto,
				FaceDist:       rand.Float64(), //nolint:gosec // test data generation crypto rand not required
				EmbeddingsJSON: embeddings.JSON(),
				X:              rand.Float32(), //nolint:gosec // test data generation crypto rand not required
				Y:              rand.Float32(), //nolint:gosec // test data generation crypto rand not required
				W:              rand.Float32(), //nolint:gosec // test data generation crypto rand not required
				H:              rand.Float32(), //nolint:gosec // test data generation crypto rand not required
				Size:           rand.IntN(600), //nolint:gosec // test data generation crypto rand not required
				Score:          rand.IntN(150), //nolint:gosec // test data generation crypto rand not required
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
			}
			if err := Db().Create(&marker).Error; err != nil {
				return err
			}
			face := entity.Face{
				ID:              randomSHA1(),
				FaceSrc:         entity.SrcImage,
				FaceKind:        1,
				FaceHidden:      false,
				SubjUID:         subject.SubjUID,
				Samples:         5,
				SampleRadius:    0.35,
				Collisions:      5,
				CollisionRadius: 0.5,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
			}
			if err := Db().Create(&face).Error; err != nil {
				return err
			}
		}

		// Add to Album
		albumSlug := fmt.Sprintf("my-photos-from-%04d", year)
		album := entity.Album{}
		if res := Db().Model(&entity.Album{}).Where("album_slug = ?", albumSlug).First(&album); res.RowsAffected == 0 {
			album = entity.Album{
				AlbumUID:         rnd.GenerateUID(entity.AlbumUID),
				AlbumSlug:        albumSlug,
				AlbumPath:        "",
				AlbumType:        entity.AlbumManual,
				AlbumTitle:       fmt.Sprintf("My Photos From %04d", year),
				AlbumLocation:    "",
				AlbumCategory:    "",
				AlbumCaption:     "",
				AlbumDescription: "A wonderful year",
				AlbumNotes:       "",
				AlbumFilter:      "",
				AlbumOrder:       "oldest",
				AlbumTemplate:    "",
				AlbumCountry:     entity.UnknownID,
				AlbumYear:        year,
				AlbumMonth:       0,
				AlbumDay:         0,
				AlbumFavorite:    false,
				AlbumPrivate:     false,
				CreatedAt:        time.Now().UTC(),
				UpdatedAt:        time.Now().UTC(),
				DeletedAt:        gorm.DeletedAt{},
			}
			if err := Db().Create(&album).Error; err != nil {
				return err
			}
		}
		photoAlbum := entity.PhotoAlbum{
			PhotoUID:  photo.PhotoUID,
			AlbumUID:  album.AlbumUID,
			Order:     0,
			Hidden:    false,
			Missing:   false,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		if err := Db().Create(photoAlbum).Error; err != nil {
			return err
		}

		details := entity.Details{
			PhotoID:     photo.ID,
			Keywords:    keywordStr,
			KeywordsSrc: entity.SrcMeta,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}
		if err := Db().Create(details).Error; err != nil {
			return err
		}
	}

	entity.File{}.RegenerateIndex()
	if err := entity.UpdateCounts(); err != nil {
		return err
	}

	log.Infof("Database Creation completed in %s", time.Since(start))
	return nil
}
