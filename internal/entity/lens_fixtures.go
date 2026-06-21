package entity

import (
	"time"
)

type LensMap map[string]Lens

func (m LensMap) Get(name string) Lens {
	if result, ok := m[name]; ok {
		return result
	}

	return *NewLens("", name)
}

func (m LensMap) Pointer(name string) *Lens {
	if result, ok := m[name]; ok {
		return &result
	}

	return NewLens("", name)
}

// LensFixtures holds the set of test fixtures for Lenses
var LensFixtures = LensMap{
	"lens-f-380": {
		ID:              1000000,
		LensSlug:        "lens-f-380",
		LensName:        "Apple F380",
		LensMake:        "Apple",
		LensModel:       "F380",
		LensType:        "",
		LensDescription: "",
		LensNotes:       "Notes",
		CreatedAt:       time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:       nil,
	},
	"4.15mm-f/2.2": {
		ID:              1000001,
		LensSlug:        "apple-iphone-se-4-15mm-f-2-2",
		LensName:        "Apple iPhone SE 4.15mm f/2.2",
		LensMake:        "Apple",
		LensModel:       "iPhone SE 4.15mm f/2.2",
		LensType:        "",
		LensDescription: "",
		LensNotes:       "Notes",
		CreatedAt:       time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:       nil,
	},
	"4-37": {
		ID:              1000002,
		LensSlug:        "4-37",
		LensName:        "4 37",
		LensMake:        "",
		LensModel:       "4 37",
		LensType:        "",
		LensDescription: "",
		LensNotes:       "",
		CreatedAt:       time.Date(2026, 06, 12, 10, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2026, 06, 12, 10, 0, 0, 0, time.UTC),
		DeletedAt:       nil,
	}}

// CreateLensFixtures inserts known entities into the database for testing.
func CreateLensFixtures() {
	for _, entity := range LensFixtures {
		Db().Create(&entity)
	}
}
