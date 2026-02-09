package entity

import (
	"fmt"
	"math"
	"math/rand/v2"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity"
)

func populatePhotoPrismStructsWithAutoIncrement(t *testing.T, db *gorm.DB) {
	savedID := uint(0)

	album := entity.Album{}
	if err := populateStructWithMin(&album); err != nil {
		t.Error(err)
	} else {
		album.ID = savedID
		if result := db.Create(&album); result.Error != nil {
			t.Error(result.Error)
		}
	}

	user := entity.User{}
	if err := populateStructWithMin(&user); err != nil {
		t.Error(err)
	} else {
		user.ID = int(savedID)
		if result := db.Create(&user); result.Error != nil {
			t.Error(result.Error)
		}
	}

	camera := entity.Camera{}
	if err := populateStructWithMin(&camera); err != nil {
		t.Error(err)
	} else {
		camera.ID = savedID
		if result := db.Create(&camera); result.Error != nil {
			t.Error(result.Error)
		}
	}

	tmpError := entity.Error{}
	if err := populateStructWithMin(&tmpError); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&tmpError); result.Error != nil {
			t.Error(result.Error)
		}
	}

	service := entity.Service{}
	if err := populateStructWithMin(&service); err != nil {
		t.Error(err)
	} else {
		service.ID = savedID
		if result := db.Create(&service); result.Error != nil {
			t.Error(result.Error)
		}
	}

	keyword := entity.Keyword{}
	if err := populateStructWithMin(&keyword); err != nil {
		t.Error(err)
	} else {
		keyword.ID = savedID
		if result := db.Create(&keyword); result.Error != nil {
			t.Error(result.Error)
		}
	}

	label := entity.Label{}
	if err := populateStructWithMin(&label); err != nil {
		t.Error(err)
	} else {
		label.ID = savedID
		if result := db.Create(&label); result.Error != nil {
			t.Error(result.Error)
		}
	}

	lens := entity.Lens{}
	if err := populateStructWithMin(&lens); err != nil {
		t.Error(err)
	} else {
		lens.ID = savedID
		if result := db.Create(&lens); result.Error != nil {
			t.Error(result.Error)
		}
	}

	photo := entity.Photo{}
	if err := populateStructWithMin(&photo); err != nil {
		t.Error(err)
	} else {
		photo.ID = savedID
		photo.CameraID = camera.ID
		photo.CellID = "zz"
		photo.LensID = lens.ID
		photo.PlaceID = "zz"
		if result := db.Create(&photo); result.Error != nil {
			t.Error(result.Error)
		}
	}

	file := entity.File{}
	if err := populateStructWithMin(&file); err != nil {
		t.Error(err)
	} else {
		file.ID = savedID
		file.PhotoID = photo.ID
		if result := db.Create(&file); result.Error != nil {
			t.Error(result.Error)
		}
	}

	if result := db.Unscoped().Delete(&file); result.Error != nil {
		t.Error(result.Error)
	}
	if result := db.Unscoped().Delete(&photo); result.Error != nil {
		t.Error(result.Error)
	}
	if result := db.Unscoped().Delete(&lens); result.Error != nil {
		t.Error(result.Error)
	}
	if result := db.Unscoped().Delete(&label); result.Error != nil {
		t.Error(result.Error)
	}
	if result := db.Unscoped().Delete(&keyword); result.Error != nil {
		t.Error(result.Error)
	}
	if result := db.Unscoped().Delete(&service); result.Error != nil {
		t.Error(result.Error)
	}
	if result := db.Unscoped().Delete(&tmpError); result.Error != nil {
		t.Error(result.Error)
	}
	if result := db.Unscoped().Delete(&camera); result.Error != nil {
		t.Error(result.Error)
	}
	if result := db.Unscoped().Delete(&user); result.Error != nil {
		t.Error(result.Error)
	}
	if result := db.Unscoped().Delete(&album); result.Error != nil {
		t.Error(result.Error)
	}
}

