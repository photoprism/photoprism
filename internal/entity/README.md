## PhotoPrism — Database Entities

**Last Updated:** September 23, 2026

### Overview

`internal/entity` holds the GORM models (Photo, File, Album, Label, Face, User, Client, Session, Service, Marker, …), their query and create/update helpers, the test fixtures (`*_fixtures.go`), and the migration helpers under `migrate/`. Models map to the database via GORM v1 (`github.com/jinzhu/gorm`) and are shared by the API, workers, and CLI.

### Account & Session Caches

`User.Save` evicts that account's cached sessions and WebDAV authentication entries after a successful database save. `FlushUserSessionCache` matches the user UID and leaves other users' entries and persisted credentials intact. `FlushSessionCache` clears both caches when a global refresh is required.

WebDAV uses `CachedWebDAVUser` and `CacheWebDAVUser` for its one-minute credential cache. Authentication captures `CurrentAuthCacheGeneration` before loading the user or session, and insertion checks that generation under the same short mutex used by invalidation. A superseded result is not cached, but the request is not canceled. Per-user invalidation leaves other accounts' pending cache writes valid; a global flush invalidates every older snapshot and clears the per-user revision table. Account changes are therefore handled at the same persistence boundary as the general session cache. Sessions created or loaded through the entity helpers retain their generation so an in-flight object cannot repopulate the general cache after eviction. A raw untracked record may be cached only before its user is resolved. This is process-local invalidation; writes through another process or directly to the database do not signal a running server.

A session read from or written to the database is only ever updated afterwards: `Session.Save` writes it through `Update`, and inserts only a session that was never stored. `Session.VerifyStored` confirms that a session's row still exists; app-password sign-in, the OAuth password and session grants, and `OIDCSessionEligible` call it before deriving a credential from a session, and a missing row evicts the session from the session and preview token caches.

### Update & Save Helpers

Prefer these helpers over hand-written GORM calls when a model is written as a whole:

- `ModelValues(m, omit...)` returns the exported column values of a model as `Values`, keyed by field name and including zero values. It skips `CreatedAt`/`UpdatedAt`, relations, maps and other non-byte slices, but keeps byte slices such as `json.RawMessage` columns. Omitted fields are returned separately, which is how the key values reach `Update`.
- `Update(m, keys...)` updates an existing row with every value from `ModelValues`, zero values included, and never inserts; GORM sets `updated_at`. When the update changes no row, it counts the rows matching the keys and returns an error unless exactly one exists, since MariaDB reports changed rather than matched rows.
- `Save(m, keys...)` tries `Update` first and falls back to GORM's `Save`, which inserts a missing row.

GORM's `Updates` with a struct skips zero values, so a field reset to its zero value is not written that way; pass a `Values` map or use `Update`. `Report()` output for users, clients, and sessions is built from `ModelValues`.

### Timestamps

Created and updated timestamps are stored as SQL `DATETIME` **without fractional seconds** (`DATETIME_PRECISION = 0`). To keep in-memory and persisted values in sync, the package sets GORM's timestamp source to second precision in `db.go`:

```go
gorm.NowFunc = Now // entity.Now() == UTC().Truncate(time.Second)
```

Time helpers in `entity_time.go`:

- `UTC()` — current time in UTC, full sub-second precision. Use for elapsed-time measurements, not for values that get persisted.
- `Now()` — UTC truncated to whole seconds. This is what GORM writes to `created_at` / `updated_at`.
- `TimeStamp()` — pointer to `Now()`, for nullable `*time.Time` columns.
- `Time(s)` — parses an RFC 3339 string to a second-precision UTC time, or `nil`.

Implications:

- **Do not rely on sub-second ordering of persisted timestamps.** Two rows created and updated within the same wall-clock second compare **equal**, so `created_at` / `updated_at` cannot disambiguate them. There is no monotonic auto-increment ID on UID-keyed models (e.g. `Client`), so there is no reliable intra-second tiebreaker — give rows distinct times when ordering must be deterministic.
- Because both SQLite and MariaDB now receive second-precision values, timestamp behavior is **identical across drivers**. A timestamp assertion that passes on SQLite will pass on MariaDB.

When a test needs to prove a write advanced a timestamp, prefer one of:

