package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
)

func TestBatchPhotosEdit(t *testing.T) {
	t.Run("FeatureDisabled", func(t *testing.T) {
		app, router, conf := NewApiTest()

		settings := conf.Settings()
		orig := settings.Features.BatchEdit
		settings.Features.BatchEdit = false
		t.Cleanup(func() {
			settings.Features.BatchEdit = orig
		})

		BatchPhotosEdit(router)

		photoUIDs := `{"photos": ["pqkm36fjqvset9uy"]}`
		resp := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/edit", photoUIDs)

		assert.Equal(t, http.StatusForbidden, resp.Code)
		assert.Contains(t, resp.Body.String(), i18n.Msg(i18n.ErrFeatureDisabled))
	})
	t.Run("SuccessNoChange", func(t *testing.T) {
		// Create new API test instance.
		app, router, _ := NewApiTest()

		// Attach POST /api/v1/batch/photos/edit request handler.
		BatchPhotosEdit(router)

		// Specify the unique IDs of the photos used for testing.
		photoUIDs := `["pqkm36fjqvset9uy", "pqkm36fjqvset9uz"]`

		// Get the photo models and current values for the batch edit form.
		editResponse := PerformRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			fmt.Sprintf(`{"photos": %s}`, photoUIDs),
		)

		// Check the edit response status code.
		assert.Equal(t, http.StatusOK, editResponse.Code)

		// Check the edit response body.
		editBody := editResponse.Body.String()
		assert.NotEmpty(t, editBody)
		assert.True(t, strings.HasPrefix(editBody, `{"models":[{"ID"`), "unexpected response")

		// Check the edit response values.
		editValues := gjson.Get(editBody, "values").Raw
		timezoneBefore := gjson.Get(editValues, "TimeZone")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", timezoneBefore.String())
		altitudeBefore := gjson.Get(editValues, "Altitude")
		assert.Equal(t, "{\"value\":0,\"mixed\":true,\"action\":\"none\"}", altitudeBefore.String())
		countryBefore := gjson.Get(editValues, "Country")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", countryBefore.String())
		latBefore := gjson.Get(editValues, "Lat")
		assert.Equal(t, "{\"value\":0,\"mixed\":true,\"action\":\"none\"}", latBefore.String())
		lngBefore := gjson.Get(editValues, "Lng")
		assert.Equal(t, "{\"value\":0,\"mixed\":true,\"action\":\"none\"}", lngBefore.String())
		typeBefore := gjson.Get(editValues, "Type")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", typeBefore.String())
		yearBefore := gjson.Get(editValues, "Year")
		assert.Equal(t, "{\"value\":-2,\"mixed\":true,\"action\":\"none\"}", yearBefore.String())
		dayBefore := gjson.Get(editValues, "Day")
		assert.Equal(t, "{\"value\":-2,\"mixed\":true,\"action\":\"none\"}", dayBefore.String())
		monthBefore := gjson.Get(editValues, "Month")
		assert.Equal(t, "{\"value\":-2,\"mixed\":true,\"action\":\"none\"}", monthBefore.String())
		titleBefore := gjson.Get(editValues, "Title")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", titleBefore.String())
		captionBefore := gjson.Get(editValues, "Caption")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", captionBefore.String())
		subjectBefore := gjson.Get(editValues, "DetailsSubject")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", subjectBefore.String())
		artistBefore := gjson.Get(editValues, "DetailsArtist")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", artistBefore.String())
		copyrightBefore := gjson.Get(editValues, "DetailsCopyright")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", copyrightBefore.String())
		licenseBefore := gjson.Get(editValues, "DetailsLicense")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", licenseBefore.String())
		favoriteBefore := gjson.Get(editValues, "Favorite")
		assert.Equal(t, "{\"value\":false,\"mixed\":true,\"action\":\"none\"}", favoriteBefore.String())
		scanBefore := gjson.Get(editValues, "Scan")
		assert.Equal(t, "{\"value\":false,\"mixed\":true,\"action\":\"none\"}", scanBefore.String())
		privateBefore := gjson.Get(editValues, "Private")
		assert.Equal(t, "{\"value\":false,\"mixed\":true,\"action\":\"none\"}", privateBefore.String())
		panoramaBefore := gjson.Get(editValues, "Panorama")
		assert.Equal(t, "{\"value\":false,\"mixed\":true,\"action\":\"none\"}", panoramaBefore.String())
		albumsBefore := gjson.Get(editValues, "Albums")
		assert.Contains(t, albumsBefore.String(), "{\"value\":\"as6sg6bipotaab19\",\"title\":\"IlikeFood\",\"mixed\":false,\"action\":\"none\"}")
		assert.Contains(t, albumsBefore.String(), "{\"value\":\"as6sg6bxpogaaba7\",\"title\":\"Christmas 2030\",\"mixed\":true,\"action\":\"none\"}")
		assert.Contains(t, albumsBefore.String(), "{\"value\":\"as6sg6bxpogaaba8\",\"title\":\"Holiday 2030\",\"mixed\":true,\"action\":\"none\"}")
		labelsBefore := gjson.Get(editValues, "Labels")
		assert.Contains(t, labelsBefore.String(), "{\"value\":\"ls6sg6b1wowuy316\",\"title\":\"\\u0026friendship\",\"mixed\":true,\"action\":\"none\"}")
		assert.Contains(t, labelsBefore.String(), "{\"value\":\"ls6sg6b1wowuy3c4\",\"title\":\"Cake\",\"mixed\":false,\"action\":\"none\"}")
		assert.Contains(t, labelsBefore.String(), "{\"value\":\"ls6sg6b1wowuy3c3\",\"title\":\"Flower\",\"mixed\":true,\"action\":\"none\"}")
		cameraBefore := gjson.Get(editValues, "CameraID")
		assert.Equal(t, "{\"value\":-2,\"mixed\":true,\"action\":\"none\"}", cameraBefore.String())
		lensBefore := gjson.Get(editValues, "LensID")
		assert.Equal(t, "{\"value\":-2,\"mixed\":true,\"action\":\"none\"}", lensBefore.String())
		isoBefore := gjson.Get(editValues, "Iso")
		assert.Equal(t, "{\"value\":0,\"mixed\":true,\"action\":\"none\"}", isoBefore.String())
		fNumberBefore := gjson.Get(editValues, "FNumber")
		assert.Equal(t, "{\"value\":3.5,\"mixed\":false,\"action\":\"none\"}", fNumberBefore.String())
		focalLengthBefore := gjson.Get(editValues, "FocalLength")
		assert.Equal(t, "{\"value\":0,\"mixed\":true,\"action\":\"none\"}", focalLengthBefore.String())
		exposureBefore := gjson.Get(editValues, "Exposure")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", exposureBefore.String())
		takenBefore := gjson.Get(editValues, "TakenAt")
		assert.Equal(t, "{\"value\":\"2018-12-01T09:08:18Z\",\"mixed\":true,\"action\":\"none\"}", takenBefore.String())
		takenLocalBefore := gjson.Get(editValues, "TakenAtLocal")
		assert.Equal(t, "{\"value\":\"2018-12-01T09:08:18Z\",\"mixed\":true,\"action\":\"none\"}", takenLocalBefore.String())
		keywordsBefore := gjson.Get(editValues, "DetailsKeywords")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", keywordsBefore.String())

		// Send the edit form values back to the same API endpoint and check for errors.
		saveResponse := PerformRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			fmt.Sprintf(`{"photos": %s, "values": %s}`, photoUIDs, editValues),
		)

		// Check the save response status code.
		assert.Equal(t, http.StatusOK, saveResponse.Code)

		// Check the save response body.
		saveBody := saveResponse.Body.String()
		assert.NotEmpty(t, saveBody)

		// Check the save response values.
		saveValues := gjson.Get(saveBody, "values").Raw
		//t.Logf("save values: %#v", saveValues)
		timezoneAfter := gjson.Get(saveValues, "TimeZone")
		assert.Equal(t, timezoneAfter.String(), timezoneBefore.String())
		altitudeAfter := gjson.Get(saveValues, "Altitude")
		assert.Equal(t, altitudeAfter.String(), altitudeBefore.String())
		countryAfter := gjson.Get(saveValues, "Country")
		assert.Equal(t, countryAfter.String(), countryBefore.String())
		latAfter := gjson.Get(saveValues, "Lat")
		assert.Equal(t, latAfter.String(), latBefore.String())
		lngAfter := gjson.Get(saveValues, "Lng")
		assert.Equal(t, lngAfter.String(), lngBefore.String())
		typeAfter := gjson.Get(saveValues, "Type")
		assert.Equal(t, typeAfter.String(), typeBefore.String())
		yearAfter := gjson.Get(saveValues, "Year")
		assert.Equal(t, yearAfter.String(), yearBefore.String())
		dayAfter := gjson.Get(saveValues, "Day")
		assert.Equal(t, dayAfter.String(), dayBefore.String())
		monthAfter := gjson.Get(saveValues, "Month")
		assert.Equal(t, monthAfter.String(), monthBefore.String())
		titleAfter := gjson.Get(saveValues, "Title")
		assert.Equal(t, titleAfter.String(), titleBefore.String())
		captionAfter := gjson.Get(saveValues, "Caption")
		assert.Equal(t, captionAfter.String(), captionBefore.String())
		subjectAfter := gjson.Get(saveValues, "DetailsSubject")
		assert.Equal(t, subjectAfter.String(), subjectBefore.String())
		artistAfter := gjson.Get(saveValues, "DetailsArtist")
		assert.Equal(t, artistAfter.String(), artistBefore.String())
		copyrightAfter := gjson.Get(saveValues, "DetailsCopyright")
		assert.Equal(t, copyrightAfter.String(), copyrightBefore.String())
		licenseAfter := gjson.Get(saveValues, "DetailsLicense")
		assert.Equal(t, licenseAfter.String(), licenseBefore.String())
		favoriteAfter := gjson.Get(saveValues, "Favorite")
		assert.Equal(t, favoriteAfter.String(), favoriteBefore.String())
		scanAfter := gjson.Get(saveValues, "Scan")
		assert.Equal(t, scanAfter.String(), scanBefore.String())
		privateAfter := gjson.Get(saveValues, "Private")
		assert.Equal(t, privateAfter.String(), privateBefore.String())
		panoramaAfter := gjson.Get(saveValues, "Panorama")
		assert.Equal(t, panoramaAfter.String(), panoramaBefore.String())
		albumsAfter := gjson.Get(saveValues, "Albums")
		assert.Equal(t, albumsAfter.String(), albumsBefore.String())
		labelsAfter := gjson.Get(saveValues, "Labels")
		assert.Equal(t, labelsAfter.String(), labelsBefore.String())
		cameraAfter := gjson.Get(saveValues, "CameraID")
		assert.Equal(t, cameraAfter.String(), cameraBefore.String())
		lensAfter := gjson.Get(saveValues, "LensID")
		assert.Equal(t, lensAfter.String(), lensBefore.String())
		isoAfter := gjson.Get(saveValues, "Iso")
		assert.Equal(t, isoAfter.String(), isoBefore.String())
		fNumberAfter := gjson.Get(saveValues, "FNumber")
		assert.Equal(t, fNumberAfter.String(), fNumberBefore.String())
		focalLengthAfter := gjson.Get(saveValues, "FocalLength")
		assert.Equal(t, focalLengthAfter.String(), focalLengthBefore.String())
		exposureAfter := gjson.Get(saveValues, "Exposure")
		assert.Equal(t, exposureAfter.String(), exposureBefore.String())
		takenAfter := gjson.Get(saveValues, "TakenAt")
		assert.Equal(t, takenAfter.String(), takenBefore.String())
		takenLocalAfter := gjson.Get(saveValues, "TakenAtLocal")
		assert.Equal(t, takenLocalAfter.String(), takenLocalBefore.String())
		//TODO Uncomment once keywords may be supported
		//keywordsAfter := gjson.Get(saveValues, "DetailsKeywords")
		//assert.Equal(t, keywordsAfter.String(), keywordsBefore.String())
		//assert.Equal(t, editValues, saveValues)
	})
	t.Run("SuccessChangeLocationValues", func(t *testing.T) {
		// Create new API test instance.
		app, router, _ := NewApiTest()

		// Attach POST /api/v1/batch/photos/edit request handler.
		BatchPhotosEdit(router)

		// Specify the unique IDs of the photos used for testing.
		photoUIDs := `["pqkm36fjqvset8uy", "pqkm36fjqvset9uz"]`

		// Get the photo models and current values for the batch edit form.
		editResponse := PerformRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			fmt.Sprintf(`{"photos": %s}`, photoUIDs),
		)

		// Check the edit response status code.
		assert.Equal(t, http.StatusOK, editResponse.Code)

		// Check the edit response body.
		editBody := editResponse.Body.String()
		assert.NotEmpty(t, editBody)

		// Check the edit response values.
		editPhotos := gjson.Get(editBody, "models").Array()
		assert.Equal(t, len(editPhotos), 2)
		editValues := gjson.Get(editBody, "values").Raw
		timezoneBefore := gjson.Get(editValues, "TimeZone")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", timezoneBefore.String())
		altitudeBefore := gjson.Get(editValues, "Altitude")
		assert.Equal(t, "{\"value\":0,\"mixed\":true,\"action\":\"none\"}", altitudeBefore.String())
		countryBefore := gjson.Get(editValues, "Country")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", countryBefore.String())
		latBefore := gjson.Get(editValues, "Lat")
		assert.Equal(t, "{\"value\":0,\"mixed\":true,\"action\":\"none\"}", latBefore.String())
		lngBefore := gjson.Get(editValues, "Lng")
		assert.Equal(t, "{\"value\":0,\"mixed\":true,\"action\":\"none\"}", lngBefore.String())
		// Send the edit form values back to the same API endpoint and check for errors.
		saveResponse := PerformRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			fmt.Sprintf(`{"photos": %s, "values": %s}`, photoUIDs,
				"{"+
					"\"Lat\":{\"value\":21.850195,\"mixed\":false,\"action\":\"update\"},"+
					"\"Lng\":{\"value\":90.18015,\"mixed\":false,\"action\":\"update\"}"+
					"}"),
		)

		// Check the save response status code.
		assert.Equal(t, http.StatusOK, saveResponse.Code)

		// Check the save response body.
		saveBody := saveResponse.Body.String()
		assert.NotEmpty(t, saveBody)

		// Check the save response values.
		saveValues := gjson.Get(saveBody, "values").Raw
		timezoneAfter := gjson.Get(saveValues, "TimeZone")
		assert.Equal(t, "{\"value\":\"Asia/Dhaka\",\"mixed\":false,\"action\":\"none\"}", timezoneAfter.String())
		altitudeAfter := gjson.Get(saveValues, "Altitude")
		assert.Equal(t, "{\"value\":0,\"mixed\":true,\"action\":\"none\"}", altitudeAfter.String())
		countryAfter := gjson.Get(saveValues, "Country")
		assert.Equal(t, "{\"value\":\"bd\",\"mixed\":false,\"action\":\"none\"}", countryAfter.String())
		latAfter := gjson.Get(saveValues, "Lat")
		assert.Equal(t, "{\"value\":21.850195,\"mixed\":false,\"action\":\"none\"}", latAfter.String())
		lngAfter := gjson.Get(saveValues, "Lng")
		assert.Equal(t, "{\"value\":90.18015,\"mixed\":false,\"action\":\"none\"}", lngAfter.String())

		GetPhoto(router)
		r1 := PerformRequest(app, "GET", "/api/v1/photos/pqkm36fjqvset9uz")
		assert.Equal(t, http.StatusOK, r1.Code)
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "PlaceSrc").String())
		assert.Equal(t, "meta", gjson.Get(r1.Body.String(), "TakenSrc").String())
		assert.Equal(t, "2018-12-01T03:08:18Z", gjson.Get(r1.Body.String(), "TakenAt").String())
		assert.Equal(t, "21.850195", gjson.Get(r1.Body.String(), "Lat").String())
		assert.Equal(t, "90.18015", gjson.Get(r1.Body.String(), "Lng").String())
		assert.Equal(t, "bd", gjson.Get(r1.Body.String(), "Country").String())
		assert.Equal(t, "Asia/Dhaka", gjson.Get(r1.Body.String(), "TimeZone").String())
	})
	t.Run("SuccessChangeValues", func(t *testing.T) {
		// Create new API test instance.
		app, router, _ := NewApiTest()

		// Attach POST /api/v1/batch/photos/edit request handler.
		BatchPhotosEdit(router)

		// Specify the unique IDs of the photos used for testing.
		photoUIDs := `["pqkm36fjqvset9uy", "pqkm36fjqvset9uz"]`

		// Get the photo models and current values for the batch edit form.
		editResponse := PerformRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			fmt.Sprintf(`{"photos": %s}`, photoUIDs),
		)

		// Check the edit response status code.
		assert.Equal(t, http.StatusOK, editResponse.Code)

		// Check the edit response body.
		editBody := editResponse.Body.String()
		assert.NotEmpty(t, editBody)

		// Check the edit response values.
		editPhotos := gjson.Get(editBody, "models").Array()
		assert.Equal(t, len(editPhotos), 2)
		editValues := gjson.Get(editBody, "values").Raw
		timezoneBefore := gjson.Get(editValues, "TimeZone")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", timezoneBefore.String())
		altitudeBefore := gjson.Get(editValues, "Altitude")
		assert.Equal(t, "{\"value\":0,\"mixed\":true,\"action\":\"none\"}", altitudeBefore.String())
		typeBefore := gjson.Get(editValues, "Type")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", typeBefore.String())
		yearBefore := gjson.Get(editValues, "Year")
		assert.Equal(t, "{\"value\":-2,\"mixed\":true,\"action\":\"none\"}", yearBefore.String())
		dayBefore := gjson.Get(editValues, "Day")
		assert.Equal(t, "{\"value\":-2,\"mixed\":true,\"action\":\"none\"}", dayBefore.String())
		monthBefore := gjson.Get(editValues, "Month")
		assert.Equal(t, "{\"value\":-2,\"mixed\":true,\"action\":\"none\"}", monthBefore.String())
		titleBefore := gjson.Get(editValues, "Title")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", titleBefore.String())
		captionBefore := gjson.Get(editValues, "Caption")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", captionBefore.String())
		subjectBefore := gjson.Get(editValues, "DetailsSubject")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", subjectBefore.String())
		artistBefore := gjson.Get(editValues, "DetailsArtist")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", artistBefore.String())
		copyrightBefore := gjson.Get(editValues, "DetailsCopyright")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", copyrightBefore.String())
		licenseBefore := gjson.Get(editValues, "DetailsLicense")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", licenseBefore.String())
		favoriteBefore := gjson.Get(editValues, "Favorite")
		assert.Equal(t, "{\"value\":false,\"mixed\":true,\"action\":\"none\"}", favoriteBefore.String())
		scanBefore := gjson.Get(editValues, "Scan")
		assert.Equal(t, "{\"value\":false,\"mixed\":true,\"action\":\"none\"}", scanBefore.String())
		privateBefore := gjson.Get(editValues, "Private")
		assert.Equal(t, "{\"value\":false,\"mixed\":true,\"action\":\"none\"}", privateBefore.String())
		panoramaBefore := gjson.Get(editValues, "Panorama")
		assert.Equal(t, "{\"value\":false,\"mixed\":true,\"action\":\"none\"}", panoramaBefore.String())
		// Send the edit form values back to the same API endpoint and check for errors.
		saveResponse := PerformRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			fmt.Sprintf(`{"photos": %s, "values": %s}`, photoUIDs,
				"{"+
					"\"TimeZone\":{\"value\":\"Europe/Vienna\",\"mixed\":false,\"action\":\"update\"},"+
					"\"Altitude\":{\"value\":145,\"mixed\":false,\"action\":\"update\"},"+
					"\"Year\":{\"value\":2000,\"mixed\":false,\"action\":\"update\"},"+
					"\"Month\":{\"value\":11,\"mixed\":true,\"action\":\"update\"},"+
					"\"Day\":{\"value\":-1,\"mixed\":true,\"action\":\"update\"},"+
					"\"Title\":{\"value\":\"My Batch Edited Title\",\"mixed\":false,\"action\":\"update\"},"+
					"\"Caption\":{\"value\":\"Batch edited caption\",\"mixed\":false,\"action\":\"update\"},"+
					"\"DetailsSubject\":{\"value\":\"Batch edited subject\",\"mixed\":false,\"action\":\"update\"},"+
					"\"DetailsArtist\":{\"value\":\"Batchie\",\"mixed\":false,\"action\":\"update\"},"+
					"\"DetailsCopyright\":{\"value\":\"Batch edited copyright\",\"mixed\":false,\"action\":\"update\"},"+
					"\"DetailsLicense\":{\"value\":\"Batch edited license\",\"mixed\":false,\"action\":\"update\"},"+
					"\"Type\":{\"value\":\"live\",\"mixed\":false,\"action\":\"update\"},"+
					"\"Favorite\":{\"value\":false,\"mixed\":false,\"action\":\"update\"},"+
					"\"Panorama\":{\"value\":true,\"mixed\":false,\"action\":\"update\"},"+
					"\"Private\":{\"value\":true,\"mixed\":false,\"action\":\"update\"},"+
					"\"Scan\":{\"value\":true,\"mixed\":false,\"action\":\"update\"}"+
					"}"),
		)

		// Check the save response status code.
		assert.Equal(t, http.StatusOK, saveResponse.Code)

		// Check the save response body.
		saveBody := saveResponse.Body.String()
		assert.NotEmpty(t, saveBody)

		// Check the save response values.
		saveValues := gjson.Get(saveBody, "values").Raw
		timezoneAfter := gjson.Get(saveValues, "TimeZone")
		assert.Equal(t, "{\"value\":\"Europe/Vienna\",\"mixed\":false,\"action\":\"none\"}", timezoneAfter.String())
		altitudeAfter := gjson.Get(saveValues, "Altitude")
		assert.Equal(t, "{\"value\":145,\"mixed\":false,\"action\":\"none\"}", altitudeAfter.String())
		typeAfter := gjson.Get(saveValues, "Type")
		assert.Equal(t, "{\"value\":\"live\",\"mixed\":false,\"action\":\"none\"}", typeAfter.String())
		yearAfter := gjson.Get(saveValues, "Year")
		assert.Equal(t, "{\"value\":2000,\"mixed\":false,\"action\":\"none\"}", yearAfter.String())
		dayAfter := gjson.Get(saveValues, "Day")
		assert.Equal(t, "{\"value\":-1,\"mixed\":false,\"action\":\"none\"}", dayAfter.String())
		monthAfter := gjson.Get(saveValues, "Month")
		assert.Equal(t, "{\"value\":11,\"mixed\":false,\"action\":\"none\"}", monthAfter.String())
		titleAfter := gjson.Get(saveValues, "Title")
		assert.Equal(t, "{\"value\":\"My Batch Edited Title\",\"mixed\":false,\"action\":\"none\"}", titleAfter.String())
		captionAfter := gjson.Get(saveValues, "Caption")
		assert.Equal(t, "{\"value\":\"Batch edited caption\",\"mixed\":false,\"action\":\"none\"}", captionAfter.String())
		subjectAfter := gjson.Get(saveValues, "DetailsSubject")
		assert.Equal(t, "{\"value\":\"Batch edited subject\",\"mixed\":false,\"action\":\"none\"}", subjectAfter.String())
		artistAfter := gjson.Get(saveValues, "DetailsArtist")
		assert.Equal(t, "{\"value\":\"Batchie\",\"mixed\":false,\"action\":\"none\"}", artistAfter.String())
		copyrightAfter := gjson.Get(saveValues, "DetailsCopyright")
		assert.Equal(t, "{\"value\":\"Batch edited copyright\",\"mixed\":false,\"action\":\"none\"}", copyrightAfter.String())
		licenseAfter := gjson.Get(saveValues, "DetailsLicense")
		assert.Equal(t, "{\"value\":\"Batch edited license\",\"mixed\":false,\"action\":\"none\"}", licenseAfter.String())
		favoriteAfter := gjson.Get(saveValues, "Favorite")
		assert.Equal(t, "{\"value\":false,\"mixed\":false,\"action\":\"none\"}", favoriteAfter.String())
		scanAfter := gjson.Get(saveValues, "Scan")
		assert.Equal(t, "{\"value\":true,\"mixed\":false,\"action\":\"none\"}", scanAfter.String())
		privateAfter := gjson.Get(saveValues, "Private")
		assert.Equal(t, "{\"value\":true,\"mixed\":false,\"action\":\"none\"}", privateAfter.String())
		panoramaAfter := gjson.Get(saveValues, "Panorama")
		assert.Equal(t, "{\"value\":true,\"mixed\":false,\"action\":\"none\"}", panoramaAfter.String())
		takenAfter := gjson.Get(saveValues, "TakenAt")
		assert.Contains(t, takenAfter.String(), "{\"value\":\"2000-11")
		takenLocalAfter := gjson.Get(saveValues, "TakenAtLocal")
		assert.Contains(t, takenLocalAfter.String(), "{\"value\":\"2000-11")

		GetPhoto(router)
		r1 := PerformRequest(app, "GET", "/api/v1/photos/pqkm36fjqvset9uz")
		assert.Equal(t, http.StatusOK, r1.Code)
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "PlaceSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "TakenSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "TypeSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "TitleSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "CaptionSrc").String())
		assert.Equal(t, "meta", gjson.Get(r1.Body.String(), "Details.KeywordsSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "Details.SubjectSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "Details.ArtistSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "Details.CopyrightSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "Details.LicenseSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "PlaceSrc").String())
		assert.Equal(t, "-1", gjson.Get(r1.Body.String(), "Day").String())
		assert.Equal(t, "11", gjson.Get(r1.Body.String(), "Month").String())
		assert.Equal(t, "2000", gjson.Get(r1.Body.String(), "Year").String())
		assert.Equal(t, "2000-11-01T08:08:18Z", gjson.Get(r1.Body.String(), "TakenAt").String())
		assert.Equal(t, "Europe/Vienna", gjson.Get(r1.Body.String(), "TimeZone").String())
		assert.Equal(t, "145", gjson.Get(r1.Body.String(), "Altitude").String())
		assert.Equal(t, "live", gjson.Get(r1.Body.String(), "Type").String())
		assert.Equal(t, "My Batch Edited Title", gjson.Get(r1.Body.String(), "Title").String())
		assert.Equal(t, "Batch edited caption", gjson.Get(r1.Body.String(), "Caption").String())
		assert.Equal(t, "Batch edited subject", gjson.Get(r1.Body.String(), "Details.Subject").String())
		assert.Equal(t, "Batchie", gjson.Get(r1.Body.String(), "Details.Artist").String())
		assert.Equal(t, "Batch edited copyright", gjson.Get(r1.Body.String(), "Details.Copyright").String())
		assert.Equal(t, "Batch edited license", gjson.Get(r1.Body.String(), "Details.License").String())
		assert.Equal(t, "true", gjson.Get(r1.Body.String(), "Panorama").String())
		assert.Equal(t, "false", gjson.Get(r1.Body.String(), "Favorite").String())
		assert.Equal(t, "true", gjson.Get(r1.Body.String(), "Private").String())
		assert.Equal(t, "true", gjson.Get(r1.Body.String(), "Scan").String())
	})
	t.Run("SuccessChangeAlbumAndLabels", func(t *testing.T) {
		// Create new API test instance.
		app, router, _ := NewApiTest()

		// Attach POST /api/v1/batch/photos/edit request handler.
		BatchPhotosEdit(router)

		// Specify the unique IDs of the photos used for testing.
		photoUIDs := `["pqkm36fjqvset9uy", "pqkm36fjqvset9uz"]`

		// Get the photo models and current values for the batch edit form.
		editResponse := PerformRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			fmt.Sprintf(`{"photos": %s}`, photoUIDs),
		)

		// Check the edit response status code.
		assert.Equal(t, http.StatusOK, editResponse.Code)

		// Check the edit response body.
		editBody := editResponse.Body.String()
		assert.NotEmpty(t, editBody)

		// Check the edit response values.
		editPhotos := gjson.Get(editBody, "models").Array()
		assert.Equal(t, len(editPhotos), 2)
		editValues := gjson.Get(editBody, "values").Raw
		//t.Logf(editValues)
		albumsBefore := gjson.Get(editValues, "Albums")
		assert.Contains(t, albumsBefore.String(), "{\"value\":\"as6sg6bipotaab19\",\"title\":\"IlikeFood\",\"mixed\":false,\"action\":\"none\"}")
		assert.Contains(t, albumsBefore.String(), "{\"value\":\"as6sg6bxpogaaba7\",\"title\":\"Christmas 2030\",\"mixed\":true,\"action\":\"none\"}")
		assert.Contains(t, albumsBefore.String(), "{\"value\":\"as6sg6bxpogaaba8\",\"title\":\"Holiday 2030\",\"mixed\":true,\"action\":\"none\"}")
		labelsBefore := gjson.Get(editValues, "Labels")
		assert.Contains(t, labelsBefore.String(), "{\"value\":\"ls6sg6b1wowuy316\",\"title\":\"\\u0026friendship\",\"mixed\":true,\"action\":\"none\"}")
		assert.Contains(t, labelsBefore.String(), "{\"value\":\"ls6sg6b1wowuy3c4\",\"title\":\"Cake\",\"mixed\":false,\"action\":\"none\"}")
		assert.Contains(t, labelsBefore.String(), "{\"value\":\"ls6sg6b1wowuy3c5\",\"title\":\"COW\",\"mixed\":false,\"action\":\"none\"}")
		assert.Contains(t, labelsBefore.String(), "{\"value\":\"ls6sg6b1wowuy3c2\",\"title\":\"Landscape\",\"mixed\":false,\"action\":\"none\"}")
		assert.Contains(t, labelsBefore.String(), "{\"value\":\"ls6sg6b1wowuy3c3\",\"title\":\"Flower\",\"mixed\":true,\"action\":\"none\"}")
		assert.Contains(t, labelsBefore.String(), "{\"value\":\"ls6sg6b1wowuy317\",\"title\":\"construction\\u0026failure\",\"mixed\":true,\"action\":\"none\"}")
		// Send the edit form values back to the same API endpoint and check for errors.
		saveResponse := PerformRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			fmt.Sprintf(`{"photos": %s, "values": %s}`, photoUIDs,
				"{"+
					"\"Labels\":{\"items\":[{\"value\":\"ls6sg6b1wowuy317\",\"title\":\"construction\\u0026failure\",\"mixed\":false,\"action\":\"remove\"},{\"value\":\"ls6sg6b1wowuy3c2\",\"title\":\"Landscape\",\"mixed\":false,\"action\":\"remove\"},{\"value\":\"ls6sg6b1wowuy3c5\",\"title\":\"COW\",\"mixed\":false,\"action\":\"remove\"},{\"value\":\"ls6sg6b1wowuy3c4\",\"title\":\"Cake\",\"mixed\":false,\"action\":\"remove\"},{\"value\":\"ls6sg6b1wowuy3c3\",\"title\":\"Flower\",\"mixed\":false,\"action\":\"add\"},{\"value\":\"ls6sg6b1wowuy316\",\"title\":\"&friendship\",\"mixed\":false,\"action\":\"remove\"},{\"value\":\"\",\"title\":\"BatchLabel\",\"mixed\":false,\"action\":\"add\"}],\"mixed\":false,\"action\":\"update\"},"+
					"\"Albums\":{\"items\":[{\"value\":\"as6sg6bipotaab19\",\"title\":\"IlikeFood\",\"mixed\":false,\"action\":\"remove\"},{\"value\":\"as6sg6bxpogaaba8\",\"title\":\"Holiday 2030\",\"mixed\":true,\"action\":\"none\"},{\"value\":\"as6sg6bxpogaaba7\",\"title\":\"Christmas 2030\",\"mixed\":false,\"action\":\"add\"}, {\"value\":\"\",\"title\":\"BatchAlbum\",\"mixed\":false,\"action\":\"add\"}],\"mixed\":true,\"action\":\"update\"}"+
					"}"),
		)

		// Check the save response status code.
		assert.Equal(t, http.StatusOK, saveResponse.Code)

		// Check the save response body.
		saveBody := saveResponse.Body.String()
		assert.NotEmpty(t, saveBody)

		// Check the save response values.
		saveValues := gjson.Get(saveBody, "values").Raw
		albumsAfter := gjson.Get(saveValues, "Albums")
		assert.Contains(t, albumsAfter.String(), "{\"value\":\"as6sg6bxpogaaba8\",\"title\":\"Holiday 2030\",\"mixed\":true,\"action\":\"none\"}")
		assert.Contains(t, albumsAfter.String(), "{\"value\":\"as6sg6bxpogaaba7\",\"title\":\"Christmas 2030\",\"mixed\":false,\"action\":\"none\"}")
		assert.Contains(t, albumsAfter.String(), "\"title\":\"BatchAlbum\",\"mixed\":false,\"action\":\"none\"}")
		assert.NotContains(t, albumsAfter.String(), "{\"value\":\"as6sg6bipotaab19\",\"title\":\"\\u0026IlikeFood\"")
		labelsAfter := gjson.Get(saveValues, "Labels")
		assert.Contains(t, labelsAfter.String(), "{\"value\":\"ls6sg6b1wowuy3c3\",\"title\":\"Flower\",\"mixed\":false,\"action\":\"none\"}")
		assert.NotContains(t, labelsAfter.String(), "{\"value\":\"ls6sg6b1wowuy3c4\",\"title\":\"Cake\"")
		assert.Contains(t, labelsAfter.String(), "\"title\":\"BatchLabel\",\"mixed\":false,\"action\":\"none\"}")
		assert.NotContains(t, labelsAfter.String(), "{\"value\":\"ls6sg6b1wowuy316\",\"title\":\"\\u0026friendship\"")
		assert.NotContains(t, labelsAfter.String(), "{\"value\":\"ls6sg6b1wowuy3c5\",\"title\":\"COW\",\"mixed\":false,\"action\":\"none\"}")
		assert.NotContains(t, labelsAfter.String(), "{\"value\":\"ls6sg6b1wowuy3c2\",\"title\":\"Landscape\",\"mixed\":false,\"action\":\"none\"}")
		assert.NotContains(t, labelsAfter.String(), "{\"value\":\"ls6sg6b1wowuy317\",\"title\":\"construction\\u0026failure\",\"mixed\":true,\"action\":\"none\"}")

		GetPhoto(router)
		r1 := PerformRequest(app, "GET", "/api/v1/photos/pqkm36fjqvset9uz")
		assert.Equal(t, http.StatusOK, r1.Code)
		assert.Equal(t, "BatchLabel", gjson.Get(r1.Body.String(), "Labels.0.Label.Name").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "Labels.0.LabelSrc").String())
		assert.Equal(t, "0", gjson.Get(r1.Body.String(), "Labels.0.Uncertainty").String())
		assert.Equal(t, "Flower", gjson.Get(r1.Body.String(), "Labels.1.Label.Name").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "Labels.1.LabelSrc").String())
		assert.Equal(t, "0", gjson.Get(r1.Body.String(), "Labels.1.Uncertainty").String())
		assert.Equal(t, "&friendship", gjson.Get(r1.Body.String(), "Labels.2.Label.Name").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "Labels.2.LabelSrc").String())
		assert.Equal(t, "100", gjson.Get(r1.Body.String(), "Labels.2.Uncertainty").String())
		assert.Equal(t, "COW", gjson.Get(r1.Body.String(), "Labels.3.Label.Name").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "Labels.3.LabelSrc").String())
		assert.Equal(t, "100", gjson.Get(r1.Body.String(), "Labels.3.Uncertainty").String())
		assert.Equal(t, "Cake", gjson.Get(r1.Body.String(), "Labels.4.Label.Name").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "Labels.4.LabelSrc").String())
		assert.Equal(t, "100", gjson.Get(r1.Body.String(), "Labels.4.Uncertainty").String())
		assert.Equal(t, "Landscape", gjson.Get(r1.Body.String(), "Labels.5.Label.Name").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "Labels.5.LabelSrc").String())
		assert.Equal(t, "100", gjson.Get(r1.Body.String(), "Labels.5.Uncertainty").String())
		assert.Equal(t, "", gjson.Get(r1.Body.String(), "Labels.6.Label.Name").String())

		batchLabelUid := gjson.Get(r1.Body.String(), "Labels.0.Label.UID").String()

		r2 := PerformRequest(app, "GET", "/api/v1/photos/pqkm36fjqvset9uy")
		assert.Equal(t, http.StatusOK, r2.Code)
		assert.Equal(t, "BatchLabel", gjson.Get(r2.Body.String(), "Labels.0.Label.Name").String())
		assert.Equal(t, "batch", gjson.Get(r2.Body.String(), "Labels.0.LabelSrc").String())
		assert.Equal(t, "0", gjson.Get(r2.Body.String(), "Labels.0.Uncertainty").String())
		assert.Equal(t, "Flower", gjson.Get(r2.Body.String(), "Labels.1.Label.Name").String())
		assert.Equal(t, "batch", gjson.Get(r2.Body.String(), "Labels.1.LabelSrc").String())
		assert.Equal(t, "0", gjson.Get(r2.Body.String(), "Labels.1.Uncertainty").String())
		assert.Equal(t, "COW", gjson.Get(r2.Body.String(), "Labels.2.Label.Name").String())
		assert.Equal(t, "batch", gjson.Get(r2.Body.String(), "Labels.2.LabelSrc").String())
		assert.Equal(t, "100", gjson.Get(r2.Body.String(), "Labels.2.Uncertainty").String())
		assert.Equal(t, "Landscape", gjson.Get(r2.Body.String(), "Labels.3.Label.Name").String())
		assert.Equal(t, "batch", gjson.Get(r2.Body.String(), "Labels.3.LabelSrc").String())
		assert.Equal(t, "100", gjson.Get(r2.Body.String(), "Labels.3.Uncertainty").String())
		assert.Equal(t, "", gjson.Get(r2.Body.String(), "Labels.4.Label.Name").String())

		// Get the photo models and current values for the batch edit form.
		editResponse2 := PerformRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			fmt.Sprintf(`{"photos": %s}`, photoUIDs),
		)

		// Check the edit response status code.
		assert.Equal(t, http.StatusOK, editResponse2.Code)

		// Send the edit form values back to the same API endpoint and check for errors.
		saveResponse2 := PerformRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			fmt.Sprintf(`{"photos": %s, "values": %s}`, photoUIDs,
				"{"+
					"\"Labels\":{\"items\":[{\"value\":\""+batchLabelUid+"\",\"title\":\"BatchLabel\",\"mixed\":false,\"action\":\"remove\"}],\"mixed\":false,\"action\":\"update\"}"+
					"}"),
		)

		// Check the save response status code.
		assert.Equal(t, http.StatusOK, saveResponse2.Code)

		// Check the save response body.
		saveBody2 := saveResponse2.Body.String()
		assert.NotEmpty(t, saveBody2)

		saveValues2 := gjson.Get(saveBody2, "values").Raw
		labelsAfter2 := gjson.Get(saveValues2, "Labels")
		assert.NotContains(t, labelsAfter2.String(), "\"title\":\"BatchLabel\",\"mixed\":false,\"action\":\"none\"}")
		assert.NotContains(t, labelsAfter2.String(), "{\"value\":\"ls6sg6b1wowuy3c4\",\"title\":\"Cake\"")
		assert.NotContains(t, labelsAfter2.String(), "{\"value\":\"ls6sg6b1wowuy316\",\"title\":\"\\u0026friendship\"")
		assert.NotContains(t, labelsAfter2.String(), "{\"value\":\"ls6sg6b1wowuy3c5\",\"title\":\"COW\",\"mixed\":false,\"action\":\"none\"}")
		assert.NotContains(t, labelsAfter2.String(), "{\"value\":\"ls6sg6b1wowuy3c2\",\"title\":\"Landscape\",\"mixed\":false,\"action\":\"none\"}")
		assert.NotContains(t, labelsAfter2.String(), "{\"value\":\"ls6sg6b1wowuy317\",\"title\":\"construction\\u0026failure\",\"mixed\":true,\"action\":\"none\"}")
	})
	t.Run("SuccessChangeCountry", func(t *testing.T) {
		// Create new API test instance.
		app, router, _ := NewApiTest()

		// Attach POST /api/v1/batch/photos/edit request handler.
		BatchPhotosEdit(router)

		// Specify the unique IDs of the photos used for testing.
		photoUIDs := `["pqkm36fjqvset9uy", "pqkm36fjqvset9uz"]`

		// Get the photo models and current values for the batch edit form.
		editResponse := PerformRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			fmt.Sprintf(`{"photos": %s}`, photoUIDs),
		)

		// Check the edit response status code.
		assert.Equal(t, http.StatusOK, editResponse.Code)

		// Check the edit response body.
		editBody := editResponse.Body.String()
		assert.NotEmpty(t, editBody)

		// Check the edit response values.
		editPhotos := gjson.Get(editBody, "models").Array()
		assert.Equal(t, len(editPhotos), 2)
		editValues := gjson.Get(editBody, "values").Raw
		timezoneBefore := gjson.Get(editValues, "TimeZone")
		assert.Equal(t, "{\"value\":\"Europe/Vienna\",\"mixed\":false,\"action\":\"none\"}", timezoneBefore.String())
		altitudeBefore := gjson.Get(editValues, "Altitude")
		assert.Equal(t, "{\"value\":145,\"mixed\":false,\"action\":\"none\"}", altitudeBefore.String())
		countryBefore := gjson.Get(editValues, "Country")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":true,\"action\":\"none\"}", countryBefore.String())
		latBefore := gjson.Get(editValues, "Lat")
		assert.Equal(t, "{\"value\":0,\"mixed\":true,\"action\":\"none\"}", latBefore.String())
		lngBefore := gjson.Get(editValues, "Lng")
		assert.Equal(t, "{\"value\":0,\"mixed\":true,\"action\":\"none\"}", lngBefore.String())
		// Send the edit form values back to the same API endpoint and check for errors.
		saveResponse := PerformRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			fmt.Sprintf(`{"photos": %s, "values": %s}`, photoUIDs,
				"{"+
					"\"Country\":{\"value\":\"gb\",\"mixed\":false,\"action\":\"update\"},"+
					"\"Lat\":{\"value\":0,\"mixed\":false,\"action\":\"update\"},"+
					"\"Lng\":{\"value\":0,\"mixed\":false,\"action\":\"update\"}"+
					"}"),
		)

		// Check the save response status code.
		assert.Equal(t, http.StatusOK, saveResponse.Code)

		// Check the save response body.
		saveBody := saveResponse.Body.String()
		assert.NotEmpty(t, saveBody)

		// Check the save response values.
		saveValues := gjson.Get(saveBody, "values").Raw
		timezoneAfter := gjson.Get(saveValues, "TimeZone")
		assert.Equal(t, "{\"value\":\"Europe/Vienna\",\"mixed\":false,\"action\":\"none\"}", timezoneAfter.String())
		altitudeAfter := gjson.Get(saveValues, "Altitude")
		assert.Equal(t, "{\"value\":145,\"mixed\":false,\"action\":\"none\"}", altitudeAfter.String())
		countryAfter := gjson.Get(saveValues, "Country")
		assert.Equal(t, "{\"value\":\"gb\",\"mixed\":false,\"action\":\"none\"}", countryAfter.String())
		latAfter := gjson.Get(saveValues, "Lat")
		assert.Equal(t, "{\"value\":0,\"mixed\":false,\"action\":\"none\"}", latAfter.String())
		lngAfter := gjson.Get(saveValues, "Lng")
		assert.Equal(t, "{\"value\":0,\"mixed\":false,\"action\":\"none\"}", lngAfter.String())

		GetPhoto(router)
		r1 := PerformRequest(app, "GET", "/api/v1/photos/pqkm36fjqvset9uz")
		assert.Equal(t, http.StatusOK, r1.Code)
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "PlaceSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "TakenSrc").String())
		assert.Equal(t, "2000-11-01T08:08:18Z", gjson.Get(r1.Body.String(), "TakenAt").String())
		assert.Equal(t, "0", gjson.Get(r1.Body.String(), "Lat").String())
		assert.Equal(t, "0", gjson.Get(r1.Body.String(), "Lng").String())
		assert.Equal(t, "gb", gjson.Get(r1.Body.String(), "Country").String())
		assert.Equal(t, "Europe/Vienna", gjson.Get(r1.Body.String(), "TimeZone").String())
	})
	t.Run("SuccessRemoveValues", func(t *testing.T) {
		// Create new API test instance.
		app, router, _ := NewApiTest()

		// Attach POST /api/v1/batch/photos/edit request handler.
		BatchPhotosEdit(router)

		// Specify the unique IDs of the photos used for testing.
		photoUIDs := `["pqkm36fjqvset9uy", "pqkm36fjqvset9uz"]`

		// Get the photo models and current values for the batch edit form.
		editResponse := PerformRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			fmt.Sprintf(`{"photos": %s}`, photoUIDs),
		)

		// Check the edit response status code.
		assert.Equal(t, http.StatusOK, editResponse.Code)

		// Check the edit response body.
		editBody := editResponse.Body.String()
		assert.NotEmpty(t, editBody)

		// Check the edit response values.
		editPhotos := gjson.Get(editBody, "models").Array()
		assert.Equal(t, len(editPhotos), 2)
		// Send the edit form values back to the same API endpoint and check for errors.
		saveResponse := PerformRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			fmt.Sprintf(`{"photos": %s, "values": %s}`, photoUIDs,
				"{"+
					"\"Altitude\":{\"value\":0,\"mixed\":false,\"action\":\"update\"},"+
					"\"Year\":{\"value\":-1,\"mixed\":false,\"action\":\"update\"},"+
					"\"Month\":{\"value\":-1,\"mixed\":false,\"action\":\"update\"},"+
					"\"Day\":{\"value\":-1,\"mixed\":false,\"action\":\"update\"},"+
					"\"Title\":{\"value\":\"\",\"mixed\":false,\"action\":\"remove\"},"+
					"\"Caption\":{\"value\":\"\",\"mixed\":false,\"action\":\"remove\"},"+
					"\"DetailsSubject\":{\"value\":\"\",\"mixed\":false,\"action\":\"remove\"},"+
					"\"DetailsArtist\":{\"value\":\"\",\"mixed\":false,\"action\":\"remove\"},"+
					"\"DetailsCopyright\":{\"value\":\"\",\"mixed\":false,\"action\":\"remove\"},"+
					"\"DetailsLicense\":{\"value\":\"\",\"mixed\":false,\"action\":\"remove\"}"+
					"}"),
		)

		// Check the save response status code.
		assert.Equal(t, http.StatusOK, saveResponse.Code)

		// Check the save response body.
		saveBody := saveResponse.Body.String()
		assert.NotEmpty(t, saveBody)

		// Check the save response values.
		saveValues := gjson.Get(saveBody, "values").Raw
		altitudeAfter := gjson.Get(saveValues, "Altitude")
		assert.Equal(t, "{\"value\":0,\"mixed\":false,\"action\":\"none\"}", altitudeAfter.String())
		yearAfter := gjson.Get(saveValues, "Year")
		assert.Equal(t, "{\"value\":-1,\"mixed\":false,\"action\":\"none\"}", yearAfter.String())
		dayAfter := gjson.Get(saveValues, "Day")
		assert.Equal(t, "{\"value\":-1,\"mixed\":false,\"action\":\"none\"}", dayAfter.String())
		monthAfter := gjson.Get(saveValues, "Month")
		assert.Equal(t, "{\"value\":-1,\"mixed\":false,\"action\":\"none\"}", monthAfter.String())
		titleAfter := gjson.Get(saveValues, "Title")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":false,\"action\":\"none\"}", titleAfter.String())
		captionAfter := gjson.Get(saveValues, "Caption")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":false,\"action\":\"none\"}", captionAfter.String())
		subjectAfter := gjson.Get(saveValues, "DetailsSubject")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":false,\"action\":\"none\"}", subjectAfter.String())
		artistAfter := gjson.Get(saveValues, "DetailsArtist")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":false,\"action\":\"none\"}", artistAfter.String())
		copyrightAfter := gjson.Get(saveValues, "DetailsCopyright")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":false,\"action\":\"none\"}", copyrightAfter.String())
		licenseAfter := gjson.Get(saveValues, "DetailsLicense")
		assert.Equal(t, "{\"value\":\"\",\"mixed\":false,\"action\":\"none\"}", licenseAfter.String())

		GetPhoto(router)
		r1 := PerformRequest(app, "GET", "/api/v1/photos/pqkm36fjqvset9uz")
		assert.Equal(t, http.StatusOK, r1.Code)
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "PlaceSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "TakenSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "TypeSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "TitleSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "CaptionSrc").String())
		assert.Equal(t, "meta", gjson.Get(r1.Body.String(), "Details.KeywordsSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "Details.SubjectSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "Details.ArtistSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "Details.CopyrightSrc").String())
		assert.Equal(t, "batch", gjson.Get(r1.Body.String(), "Details.LicenseSrc").String())
		assert.Equal(t, "-1", gjson.Get(r1.Body.String(), "Day").String())
		assert.Equal(t, "-1", gjson.Get(r1.Body.String(), "Month").String())
		assert.Equal(t, "-1", gjson.Get(r1.Body.String(), "Year").String())
		assert.Equal(t, "2000-11-01T08:08:18Z", gjson.Get(r1.Body.String(), "TakenAt").String())
		assert.Equal(t, "Europe/Vienna", gjson.Get(r1.Body.String(), "TimeZone").String())
		assert.Equal(t, "0", gjson.Get(r1.Body.String(), "Altitude").String())
		assert.Equal(t, "", gjson.Get(r1.Body.String(), "Title").String())
		assert.Equal(t, "", gjson.Get(r1.Body.String(), "Caption").String())
		assert.Equal(t, "", gjson.Get(r1.Body.String(), "Details.Subject").String())
		assert.Equal(t, "", gjson.Get(r1.Body.String(), "Details.Artist").String())
		assert.Equal(t, "", gjson.Get(r1.Body.String(), "Details.Copyright").String())
		assert.Equal(t, "", gjson.Get(r1.Body.String(), "Details.License").String())

		r2 := PerformRequest(app, "GET", "/api/v1/photos/pqkm36fjqvset9uy")
		assert.Equal(t, http.StatusOK, r2.Code)
		assert.Equal(t, "batch", gjson.Get(r2.Body.String(), "PlaceSrc").String())
		assert.Equal(t, "batch", gjson.Get(r2.Body.String(), "TakenSrc").String())
		assert.Equal(t, "batch", gjson.Get(r2.Body.String(), "TypeSrc").String())
		assert.Equal(t, "batch", gjson.Get(r2.Body.String(), "TitleSrc").String())
		assert.Equal(t, "batch", gjson.Get(r2.Body.String(), "CaptionSrc").String())
		assert.Equal(t, "meta", gjson.Get(r2.Body.String(), "Details.KeywordsSrc").String())
		assert.Equal(t, "batch", gjson.Get(r2.Body.String(), "Details.SubjectSrc").String())
		assert.Equal(t, "batch", gjson.Get(r2.Body.String(), "Details.ArtistSrc").String())
		assert.Equal(t, "batch", gjson.Get(r2.Body.String(), "Details.CopyrightSrc").String())
		assert.Equal(t, "batch", gjson.Get(r2.Body.String(), "Details.LicenseSrc").String())
		assert.Equal(t, "-1", gjson.Get(r2.Body.String(), "Day").String())
		assert.Equal(t, "-1", gjson.Get(r2.Body.String(), "Month").String())
		assert.Equal(t, "-1", gjson.Get(r2.Body.String(), "Year").String())
		assert.Equal(t, "2000-11-01T08:08:18Z", gjson.Get(r2.Body.String(), "TakenAt").String())
		assert.Equal(t, "Europe/Vienna", gjson.Get(r2.Body.String(), "TimeZone").String())
		assert.Equal(t, "0", gjson.Get(r2.Body.String(), "Altitude").String())
		assert.Equal(t, "", gjson.Get(r2.Body.String(), "Title").String())
		assert.Equal(t, "", gjson.Get(r2.Body.String(), "Caption").String())
		assert.Equal(t, "", gjson.Get(r2.Body.String(), "Details.Subject").String())
		assert.Equal(t, "", gjson.Get(r2.Body.String(), "Details.Artist").String())
		assert.Equal(t, "", gjson.Get(r2.Body.String(), "Details.Copyright").String())
		assert.Equal(t, "", gjson.Get(r2.Body.String(), "Details.License").String())
	})
	t.Run("ReturnPhotosAndValues", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		authToken := AuthenticateUser(app, router, "alice", "Alice123!")

		// Attach POST /api/v1/batch/photos/edit request handler.
		BatchPhotosEdit(router)

		response := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/batch/photos/edit",
			`{"photos": ["ps6sg6be2lvl0yh7","ps6sg6be2lvl0yh8","ps6sg6byk7wrbk47","ps6sg6be2lvl0yh0"], "return": true, "values": {}}`,
			authToken)

		body := response.Body.String()

		assert.NotEmpty(t, body)
		assert.True(t, strings.HasPrefix(body, `{"models":[{"ID"`), "unexpected response")

		// fmt.Println(body)
		/* models := gjson.Get(body, "models")
		values := gjson.Get(body, "values")
		t.Logf("models: %#v", models)
		t.Logf("values: %#v", values) */

		assert.Equal(t, http.StatusOK, response.Code)
	})
	t.Run("MissingSelection", func(t *testing.T) {
		app, router, _ := NewApiTest()

		// Attach POST /api/v1/batch/photos/edit request handler.
		BatchPhotosEdit(router)

		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/edit", `{"photos": [], "return": true}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrNoItemsSelected), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()

		// Attach POST /api/v1/batch/photos/edit request handler.
		BatchPhotosEdit(router)

		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/edit", `{"photos": 123, "return": true}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("ReturnValuesAsAdmin", func(t *testing.T) {
		app, router, conf := NewApiTest()

		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		// Attach POST /api/v1/batch/photos/edit request handler.
		BatchPhotosEdit(router)

		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		response := AuthenticatedRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			`{"photos": ["ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh8"]}`,
			sessId,
		)

		body := response.Body.String()

		assert.NotEmpty(t, body)
		assert.True(t, strings.HasPrefix(body, `{"models":[{"ID"`), "unexpected response")

		assert.Equal(t, http.StatusOK, response.Code)
	})
	t.Run("ReturnValuesAsGuest", func(t *testing.T) {
		app, router, conf := NewApiTest()

		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		// Attach POST /api/v1/batch/photos/edit request handler.
		BatchPhotosEdit(router)

		sessId := AuthenticateUser(app, router, "gandalf", "Gandalf123!")

		response := AuthenticatedRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			`{"photos": ["ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh8"]}`,
			sessId,
		)

		if response.Code != http.StatusForbidden {
			t.Fatal(response.Body.String())
		}

		val := gjson.Get(response.Body.String(), "error")
		assert.Equal(t, "Permission denied", val.String())
	})

	// This covers the case where a label was added via batch (uncertainty=0, source=batch),
	// then removed, and later another batch edit is performed. Previously, the removed label
	// could reappear with 75% confidence and source=keyword because Details.Keywords were not
	// persisted before reload. The fix persists Details immediately after keyword removal.
	t.Run("RemovedLabelDoesNotReappearFromKeyword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		authToken := AuthenticateUser(app, router, "alice", "Alice123!")

		BatchPhotosEdit(router)

		photoUID := "pqkm36fjqvset9uz"
		flowerLabelPtr, err := entity.FindLabel("Flower", false)
		if err != nil || flowerLabelPtr == nil || !flowerLabelPtr.HasID() {
			t.Fatalf("fixture label 'Flower' not found: %v", err)
		}
		cakeLabelPtr, err := entity.FindLabel("Cake", false)
		if err != nil || cakeLabelPtr == nil || !cakeLabelPtr.HasID() {
			t.Fatalf("fixture label 'Cake' not found: %v", err)
		}

		addBody := fmt.Sprintf(`{"photos":["%s"],"values":{"Labels":{"action":"update","items":[{"action":"add","value":"%s"}]}}}`, photoUID, flowerLabelPtr.LabelUID)
		resp1 := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/batch/photos/edit", addBody, authToken)
		if resp1.Code != http.StatusOK {
			t.Fatalf("add label failed: %s", resp1.Body.String())
		}

		p, err := query.PhotoPreloadByUID(clean.UID(photoUID))
		if err != nil {
			t.Fatal(err)
		}

		var pl entity.PhotoLabel
		if err := entity.Db().Where("photo_id = ? AND label_id = ?", p.ID, flowerLabelPtr.ID).First(&pl).Error; err != nil {
			t.Fatalf("photo-label missing after add: %v", err)
		}
		if pl.Uncertainty != 0 {
			t.Fatalf("expected uncertainty 0 after add, got %d", pl.Uncertainty)
		}
		if pl.LabelSrc != entity.SrcBatch {
			t.Fatalf("expected label src 'batch' after add, got %s", pl.LabelSrc)
		}

		removeBody := fmt.Sprintf(`{"photos":["%s"],"values":{"Labels":{"action":"update","items":[{"action":"remove","value":"%s"}]}}}`, photoUID, flowerLabelPtr.LabelUID)
		resp2 := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/batch/photos/edit", removeBody, authToken)
		if resp2.Code != http.StatusOK {
			t.Fatalf("remove label failed: %s", resp2.Body.String())
		}

		var removed entity.PhotoLabel
		err = entity.Db().Where("photo_id = ? AND label_id = ?", p.ID, flowerLabelPtr.ID).First(&removed).Error
		if err == nil {
			t.Fatalf("expected photo-label to be deleted, but it exists")
		}

		addSecondBody := fmt.Sprintf(`{"photos":["%s"],"values":{"Labels":{"action":"update","items":[{"action":"add","value":"%s"}]}}}`, photoUID, cakeLabelPtr.LabelUID)
		resp3 := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/batch/photos/edit", addSecondBody, authToken)
		if resp3.Code != http.StatusOK {
			t.Fatalf("add second label failed: %s", resp3.Body.String())
		}

		var re entity.PhotoLabel
		reErr := entity.Db().Where("photo_id = ? AND label_id = ?", p.ID, flowerLabelPtr.ID).First(&re).Error
		if reErr == nil {
			t.Fatalf("removed label reappeared unexpectedly (src=%s uncertainty=%d)", re.LabelSrc, re.Uncertainty)
		}
	})
}
