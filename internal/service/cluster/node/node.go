/*
Package node bootstraps a PhotoPrism node (app or service) that joins a cluster Portal.

Responsibilities include:

  - Initializing runtime configuration derived from options and environment.
  - Registering the node with the Portal over HTTP(S), handling non-transient
    errors (401/403/404) as terminal and bounding retries on transient failures.
  - Persisting returned registration data such as the Node secret and, when
    appropriate, database settings (never for SQLite), by merging into options.yml.
  - Installing a theme from the Portal if the local theme is missing, without
    overwriting existing installations.

The package deliberately avoids importing Portal internals and communicates
with the Portal using HTTP-only APIs.

Copyright (c) 2018 - 2025 PhotoPrism UG. All rights reserved.

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
package node