- Seed the starting value clearly in the past (e.g. `Now().Add(-time.Hour)`) and assert the new value is greater. This stays meaningful and distinguishes a real bump from a no-op.
- Compare with `Time.Sub()` and assert the difference falls in a sane range, rather than a strict `Before`/`After`. A same-second save legitimately yields a zero delta:

  ```go
  elapsed := after.Sub(before)
  assert.GreaterOrEqual(t, elapsed, time.Duration(0))
  assert.Less(t, elapsed, time.Minute)
  ```

### Testing

Tests default to SQLite. To exercise the models against MariaDB (which is stricter and is the production database for some subsystems such as the cluster registry):

```bash
mariadb < scripts/sql/reset-acceptance.sql
PHOTOPRISM_TEST_DRIVER="mysql" \
PHOTOPRISM_TEST_DSN="root:photoprism@tcp(mariadb:4001)/acceptance?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true" \
go test ./internal/entity/... -count=1 -tags="slow,develop"
```

`make test-mariadb` runs the whole backend suite this way. Each package gets a **database of its own**, named after its source directory (`acceptance_query_…`) and created on demand by `entity.TestDbDSN`, which mirrors the file-per-package isolation SQLite gets for free. Without it the packages would share one schema, and since every `TestMain` truncates the tables and re-seeds the fixtures, they would pull the database out from under each other as soon as `go test` runs them in parallel. `make reset-acceptance` drops these databases along with `acceptance` itself. The configured account needs `CREATE` privileges for this; without them the packages fall back to sharing `acceptance` (logged as a warning) and must not run in parallel. A test that needs a database of its own **within** a package must set `PHOTOPRISM_TEST_DRIVER` as well as `PHOTOPRISM_TEST_DSN`, as a SQLite path alone is parsed as a MySQL DSN and aborts the package.

MariaDB strict mode rejects inserts that SQLite quietly accepts, so a test that only ran on SQLite can fail here:

- **Primary keys must be set.** An empty PK (`""` UID, zero ID) triggers `Error 1364: Field '<col>' doesn't have a default value`. Use a valid ID/UID, not a placeholder like `"1234"`.
- **Values must fit the column.** Oversized strings give `Error 1406: Data too long`; out-of-range integers give `Error 1264: Out of range value` (e.g. `photo_id` is `INT UNSIGNED`, max 4294967295).
- UID format (see `pkg/rnd/uid.go`): a one-byte prefix + 6 base36 time chars + 9 base36 random, 16 chars total (`p…` photo, `a…` album, `c…` client, `u…` user, `l…` label, `d…` folder). Reuse existing fixtures for foreign-key safety; use a throwaway but in-range value only where a real reference would overwrite seeded data (e.g. a synthetic `photo_id` so a Details row does not attach to a real photo).
- Fixtures live in `*_fixtures.go`, but some join rows are created **indirectly** from a parent fixture's embedded slice (e.g. a `photos_labels` row from a `Photo` fixture's `Labels`). Verify a combination is free against the **seeded database**, not just the fixtures file.
- Face and marker embeddings are the exception to "fixtures are literals": `GenerateFaceFixtureVectors` (in `face_fixtures_vectors.go`) generates them for the configured embedding model just before the rows are written, because a stored vector has one model's width and no usable provenance under any other. `faceFixtureSeeds` gives each fixture person a centroid, and `markerFixtureVectors` places each face marker at a fraction of the distance its cluster accepts, so the geometry survives both a change of model and a recalibration.
- `List`-style global queries (`WHERE … <> ''` with no per-test scope) see everything the package has written: rows from other tests in the same package leak in, so a `len(list) == N` assertion that holds against a per-test SQLite file can fail on MariaDB, where the whole package shares one database.
- **Sort order is collation-dependent.** `utf8mb4_unicode_ci` sorts case-insensitively and weights punctuation by Unicode rules, while SQLite compares byte values, so `ORDER BY` on a text column yields a different sequence. Give rows a deterministic tiebreaker, or assert per dialect (`entity.Db().Dialect().GetName()`).
- **Generated IDs restart at 1.** `Tables.Truncate` issues `TRUNCATE` where supported, which resets `AUTO_INCREMENT`, so a fixture without an explicit ID gets the same value it would in a fresh database. Plain `DELETE` would not, and IDs would drift with every reset.

### Collation & Emoji

