package entity

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/pkg/rnd"
)

const (
	GenderMale    = "male"
	GenderFemale  = "female"
	GenderOther   = "other"
	GenderUnknown = ""
)

// UserDetails represents user profile information.
type UserDetails struct {
	UserUID     string    `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;" json:"-" yaml:"UserUID"`
	SubjUID     string    `gorm:"type:VARBINARY(42);index;" json:"SubjUID,omitempty" yaml:"SubjUID,omitempty"`
	SubjSrc     string    `gorm:"type:VARBINARY(8);default:'';" json:"SubjSrc,omitempty" yaml:"SubjSrc,omitempty"`
	PlaceID     string    `gorm:"type:VARBINARY(42);index;default:'zz'" json:"PlaceID,omitempty" yaml:"-"`
	PlaceSrc    string    `gorm:"type:VARBINARY(8);" json:"PlaceSrc,omitempty" yaml:"PlaceSrc,omitempty"`
	CellID      string    `gorm:"type:VARBINARY(42);index;default:'zz'" json:"CellID,omitempty" yaml:"CellID,omitempty"`
	IdURL       string    `gorm:"type:VARBINARY(512);column:id_url;" json:"IdURL,omitempty" yaml:"IdURL,omitempty"`
	AvatarURL   string    `gorm:"type:VARBINARY(512);column:avatar_url" json:"AvatarURL,omitempty" yaml:"AvatarURL,omitempty"`
	SiteURL     string    `gorm:"type:VARBINARY(512);column:site_url" json:"SiteURL,omitempty" yaml:"SiteURL,omitempty"`
	FeedURL     string    `gorm:"type:VARBINARY(512);column:feed_url" json:"FeedURL,omitempty" yaml:"FeedURL,omitempty"`
	UserGender  string    `gorm:"size:16;" json:"Gender,omitempty" yaml:"Gender,omitempty"`
	NamePrefix  string    `gorm:"size:32;" json:"NamePrefix,omitempty" yaml:"NamePrefix,omitempty"`
	GivenName   string    `gorm:"size:64;" json:"GivenName,omitempty" yaml:"GivenName,omitempty"`
	MiddleName  string    `gorm:"size:64;" json:"MiddleName,omitempty" yaml:"MiddleName,omitempty"`
	FamilyName  string    `gorm:"size:64;" json:"FamilyName,omitempty" yaml:"FamilyName,omitempty"`
	NameSuffix  string    `gorm:"size:32;" json:"NameSuffix,omitempty" yaml:"NameSuffix,omitempty"`
	NickName    string    `gorm:"size:64;" json:"NickName,omitempty" yaml:"NickName,omitempty"`
	UserPhone   string    `gorm:"size:32;" json:"Phone,omitempty" yaml:"Phone,omitempty"`
	UserAddress string    `gorm:"size:512;" json:"Address,omitempty" yaml:"Address,omitempty"`
	UserCountry string    `gorm:"type:VARBINARY(2);" json:"Country,omitempty" yaml:"Country,omitempty"`
	UserBio     string    `gorm:"size:512;" json:"Bio,omitempty" yaml:"Bio,omitempty"`
	JobTitle    string    `gorm:"size:64;" json:"JobTitle,omitempty" yaml:"JobTitle,omitempty"`
	Department  string    `gorm:"size:128;" json:"Department,omitempty" yaml:"Department,omitempty"`
	Company     string    `gorm:"size:128;" json:"Company,omitempty" yaml:"Company,omitempty"`
	CompanyURL  string    `gorm:"type:VARBINARY(512);column:company_url" json:"CompanyURL,omitempty" yaml:"CompanyURL,omitempty"`
	BirthYear   int       `gorm:"default:-1;" json:"BirthYear,omitempty" yaml:"BirthYear,omitempty"`
	BirthMonth  int       `gorm:"default:-1;" json:"BirthMonth,omitempty" yaml:"BirthMonth,omitempty"`
	BirthDay    int       `gorm:"default:-1;" json:"BirthDay,omitempty" yaml:"BirthDay,omitempty"`
	CreatedAt   time.Time `json:"CreatedAt" yaml:"-"`
	UpdatedAt   time.Time `json:"UpdatedAt" yaml:"-"`
}

// TableName returns the entity table name.
func (UserDetails) TableName() string {
	return "auth_users_details"
}

// NewUserDetails creates new user details.
func NewUserDetails(uid string) *UserDetails {
	return &UserDetails{UserUID: uid}
}

// CreateUserDetails creates new user details or returns nil on error.
func CreateUserDetails(user *User) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	if user.UID() == "" {
		return fmt.Errorf("empty user uid")
	}

	user.UserDetails = NewUserDetails(user.UID())

	if err := Db().Where("user_uid = ?", user.UID()).First(user.UserDetails).Error; err == nil {
		return nil
	}

	return user.UserDetails.Create()
}

// HasID tests if the entity has a valid uid.
func (m *UserDetails) HasID() bool {
	return rnd.IsUID(m.UserUID, UserUID)
}

// Create new entity in the database.
func (m *UserDetails) Create() error {
	return Db().Create(m).Error
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *UserDetails) Save() error {
	return Db().Save(m).Error
}

// Updates multiple properties in the database.
func (m *UserDetails) Updates(values interface{}) error {
	return UnscopedDb().Model(m).Updates(values).Error
}
