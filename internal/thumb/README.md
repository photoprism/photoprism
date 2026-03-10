## PhotoPrism — Thumbnails Package

**Last Updated:** March 6, 2026

### Overview

`internal/thumb` builds thumbnails with libvips, handling resize/crop options, color management, metadata stripping, and format export (JPEG/PNG). It is used by PhotoPrism’s workers and CLI to generate cached thumbs consistently.

### Constraints

- Uses libvips via govips; initialization is centralized in `VipsInit`.
- Requires libvips 8.14+ with the current govips bindings (`github.com/davidbyttow/govips/v2`).
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

### Go 1.26 JPEG Notes

Go `1.26.0` replaced the standard `image/jpeg` encoder and decoder. This package primarily relies on libvips for thumbnail generation, but helper paths and tests that decode or encode JPEGs through Go libraries can observe behavior changes after toolchain upgrades.

Observed impact during internal comparison runs (Go `1.25.4` vs `1.26.0`):

- **Compatibility** — No decode failures for 55/55 JPEG fixtures in `assets/samples` on either version.
- **Decoded Pixels** — All scanned JPEG fixtures produced different decoded pixel hashes across versions, even though dimensions were unchanged.
- **Re-Encode Size** — Re-encoded JPEG sizes changed slightly in both directions; aggregate deltas were small (about `+0.014%` at default quality, about `+0.017%` at quality 95 in our fixture scan).
- **Throughput** — Micro-benchmark runs showed modest improvements in decode and decode+encode throughput in Go `1.26.0`.

Testing guidance:

- Do not rely on bit-for-bit JPEG output across Go toolchain upgrades.
- Prefer assertions on image dimensions, error-free processing, and perceptual/tolerance metrics where appropriate.
