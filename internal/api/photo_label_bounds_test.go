package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// photoLabelErrorLog captures handler error messages while preserving the existing log hooks.
func photoLabelErrorLog(t *testing.T) *logtest.Hook {
	t.Helper()
	logger, ok := log.(*logrus.Logger)
	require.True(t, ok)
	saved := make(logrus.LevelHooks, len(logger.Hooks))
	for level, hooks := range logger.Hooks {
		saved[level] = append([]logrus.Hook(nil), hooks...)
	}
	hook := logtest.NewLocal(logger)
	t.Cleanup(func() { logger.ReplaceHooks(saved) })
	return hook
}

// assertPhotoLabelError checks that a failed write is logged once at error level.
func assertPhotoLabelError(t *testing.T, hook *logtest.Hook, message string) {
	t.Helper()
	count := 0
	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.ErrorLevel && strings.Contains(entry.Message, message) {
			count++
		}
	}
	assert.Equal(t, 1, count)
}

// photoLabelBoundFixture stores a picture and one image-sourced label assignment.
func photoLabelBoundFixture(t *testing.T) (entity.Photo, entity.Label) {
	t.Helper()
	entity.WaitForAsyncJobs()
	photo := entity.NewPhoto(false)
	require.NoError(t, photo.Save())
	label := entity.NewLabel("Label Control "+rnd.GenerateUID('l'), 3)
	label.LabelDescription = "Original description"
	label.LabelNotes = "Original notes"
	require.NoError(t, label.Create())
	link := entity.NewPhotoLabel(photo.ID, label.ID, 40, entity.SrcImage)
	require.NoError(t, link.Create())
	t.Cleanup(func() {
		entity.WaitForAsyncJobs()
		entity.UnscopedDb().Delete(&entity.PhotoLabel{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Delete(&photo)
		entity.UnscopedDb().Delete(&entity.Label{}, "id = ?", label.ID)
	})
	var stored entity.Label
	require.NoError(t, entity.Db().First(&stored, label.ID).Error)
	return photo, stored
}

// TestPhotoLabelUpdateBounds checks the editable fields and existing assignment operations.
func TestPhotoLabelUpdateBounds(t *testing.T) {
	app, router, conf := NewApiTest()
	UpdatePhotoLabel(router)
	RemovePhotoLabel(router)
	mode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePublic)
	t.Cleanup(func() { conf.SetAuthMode(mode) })
	t.Run("SharedFieldsStay", func(t *testing.T) {
		photo, before := photoLabelBoundFixture(t)
		data, err := json.Marshal(map[string]any{"Uncertainty": 30, "Label": map[string]any{
			"UID": rnd.GenerateUID('l'), "Slug": "submitted-" + rnd.GenerateUID('l'),
			"Description": "Submitted description", "Notes": "Submitted notes", "Priority": 7, "Favorite": true,
			"Thumb": "submitted-thumb", "ThumbSrc": entity.SrcManual,
		}})
		require.NoError(t, err)
		r := PerformRequestWithBody(app, http.MethodPut, fmt.Sprintf("/api/v1/photos/%s/label/%d", photo.PhotoUID, before.ID), string(data))
		require.Equal(t, http.StatusOK, r.Code)
		entity.WaitForAsyncJobs()
		var after entity.Label
		require.NoError(t, entity.UnscopedDb().First(&after, before.ID).Error)
		assert.Equal(t, before.LabelUID, after.LabelUID)
		assert.Equal(t, before.LabelSlug, after.LabelSlug)
		assert.Equal(t, before.CustomSlug, after.CustomSlug)
		assert.Equal(t, before.LabelName, after.LabelName)
		assert.Equal(t, before.LabelDescription, after.LabelDescription)
		assert.Equal(t, before.LabelNotes, after.LabelNotes)
		assert.Equal(t, before.LabelPriority, after.LabelPriority)
		assert.Equal(t, before.LabelFavorite, after.LabelFavorite)
		assert.Equal(t, before.Thumb, after.Thumb)
		assert.Equal(t, before.ThumbSrc, after.ThumbSrc)
	})
	t.Run("Rename", func(t *testing.T) {
		photo, before := photoLabelBoundFixture(t)
		r := PerformRequestWithBody(app, http.MethodPut, fmt.Sprintf("/api/v1/photos/%s/label/%d", photo.PhotoUID, before.ID), `{"Label":{"Name":"Label Rename Control"}}`)
		require.Equal(t, http.StatusOK, r.Code)
		entity.WaitForAsyncJobs()
		var after entity.Label
		require.NoError(t, entity.UnscopedDb().First(&after, before.ID).Error)
		assert.Equal(t, "Label Rename Control", after.LabelName)
		assert.Equal(t, "label-rename-control", after.CustomSlug)
		assert.Equal(t, before.LabelSlug, after.LabelSlug)
		assert.Equal(t, before.LabelUID, after.LabelUID)
		var link entity.PhotoLabel
		require.NoError(t, entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, before.ID).First(&link).Error)
		assert.Equal(t, 40, link.Uncertainty)
		assert.Equal(t, entity.SrcImage, link.LabelSrc)
		assert.Equal(t, "Label Rename Control", gjson.Get(r.Body.String(), fmt.Sprintf("Labels.#(LabelID==%d).Label.Name", before.ID)).String())
	})
	t.Run("Accept", func(t *testing.T) {
		photo, label := photoLabelBoundFixture(t)
		require.NoError(t, entity.Db().Model(&entity.PhotoLabel{}).Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).UpdateColumn("uncertainty", 100).Error)
		r := PerformRequestWithBody(app, http.MethodPut, fmt.Sprintf("/api/v1/photos/%s/label/%d", photo.PhotoUID, label.ID), `{"Uncertainty":0}`)
		require.Equal(t, http.StatusOK, r.Code)
		entity.WaitForAsyncJobs()
		var link entity.PhotoLabel
		require.NoError(t, entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&link).Error)
		assert.Zero(t, link.Uncertainty)
		assert.Equal(t, entity.SrcManual, link.LabelSrc)
	})
	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		t.Run(method+"WriteFailure", func(t *testing.T) {
			photo, label := photoLabelBoundFixture(t)
			hook := photoLabelErrorLog(t)
			fired := false
			entity.Db().Callback().Update().Before("gorm:begin_transaction").Register("test:label-write-error", func(db *gorm.DB) {
				if db.Statement.Table == (entity.PhotoLabel{}).TableName() && !fired {
					fired = true
					_ = db.Statement.AddError(errors.New("photo label write control"))
				}
			})
			t.Cleanup(func() { entity.Db().Callback().Update().Remove("test:label-write-error") })
			r := PerformRequestWithBody(app, method, fmt.Sprintf("/api/v1/photos/%s/label/%d", photo.PhotoUID, label.ID), `{"Uncertainty":20}`)
			entity.WaitForAsyncJobs()
			require.True(t, fired)
			assert.Equal(t, http.StatusInternalServerError, r.Code)
			assert.NotContains(t, strings.ToLower(r.Body.String()), "photo label write control")
			assertPhotoLabelError(t, hook, "photo label write control")
		})
	}
}

