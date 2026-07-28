/*
Package tokens signs and verifies the bunny.net-compatible URL tokens PhotoPrism embeds in its
header-less media URLs, so endpoints loaded through <img src> / <a href download> contexts (which
carry only a "?t=" query token and no Authorization header, possibly through a CDN) can be scoped
without a server-side lookup.

Each token kind is a package-level Signer (see signer.go) configured by config.Propagate with its own
key and signature path — downloads today (see download.go), thumbnail and video previews next. The
kind's own policy (per-session vs. per-path scope, sliding vs. bucketed expiry, compact vs. path
encoding) lives in its thin per-kind sign and verify wrapper, so a new kind slots in beside Download
without a second package. This is a Propagate-configured leaf like thumb/dl/ttl and imports neither
config nor get.

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
package tokens
