package entity

import (
	"fmt"
	"time"
)

type Addresses []Address

// Address represents a postal address.
type Address struct {
	ID             int        `gorm:"primary_key" json:"ID" yaml:"ID"`
	CellID         string     `gorm:"type:VARBINARY(42);index;default:'zz'" json:"CellID" yaml:"CellID"`
	AddressSrc     string     `gorm:"type:VARBINARY(8);" json:"Src" yaml:"Src"`
	AddressLat     float32    `gorm:"type:FLOAT;index;" json:"Lat" yaml:"Lat,omitempty"`
	AddressLng     float32    `gorm:"type:FLOAT;index;" json:"Lng" yaml:"Lng,omitempty"`
	AddressLine1   string     `gorm:"size:255;" json:"Line1" yaml:"Line1,omitempty"`
	AddressLine2   string     `gorm:"size:255;" json:"Line2" yaml:"Line2,omitempty"`
	AddressZip     string     `gorm:"size:32;" json:"Zip" yaml:"Zip,omitempty"`
	AddressCity    string     `gorm:"size:128;" json:"City" yaml:"City,omitempty"`
	AddressState   string     `gorm:"size:128;" json:"State" yaml:"State,omitempty"`
	AddressCountry string     `gorm:"type:VARBINARY(2);default:'zz'" json:"Country" yaml:"Country,omitempty"`
	AddressNotes   string     `gorm:"type:TEXT;" json:"Notes" yaml:"Notes,omitempty"`
	CreatedAt      time.Time  `json:"CreatedAt" yaml:"-"`
	UpdatedAt      time.Time  `json:"UpdatedAt" yaml:"-"`
	DeletedAt      *time.Time `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
}

// TableName returns the entity database table name.
func (Address) TableName() string {
	return "addresses"
}

// UnknownAddress may be used as placeholder for an unknown address.
var UnknownAddress = Address{
	CellID:         UnknownLocation.ID,
	AddressSrc:     SrcDefault,
	AddressCountry: UnknownCountry.ID,
}

// CreateUnknownAddress creates the default address if not exists.
func CreateUnknownAddress() {
	UnknownAddress = *FirstOrCreateAddress(&UnknownAddress)
}

// FirstOrCreateAddress returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateAddress(m *Address) *Address {
	result := Address{}

	if err := Db().Where("cell_id = ?", m.CellID).First(&result).Error; err == nil {
		return &result
	} else if createErr := m.Create(); createErr == nil {
		return m
	} else if err := Db().Where("cell_id = ?", m.CellID).First(&result).Error; err == nil {
		return &result
	} else {
		log.Errorf("address: %s (find or create %s)", createErr, m.String())
	}

	return nil
}

// String returns an identifier that can be used in logs.
func (m *Address) String() string {
	return fmt.Sprintf("%s, %s %s, %s", m.AddressLine1, m.AddressZip, m.AddressCity, m.AddressCountry)
}

// Unknown returns true if the address is unknown.
func (m *Address) Unknown() bool {
	return m.ID < 0
}

// Create inserts a new row to the database.
func (m *Address) Create() error {
	return Db().Create(m).Error
}

// Saves the new row to the database.
func (m *Address) Save() error {
	return Db().Save(m).Error
}
