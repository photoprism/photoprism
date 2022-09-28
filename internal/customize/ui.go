package customize

// UISettings represents user interface settings.
type UISettings struct {
	Scrollbar bool   `json:"scrollbar" yaml:"Scrollbar"`
	Zoom      bool   `json:"zoom" yaml:"Zoom"`
	Theme     string `json:"theme" yaml:"Theme"`
	Language  string `json:"language" yaml:"Language"`
	TimeZone  string `json:"timeZone" yaml:"TimeZone"`
}
