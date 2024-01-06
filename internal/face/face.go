/*
Package face provides face recognition.

Copyright (c) 2018 - 2024 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package face

import (
	"encoding/json"

	"github.com/photoprism/photoprism/internal/crop"
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

// Face represents a face detected.
type Face struct {
	Rows       int        `json:"rows,omitempty"`
	Cols       int        `json:"cols,omitempty"`
	Score      int        `json:"score,omitempty"`
	Area       Area       `json:"face,omitempty"`
	Eyes       Areas      `json:"eyes,omitempty"`
	Landmarks  Areas      `json:"landmarks,omitempty"`
	Embeddings Embeddings `json:"embeddings,omitempty"`
}

// Size returns the absolute face size in pixels.
func (f *Face) Size() int {
	return f.Area.Scale
}

// Dim returns the max number of rows and cols as float32 to calculate relative coordinates.
func (f *Face) Dim() float32 {
	if f.Cols > 0 {
		return float32(f.Cols)
	}

	return float32(1)
}

// CropArea returns the relative image area for cropping.
func (f *Face) CropArea() crop.Area {
	if f.Rows < 1 {
		f.Cols = 1
	}

	if f.Cols < 1 {
		f.Cols = 1
	}

	x := float32(f.Area.Col-f.Area.Scale/2) / float32(f.Cols)
	y := float32(f.Area.Row-f.Area.Scale/2) / float32(f.Rows)

	return crop.NewArea(
		f.Area.Name,
		x,
		y,
		float32(f.Area.Scale)/float32(f.Cols),
		float32(f.Area.Scale)/float32(f.Rows),
	)
}

// EyesMidpoint returns the point in between the eyes.
func (f *Face) EyesMidpoint() Area {
	if len(f.Eyes) != 2 {
		return Area{
			Name:  "midpoint",
			Row:   f.Area.Row,
			Col:   f.Area.Col,
			Scale: f.Area.Scale,
		}
	}

	return Area{
		Name:  "midpoint",
		Row:   (f.Eyes[0].Row + f.Eyes[1].Row) / 2,
		Col:   (f.Eyes[0].Col + f.Eyes[1].Col) / 2,
		Scale: (f.Eyes[0].Scale + f.Eyes[1].Scale) / 2,
	}
}

// RelativeLandmarks returns relative face areas.
func (f *Face) RelativeLandmarks() crop.Areas {
	p := f.EyesMidpoint()

	m := f.Landmarks.Relative(p, float32(f.Rows), float32(f.Cols))
	m = append(m, f.Eyes.Relative(p, float32(f.Rows), float32(f.Cols))...)

	return m
}

// RelativeLandmarksJSON returns relative face areas as JSON.
func (f *Face) RelativeLandmarksJSON() (b []byte) {
	var noResult = []byte("")

	l := f.RelativeLandmarks()

	if len(l) < 1 {
		return noResult
	}

	if result, err := json.Marshal(l); err != nil {
		log.Errorf("faces: %s", err)
		return noResult
	} else {
		return result
	}
}

// EmbeddingsJSON returns detected face embeddings as JSON array.
func (f *Face) EmbeddingsJSON() (b []byte) {
	return f.Embeddings.JSON()
}

// HasEmbedding tests if the face has at least one embedding.
func (f *Face) HasEmbedding() bool {
	return len(f.Embeddings) > 0
}

// NoEmbedding tests if the face has no embeddings.
func (f *Face) NoEmbedding() bool {
	if f.Embeddings == nil {
		return true
	}

	return f.Embeddings.Empty()
}
