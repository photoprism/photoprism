package pro

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/pro/places"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
	"gopkg.in/yaml.v2"
)

// Config represents photoprism.pro api credentials for maps & geodata.
type Config struct {
	Key     string `json:"key" yaml:"key"`
	Secret  string `json:"secret" yaml:"secret"`
	Session string `json:"session" yaml:"session"`
	Status  string `json:"status" yaml:"status"`
	Version string `json:"version" yaml:"version"`
}

// NewConfig creates a new photoprism.pro api credentials instance.
func NewConfig(version string) *Config {
	return &Config{
		Key:     "",
		Secret:  "",
		Session: "",
		Status:  "",
		Version: version,
	}
}

// Propagate updates photoprism.pro api credentials in other packages.
func (p *Config) Propagate() {
	places.Key = p.Key
	places.Secret = p.Secret
}

// Sanitize verifies and sanitizes photoprism.pro api credentials.
func (p *Config) Sanitize() {
	p.Key = strings.ToLower(p.Key)

	if p.Secret != "" {
		if p.Key != fmt.Sprintf("%x", sha1.Sum([]byte(p.Secret))) {
			p.Key = ""
			p.Secret = ""
			p.Session = ""
			p.Status = ""
		}
	}
}

// DecodeSession decodes photoprism.pro api session data.
func (p *Config) DecodeSession() (Session, error) {
	p.Sanitize()

	result := Session{}

	if p.Session == "" {
		return result, fmt.Errorf("empty session")
	}

	s, err := hex.DecodeString(p.Session)

	if err != nil {
		return result, err
	}

	hash := sha256.New()
	hash.Write([]byte(p.Secret))

	var b []byte

	block, err := aes.NewCipher(hash.Sum(b))

	if err != nil {
		return result, err
	}

	iv := s[:aes.BlockSize]

	plaintext := make([]byte, len(s))

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, s[aes.BlockSize:])

	plaintext = bytes.Trim(plaintext, "\x00")

	if err := json.Unmarshal(plaintext, &result); err != nil {
		return result, err
	}

	return result, nil
}

// Refresh updates photoprism.pro api credentials.
func (p *Config) Refresh() (err error) {
	p.Sanitize()
	client := &http.Client{Timeout: 60 * time.Second}
	url := ApiURL
	method := http.MethodPost
	var req *http.Request

	if p.Key != "" {
		url = fmt.Sprintf(ApiURL+"/%s", p.Key)
		method = http.MethodPut
		log.Debugf("pro: updating api key for maps & places")
	} else {
		log.Debugf("pro: requesting api key for maps & places")
	}

	if j, err := json.Marshal(NewRequest(p.Version)); err != nil {
		return err
	} else if req, err = http.NewRequest(method, url, bytes.NewReader(j)); err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	var r *http.Response

	for i := 0; i < 3; i++ {
		r, err = client.Do(req)

		if err == nil {
			break
		}
	}

	if err != nil {
		log.Errorf("pro: %s", err.Error())
		return err
	} else if r.StatusCode >= 400 {
		err = fmt.Errorf("api key request for maps & places failed with code %d", r.StatusCode)
		return err
	}

	err = json.NewDecoder(r.Body).Decode(p)

	if err != nil {
		log.Errorf("pro: %s", err.Error())
		return err
	}

	return nil
}

// Load photoprism.pro api credentials from a YAML file.
func (p *Config) Load(fileName string) error {
	if !fs.FileExists(fileName) {
		return fmt.Errorf("api key file not found: %s", txt.Quote(fileName))
	}

	yamlConfig, err := ioutil.ReadFile(fileName)

	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(yamlConfig, p); err != nil {
		return err
	}

	p.Sanitize()
	p.Propagate()

	return nil
}

// Save photoprism.pro api credentials to a YAML file.
func (p *Config) Save(fileName string) error {
	p.Sanitize()

	data, err := yaml.Marshal(p)

	if err != nil {
		return err
	}

	p.Propagate()

	if err := ioutil.WriteFile(fileName, data, os.ModePerm); err != nil {
		return err
	}

	p.Propagate()

	return nil
}
