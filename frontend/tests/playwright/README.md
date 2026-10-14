# PhotoPrism Export Browser Checks

**Last Updated:** October 14, 2026

Run `node frontend/tests/playwright/export-images.mjs` and then
`node frontend/tests/playwright/export-controls.mjs` from the repository root
against an isolated development library. The runner uses Playwright, Chromium,
and Python Pillow to verify real downloaded bytes, filenames, and JPEG dimensions
at 1280×900 and 375×667. The image matrix exercises multiple outputs per original, manual links, single
Escape closure, and focus return. The controls runner checks Tab, Shift+Tab,
Space, and Enter, cancellation, permission loss, photo changes, Lightbox closure,
and the existing Download menu, Ctrl/Command+D, and RAW file filter. It changes
client settings only within its disposable browser contexts. The instance must
permit fresh browser contexts to open the library with unrestricted file access
and use English UI; neither runner authenticates or loads saved session state.
Download must be disabled and Export enabled in the test instance.
Enable `PHOTOPRISM_THUMB_UNCACHED=true` to
exercise advertised sizes beyond the pre-cached limit.

The default fixture root is `.local/export-flow`; override it with
`PHOTOPRISM_EXPORT_FIXTURE_PATH`. Override the server URL with
`PHOTOPRISM_TEST_URL` and the Chromium executable with
`PHOTOPRISM_BROWSER_PATH`. The image runner saves verified downloads and viewport
screenshots under the fixture root's `downloads/` directory; the controls runner
checks download events without retaining files. Both runners close their browser
contexts and tabs even when an assertion fails.

The fixture root must contain `originals/`, `sidecar/`, and `cases.json`. Each
manifest entry supplies `label` and the real photo `uid`. Required labels are
`JPEG`, `PNG`, `RAW`, `HEIC`, `SVG`, `PDF`, `TwoJPEG`, `RAWPlusJPEG`, `TwoRAW`,
`SameBasename`, and `AmbiguousPreview`. JPEG/PNG and large previews use 4000×3000
sources; RAW preview cases use smaller images with no surviving Full HD choice.
The ambiguous stack contains two RAW originals and a JPEG sharing their stem.
The case-insensitive filename stack contains two JPEG originals with the same
basename. SVG requires an existing JPG/PNG preview; the check does not require
an SVG ingest converter. File records and source files must already exist in
the isolated instance; this runner does not modify its database or production
media.

## Media Client Contract Checks

`node frontend/tests/playwright/export-media-client.mjs` exercises the media dialog
at both viewports against the isolated image library. It intercepts photo-detail
and planned probe/prepare requests with fixture responses. A temporary loopback
HTTP proxy serves fixture attachments/thumbnails and forwards other requests to
the isolated application; browser download requests are not intercepted.
It generates real AVC/HEVC video and GIF/APNG/WebP animation bytes using FFmpeg
and Pillow, then verifies browser attachments, bytes, dimensions, and `ffprobe`
codec/container results. Cases cover AVC, HEVC, Apple JPEG/HEIC pairs, embedded
HEVC motion photos, animations, mixed stacks, delayed preparation, and Retry.
Cancel/Escape during preparation must ignore late responses, and reopening must
discover the completed fixture output with the initial selection restored.
These checks validate the client contract and rendering; they do not verify
application conversion, native segment delivery, sidecar publication, or server authorization.

The default output root is `.local/export-media-client`; override it with
`PHOTOPRISM_EXPORT_CLIENT_PATH`. The same URL/browser/library environment variables
apply. Requires FFmpeg with `libx264`/`libx265`, `ffprobe`, Pillow, and the isolated
library's JPEG/HEIC originals. The runner closes contexts, the browser, and its
temporary HTTP proxy in `finally` blocks. These checks use fixture API responses;
real backend acceptance uses the separate runner below.

## Real Media Backend Checks

`node frontend/tests/playwright/export-media-backend.mjs` uses real indexed photo
details, capability probes, preparation, and attachment downloads without API
fixtures. It checks AVC deduplication, HEVC conversion, Apple JPEG/HEIC pairs,
native motion-video bytes, GIF/APNG/WebP animation, and mixed stacks at both
viewports. Original/preview hashes, actual AVC codecs, and native MOV bytes are
verified after saving downloads. Browser contexts close in `finally` blocks.
Retry and cancellation checks abort one network request or delay a real server
response; capability JSON and prepared bytes still come from the application.
Completed exports and Retry use the always-visible title Close control and verify
return to Lightbox. Preparation cancellation uses desktop Cancel or mobile title
Close and verifies that late responses do not hand off downloads.

Provide an isolated library with the same access/UI requirements as the image
runner and a manifest via `PHOTOPRISM_EXPORT_BACKEND_MANIFEST` (default
`.local/export-backend/cases.json`). Entries have `label`, a real indexed photo
`uid`, and file descriptions. Required labels are `HEVC`, `AVC`, `AppleJPEG`,
`AppleHEIC`, `Motion`, `Animationgif`, `Animationpng`, `Animationwebp`, and `Mixed`.
The fixture library contains `originals/backend-hevc.mov`, the exact native video
embedded in the Motion fixture. `PHOTOPRISM_EXPORT_BACKEND_PATH` selects the saved
download directory. Other library/URL/browser environment variables match the
image runner. The runner does not create or modify database records.

## Export Layout & Themes

`node frontend/tests/playwright/export-ui.mjs` checks the 560 px desktop dialog
and fullscreen mobile layout against an isolated image library at both viewports.
It intercepts photo details, capability probes, and header thumbnails with UI
fixtures. All built-in selectable themes and the shared custom dark/light theme
fixtures are exercised from inside Lightbox: the Export surface must use the
selected application theme. Checks cover long Unicode filenames, wrapping, fixed
header/actions, checkbox keyboard selection, Escape focus return, separate Apple
photo/MOV groups, and RAW/JPEG/SVG/PDF/AVC/HEVC/animation/motion/mixed layouts.
This runner validates client rendering; media conversion and download behavior
are covered by the runners above.

It uses the same URL/browser/library environment variables and requires the
`JPEG` manifest entry plus `originals/single.jpg`. JPEG details are replaced only
inside disposable browser contexts; the database is unchanged. Screenshots and
measurements are saved under `.local/screenshots/export-ui`, overridable with
`PHOTOPRISM_EXPORT_UI_PATH`. Contexts, pages, and Chromium close on failure.
