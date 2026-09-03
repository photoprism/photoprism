package entity

import (
	"time"
)

// JustDeletedAt represents the structure of the DeletedAt field prior to Gorm V1
// yaml.Unmarshal errors when the yaml file contains the Gorm V1 structure, but does populate all other fields
// This is used to grab the lost DeletedAt when yaml.Unmarshal reports the DeletedAt error.

type JustDeletedAt struct {
	DeletedAt time.Time `json:"DeletedAt" yaml:"DeletedAt,omitempty"`
}