// TestUpdatePhotoLabelScope checks assignment and name edits with resource-scoped credentials.
func TestUpdatePhotoLabelScope(t *testing.T) {
	app, router, conf := NewApiTest()
	UpdatePhotoLabel(router)
	mode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(mode) })
	for _, grant := range []authn.GrantType{authn.GrantClientCredentials, authn.GrantPassword} {
		t.Run(grant.String(), func(t *testing.T) {
			for _, tc := range []struct {
				name, scope string
				rename      bool
				status      int
			}{
				{"Assignment", "photos", false, http.StatusOK},
				{"IgnoredLabelFields", "photos", false, http.StatusOK},
				{"NameRequiresLabels", "photos", true, http.StatusForbidden},
				{"NamePermitted", "photos labels", true, http.StatusOK},
				{"WriteNamePermitted", "write photos labels", true, http.StatusOK},
				{"PhotoRequired", "labels", true, http.StatusForbidden},
				{"ReadOnly", "read photos labels", true, http.StatusForbidden},
			} {
				t.Run(tc.name, func(t *testing.T) {
					photo, before := photoLabelBoundFixture(t)
					var user *entity.User
					if grant == authn.GrantPassword {
						user = entity.UserFixtures.Pointer("alice")
					}
					sess, err := entity.AddClientSession("label-scope-control", conf.SessionMaxAge(), tc.scope, grant, user)
					require.NoError(t, err)
					t.Cleanup(func() { _ = sess.Delete() })
					body := `{"Uncertainty":0}`
					if tc.name == "IgnoredLabelFields" {
						body = `{"Uncertainty":0,"Label":{"UID":"ltm3000000000001","Priority":99}}`
					}
					if tc.rename {
						body = `{"Uncertainty":0,"Label":{"Name":"Scoped Name Control"}}`
					}
					r := AuthenticatedRequestWithBody(app, http.MethodPut, fmt.Sprintf("/api/v1/photos/%s/label/%d", photo.PhotoUID, before.ID), body, sess.AuthToken())
					require.Equal(t, tc.status, r.Code, "%s", r.Body.String())
					entity.WaitForAsyncJobs()
					var after entity.Label
					require.NoError(t, entity.UnscopedDb().First(&after, before.ID).Error)
					assert.Equal(t, before.LabelUID, after.LabelUID)
					assert.Equal(t, before.LabelPriority, after.LabelPriority)
					var link entity.PhotoLabel
					require.NoError(t, entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, before.ID).First(&link).Error)
					if tc.status == http.StatusOK {
						assert.Zero(t, link.Uncertainty)
						assert.Equal(t, entity.SrcManual, link.LabelSrc)
						if tc.rename {
							assert.Equal(t, "Scoped Name Control", after.LabelName)
						} else {
							assert.Equal(t, before.LabelName, after.LabelName)
						}
					} else {
						assert.Equal(t, 40, link.Uncertainty)
						assert.Equal(t, entity.SrcImage, link.LabelSrc)
						assert.Equal(t, before.LabelName, after.LabelName)
					}
				})
			}
		})
	}
}

