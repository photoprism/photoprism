package hub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt"
)

var FeedbackURL = ServiceURL + "/%s/feedback"

type Feedback struct {
	Category      string `json:"Category"`
	Subject       string `json:"Subject"`
	Message       string `json:"Message"`
	UserName      string `json:"UserName"`
	UserEmail     string `json:"UserEmail"`
	UserAgent     string `json:"UserAgent"`
	ApiKey        string `json:"ApiKey"`
	PartnerID     string `json:"PartnerID"`
	ClientVersion string `json:"ClientVersion"`
	ClientSerial  string `json:"ClientSerial"`
	ClientOS      string `json:"ClientOS"`
	ClientArch    string `json:"ClientArch"`
	ClientCPU     int    `json:"ClientCPU"`
}

// NewFeedback creates a new hub feedback instance.
func NewFeedback(version, serial, partner string) *Feedback {
	return &Feedback{
		PartnerID:     partner,
		ClientVersion: version,
		ClientSerial:  serial,
		ClientOS:      runtime.GOOS,
		ClientArch:    runtime.GOARCH,
		ClientCPU:     runtime.NumCPU(),
	}
}

func (c *Config) SendFeedback(f form.Feedback) (err error) {
	feedback := NewFeedback(c.Version, c.Serial, c.PartnerID)
	feedback.Category = f.Category
	feedback.Subject = txt.Shorten(f.Message, 50, "...")
	feedback.Message = f.Message
	feedback.UserName = f.UserName
	feedback.UserEmail = f.UserEmail
	feedback.UserAgent = f.UserAgent
	feedback.ApiKey = c.Key

	client := &http.Client{Timeout: 60 * time.Second}
	url := fmt.Sprintf(FeedbackURL, c.Key)
	method := http.MethodPost

	var req *http.Request

	log.Debugf("sending feedback to %s", ApiHost())

	if j, err := json.Marshal(feedback); err != nil {
		return err
	} else if req, err = http.NewRequest(method, url, bytes.NewReader(j)); err != nil {
		return err
	}

	req.Header.Add("Accept-Language", f.UserLocales)
	req.Header.Add("Content-Type", "application/json")

	var r *http.Response

	for i := 0; i < 3; i++ {
		r, err = client.Do(req)

		if err == nil {
			break
		}
	}

	if err != nil {
		return err
	} else if r.StatusCode >= 400 {
		err = fmt.Errorf("sending feedback to %s failed (error %d)", ApiHost(), r.StatusCode)
		return err
	}

	err = json.NewDecoder(r.Body).Decode(c)

	if err != nil {
		return err
	}

	return nil
}