func populatePhotoPrismStructsWithMin(t *testing.T, db *gorm.DB) {
	savedID := uint(123456789)

	albumUser := entity.AlbumUser{}
	if err := populateStructWithMin(&albumUser); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&albumUser); result.Error != nil {
			t.Error(result.Error)
		}
	}

	album := entity.Album{}
	if err := populateStructWithMin(&album); err != nil {
		t.Error(err)
	} else {
		album.ID = savedID
		if result := db.Create(&album); result.Error != nil {
			t.Error(result.Error)
		}
	}

	client := entity.Client{}
	if err := populateStructWithMin(&client); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&client); result.Error != nil {
			t.Error(result.Error)
		}
	}

	session := entity.Session{}
	if err := populateStructWithMin(&session); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&session); result.Error != nil {
			t.Error(result.Error)
		}
	}

	user := entity.User{}
	if err := populateStructWithMin(&user); err != nil {
		t.Error(err)
	} else {
		user.ID = int(savedID)
		if result := db.Create(&user); result.Error != nil {
			t.Error(result.Error)
		}
	}

	userDetails := entity.UserDetails{}
	if err := populateStructWithMin(&userDetails); err != nil {
		t.Error(err)
	} else {
		userDetails.UserUID = user.UserUID
		if result := db.Create(&userDetails); result.Error != nil {
			t.Error(result.Error)
		}
	}

	userSettings := entity.UserSettings{}
	if err := populateStructWithMin(&userSettings); err != nil {
		t.Error(err)
	} else {
		userSettings.UserUID = user.UserUID
		if result := db.Create(&userSettings); result.Error != nil {
			t.Error(result.Error)
		}
	}

	userShare := entity.UserShare{}
	if err := populateStructWithMin(&userShare); err != nil {
		t.Error(err)
	} else {
		userShare.UserUID = user.UserUID
		if result := db.Create(&userShare); result.Error != nil {
			t.Error(result.Error)
		}
	}

	camera := entity.Camera{}
	if err := populateStructWithMin(&camera); err != nil {
		t.Error(err)
	} else {
		camera.ID = savedID
		if result := db.Create(&camera); result.Error != nil {
			t.Error(result.Error)
		}
	}

	place := entity.Place{}
	if err := populateStructWithMin(&place); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&place); result.Error != nil {
			t.Error(result.Error)
		}
	}

	cell := entity.Cell{}
	if err := populateStructWithMin(&cell); err != nil {
		t.Error(err)
	} else {
		cell.PlaceID = place.ID
		if result := db.Create(&cell); result.Error != nil {
			t.Error(result.Error)
		}
	}

	country := entity.Country{}
	if err := populateStructWithMin(&country); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&country); result.Error != nil {
			t.Error(result.Error)
		}
	}

	duplicate := entity.Duplicate{}
	if err := populateStructWithMin(&duplicate); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&duplicate); result.Error != nil {
			t.Error(result.Error)
		}
	}

	tmpError := entity.Error{}
	if err := populateStructWithMin(&tmpError); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&tmpError); result.Error != nil {
			t.Error(result.Error)
		}
	}

	face := entity.Face{}
	if err := populateStructWithMin(&face); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&face); result.Error != nil {
			t.Error(result.Error)
		}
	}

	service := entity.Service{}
	if err := populateStructWithMin(&service); err != nil {
		t.Error(err)
	} else {
		service.ID = savedID
		if result := db.Create(&service); result.Error != nil {
			t.Error(result.Error)
		}
	}

	fileSync := entity.FileSync{}
	if err := populateStructWithMin(&fileSync); err != nil {
		t.Error(err)
	} else {
		fileSync.ServiceID = savedID
		if result := db.Create(&fileSync); result.Error != nil {
			t.Error(result.Error)
		}
	}

	folder := entity.Folder{}
	if err := populateStructWithMin(&folder); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&folder); result.Error != nil {
			t.Error(result.Error)
		}
	}

	keyword := entity.Keyword{}
	if err := populateStructWithMin(&keyword); err != nil {
		t.Error(err)
	} else {
		keyword.ID = savedID
		if result := db.Create(&keyword); result.Error != nil {
			t.Error(result.Error)
		}
	}

	label := entity.Label{}
	if err := populateStructWithMin(&label); err != nil {
		t.Error(err)
	} else {
		label.ID = savedID
		if result := db.Create(&label); result.Error != nil {
			t.Error(result.Error)
		}
	}

	category := entity.Category{}
	if err := populateStructWithMin(&category); err != nil {
		t.Error(err)
	} else {
		category.LabelID = savedID
		category.CategoryID = savedID
		if result := db.Create(&category); result.Error != nil {
			t.Error(result.Error)
		}
	}

	lens := entity.Lens{}
	if err := populateStructWithMin(&lens); err != nil {
		t.Error(err)
	} else {
		lens.ID = savedID
		if result := db.Create(&lens); result.Error != nil {
			t.Error(result.Error)
		}
	}

	link := entity.Link{}
	if err := populateStructWithMin(&link); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&link); result.Error != nil {
			t.Error(result.Error)
		}
	}

	passcode := entity.Passcode{}
	if err := populateStructWithMin(&passcode); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&passcode); result.Error != nil {
			t.Error(result.Error)
		}
	}

	password := entity.Password{}
	if err := populateStructWithMin(&password); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&password); result.Error != nil {
			t.Error(result.Error)
		}
	}

	photoUser := entity.PhotoUser{}
	if err := populateStructWithMin(&photoUser); err != nil {
		t.Error(err)
	} else {
		photoUser.UserUID = user.UserUID
		if result := db.Create(&photoUser); result.Error != nil {
			t.Error(result.Error)
		}
	}

	reaction := entity.Reaction{}
	if err := populateStructWithMin(&reaction); err != nil {
		t.Error(err)
	} else {
		reaction.UserUID = user.UserUID
		if result := db.Create(&reaction); result.Error != nil {
			t.Error(result.Error)
		}
	}

	subject := entity.Subject{}
	if err := populateStructWithMin(&subject); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&subject); result.Error != nil {
			t.Error(result.Error)
		}
	}

	photo := entity.Photo{}
	if err := populateStructWithMin(&photo); err != nil {
		t.Error(err)
	} else {
		photo.ID = savedID
		photo.CameraID = camera.ID
		photo.CellID = cell.ID
		photo.LensID = lens.ID
		photo.PlaceID = place.ID
		if result := db.Create(&photo); result.Error != nil {
			t.Error(result.Error)
		}
	}

	photoAlbum := entity.PhotoAlbum{}
	if err := populateStructWithMin(&photoAlbum); err != nil {
		t.Error(err)
	} else {
		photoAlbum.PhotoUID = photo.PhotoUID
		photoAlbum.AlbumUID = album.AlbumUID
		if result := db.Create(&photoAlbum); result.Error != nil {
			t.Error(result.Error)
		}
	}

	photoKeyword := entity.PhotoKeyword{}
	if err := populateStructWithMin(&photoKeyword); err != nil {
		t.Error(err)
	} else {
		photoKeyword.PhotoID = photo.ID
		photoKeyword.KeywordID = keyword.ID
		if result := db.Create(&photoKeyword); result.Error != nil {
			t.Error(result.Error)
		}
	}

	photoLabel := entity.PhotoLabel{}
	if err := populateStructWithMin(&photoLabel); err != nil {
		t.Error(err)
	} else {
		photoLabel.PhotoID = photo.ID
		photoLabel.LabelID = label.ID
		if result := db.Create(&photoLabel); result.Error != nil {
			t.Error(result.Error)
		}
	}

	file := entity.File{}
	if err := populateStructWithMin(&file); err != nil {
		t.Error(err)
	} else {
		file.ID = savedID
		file.PhotoID = savedID
		if result := db.Create(&file); result.Error != nil {
			t.Error(result.Error)
		}
	}

	fileShare := entity.FileShare{}
	if err := populateStructWithMin(&fileShare); err != nil {
		t.Error(err)
	} else {
		fileShare.FileID = savedID
		fileShare.ServiceID = savedID
		if result := db.Create(&fileShare); result.Error != nil {
			t.Error(result.Error)
		}
	}

	marker := entity.Marker{}
	if err := populateStructWithMin(&marker); err != nil {
		t.Error(err)
	} else {
		marker.FileUID = file.FileUID
		marker.SubjUID = subject.SubjUID
		marker.FaceID = face.ID
		if result := db.Create(&marker); result.Error != nil {
			t.Error(result.Error)
		}
	}

	details := entity.Details{}
	if err := populateStructWithMin(&details); err != nil {
		t.Error(err)
	} else {
		details.PhotoID = savedID
		if result := db.Create(&details); result.Error != nil {
			t.Error(result.Error)
		}
	}

}

