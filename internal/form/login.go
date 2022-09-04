package form

type Login struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

func (f Login) HasToken() bool {
	return f.Token != ""
}

func (f Login) HasUsername() bool {
	return f.Username != "" && len(f.Username) <= 255
}

func (f Login) HasPassword() bool {
	return f.Password != "" && len(f.Password) <= 255
}

func (f Login) HasCredentials() bool {
	return f.HasUsername() && f.HasPassword()
}
