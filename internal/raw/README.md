## PhotoPrism — RAW Image Conversion

**Last Updated:** June 30, 2026

### Overview

`internal/raw` provides command builders and heuristics for converting camera RAW images (including Adobe Digital Negative, `.dng`) to JPEG with Darktable, RawTherapee, and ExifTool. It mirrors the `internal/ffmpeg` pattern: the builders return `*exec.Cmd` and the caller passes the binary path plus options, while the orchestration that runs the commands and accepts their output stays in `internal/photoprism` (`Convert.JpegConvertCmds` and the `ToImage` convert loop).

The package owns the small amount of RAW-specific knowledge that would otherwise be scattered through the converter: the command flags, the stderr patterns that mark an untrustworthy decode, and the list of formats whose embedded preview is unusable.

#### Builders

- `DarktableCmd(DarktableOptions) (*exec.Cmd, bool)` — also reports whether the command needs a global mutex (presets mode runs one instance at a time).
- `TherapeeCmd(bin, rawName, jpegName, profile string, quality int) *exec.Cmd` — RawTherapee render to JPEG (named without the `Raw` prefix to avoid a `raw.RawTherapee…` stutter).
- `ExifToolJpgFromRawCmd` / `ExifToolPreviewImageCmd` — extract the full-resolution / smaller embedded preview to stdout.

### Converter Priority

For a RAW input, `JpegConvertCmds` appends commands in this order, and the convert loop accepts the first whose output passes:

1. **Darktable** — full RAW developer (preferred).
2. **RawTherapee** — full RAW developer (when Darktable is unavailable or fails).
3. **Embedded camera preview** via ExifTool (largest first) — the last resort.

The embedded preview is last because a full render is higher quality, but it has correct colors when the RAW developers cannot identify the sensor (e.g. very recent Canon CR3 bodies otherwise come out magenta), so it wins when the developers are unavailable or — for the gated formats described below — produce an untrustworthy render.

### Conversion Gating (`--disable-raw`)

`PHOTOPRISM_DISABLE_RAW` disables both **indexing** and **conversion** of RAW, via two independent gates:

- **Indexing/import/upload** skip RAW entirely before it reaches the converter (`index.go`, `mediafile_related.go`, `import.go`, `users_upload.go`).
- **Conversion** gates only the RAW *renderers* (Darktable, RawTherapee, and the macOS `sips` path) on `RawEnabled()`. ExifTool preview extraction runs regardless, so an already-embedded preview can still be extracted when rendering is disabled (e.g. via the `convert` CLI).

### Error Handling

The three tools fail differently, so they are guarded differently:

- **Darktable** signals an unsupported sensor with a **non-zero exit code** and writes no file (its diagnostics go to stdout, e.g. `[libraw_open] detected unsupported image`). The convert loop's exit-code check skips it, so no stderr inspection is needed.
- **RawTherapee** can **exit 0 yet produce wrong colors** (e.g. a default white balance for an unidentified sensor) while printing a dcraw-derived message to stderr. For formats in the discard set (`DiscardRenderOnWarning`, default `.cr3`) its command is tagged with `ConvertCmd.WithStderrRejection(raw.DecoderErrors...)`; a match discards the output and the loop falls through to the embedded preview. `WhiteBalanceError` is the canonical case; `DecoderErrors` adds the other untrustworthy-decode messages (sourced from `dcraw.c`). Other RAW formats keep the render — see *Discard-on-Warning Gate* below.
- **Embedded previews** can be header-valid yet fail the thumbnailer (a bogus Huffman table only surfaces during shrink-on-load), so they are tagged `WithImageVerification`, which runs `thumb.Verify` to force the decode before acceptance.

Because of this, `StderrRejected` is RawTherapee-specific by design — Darktable does not need it (it fails via exit code, not via accepted-but-wrong output).

### Discard-on-Warning Gate

`DiscardRenderOnWarning(ext)` decides, per RAW format, whether a RawTherapee render is discarded when its stderr matches `DecoderErrors`. It defaults to Canon CR3 (`.cr3`): recent CR3 bodies, which neither RAW developer can demosaic, render magenta, but they reliably embed a good preview to fall back to. The white-balance warning alone cannot tell a magenta render from a fine one — RawTherapee emits it for any sensor whose coefficients it cannot read, including obscure bodies whose default-WB render is perfectly fine — so the rejection is gated on the format as well as the warning.

Formats RawTherapee alone can decode (e.g. `.raw`, `.kdc`) are deliberately left out: they have no embedded preview to fall back to, so discarding the render would leave nothing to index. The set is overridable via `PHOTOPRISM_RAW_DISCARD_ON_WARNING_EXT` (a comma-separated list that **replaces** the default), an escape hatch — outside the CLI/config surface — for adding a newly found magenta-prone format without a rebuild. A gated format must allow embedded-preview extraction (`PreviewExtAllowed`); otherwise its render would be discarded with no fallback, so PhotoPrism logs a warning at startup if an override names such a format.

### Preview-Unsafe Formats

`previewUnsafeExt` lists RAW extensions whose embedded JPEG preview is known to be unusable, exposed via `PreviewExtAllowed(ext)`. Currently `.mos` (Leaf), whose preview is a bogus-Huffman JPEG that passes a MIME sniff but fails thumbnailing; skipping it forces a full RAW render.

### PNG Output (Deferred)

RAW is not converted to PNG today: `ToImage` writes a `.jpg` sidecar for RAW, and `PngConvertCmds` has no RAW renderer at all (it logs and skips RAW). A lossless RAW export/preview (e.g. a future UI "export" tool) is the case that would need RAW→PNG. When that lands, build the PNG path to its own spec rather than copying `JpegConvertCmds`: it should render with Darktable/RawTherapee in **PNG mode** (RawTherapee needs `-n`/`-b16`, a new builder), and it should **not** use the embedded-JPEG preview fallback, which is lossy and would defeat a lossless export.

### Known Gaps

- RAW→PNG is unimplemented (see *PNG Output* above); `PngConvertCmds` logs and skips RAW. RAW rendering to JPEG (Darktable, RawTherapee, `sips`) is gated on `RawEnabled()`.
