## Import/Index

- ImportWorker may skip files if an identical one already exists (duplicate detection). Use unique copies or assert DB rows after ensuring a non-duplicate destination.
- Mixed roots: keep `SamplesPath()/ImportPath()/OriginalsPath()` consistent so `RelatedFiles` and `AllowExt` behave as expected.
- `IndexOptions*` helpers require a `*config.Config`; pass the active config (or `config.NewMinimalTestConfig(t.TempDir())` in unit tests) so face/label/NSFW scheduling matches the current run.
- Folder albums use path-first lookup/update (`album_path`) to avoid slug collisions for emoji child paths.
- Label/label-search logic should reuse `entity.FindLabels(...)`, `entity.FindLabelIDs(...)`, and `entity.LabelSlugs(...)` for homophone-aware exact-name matching — avoid ad-hoc slug SQL in search code.
- Vision worker scheduling is controlled via `VisionSchedule` / `VisionFilter` and `Run` in `vision.yml`. Use `vision.FilterModels` and `entity.Photo.ShouldGenerateLabels/Caption` to decide when work is required before loading media files.

## Download CLI Workbench (yt-dlp, remux, importer)

**Code anchors:** CLI flags/examples in `internal/commands/download.go`; core impl in `internal/commands/download_impl.go`; yt-dlp helpers in `internal/photoprism/dl/*` (`options.go`, `info.go`, `file.go`, `meta.go`); importer entry in `internal/photoprism/get/import.go`; import options in `internal/photoprism/import_options.go`.

**Fast test runs:**
- yt-dlp package: `go test ./internal/photoprism/dl -run 'Options|Created|PostprocessorArgs' -count=1`
- CLI command: `go test ./internal/commands -run 'DownloadImpl|HelpFlags' -count=1`

**FFmpeg-less tests:** set `c.Options().FFmpegBin = "/bin/false"` and `c.Settings().Index.Convert = false` to avoid ffmpeg dependencies when not validating remux.

**Stubbing yt-dlp (no network):** use a shell script that prints minimal JSON for `--dump-single-json` and creates a file + prints its path when `--print` is requested. Harness env vars: `YTDLP_ARGS_LOG` (append final args for assertion), `YTDLP_OUTPUT_FILE` (absolute file path to create for `--print`), `YTDLP_DUMMY_CONTENT` (file contents to avoid importer duplicate detection between tests).

**Remux policy:**
- Pipe method: PhotoPrism remux (ffmpeg) always embeds title/description/created.
- File method: yt-dlp writes files; we pass `--postprocessor-args 'ffmpeg:-metadata creation_time=<RFC3339>'` so imports get `Created` even without local remux.
- Default policy `auto`; use `always` for the most complete metadata.
- CLI defaults: `photoprism dl` defaults to `--method pipe` and `--impersonate firefox`; pass `-i none` to disable impersonation.