func populatePhotoPrismStructsWithMax(t *testing.T, db *gorm.DB) {
	var uintMaxInt64 bool

	switch db.Dialector.Name() {
	case entity.SQLite3, entity.Postgres:
		uintMaxInt64 = true
	default:
		uintMaxInt64 = false
	}

	savedID := uint(math.MaxUint)
	if uintMaxInt64 {
		savedID = uint(math.MaxInt64)
	}

	albumUser := entity.AlbumUser{}
	if err := populateStructWithMax(&albumUser, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&albumUser); result.Error != nil {
			t.Error(result.Error)
		}
	}

	album := entity.Album{}
	if err := populateStructWithMax(&album, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		album.ID = savedID
		if result := db.Create(&album); result.Error != nil {
			t.Error(result.Error)
		}
	}

	client := entity.Client{}
	if err := populateStructWithMax(&client, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&client); result.Error != nil {
			t.Error(result.Error)
		}
	}

	session := entity.Session{}
	if err := populateStructWithMax(&session, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&session); result.Error != nil {
			t.Error(result.Error)
		}
	}

	user := entity.User{}
	if err := populateStructWithMax(&user, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		user.ID = math.MaxInt
		if result := db.Create(&user); result.Error != nil {
			t.Error(result.Error)
		}
	}

	userDetails := entity.UserDetails{}
	if err := populateStructWithMax(&userDetails, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		userDetails.UserUID = user.UserUID
		if result := db.Create(&userDetails); result.Error != nil {
			t.Error(result.Error)
		}
	}

	userSettings := entity.UserSettings{}
	if err := populateStructWithMax(&userSettings, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		userSettings.UserUID = user.UserUID
		if result := db.Create(&userSettings); result.Error != nil {
			t.Error(result.Error)
		}
	}

	userShare := entity.UserShare{}
	if err := populateStructWithMax(&userShare, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		userShare.UserUID = user.UserUID
		if result := db.Create(&userShare); result.Error != nil {
			t.Error(result.Error)
		}
	}

	camera := entity.Camera{}
	if err := populateStructWithMax(&camera, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		camera.ID = savedID
		if result := db.Create(&camera); result.Error != nil {
			t.Error(result.Error)
		}
	}

	place := entity.Place{}
	if err := populateStructWithMax(&place, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&place); result.Error != nil {
			t.Error(result.Error)
		}
	}

	cell := entity.Cell{}
	if err := populateStructWithMax(&cell, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		cell.PlaceID = place.ID
		if result := db.Create(&cell); result.Error != nil {
			t.Error(result.Error)
		}
	}

	country := entity.Country{}
	if err := populateStructWithMax(&country, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&country); result.Error != nil {
			t.Error(result.Error)
		}
	}

	duplicate := entity.Duplicate{}
	if err := populateStructWithMax(&duplicate, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&duplicate); result.Error != nil {
			t.Error(result.Error)
		}
	}

	tmpError := entity.Error{}
	if err := populateStructWithMax(&tmpError, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&tmpError); result.Error != nil {
			t.Error(result.Error)
		}
	}

	face := entity.Face{}
	if err := populateStructWithMax(&face, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&face); result.Error != nil {
			t.Error(result.Error)
		}
	}

	service := entity.Service{}
	if err := populateStructWithMax(&service, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		service.ID = savedID
		if result := db.Create(&service); result.Error != nil {
			t.Error(result.Error)
		}
	}

	fileSync := entity.FileSync{}
	if err := populateStructWithMax(&fileSync, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		fileSync.ServiceID = savedID
		if result := db.Create(&fileSync); result.Error != nil {
			t.Error(result.Error)
		}
	}

	folder := entity.Folder{}
	if err := populateStructWithMax(&folder, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&folder); result.Error != nil {
			t.Error(result.Error)
		}
	}

	keyword := entity.Keyword{}
	if err := populateStructWithMax(&keyword, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		keyword.ID = savedID
		if result := db.Create(&keyword); result.Error != nil {
			t.Error(result.Error)
		}
	}

	label := entity.Label{}
	if err := populateStructWithMax(&label, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		label.ID = savedID
		if result := db.Create(&label); result.Error != nil {
			t.Error(result.Error)
		}
	}

	category := entity.Category{}
	if err := populateStructWithMax(&category, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		category.LabelID = savedID
		category.CategoryID = savedID
		if result := db.Create(&category); result.Error != nil {
			t.Error(result.Error)
		}
	}

	lens := entity.Lens{}
	if err := populateStructWithMax(&lens, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		lens.ID = savedID
		if result := db.Create(&lens); result.Error != nil {
			t.Error(result.Error)
		}
	}

	link := entity.Link{}
	if err := populateStructWithMax(&link, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&link); result.Error != nil {
			t.Error(result.Error)
		}
	}

	passcode := entity.Passcode{}
	if err := populateStructWithMax(&passcode, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&passcode); result.Error != nil {
			t.Error(result.Error)
		}
	}

	password := entity.Password{}
	if err := populateStructWithMax(&password, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&password); result.Error != nil {
			t.Error(result.Error)
		}
	}

	photoUser := entity.PhotoUser{}
	if err := populateStructWithMax(&photoUser, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		photoUser.UserUID = user.UserUID
		if result := db.Create(&photoUser); result.Error != nil {
			t.Error(result.Error)
		}
	}

	reaction := entity.Reaction{}
	if err := populateStructWithMax(&reaction, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		reaction.UserUID = user.UserUID
		if result := db.Create(&reaction); result.Error != nil {
			t.Error(result.Error)
		}
	}

	subject := entity.Subject{}
	if err := populateStructWithMax(&subject, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		if result := db.Create(&subject); result.Error != nil {
			t.Error(result.Error)
		}
	}

	photo := entity.Photo{}
	if err := populateStructWithMax(&photo, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		photo.ID = savedID
		photo.CameraID = camera.ID
		photo.CellID = cell.ID
		photo.LensID = lens.ID
		photo.PlaceID = place.ID
		if result := db.Create(&photo); result.Error != nil {
			t.Error(result.Error)
		}
	}

	photoAlbum := entity.PhotoAlbum{}
	if err := populateStructWithMax(&photoAlbum, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		photoAlbum.PhotoUID = photo.PhotoUID
		photoAlbum.AlbumUID = album.AlbumUID
		if result := db.Create(&photoAlbum); result.Error != nil {
			t.Error(result.Error)
		}
	}

	photoKeyword := entity.PhotoKeyword{}
	if err := populateStructWithMax(&photoKeyword, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		photoKeyword.PhotoID = photo.ID
		photoKeyword.KeywordID = keyword.ID
		if result := db.Create(&photoKeyword); result.Error != nil {
			t.Error(result.Error)
		}
	}

	photoLabel := entity.PhotoLabel{}
	if err := populateStructWithMax(&photoLabel, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		photoLabel.PhotoID = photo.ID
		photoLabel.LabelID = label.ID
		if result := db.Create(&photoLabel); result.Error != nil {
			t.Error(result.Error)
		}
	}

	file := entity.File{}
	if err := populateStructWithMax(&file, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		file.ID = savedID
		file.PhotoID = savedID
		if result := db.Create(&file); result.Error != nil {
			t.Error(result.Error)
		}
	}

	fileShare := entity.FileShare{}
	if err := populateStructWithMax(&fileShare, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		fileShare.FileID = savedID
		fileShare.ServiceID = savedID
		if result := db.Create(&fileShare); result.Error != nil {
			t.Error(result.Error)
		}
	}

	marker := entity.Marker{}
	if err := populateStructWithMax(&marker, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		marker.FileUID = file.FileUID
		marker.SubjUID = subject.SubjUID
		marker.FaceID = face.ID
		if result := db.Create(&marker); result.Error != nil {
			t.Error(result.Error)
		}
	}

	details := entity.Details{}
	if err := populateStructWithMax(&details, uintMaxInt64); err != nil {
		t.Error(err)
	} else {
		details.PhotoID = savedID
		if result := db.Create(&details); result.Error != nil {
			t.Error(result.Error)
		}
	}

}

