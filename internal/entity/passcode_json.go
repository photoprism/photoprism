package entity

import (
	"encoding/json"
	"time"

	"github.com/photoprism/photoprism/pkg/media"
)

// MarshalJSON returns the JSON encoding.
func (m *Passcode) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		UID          string     `json:"UID"`
		Type         string     `json:"Type"`
		Secret       string     `json:"Secret"`
		QRCode       string     `json:"QRCode"`
		RecoveryCode string     `json:"RecoveryCode"`
		CreatedAt    time.Time  `json:"CreatedAt"`
		UpdatedAt    time.Time  `json:"UpdatedAt"`
		VerifiedAt   *time.Time `json:"VerifiedAt"`
		ActivatedAt  *time.Time `json:"ActivatedAt"`
	}{
		UID:          m.UID,
		Type:         m.KeyType,
		Secret:       m.Secret(),
		QRCode:       media.Base64(m.PNG(350)),
		RecoveryCode: m.RecoveryCode,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		VerifiedAt:   m.VerifiedAt,
		ActivatedAt:  m.ActivatedAt,
	})
}
