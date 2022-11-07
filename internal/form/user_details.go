package form

// UserDetails represents a user details form.
type UserDetails struct {
	BirthYear    int    `json:"BirthYear"`
	BirthMonth   int    `json:"BirthMonth"`
	BirthDay     int    `json:"BirthDay"`
	NameTitle    string `json:"NameTitle"`
	GivenName    string `json:"GivenName"`
	MiddleName   string `json:"MiddleName"`
	FamilyName   string `json:"FamilyName"`
	NameSuffix   string `json:"NameSuffix"`
	NickName     string `json:"NickName"`
	NameSrc      string `json:"NameSrc"`
	UserGender   string `json:"Gender"`
	UserAbout    string `json:"About"`
	UserBio      string `json:"Bio"`
	UserLocation string `json:"Location"`
	UserCountry  string `json:"Country"`
	UserPhone    string `json:"Phone"`
	SiteURL      string `json:"SiteURL"`
	ProfileURL   string `json:"ProfileURL"`
	FeedURL      string `json:"FeedURL"`
	OrgTitle     string `json:"OrgTitle"`
	OrgName      string `json:"OrgName"`
	OrgEmail     string `json:"OrgEmail"`
	OrgPhone     string `json:"OrgPhone"`
	OrgURL       string `json:"OrgURL"`
}
