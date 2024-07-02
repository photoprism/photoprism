/*
Package get provides a registry for common application services.

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
package get

import (
	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/auth/oidc"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/session"

	gc "github.com/patrickmn/go-cache"
)

var conf *config.Config

var services struct {
	FolderCache *gc.Cache
	CoverCache  *gc.Cache
	ThumbCache  *gc.Cache
	Classify    *classify.TensorFlow
	Convert     *photoprism.Convert
	Files       *photoprism.Files
	Photos      *photoprism.Photos
	Import      *photoprism.Import
	Index       *photoprism.Index
	Moments     *photoprism.Moments
	Faces       *photoprism.Faces
	Places      *photoprism.Places
	Purge       *photoprism.Purge
	CleanUp     *photoprism.CleanUp
	Nsfw        *nsfw.Detector
	FaceNet     *face.Net
	Query       *query.Query
	Thumbs      *photoprism.Thumbs
	Session     *session.Session
	OIDC        *oidc.Client
}

func SetConfig(c *config.Config) {
	if c == nil {
		panic("config is nil")
	}

	conf = c

	photoprism.SetConfig(c)
}

func Config() *config.Config {
	if conf == nil {
		panic("config is nil")
	}

	return conf
}