MariaDB's `utf8mb4_unicode_ci` assigns most emoji the **same collation weight**, so an SQL `=`, `<>`, or `LIKE` on a `utf8mb4` column treats distinct emoji as equal (e.g. `test/🪞` matches `test/🎃`). SQLite compares text byte-exact, so this only reproduces on MariaDB.

- `utf8mb4` columns that collapse: `albums.album_title`, display/name text (`*_name`, `*_title`).
- `VARBINARY` columns that stay byte-exact: `albums.album_slug`, `albums.album_filter`, `albums.album_path`, `photos.photo_path`, and every `*_uid`. A `utf8mb4` column compared against a `VARBINARY` column is byte-exact (the binary operand wins).

Byte-exact also means **case-sensitive**, which is the one place `VARBINARY` bites on a search path: SQLite's `LIKE` folds ASCII case, so `album_slug LIKE 'Forrest%'` finds the `forrest` slug there but nothing on MariaDB. Slugs are always generated lowercase, so fold the pattern before comparing (`strings.ToLower`), as the album filter in `search.searchPhotos` does.

The durable fix for an identity/path column is to make it `VARBINARY` — `album_path` is `VARBINARY(1024)` so it matches `photos.photo_path` and `album_path = ?` lookups are byte-exact at the database. Where a `utf8mb4` column must stay, keep the SQL but re-verify the match byte-exact in Go before accepting it (see `FindFolderAlbum` / `findFolderAlbumByPath`, whose Go re-check is retained as defense-in-depth even now that `album_path` is `VARBINARY`). For self-join SQL where a Go re-check is awkward, `HEX(col) = HEX(col)` compares byte-exact on both MariaDB and SQLite. Legacy folder slugs drop emoji entirely (`slug.Make("ins/🪞") == "ins"`) and long paths truncate to `ClipSlug` runes, so distinct folders can still collide on `album_slug`; folder albums are therefore deduplicated by `album_filter` (the byte-exact serialized path), not by slug (see `query.RemoveDuplicateMoments`).

### Per-Dialect SQL

Several helpers under `query/` write raw SQL because GORM v1 cannot express the statement. Where one has to differ per driver — MariaDB's multi-table `UPDATE … JOIN` against SQLite's correlated subquery — keep the branches to what the dialects genuinely require and every selection rule identical, ordering included. A statement that can be written once for both should be: divergence between two hand-written branches is invisible to a SQLite-only run, and only `make test-mariadb` exercises the other.

An `ORDER BY` in such a statement needs a total order. A prefix that leaves ties resolves them by plan, and `GROUP BY` guarantees no order at all on MariaDB, so a result that depends on either can change with a version or plan change and no code change. End the ordering on a unique column, as `query.UpdateSubjectCovers` does on `markers.marker_uid`.

### VARBINARY Index Prefix Limit

InnoDB caps an index key prefix at **767 bytes** on the `COMPACT`/`REDUNDANT` row formats, and only allows up to 3072 bytes on `DYNAMIC`/`COMPRESSED`. On a `VARBINARY` column the prefix is counted in **bytes** (on `utf8mb4` it is counted in characters, i.e. up to 4 bytes each), so converting a long text column to `VARBINARY` can push an existing prefix index over the limit on older or non-`DYNAMIC` installs. Keep prefix indexes on long `VARBINARY` path/filter columns at **≤ 767 bytes**; the project convention is **512** (`albums.album_filter(512)`, `albums.album_path(512)`). A prefix index only narrows candidate rows — the full-column comparison stays exact — so a shorter prefix costs nothing for correctness.

### File Export Eligibility

`File.Exportable` applies the YAML export policy after download admission and row visibility.
YAML is recognized by its filename or recorded file type, including `.yml` and `.yaml`.
An identified registered reader or client must have effective read permission on photos or
files. Share-link visitors and unidentified downloads do not export YAML. Registered guests,
Contributors, and files-only readers keep their existing eligible downloads. Other file
formats are unchanged. `SelectedFilesForSession` uses the same decision for download selections as direct downloads;
`SelectedFiles` remains unrestricted for internal workflows.

### Ignored Transfers

Excluded queued service shares use `FileShareIgnore`, with no retry error. Automatic sync
uses `FileSyncIgnore`; upload dispositions have a per-FileID internal key under the reserved
`.photoprism/sync` namespace so matching filenames in different roots cannot collide. These
keys are state identifiers, never remote transfer destinations. Existing deduplication rules
and retry handling for genuine transfer failures remain unchanged.
