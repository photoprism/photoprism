package entity

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/pquerna/otp"
	"github.com/stretchr/testify/assert"
	"github.com/ulule/deepcopier"
)

func TestModelValues(t *testing.T) {
	t.Run("NoInterface", func(t *testing.T) {
		m := Photo{}
		values, keys, err := ModelValues(m, "ID", "PhotoUID")

		assert.Error(t, err)
		assert.IsType(t, map[string]interface{}{}, values)
		assert.Len(t, keys, 0)
	})
	t.Run("NewPhoto", func(t *testing.T) {
		m := &Photo{}
		values, keys, err := ModelValues(m, "ID", "PhotoUID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 0)
		assert.NotNil(t, values)
		assert.IsType(t, map[string]interface{}{}, values)
	})
	t.Run("ExistingPhoto", func(t *testing.T) {
		m := PhotoFixtures.Pointer("Photo01")
		values, keys, err := ModelValues(m, "ID", "PhotoUID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 2)
		assert.NotNil(t, values)
		assert.IsType(t, map[string]interface{}{}, values)
	})
	t.Run("NewFace", func(t *testing.T) {
		m := &Face{}
		values, keys, err := ModelValues(m, "ID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 0)
		assert.NotNil(t, values)
		assert.IsType(t, map[string]interface{}{}, values)
	})
	t.Run("ExistingFace", func(t *testing.T) {
		m := FaceFixtures.Pointer("john-doe")
		values, keys, err := ModelValues(m, "ID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 1)
		assert.NotNil(t, values)
		assert.IsType(t, map[string]interface{}{}, values)
	})
}

// Structs to test all the supported data types
type ParentStruct struct {
	ID                   int
	NullTime             sql.NullTime
	NullTimePointer      *sql.NullTime
	LocalTime            time.Time
	LocalTimePointer     *time.Time
	Duration             time.Duration
	DurationPointer      *time.Duration
	UnsignedInt          uint
	UnsignedIntPointer   *uint
	UnsignedInt32        uint32
	UnsignedIntPointer32 *uint32
	UnsignedInt64        uint64
	UnsignedIntPointer64 *uint64
	SignedInt            int
	SignedIntPointer     *int
	SignedInt32          int32
	SignedIntPointer32   *int32
	SignedInt64          int64
	SignedIntPointer64   *int64
	StringData           string
	StringPointer        *string
	Float32              float32
	Float32Pointer       *float32
	Float64              float64
	Float64Pointer       *float64
	Boolean              bool
	BooleanPointer       *bool
	OtpKey               otp.Key
	OtpKeyPointer        *otp.Key
	EmbeddingJSON        json.RawMessage
	EmbeddingJSONPointer *json.RawMessage
	ChildStruct1ID       int
	ChildStruct1         *ChildStruct1
	ChildStruct2s        []ChildStruct2
}

type ChildStruct1 struct {
	ID      int
	AColumn string
}

type ParentStruct2 struct {
	ID                int
	AnotherDataColumn string
	ChildStruct2s     []ChildStruct2
}

type ChildStruct2 struct {
	ParentID  int
	Parent2ID int
}

func TestModelValuesStructOption(t *testing.T) {
	t.Run("NoInterface", func(t *testing.T) {
		m := Photo{}
		values, keys, err := ModelValuesStructOption(m, true, "ID", "PhotoUID")

		assert.Error(t, err)
		assert.IsType(t, map[string]interface{}{}, values)
		assert.Len(t, keys, 0)

		values, keys, err = ModelValuesStructOption(m, false, "ID", "PhotoUID")

		assert.Error(t, err)
		assert.IsType(t, map[string]interface{}{}, values)
		assert.Len(t, keys, 0)
	})
	t.Run("NewPhoto", func(t *testing.T) {
		m := &Photo{}
		values, keys, err := ModelValuesStructOption(m, true, "ID", "PhotoUID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 0)
		assert.NotNil(t, values)
		assert.IsType(t, map[string]interface{}{}, values)

		values, keys, err = ModelValuesStructOption(m, false, "ID", "PhotoUID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 0)
		assert.NotNil(t, values)
		assert.IsType(t, map[string]interface{}{}, values)
	})
	t.Run("ExistingPhoto", func(t *testing.T) {
		original := PhotoFixtures.Pointer("Photo01")
		m := &Photo{}
		//log.Debugf("backup = %v", backup)
		if err := deepcopier.Copy(original).To(m); err != nil {
			assert.Nil(t, err)
			t.FailNow()
		}

		values, keys, err := ModelValuesStructOption(m, true, "ID", "PhotoUID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 2)
		assert.NotNil(t, values)
		assert.IsType(t, map[string]interface{}{}, values)
		assert.NotNil(t, m.Camera)
		assert.NotNil(t, m.Cell)
		assert.NotNil(t, m.Lens)
		assert.NotNil(t, m.Place)
		assert.NotNil(t, m.Keywords)
		assert.NotNil(t, m.Albums)
		assert.NotNil(t, m.Files)
		assert.NotNil(t, m.Labels)

		values, keys, err = ModelValuesStructOption(m, false, "ID", "PhotoUID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 2)
		assert.NotNil(t, values)
		assert.IsType(t, map[string]interface{}{}, values)
		assert.Nil(t, m.Camera)
		assert.Nil(t, m.Cell)
		assert.Nil(t, m.Lens)
		assert.Nil(t, m.Place)
		assert.Nil(t, m.Keywords)
		assert.Nil(t, m.Albums)
		assert.Nil(t, m.Files)
		assert.Nil(t, m.Labels)
	})
	t.Run("NewFace", func(t *testing.T) {
		m := &Face{}
		values, keys, err := ModelValuesStructOption(m, true, "ID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 0)
		assert.NotNil(t, values)
		assert.IsType(t, map[string]interface{}{}, values)

		values, keys, err = ModelValuesStructOption(m, false, "ID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 0)
		assert.NotNil(t, values)
		assert.IsType(t, map[string]interface{}{}, values)

	})

	t.Run("ExistingFace", func(t *testing.T) {
		original := FaceFixtures.Pointer("unknown")
		m := &Face{}
		//log.Debugf("backup = %v", backup)
		if err := deepcopier.Copy(original).To(m); err != nil {
			assert.Nil(t, err)
			t.FailNow()
		}

		values, keys, err := ModelValuesStructOption(m, true, "ID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 1)
		assert.NotNil(t, values)
		assert.IsType(t, map[string]interface{}{}, values)
		assert.Equal(t, original.FaceSrc, m.FaceSrc)
		assert.Equal(t, original.MatchedAt, m.MatchedAt)
		assert.Equal(t, original.CreatedAt, m.CreatedAt)
		assert.Equal(t, original.FaceHidden, m.FaceHidden)
		assert.NotNil(t, m.EmbeddingJSON)

		values, keys, err = ModelValuesStructOption(m, false, "ID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 1)
		assert.NotNil(t, values)
		assert.IsType(t, map[string]interface{}{}, values)
		assert.Equal(t, original.FaceSrc, m.FaceSrc)
		assert.Equal(t, original.MatchedAt, m.MatchedAt)
		assert.Equal(t, original.CreatedAt, m.CreatedAt)
		assert.Equal(t, original.FaceHidden, m.FaceHidden)
		assert.Nil(t, m.EmbeddingJSON)
	})

	t.Run("AllTypes", func(t *testing.T) {
		sqlNullTime := sql.NullTime{Time: time.Now().UTC(), Valid: true}
		localTime := time.Now().UTC()
		duration := time.Second * 60
		unsignedInt := uint(1)
		unsignedInt32 := uint32(1)
		unsignedInt64 := uint64(1)
		signedInt := int(1)
		signedInt32 := int32(1)
		signedInt64 := int64(1)
		stringData := "IAmAString"
		float32 := float32(12.34)
		float64 := float64(12.3456789)
		boolean := true
		otpKey, err := otp.NewKeyFromURL(`otpauth://totp/Example:alice@google.com?secret=JBSWY3DPEHPK3PXP&issuer=Example&algorithm=sha256&digits=8`)
		embeddingJSON := json.RawMessage(`{"precomputed": true}`)
		childStruct := ChildStruct1{ID: 12345, AColumn: "SomeData"}

		original := ParentStruct{ID: 1,
			NullTime:             sqlNullTime,
			NullTimePointer:      &sqlNullTime,
			LocalTime:            localTime,
			LocalTimePointer:     &localTime,
			Duration:             duration,
			DurationPointer:      &duration,
			UnsignedInt:          unsignedInt,
			UnsignedIntPointer:   &unsignedInt,
			UnsignedInt32:        unsignedInt32,
			UnsignedIntPointer32: &unsignedInt32,
			UnsignedInt64:        unsignedInt64,
			UnsignedIntPointer64: &unsignedInt64,
			SignedInt:            signedInt,
			SignedIntPointer:     &signedInt,
			SignedInt32:          signedInt32,
			SignedIntPointer32:   &signedInt32,
			SignedInt64:          signedInt64,
			SignedIntPointer64:   &signedInt64,
			StringData:           stringData,
			StringPointer:        &stringData,
			Float32:              float32,
			Float32Pointer:       &float32,
			Float64:              float64,
			Float64Pointer:       &float64,
			Boolean:              boolean,
			BooleanPointer:       &boolean,
			OtpKey:               *otpKey,
			OtpKeyPointer:        otpKey,
			EmbeddingJSON:        embeddingJSON,
			EmbeddingJSONPointer: &embeddingJSON,
			ChildStruct1ID:       12345,
			ChildStruct1:         &childStruct,
			ChildStruct2s:        []ChildStruct2{{Parent2ID: 33}},
		}
		m := &ParentStruct{}
		//log.Debugf("backup = %v", backup)
		if err := deepcopier.Copy(original).To(m); err != nil {
			assert.Nil(t, err)
			t.FailNow()
		}

		values, keys, err := ModelValuesStructOption(m, true, "ID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 1)
		assert.NotNil(t, values)
		assert.IsType(t, map[string]interface{}{}, values)
		assert.Equal(t, original.ID, m.ID)
		assert.Equal(t, original.NullTime, m.NullTime)
		assert.Equal(t, original.NullTimePointer, m.NullTimePointer)
		assert.Equal(t, original.LocalTime, m.LocalTime)
		assert.Equal(t, original.LocalTimePointer, m.LocalTimePointer)
		assert.Equal(t, original.Duration, m.Duration)
		assert.Equal(t, original.DurationPointer, m.DurationPointer)
		assert.Equal(t, original.UnsignedInt, m.UnsignedInt)
		assert.Equal(t, original.UnsignedIntPointer, m.UnsignedIntPointer)
		assert.Equal(t, original.UnsignedInt32, m.UnsignedInt32)
		assert.Equal(t, original.UnsignedIntPointer32, m.UnsignedIntPointer32)
		assert.Equal(t, original.UnsignedInt64, m.UnsignedInt64)
		assert.Equal(t, original.UnsignedIntPointer64, m.UnsignedIntPointer64)
		assert.Equal(t, original.SignedInt, m.SignedInt)
		assert.Equal(t, original.SignedIntPointer, m.SignedIntPointer)
		assert.Equal(t, original.SignedInt32, m.SignedInt32)
		assert.Equal(t, original.SignedIntPointer32, m.SignedIntPointer32)
		assert.Equal(t, original.SignedInt64, m.SignedInt64)
		assert.Equal(t, original.SignedIntPointer64, m.SignedIntPointer64)
		assert.Equal(t, original.StringData, m.StringData)
		assert.Equal(t, original.StringPointer, m.StringPointer)
		assert.Equal(t, original.Float32, m.Float32)
		assert.Equal(t, original.Float32Pointer, m.Float32Pointer)
		assert.Equal(t, original.Float64, m.Float64)
		assert.Equal(t, original.Float64Pointer, m.Float64Pointer)
		assert.Equal(t, original.Boolean, m.Boolean)
		assert.Equal(t, original.BooleanPointer, m.BooleanPointer)
		assert.Equal(t, original.OtpKey, m.OtpKey)
		assert.Equal(t, original.OtpKeyPointer, m.OtpKeyPointer)
		assert.Equal(t, original.EmbeddingJSON, m.EmbeddingJSON)
		assert.Equal(t, original.EmbeddingJSONPointer, m.EmbeddingJSONPointer)
		assert.Equal(t, original.ChildStruct1ID, m.ChildStruct1ID)
		assert.Equal(t, original.ChildStruct1, m.ChildStruct1)
		assert.Equal(t, original.ChildStruct2s, m.ChildStruct2s)

		values, keys, err = ModelValuesStructOption(m, false, "ID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 1)
		assert.NotNil(t, values)
		assert.IsType(t, map[string]interface{}{}, values)
		assert.Equal(t, original.ID, m.ID)
		assert.Equal(t, original.NullTime, m.NullTime)
		assert.Equal(t, original.NullTimePointer, m.NullTimePointer)
		assert.Equal(t, original.LocalTime, m.LocalTime)
		assert.Equal(t, original.LocalTimePointer, m.LocalTimePointer)
		assert.Equal(t, original.Duration, m.Duration)
		assert.Equal(t, original.DurationPointer, m.DurationPointer)
		assert.Equal(t, original.UnsignedInt, m.UnsignedInt)
		assert.Equal(t, original.UnsignedIntPointer, m.UnsignedIntPointer)
		assert.Equal(t, original.UnsignedInt32, m.UnsignedInt32)
		assert.Equal(t, original.UnsignedIntPointer32, m.UnsignedIntPointer32)
		assert.Equal(t, original.UnsignedInt64, m.UnsignedInt64)
		assert.Equal(t, original.UnsignedIntPointer64, m.UnsignedIntPointer64)
		assert.Equal(t, original.SignedInt, m.SignedInt)
		assert.Equal(t, original.SignedIntPointer, m.SignedIntPointer)
		assert.Equal(t, original.SignedInt32, m.SignedInt32)
		assert.Equal(t, original.SignedIntPointer32, m.SignedIntPointer32)
		assert.Equal(t, original.SignedInt64, m.SignedInt64)
		assert.Equal(t, original.SignedIntPointer64, m.SignedIntPointer64)
		assert.Equal(t, original.StringData, m.StringData)
		assert.Equal(t, original.StringPointer, m.StringPointer)
		assert.Equal(t, original.Float32, m.Float32)
		assert.Equal(t, original.Float32Pointer, m.Float32Pointer)
		assert.Equal(t, original.Float64, m.Float64)
		assert.Equal(t, original.Float64Pointer, m.Float64Pointer)
		assert.Equal(t, original.Boolean, m.Boolean)
		assert.Equal(t, original.BooleanPointer, m.BooleanPointer)
		assert.Equal(t, original.OtpKey, m.OtpKey)
		assert.Equal(t, original.OtpKeyPointer, m.OtpKeyPointer)
		assert.Equal(t, original.EmbeddingJSON, m.EmbeddingJSON)
		assert.Equal(t, original.EmbeddingJSONPointer, m.EmbeddingJSONPointer)
		assert.Equal(t, original.ChildStruct1ID, m.ChildStruct1ID)
		assert.Nil(t, m.ChildStruct1)
		assert.Nil(t, m.ChildStruct2s)
	})
}
