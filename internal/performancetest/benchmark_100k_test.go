package performancetest

import (
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

func Benchmark100k_SQLite(b *testing.B) {
	// Setup here
	loglevel := event.Log.GetLevel()
	event.Log.SetLevel(logrus.ErrorLevel)
	testDbOriginal := "../../storage/test-100k.original.sqlite"

	if !fs.FileExists(testDbOriginal) {
		log.Info("Generating SQLite database with 100000 records")
		require.NoError(b, generateDatabase(100000, "sqlite", testDbOriginal, true, true))
	}

	// Prepare temporary sqlite db.
	testDbTemp := "../../storage/test-100k.db"
	dumpName, err := filepath.Abs(testDbTemp)
	_ = os.Remove(dumpName)
	if err != nil {
		b.Fatal(err)
	} else if err = fs.Copy(testDbOriginal, dumpName, true); err != nil {
		b.Fatal(err)
	}
	defer os.Remove(dumpName)

	// Force the dbConn to nil so that a new database can be connected to.
	entity.SetDbProvider(nil)

	// Create gorm.DB connection provider.
	db := &entity.DbConn{
		Driver: "sqlite",
		Dsn:    dumpName,
	}

	// Insert test fixtures into the database.
	entity.SetDbProvider(db)

	entity.InitDb(migrate.Opt(true, false, nil))

	defer db.Close()

	// tests here

	runTests(b)

	// teardown here
	event.Log.SetLevel(loglevel)
}

func Benchmark100k_MySQL(b *testing.B) {
	// Setup here
	loglevel := event.Log.GetLevel()
	event.Log.SetLevel(logrus.ErrorLevel)
	testDbOriginal := "../../storage/test-100k.original.mysql"
	mysqlD := dsn.TestDSNFromEnv(dsn.DriverMariaDB, "migrate")
	mysqlDSN := mysqlD.ToString()

	// Prepare temporary mariadb db.
	if !fs.FileExists(testDbOriginal) {
		log.Info("Generating Mariadb database with 100000 records")
		require.NoError(b, generateDatabase(100000, "mysql", mysqlDSN, true, true))
		resultFile := "--result-file=" + testDbOriginal
		if err := exec.Command("mariadb-dump", "--user=migrate", "--password=migrate", "--lock-tables", "--add-drop-database", "--databases", "migrate", resultFile).Run(); err != nil { //nolint:gosec // test generated input, test only credentials
			b.Fatal(err)
		}
	}

	// Prepare migrate mariadb db.
	if dumpName, err := filepath.Abs(testDbOriginal); err != nil {
		b.Fatal(err)
	} else if err = exec.Command("mariadb", "-u", "migrate", "-pmigrate", "migrate", //nolint:gosec // test generated input, test only credentials
		"-e", "source "+dumpName).Run(); err != nil {
		b.Fatal(err)
	}

	// Force the dbConn to nil so that a new database can be connected to.
	entity.SetDbProvider(nil)

	// Create gorm.DB connection provider.
	db := &entity.DbConn{
		Driver: "mysql",
		Dsn:    mysqlDSN,
	}

	// Insert test fixtures into the database.
	entity.SetDbProvider(db)

	entity.InitDb(migrate.Opt(true, false, nil))

	defer db.Close()

	// tests here

	runTests(b)

	// teardown here
	event.Log.SetLevel(loglevel)
}

func Benchmark100k_PostgreSQL(b *testing.B) {
	// Setup here
	loglevel := event.Log.GetLevel()
	event.Log.SetLevel(logrus.ErrorLevel)
	testDbOriginal := "../../storage/test-100k.original.postgresql"
	oDSN := testextras.TestDbDSN(dsn.DriverPostgreSQL, "migrate")
	mDSN := dsn.Parse(oDSN)
	pDSN := mDSN
	pDSN.User = "photoprism"     //nolint:gosec // test only credentials
	pDSN.Password = "photoprism" //nolint:gosec // test only credentials
	pDSN.Name = "postgres"

	// Prepare temporary PostgreSQL db.
	if !fs.FileExists(testDbOriginal) {
		log.Info("Generating PostgreSQL database with 100000 records")
		require.NoError(b, generateDatabase(100000, Postgres, mDSN.ToString(), true, true))
		if err := exec.Command("pg_dump", "-d", mDSN.ForPSQL(), "-F c", "-f", testDbOriginal).Run(); err != nil { //nolint:gosec // test generated input
			b.Fatal(err)
		}
	}

	// Prepare migrate PostgreSQL db.
	if dumpName, err := filepath.Abs(testDbOriginal); err != nil {
		b.Fatal(err)
	} else if err = exec.Command("dropdb", fmt.Sprintf("--maintenance-db=%s", pDSN.ForPSQL()), "--force", "--if-exists", "migrate").Run(); err != nil { //nolint:gosec // test generated input, test only credentials
		b.Fatal(err)
	} else if err = exec.Command("createdb", fmt.Sprintf("--maintenance-db=%s", pDSN.ForPSQL()), "-O", "migrate", "-T", "template0", "migrate").Run(); err != nil { //nolint:gosec // test generated input, test only credentials
		b.Fatal(err)
	} else if err = exec.Command("pg_restore", "-d", mDSN.ForPSQL(), dumpName).Run(); err != nil { //nolint:gosec // test generated input, test only credentials
		b.Fatal(err)
	}

	// Force the dbConn to nil so that a new database can be connected to.
	entity.SetDbProvider(nil)

	// Create gorm.DB connection provider.
	db := &entity.DbConn{
		Driver: Postgres,
		Dsn:    mDSN.ToString(),
	}

	// Insert test fixtures into the database.
	entity.SetDbProvider(db)

	entity.InitDb(migrate.Opt(true, false, nil))

	defer db.Close()

	// tests here

	runTests(b)

	// teardown here
	event.Log.SetLevel(loglevel)
}

// The following is the tests being executed

func runTests(b *testing.B) {

	b.Run("CreateDeleteAlbum", func(b *testing.B) {
		for b.Loop() {
			createDeleteAlbum(b)
		}
	})

	b.Run("ListAlbums", func(b *testing.B) {
		for b.Loop() {
			listAlbums(b)
		}
	})

	b.Run("CreateDeleteCamera", func(b *testing.B) {
		for b.Loop() {
			createDeleteCamera(b)
		}
	})

	b.Run("CreateDeleteCellAndPlace", func(b *testing.B) {
		for b.Loop() {
			createDeleteCellAndPlace(b)
		}
	})

	b.Run("FileRegenerateIndex", func(b *testing.B) {
		for b.Loop() {
			fileRegenerateIndex(b)
		}
	})

	b.Run("CreateDeletePhoto", func(b *testing.B) {
		for b.Loop() {
			createDeletePhoto(b)
		}
	})

	b.Run("ListPhotos", func(b *testing.B) {
		for b.Loop() {
			listPhotos(b)
		}
	})

}

// The following are the functions to be executed by the tests above, so that PostgreSQL, MariaDB and SQLite can have the same test.

func createDeleteAlbum(b *testing.B) {
	album := entity.NewAlbum("BenchMarkAlbum", entity.AlbumManual)
	if err := album.Create(); err != nil {
		b.Fatal(err)
	}
	if err := album.DeletePermanently(); err != nil {
		b.Fatal(err)
	}
	entity.FlushAlbumCache()
}

func listAlbums(b *testing.B) {
	year := rand.IntN(45) + 1980 //nolint:gosec // test data generation crypto rand not required
	frm := form.SearchAlbums{
		Year: strconv.Itoa(year),
	}
	_, err := search.Albums(frm)
	if err != nil {
		b.Fatal(err)
	}
	entity.FlushAlbumCache()
	albumSlug := fmt.Sprintf("slug:my-photos-from-%04d", year)
	query := form.NewAlbumSearch(albumSlug)
	_, err = search.Albums(query)
	if err != nil {
		b.Fatal(err)
	}
	entity.FlushAlbumCache()
}

func createDeleteCamera(b *testing.B) {
	camera := entity.NewCamera("Palasonic", "Palasonic Dumix")

	if err := camera.Create(); err != nil {
		b.Fatal(err)
	}
	if err := entity.UnscopedDb().Delete(&camera).Error; err != nil {
		b.Fatal(err)
	}
	entity.FlushCameraCache()
}

func createDeleteCellAndPlace(b *testing.B) {
	lat := randRange(-90, 90)
	lng := randRange(-180, 180)
	cell := entity.NewCell(lat, lng)
	place := &entity.Place{
		ID:            randomString(12),
		PlaceLabel:    randomString(20),
		PlaceDistrict: randomString(30),
		PlaceCity:     randomString(30),
		PlaceState:    randomString(30),
		PlaceCountry:  randomString(2),
		PlaceKeywords: randomString(10),
		PlaceFavorite: false,
	}

	if cell.Place = entity.FirstOrCreatePlace(place); cell.Place == nil {
		b.Fatal("unable to find/create place")
	}

	cell.PlaceID = cell.Place.ID

	if entity.FirstOrCreateCell(cell) == nil {
		b.Fatal("unable to find/create cell")
	}
	if err := cell.Delete(); err != nil {
		b.Fatal(err)
	}

	if err := place.Delete(); err != nil {
		b.Fatal(err)
	}

}

func fileRegenerateIndex(b *testing.B) {
	fileId := uint(rand.IntN(100000)) //nolint:gosec // test data generation crypto rand not required

	file := entity.File{ID: fileId}
	require.NoError(b, entity.Db().First(&file).Error)

	file.RegenerateIndex()
}

func listPhotos(b *testing.B) {
	year := rand.IntN(45) + 1980 //nolint:gosec // test data generation crypto rand not required
	frm := form.SearchPhotos{
		Year: strconv.Itoa(year),
	}
	_, _, err := search.Photos(frm)
	if err != nil {
		b.Fatal(err)
	}
	albumSlug := fmt.Sprintf("slug:my-photos-from-%04d", year)
	var f form.SearchPhotos
	f.Query = ""
	f.Albums = albumSlug
	_, _, err = search.Photos(f)
	if err != nil {
		b.Fatal(err)
	}
}