func populateStructWithMin(m interface{}) (err error) {

	r := reflect.ValueOf(m)

	if r.Kind() != reflect.Pointer {
		return fmt.Errorf("model interface expected")
	}

	values := r.Elem()

	if kind := values.Kind(); kind != reflect.Struct {
		return fmt.Errorf("model expected")
	}

	reType := regexp.MustCompile(`(?:type:)(?P<type>[a-zA-Z]*)`)
	reSize := regexp.MustCompile(`(?:size:)(?P<size>[0-9]*)`)

	vT := values.Type()
	num := vT.NumField()
	for i := 0; i < num; i++ {
		field := vT.Field(i)

		// Skip non-exported fields.
		if !field.IsExported() {
			continue
		}

		// fieldName := field.Name

		v := values.Field(i)
		tag, _ := readTag(field, "gorm")
		// log.Debugf("field %v is %v with name %v or string %v and is exported %v with tag %v", fieldName, v.Kind(), v.Type().Name(), v.Type().String(), field.IsExported(), tag)

		// Skip read-only fields.
		if !v.CanSet() {
			continue
		}

		// Skip ignored fields by tag
		if tag == "-" {
			continue
		}

		if strings.Contains(tag, "type") {
			typeString := reFindStringGroup(reType, tag, "type")
			switch typeString {
			// case "bytes":
			// 	v.Set(reflect.ValueOf(""))
			case "float":
				strLen, _ := strconv.Atoi(reFindStringGroup(reSize, tag, "size"))
				if strLen == 32 {
					v.SetFloat(math.SmallestNonzeroFloat32)
				} else {
					v.SetFloat(math.SmallestNonzeroFloat64)
				}
			case "int":
				strLen, _ := strconv.Atoi(reFindStringGroup(reSize, tag, "size"))
				switch strLen {
				case 8:
					v.SetInt(math.MinInt8)
				case 16:
					v.SetInt(math.MinInt16)
				case 32:
					v.SetInt(math.MinInt32)
				case 64:
					v.SetInt(math.MinInt64)
				}
			}
		} else {
			switch v.Kind() {
			case reflect.Struct:
				switch v.Type().String() {
				case "time.Time":
					v.Set(reflect.ValueOf(time.Date(1000, 01, 01, 0, 0, 0, 0, time.UTC))) // Mariadb limitation
				default:
					// case "sql.NullTime", "time.Duration":
					log.Debugf("reflect.Struct with Unhandled type %s for field %s", v.Type().String(), field.Name)

				}
			case reflect.Pointer:
				switch v.Type().String() {
				case "*time.Time":
					timeTime := time.Date(1000, 01, 01, 0, 0, 0, 0, time.UTC)
					v.Set(reflect.ValueOf(&timeTime)) // Mariadb limitation
				default:
					//case "*time.Time", "*time.Duration", "*bool", "*uint", "*uint64", "*uint32", "*int", "*int64", "*int32", "*string", "*float32", "*float64", "*otp.Key", "*sql.NullTime", "*json.RawMessage":
					log.Debugf("reflect.Pointer with Unhandled type %s for field %s", v.Type().String(), field.Name)
				}
			case reflect.Uint:
				v.Set(reflect.ValueOf(uint(0)))
			case reflect.Int8:
				v.SetInt(math.MinInt8)
			case reflect.Int16:
				v.SetInt(math.MinInt16)
			case reflect.Int32:
				v.SetInt(math.MinInt32)
			case reflect.Int64:
				v.SetInt(math.MinInt64)
			case reflect.Int:
				if strings.Contains(tag, "size") {
					tagSize, _ := strconv.Atoi(reFindStringGroup(reSize, tag, "size"))
					switch tagSize {
					case 8:
						v.SetInt(math.MinInt8)
					case 16:
						v.SetInt(math.MinInt16)
					case 32:
						v.SetInt(math.MinInt32)
					case 64:
						v.SetInt(math.MinInt64)
					}
				} else {
					v.SetInt(math.MinInt)
				}
			case reflect.Float32:
				v.SetFloat(math.SmallestNonzeroFloat32)
			case reflect.Float64:
				v.SetFloat(math.SmallestNonzeroFloat64)
			case reflect.String:
				if strings.Contains(tag, "type") {
					typeString := reFindStringGroup(reType, tag, "type")
					switch typeString {
					case "bytes":
						v.Set(reflect.ValueOf(""))
					case "float":
						strLen, _ := strconv.Atoi(reFindStringGroup(reSize, tag, "size"))
						if strLen == 32 {
							v.SetFloat(math.SmallestNonzeroFloat32)
						} else {
							v.SetFloat(math.SmallestNonzeroFloat64)
						}
					case "int":
						strLen, _ := strconv.Atoi(reFindStringGroup(reSize, tag, "size"))
						switch strLen {
						case 8:
							v.SetInt(math.MinInt8)
						case 16:
							v.SetInt(math.MinInt16)
						case 32:
							v.SetInt(math.MinInt32)
						case 64:
							v.SetInt(math.MinInt64)
						}
					}
				} else if strings.Contains(tag, "size") {
					v.Set(reflect.ValueOf(""))
				}
			case reflect.Bool:
				v.SetBool(false)
			default:
				log.Debugf("vKind %s with Unhandled type %s for field %s", v.Kind(), v.Type().String(), field.Name)
			}
		}
	}
	return nil
}

