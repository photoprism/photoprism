package form

// UserPasscode represents a multi-factor authentication key setup form.
type UserPasscode struct {
	Type     string `form:"type" json:"type,omitempty"`
	Passcode string `form:"passcode" json:"passcode,omitempty"`
	Password string `form:"password" json:"password,omitempty"`
}

// HasPassword checks if a password is set.
func (f UserPasscode) HasPassword() bool {
	return f.Password != "" && len(f.Password) <= 255
}
