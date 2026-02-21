package performancetest

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const characterRunes = " abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const sha1Runes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randRange(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func randomString(len int) string {
	sb := strings.Builder{}
	sb.Grow(len)
	for i := 0; i < len; {
		sb.WriteByte(characterRunes[rand.IntN(53)])
		i++
	}

	return sb.String()
}

func createDeletePhoto(b *testing.B) {
	month := rand.IntN(11) + 1
	day := rand.IntN(28) + 1
	year := rand.IntN(45) + 1980
	takenAt := time.Date(year, time.Month(month), day, rand.IntN(24), rand.IntN(60), rand.IntN(60), rand.IntN(1000), time.UTC)
	labelCount := rand.IntN(5)

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

	if place := entity.FirstOrCreatePlace(place); place == nil {
		b.Fatal("unable to find/create place")
	}
	placeId := place.ID

	// Create the cell for the Photo's location
	lat := (rand.Float64() * 180.0) - 90.0
	lng := (rand.Float64() * 360.0) - 180.0
	cell := entity.NewCell(lat, lng)
	cell.PlaceID = placeId
	cell.Place = place
	entity.Db().FirstOrCreate(cell)

	folder := entity.Folder{}
	if res := entity.Db().Model(&entity.Folder{}).Where("path = ?", fmt.Sprintf("%04d", year)).First(&folder); res.RowsAffected == 0 {
		folder = entity.NewFolder("/", fmt.Sprintf("%04d", year), time.Now().UTC())
		folder.Create()
	}
	folder = entity.Folder{}
	if res := entity.Db().Model(&entity.Folder{}).Where("path = ?", fmt.Sprintf("%04d/%02d", year, month)).First(&folder); res.RowsAffected == 0 {
		folder = entity.NewFolder("/", fmt.Sprintf("%04d/%02d", year, month), time.Now().UTC())
		folder.Create()
	}

	camera := entity.NewCamera("Palasonic", "Palasonic Dumix")

	if err := camera.Create(); err != nil {
		b.Fatal(err)
	}

	lens := entity.NewLens("Palasonic", "Super Zoom")

	if err := lens.Create(); err != nil {
		b.Fatal(err)
	}

	i := rand.Int64N(60000)

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
		PhotoCountry:     "au",
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
		CameraID:     camera.ID,
		CameraSerial: "",
		CameraSrc:    "",
		LensID:       lens.ID,
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

	photo.Create()

	// Allocate the labels for this photo
	labels := make([]uint, labelCount)
	for i := range labelCount {
		label := entity.Label{
			LabelSlug:        strings.ToLower(randomString(32)),
			CustomSlug:       strings.ToLower(randomString(32)),
			LabelName:        strings.ToLower(randomString(32)),
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
		label.Create()
		labels[i] = label.ID

		photoLabel := entity.NewPhotoLabel(photo.ID, label.ID, 0, entity.SrcMeta)
		entity.Db().FirstOrCreate(photoLabel)
	}
	// Allocate the keywords for this photo
	keywordCount := rand.IntN(5)
	keywords := make([]uint, keywordCount)
	keywordStr := ""
	for i := range keywordCount {
		keyword := entity.Keyword{
			Keyword: randomString(32),
			Skip:    false,
		}
		keywords[i] = keyword.ID
		photoKeyword := entity.PhotoKeyword{PhotoID: photo.ID, KeywordID: keyword.ID}

		entity.Db().FirstOrCreate(&photoKeyword)
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
		FileSize:     rand.Int64N(1000000),
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
		//markers
	}
	entity.Db().Create(&file)

	// Add Markers
	markersToCreate := rand.IntN(5)
	subjects := make([]string, markersToCreate)
	for i := range markersToCreate {
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
		entity.Db().Create(&subject)
		subjects[i] = subject.SubjUID
		marker := entity.Marker{
			MarkerUID:     rnd.GenerateUID('m'),
			FileUID:       file.FileUID,
			MarkerType:    entity.MarkerFace,
			MarkerName:    subject.SubjName,
			MarkerReview:  false,
			MarkerInvalid: false,
			SubjUID:       subject.SubjUID,
			SubjSrc:       subject.SubjSrc,
			X:             rand.Float32() * 1024.0,
			Y:             rand.Float32() * 2048.0,
			W:             rand.Float32() * 10.0,
			H:             rand.Float32() * 20.0,
			Q:             10,
			Size:          100,
			Score:         10,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}
		entity.Db().Create(&marker)
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
		entity.Db().Create(&face)
	}

	// Add to Album
	albumSlug := fmt.Sprintf("my-photos-from-%04d", year)
	album := entity.Album{}
	if res := entity.Db().Model(&entity.Album{}).Where("album_slug = ?", albumSlug).First(&album); res.RowsAffected == 0 {
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
		entity.Db().Create(&album)
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
	entity.Db().Create(photoAlbum)

	details := entity.Details{
		PhotoID:     photo.ID,
		Keywords:    keywordStr,
		KeywordsSrc: entity.SrcMeta,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	entity.Db().Create(details)

	photo.DeletePermanently()

	if err := entity.UnscopedDb().Delete(lens).Error; err != nil {
		b.Fatal(err)
	}
	entity.FlushLensCache()

	if err := entity.UnscopedDb().Delete(camera).Error; err != nil {
		b.Fatal(err)
	}
	entity.FlushCameraCache()

	if err := cell.Delete(); err != nil {
		b.Fatal(err)
	}

	if err := place.Delete(); err != nil {
		b.Fatal(err)
	}

	for i := range labelCount {
		entity.UnscopedDb().Where("label_id = ?", labels[i]).Delete(&entity.Label{})
	}

	for i := range keywordCount {
		entity.UnscopedDb().Where("keyword_id = ?", keywords[i]).Delete(&entity.Keyword{})
	}
}
