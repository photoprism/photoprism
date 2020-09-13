package entity

import (
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Camera model and make (as extracted from UpdateExif metadata)
type Camera struct {
	ID                uint       `gorm:"primary_key" json:"ID" yaml:"ID"`
	CameraSlug        string     `gorm:"type:VARBINARY(255);unique_index;" json:"Slug" yaml:"-"`
	CameraName        string     `gorm:"type:VARCHAR(255);" json:"Name" yaml:"Name"`
	CameraMake        string     `gorm:"type:VARCHAR(255);" json:"Make" yaml:"Make,omitempty"`
	CameraModel       string     `gorm:"type:VARCHAR(255);" json:"Model" yaml:"Model,omitempty"`
	CameraType        string     `gorm:"type:VARCHAR(255);" json:"Type,omitempty" yaml:"Type,omitempty"`
	CameraDescription string     `gorm:"type:TEXT;" json:"Description,omitempty" yaml:"Description,omitempty"`
	CameraNotes       string     `gorm:"type:TEXT;" json:"Notes,omitempty" yaml:"Notes,omitempty"`
	CreatedAt         time.Time  `json:"-" yaml:"-"`
	UpdatedAt         time.Time  `json:"-" yaml:"-"`
	DeletedAt         *time.Time `sql:"index" json:"-" yaml:"-"`
}

var UnknownCamera = Camera{
	CameraSlug:  "zz",
	CameraName:  "Unknown",
	CameraMake:  "",
	CameraModel: "Unknown",
}

var CameraMakes = map[string]string{
	"OLYMPUS OPTICAL CO.,LTD": "Olympus",
	"samsung":                 "Samsung",
}

var CameraModels = map[string]string{
	"WAS-LX1":    "P10 lite",
	"WAS-LX2":    "P10 lite",
	"WAS-LX3":    "P10 lite",
	"WAS-LX1A":   "P10 lite",
	"WAS-LX2J":   "P10 lite",
	"WAS-L03T":   "P10 lite",
	"WAS-AL00":   "P10 lite",
	"WAS-TL10":   "P10 lite",
	"VTR-L29":    "P10",
	"VTR-AL00":   "P10",
	"VTR-TL00":   "P10",
	"VTR-L09":    "P10",
	"EML-AL00":   "P20",
	"EML-L09":    "P20",
	"EML-L09C":   "P20",
	"EML-L29":    "P20",
	"EML-L29C":   "P20",
	"CLT-AL00":   "P20 Pro",
	"CLT-AL01":   "P20 Pro",
	"CLT-TL01":   "P20 Pro",
	"CLT-L09":    "P20 Pro",
	"CLT-L29":    "P20 Pro",
	"ELE-L29":    "P30",
	"ELE-AL00":   "P30",
	"ELE-L04":    "P30",
	"ELE-L09":    "P30",
	"ELE-TL00":   "P30",
	"VOG-L29":    "P30 Pro",
	"VOG-L09":    "P30 Pro",
	"VOG-L04":    "P30 Pro",
	"VOG-AL00":   "P30 Pro",
	"VOG-AL10":   "P30 Pro",
	"VOG-TL00":   "P30 Pro",
	"MAR-L01A":   "P30 lite",
	"MAR-L21A":   "P30 lite",
	"MAR-LX1A":   "P30 lite",
	"MAR-LX1M":   "P30 lite",
	"MAR-LX2":    "P30 lite",
	"MAR-L21MEA": "P30 lite",
	"MAR-L22A":   "P30 lite",
	"MAR-L22B":   "P30 lite",
	"MAR-LX3A":   "P30 lite",
	"ANA-AN00":   "P40",
	"ANA-TN00":   "P40",
	"ELS-AN00":   "P40 Pro",
	"ELS-TN00":   "P40 Pro",
	"ELS-NX9":    "P40 Pro",
	"ELS-N04":    "P40 Pro",
	"JNY-L21A":   "P40 lite",
	"JNY-L01A":   "P40 lite",
	"JNY-L21B":   "P40 lite",
	"JNY-L22A":   "P40 lite",
	"JNY-L02A":   "P40 lite",
	"JNY-L22B":   "P40 lite",
	"STK-LX1":    "Honor 9X",
	"HLK-AL00":   "Honor 9X",
	"HLK-TL00":   "Honor 9X",
	"SNE-AL00":   "Mate 20 lite",
	"SNE-LX1":    "Mate 20 lite",
	"SNE-LX2":    "Mate 20 lite",
	"SNE-LX3":    "Mate 20 lite",
	"INE-LX2":    "Mate 20 lite",
	"HMA-L29":    "Mate 20",
	"HMA-L09":    "Mate 20",
	"HMA-LX9":    "Mate 20",
	"HMA-AL00":   "Mate 20",
	"HMA-TL00":   "Mate 20",
	"LYA-L09":    "Mate 20 Pro",
	"LYA-L29":    "Mate 20 Pro",
	"LYA-AL00":   "Mate 20 Pro",
	"LYA-AL10":   "Mate 20 Pro",
	"LYA-TL00":   "Mate 20 Pro",
	"LYA-L0C":    "Mate 20 Pro",
	"TAS-L09":    "Mate 30",
	"TAS-L29":    "Mate 30",
	"TAS-AL00":   "Mate 30",
	"TAS-TL00":   "Mate 30",
	"LIO-L09":    "Mate 30 Pro",
	"LIO-L29":    "Mate 30 Pro",
	"LIO-AL00":   "Mate 30 Pro",
	"LIO-TL00":   "Mate 30 Pro",
}

// CreateUnknownCamera initializes the database with an unknown camera if not exists
func CreateUnknownCamera() {
	FirstOrCreateCamera(&UnknownCamera)
}

// NewCamera creates a camera entity from a model name and a make name.
func NewCamera(modelName string, makeName string) *Camera {
	modelName = txt.Clip(modelName, txt.ClipDefault)
	makeName = txt.Clip(makeName, txt.ClipDefault)

	if modelName == "" && makeName == "" {
		return &UnknownCamera
	} else if strings.HasPrefix(modelName, makeName) {
		modelName = strings.TrimSpace(modelName[len(makeName):])
	}

	if n, ok := CameraMakes[makeName]; ok {
		makeName = n
	}

	if n, ok := CameraModels[modelName]; ok {
		modelName = n
	}

	var name []string

	if makeName != "" {
		name = append(name, makeName)
	}

	if modelName != "" {
		name = append(name, modelName)
	}

	cameraName := strings.Join(name, " ")
	cameraSlug := slug.Make(txt.Clip(cameraName, txt.ClipSlug))

	result := &Camera{
		CameraSlug:  cameraSlug,
		CameraName:  cameraName,
		CameraMake:  makeName,
		CameraModel: modelName,
	}

	return result
}

// Create inserts a new row to the database.
func (m *Camera) Create() error {
	return Db().Create(m).Error
}

// FirstOrCreateCamera returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateCamera(m *Camera) *Camera {
	result := Camera{}

	if err := Db().Where("camera_model = ? AND camera_make = ?", m.CameraModel, m.CameraMake).First(&result).Error; err == nil {
		return &result
	} else if createErr := m.Create(); createErr == nil {
		if !m.Unknown() {
			event.EntitiesCreated("cameras", []*Camera{m})

			event.Publish("count.cameras", event.Data{
				"count": 1,
			})
		}

		return m
	} else if err := Db().Where("camera_model = ? AND camera_make = ?", m.CameraModel, m.CameraMake).First(&result).Error; err == nil {
		return &result
	} else {
		log.Errorf("camera: %s (first or create %s)", createErr, m.String())
	}

	return nil
}

// String returns an identifier that can be used in logs.
func (m *Camera) String() string {
	return m.CameraName
}

// Unknown returns true if the camera is not a known make or model.
func (m *Camera) Unknown() bool {
	return m.CameraSlug == "" || m.CameraSlug == UnknownCamera.CameraSlug
}
