/*
Package nsfw provides detection of images that are "not safe for work" based on various categories.

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark/>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package nsfw

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

// Sentinel errors distinguish an unconfigured detector from one that failed.
var (
	ErrNotConfigured       = errors.New("no nsfw detector configured")
	ErrDetectorUnavailable = errors.New("nsfw detector unavailable")
)

// StatusUnavailableName is how the undecided status renders as text.
const StatusUnavailableName = "unavailable"

// DefaultThreshold is the conservative fallback used until ONNX corpus calibration is complete.
const DefaultThreshold float32 = 0.98

// UploadThreshold preserves the established upload-screening operating point.
const UploadThreshold float32 = 0.75

// Status is the three-valued outcome of an NSFW check.
// Its zero value is unavailable so an unfilled result is never a clearance.
type Status string

const (
	// StatusUnavailable means no detector produced a decision for this image.
	StatusUnavailable Status = ""
	// StatusSafe means a detector scored the image below the applied threshold.
	StatusSafe Status = "safe"
	// StatusUnsafe means a detector scored the image at or above the applied threshold.
	StatusUnsafe Status = "unsafe"
)

// String returns the status name, rendering the zero value as "unavailable".
func (s Status) String() string {
	if s == StatusUnavailable {
		return StatusUnavailableName
	}

	return string(s)
}

// MarshalJSON writes the status name so an undecided result is explicit on the wire.
func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON reads a status name and maps anything unrecognized to StatusUnavailable, so a
// value from a newer or foreign service is never read as a clearance.
func (s *Status) UnmarshalJSON(b []byte) error {
	var name string

	if err := json.Unmarshal(b, &name); err != nil {
		return err
	}

	switch Status(name) {
	case StatusSafe:
		*s = StatusSafe
	case StatusUnsafe:
		*s = StatusUnsafe
	default:
		*s = StatusUnavailable
	}

	return nil
}

// Result is the outcome of one NSFW check.
type Result struct {
	// Status is the decision, and the only field a caller may act on. The tag spells the wire
	// values out because the zero value serializes as "unavailable" rather than as itself.
	Status Status `yaml:"Status" json:"status" swaggertype:"string" enums:"safe,unsafe,unavailable"`
	// Score is the probability that the image is unsafe, from 0 to 1.
	Score float32 `yaml:"Score,omitempty" json:"score,omitempty"`
	// Threshold is the value Score was compared against, recorded so a decision can be
	// explained, or re-derived by a caller, without running inference again.
	Threshold float32 `yaml:"Threshold,omitempty" json:"threshold,omitempty"`
	// Reason names why no decision was made, and is empty otherwise.
	Reason string `yaml:"Reason,omitempty" json:"reason,omitempty"`

	// Class probabilities used by legacy and remote models. Detectors with a binary taxonomy
	// leave these zero, so a client must read Status. The JSON names are capitalized because
	// that is what this endpoint has always emitted for these untagged fields.
	Drawing float32 `yaml:"Drawing,omitempty" json:"Drawing"`
	Hentai  float32 `yaml:"Hentai,omitempty" json:"Hentai"`
	Neutral float32 `yaml:"Neutral,omitempty" json:"Neutral"`
	Porn    float32 `yaml:"Porn,omitempty" json:"Porn"`
	Sexy    float32 `yaml:"Sexy,omitempty" json:"Sexy"`
}

// Unavailable returns a Result recording that no decision could be made, and why.
func Unavailable(reason string) Result {
	return Result{Status: StatusUnavailable, Reason: reason}
}

// NewResult returns a decided Result for an unsafe probability and threshold.
func NewResult(score, threshold float32) Result {
	if err := ValidateScore(score); err != nil {
		return Unavailable(err.Error())
	}

	result := Result{Score: score, Threshold: threshold}

	if score >= threshold {
		result.Status = StatusUnsafe
	} else {
		result.Status = StatusSafe
	}

	return result
}

// Decide returns a copy of r decided against threshold.
// A result without class scores stays unavailable.
func (r Result) Decide(threshold float32) Result {
	if !r.HasScores() {
		if r.Status == StatusUnavailable && r.Reason == "" {
			r.Reason = "no scores"
		}

		return r
	}

	decided := NewResult(r.UnsafeScore(), threshold)
	decided.Drawing, decided.Hentai = r.Drawing, r.Hentai
	decided.Neutral, decided.Porn, decided.Sexy = r.Neutral, r.Porn, r.Sexy

	return decided
}

// UnsafeScore returns the highest unsafe class probability without a neutral-class veto.
func (r Result) UnsafeScore() float32 {
	return max(r.Porn, r.Sexy, r.Hentai)
}

// HasScores reports whether the class probabilities carry any signal.
func (r Result) HasScores() bool {
	return r.Drawing > 0 || r.Hentai > 0 || r.Neutral > 0 || r.Porn > 0 || r.Sexy > 0
}

// IsSafe reports whether a detector decided the image is safe for work.
func (r Result) IsSafe() bool {
	return r.Status == StatusSafe
}

// IsUnsafe reports whether a detector decided the image is not safe for work.
func (r Result) IsUnsafe() bool {
	return r.Status == StatusUnsafe
}

// IsUnavailable reports whether no detector produced a decision.
//
// The zero value reports true, so a result no detector filled in is never a clearance.
func (r Result) IsUnavailable() bool {
	return r.Status == StatusUnavailable
}

// ValidateScore rejects an unsafe probability that is not a finite value from 0 to 1.
func ValidateScore(score float32) error {
	value := float64(score)

	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("model returned a non-finite score")
	} else if value < 0 || value > 1 {
		return fmt.Errorf("model returned score %.4f, expected 0 to 1", value)
	}

	return nil
}
