package api

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/i18n"
)

func TestGetPhoto(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetPhoto(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Iso")
		assert.Equal(t, "200", val.String())
	})
	t.Run("AliceAppPassword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetPhoto(router)
		r := AuthenticatedRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7", "X3B6IU-hfeLG5-HpVxkT-ctCY3M")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Iso")
		assert.Equal(t, "200", val.String())
	})
	t.Run("AliceAppPasswordWebdav", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetPhoto(router)
		r := AuthenticatedRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7", "v2wS72-OkqEzm-MQ63Z2-TEhU0w")
		assert.Equal(t, http.StatusForbidden, r.Code)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Permission denied", val.String())
	})
	t.Run("AccessToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetPhoto(router)
		r := AuthenticatedRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7", "8e154d323800393faf5177ce7392116feebbf674e6c2d39e")
		assert.Equal(t, http.StatusOK, r.Code)
		val := gjson.Get(r.Body.String(), "Iso")
		assert.Equal(t, "200", val.String())
	})
	t.Run("InvalidAppPassword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetPhoto(router)
		r := AuthenticatedRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7", "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7xxx")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
		val := gjson.Get(r.Body.String(), "Iso")
		assert.Equal(t, "", val.String())
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetPhoto(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/xxx")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("GuestDeniedPrivate", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetPhoto(router)
		sessId := AuthenticateUser(app, router, "gandalf", "Gandalf123!")
		// A private picture outside the guest's shared scope is reported as not found.
		r := AuthenticatedRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0y13", sessId)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("AdminSeesPrivate", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetPhoto(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")
		r := AuthenticatedRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0y13", sessId)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestUpdatePhoto(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhoto(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/ps6sg6be2lvl0y13", `{"Title": "Updated01", "Country": "de"}`)
		val := gjson.Get(r.Body.String(), "Title")
		assert.Equal(t, "Updated01", val.String())
		val2 := gjson.Get(r.Body.String(), "Country")
		assert.Equal(t, "de", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("BadRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhoto(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/ps6sg6be2lvl0y13", `{"Name": "Updated01", "Country": 123}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdatePhoto(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/photos/xxx", `{"Name": "Updated01", "Country": "de"}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrEntityNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestGetPhotoDownload(t *testing.T) {
	t.Run("OriginalMissing", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetPhotoDownload(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7/dl?t="+conf.DownloadToken())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetPhotoDownload(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/xxx/dl?t="+conf.DownloadToken())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("InvalidToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		GetPhotoDownload(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7/dl?t=xxx")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}

// TestLikePhoto checks photo reactions and scoped response shaping.
func TestLikePhoto(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		LikePhoto(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh9/like")
		assert.Equal(t, http.StatusOK, r.Code)
		GetPhoto(router)
		r2 := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh9")
		val := gjson.Get(r2.Body.String(), "Favorite")
		assert.Equal(t, "true", val.String())
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		LikePhoto(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/xxx/like")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("GuestDeniedOutOfScope", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		LikePhoto(router)
		// A guest holds ActionReact and so passes the route ACL, but a picture outside its shared
		// scope must be reported as not found rather than returning the (unredacted) photo record.
		sessId := AuthenticateUser(app, router, "gandalf", "Gandalf123!")
		r := AuthenticatedRequest(app, "POST", "/api/v1/photos/ps6sg6be2lvl0y13/like", sessId)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("GuestInScopeRedacted", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		guest := entity.FindUserByName("guest")
		if guest == nil {
			t.Fatal("guest fixture is missing")
		}
		var priorPhoto entity.Photo
		if err := entity.UnscopedDb().Where("photo_uid = ?", "ps6sg6be2lvl0y21").First(&priorPhoto).Error; err != nil {
			t.Fatal(err)
		}
		shareUID := "as6sg6bxpogaaba8"
		prior := entity.FindUserShare(entity.UserShare{UserUID: guest.UserUID, ShareUID: shareUID})
		var link entity.Link
		if err := entity.Db().Where("link_token = ?", "1jxf3jfn2k").First(&link).Error; err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if prior == nil {
				assert.NoError(t, entity.UnscopedDb().Delete(&entity.UserShare{}, "user_uid = ? AND share_uid = ?", guest.UserUID, shareUID).Error)
			} else {
				assert.NoError(t, entity.UnscopedDb().Save(prior).Error)
			}
			guest.RefreshShares()
			entity.UserFixtures.Pointer("guest").RefreshShares()
			assert.NoError(t, entity.UnscopedDb().Model(&entity.Link{}).Where("link_uid = ?", link.LinkUID).
				UpdateColumn("link_views", link.LinkViews).Error)
			var restored entity.Link
			assert.NoError(t, entity.Db().Where("link_uid = ?", link.LinkUID).First(&restored).Error)
			assert.Equal(t, link.LinkViews, restored.LinkViews)
		})

		// A guest with an album share receives the picture with identifying details redacted.
		sess := entity.NewSession(conf.SessionMaxAge(), 0)
		sess.SetUser(guest)
		sess.RedeemToken("1jxf3jfn2k")
		if err := sess.Save(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { assert.NoError(t, sess.Delete()) })

		LikePhoto(router)
		r := AuthenticatedRequest(app, "POST", "/api/v1/photos/ps6sg6be2lvl0y21/like", sess.AuthToken())
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "ps6sg6be2lvl0y21", gjson.Get(r.Body.String(), "photo.UID").String())
		var liked entity.Photo
		assert.NoError(t, entity.Db().Where("id = ?", priorPhoto.ID).First(&liked).Error)
		assert.Equal(t, priorPhoto.PhotoFavorite, liked.PhotoFavorite)
		// A non-redacted field stays present (the genuine record is returned), while the labels are
		// stripped — the fixture attaches one, so the empty result proves redaction.
		assert.Equal(t, "Title", gjson.Get(r.Body.String(), "photo.Title").String())
		assert.Empty(t, gjson.Get(r.Body.String(), "photo.Labels").Array())
		// The storage path stays: search returns it to every in-scope session.
		assert.Equal(t, "2018/11", gjson.Get(r.Body.String(), "photo.Path").String())
	})
}

func TestDislikePhoto(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		DislikePhoto(router)
		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh8/like")
		assert.Equal(t, http.StatusOK, r.Code)
		GetPhoto(router)
		r2 := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh8")
		val := gjson.Get(r2.Body.String(), "Favorite")
		assert.Equal(t, "false", val.String())
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		DislikePhoto(router)
		r := PerformRequest(app, "DELETE", "/api/v1/photos/xxx/like")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("GuestDeniedOutOfScope", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DislikePhoto(router)
		sessId := AuthenticateUser(app, router, "gandalf", "Gandalf123!")
		r := AuthenticatedRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0y13/like", sessId)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestPhotoPrimary(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PhotoPrimary(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh8/files/fs6sg6bw45bn0003/primary")
		assert.Equal(t, http.StatusOK, r.Code)
		GetFile(router)
		r2 := PerformRequest(app, "GET", "/api/v1/files/ocad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		val := gjson.Get(r2.Body.String(), "Primary")
		assert.Equal(t, "true", val.String())
		r3 := PerformRequest(app, "GET", "/api/v1/files/3cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		val2 := gjson.Get(r3.Body.String(), "Primary")
		assert.Equal(t, "false", val2.String())
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PhotoPrimary(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/xxx/files/fs6sg6bw45bnlqdw/primary")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrEntityNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestGetPhotoYaml(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetPhotoYaml(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0yh7/yaml")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetPhotoYaml(router)
		r := PerformRequest(app, "GET", "/api/v1/photos/xxx/yaml")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestApprovePhoto(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetPhoto(router)
		r3 := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6bexxvl0y20")
		val2 := gjson.Get(r3.Body.String(), "Quality")
		assert.Equal(t, "1", val2.String())
		ApprovePhoto(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/ps6sg6bexxvl0y20/approve")
		assert.Equal(t, http.StatusOK, r.Code)
		r2 := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6bexxvl0y20")
		val := gjson.Get(r2.Body.String(), "Quality")
		assert.Equal(t, "3", val.String())
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		ApprovePhoto(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/xxx/approve")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

// createTimeZoneTestPhoto creates a photo with a manual capture time and removes it after the test.
func createTimeZoneTestPhoto(t *testing.T, zone, taken, local string, year, month, day int) entity.Photo {
	t.Helper()
	photo := entity.NewPhoto(false)
	photo.TakenAt, _ = time.Parse(time.RFC3339, taken)
	photo.TakenAtLocal, _ = time.Parse(time.RFC3339, local)
	photo.TimeZone, photo.TakenSrc = zone, entity.SrcManual
	photo.PhotoYear, photo.PhotoMonth, photo.PhotoDay = year, month, day
	require.NoError(t, photo.Create())
	t.Cleanup(func() {
		entity.UnscopedDb().Delete(&entity.PhotoKeyword{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Delete(&entity.PhotoLabel{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Delete(&photo)
	})
	return photo
}

// TestUpdatePhoto_TimeZone checks the requests the photo editors send for date and zone edits, followed by unrelated saves.
func TestUpdatePhoto_TimeZone(t *testing.T) {
	cases := []struct {
		name, zone, taken, local, body, expectedTaken, expectedLocal string
		year, month, day                                             int
	}{
		{"UTCToBerlin", "UTC", "2025-06-30T23:30:00Z", "2025-06-30T23:30:00Z", `{"TakenAtLocal":"2025-07-01T01:30:00Z","TimeZone":"Europe/Berlin","Month":7,"Day":1,"TakenSrc":"manual"}`, "2025-06-30T23:30:00Z", "2025-07-01T01:30:00Z", 2025, 7, 1},
		{"UTCMinus12ToUTC", "UTC-12", "2020-01-01T11:45:00Z", "2019-12-31T23:45:00Z", `{"TakenAt":"2019-12-31T23:45:00Z","TimeZone":"UTC","TakenSrc":"manual"}`, "2019-12-31T23:45:00Z", "2019-12-31T23:45:00Z", 2019, 12, 31},
		{"UTCToPreviousYear", "UTC", "2020-01-01T00:30:00Z", "2020-01-01T00:30:00Z", `{"TakenAtLocal":"2019-12-31T12:30:00Z","TimeZone":"UTC-12","Year":2019,"Month":12,"Day":31,"TakenSrc":"manual"}`, "2020-01-01T00:30:00Z", "2019-12-31T12:30:00Z", 2019, 12, 31},
		{"BerlinToUTC", "Europe/Berlin", "2024-02-29T23:30:00Z", "2024-03-01T00:30:00Z", `{"TakenAt":"2024-03-01T00:30:00Z","TimeZone":"UTC","TakenSrc":"manual"}`, "2024-03-01T00:30:00Z", "2024-03-01T00:30:00Z", 2024, 3, 1},
		{"BerlinToNewYork", "Europe/Berlin", "2025-06-30T12:00:00Z", "2025-06-30T14:00:00Z", `{"TakenAt":"2025-06-30T18:00:00Z","TimeZone":"America/New_York","TakenSrc":"manual"}`, "2025-06-30T18:00:00Z", "2025-06-30T14:00:00Z", 2025, 6, 30},
		{"RepeatedHour", "UTC", "2025-10-26T00:30:00Z", "2025-10-26T00:30:00Z", `{"TakenAtLocal":"2025-10-26T02:30:00Z","TimeZone":"Europe/Berlin","TakenSrc":"manual"}`, "2025-10-26T00:30:00Z", "2025-10-26T02:30:00Z", 2025, 10, 26},
		{"ManualDateWithZone", "UTC", "2025-06-30T23:30:00Z", "2025-06-30T23:30:00Z", `{"TimeZone":"Europe/Berlin","TakenAt":"2025-07-02T01:45:00Z","TakenAtLocal":"2025-07-02T03:45:00Z","Year":2025,"Month":7,"Day":2}`, "2025-07-02T01:45:00Z", "2025-07-02T03:45:00Z", 2025, 7, 2},
		{"ZoneOnlyKeepsWallTime", "Europe/Berlin", "2025-06-30T12:00:00Z", "2025-06-30T14:00:00Z", `{"TimeZone":"America/New_York"}`, "2025-06-30T18:00:00Z", "2025-06-30T14:00:00Z", 2025, 6, 30},
		{"ZoneOnlyToUTCKeepsWallTime", "Europe/Berlin", "2025-06-30T12:00:00Z", "2025-06-30T14:00:00Z", `{"TimeZone":"UTC"}`, "2025-06-30T14:00:00Z", "2025-06-30T14:00:00Z", 2025, 6, 30},
		{"ZoneOnlyFromUTCConverts", "UTC", "2025-06-30T23:30:00Z", "2025-06-30T23:30:00Z", `{"TimeZone":"Europe/Berlin","TakenSrc":"manual"}`, "2025-06-30T23:30:00Z", "2025-07-01T01:30:00Z", 2025, 7, 1},
		{"ZoneOnlyFromUTCUnknownDay", "UTC", "2020-01-01T03:00:00Z", "2020-01-01T03:00:00Z", `{"TimeZone":"America/New_York"}`, "2020-01-01T03:00:00Z", "2019-12-31T22:00:00Z", 2020, 1, -1},
		{"AssignLocalZone", "Local", "2025-06-30T23:30:00Z", "2025-06-30T23:30:00Z", `{"TakenAt":"2025-06-30T21:30:00Z","TimeZone":"Europe/Berlin"}`, "2025-06-30T21:30:00Z", "2025-06-30T23:30:00Z", 2025, 6, 30},
		{"RemoveZone", "Europe/Berlin", "2025-06-30T21:30:00Z", "2025-06-30T23:30:00Z", `{"TakenAt":"2025-06-30T23:30:00Z","TimeZone":"Local"}`, "2025-06-30T23:30:00Z", "2025-06-30T23:30:00Z", 2025, 6, 30},
		{"UnknownDay", "UTC", "2020-01-01T03:00:00Z", "2020-01-01T03:00:00Z", `{"TakenAtLocal":"2019-12-31T22:00:00Z","TimeZone":"America/New_York"}`, "2020-01-01T03:00:00Z", "2019-12-31T22:00:00Z", 2020, 1, -1},
		{"UnknownDate", "UTC", "2025-06-30T23:30:00Z", "2025-06-30T23:30:00Z", `{"TakenAtLocal":"2025-07-01T01:30:00Z","TimeZone":"Europe/Berlin"}`, "2025-06-30T23:30:00Z", "2025-07-01T01:30:00Z", -1, -1, -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app, router, _ := NewApiTest()
			UpdatePhoto(router)
			local, _ := time.Parse(time.RFC3339, tc.local)
			year, month, day := local.Year(), int(local.Month()), local.Day()
			if tc.year <= 0 {
				year = tc.year
			}
			if tc.month <= 0 {
				month = tc.month
			}
			if tc.day <= 0 {
				day = tc.day
			}
			photo := createTimeZoneTestPhoto(t, tc.zone, tc.taken, tc.local, year, month, day)
			uri := "/api/v1/photos/" + photo.PhotoUID
			for _, body := range []string{tc.body, `{"Caption":"Updated caption"}`, `{"Title":"Updated title","TitleSrc":"manual"}`} {
				r := PerformRequestWithBody(app, "PUT", uri, body)
				require.Equal(t, http.StatusOK, r.Code, r.Body.String())
				var result entity.Photo
				require.NoError(t, json.Unmarshal(r.Body.Bytes(), &result))
				assert.Equal(t, tc.expectedTaken, result.TakenAt.Format(time.RFC3339), body)
				assert.Equal(t, tc.expectedLocal, result.TakenAtLocal.Format(time.RFC3339), body)
				assert.Equal(t, []int{tc.year, tc.month, tc.day}, []int{result.PhotoYear, result.PhotoMonth, result.PhotoDay}, body)
				var stored entity.Photo
				require.NoError(t, entity.Db().First(&stored, photo.ID).Error)
				assert.True(t, stored.TakenAt.Equal(result.TakenAt))
				assert.True(t, stored.TakenAtLocal.Equal(result.TakenAtLocal))
			}
		})
	}
}
