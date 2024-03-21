package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const (
	GenderMale   = "male"
	GenderFemale = "female"
	GenderOther  = "other"
)

// UserDetails represents user profile information.
type UserDetails struct {
	UserUID      string    `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;" json:"-" yaml:"-"`
	SubjUID      string    `gorm:"type:VARBINARY(42);index;" json:"SubjUID,omitempty" yaml:"SubjUID,omitempty"`
	SubjSrc      string    `gorm:"type:VARBINARY(8);default:'';" json:"-" yaml:"SubjSrc,omitempty"`
	PlaceID      string    `gorm:"type:VARBINARY(42);index;default:'zz'" json:"-" yaml:"-"`
	PlaceSrc     string    `gorm:"type:VARBINARY(8);" json:"-" yaml:"PlaceSrc,omitempty"`
	CellID       string    `gorm:"type:VARBINARY(42);index;default:'zz'" json:"-" yaml:"CellID,omitempty"`
	BirthYear    int       `json:"BirthYear" yaml:"BirthYear,omitempty"`
	BirthMonth   int       `json:"BirthMonth" yaml:"BirthMonth,omitempty"`
	BirthDay     int       `json:"BirthDay" yaml:"BirthDay,omitempty"`
	NameTitle    string    `gorm:"size:32;" json:"NameTitle" yaml:"NameTitle,omitempty"`
	GivenName    string    `gorm:"size:64;" json:"GivenName" yaml:"GivenName,omitempty"`
	MiddleName   string    `gorm:"size:64;" json:"MiddleName" yaml:"MiddleName,omitempty"`
	FamilyName   string    `gorm:"size:64;" json:"FamilyName" yaml:"FamilyName,omitempty"`
	NameSuffix   string    `gorm:"size:32;" json:"NameSuffix" yaml:"NameSuffix,omitempty"`
	NickName     string    `gorm:"size:64;" json:"NickName" yaml:"NickName,omitempty"`
	NameSrc      string    `gorm:"type:VARBINARY(8);" json:"NameSrc" yaml:"NameSrc,omitempty"`
	UserGender   string    `gorm:"size:16;" json:"Gender" yaml:"Gender,omitempty"`
	UserAbout    string    `gorm:"size:512;" json:"About" yaml:"About,omitempty"`
	UserBio      string    `gorm:"size:2048;" json:"Bio" yaml:"Bio,omitempty"`
	UserLocation string    `gorm:"size:512;" json:"Location" yaml:"Location,omitempty"`
	UserCountry  string    `gorm:"type:VARBINARY(2);" json:"Country" yaml:"Country,omitempty"`
	UserPhone    string    `gorm:"size:32;" json:"Phone" yaml:"Phone,omitempty"`
	SiteURL      string    `gorm:"type:VARBINARY(512);column:site_url" json:"SiteURL" yaml:"SiteURL,omitempty"`
	ProfileURL   string    `gorm:"type:VARBINARY(512);column:profile_url" json:"ProfileURL" yaml:"ProfileURL,omitempty"`
	FeedURL      string    `gorm:"type:VARBINARY(512);column:feed_url" json:"FeedURL,omitempty" yaml:"FeedURL,omitempty"`
	AvatarURL    string    `gorm:"type:VARBINARY(512);column:avatar_url" json:"AvatarURL,omitempty" yaml:"AvatarURL,omitempty"`
	OrgTitle     string    `gorm:"size:64;" json:"OrgTitle" yaml:"OrgTitle,omitempty"`
	OrgName      string    `gorm:"size:128;" json:"OrgName" yaml:"OrgName,omitempty"`
	OrgEmail     string    `gorm:"size:255;index;" json:"OrgEmail" yaml:"OrgEmail,omitempty"`
	OrgPhone     string    `gorm:"size:32;" json:"OrgPhone" yaml:"OrgPhone,omitempty"`
	OrgURL       string    `gorm:"type:VARBINARY(512);column:org_url" json:"OrgURL" yaml:"OrgURL,omitempty"`
	IdURL        string    `gorm:"type:VARBINARY(512);column:id_url;" json:"IdURL,omitempty" yaml:"IdURL,omitempty"`
	CreatedAt    time.Time `json:"CreatedAt" yaml:"-"`
	UpdatedAt    time.Time `json:"UpdatedAt" yaml:"-"`
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

// DisplayName returns a display name based on the user details.
func (m *UserDetails) DisplayName() string {
	if m == nil {
		return ""
	}

	n := make([]string, 0, 3)

	// Add title if exists.
	if m.NameTitle != "" {
		n = append(n, m.NameTitle)
	}

	// Add given name if exists.
	if m.GivenName != "" {
		n = append(n, m.GivenName)
	}

	// Add family name if exists.
	if m.FamilyName != "" {
		n = append(n, m.FamilyName)
	}

	// Default to nick name.
	if len(n) == 0 {
		return m.NickName
	}

	return strings.Join(n, " ")
}

// SetGivenName updates the user's given name.
func (m *UserDetails) SetGivenName(name string) *UserDetails {
	name = clean.Name(name)

	if name == "" || SrcPriority[SrcAuto] < SrcPriority[m.NameSrc] {
		return m
	}

	m.GivenName = name

	return m
}

// SetFamilyName updates the user's family name.
func (m *UserDetails) SetFamilyName(name string) *UserDetails {
	name = clean.Name(name)

	if name == "" || SrcPriority[SrcAuto] < SrcPriority[m.NameSrc] {
		return m
	}

	m.FamilyName = name

	return m
}
