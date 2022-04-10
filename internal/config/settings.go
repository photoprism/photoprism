package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// UISettings represents user interface settings.
type UISettings struct {
	Scrollbar bool   `json:"scrollbar" yaml:"Scrollbar"`
	Zoom      bool   `json:"zoom" yaml:"Zoom"`
	Theme     string `json:"theme" yaml:"Theme"`
	Language  string `json:"language" yaml:"Language"`
}

// SearchSettings represents search UI preferences.
type SearchSettings struct {
	BatchSize int `json:"batchSize" yaml:"BatchSize"`
}

// TemplateSettings represents template settings for the UI and messaging.
type TemplateSettings struct {
	Default string `json:"default" yaml:"Default"`
}

// MapsSettings represents maps settings (for places).
type MapsSettings struct {
	Animate int    `json:"animate" yaml:"Animate"`
	Style   string `json:"style" yaml:"Style"`
}

// FeatureSettings represents feature flags, mainly for the Web UI.
type FeatureSettings struct {
	Upload    bool `json:"upload" yaml:"Upload"`
	Download  bool `json:"download" yaml:"Download"`
	Private   bool `json:"private" yaml:"Private"`
	Review    bool `json:"review" yaml:"Review"`
	Files     bool `json:"files" yaml:"Files"`
	Videos    bool `json:"videos" yaml:"Videos"`
	Folders   bool `json:"folders" yaml:"Folders"`
	Albums    bool `json:"albums" yaml:"Albums"`
	Moments   bool `json:"moments" yaml:"Moments"`
	Estimates bool `json:"estimates" yaml:"Estimates"`
	People    bool `json:"people" yaml:"People"`
	Labels    bool `json:"labels" yaml:"Labels"`
	Places    bool `json:"places" yaml:"Places"`
	Edit      bool `json:"edit" yaml:"Edit"`
	Archive   bool `json:"archive" yaml:"Archive"`
	Delete    bool `json:"delete" yaml:"Delete"`
	Share     bool `json:"share" yaml:"Share"`
	Library   bool `json:"library" yaml:"Library"`
	Import    bool `json:"import" yaml:"Import"`
	Logs      bool `json:"logs" yaml:"Logs"`
}

// ImportSettings represents import settings.
type ImportSettings struct {
	Path string `json:"path" yaml:"Path"`
	Move bool   `json:"move" yaml:"Move"`
}

// IndexSettings represents indexing settings.
type IndexSettings struct {
	Path    string `json:"path" yaml:"Path"`
	Convert bool   `json:"convert" yaml:"Convert"`
	Rescan  bool   `json:"rescan" yaml:"Rescan"`
}

// StackSettings represents settings for files that belong to the same photo.
type StackSettings struct {
	UUID bool `json:"uuid" yaml:"UUID"`
	Meta bool `json:"meta" yaml:"Meta"`
	Name bool `json:"name" yaml:"Name"`
}

// ShareSettings represents content sharing settings.
type ShareSettings struct {
	Title string `json:"title" yaml:"Title"`
}

// DownloadSettings represents content download settings.
type DownloadSettings struct {
	Name entity.DownloadName `json:"name" yaml:"Name"`
	Raw  bool                `json:"raw" yaml:"Raw"`
}

// Settings represents user settings for Web UI, indexing, and import.
type Settings struct {
	UI        UISettings       `json:"ui" yaml:"UI"`
	Search    SearchSettings   `json:"search" yaml:"Search"`
	Maps      MapsSettings     `json:"maps" yaml:"Maps"`
	Features  FeatureSettings  `json:"features" yaml:"Features"`
	Import    ImportSettings   `json:"import" yaml:"Import"`
	Index     IndexSettings    `json:"index" yaml:"Index"`
	Stack     StackSettings    `json:"stack" yaml:"Stack"`
	Share     ShareSettings    `json:"share" yaml:"Share"`
	Download  DownloadSettings `json:"download" yaml:"Download"`
	Templates TemplateSettings `json:"templates" yaml:"Templates"`
}

// NewSettings creates a new Settings instance.
func NewSettings(c *Config) *Settings {
	return &Settings{
		UI: UISettings{
			Scrollbar: true,
			Zoom:      false,
			Theme:     c.DefaultTheme(),
			Language:  c.DefaultLocale(),
		},
		Search: SearchSettings{
			BatchSize: 0,
		},
		Maps: MapsSettings{
			Animate: 0,
			Style:   "streets",
		},
		Features: FeatureSettings{
			Upload:    true,
			Download:  true,
			Archive:   true,
			Review:    true,
			Private:   true,
			Files:     true,
			Videos:    true,
			Folders:   true,
			Albums:    true,
			Moments:   true,
			Estimates: true,
			People:    true,
			Labels:    true,
			Places:    true,
			Edit:      true,
			Share:     true,
			Library:   true,
			Import:    true,
			Logs:      true,
		},
		Import: ImportSettings{
			Path: entity.RootPath,
			Move: false,
		},
		Index: IndexSettings{
			Path:    entity.RootPath,
			Rescan:  false,
			Convert: true,
		},
		Stack: StackSettings{
			UUID: true,
			Meta: true,
			Name: false,
		},
		Share: ShareSettings{
			Title: "",
		},
		Download: DownloadSettings{
			Name: entity.DownloadNameDefault,
		},
		Templates: TemplateSettings{
			Default: "index.tmpl",
		},
	}
}

// Propagate updates settings in other packages as needed.
func (s *Settings) Propagate() {
	i18n.SetLocale(s.UI.Language)
}

// StackSequences checks if files should be stacked based on their file name prefix (sequential names).
func (s Settings) StackSequences() bool {
	return s.Stack.Name
}

// StackUUID checks if files should be stacked based on unique image or instance id.
func (s Settings) StackUUID() bool {
	return s.Stack.UUID
}

// StackMeta checks if files should be stacked based on their place and time metadata.
func (s Settings) StackMeta() bool {
	return s.Stack.Meta
}

// Load user settings from file.
func (s *Settings) Load(fileName string) error {
	if !fs.FileExists(fileName) {
		return fmt.Errorf("settings file not found: %s", sanitize.Log(fileName))
	}

	yamlConfig, err := os.ReadFile(fileName)

	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(yamlConfig, s); err != nil {
		return err
	}

	s.Propagate()

	return nil
}

// Save user settings to a file.
func (s *Settings) Save(fileName string) error {
	data, err := yaml.Marshal(s)

	if err != nil {
		return err
	}

	s.Propagate()

	if err := os.WriteFile(fileName, data, os.ModePerm); err != nil {
		return err
	}

	return nil
}

// initSettings initializes user settings from a config file.
func (c *Config) initSettings() {
	if c.settings != nil {
		return
	}

	c.settings = NewSettings(c)
	fileName := c.SettingsFile()

	if err := c.settings.Load(fileName); err == nil {
		log.Debugf("settings: loaded from %s ", fileName)
	} else if err := c.settings.Save(fileName); err != nil {
		log.Errorf("settings: could not create %s (%s)", fileName, err)
	} else {
		log.Debugf("settings: saved to %s ", fileName)
	}

	i18n.SetDir(c.LocalesPath())

	c.settings.Propagate()
}

// Settings returns the current user settings.
func (c *Config) Settings() *Settings {
	c.initSettings()

	if c.DisablePlaces() {
		c.settings.Features.Places = false
	}

	return c.settings
}
