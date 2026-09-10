package performancetest

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
)

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
		var maxID uint
		entity.Db().Model(&entity.File{}).Pluck("max(ID)", &maxID)
		for b.Loop() {
			fileRegenerateIndex(b, int(maxID))
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

func fileRegenerateIndex(b *testing.B, maxID int) {
	fileId := uint(rand.IntN(maxID)) //nolint:gosec // test data generation crypto rand not required

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
