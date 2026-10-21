/*
Package tokens signs and verifies the bunny.net-compatible URL tokens PhotoPrism embeds in its
header-less media URLs, so endpoints loaded through <img src> / <a href download> contexts (which
carry only a "?t=" query token and no Authorization header, possibly through a CDN) can be scoped
without a server-side lookup, and the signed authorization codes and refresh tokens of the OAuth app
flow, so a value without a valid signature is refused without a lookup.

Each token kind has a package-level Signer (see signer.go) whose key config.Propagate sets from the
instance signing key: Download and Preview share the key and signature path and are kept apart by a
signed kind parameter, while OAuth signs codes and refresh tokens with a purpose per kind (oauth.go).
The kind's own policy (per-session vs. per-path scope, sliding vs. bucketed expiry, compact vs. path
encoding) lives in its thin per-kind sign and verify wrapper. This is a Propagate-configured leaf like
thumb/dl/ttl and imports neither config nor get.

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
