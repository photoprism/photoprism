package config

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/photoprism/photoprism/internal/maps/places"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
	"gopkg.in/yaml.v2"
)

// Credentials represents api credentials for hosted services like maps & places.
type Credentials struct {
	Key     string `json:"key" yaml:"key"`
	Secret  string `json:"secret" yaml:"secret"`
	Session string `json:"session" yaml:"session"`
}

// NewCredentials creates a new Credentials instance.
func NewCredentials() *Credentials {
	return &Credentials{
		Key:     "",
		Secret:  "",
		Session: "",
	}
}

// Propagate updates api credentials in other packages.
func (a *Credentials) Propagate() {
	places.Key = a.Key
}

// Sanitize verifies and sanitizes api credentials;
func (a *Credentials) Sanitize() {
	a.Key = strings.ToLower(a.Key)

	if a.Secret != "" {
		if a.Key != fmt.Sprintf("%x", sha1.Sum([]byte(a.Secret))) {
			a.Secret = ""
			a.Session = ""
		}
	}
}

// Load api credentials from a file.
func (a *Credentials) Load(fileName string) error {
	if !fs.FileExists(fileName) {
		return fmt.Errorf("credentials file not found: %s", txt.Quote(fileName))
	}

	yamlConfig, err := ioutil.ReadFile(fileName)

	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(yamlConfig, a); err != nil {
		return err
	}

	a.Sanitize()
	a.Propagate()

	return nil
}

// Save api credentials to a file.
func (a *Credentials) Save(fileName string) error {
	a.Sanitize()

	data, err := yaml.Marshal(a)

	if err != nil {
		return err
	}

	a.Propagate()

	if err := ioutil.WriteFile(fileName, data, os.ModePerm); err != nil {
		return err
	}

	a.Propagate()

	return nil
}

// initCredentials initializes the api credentials.
func (c *Config) initCredentials() {
	c.credentials = NewCredentials()
	p := c.CredentialsFile()

	if err := c.credentials.Load(p); err != nil {
		log.Traceln(err)
	}

	c.credentials.Propagate()
}

// Credentials returns the api key instance.
func (c *Config) Credentials() *Credentials {
	return c.credentials
}
