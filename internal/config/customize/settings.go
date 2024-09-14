package customize

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
)

const (
	RootPath = "/"
)

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

// NewDefaultSettings creates a new default Settings instance.
func NewDefaultSettings() *Settings {
	return NewSettings(DefaultTheme, DefaultLocale)
}

// NewSettings creates a new Settings instance.
func NewSettings(theme, lang string) *Settings {
	return &Settings{
		UI: UISettings{
			Scrollbar: true,
			Zoom:      false,
			Theme:     theme,
			Language:  lang,
		},
		Search: SearchSettings{
			BatchSize: 0,
			ListView:  true,
		},
		Maps: MapsSettings{
			Animate: 0,
			Style:   "",
		},
		Features: FeatureSettings{
			Favorites: true,
			Reactions: true,
			Ratings:   true,
			Upload:    true,
			Download:  true,
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
			Archive:   true,
			Review:    true,
			Share:     true,
			Library:   true,
			Import:    true,
			Logs:      true,
			Search:    true,
			Settings:  true,
			Services:  true,
			Account:   true,
			Delete:    true,
		},
		Import: ImportSettings{
			Path: RootPath,
			Move: false,
		},
		Index: IndexSettings{
			Path:    RootPath,
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
		Download: NewDownloadSettings(),
		Templates: TemplateSettings{
			Default: "index.gohtml",
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
	if fileName == "" {
		return fmt.Errorf("no settings filename provided")
	} else if !fs.FileExists(fileName) {
		return fmt.Errorf("settings file not found: %s", clean.Log(fileName))
	}

	yamlConfig, err := os.ReadFile(fileName)

	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(yamlConfig, s); err != nil {
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

	if err = os.WriteFile(fileName, data, fs.ModeFile); err != nil {
		return err
	}

	return nil
}
