package search

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// createPlaceTestPhoto adds a photo with coordinates and a place of its own with the city and state.
// The rows are removed again so shared fixture counts other tests assert on stay stable.
func createPlaceTestPhoto(t *testing.T, city, state string) entity.Photo {
	t.Helper()

	place := entity.Place{
		ID:           rnd.UUIDv7(),
		PlaceLabel:   city + ", " + state,
		PlaceCity:    city,
		PlaceState:   state,
		PlaceCountry: "de",
	}

	require.NoError(t, entity.Db().Create(&place).Error)

	photo := entity.Photo{
		PhotoUID:     rnd.GenerateUID(entity.PhotoUID),
		PhotoName:    "20221119_120000_PLACE",
		PhotoPath:    "place",
		PhotoType:    entity.MediaImage,
		PhotoQuality: 3,
		PhotoLat:     52.52,
		PhotoLng:     13.40,
		PlaceID:      place.ID,
		PlaceSrc:     entity.SrcMeta,
		TakenAt:      entity.Now(),
		TakenAtLocal: entity.Now(),
		TakenSrc:     entity.SrcMeta,
	}

	require.NoError(t, entity.Db().Create(&photo).Error)

	file := entity.File{
		PhotoID:     photo.ID,
		PhotoUID:    photo.PhotoUID,
		FileUID:     rnd.GenerateUID(entity.FileUID),
		FileName:    "place/" + photo.PhotoUID + ".jpg",
		FileRoot:    entity.RootOriginals,
		FileHash:    rnd.UUIDv7(),
		FileType:    "jpg",
		MediaType:   media.Image.String(),
		FilePrimary: true,
	}

	require.NoError(t, entity.Db().Create(&file).Error)

	entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID}.RegenerateIndex()

	t.Cleanup(func() {
		entity.UnscopedDb().Where("photo_id = ?", photo.ID).Delete(&entity.File{})
		entity.UnscopedDb().Where("id = ?", photo.ID).Delete(&entity.Photo{})
		entity.UnscopedDb().Where("id = ?", place.ID).Delete(&entity.Place{})
	})

	return photo
}

// placeFilterUIDs returns the UIDs that the photo and the geo search find for a search form.
func placeFilterUIDs(t *testing.T, f form.SearchPhotos) (photoUIDs, geoUIDs []string) {
	t.Helper()

	f.Merged = true

	photos, _, err := Photos(f)
	require.NoError(t, err)

	for _, p := range photos {
		photoUIDs = append(photoUIDs, p.PhotoUID)
	}

	geo, err := PhotosGeo(form.SearchPhotosGeo{City: f.City, State: f.State, Query: f.Query})
	require.NoError(t, err)

	for _, p := range geo {
		geoUIDs = append(geoUIDs, p.PhotoUID)
	}

	return photoUIDs, geoUIDs
}

