package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// TestPhotoLabelPartialAcceptance checks partial input and subsequent removal semantics.
func TestPhotoLabelPartialAcceptance(t *testing.T) {
	app, router, conf := NewApiTest()
	UpdatePhotoLabel(router)
	RemovePhotoLabel(router)
	mode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePublic)
	t.Cleanup(func() { conf.SetAuthMode(mode) })
	for _, tc := range []struct {
		name, body     string
		rename, accept bool
	}{
		{"Empty", `{}`, false, false},
		{"Null", `null`, false, false},
		{"NullUncertainty", `{"Uncertainty":null}`, false, false},
		{"NullLabel", `{"Label":null}`, false, false},
		{"EmptyLabel", `{"Label":{}}`, false, false},
		{"NullName", `{"Label":{"Name":null}}`, false, false},
		{"Rename", `{"Label":{"Name":"Partial Name Control"}}`, true, false},
		{"RenameNullUncertainty", `{"Uncertainty":null,"Label":{"Name":"Partial Name Control"}}`, true, false},
		{"Accept", `{"Uncertainty":0}`, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			photo, label := photoLabelBoundFixture(t)
			require.NoError(t, entity.Db().Model(&entity.PhotoLabel{}).Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).UpdateColumn("uncertainty", 0).Error)
			url := fmt.Sprintf("/api/v1/photos/%s/label/%d", photo.PhotoUID, label.ID)
			r := PerformRequestWithBody(app, http.MethodPut, url, tc.body)
			require.Equal(t, http.StatusOK, r.Code, r.Body.String())
			entity.WaitForAsyncJobs()
			var link entity.PhotoLabel
			require.NoError(t, entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&link).Error)
			assert.Zero(t, link.Uncertainty)
			if tc.accept {
				assert.Equal(t, entity.SrcManual, link.LabelSrc)
			} else {
				assert.Equal(t, entity.SrcImage, link.LabelSrc)
			}
			var stored entity.Label
			require.NoError(t, entity.Db().First(&stored, label.ID).Error)
			if tc.rename {
				assert.Equal(t, "Partial Name Control", stored.LabelName)
				assert.Equal(t, "partial-name-control", stored.CustomSlug)
			} else {
				assert.Equal(t, label, stored)
			}
			r = PerformRequest(app, http.MethodDelete, url)
			require.Equal(t, http.StatusOK, r.Code, r.Body.String())
			entity.WaitForAsyncJobs()
			err := entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&link).Error
			if tc.accept {
				assert.True(t, gorm.IsRecordNotFoundError(err))
			} else {
				require.NoError(t, err)
				assert.Equal(t, 100, link.Uncertainty)
				assert.Equal(t, entity.SrcManual, link.LabelSrc)
			}
		})
	}
}

// TestPhotoLabelUncertaintyBounds checks numeric bounds before any name or assignment write.
func TestPhotoLabelUncertaintyBounds(t *testing.T) {
	app, router, conf := NewApiTest()
	UpdatePhotoLabel(router)
	mode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePublic)
	t.Cleanup(func() { conf.SetAuthMode(mode) })
	for _, tc := range []struct {
		name, body string
	}{
		{"Negative", `{"Uncertainty":-1}`},
		{"LargeNegative", `{"Uncertainty":-50,"Label":{"Name":"Bounded Name Control"}}`},
		{"AboveMaximum", `{"Uncertainty":101,"Label":{"Name":"Bounded Name Control"}}`},
		{"LargePositive", `{"Uncertainty":100000,"Label":{"Name":"Bounded Name Control"}}`},
		{"NullLabel", `{"Uncertainty":101,"Label":null}`},
		{"NullName", `{"Uncertainty":-1,"Label":{"Name":null}}`},
		{"String", `{"Uncertainty":"0","Label":{"Name":"Bounded Name Control"}}`},
		{"Fraction", `{"Uncertainty":0.5,"Label":{"Name":"Bounded Name Control"}}`},
		{"Boolean", `{"Uncertainty":false,"Label":{"Name":"Bounded Name Control"}}`},
		{"Overflow", `{"Uncertainty":9223372036854775808,"Label":{"Name":"Bounded Name Control"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			photo, label := photoLabelBoundFixture(t)
			var before, after entity.PhotoLabel
			require.NoError(t, entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&before).Error)
			writes := 0
			entity.Db().Callback().Update().Before("gorm:begin_transaction").Register("test:input-writes", func(scope *gorm.Scope) {
				switch scope.TableName() {
				case (entity.Label{}).TableName(), (entity.PhotoLabel{}).TableName(), (entity.Photo{}).TableName():
					writes++
				}
			})
			t.Cleanup(func() { entity.Db().Callback().Update().Remove("test:input-writes") })
			r := PerformRequestWithBody(app, http.MethodPut, fmt.Sprintf("/api/v1/photos/%s/label/%d", photo.PhotoUID, label.ID), tc.body)
			entity.WaitForAsyncJobs()
			assert.Equal(t, http.StatusBadRequest, r.Code, r.Body.String())
			assert.Zero(t, writes)
			require.NoError(t, entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&after).Error)
			assert.Equal(t, before, after)
			var stored entity.Label
			require.NoError(t, entity.Db().First(&stored, label.ID).Error)
			assert.Equal(t, label, stored)
		})
	}
}

// TestPhotoLabelExplicitUncertainty checks inclusive bounds and source transitions.
func TestPhotoLabelExplicitUncertainty(t *testing.T) {
	app, router, conf := NewApiTest()
	UpdatePhotoLabel(router)
	mode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePublic)
	t.Cleanup(func() { conf.SetAuthMode(mode) })
	for _, source := range []struct{ name, value string }{
		{"Image", entity.SrcImage}, {"Manual", entity.SrcManual}, {"Batch", entity.SrcBatch},
	} {
		for _, uncertainty := range []int{0, 1, 100} {
			t.Run(fmt.Sprintf("%s%d", source.name, uncertainty), func(t *testing.T) {
				photo, label := photoLabelBoundFixture(t)
				require.NoError(t, entity.Db().Model(&entity.PhotoLabel{}).Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).UpdateColumn("label_src", source.value).Error)
				r := PerformRequestWithBody(app, http.MethodPut, fmt.Sprintf("/api/v1/photos/%s/label/%d", photo.PhotoUID, label.ID), fmt.Sprintf(`{"Uncertainty":%d}`, uncertainty))
				require.Equal(t, http.StatusOK, r.Code, r.Body.String())
				entity.WaitForAsyncJobs()
				var link entity.PhotoLabel
				require.NoError(t, entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&link).Error)
				assert.Equal(t, uncertainty, link.Uncertainty)
				expected := source.value
				if uncertainty == 0 {
					expected = entity.SrcManual
				}
				assert.Equal(t, expected, link.LabelSrc)
			})
		}
	}
}
