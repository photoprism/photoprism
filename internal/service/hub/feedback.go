package hub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/txt"
)

// FeedbackURL is the service endpoint for submitting user feedback.
var FeedbackURL = ServiceURL + "/%s/feedback"

// Feedback represents user feedback submitted through the user interface.
type Feedback struct {
	Category      string `json:"Category"`
	Subject       string `json:"Subject"`
	Message       string `json:"Message"`
	UserName      string `json:"UserName"`
	UserEmail     string `json:"UserEmail"`
	UserAgent     string `json:"UserAgent"`
	ApiKey        string `json:"ApiKey"`
	ClientVersion string `json:"ClientVersion"`
	ClientSerial  string `json:"ClientSerial"`
	ClientOS      string `json:"ClientOS"`
	ClientArch    string `json:"ClientArch"`
	ClientCPU     int    `json:"ClientCPU"`
	ClientEnv     string `json:"ClientEnv"`
	PartnerID     string `json:"PartnerID"`
}

// NewFeedback creates a new hub feedback instance.
func NewFeedback(version, serial, env, partnerId string) *Feedback {
	return &Feedback{
		ClientVersion: version,
		ClientSerial:  serial,
		ClientOS:      runtime.GOOS,
		ClientArch:    runtime.GOARCH,
		ClientCPU:     runtime.NumCPU(),
		ClientEnv:     env,
		PartnerID:     partnerId,
	}
}

// SendFeedback sends a feedback message.
func (c *Config) SendFeedback(f form.Feedback) (err error) {
	feedback := NewFeedback(c.Version, c.Serial, c.Env, c.PartnerID)
	feedback.Category = f.Category
	feedback.Subject = txt.Shorten(f.Message, 50, "...")
	feedback.Message = f.Message
	feedback.UserName = f.UserName
	feedback.UserEmail = f.UserEmail
	feedback.UserAgent = f.UserAgent
	feedback.ApiKey = c.Key

	// Create new http.Client instance.
	//
	// NOTE: Timeout specifies a time limit for requests made by
	// this Client. The timeout includes connection time, any
	// redirects, and reading the response body. The timer remains
	// running after Get, Head, Post, or Do return and will
	// interrupt reading of the Response.Body.
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

	// Set user agent.
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	} else {
		req.Header.Set("User-Agent", "PhotoPrism/Test")
	}

	req.Header.Add("Accept-Language", f.UserLocales)
	req.Header.Add(header.ContentType, header.ContentTypeJson)

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
