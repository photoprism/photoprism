## PhotoPrism — Metadata Pipeline

**Last Updated:** June 30, 2026

### Overview

The `internal/meta` package extracts, normalizes, and reports metadata from images, videos, and sidecars (Exif, XMP, JSON). It produces a `meta.Data` struct that downstream components (indexer, UI, API) consume for dates, GPS, camera/lens info, keywords, and motion-photo flags. The package aims to be loss-tolerant (accepts imperfect files), deterministic (stable parsing order), and explicit about fallbacks.

### Guidelines

- Keep nanosecond precision in `meta.Data`; adjust consumers/tests instead of truncating here.
- When comparing or persisting times, be aware of second-only storage in entity and DB layers.
- For stacking or dedupe features, use second-based keys unless the DB schema is upgraded.
- When adding new parsers, ensure they fail softly and add test fixtures mirroring real-world oddities.

### Time & Precision

- Parsers preserve sub-second timestamps found in Exif/XMP/JSON (`TakenAt`, `TakenAtLocal`, `TakenNs`). Tests expect nanosecond precision where available.
- Downstream persistence truncates to whole seconds:
  - `MediaFile.TakenAt()` truncates `meta.Data` timestamps to seconds before caching.
  - Entity columns `photos.taken_at`, `photos.taken_at_local`, and `files.photo_taken_at` are `DATETIME` (no fractional seconds).
  - YAML metadata backups serialize the entity values, so they also lose sub-second precision.
- Stack/search logic remains second-based (`MapKey` uses `takenAt.Unix()`), so nanoseconds do not affect grouping or comparisons.
- If future work needs sub-second storage, columns must switch to `DATETIME(6)` (or similar) and the truncation in `MediaFile.TakenAt()` removed.

### Parsing Order & Fallbacks

- Exif → XMP → JSON (ExifTool/GPhotos/motion) → filename → filesystem mtime. Each stage logs source and errors but continues when safe.
- Brute-force Exif search is used when native parsers fail; errors are logged with context.
- GPS parsing supports decimal, DMS (`51 deg 15' 17.47" N`), and the 2-component Adobe XMP form (`52,30.4567N`); regexes are kept simple and precompiled.

### XMP Sidecar Reader

The `.xmp` sidecar reader (`xmp.go` + `xmp_document.go`) is XPath-based on `antchfx/xmlquery` and namespace-aware via `xpath.CompileWithNS`. Each accessor declares a `chainXPath` priority list; the engine evaluates links left-to-right and returns the first non-empty match. Composition (Lat sign from `GPSLatitudeRef`, sub-second join from `SubSecTimeOriginal`, etc.) lives in the relevant accessor — never in the chain engine.

- **Loader security guards.** `Load` rejects sidecars larger than 1 MiB (`ErrXmpFileTooLarge`) and documents nesting deeper than 64 elements (`ErrXmpTooDeep`). XXE and DTD attacks are mitigated by `encoding/xml`'s default behavior (no external entity resolution); `xmp_security_test.go` is the regression guard.
- **Element-or-attribute helper.** RDF/XML allows scalar properties to be expressed as either child elements or attributes on `rdf:Description`. The `elemOrAttr(qname)` helper builds a union XPath that matches both — required because digiKam emits `xmpMM:*`/`exif:*`/`tiff:*` as attributes while Adobe writes them as child elements.
- **Adding an accessor.** Declare a `chainXPath` at package init using `mustCompile` (or `elemOrAttr` for scalar fields), document the priority chain in a one-line comment, then add the accessor that calls `firstNonEmpty` (for scalars) or `queryAll` (for `rdf:Bag`/`rdf:Seq`). Wire the new field into `xmp.go` with the existing "set only when non-empty" pattern.
- **Source priority.** Sidecar values are tagged `SrcXmp` (priority 32), which outranks `SrcMeta` (priority 16) at the entity layer. Re-indexing a photo after the sidecar has been added overwrites previously-`SrcMeta` values without a database wipe.
- **Keywords vs. Subject.** `dc:subject` (Adobe's "Keywords" panel) maps to the descriptive `Details.Subject` field — never the `Details.Keywords` field — matching the embedded/ExifTool path where `data.Subject` comes from the `dc:subject`-backed `Subject` tag and `data.Keywords` from IPTC `Keywords`. The XMP path adds only the derived `flash`/`panorama`/`hdr` keywords. Hierarchical-label and face-region parsing (Labels, people Subjects) is a planned extension tracked under epic [#2260](https://github.com/photoprism/photoprism/issues/2260).
- **Coverage.** The fixture corpus under `testdata/xmp/{adobe,darktable,digikam,synthetic}/` documents the full set of supported tags and their per-fixture provenance.

### Motion Photos & Embedded Media

- Motion-photo JSON readers set `HasThumbEmbedded` / `HasVideoEmbedded`, `Codec`, `Duration`, and capture accurate timestamps (including ns) when present.
- Time zones from motion metadata are respected; missing zones fall back to UTC.

### Sanitization

- `SanitizeString`, `SanitizeUnicode`, and related helpers strip binary markers, quotes, and invalid Unicode; filenames and keywords use lower-case, dash/underscore-safe regexes.
- Lower-case regex and quote removal now use `ReplaceAll` and raw strings to avoid double escaping.
- Google Photos JSON coordinates are clamped to hard latitude/longitude bounds with `geo.ClampCoordinateBounds` before assigning `meta.Data.Lat` and `meta.Data.Lng`.

### Docs & References

- External tag references are listed in `docs.go`.
- Tests under `internal/meta/testdata` cover Exif, XMP, motion photos, and edge cases (missing headers, panoramas, time offsets).
