## PhotoPrism — Metadata Pipeline

**Last Updated:** November 22, 2025

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
- GPS parsing supports decimal and DMS formats; regexes are kept simple and precompiled.

### Motion Photos & Embedded Media

- Motion-photo JSON readers set `HasThumbEmbedded` / `HasVideoEmbedded`, `Codec`, `Duration`, and capture accurate timestamps (including ns) when present.
- Time zones from motion metadata are respected; missing zones fall back to UTC.

### Sanitization

- `SanitizeString`, `SanitizeUnicode`, and related helpers strip binary markers, quotes, and invalid Unicode; filenames and keywords use lower-case, dash/underscore-safe regexes.
- Lower-case regex and quote removal now use `ReplaceAll` and raw strings to avoid double escaping.

### Docs & References

- External tag references are listed in `docs.go`.
- Tests under `internal/meta/testdata` cover Exif, XMP, motion photos, and edge cases (missing headers, panoramas, time offsets).
