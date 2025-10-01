package entity

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestGorm(t *testing.T) {
	dbtestMutex.Lock()
	defer dbtestMutex.Unlock()

	// This test shows the underlying way that PhotoPrism creates a new Photo with GormV1
	t.Run("Photo_Gorm1", func(t *testing.T) {
		m := entity.PhotoFixtures.Pointer("Photo09")

		values, keys, err := modelValuesStructOption(m, false, "ID", "PhotoUID")
		log.Debugf("Keys = %v", keys)
		if err == nil {
			for k, v := range values {
				s := fmt.Sprintf("%v = [%#v]", k, v)
				log.Debug(s)
			}
		}
		log.Debugf("Photo = %v", m)

		values, keys, err = modelValues(m, "ID", "PhotoUID")
		log.Debugf("Keys = %v", keys)
		if err == nil {
			for k, v := range values {
				s := fmt.Sprintf("%v = [%#v]", k, v)
				log.Debug(s)
			}
		}

	})

	t.Run("File_Gorm1", func(t *testing.T) {
		m := entity.FileFixtures.Pointer("exampleFileName.jpg")

		values, keys, err := modelValuesStructOption(m, false, "ID", "PhotoUID")
		log.Debugf("Keys = %v", keys)
		if err == nil {
			for k, v := range values {
				s := fmt.Sprintf("%v = [%#v]", k, v)
				log.Debug(s)
			}
		}

	})

	t.Run("Passcode_Gorm1", func(t *testing.T) {
		m := entity.PasscodeFixtures.Pointer("alice")

		values, keys, err := modelValuesStructOption(m, false, "ID", "PhotoUID")
		log.Debugf("Keys = %v", keys)
		if err == nil {
			for k, v := range values {
				s := fmt.Sprintf("%v = [%#v]", k, v)
				log.Debug(s)
			}
		}

	})

	t.Run("Session_Gorm1", func(t *testing.T) {
		m := entity.SessionFixtures.Pointer("alice")

		values, keys, err := modelValuesStructOption(m, false, "ID", "PhotoUID")
		log.Debugf("Keys = %v", keys)
		if err == nil {
			for k, v := range values {
				s := fmt.Sprintf("%v = [%#v]", k, v)
				log.Debug(s)
			}
		}

	})

	t.Run("Service_Gorm1", func(t *testing.T) {
		m := entity.ServiceFixtures["dummy-webdav"]
		values, keys, err := modelValues(&m, "ID", "PhotoUID")
		log.Debugf("Keys = %v", keys)
		if err == nil {
			for k, v := range values {
				s := fmt.Sprintf("%v = [%#v]", k, v)
				log.Debug(s)
			}
		}

		m = entity.ServiceFixtures["dummy-webdav"]
		values, keys, err = modelValuesStructOption(&m, false, "ID", "PhotoUID")
		log.Debugf("Keys = %v", keys)
		if err == nil {
			for k, v := range values {
				s := fmt.Sprintf("%v = [%#v]", k, v)
				log.Debug(s)
			}
		}
	})

	t.Run("Country_Gorm1", func(t *testing.T) {
		cid := uint(12345)
		m := entity.Country{
			ID:                 "de",
			CountrySlug:        "germany",
			CountryName:        "Germany",
			CountryDescription: "Country description",
			CountryNotes:       "Country Notes",
			CountryPhoto:       nil,
			CountryPhotoID:     &cid,
			New:                false,
		}

		values, keys, err := modelValuesStructOption(&m, false, "ID", "PhotoUID")
		log.Debugf("Keys = %v", keys)
		if err == nil {
			for k, v := range values {
				s := fmt.Sprintf("%v = [%#v]", k, v)
				log.Debug(s)
			}
		} else {
			log.Debugf("error was %v", err)
		}

	})

}

func modelValues(m interface{}, omit ...string) (result entity.Values, omitted []interface{}, err error) {
	return modelValuesStructOption(m, true, omit...)
}

// ModelValues extracts Values from an entity model.
func modelValuesStructOption(m interface{}, includeStruct bool, omit ...string) (result entity.Values, omitted []interface{}, err error) {
	mustOmit := func(name string) bool {
		for _, s := range omit {
			if name == s {
				return true
			}
		}

		return false
	}

	r := reflect.ValueOf(m)

	if r.Kind() != reflect.Pointer {
		return result, omitted, fmt.Errorf("model interface expected")
	}

	values := r.Elem()

	if kind := values.Kind(); kind != reflect.Struct {
		return result, omitted, fmt.Errorf("model expected")
	}

	t := values.Type()
	num := t.NumField()

	omitted = make([]interface{}, 0, len(omit))
	result = make(entity.Values, num)

	// Add exported fields to result.
	for i := 0; i < num; i++ {
		field := t.Field(i)

		// Skip non-exported fields.
		if !field.IsExported() {
			continue
		}

		fieldName := field.Name

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
			whitelist := false
			switch v.Type().Name() {
			case "NullTime", "Time":
				whitelist = true
			}
			if !whitelist && !includeStruct {
				log.Debugf("Struct Zeroing %v of type %v or string %v", fieldName, v.Type().Name(), v.Type().String())
				v.SetZero()
				continue
			}
		case reflect.Pointer:
			whitelist := false
			switch v.Type().String() {
			case "*time.Time", "*uint", "*uint64", "*uint32", "*int", "*int64", "*int32", "*string", "*float32", "*float64", "*otp.Key", "sql.NullTime":
				whitelist = true
			}
			if !whitelist && !includeStruct {
				log.Debugf("Pointer Zeroing %v of type %v or string %v", fieldName, v.Type().Name(), v.Type().String())
				v.SetZero()
				continue
			}
		}

		// Skip read-only fields.
		if !v.CanSet() {
			continue
		}

		// Skip omitted.
		if mustOmit(fieldName) {
			if !v.IsZero() {
				omitted = append(omitted, v.Interface())
			}
			continue
		}

		// Add value to result.
		result[fieldName] = v.Interface()
	}

	if len(result) == 0 {
		return result, omitted, fmt.Errorf("no values")
	}

	return result, omitted, nil
}
