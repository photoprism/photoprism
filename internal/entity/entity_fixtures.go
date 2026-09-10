package entity

import (
	"fmt"
	"maps"
	"os"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CreateDefaultFixtures inserts default fixtures for test and production.
func CreateDefaultFixtures() {
	CreateDefaultUsers()
	CreateUnknownPlace()
	CreateUnknownLocation()
	CreateUnknownCountry()
	CreateUnknownCamera()
	CreateUnknownLens()
}

// ResetTestFixtures recreates database tables and test fixtures.
func ResetTestFixtures() {
	start := time.Now()

	// Make sure that the migrations and versions tables are already there, as once prevents these from being handled correctly in tests.
	if (!Db().HasTable(&migrate.Migration{})) {
		Db().AutoMigrate(&migrate.Migration{})
	}
	if (!Db().HasTable(&migrate.Version{})) {
		Db().AutoMigrate(&migrate.Version{})
	}

	Entities.Migrate(Db(), migrate.Opt(true, false, nil))

	if err := Entities.WaitForMigration(Db()); err != nil {
		log.Errorf("migrate: %s [%s]", err, time.Since(start))
	}

	Entities.Truncate(Db())

	CreateDefaultFixtures()

	CreateTestFixtures()

	FlushCaches()

	File{}.RegenerateIndex()

	log.Debugf("migrate: recreated test fixtures [%s]", time.Since(start))
}

// ValidateFixtures checks that all the Fixtures are in the database, and that there are not extra records in place.
func ValidateFixtures(t *testing.T) {
	t.Helper()
	if os.Getenv("PHOTOPRISM_TEST_CLEANUP") != "" {
		var photoLabelFixtures []PhotoLabel
		require.NoError(t, UnscopedDb().Find(&photoLabelFixtures).Error)
		var categoryFixtures []Category
		require.NoError(t, UnscopedDb().Find(&categoryFixtures).Error)

		t.Cleanup(func() {
			t.Helper()

			var albumDBFixtures []Album
			albumFixtures := slices.Collect(maps.Values(AlbumFixtures))
			require.NoError(t, UnscopedDb().Find(&albumDBFixtures).Error)
			assert.Equal(t, len(albumFixtures), len(albumDBFixtures), "Number of DB Albums is wrong")
			validateFixture(t, "Album", albumFixtures, albumDBFixtures)

			var clientDBFixtures []Client
			clientFixtures := slices.Collect(maps.Values(ClientFixtures))
			require.NoError(t, UnscopedDb().Find(&clientDBFixtures).Error)
			assert.Equal(t, len(clientFixtures), len(clientDBFixtures), "Number of DB Clients is wrong")
			validateFixture(t, "Client", clientFixtures, clientDBFixtures)

			var sessionDBFixtures []Session
			sessionFixtures := slices.Collect(maps.Values(SessionFixtures))
			require.NoError(t, UnscopedDb().Find(&sessionDBFixtures).Error)
			assert.Equal(t, len(sessionFixtures), len(sessionDBFixtures), "Number of DB Sessions is wrong")
			validateFixture(t, "Session", sessionFixtures, sessionDBFixtures)

			var userDBFixtures []User
			userFixtures := slices.Collect(maps.Values(UserFixtures))
			require.NoError(t, UnscopedDb().Where("id in (1, -2, -1)").Find(&userDBFixtures).Error)
			userFixtures = append(userFixtures, userDBFixtures...)
			require.NoError(t, UnscopedDb().Find(&userDBFixtures).Error)
			assert.Equal(t, len(userFixtures), len(userDBFixtures), "Number of DB Users is wrong")
			validateFixture(t, "User", userFixtures, userDBFixtures, "UserDetails", "UserSettings", "PreviewToken", "RefID")

			var userShareDBFixtures []UserShare
			userShareFixtures := slices.Collect(maps.Values(UserShareFixtures))
			require.NoError(t, UnscopedDb().Find(&userShareDBFixtures).Error)
			assert.Equal(t, len(userShareFixtures), len(userShareDBFixtures), "Number of DB UserShares is wrong")
			validateFixture(t, "UserShare", userShareFixtures, userShareDBFixtures)

			var cameraDBFixtures []Camera
			cameraFixtures := slices.Collect(maps.Values(CameraFixtures))
			require.NoError(t, UnscopedDb().Where("id in (1)").Find(&cameraDBFixtures).Error)
			cameraFixtures = append(cameraFixtures, cameraDBFixtures...)
			require.NoError(t, UnscopedDb().Find(&cameraDBFixtures).Error)
			assert.Equal(t, len(cameraFixtures), len(cameraDBFixtures), "Number of DB Cameras is wrong")
			validateFixture(t, "Camera", cameraFixtures, cameraDBFixtures)

			var countryDBFixtures []Country
			countryFixtures := slices.Collect(maps.Values(CountryFixtures))
			require.NoError(t, UnscopedDb().Where("id = 'zz'").Find(&countryDBFixtures).Error)
			countryFixtures = append(countryFixtures, countryDBFixtures...)
			require.NoError(t, UnscopedDb().Find(&countryDBFixtures).Error)
			assert.Equal(t, len(countryFixtures), len(countryDBFixtures), "Number of DB Countrys is wrong")
			validateFixture(t, "Country", countryFixtures, countryDBFixtures)

			var detailsDBFixtures []Details
			detailsFixtures := slices.Collect(maps.Values(DetailsFixtures))
			require.NoError(t, UnscopedDb().Find(&detailsDBFixtures).Error)
			assert.Equal(t, len(detailsFixtures), len(detailsDBFixtures), "Number of DB Detailss is wrong")
			validateFixture(t, "Details", detailsFixtures, detailsDBFixtures)

			var duplicateDBFixtures []Duplicate
			duplicateFixtures := []Duplicate{}
			require.NoError(t, UnscopedDb().Find(&duplicateDBFixtures).Error)
			assert.Equal(t, len(duplicateFixtures), len(duplicateDBFixtures), "Number of DB Duplicates is wrong")
			validateFixture(t, "Duplicates", duplicateFixtures, duplicateDBFixtures)

			var faceDBFixtures []Face
			faceFixtures := slices.Collect(maps.Values(FaceFixtures))
			require.NoError(t, UnscopedDb().Find(&faceDBFixtures).Error)
			assert.Equal(t, len(faceFixtures), len(faceDBFixtures), "Number of DB Faces is wrong")
			validateFixture(t, "Face", faceFixtures, faceDBFixtures)

			var fileDBFixtures []File
			fileFixtures := slices.Collect(maps.Values(FileFixtures))
			require.NoError(t, UnscopedDb().Find(&fileDBFixtures).Error)
			assert.Equal(t, len(fileFixtures), len(fileDBFixtures), "Number of DB Files is wrong")
			validateFixture(t, "File", fileFixtures, fileDBFixtures, "Photo", "PhotoTakenAt", "TimeIndex", "MediaID", "FileColors")

			var fileShareDBFixtures []FileShare
			fileShareFixtures := slices.Collect(maps.Values(FileShareFixtures))
			require.NoError(t, UnscopedDb().Find(&fileShareDBFixtures).Error)
			assert.Equal(t, len(fileShareFixtures), len(fileShareDBFixtures), "Number of DB FileShares is wrong")
			validateFixture(t, "FileShare", fileShareFixtures, fileShareDBFixtures, "Account")

			var fileSyncDBFixtures []FileSync
			fileSyncFixtures := slices.Collect(maps.Values(FileSyncFixtures))
			require.NoError(t, UnscopedDb().Find(&fileSyncDBFixtures).Error)
			assert.Equal(t, len(fileSyncFixtures), len(fileSyncDBFixtures), "Number of DB FileSyncs is wrong")
			validateFixture(t, "FileSync", fileSyncFixtures, fileSyncDBFixtures, "Account", "File")

			var folderDBFixtures []Folder
			folderFixtures := slices.Collect(maps.Values(FolderFixtures))
			require.NoError(t, UnscopedDb().Find(&folderDBFixtures).Error)
			assert.Equal(t, len(folderFixtures), len(folderDBFixtures), "Number of DB Folders is wrong")
			validateFixture(t, "Folder", folderFixtures, folderDBFixtures)

			var keywordDBFixtures []Keyword
			keywordFixtures := slices.Collect(maps.Values(KeywordFixtures))
			require.NoError(t, UnscopedDb().Find(&keywordDBFixtures).Error)
			assert.Equal(t, len(keywordFixtures), len(keywordDBFixtures), "Number of DB Keywords is wrong")
			validateFixture(t, "Keyword", keywordFixtures, keywordDBFixtures)

			var labelDBFixtures []Label
			labelFixtures := slices.Collect(maps.Values(LabelFixtures))
			require.NoError(t, UnscopedDb().Find(&labelDBFixtures).Error)
			assert.Equal(t, len(labelFixtures), len(labelDBFixtures), "Number of DB Labels is wrong")
			validateFixture(t, "Label", labelFixtures, labelDBFixtures)

			var lensDBFixtures []Lens
			lensFixtures := slices.Collect(maps.Values(LensFixtures))
			require.NoError(t, UnscopedDb().Where("id = 1").Find(&lensDBFixtures).Error)
			lensFixtures = append(lensFixtures, lensDBFixtures...)
			require.NoError(t, UnscopedDb().Find(&lensDBFixtures).Error)
			assert.Equal(t, len(lensFixtures), len(lensDBFixtures), "Number of DB Lenss is wrong")
			validateFixture(t, "Lens", lensFixtures, lensDBFixtures)

			var linkDBFixtures []Link
			linkFixtures := slices.Collect(maps.Values(LinkFixtures))
			require.NoError(t, UnscopedDb().Find(&linkDBFixtures).Error)
			assert.Equal(t, len(linkFixtures), len(linkDBFixtures), "Number of DB Links is wrong")
			validateFixture(t, "Link", linkFixtures, linkDBFixtures, "RefID")

			var markerDBFixtures []Marker
			markerFixtures := slices.Collect(maps.Values(MarkerFixtures))
			require.NoError(t, UnscopedDb().Find(&markerDBFixtures).Error)
			assert.Equal(t, len(markerFixtures), len(markerDBFixtures), "Number of DB Markers is wrong")
			validateFixture(t, "Marker", markerFixtures, markerDBFixtures, "FaceDist")

			var passcodeDBFixtures []Passcode
			passcodeFixtures := slices.Collect(maps.Values(PasscodeFixtures))
			require.NoError(t, UnscopedDb().Find(&passcodeDBFixtures).Error)
			assert.Equal(t, len(passcodeFixtures), len(passcodeDBFixtures), "Number of DB Passcodes is wrong")
			validateFixture(t, "Passcode", passcodeFixtures, passcodeDBFixtures)

			var passwordDBFixtures []Password
			passwordFixtures := slices.Collect(maps.Values(PasswordFixtures))
			require.NoError(t, UnscopedDb().Where("uid = (select user_uid from auth_users where id = 1)").Find(&passwordDBFixtures).Error)
			passwordFixtures = append(passwordFixtures, passwordDBFixtures...)
			require.NoError(t, UnscopedDb().Find(&passwordDBFixtures).Error)
			assert.Equal(t, len(passwordFixtures), len(passwordDBFixtures), "Number of DB Passwords is wrong")
			validateFixture(t, "Password", passwordFixtures, passwordDBFixtures)

			var photoAlbumDBFixtures []PhotoAlbum
			photoAlbumFixtures := slices.Collect(maps.Values(PhotoAlbumFixtures))
			require.NoError(t, UnscopedDb().Find(&photoAlbumDBFixtures).Error)
			assert.Equal(t, len(photoAlbumFixtures), len(photoAlbumDBFixtures), "Number of DB PhotoAlbums is wrong")
			validateFixture(t, "PhotoAlbum", photoAlbumFixtures, photoAlbumDBFixtures, "Photo", "Album")

			var photoDBFixtures []Photo
			photoFixtures := slices.Collect(maps.Values(PhotoFixtures))
			require.NoError(t, UnscopedDb().Find(&photoDBFixtures).Error)
			assert.Equal(t, len(photoFixtures), len(photoDBFixtures), "Number of DB Photos is wrong")
			validateFixture(t, "Photo", photoFixtures, photoDBFixtures, "Details", "Camera", "Cell", "Lens", "Place", "Keywords", "Albums", "Files", "Labels", "PhotoDescription", "DescriptionSrc", "TakenAt", "TakenAtLocal")

			var photoKeywordDBFixtures []PhotoKeyword
			photoKeywordFixtures := slices.Collect(maps.Values(PhotoKeywordFixtures))
			require.NoError(t, UnscopedDb().Find(&photoKeywordDBFixtures).Error)
			assert.Equal(t, len(photoKeywordFixtures), len(photoKeywordDBFixtures), "Number of DB PhotoKeywords is wrong")
			validateFixture(t, "PhotoKeyword", photoKeywordFixtures, photoKeywordDBFixtures)

			var cellDBFixtures []Cell
			cellFixtures := slices.Collect(maps.Values(CellFixtures))
			cellFixtures = append(cellFixtures, UnknownLocation)
			require.NoError(t, UnscopedDb().Find(&cellDBFixtures).Error)
			assert.Equal(t, len(cellFixtures), len(cellDBFixtures), "Number of DB Cells is wrong")
			validateFixture(t, "Cell", cellFixtures, cellDBFixtures, "Place")

			var placeDBFixtures []Place
			placeFixtures := slices.Collect(maps.Values(PlaceFixtures))
			require.NoError(t, UnscopedDb().Where("id = 'zz'").Find(&placeDBFixtures).Error)
			placeFixtures = append(placeFixtures, placeDBFixtures...)
			require.NoError(t, UnscopedDb().Find(&placeDBFixtures).Error)
			assert.Equal(t, len(placeFixtures), len(placeDBFixtures), "Number of DB Places is wrong")
			validateFixture(t, "Place", placeFixtures, placeDBFixtures)

			var reactionDBFixtures []Reaction
			reactionFixtures := slices.Collect(maps.Values(ReactionFixtures))
			require.NoError(t, UnscopedDb().Find(&reactionDBFixtures).Error)
			assert.Equal(t, len(reactionFixtures), len(reactionDBFixtures), "Number of DB Reactions is wrong")
			validateFixture(t, "Reaction", reactionFixtures, reactionDBFixtures)

			var serviceDBFixtures []Service
			serviceFixtures := slices.Collect(maps.Values(ServiceFixtures))
			require.NoError(t, UnscopedDb().Find(&serviceDBFixtures).Error)
			assert.Equal(t, len(serviceFixtures), len(serviceDBFixtures), "Number of DB Services is wrong")
			validateFixture(t, "Service", serviceFixtures, serviceDBFixtures, "SyncDate")

			var subjectDBFixtures []Subject
			subjectFixtures := slices.Collect(maps.Values(SubjectFixtures))
			require.NoError(t, UnscopedDb().Find(&subjectDBFixtures).Error)
			assert.Equal(t, len(subjectFixtures), len(subjectDBFixtures), "Number of DB Subjects is wrong")
			validateFixture(t, "Subject", subjectFixtures, subjectDBFixtures)

			var photoLabelDBFixtures []PhotoLabel
			require.NoError(t, UnscopedDb().Find(&photoLabelDBFixtures).Error)
			assert.Equal(t, len(photoLabelFixtures), len(photoLabelDBFixtures), "Number of DB PhotoLabels is wrong")
			validateFixture(t, "PhotoLabel", photoLabelFixtures, photoLabelDBFixtures)

			var categoryDBFixtures []Category
			require.NoError(t, UnscopedDb().Find(&categoryDBFixtures).Error)
			assert.Equal(t, len(categoryFixtures), len(categoryDBFixtures), "Number of DB Categories is wrong")
			validateFixture(t, "Category", categoryFixtures, categoryDBFixtures)
		})
	}
}

func validateFixture[T any](t *testing.T, name string, fixture []T, data []T, omit ...string) {
	t.Helper()
	var err error
	var aV, bV Values
	failed := false
	// Validate fixture against data
	for _, a := range fixture {
		matched := false
		if aV, err = structToMap(&a, omit...); err != nil {
			require.NoError(t, err, "processing %+v", a)
		}
		for _, b := range data {
			if bV, err = structToMap(&b, omit...); err != nil {
				require.NoError(t, err, "processing %+v", b)
			}

			if maps.Equal(aV, bV) {
				matched = true
				break
			}
		}
		assert.True(t, matched, "unable to match Fixture %s from (%+v)", name, a)
		if !matched {
			failed = true
		}
	}
	// Validate data against fixture
	for _, a := range data {
		matched := false
		if aV, err = structToMap(&a, omit...); err != nil {
			require.NoError(t, err, "processing %+v", a)
		}
		for _, b := range fixture {
			if bV, err = structToMap(&b, omit...); err != nil {
				require.NoError(t, err, "processing %+v", b)
			}

			if maps.Equal(aV, bV) {
				matched = true
				break
			}
		}
		assert.True(t, matched, "unable to match DB value %s from (%+v)", name, a)
		if !matched {
			failed = true
		}
	}
	if failed {
		t.Fatalf("failed was true for %s", name)
	}
}

// structToMap extracts struct fields into a Values map, optionally excluding selected names.
func structToMap(m any, omit ...string) (result Values, err error) {
	mustOmit := func(name string) bool {
		return slices.Contains(omit, name)
	}

	r := reflect.ValueOf(m)

	if r.Kind() != reflect.Pointer {
		return result, fmt.Errorf("model interface expected")
	}

	values := r.Elem()

	if kind := values.Kind(); kind != reflect.Struct {
		return result, fmt.Errorf("model expected")
	}

	t := values.Type()
	num := t.NumField()

	result = make(Values, num)

	// Add exported fields to result.
	for i := range num {
		field := t.Field(i)

		fieldName := field.Name

		// Skip non-exported fields, because reflect can't get at them by design.
		if !field.IsExported() {
			continue
		}

		// Skip timestamps.
		if fieldName == "" || fieldName == "UpdatedAt" || fieldName == "CreatedAt" {
			continue
		}

		v := values.Field(i)

		switch v.Kind() {
		case reflect.Slice, reflect.Chan, reflect.Func, reflect.Map, reflect.UnsafePointer:
			continue
		case reflect.Struct:
			if v.IsZero() {
				continue
			}
		}

		// Skip read-only fields.
		if !v.CanSet() {
			continue
		}

		// Skip omitted.
		if mustOmit(fieldName) {
			continue
		}

		if field.Type == reflect.TypeOf(time.Time{}) {
			if t, err := timeReflectToString(v); err == nil {
				result[fieldName] = t
				continue
			}
		}
		if v.Kind() == reflect.Pointer && v.Type().String() == "*time.Time" {
			if t, err := timeReflectToString(v.Elem()); err == nil {
				result[fieldName] = t
				continue
			}
		}

		// Add value to result.
		result[fieldName] = v.Interface()
	}

	if len(result) == 0 {
		return result, fmt.Errorf("no values")
	}

	return result, nil
}

func timeReflectToString(v reflect.Value) (string, error) {
	// Ensure the underlying type matches time.Time or a named type based on it
	if !v.IsValid() {
		return "", fmt.Errorf("invalid reflect.Value")
	}

	// Extract the underlying interface and assert it to time.Time
	if t, ok := v.Interface().(time.Time); ok {
		// Use t.String() or t.Format("2006-01-02 15:04:05") for custom formats
		return t.Format(time.RFC3339), nil
	}

	return "", fmt.Errorf("value is not of type time.Time")
}
