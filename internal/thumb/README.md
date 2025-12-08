## PhotoPrism — Thumbnails Package

**Last Updated:** November 24, 2025

### Overview

`internal/thumb` builds thumbnails with libvips, handling resize/crop options, color management, metadata stripping, and format export (JPEG/PNG). It is used by PhotoPrism’s workers and CLI to generate cached thumbs consistently.

### Context & Constraints

- Uses libvips via govips; initialization is centralized in `VipsInit`.
- Works on files or in-memory buffers; writes outputs with `fs.ModeFile`.
- ICC handling: if a JPEG lacks an embedded profile but sets EXIF `InteroperabilityIndex` (`R03`/Adobe RGB, `R98`/sRGB, `THM`/thumbnail), we embed an Adobe-compatible profile; otherwise we leave color untouched.
- Metadata is removed from outputs to keep thumbs small.

### Goals

- Produce consistent thumbnails for all configured sizes and resample modes.
- Preserve color fidelity when cameras signal color space through EXIF interop tags.
- Keep error paths non-fatal: invalid sizes, missing files, or absent profiles should return errors (not panics).

### Non-Goals

- Serving or caching thumbnails (handled elsewhere).
- Full ICC workflow management; only minimal embedding for interop-index cases.

### Package Layout (Code Map)

- `vips.go` — main `Vips` entry: load, resize/crop, strip metadata, export.
- `vips_icc.go` — EXIF InteroperabilityIndex handling and ICC embedding.
- `icc.go` — lists bundled ICC filenames (`IccProfiles`) and `GetIccProfile` helper.
- `resample.go`, `sizes.go` — resample options and predefined sizes.
- `thumb.go` and helpers — naming, caching, file info.
- Tests live alongside sources (`*_test.go`, fixtures under `testdata/`).

### ICC & Interop Handling

- EXIF `InteroperabilityIndex` codes we honor (per EXIF TagNames and regex.info):
  - `R03` → Adobe RGB (1998) compatible (`a98.icc`, etc.)
  - `R98` → sRGB (assumed default; no embed)
  - `THM` → Thumbnail (treated as sRGB; no embed)
- If an ICC profile already exists, we skip embedding.
- Exiftool name differs (libvips expects `exif-ifd4-InteroperabilityIndex`):
  ```
  exiftool -InteropIndex -InteropVersion -icc_profile:all -G -s file.jpg
  ```
- Test Files:
  - `testdata/interop_index.jpg` — R03 interop tag, no ICC (expects Adobe profile embed).
  - `testdata/interop_index_srgb_icc.jpg` — R03 tag with embedded ICC (must remain unchanged).
  - `testdata/interop_index_r98.jpg` — R98 interop tag, no ICC (should stay sRGB without embedding).
  - `testdata/interop_index_thm.jpg` — THM interop tag, no ICC (thumbnail; should remain unchanged).
- References:
  - [EXIF TagNames (InteroperabilityIndex)](https://unpkg.com/exiftool-vendored.pl@10.50.0/bin/html/TagNames/EXIF.html)
  - [Digital-Image Color Spaces: Recommendations and Links](https://regex.info/blog/photo-tech/color-spaces-page7)

### Tests

- Fast scoped: `go test ./internal/thumb -run 'Icc|Vips' -count=1`
- Full: `go test ./internal/thumb -count=1`

### Lint & Formatting

- Format: `make fmt-go`
- Lint: `make lint-go` or `golangci-lint run ./internal/thumb/...`

### Notes

- When adding ICC files, place them in `assets/profiles/icc/` and append to `IccProfiles`.
- Comments for exported identifiers must start with the identifier name (Go style).
