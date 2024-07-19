/*
Package api provides REST-API authentication and request handlers.

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
package api

import (
	_ "net/http"

	_ "github.com/gin-gonic/gin"

	_ "github.com/photoprism/photoprism/internal/auth/acl"
	_ "github.com/photoprism/photoprism/internal/entity"
	_ "github.com/photoprism/photoprism/internal/entity/query"
	_ "github.com/photoprism/photoprism/internal/event"
	_ "github.com/photoprism/photoprism/internal/form"
	_ "github.com/photoprism/photoprism/internal/photoprism"
	_ "github.com/photoprism/photoprism/internal/photoprism/get"
	_ "github.com/photoprism/photoprism/pkg/clean"
	_ "github.com/photoprism/photoprism/pkg/fs"
	_ "github.com/photoprism/photoprism/pkg/i18n"
)

//	@title						PhotoPrism API
//	@description				API request bodies and responses are usually JSON-encoded, except for binary data and some of the OAuth2 endpoints. Note that the `Content-Type` header must be set to `application/json` for this, as the request may otherwise fail with error 400.
//	@description				When clients have a valid access token, e.g. obtained through the `POST /api/v1/session` or `POST /api/v1/oauth/token` endpoint, they can use a standard Bearer Authorization header to authenticate their requests. Submitting the access token with a custom `X-Auth-Token` header is supported as well.
//	@externalDocs.description	Learn more ›
//	@externalDocs.url			https://docs.photoprism.app/developer-guide/api/
//	@version					v1
//	@host						demo.photoprism.app
//	@query.collection.format	multi
