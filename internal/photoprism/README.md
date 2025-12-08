## PhotoPrism — Core Package

**Last Updated:** November 22, 2025

### Overview

`internal/photoprism` contains the core application logic for scanning originals, extracting metadata, generating thumbnails, importing/stacking media, and orchestrating converters (FFmpeg/ImageMagick/ExifTool). It wires configuration, indexer, converters, files/photos repositories, and background workers into a single package that other layers (CLI, API, workers) call.

#### Goals

- Provide a single, tested entrypoint for indexing/importing media files (`Index`, `IndexMain`, `ImportWorker`).
- Normalize metadata (including sub-second timestamps) before persisting to entities and sidecars.
- Keep converters and thumbnails consistent across CLI, background jobs, and tests.

#### Non-Goals

- Direct HTTP handling (lives in `internal/server`).
- Database models (lives in `internal/entity`).
- UI concerns (handled by frontend/JS).

### Package Layout (Code Map)

- Indexing/import: `index.go`, `index_main.go`, `index_mediafile.go`, `index_related.go`, `import_worker.go`, `files.go`, `photos.go`.
- Media files & helpers: `mediafile*.go`, `mediafile_thumbs.go`, `mediafile_vision.go`, `convert_*.go`, `colors.go`, `label.go`.
- Faces/people: `faces_*.go` (audit, clustering, matching, optimize).
- Backups: `backup/` (database and sidecar YAML backup/restore helpers).
- Downloads: `dl/` (export and download handlers/helpers).
- Service registry: `get/` (registry lookups and helper commands).
- Tests & fixtures: `*_test.go`, `testdata/`, uses shared test config (`config.TestConfig()`).

### Related Packages & Docs

- [`internal/entity`](../entity) — persistence models and DB helpers used by the indexer.
- [`internal/server`](../server/README.md) — HTTP routing that calls into this package.
- [`internal/meta`](../meta/README.md) — metadata extraction (EXIF/JSON) feeding `MediaFile.MetaData()`.
- [`internal/ffmpeg`](../ffmpeg/README.md) — media transcoding helpers used by converters.
- [`internal/thumb`](../thumb) — thumbnail generation helpers.

### Usage & Test Guidelines

- Indexing: use `IndexMain` / `IndexRelated` via `IndexMediaFile` helpers; prefer `IndexOptions` factories.
- Import: run via `ImportWorker` with `ImportOptions`; stacked handling is driven by metadata and document IDs.
- Converters: use `Convert.ToImage` / `Convert.ToVideo` / `Convert.ToJson`; options come from `config.Config`.
- Vision: thumbnails for vision models are selected in `mediafile_vision.go`; ensure models exist in `internal/ai/vision`.
- Tests: targeted runs keep iteration fast, e.g.  
  - `go test ./internal/photoprism -run TestMediaFile_ -count=1`  
  - `go test ./internal/photoprism/index_mediafile_test.go -run TestIndexMediaFile`  
  Full suite: `go test ./internal/photoprism/...` (heavy; migrates fixtures).
- Fixtures live under `storage/testdata`; tests expect initialized config (`config.TestConfig()` / `config.NewMinimalTestConfigWithDb`).

### Operational Notes

- Sub-second EXIF timestamps are preserved through metadata parsing and visible in `MediaFile.MetaData()`; database columns remain second-precision.
- File I/O permissions must use `pkg/fs` modes; overwrite requires explicit `force` flags.
- Exec calls to external tools are parameterized by config paths/binaries (`config.Config`).
- Stacking rules honor document IDs, time/place proximity, and configuration (`StackUUID`, `StackMeta`).
