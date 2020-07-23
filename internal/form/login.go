package form

type Login struct {
	Email    string `json:"email"`
	UserName string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

func (f Login) HasToken() bool {
	return f.Token != ""
}

func (f Login) HasUserName() bool {
	return f.UserName != "" && len(f.UserName) <= 255
}

func (f Login) HasPassword() bool {
	return f.Password != "" && len(f.Password) <= 255
}

func (f Login) HasCredentials() bool {
	return f.HasUserName() && f.HasPassword()
}