func TestPhotosFilterPlaceCase(t *testing.T) {
	berlin := createPlaceTestPhoto(t, "Berlin", "Baden-Württemberg")
	hamburg := createPlaceTestPhoto(t, "Hamburg", "Hamburg")
	literal := createPlaceTestPhoto(t, "Bad_Ort", "Rand%Land")

	// MariaDB's collation and PostgreSQL's ILIKE ignore case, while SQLite keeps comparing exactly.
	ignoresCase := entity.DbDialect() != dsn.DialectSQLite

	// expect asserts that both searches find exactly the wanted photos among the test photos.
	expect := func(t *testing.T, f form.SearchPhotos, want ...entity.Photo) {
		t.Helper()

		photoUIDs, geoUIDs := placeFilterUIDs(t, f)

		for _, p := range []entity.Photo{berlin, hamburg, literal} {
			wanted := false

			for _, w := range want {
				wanted = wanted || w.PhotoUID == p.PhotoUID
			}

			assert.Equal(t, wanted, slices.Contains(photoUIDs, p.PhotoUID), "photos %s", p.PhotoName)
			assert.Equal(t, wanted, slices.Contains(geoUIDs, p.PhotoUID), "geo %s", p.PhotoName)
		}
	}

	// folded returns the photos if the database ignores case, and none otherwise.
	folded := func(p ...entity.Photo) []entity.Photo {
		if ignoresCase {
			return p
		}

		return nil
	}

	t.Run("CityExact", func(t *testing.T) {
		expect(t, form.SearchPhotos{City: "Berlin"}, berlin)
		expect(t, form.SearchPhotos{Query: "city:Berlin"}, berlin)
	})
	t.Run("CityLowerUpperMixed", func(t *testing.T) {
		for _, v := range []string{"berlin", "BERLIN", "bErLiN"} {
			expect(t, form.SearchPhotos{City: v}, folded(berlin)...)
			expect(t, form.SearchPhotos{Query: "city:" + v}, folded(berlin)...)
		}
	})
	t.Run("StateExact", func(t *testing.T) {
		expect(t, form.SearchPhotos{State: "Baden-Württemberg"}, berlin)
		expect(t, form.SearchPhotos{Query: `state:"Baden-Württemberg"`}, berlin)
	})
	t.Run("StateLowerUpperMixed", func(t *testing.T) {
		for _, v := range []string{"baden-württemberg", "BADEN-WÜRTTEMBERG", "bAdEn-wÜrTtEmBeRg"} {
			expect(t, form.SearchPhotos{State: v}, folded(berlin)...)
			expect(t, form.SearchPhotos{Query: `state:"` + v + `"`}, folded(berlin)...)
		}
	})
	t.Run("OrList", func(t *testing.T) {
		expect(t, form.SearchPhotos{City: "Berlin|Hamburg"}, berlin, hamburg)
		expect(t, form.SearchPhotos{Query: `city:"Berlin|Hamburg"`}, berlin, hamburg)
		expect(t, form.SearchPhotos{State: "Baden-Württemberg | Hamburg"}, berlin, hamburg)
		expect(t, form.SearchPhotos{City: "berlin|Hamburg"}, append(folded(berlin), hamburg)...)
		expect(t, form.SearchPhotos{Query: `state:"BADEN-WÜRTTEMBERG|hamburg"`}, folded(berlin, hamburg)...)
	})
	t.Run("NoPrefixOrWildcardMatch", func(t *testing.T) {
		for _, v := range []string{"Berl", "berl", "Berl%", "Berl*", "Berl_n", "_", "B%n", "Baden%", "Baden*"} {
			expect(t, form.SearchPhotos{City: v})
			expect(t, form.SearchPhotos{State: v})
		}
	})
	t.Run("MaxSearchValues", func(t *testing.T) {
		// PostgreSQL compares only the first MaxSearchValues values; MariaDB and SQLite compare all of them.
		names := make([]string, MaxSearchValues)

		for i := range names {
			names[i] = fmt.Sprintf("Nowhere%d", i)
		}

		capped := entity.DbDialect() == dsn.DialectPostgreSQL
		last := strings.Join(append(names, "Berlin"), "|")
		first := strings.Join(append([]string{"Berlin"}, names...), "|")

		if capped {
			expect(t, form.SearchPhotos{City: last})
			expect(t, form.SearchPhotos{State: strings.Join(append(names, "Hamburg"), "|")})
		} else {
			expect(t, form.SearchPhotos{City: last}, berlin)
			expect(t, form.SearchPhotos{State: strings.Join(append(names, "Hamburg"), "|")}, hamburg)
		}

		expect(t, form.SearchPhotos{City: first}, berlin)
	})
	t.Run("LiteralWildcardCharacters", func(t *testing.T) {
		expect(t, form.SearchPhotos{City: "Bad_Ort"}, literal)
		expect(t, form.SearchPhotos{State: "Rand%Land"}, literal)
		expect(t, form.SearchPhotos{City: "bad_ort"}, folded(literal)...)
		expect(t, form.SearchPhotos{State: "RAND%LAND"}, folded(literal)...)
		expect(t, form.SearchPhotos{City: "BadXOrt"})
		expect(t, form.SearchPhotos{State: "RandXYLand"})
	})
}
