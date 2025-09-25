/*
Package provisioner manages per-node database provisioning for cluster setups.

It runs on the Portal and is responsible for:

  - Generating deterministic database and user names for nodes based on the
    portal's Cluster UUID and the node name, while keeping lengths within
    engine limits.
  - Creating the database schema if missing and granting minimal privileges
    to the node's database user.
  - Creating the user if needed and rotating its password on demand, returning
    credentials (and a ready-to-use DSN) to the caller.

The implementation uses GORM v1 for execution and binds parameters safely
instead of string formatting for values, while quoting identifiers where
necessary for portability and security.

Admin DSN: Provisioning connects via a dedicated admin DSN (mysql driver),
independent of the application's main database/driver. This follows least
privilege best practices and allows the app DB to be SQLite while provisioning
against MariaDB/MySQL.

Testing notes: Package includes a lightweight MariaDB integration test that
uses the Admin DSN and skips automatically if the DSN cannot be opened/pinged.
Historically, behavior was also validated via Docker Compose and broader repo
targets such as "make run-test-mariadb".

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
package provisioner

import "github.com/photoprism/photoprism/internal/event"

var log = event.Log