func populateStructWithMax(m interface{}, uintMaxInt64 bool) (err error) {

	r := reflect.ValueOf(m)

	if r.Kind() != reflect.Pointer {
		return fmt.Errorf("model interface expected")
	}

	values := r.Elem()

	if kind := values.Kind(); kind != reflect.Struct {
		return fmt.Errorf("model expected")
	}

	reType := regexp.MustCompile(`(?:type:)(?P<type>[a-zA-Z]*)`)
	reSize := regexp.MustCompile(`(?:size:)(?P<size>[0-9]*)`)

	vT := values.Type()
	num := vT.NumField()
	for i := 0; i < num; i++ {
		field := vT.Field(i)

		// Skip non-exported fields.
		if !field.IsExported() {
			continue
		}

		// fieldName := field.Name

		v := values.Field(i)
		tag, _ := readTag(field, "gorm")
		// log.Debugf("field %v is %v with name %v or string %v and is exported %v with tag %v", fieldName, v.Kind(), v.Type().Name(), v.Type().String(), field.IsExported(), tag)

		// Skip read-only fields.
		if !v.CanSet() {
			continue
		}

		switch v.Kind() {
		case reflect.Struct:
			switch v.Type().String() {
			case "time.Time":
				v.Set(reflect.ValueOf(time.Date(9999, 12, 31, 23, 59, 59, 999999, time.UTC))) // Mariadb limitation
				//case "sql.NullTime", "time.Duration":

			}
		case reflect.Pointer:
			switch v.Type().String() {
			case "*time.Time":
				timeTime := time.Date(9999, 12, 31, 23, 59, 59, 999999, time.UTC)
				v.Set(reflect.ValueOf(&timeTime)) // Mariadb limitation

				// case "*time.Time", "*time.Duration", "*bool", "*uint", "*uint64", "*uint32", "*int", "*int64", "*int32", "*string", "*float32", "*float64", "*otp.Key", "*sql.NullTime", "*json.RawMessage":
			}
		case reflect.Uint:
			if strings.Contains(tag, "size") {
				tagSize, _ := strconv.Atoi(reFindStringGroup(reSize, tag, "size"))
				switch tagSize {
				case 8:
					v.Set(reflect.ValueOf(uint(math.MaxUint8)))
				case 16:
					v.Set(reflect.ValueOf(uint(math.MaxUint16)))
				case 32:
					v.Set(reflect.ValueOf(uint(math.MaxUint32)))
				case 64:
					if uintMaxInt64 {
						v.Set(reflect.ValueOf(uint(math.MaxUint32)))
					} else {
						v.Set(reflect.ValueOf(uint(math.MaxUint64)))
					}
				}
			} else {
				if uintMaxInt64 {
					v.Set(reflect.ValueOf(uint(math.MaxUint32)))
				} else {
					v.Set(reflect.ValueOf(uint(math.MaxUint64)))
				}
			}
		case reflect.Int8:
			v.SetInt(math.MaxInt8)
		case reflect.Int16:
			v.SetInt(math.MaxInt16)
		case reflect.Int32:
			v.SetInt(math.MaxInt32)
		case reflect.Int64:
			v.SetInt(math.MaxInt64)
		case reflect.Int:
			if strings.Contains(tag, "size") {
				tagSize, _ := strconv.Atoi(reFindStringGroup(reSize, tag, "size"))
				switch tagSize {
				case 8:
					v.SetInt(math.MaxInt8)
				case 16:
					v.SetInt(math.MaxInt16)
				case 32:
					v.SetInt(math.MaxInt32)
				case 64:
					v.SetInt(math.MaxInt64)
				}
			} else {
				v.SetInt(math.MaxInt)
			}
		case reflect.Float32:
			v.SetFloat(math.MaxFloat32)
		case reflect.Float64:
			v.SetFloat(math.MaxFloat64)
		case reflect.String:
			// tag
			if tag == "-" {
				continue
			}
			if strings.Contains(tag, "type") {
				typeString := reFindStringGroup(reType, tag, "type")
				switch typeString {
				case "bytes":
					strLen, _ := strconv.Atoi(reFindStringGroup(reSize, tag, "size"))
					if strLen <= 0 || strLen > 8192 {
						v.Set(reflect.ValueOf(randomString(8192)))
					} else {
						v.Set(reflect.ValueOf(randomString(strLen)))
					}
				case "float":
					strLen, _ := strconv.Atoi(reFindStringGroup(reSize, tag, "size"))
					if strLen == 32 {
						v.SetFloat(math.SmallestNonzeroFloat32)
					} else {
						v.SetFloat(math.SmallestNonzeroFloat64)
					}
				case "int":
					strLen, _ := strconv.Atoi(reFindStringGroup(reSize, tag, "size"))
					switch strLen {
					case 8:
						v.SetInt(math.MaxInt8)
					case 16:
						v.SetInt(math.MaxInt16)
					case 32:
						v.SetInt(math.MaxInt32)
					case 64:
						v.SetInt(math.MaxInt64)
					}
				}
			} else if strings.Contains(tag, "size") {
				strLen, _ := strconv.Atoi(reFindStringGroup(reSize, tag, "size"))
				if strLen <= 0 || strLen > 8192 {
					v.Set(reflect.ValueOf(randomString(8192)))
				} else {
					v.Set(reflect.ValueOf(randomString(strLen)))
				}
			}
		case reflect.Bool:
			v.SetBool(true)

		}
	}
	return nil
}

func reFindStringGroup(expression *regexp.Regexp, textToSearch string, groupName string) string {
	match := expression.FindStringSubmatch(textToSearch)
	groupIndex := expression.SubexpIndex(groupName)

	return match[groupIndex]
}

func readTag(f reflect.StructField, id string) (result string, valid bool) {
	val, ok := f.Tag.Lookup(id)
	if !ok {
		return id, false
	}
	return val, true
}

const characterRunes = " abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(len int) string {
	sb := strings.Builder{}
	sb.Grow(len)
	for i := 0; i < len; {
		sb.WriteByte(characterRunes[rand.IntN(53)])
		i++
	}

	return sb.String()
}
