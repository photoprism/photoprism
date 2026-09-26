## PhotoPrism — Core Package

**Last Updated:** September 26, 2026

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
- 360° originals: `mediafile_insta360.go` and `mediafile_projection.go` detect fisheye/dual-fisheye sources, `convert_image*.go` / `convert_video_avc.go` dewarp them to equirectangular derivatives, `index_mediafile.go` stacks late lens and proxy files by their stack name (`fs.StackPrefix`), and `index_insta360.go` merges photos that older versions created for separate lens files whenever a capture is indexed in rescan mode, keeping the lowest-ID photo's archive state. `import_worker.go` recognizes imported captures by their original names and rebuilds their combined preview after indexing.
- Faces/people: `faces_*.go` (audit, clustering, matching, consensus naming, optimize, migrate, crop source); face-marker persistence and XMP face-tag import in `index_faces.go` / `index_faces_xmp.go` (gated by `PHOTOPRISM_XMP_FACES`).
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
- Import: run via `ImportWorker` with `ImportOptions`; files imported together are stacked as one related set, files from separate batches only through a shared document ID or matching capture metadata.
- Converters: use `Convert.ToImage` / `Convert.ToVideo` / `Convert.ToJson`; options come from `config.Config`.
- Vision: thumbnails for vision models are selected in `mediafile_vision.go`; ensure models exist in `internal/ai/vision`.
- NSFW: `index_mediafile.go` flags new photos as `PhotoPrivate` when the labels-path NSFW shortcut (LLM with `DETECT_NSFW=true && EXPERIMENTAL=true`) hits or, as a fallback, when `m.DetectNSFW()` returns true and `PHOTOPRISM_DETECT_NSFW=true`. Both promotions short-circuit when `DetectNSFW()` is false. Full call-graph + flag matrix in [`internal/ai/nsfw/README.md`](../ai/nsfw/README.md).
- Tests: targeted runs keep iteration fast, e.g.  
  - `go test ./internal/photoprism -run TestMediaFile_ -count=1`  
  - `go test ./internal/photoprism/index_mediafile_test.go -run TestIndexMediaFile`  
  Full suite: `go test ./internal/photoprism/...` (heavy; migrates fixtures).
- Fixtures live under `storage/testdata`; tests expect initialized config (`config.TestConfig()` / `config.NewMinimalTestConfigWithDb`).
- `internal/photoprism` tests isolate package-level storage and SQLite DSN in `TestMain` using temporary per-process paths (`PHOTOPRISM_STORAGE_PATH`, `PHOTOPRISM_TEST_DSN`) to avoid flaky cross-process collisions on macOS/Linux when multiple `go test` processes run in parallel.
- Stateful tests that import/index media files should prefer isolated helpers like `config.NewMinimalTestConfigWithDb("<name>", filepath.Join(t.TempDir(), "storage"))` instead of shared `config.TestConfig()`.

### Transport-Stream Coordination

Transport-stream conversions share compatible in-flight work by output destination across `Convert`
instances in the same process. Coordination includes remuxing and any fallback transcode. Requests
with different source versions, configuration objects, FFmpeg exclusion snapshots, encoders or
force/mutex settings queue for that destination and recheck the output; unrelated destinations retain
their existing scheduling. Each caller receives its own media object. Completed operations are
released for later retries; overlapping compatible callers share failures.
This is not a cross-process lock and does not alter animated-WebP or encoder mutex behavior.

### Operational Notes

- Sub-second EXIF timestamps are preserved through metadata parsing and visible in `MediaFile.MetaData()`; database columns remain second-precision.
- File I/O permissions must use `pkg/fs` modes; overwrite requires explicit `force` flags.
- Exec calls to external tools are parameterized by config paths/binaries (`config.Config`).
- Stacking: a new file joins the photo in its folder named after its stack name (`fs.StackPrefix`); `StackSequences`, `StackUUID`, and `StackMeta` add matching by sequence-stripped name, shared document ID, and exact capture time, place, and camera serial.
- Forced rescans (`IndexOptions.Rescan=true`) run folder album reconciliation at the end of indexing via `entity.ReconcileOriginalsFolderAlbums(...)`; normal incremental runs skip this pass.
- Updated or newly added XMP sidecars next to originals are re-read on normal incremental passes. The filesystem walk compares each sidecar's modification time with `files.mod_time`, resolves its main media file from the Files cache, and queues deduplicated main-file jobs only after a successful walk; on forced rescans this detection is skipped because every main file is reindexed and re-reads its sidecar anyway. External XMP edits merge with `SrcXmp` priority, while `SrcManual` values are preserved. A sidecar that fails to parse records the error and advances its `mod_time`, so it is retried only after another edit instead of on every pass. Incremental sidecar deletion is not supported, and automatic removal of stale XMP-derived metadata is not guaranteed by a forced rescan: fields such as `UUID`, `CameraSerial`, and primary `InstanceID` do not retain enough source information for complete reconciliation.
- Folder create/index conflict lookup uses unscoped folder reads in `internal/entity/folder.go` so soft-deleted rows are detectable for troubleshooting instead of causing repeated create/find mismatches.

### JSON Metadata Output

ExifTool JSON output is captured with the shared `meta.JSONFileLimit()` limit (1 MiB by default).
Output beyond that limit returns `meta.ErrJSONFileTooLarge` without publishing a partial
cache file. Diagnostic output is retained up to 64 KiB, with a truncation marker on errors;
normal conversion timeouts and process cleanup remain in effect. Cached/external sidecars
are independently bounded by the [metadata JSON reader](../meta/README.md#json-sidecar-reader).

Operators can override the shared JSON byte limit with `PHOTOPRISM_JSON_LIMIT` (positive
decimal bytes, for example `4194304` for 4 MiB). Empty, invalid, zero, negative, or
out-of-range values retain the 1 MiB default; there is no unlimited setting. The override
applies to both sidecar reads and ExifTool stdout capture, not the 64 KiB stderr bound.
