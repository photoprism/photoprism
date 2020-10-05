package pro

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt"
	"net/http"
	"runtime"
	"time"
)

var FeedbackURL = ApiURL + "/%s/feedback"

type Feedback struct {
	Category      string `json:"Category"`
	Subject       string `json:"Subject"`
	Message       string `json:"Message"`
	UserName      string `json:"UserName"`
	UserEmail     string `json:"UserEmail"`
	UserAgent     string `json:"UserAgent"`
	ApiKey        string `json:"ApiKey"`
	ClientVersion string `json:"ClientVersion"`
	ClientOS      string `json:"ClientOS"`
	ClientArch    string `json:"ClientArch"`
	ClientCPU     int    `json:"ClientCPU"`
}

// NewFeedback creates a new photoprism.pro key request instance.
func NewFeedback(version string) *Feedback {
	return &Feedback{
		ClientVersion: version,
		ClientOS:      runtime.GOOS,
		ClientArch:    runtime.GOARCH,
		ClientCPU:     runtime.NumCPU(),
	}
}

func (c *Config) SendFeedback(f form.Feedback) (err error) {
	feedback := NewFeedback(c.Version)
	feedback.Category = f.Category
	feedback.Subject = txt.TrimLen(f.Message, 50)
	feedback.Message = f.Message
	feedback.UserName = f.UserName
	feedback.UserEmail = f.UserEmail
	feedback.UserAgent = f.UserAgent
	feedback.ApiKey = c.Key

	client := &http.Client{Timeout: 60 * time.Second}
	url := fmt.Sprintf(FeedbackURL, c.Key)
	method := http.MethodPost
	var req *http.Request

	log.Debugf("pro: sending feedback")

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
		log.Errorf("pro: %s", err.Error())
		return err
	} else if r.StatusCode >= 400 {
		err = fmt.Errorf("sending feedback failed with code %d", r.StatusCode)
		return err
	}

	err = json.NewDecoder(r.Body).Decode(c)

	if err != nil {
		log.Errorf("pro: %s", err.Error())
		return err
	}

	return nil
}