// TestUpdatePhotoLabelRouteIdentity checks that submitted identifiers do not select a different assignment.
func TestUpdatePhotoLabelRouteIdentity(t *testing.T) {
	app, router, _ := NewApiTest()
	UpdatePhotoLabel(router)
	photo, label := photoLabelBoundFixture(t)
	otherPhoto, otherLabel := photoLabelBoundFixture(t)
	body, err := json.Marshal(map[string]any{"PhotoID": otherPhoto.ID, "LabelID": otherLabel.ID, "LabelSrc": entity.SrcYaml, "Topicality": 99, "NSFW": 90, "Uncertainty": 25, "Label": map[string]any{"ID": otherLabel.ID, "UID": otherLabel.LabelUID}})
	require.NoError(t, err)
	r := PerformRequestWithBody(app, http.MethodPut, fmt.Sprintf("/api/v1/photos/%s/label/%d", photo.PhotoUID, label.ID), string(body))
	require.Equal(t, http.StatusOK, r.Code)
	entity.WaitForAsyncJobs()
	var link, other entity.PhotoLabel
	require.NoError(t, entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&link).Error)
	require.NoError(t, entity.Db().Where("photo_id = ? AND label_id = ?", otherPhoto.ID, otherLabel.ID).First(&other).Error)
	assert.Equal(t, 25, link.Uncertainty)
	assert.Equal(t, entity.SrcImage, link.LabelSrc)
	assert.Zero(t, link.Topicality)
	assert.Zero(t, link.NSFW)
	assert.Equal(t, 40, other.Uncertainty)
	var stored entity.Label
	require.NoError(t, entity.Db().First(&stored, otherLabel.ID).Error)
	assert.Equal(t, otherLabel.LabelUID, stored.LabelUID)
	assert.Equal(t, otherLabel.LabelName, stored.LabelName)
}

