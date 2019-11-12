package config

type Settings struct {
	Theme    string `json:"theme" yaml:"theme" flag:"theme"`
	Language string `json:"language" yaml:"language" flag:"language"`
}
