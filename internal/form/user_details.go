package form

// UserDetails represents a user details form.
type UserDetails struct {
	BirthYear    int    `json:"BirthYear"`
	BirthMonth   int    `json:"BirthMonth"`
	BirthDay     int    `json:"BirthDay"`
	NamePrefix   string `json:"NamePrefix"`
	GivenName    string `json:"GivenName"`
	MiddleName   string `json:"MiddleName"`
	FamilyName   string `json:"FamilyName"`
	NameSuffix   string `json:"NameSuffix"`
	NickName     string `json:"NickName"`
	UserGender   string `json:"Gender"`
	UserBio      string `json:"Bio"`
	UserLocation string `json:"Location"`
	UserCountry  string `json:"Country"`
	UserPhone    string `json:"Phone"`
	SiteURL      string `json:"SiteURL"`
	ProfileURL   string `json:"ProfileURL"`
	FeedURL      string `json:"FeedURL"`
	OrgName      string `json:"OrgName"`
	OrgTitle     string `json:"OrgTitle"`
	OrgEmail     string `json:"OrgEmail"`
	OrgPhone     string `json:"OrgPhone"`
	OrgURL       string `json:"OrgURL"`
}