// TestUpdatePhotoLabelNameFields checks that a permitted rename only changes the naming columns.
func TestUpdatePhotoLabelNameFields(t *testing.T) {
	app, router, _ := NewApiTest()
	UpdatePhotoLabel(router)
	photo, before := photoLabelBoundFixture(t)
	body, err := json.Marshal(map[string]any{"Label": map[string]any{
		"Name": "Field Name Control", "UID": rnd.GenerateUID('l'), "Slug": "submitted-slug", "CustomSlug": "submitted-custom",
		"Description": "Submitted description", "Notes": "Submitted notes", "Priority": 99, "Favorite": true, "NSFW": true,
		"PhotoCount": 321, "Thumb": "submitted-thumb", "ThumbSrc": entity.SrcManual,
		"CreatedAt": "2000-01-01T00:00:00Z", "UpdatedAt": "2000-01-01T00:00:00Z", "DeletedAt": "2000-01-01T00:00:00Z", "PublishedAt": "2000-01-01T00:00:00Z",
	}})
	require.NoError(t, err)
	r := PerformRequestWithBody(app, http.MethodPut, fmt.Sprintf("/api/v1/photos/%s/label/%d", photo.PhotoUID, before.ID), string(body))
	require.Equal(t, http.StatusOK, r.Code)
	entity.WaitForAsyncJobs()
	var after entity.Label
	require.NoError(t, entity.UnscopedDb().First(&after, before.ID).Error)
	assert.Equal(t, "Field Name Control", after.LabelName)
	assert.Equal(t, "field-name-control", after.CustomSlug)
	before.LabelName, before.CustomSlug = after.LabelName, after.CustomSlug
	assert.Equal(t, before, after)
}

// TestUpdatePhotoLabelInvalidName checks that validation precedes assignment updates.
func TestUpdatePhotoLabelInvalidName(t *testing.T) {
	app, router, _ := NewApiTest()
	UpdatePhotoLabel(router)
	photo, label := photoLabelBoundFixture(t)
	r := PerformRequestWithBody(app, http.MethodPut, fmt.Sprintf("/api/v1/photos/%s/label/%d", photo.PhotoUID, label.ID), `{"Uncertainty":0,"Label":{"Name":""}}`)
	require.Equal(t, http.StatusBadRequest, r.Code)
	var link entity.PhotoLabel
	require.NoError(t, entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&link).Error)
	assert.Equal(t, 40, link.Uncertainty)
	assert.Equal(t, entity.SrcImage, link.LabelSrc)
}

// TestUpdatePhotoLabelPartialInput preserves omitted values for a photos-scoped credential.
func TestUpdatePhotoLabelPartialInput(t *testing.T) {
	app, router, conf := NewApiTest()
	UpdatePhotoLabel(router)
	mode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(mode) })
	sess, err := entity.AddClientSession("partial-label-control", conf.SessionMaxAge(), "photos", authn.GrantClientCredentials, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sess.Delete() })
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"Empty", `{}`, http.StatusOK}, {"Null", `null`, http.StatusOK},
		{"NullLabel", `{"Label":null}`, http.StatusOK}, {"EmptyLabel", `{"Label":{}}`, http.StatusOK},
		{"NullName", `{"Label":{"Name":null}}`, http.StatusOK},
		{"WrongNameType", `{"Uncertainty":0,"Label":{"Name":123}}`, http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			photo, label := photoLabelBoundFixture(t)
			r := AuthenticatedRequestWithBody(app, http.MethodPut, fmt.Sprintf("/api/v1/photos/%s/label/%d", photo.PhotoUID, label.ID), tc.body, sess.AuthToken())
			require.Equal(t, tc.status, r.Code)
			entity.WaitForAsyncJobs()
			var link entity.PhotoLabel
			require.NoError(t, entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&link).Error)
			assert.Equal(t, 40, link.Uncertainty)
			assert.Equal(t, entity.SrcImage, link.LabelSrc)
			var stored entity.Label
			require.NoError(t, entity.Db().First(&stored, label.ID).Error)
			assert.Equal(t, label, stored)
		})
	}
}

