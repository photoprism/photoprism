package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetDbFieldNameMap_ExistingFields(t *testing.T) {
	type MyEntity struct {
		ID               uint         `gorm:"primary_key" yaml:"-"`
		UUID             string       `gorm:"type:VARBINARY(64);index;" json:"DocumentID,omitempty" yaml:"DocumentID,omitempty"`
		LensID           uint         `gorm:"index:idx_photos_camera_lens;default:1" json:"LensID" yaml:"-"`
		Details          *Details     `gorm:"association_autoupdate:false;association_autocreate:false;association_save_reference:false" json:"Details" yaml:"Details"`
		Albums           []Album      `json:"Albums" yaml:"-"`
		Files            []File       `yaml:"-"`
		Labels           []PhotoLabel `yaml:"-"`
		PhotoFNumber     float32      `gorm:"type:FLOAT;" json:"FNumber" yaml:"FNumber,omitempty"`
		PhotoFocalLength int          `json:"FocalLength" yaml:"FocalLength,omitempty"`
		SomeDbTime       time.Time    `json:"SomeTime" yaml:"SomeTime,omitempty"`
		SomeGormField    time.Time    `gorm:"column:a_gorm_field;" json:"TheGormField" yaml:"SomeTime,omitempty"`
		TypedGormField   string       `gorm:"column:a_typed_gorm_field;type:VARBINARY(64);" json:"TheTypedGormField" yaml:"SomeTime,omitempty"`
		TypedGormField2  string       `gorm:"type:VARBINARY(64);column:second_typed_gorm_field;" json:"TheSecondTypedGormField" yaml:"SomeTime,omitempty"`
		TypedGormField3  string       `gorm:"type:VARBINARY(64);column:third_typed_gorm_field;index" json:"TheThirdypedGormField" yaml:"SomeTime,omitempty"`
	}
	t.Run("get db fieldname map for typical entity struct", func(t *testing.T) {
		m := GetDbFieldMap(MyEntity{})
		assert.Equal(t, "uuid", m["DocumentID"].FieldName)
		assert.Equal(t, "lens_id", m["LensID"].FieldName)
		assert.Equal(t, "details", m["Details"].FieldName)
		assert.Equal(t, "albums", m["Albums"].FieldName)
		assert.Equal(t, "photo_f_number", m["FNumber"].FieldName)
		assert.Equal(t, "photo_focal_length", m["FocalLength"].FieldName)
		assert.Equal(t, "some_db_time", m["SomeTime"].FieldName)
		assert.Equal(t, "a_gorm_field", m["TheGormField"].FieldName)
		assert.Equal(t, "a_typed_gorm_field", m["TheTypedGormField"].FieldName)
		assert.Equal(t, "second_typed_gorm_field", m["TheSecondTypedGormField"].FieldName)
		assert.Equal(t, "third_typed_gorm_field", m["TheThirdypedGormField"].FieldName)
	})
}

func TestGetDbFieldNameMap_FieldsWithoutJsonTag(t *testing.T) {
	type MyEntity struct {
		Files            []File       `yaml:"-"`
		Labels           []PhotoLabel `yaml:"-"`
		PhotoFNumber     float32      `gorm:"type:FLOAT;" yaml:"FNumber,omitempty"`
		PhotoFocalLength int          `yaml:"FocalLength,omitempty"`
	}
	t.Run("get db fieldname map for entity struct without JSON tags", func(t *testing.T) {
		m := GetDbFieldMap(MyEntity{})
		assert.Zero(t, len(m))
	})
}

func TestSubstDbFields_ExistingFields(t *testing.T) {
	type MyEntity struct {
		ID               uint      `gorm:"primary_key" yaml:"-"`
		UUID             string    `gorm:"type:VARBINARY(64);index;" json:"DocumentID,omitempty" yaml:"DocumentID,omitempty"`
		LensID           uint      `gorm:"index:idx_photos_camera_lens;default:1" json:"LensID" yaml:"-"`
		Details          *Details  `gorm:"association_autoupdate:false;association_autocreate:false;association_save_reference:false" json:"Details" yaml:"Details"`
		Albums           []Album   `json:"Albums" yaml:"-"`
		TakenAtLocal     time.Time `gorm:"type:DATETIME;" json:"TakenAtLocal" yaml:"TakenAtLocal"`
		PhotoFNumber     float32   `gorm:"type:FLOAT;" json:"FNumber" yaml:"FNumber,omitempty"`
		PhotoFocalLength int       `json:"FocalLength" yaml:"FocalLength,omitempty"`
		SomeBoolColumn   bool      `json:"TheBoolean" yaml:"YesTheBoolean,omitempty"`
	}
	changesMap := map[string]interface{}{"DocumentID": "98upqwe89", "TakenAtLocal": "2023-02-28T14:15:16Z", "LensID": 1, "FNumber": 1.234, "TheBoolean": true}
	t.Run("substitute db field names and values", func(t *testing.T) {
		m := GetDbFieldMap(MyEntity{})
		s := SubstDbFields(changesMap, m)
		assert.Equal(t, "98upqwe89", s["uuid"])
		assert.Equal(t, time.Date(2023, time.Month(02), 28, 14, 15, 16, 0, time.UTC), s["taken_at_local"])
		assert.Equal(t, 1, s["lens_id"])
		assert.Equal(t, 1.234, s["photo_f_number"])
		assert.Equal(t, true, s["some_bool_column"])
	})
}

func TestSubstDbFields_UnknownFields(t *testing.T) {
	type MyEntity struct {
		ID               uint      `gorm:"primary_key" yaml:"-"`
		UUID             string    `gorm:"type:VARBINARY(64);index;" json:"DocumentID,omitempty" yaml:"DocumentID,omitempty"`
		LensID           uint      `gorm:"index:idx_photos_camera_lens;default:1" json:"LensID" yaml:"-"`
		Details          *Details  `gorm:"association_autoupdate:false;association_autocreate:false;association_save_reference:false" json:"Details" yaml:"Details"`
		Albums           []Album   `json:"Albums" yaml:"-"`
		TakenAtLocal     time.Time `gorm:"type:DATETIME;" json:"TakenAtLocal" yaml:"TakenAtLocal"`
		PhotoFNumber     float32   `gorm:"type:FLOAT;" json:"FNumber" yaml:"FNumber,omitempty"`
		PhotoFocalLength int       `json:"FocalLength" yaml:"FocalLength,omitempty"`
		SomeBoolColumn   bool      `json:"TheBoolean" yaml:"YesTheBoolean,omitempty"`
	}
	changesMap := map[string]interface{}{"Unknown1": "98upqwe89", "Unknown2": "2023-02-28T14:15:16Z", "Unknown3": 1, "Unknown4": 1.234, "Unknown5": true}
	t.Run("substitute db field names and values - no fields known", func(t *testing.T) {
		m := GetDbFieldMap(MyEntity{})
		s := SubstDbFields(changesMap, m)
		assert.NotZero(t, len(m))
		assert.Zero(t, len(s))
	})
}
