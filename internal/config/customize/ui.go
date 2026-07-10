package customize

// UISettings represents user interface settings.
type UISettings struct {
	Scrollbar    bool   `json:"scrollbar" yaml:"Scrollbar"`
	Zoom         bool   `json:"zoom" yaml:"Zoom"`
	OpenOnHover  bool   `json:"openOnHover" yaml:"OpenOnHover"`
	ReduceMotion bool   `json:"reduceMotion" yaml:"ReduceMotion"`
	Theme        string `json:"theme" yaml:"Theme"`
	Language     string `json:"language" yaml:"Language"`
	TimeZone     string `json:"timeZone" yaml:"TimeZone"`
	StartPage    string `json:"startPage" yaml:"StartPage"`
}
