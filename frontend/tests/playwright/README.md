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

Create disposable fixtures from the repository root, with the server stopped:

```sh
python3 frontend/tests/playwright/export-fixtures.py --root .local/export-backend
```

Requires a development `photoprism` binary, Go, FFmpeg with `libx264`/`libx265`,
`ffprobe`, Python Pillow, and the tracked DNG/HEIC samples. Use `--binary` for a
side-by-side edition build and `--edition ce|plus|pro` for its real role set.
Paid role setup requires a registered edition; unsupported roles fail setup.
The script confines all generated media/config/SQLite data to a marked directory
under `.local/`, indexes a seed, and installs file/stack fixture rows in that
isolated database. Repeating it replaces owned rows, preserving photo UIDs and
file counts. It creates real password accounts and stores the generated test
password only in local `password.env` (mode 0600). Outputs are never committed.

`environment.json` supplies the isolated paths and public-auth environment for
the media server. Apply its values to the server process, set a free loopback
`PHOTOPRISM_HTTP_PORT` and matching `PHOTOPRISM_SITE_URL`, and start `photoprism`.
For example, start the disposable media server in a separate terminal:

```sh
python3 - <<'PYTHON'
import json, os, subprocess
env = os.environ.copy()
env.update(json.load(open(".local/export-backend/environment.json")))
env.update(PHOTOPRISM_HTTP_PORT="2344", PHOTOPRISM_SITE_URL="http://localhost:2344/")
subprocess.run(["./photoprism", "start"], env=env, check=True)
PYTHON
```

Wait for `/readyz`, then run the backend check with `PHOTOPRISM_TEST_URL` set to that URL. The backend
runner defaults to `.local/export-backend` for both library and `cases.json`;
override them with `PHOTOPRISM_EXPORT_FIXTURE_PATH` and
`PHOTOPRISM_EXPORT_BACKEND_MANIFEST`. `PHOTOPRISM_EXPORT_BACKEND_PATH` selects the
saved-download directory. Use English UI, Export enabled, and uncached thumbnails.
The runner itself creates no database records.

Every required label is checked before opening Chromium: `HEVC`, `AVC`,
`AppleJPEG`, `AppleHEIC`, `Motion`, `Animationgif`, `Animationpng`, `Animationwebp`,
`Mixed`, `TwoImage`, `TwoRAW`, `TwoSVG`, `TwoPDF`, `TwoGIF`, `TwoAPNG`, `TwoWebP`,
`TwoVideo`, `TwoMotion`, `TwoApple`, and `RAWPlusJPEG`. Setup also creates a
single-image `RoleJPEG` fixture for role downloads. `TwoApple` contains two
image/MOV pairs (four originals); the other two-source stacks contain two
originals. Each manifest entry supplies `label`, a real photo `uid`, and `files`
with source name/root/format/SHA1. Motion entries also map original names to their
exact native movie filenames in `native`. Fixtures include both JPG and PNG
previews to exercise format preference. Missing fixtures, incomplete original
stacks, or absent expected choices fail rather than skip.

Every selected browser attachment must have a unique final name, the expected
real format, and original/preview SHA1 or decoded JPEG dimensions, video codec,
and container. Native motion bytes must match the independent movie source.
Retry also verifies the real AVC attachment. Both viewports run every case and
return to Lightbox after closure.

## Real Session & Role Checks

`node frontend/tests/playwright/export-roles.mjs` authenticates real accounts;
it does not replace ACLs or photo API responses. Run it separately, after stopping
the media server: retain the isolated paths from `environment.json`, override
`PHOTOPRISM_AUTH_MODE=password` after loading the environment file (add it to
the example's `env.update(...)` call), enable both Export and Download in the fixture
`config/settings.yml`, and restart. Load the local password environment without
printing it:

```sh
set -a
. .local/export-backend/password.env
set +a
node frontend/tests/playwright/export-roles.mjs
```

Supply the same URL/library/browser variables as the backend runner;
`PHOTOPRISM_EXPORT_ROLE_MANIFEST` overrides the default library `roles.json`.
CE requires admin, guest, and share-link sessions. Plus requires admin, user,
viewer, guest, and share-link. Pro requires admin, manager, user, viewer,
contributor, guest, and share-link. Accounts retain their requested role after
login. Only admin (CE), admin/user (Plus), and admin/manager/user (Pro) can Export.
Viewer/contributor/guest/share-link checks verify the missing menu action and
blocked shortcut, while existing Download menu and Ctrl+D still deliver the
expected source bytes. Each allowed role performs a real original Export too.
All cases run at both viewports. Download expectations are explicit in the
manifest and checked against the real session.

The runner creates a temporary share album/link through the authenticated admin
API and removes them in `finally`; run it only against the disposable fixture
instance. Passwords are read from named environment variables, and session tokens
are never printed. A missing account, role case, password, photo, or registered
Plus/Pro feature set fails explicitly. An unregistered instance cannot verify the
paid-edition matrix. Treat such a run as blocked verification, never as a pass.

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
