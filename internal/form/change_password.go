package form

// ChangePassword represents a password update form.
type ChangePassword struct {
	OldPassword string `json:"old"`
	NewPassword string `json:"new"`
}
