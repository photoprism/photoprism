# PhotoPrism Core Import & Index Guidelines

**Last Updated:** April 9, 2026

## Import & Index Behavior

- `ImportWorker` may skip a file when an identical file already exists. Use unique copies or assert DB rows only after ensuring the destination is not a duplicate.
- Keep `SamplesPath()`, `ImportPath()`, and `OriginalsPath()` consistent when testing related files so `RelatedFiles` and `AllowExt` behave correctly.
- `IndexOptions*` helpers now require a `*config.Config`; pass the active config, or `config.NewMinimalTestConfig(t.TempDir())` in unit tests, so face, label, and NSFW scheduling matches the current run.
- Vision scheduling is controlled by `VisionSchedule`, `VisionFilter`, and the `Run` property in `vision.yml`. Use helpers such as `vision.FilterModels` and `entity.Photo.ShouldGenerateLabels` or `ShouldGenerateCaption` before loading media.
- Folder albums use path-first lookup and update via `album_path` to avoid slug collisions for emoji child paths; re-indexing may repair stale collision titles while preserving user-custom titles.
- Reuse `entity.FindLabels(...)`, `entity.FindLabelIDs(...)`, and `entity.LabelSlugs(...)` for label lookup so homophone-aware exact-name matching stays aligned across entity and search packages.

## File I/O Overwrite Policy

- Default to safety: callers must not overwrite non-empty destination files unless they opt in with `force=true`.
- Replacing an empty destination file is allowed without `force=true`.
- Open overwriting destinations with `O_WRONLY|O_CREATE|O_TRUNC` so stale trailing bytes cannot survive; use `O_EXCL` when the caller must detect collisions.
- App-level overwrite helpers live in `internal/photoprism/mediafile.go`; reusable helpers live in `pkg/fs/copy.go` and `pkg/fs/move.go`.
- Set `force=true` only for explicit replace flows or admin tools with confirmed overwrite, not for import or index flows that touch Originals.

## Tests

- In `internal/photoprism` tests, rely on `photoprism.Config()` for runtime-accurate behavior; only build a new config if you replace it with `photoprism.SetConfig`.