// TestPhotoLabelOperationErrors checks generic responses and unchanged state after rename and delete failures.
func TestPhotoLabelOperationErrors(t *testing.T) {
	for _, operation := range []string{"Rename", "DeleteManual", "DeleteBatch"} {
		t.Run(operation, func(t *testing.T) {
			app, router, _ := NewApiTest()
			UpdatePhotoLabel(router)
			RemovePhotoLabel(router)
			photo, label := photoLabelBoundFixture(t)
			method, body := http.MethodPut, `{"Uncertainty":0,"Label":{"Name":"Failed Name Control"}}`
			if operation != "Rename" {
				method = http.MethodDelete
				source := entity.SrcManual
				if operation == "DeleteBatch" {
					source = entity.SrcBatch
				}
				require.NoError(t, entity.Db().Model(&entity.PhotoLabel{}).Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).UpdateColumn("label_src", source).Error)
			}
			hook := photoLabelErrorLog(t)
			fired := false
			callback := func(db *gorm.DB) {
				if operation == "Rename" && db.Statement.Table == (entity.Label{}).TableName() || operation != "Rename" && db.Statement.Table == (entity.PhotoLabel{}).TableName() {
					fired = true
					_ = db.Statement.AddError(errors.New("operation write control"))
				}
			}
			if operation == "Rename" {
				entity.Db().Callback().Update().Before("gorm:begin_transaction").Register("test:label-operation-failure", callback)
				t.Cleanup(func() { entity.Db().Callback().Update().Remove("test:label-operation-failure") })
			} else {
				entity.Db().Callback().Delete().Before("gorm:begin_transaction").Register("test:label-operation-failure", callback)
				t.Cleanup(func() { entity.Db().Callback().Delete().Remove("test:label-operation-failure") })
			}
			r := PerformRequestWithBody(app, method, fmt.Sprintf("/api/v1/photos/%s/label/%d", photo.PhotoUID, label.ID), body)
			require.True(t, fired)
			assert.Equal(t, http.StatusInternalServerError, r.Code)
			assert.NotContains(t, strings.ToLower(r.Body.String()), "operation write control")
			assertPhotoLabelError(t, hook, "operation write control")
			var stored entity.Label
			require.NoError(t, entity.Db().First(&stored, label.ID).Error)
			assert.Equal(t, label, stored)
			var link entity.PhotoLabel
			require.NoError(t, entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&link).Error)
			assert.Equal(t, 40, link.Uncertainty)
		})
	}
}

// TestUpdatePhotoLabelClientRole checks the client role ceiling independently of credential scope.
func TestUpdatePhotoLabelClientRole(t *testing.T) {
	app, router, conf := NewApiTest()
	UpdatePhotoLabel(router)
	mode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(mode) })
	token := mixedPrincipalToken(t, acl.RolePortal, "alice")
	require.True(t, acl.Rules.Allow(acl.ResourcePhotos, acl.RolePortal, acl.ActionUpdate))
	require.False(t, acl.Rules.Allow(acl.ResourceLabels, acl.RolePortal, acl.ActionUpdate))
	photo, label := photoLabelBoundFixture(t)
	path := fmt.Sprintf("/api/v1/photos/%s/label/%d", photo.PhotoUID, label.ID)
	r := AuthenticatedRequestWithBody(app, http.MethodPut, path, `{"Uncertainty":20}`, token)
	require.Equal(t, http.StatusOK, r.Code)
	entity.WaitForAsyncJobs()
	r = AuthenticatedRequestWithBody(app, http.MethodPut, path, `{"Uncertainty":0,"Label":{"Name":"Client Role Name"}}`, token)
	require.Equal(t, http.StatusForbidden, r.Code)
	var stored entity.Label
	require.NoError(t, entity.Db().First(&stored, label.ID).Error)
	assert.Equal(t, label.LabelName, stored.LabelName)
	var link entity.PhotoLabel
	require.NoError(t, entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&link).Error)
	assert.Equal(t, 20, link.Uncertainty)
	assert.Equal(t, entity.SrcImage, link.LabelSrc)
}
