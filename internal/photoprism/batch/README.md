## PhotoPrism — Batch Edit Package

**Last Updated:** November 23, 2025

### Overview

Batch editing allows signed-in users to update the metadata, albums, and labels of multiple photos without having to load and edit each one individually.

The `internal/photoprism/batch` package implements the form schema (`PhotosForm`), validation helpers, album/label mutation helpers, and persistence functions (`SavePhotos`) that power the `/api/v1/batch/photos/edit` endpoint. It exists so the API can keep responses consistent with the UI: mixed values stay round-trippable, add/remove operations track intent, and only changed columns hit the database.

#### Context & Constraints

- Community requests such as [Issue #271](https://github.com/photoprism/photoprism/issues/271) emphasized the need to bulk-edit core metadata (location, time zone, titles) instead of repeating the same change photo by photo.
- [PR #5324](https://github.com/photoprism/photoprism/pull/5324) introduced the modern batch dialog, chip controls, and validation rules that this package still serves.
- Batch edits run inside regular API workers; there is no dedicated job queue. We therefore optimize for O(n) database work across selected photos, avoid global locks, and offload heavy recomputation to existing workers (meta, labels, search index).
- Batch edit requests run under a dedicated `mutex.BatchEdit` activity to serialize concurrent edits and cancel the background meta worker while changes are applied.
- Frontend components expect round-trip metadata even for fields that are not yet editable (ISO, focal length, copyright, etc.), so the form structs intentionally contain more data than the dialog renders.

#### Goals

- Provide a single schema (`PhotosForm`) for reading and writing mixed selections.
- Guarantee that album/label updates obey ACLs and deduplicate creations, even when multiple requests fire concurrently.
- Minimize writes by only persisting changed columns and deferring derived work (labels, keyword re-indexing) to background workers.
- Return refreshed photo models so the UI can immediately render the persisted state without extra queries.
- Serialize batch edits with `mutex.BatchEdit` so only one batch edit runs at a time and shutdown can cancel ongoing requests; the meta worker is canceled while edits are applied.

#### Non-Goals

- Defining frontend validation or presentation; Vue components own layout, tooltips, and translations.
- Replacing specialized batch routes (`batch/photos/delete`, `…/archive`, etc.). This package focuses on metadata edits only.
- Managing worker scheduling. We simply mark `checked_at = NULL` so the metadata worker decides when to reprocess.

### Architecture & Request Flows

1. The router registers `BatchPhotosEdit` (`internal/api/batch_photos_edit.go`) under `/api/v1/batch/photos/edit`.
2. The handler authenticates via `Auth(..., acl.ResourcePhotos, acl.ActionUpdate)` before accepting JSON payloads shaped like `batch.PhotosRequest { photos: [], values: {} }`.
3. The handler always reuses the ordered `search.BatchPhotos` results when serializing the `models` array so every response mirrors the original selection and exposes the full `search.Photo` schema (thumbnail hashes, files, etc.) required by the lightbox.
4. After persisting updates, the handler issues a follow-up `query.PhotoPreloadByUIDs` call so `batch.PrepareAndSavePhotos` gets hydrated entities for album/label mutations without disrupting the frontend-facing payload.
5. `batch.PrepareAndSavePhotos` iterates over the preloaded entities, applies requested album/label changes, builds `PhotoSaveRequest` instances via `batch.NewPhotoSaveRequest`, and persists the updates before returning a summary (requests, results, updated count, `MutationStats`) to the API layer.
6. `resolveBatchItemValues` runs before per-photo work so album/label additions referenced by title are looked up or created once per batch (rather than per photo) and deleted albums/labels are restored before use.
7. `SavePhotos` (invoked by the helper) loops once per request, updates only the columns that changed, clears `checked_at`, touches `edited_at`, and queues `entity.UpdateCountsAsync()` once if any photo saved. When album mutations occurred and YAML backups are enabled, the resolved album list is written back to disk via `updateAlbumBackups` after all database work succeeds.
8. Refreshed models and values are sent back in the response form so the frontend can merge and display the changes, and the mutation stats drive the production log line (`updated photo metadata (1/3) and labels (3/3)`) so operators can see which parts of the request succeeded even when metadata columns remained untouched.

### Batch Edit API Endpoint

`POST /api/v1/batch/photos/edit` accepts a `PhotosRequest` payload and returns the refreshed photo models plus a `PhotosForm` snapshot. All fields follow JSON casing from `internal/photoprism/batch/request.go` and `photos.go`.

| Field                   | Type             | Examples                                                     | Notes                                                                                         |
|-------------------------|------------------|--------------------------------------------------------------|-----------------------------------------------------------------------------------------------|
| `photos`                | `string[]`       | `['pq1z9t3', 'px4y2k0']`                                     | Required. Contains `PhotoUID` values selected in the UI. Empty lists are rejected with `400`. |
| `values`                | `PhotosForm`     | `{ "Title": { "action": "update", "value": "Vacation" } }`   | Optional. When omitted, the endpoint only loads the selection + aggregated form data.         |
| `values.Albums.items[]` | `Items` entry    | `{ "value": "ab1c", "title": "Favorites", "action": "add" }` | Action must be `add`/`remove`; `title` is used to create albums when `value` is empty.        |
| `values.Labels.items[]` | `Items` entry    | `{ "value": "lb2d", "action": "remove" }`                    | Removing requires a valid label UID. Adds accept either UID or plain title.                   |
| `values.TimeZone`       | `String` wrapper | `{ "value": "Europe/Berlin", "action": "update" }`           | Paired with `Day/Month/Year` to recompute `TakenAt` and `TakenAtLocal`.                       |
| `values.DetailsSubject` | `String` wrapper | `{ "action": "remove" }`                                     | Setting `action="remove"` clears the field without needing `value`.                           |

#### Request Example

```json
{
  "photos": ["pt1abcd", "pt2efgh"],
  "values": {
    "Title": { "value": "Sunset Cruise", "action": "update" },
    "Caption": { "action": "remove" },
    "TimeZone": { "value": "America/Los_Angeles", "action": "update" },
    "Day": { "value": 14, "action": "update" },
    "Month": { "value": 11, "action": "update" },
    "Year": { "value": 2025, "action": "update" },
    "Albums": {
      "action": "update",
      "items": [
        { "title": "Trips 2025", "action": "add" },
        { "value": "abcf1234", "action": "remove" }
      ]
    },
    "Labels": {
      "action": "update",
      "items": [
        { "value": "lbtravel", "action": "add" },
        { "value": "lbbeta", "action": "remove" }
      ]
    }
  }
}
```

#### Response Example

```json
{
  "models": [
    {
      "UID": "pt1abcd",
      "Title": "Sunset Cruise",
      "Favorite": true,
      "Albums": [
        { "UID": "trips25", "Title": "Trips 2025" }
      ],
      "Labels": [
        { "UID": "lbtravel", "Name": "Travel" }
      ]
    }
  ],
  "values": {
    "Title": { "value": "Sunset Cruise", "mixed": false, "action": "update" },
    "Caption": { "value": "", "mixed": false, "action": "remove" },
    "Albums": {
      "action": "update",
      "mixed": false,
      "items": [
        { "value": "trips25", "title": "Trips 2025", "mixed": false, "action": "none" }
      ]
    }
  }
}
```

### Frontend Integration

The SPA consumes the endpoint through a dedicated REST model, dialog component, and Vitest suites.

- **REST Model**: `frontend/src/model/batch.js` exports the `Batch` class, which encapsulates the `/batch/photos/edit` POST calls, manages hydrated `Photo` instances, tracks the current selection, and ensures mixed-value defaults match the backend schema.
- **Vue Components**: `frontend/src/component/photo/batch-edit.vue` renders the dialog, binding the model’s `values` to chips, combo boxes, and toggles. It is mounted via `frontend/src/component/dialogs.vue`, which exposes `<p-photo-batch-edit>` so any view can trigger batch edits.
- **Vitest Coverage**: `frontend/tests/vitest/model/batch.test.js` mocks Axios to verify that the model posts the correct payloads, updates cached photos, and handles no-op responses. `frontend/tests/vitest/component/photo/batch-edit.test.js` renders the dialog with Vue Test Utils to confirm field bindings, validation flows, and selection toggling behavior.

> **Note:** The frontend `model/batch.js` drops selection IDs that no longer have editable models, so dialog counters and clipboard actions match the set of photos the backend actually returned (archived/deleted photos are silently skipped).

### Gating & Configuration

- ACL: only sessions allowed to `update` the `photos` resource may call this endpoint. That includes administrators and contributors with write access; read-only tokens fail early.
- Selection limits: `search.BatchPhotos` caps the request using `form.SearchPhotos.MaxResults` (default 5,000) to prevent runaway updates, and the ordered list returned there is reused verbatim for the API response so we avoid desynchronizing the frontend selection.
- Workers: clearing `CheckedAt` ensures the metadata worker (`internal/workers/meta`) and downstream indexers revisit the files within the configured worker interval (default 10–20 minutes).
- Environment flags: standard safety toggles (`PHOTOPRISM_READONLY`, maintenance mode, etc.) still apply because the handler runs in the main API process.

#### Feature Flags & Permissions

- Batch edit is controlled via the `Features.BatchEdit` flag exposed in `customize.FeatureSettings`. The flag defaults to `true` alongside `Features.Edit`, but administrators can disable it in settings.
- The `/api/v1/batch/photos/edit` handler returns `ErrFeatureDisabled` (HTTP 403) whenever the flag is off, so automation cannot bypass the UI toggle.
- `Settings.ApplyACL` and `Settings.ApplyScope` only keep `BatchEdit` enabled when the current role can update photos **and** has `acl.AccessAll`; this prevents scoped API clients from invoking bulk edits outside their visibility window.
- The clipboard action (`component/photo/clipboard.vue`) checks the same flag and requires `photos/access_all` before publishing `dialog.batchedit`. If either requirement fails—or the selection only includes a single photo—the component falls back to the single-photo edit dialog so metadata edits remain available.
- Because the clipboard is our only UI entry point, disabling the flag hides the floating button, un-subscribes the dialog, and keeps backend enforcement consistent with the visible capabilities.

### Supported Fields & Values

`PhotosForm` currently carries:

- Core descriptive metadata: title, caption, type, favorite/private flags, scan/panorama toggles.
- Temporal data: exact timestamps, local offsets, broken-out year/month/day, and time zone identifiers.
- Location fields: latitude, longitude, altitude, ISO country, derived place/cell IDs (updated via `UpdateLocation`).
- Equipment identifiers: camera and lens IDs/serials for future UI expansion.
- Details block: subject, artist, copyright, license, keywords (ensuring context from Issue #271 remains editable once the UI exposes the rows).
- Albums & labels as `Items` lists, with `Mixed` markers and per-item actions.

Each field embeds one of the typed wrappers (`String`, `Bool`, `Time`, `Int`, etc.) so the UI knows whether a value is mixed, unchanged, updated, or removed.

### Overriding Values with Sources & Priorities

- `Action` enums (`none`, `update`, `add`, `remove`) describe intent. Strings treat `remove` the same as `update` plus empty values, allowing the backend to wipe titles/captions clean.
- Source columns (`TitleSrc`, `CaptionSrc`, `TypeSrc`, `PlaceSrc`, details `*_src`) keep track of provenance. `SavePhotos` updates them whenever batch edits win over prior metadata (EXIF, AI, manual, etc.).
- Album & label updates respect UID validation: `ApplyAlbums` verifies `PhotoUID` / `AlbumUID`, creates albums by title when needed, and delegates to `entity.AddPhotoToAlbums`, which now uses per-album keyed locks to avoid blocking unrelated requests. `Items.ResolveValuesByTitle` plus `resolveBatchItemValues` ensure those creations happen once per batch, so per-photo calls operate on cached UIDs instead of repeating lookups.
- Label writes reuse existing `PhotoLabel` rows when possible, force 100 % confidence for manual/batch additions, and demote AI suggestions by setting `uncertainty = 100` when users explicitly remove them.
- Keyword keywords stay consistent because label removals call `photo.RemoveKeyword` and `SaveDetails` immediately, while location edits append unique place keywords via `txt.UniqueWords`.

#### Rules for Deleting Photo Labels

The following shows the actions that Batch Edit is expected to perform when a user requests to remove a label. The **Outcome** can be to **Keep** the current `PhotoLabel`, to permanently **Delete** the `PhotoLabel`, or to **Update** the `PhotoLabel`'s `LabelSrc` or `Uncertainty`.

| `LabelSrc`            | `SrcPriority` | `Uncertainty` | Expected Outcome                                 |
|-----------------------|---------------|---------------|--------------------------------------------------|
| image, openai, ollama | 8             | 0             | **Update** LabelSrc: `batch`, Uncertainty: `100` |
| image, openai, ollama | 8             | 1-99          | **Update** LabelSrc: `batch` ,Uncertainty: `100` |
| image, openai, ollama | 8             | 100           | **Update** LabelSrc: `batch` ,Uncertainty: `100` |
| *                     | < 64          | 0             | **Update** LabelSrc: `batch`, Uncertainty: `100` |
| *                     | < 64          | 1-99          | **Update** LabelSrc: `batch` ,Uncertainty: `100` |
| *                     | < 64          | 100           | **Update** LabelSrc: `batch` ,Uncertainty: `100` |
| manual                | 64            | 0             | **Delete**                                       |
| manual                | 64            | 1-99          | **Delete**                                       |
| manual                | 64            | 100           | **Keep**                                         |
| vision                | 64            | 0             | **Update** LabelSrc: `batch`, Uncertainty: `100` |
| vision                | 64            | 1-99          | **Update** LabelSrc: `batch` ,Uncertainty: `100` |
| vision                | 64            | 100           | **Keep**                                         |
| batch                 | 64            | 0             | **Delete**                                       |
| batch                 | 64            | 1-99          | **Delete**                                       |
| batch                 | 64            | 100           | **Keep**                                         |
| admin                 | 128           | 0             | **Keep**                                         |
| admin                 | 128           | 1-99          | **Keep**                                         |
| admin                 | 128           | 100           | **Keep**                                         |

Based on the above examples, the following rules apply in the given order when processing label removals via batch edit:

- **Keep** Photo labels added from sources with a higher `SrcPriority` than `batch` (> 64), e.g., `admin` (128) i.e. skip/ignore while processing.
- **Keep** photo labels (`entity.PhotoLabel`) with the same `SrcPriority` as `batch` (64) and an `Uncertainty` of `100` (zero probability), to prevent unnecessary update queries.
  - This also prevents lower-priority sources, such as `image`, `openai`, or `ollama`, from adding the same labels again.
  - In practice, there should not be any `vision` labels with an `Uncertainty` value of `100`, since models would not suggest them with zero probability.
- **Update** photo labels added from sources with a lower `SrcPriority` than `batch` (< 64), as well as `vision` photo labels (!), which are typically auto-generated by setting the `LabelSrc` to `batch` and the `Uncertainty` to `100` (zero probability).
- **Delete** the remaining photo labels with the same `SrcPriority` as `batch` (64) and an `Uncertainty` of less than `100` (non-zero probability).
  - If the above rules were properly followed, no `vision` labels with an `Uncertainty` of less than `100` should exist at this stage.

### Performance & Concurrency

- `SavePhotos` only writes dirty columns and updates `photo_details` rows separately, reducing contention and avoiding `entity.SavePhotoForm`’s per-photo cache busts. The API keeps reusing the ordered `search.BatchPhotos` result for serialization, accepting the extra `query.PhotoPreloadByUIDs` call (post-save) until the lightbox can consume entity-only responses or leverage the ordered list helper in `pkg/list/ordered` for a unified flow.
- Batch responses reuse the same hydrated entities for both persistence and response rendering, so even selections with hundreds of photos issue a constant number of queries.
- Album mutations leverage `entity.lockAlbumKey()` (per-album mutex) so two batches editing disjoint albums proceed in parallel instead of waiting on the global lock used before PR #5324’s follow-up work.
- Label operations operate on preloaded associations (`indexPhotoLabels`) to avoid hitting the join table repeatedly.
- Background costs (keyword indexing, metadata regeneration) are deferred: clearing `CheckedAt` lets workers refresh derived data asynchronously, and `entity.UpdateCountsAsync()` runs once per batch regardless of size.

#### Database Locking

Testers reported intermittent `Error 1213 (40001)` deadlocks when multiple batch edits inserted or removed `photos_labels` rows at once. The label helpers now wrap `Save` / `Delete` calls in a tiny retry loop (`deadlockRetryAttempts=3`, `deadlockRetryDelay=25ms`) so most conflicts resolve transparently while still surfacing unexpected errors to the API layer. Monitor logs for repeated warnings to decide whether we need higher backoff values or additional transaction-level tuning in the future.

### Known Issues & Limitations

- UI coverage still lags the schema: EXIF controls such as ISO/f-number remain hidden, so users cannot yet set them even though backend transport exists (tracked in Issue #271).
- The endpoint assumes all selected photos are still readable; deleted originals during a batch run lead to warnings and skipped saves rather than hard failures.
- Keyword synchronization after label removal is best-effort; if `SaveDetails()` fails, the UI might display stale keywords until the next background refresh.

### Observability & Testing

- **Unit Tests**  
  - `internal/photoprism/batch/apply_albums_test.go` validates album mutations handling.  
  - `internal/photoprism/batch/apply_labels_test.go` validates label mutations, UID validation, and keyword handling.
  - `internal/photoprism/batch/convert_test.go` and `photos_test.go` cover form aggregation and mixed-value detection.  
  - `internal/photoprism/batch/datelogic_test.go` ensures cross-field dependencies (local time vs. UTC) stay consistent.  
  - `internal/photoprism/batch/save_test.go` exercises partial updates, detail edits, `CheckedAt` resets, and the `PreparePhotoSaveRequests` / `PrepareAndSavePhotos` helpers.  
  - `internal/api/batch_photos_edit_test.go` provides end-to-end coverage for response envelopes (`SuccessNoChange`, `SuccessRemoveValues`, etc.).
  - `internal/photoprism/batch/save_resolve_test.go` validates pre-resolution helpers for albums/labels, while `save_backup_test.go` covers the YAML backup flow controlled by `updateAlbumBackups`.
- **Logging**  
  - The package uses the shared `event.Log` logger. Debug logs trace selections, album/label changes, and dirty-field sets; warnings/errors surface failed queries so operators can inspect database health. The final `INFO` line now reports metadata success counts alongside album and label mutations (including error tallies) so label-only edits no longer read as “0 out of N photos”.
- **Metrics & Alerts**  
  - The API shares the `/api/v1/metrics` Prometheus endpoint; batch edits increment the standard HTTP counters/latencies. Consider dashboarding 5xx/4xx spikes for `/batch/photos/edit` if you rely heavily on automation.

> When adjusting `internal/photoprism/batch/apply_labels*.go`, remember tests assert cache behavior. Call `photo.PreloadLabels()` after deleting existing label relations and set `Items.Action = ActionUpdate` whenever labels are (re)added/removed in tests; otherwise cached joins may cause flakiness when subtests run together.

### Documentation & References

- <https://docs.photoprism.app/developer-guide/api/> — API Endpoints & Authentication 
- <https://docs.photoprism.dev/> — Swagger REST API Documentation
- [GitHub Issue #271: Add batch edit dialog to change the metadata of multiple pictures](https://github.com/photoprism/photoprism/issues/271)
- [GitHub PR #5324: Implements #271 by adding a batch edit dialog and API endpoint](https://github.com/photoprism/photoprism/pull/5324)

### Code Map

- `batch.go` — package doc + logger bindings.
- `request.go` / `response.go` — transport structs for the API payload/response.
- `photos.go` — form aggregation from `search.PhotoResults` and bulk selection helpers.
- `convert.go` — translates `PhotosForm` into `form.Photo` instances for persistence.
- `apply_albums.go` / `apply_labels.go` — album and label mutation helpers shared across API endpoints.
- `save.go` — differential persistence, `PreparePhotoSaveRequests`, `PrepareAndSavePhotos`, `NewPhotoSaveRequest`, `PhotoSaveRequest`, background worker triggers.
- `save_photo.go` — `savePhoto` applies a single request, compares old/new values, and writes only the changed columns (indirectly invoked by `SavePhotos`).
- `save_resolve.go` — album/label title resolution helpers that run before persistence so per-photo work only receives resolved UIDs.
- `save_backup.go` — YAML backup synchronisation for albums whenever batch edits touch them and backups are enabled.
- `datelogic.go` — helpers for reconciling time zones and date parts when the UI only supplies partial values.
- `values.go` — typed wrappers for request fields (value + action + mixed flag).

### Next Steps

- [ ] Surface the dormant EXIF controls (ISO, focal length, lens/camera IDs) in the frontend and wire them to `PhotosForm` once the UI/UX is ready.
- [ ] Evaluate batching `ApplyLabels`/`ApplyAlbums` at the SQL level for very large selections while keeping validation safeguards.
- [ ] Document worker SLA guarantees (metadata refresh latency, label indexing) once observability data is available.
- [ ] Keep an eye on MySQL deadlock frequency with the new retry/backoff helpers and bump `deadlockRetryDelay` if busy installations still hit user-visible 500s.
