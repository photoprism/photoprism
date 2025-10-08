/*
Package dl provides helpers to discover media metadata and download content
from remote sources via yt-dlp. It underpins the `photoprism dl` CLI.

Two download methods are supported:

 1. Pipe method (stdout): yt-dlp streams media to stdout and PhotoPrism
    writes it to a temporary file. After writing, PhotoPrism remuxes the
    file with ffmpeg to ensure a valid MP4 container and to embed basic
    metadata such as title, description, author, source URL (as comment),
    and creation timestamp when available. This method is simple and works
    well for sources that provide a combined A/V stream. Some sites split
    audio and video; in those cases yt-dlp cannot always mux when piping.

 2. File method (on-disk): yt-dlp writes output files directly using
    `--output` templates and built-in post-processors (merge/remux/metadata).
    PhotoPrism captures the final file paths (via `--print after_move:filepath`)
    and then optionally runs a final ffmpeg remux to normalize the container
    and embed metadata if necessary. This method is recommended for sources
    that deliver separate audio/video streams or require post-processing.

Both methods accept the same authentication-related options and headers.
Cookies can be supplied via a file or a browser profile, and custom headers
(e.g. Authorization) are forwarded to yt-dlp for both metadata discovery and
downloading. Secrets are not logged; header values are redacted in traces.

The package exposes convenience constructors around yt-dlp invocation as
well as small utilities for safer logging and remux metadata preparation.

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

This code is copied and modified in part from:

  - https://github.com/wader/goutubedl
    MIT License, Copyright (c) 2019 Mattias Wadman
    see https://github.com/wader/goutubedl?tab=MIT-1-ov-file#readme

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package dl

import (
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log
