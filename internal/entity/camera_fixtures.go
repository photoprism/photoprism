package entity

import (
	"time"
)

type CameraMap map[string]Camera

func (m CameraMap) Get(name string) Camera {
	if result, ok := m[name]; ok {
		return result
	}

	return *NewCamera(name, "")
}

func (m CameraMap) Pointer(name string) *Camera {
	if result, ok := m[name]; ok {
		return &result
	}

	return NewCamera(name, "")
}

var CameraFixtures = CameraMap{
	"apple-iphone-se": {
		ID:                1000000,
		CameraSlug:        "apple-iphone-se",
		CameraModel:       "iPhone SE",
		CameraMake:        "Apple",
		CameraType:        "",
		CameraDescription: "",
		CameraNotes:       "",
		CreatedAt:         time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:         time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:         nil,
	},
	"canon-eos-5d": {
		ID:                1000001,
		CameraSlug:        "canon-eos-5d",
		CameraModel:       "EOS 5D",
		CameraMake:        "Canon",
		CameraType:        "",
		CameraDescription: "",
		CameraNotes:       "",
		CreatedAt:         time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:         time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:         nil,
	},
	"canon-eos-7d": {
		ID:                1000002,
		CameraSlug:        "canon-eos-7d",
		CameraModel:       "EOS 7D",
		CameraMake:        "Canon",
		CameraType:        "",
		CameraDescription: "",
		CameraNotes:       "",
		CreatedAt:         time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:         time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:         nil,
	},
	"canon-eos-6d": {
		ID:                1000003,
		CameraSlug:        "canon-eos-6d",
		CameraModel:       "EOS 6D",
		CameraMake:        "Canon",
		CameraType:        "",
		CameraDescription: "",
		CameraNotes:       "",
		CreatedAt:         time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:         time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:         nil,
	},
	"apple-iphone-6": {
		ID:                1000004,
		CameraSlug:        "apple-iphone-6",
		CameraModel:       "iPhone 6",
		CameraMake:        "Apple",
		CameraType:        "",
		CameraDescription: "",
		CameraNotes:       "",
		CreatedAt:         time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:         time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:         nil,
	},
	"apple-iphone-7": {
		ID:                1000005,
		CameraSlug:        "apple-iphone-7",
		CameraModel:       "iPhone 7",
		CameraMake:        "Apple",
		CameraType:        "",
		CameraDescription: "",
		CameraNotes:       "",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		DeletedAt:         nil,
	},
}

// CreateCameraFixtures inserts known entities into the database for testing.
func CreateCameraFixtures() {
	for _, entity := range CameraFixtures {
		Db().Create(&entity)
	}
}
