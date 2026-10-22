## PhotoPrism — Database Entities

**Last Updated:** October 21, 2026

### Overview

`internal/entity` holds the GORM models (Photo, File, Album, Label, Face, User, Client, Session, Service, Marker, …), their query and create/update helpers, the test fixtures (`*_fixtures.go`), and the migration helpers under `migrate/`. Models map to the database via GORM v2 (`gorm.io/gorm` with the `gorm.io/driver/mysql`, `gorm.io/driver/sqlite` and `gorm.io/driver/postgres` dialectors) and are shared by the API, workers, and CLI. PostgreSQL support is experimental and tested with PostgreSQL 18; startup refuses versions older than 16.

### Account & Session Caches

`User.Save` evicts that account's cached sessions and WebDAV authentication entries after a successful database save. `FlushUserSessionCache` matches the user UID and leaves other users' entries and persisted credentials intact. `FlushSessionCache` clears both caches when a global refresh is required. `DeleteSession` evicts the session again after its row is deleted and remembers its ID with a new cache version for a minute, so a lookup that read the row earlier does not cache it, and a lookup older than that minute caches nothing; other cache entries are unaffected.

WebDAV uses `CachedWebDAVUser` and `CacheWebDAVUser` for its one-minute credential cache. Authentication captures `CurrentAuthCacheGeneration` before loading the user or session, and insertion checks that generation under the same short mutex used by invalidation. A superseded result is not cached, but the request is not canceled. Per-user invalidation leaves other accounts' pending cache writes valid; a global flush invalidates every older snapshot and clears the per-user revision table. Account changes are therefore handled at the same persistence boundary as the general session cache. Sessions created or loaded through the entity helpers retain their generation so an in-flight object cannot repopulate the general cache after eviction. A raw untracked record may be cached only before its user is resolved. This is process-local invalidation; writes through another process or directly to the database do not signal a running server.

`CachedClientByUID` caches client lookups on the request path, such as `Session.ClientDisabled`, `Session.GetClient` and the client role of a session, for one minute (`ClientCacheDuration`). It returns an independent copy without the account, the parsed data or the matched secret hash, caches neither a missing row nor an error, and uses its own version guard, so a read that started before a write is never stored. The client write methods flush the client's entry after writing, `User.Rename` and `FlushCaches` flush them all, and a change written by another process is seen once the entry expires. Lookups that act on a client, such as the client credentials grant, admin endpoints and the registry, read the database with `FindClientByUID`; the Portal authorize endpoint and the app flow (`oauth.FindClient`) read the cached client, so an access change made on another replica applies there within a minute. A client session loaded from the database gets its client assigned before it enters the session cache (`Session.resolveClient`), which applies the client's scope and owner but keeps the auth provider and method stored with the session, as `GetClient` and the client role do, and keeps that copy until it is reloaded; if the client or its owner cannot be read, the session is not cached, and a session of a client whose owner no longer exists is refused on each request. `SetClient` keeps the client's owner when the owner's account cannot be read.

A session read from or written to the database is only ever updated afterwards: `Session.Save` writes it through `Update`, and inserts only a session that was never stored. `Session.VerifyStored` confirms that a session's row still exists; app-password sign-in, the OAuth password and session grants, and `OIDCSessionEligible` call it before deriving a credential from a session, and a missing row evicts the session from the session and preview token caches.

### Update & Save Helpers

Prefer these helpers over hand-written GORM calls when a model is written as a whole:

- `ModelValues(m, omit...)` returns the exported column values of a model as `Values`, keyed by field name and including zero values. It skips `CreatedAt`/`UpdatedAt`, relations, maps and other non-byte slices, but keeps byte slices such as `json.RawMessage` columns. Omitted fields are returned separately, which is how the key values reach `Update`.
- `Update(m, keys...)` updates an existing row with every value from `ModelValues`, zero values included, and never inserts; GORM sets `updated_at`. When the update changes no row, it counts the rows matching the keys and returns an error unless exactly one exists, since MariaDB reports changed rather than matched rows.
- `Save(m, keys...)` tries `Update` first and falls back to GORM's `Save`, which inserts a missing row.

GORM's `Updates` with a struct skips zero values, so a field reset to its zero value is not written that way; pass a `Values` map or use `Update`. `Report()` output for users, clients, and sessions is built from `ModelValues`.

### ORM Configuration

`OpenDb` opens connections with `GormConfig()` and `GormDialector()`, which keep the existing schema and the write semantics the code relies on:

- Struct tags state the existing column types and index names (`size:32` on integers, `type:bytes;size:<n>` for `VARBINARY`, named unique indexes), and the MySQL dialector creates untagged strings as `VARCHAR(255)` and times as `DATETIME` without fractional seconds. On SQLite, strings are `varchar(n)` as before and `VARBINARY` columns are `blob(0)`: blob affinity, so numeric-looking values stay strings, and the row width estimate of the former type, as SQLite plans joins with these estimates. On MySQL, entity tables are created with `ENGINE=InnoDB` and `utf8mb4`, using the most common collation of the existing PhotoPrism tables, preferring `utf8mb4_unicode_ci` and then other `utf8mb4` collations on a tie, or `utf8mb4_unicode_ci` in a new database and if the result is not `utf8mb4`. The `migrations` and `versions` tables are created first with the database default and are not compared with other tables, so neither decision nor warning includes them. Every migration run logs a warning naming tables that do not use InnoDB, InnoDB tables with the `Compact` or `Redundant` row format, whose index entries are limited to 767 bytes (with the `ALTER TABLE <table> ROW_FORMAT=DYNAMIC;` statement that converts them; PhotoPrism does not convert them itself, as rebuilding a large table can take long), and the tables to convert when a table does not use `utf8mb4` or the collations differ, with the `ALTER TABLE` statement that converts them to the collation new tables get.
- Relation fields carry `-:migration`, so migrating one table never creates another, and `DisableAssociationSaves` removes the association callbacks for the whole connection: related rows are saved explicitly, e.g. `Album.Create` and `Album.Save` save the album's photo links. Related rows such as cameras, lenses and places are shared by many photos, so a photo save must not write its possibly outdated copies of them, and a photo's details are saved with their own retry and error (`ErrDetailsNotSaved`). The `Association` API is for reading only: `Append` and `Replace` save through the removed callbacks, so `Append` writes nothing without an error, and `Replace` adds nothing but removes the existing links (`TestDisableAssociationSaves`). On a has-many relation such as `Album.Photos`, `Replace`, `Delete` and `Clear` set the foreign key to `NULL`, which SQLite accepts and MariaDB and PostgreSQL reject. A `Select` list also limits the model's own columns, so `Select("Camera").Save(&photo)` writes no photo column.
- Two callbacks keep `Save` semantics: a blank `CreatedAt` and blank belongs-to keys with a default are left out of the statement, and a `Save` that changes no row creates the row with hooks unless it already exists.
- Updates and deletes without conditions fail with `gorm.ErrMissingWhereClause`, also when a model has a zero primary key; bulk statements state `Where("1 = 1")`. A model with a composite key whose first key part is set also binds its zero key parts, in updates and queries. `UnscopedDb()` returns a new session, so its conditions do not carry over to the next statement.
- Each condition of a `WHERE` clause is put in parentheses when there are several (`buildWhere`), so a condition with `OR` does not combine with its neighbors.
- `DeletedAt` is a `sql.NullTime` with the soft-delete clauses of `gorm.DeletedAt`; it encodes as `null` in JSON and as a timestamp in YAML.
- On MariaDB, auto-migration only lengthens a shorter column of the same base type (keeping its character set and collation), sets differing defaults in place (not a negative default on an unsigned column), and lets a `NOT NULL` column accept `NULL` (`mysqlMigrator.AlterColumn`). A longer column or another base type is kept and logged, as databases created by older releases may hold values that would be truncated or converted; any other change, such as `NOT NULL` or `UNSIGNED`, needs a SQL migration. On PostgreSQL, whose driver does not migrate primary key columns, `postgresMigrator.MigrateColumn` lengthens a `varchar` primary key column to its field size and never shortens one. Preloads of an unscoped query are unscoped too, so they add `deleted_at IS NULL` for soft-deleted models.

`migrate.ConvertSQLiteSchema` rebuilds SQLite tables created with the previous column types before auto-migration, once per table: it creates the table from its model, copies the rows with numeric values in text columns converted to text, and keeps the `sqlite_sequence` counter. The previous types stored numeric-looking strings as numbers (`0001` as `1`), so photo paths and names, top-level folder paths, folder album paths, and the slugs of albums, labels, subjects, cameras, lenses and share links are restored from the file, folder and display names they were derived from, when exactly one of those names has the same numeric value. Share link tokens cannot be restored; on SQLite, an all-digit token is also found without its leading zeros (`FindLinks`).

If auto-migration fails because a unique index cannot be created over duplicate values, the system log names the table and index, with the query that lists the duplicates.

### Label Count Refresh

`UpdateLabelCounts` keeps each driver's counting query and updates the refresh timestamp only after
success. Its MySQL and PostgreSQL writes use `RetryDeadlock`, shared with the label writes described
below: at most three attempts with bounded backoff for recognized deadlock errors. Other errors return
immediately. Retries apply to the individual write, not to the entire HTTP handler or its preceding operations.

### Write Retries

`Photo.SaveDetails` writes through `RetryLock`, which retries once after any lock error, including lock wait
timeouts and a locked SQLite database. It returns an error wrapping `ErrDetailsNotSaved` when the details could
not be stored; creating a missing row counts as stored. `Photo.Save` and `Photo.Create` keep the stored
photo row in that case. The indexer then adds or updates the file with a modification time of -1, so the next
index run indexes it again and saves the details, while edits through the API return the error, except unstacking, which keeps the new photo and saves its details when it reindexes the moved files.

`PhotoLabel.Update`, `PhotoLabel.Updates`, `PhotoLabel.Delete` and `Label.Rename` write through `RetryDeadlock`,
since a label count refresh started by an earlier edit can hold locks on `labels` and `photos_labels` that a
concurrent label edit needs in the opposite order. Batch edits and the photo label API call these methods
directly, so they must not wrap them in another retry. Lock wait timeouts are not retried there: the statement has
already waited for the configured timeout, so a retry would double how long the request takes. The photo row
written by `Photo.SaveLabels` goes through `Save`, which replays a failed update as an upsert, and once more
after a lock error.

`Details` removes control characters from its text (keeping line breaks in notes, copyright and license) and
clips it to the column sizes before every write, since MariaDB rejects longer values in strict
mode while SQLite stores them: keywords are cut after the last whole keyword that fits, and source names longer
than their column are cleared.

### Timestamps

Created and updated timestamps are stored as SQL `DATETIME` **without fractional seconds** (`DATETIME_PRECISION = 0`). To keep in-memory and persisted values in sync, `GormConfig()` in `db.go` sets GORM's timestamp source to second precision:

```go
NowFunc: Now, // entity.Now() == UTC().Truncate(time.Second)
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
mariadb < scripts/sql/reset-testdb.sql
PHOTOPRISM_TEST_DRIVER="mysql" \
PHOTOPRISM_TEST_DSN="root:photoprism@tcp(mariadb:4001)/testdb?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true" \
go test ./internal/entity/... -count=1 -tags="slow,develop"
```

`make test-mariadb` runs the whole backend suite this way. Each package gets a **database of its own**, named after its source directory (`testdb_query_…`) and created on demand by `entity.TestDbDSN`, which mirrors the file-per-package isolation SQLite gets for free. Without it the packages would share one schema, and since every `TestMain` truncates the tables and re-seeds the fixtures, they would pull the database out from under each other as soon as `go test` runs them in parallel. `make reset-mariadb-testdb` drops these databases and recreates `testdb` itself. The configured account needs `CREATE` privileges for this; without them the packages fall back to sharing `testdb` (logged as a warning) and must not run in parallel. A test that needs a database of its own **within** a package must set `PHOTOPRISM_TEST_DRIVER` as well as `PHOTOPRISM_TEST_DSN`, as a SQLite path alone is parsed as a MySQL DSN and aborts the package.

`make test-postgres` runs the suite against the PostgreSQL service of the development environment (`docker compose --profile postgres up -d`) in the same way, with `PHOTOPRISM_TEST_DRIVER="postgres"` and a `postgres://` connection URI whose user may create databases; `make reset-postgres-testdb` drops the per-package databases.

MariaDB strict mode rejects inserts that SQLite quietly accepts, so a test that only ran on SQLite can fail here:

- **Primary keys must be set.** An empty PK (`""` UID, zero ID) triggers `Error 1364: Field '<col>' doesn't have a default value`. Use a valid ID/UID, not a placeholder like `"1234"`.
- **Values must fit the column.** Oversized strings give `Error 1406: Data too long`; out-of-range integers give `Error 1264: Out of range value` (e.g. `photo_id` is `INT UNSIGNED`, max 4294967295).
- UID format (see `pkg/rnd/uid.go`): a one-byte prefix + 6 base36 time chars + 9 base36 random, 16 chars total (`p…` photo, `a…` album, `c…` client, `u…` user, `l…` label, `d…` folder). Reuse existing fixtures for foreign-key safety; use a throwaway but in-range value only where a real reference would overwrite seeded data (e.g. a synthetic `photo_id` so a Details row does not attach to a real photo).
- Fixtures live in `*_fixtures.go`, but some join rows are created **indirectly** from a parent fixture's embedded slice (e.g. a `photos_labels` row from a `Photo` fixture's `Labels`). Verify a combination is free against the **seeded database**, not just the fixtures file.
- Face and marker embeddings are the exception to "fixtures are literals": `GenerateFaceFixtureVectors` (in `face_fixtures_vectors.go`) generates them for the configured embedding model just before the rows are written, because a stored vector has one model's width and no usable provenance under any other. `faceFixtureSeeds` gives each fixture person a centroid, and `markerFixtureVectors` places each face marker at a fraction of the distance its cluster accepts, so the geometry survives both a change of model and a recalibration.
- `List`-style global queries (`WHERE … <> ''` with no per-test scope) see everything the package has written: rows from other tests in the same package leak in, so a `len(list) == N` assertion that holds against a per-test SQLite file can fail on MariaDB, where the whole package shares one database.
- **Sort order is collation-dependent.** `utf8mb4_unicode_ci` sorts case-insensitively and weights punctuation by Unicode rules, while SQLite compares byte values, so `ORDER BY` on a text column yields a different sequence. Give rows a deterministic tiebreaker, or assert per dialect (`entity.DbDialect()`).
- **PostgreSQL differs again** (experimental): `VARBINARY` string fields are `varchar(n) COLLATE "C"` there, so they compare and sort byte by byte like on MariaDB but cannot hold invalid UTF-8 or NUL bytes; text columns compare case-sensitively with `=` and sort by the database collation, which is why `entity.TestDbDSN` creates the test databases with an ICU `und` locale; NULL sorts last in ascending order unless an order string adds `NULLS FIRST` (`search.OrderExprNulls` for lists sorted by nullable columns); and a failed statement aborts the rest of its transaction, so fixtures and transactions must not rely on an ignored insert error.
- **Affected rows differ.** An `UPDATE` reports the rows it changed on MariaDB, but the rows it matched on SQLite and PostgreSQL, so a write that sets a value a row already has returns `0` on MariaDB only. A conditional write that tests `RowsAffected` must hold on all three.
- **Generated IDs restart at 1.** On MySQL/MariaDB, `Tables.Truncate` deletes the rows and then resets `AUTO_INCREMENT` on the tables that have such a column, so a default fixture without an explicit ID, such as `UnknownCamera` and `UnknownLens`, gets the same value it would in a fresh database. `TRUNCATE` would do the same, but it is a DDL statement and several times slower per reset. SQLite keeps its counters, so tests compare against `UnknownCamera.ID` and `UnknownLens.ID` rather than a literal `1`.

### Shape & Statement Pins

Some tests pin behavior an ORM upgrade could change without breaking anything else, and are meant to stay unchanged across one:

- `entity_shape_test.go`: how each model encodes `DeletedAt` in JSON and YAML, that generated timestamps are whole seconds in memory and in the database, that cleared values persist (including fields with a `default` tag), that blank fields with a `default` tag get the default on create, and how a photo's preloaded files, albums and labels encode.
- `results_shape_test.go` in `search` and `query`: searches without matches encode as `[]`. `client_config_shape_test.go` in `internal/config` pins the client config lists of an empty library, where album categories and years are `null`.
- `TestPhoto_SaveStatements` (in `entity_shape_test.go`) and `TestIndex_UserMediaFileStatements` (in `internal/photoprism`) set an upper bound for the SQL statements an operation issues, per driver. They count through `internal/entity/sqlcount`, which wraps the `database/sql` driver, so the count does not depend on GORM callbacks or logging. Swap the counted provider in with `sqlcount.OpenGorm`, flush the entity caches, and restore the previous provider afterwards. Import `sqlcount` from tests only.

Lower a ceiling when a change reduces the statements; raising one needs an explained cause.

### Colors, Chroma & Diff

`photos.photo_color`, `files.file_chroma` and `files.file_diff` store measured values that are never 0, so their `default:-1` tags cannot turn a measurement into "unknown" when a row is created: GORM leaves a blank field with a `default` tag out of the `INSERT`.

- `photo_color` is the ID of the main color from `pkg/media/colors`, where `Black` is 16; color hex strings such as `file_colors` still render it as `0`.
- `file_chroma` is the rounded colorfulness in percent plus one, clamped to 100, so 1 means monochrome and 2 nearly gray. The `mono` search filter matches 1; `chroma:N` takes a percentage, so `chroma:5` matches 2 to 6. Migration `20261011-000001` shifts the percentages stored by earlier versions and maps 0 to 1, or to -1 without colors; libraries that ran an earlier preview keep 1 for both 0 and 1 percent until their files are indexed again.
- `file_diff` is a 10-bit perceptual hash from 512 to 1023.
- -1 means unknown. `NewUserPhoto` and the indexer set it before the values are computed. `entity_colors_test.go` pins the create path and the migrations that converted stored values.

### Collation & Emoji

MariaDB's `utf8mb4_unicode_ci` assigns most emoji the **same collation weight**, so an SQL `=`, `<>`, or `LIKE` on a `utf8mb4` column treats distinct emoji as equal (e.g. `test/🪞` matches `test/🎃`). SQLite compares text byte-exact, so this only reproduces on MariaDB.

- `utf8mb4` columns that collapse: `albums.album_title`, display/name text (`*_name`, `*_title`).
- `VARBINARY` columns that stay byte-exact: `albums.album_slug`, `albums.album_filter`, `albums.album_path`, `photos.photo_path`, and every `*_uid`. A `utf8mb4` column compared against a `VARBINARY` column is byte-exact (the binary operand wins).

Byte-exact also means **case-sensitive**, which is the one place `VARBINARY` bites on a search path: SQLite's `LIKE` folds ASCII case, so `album_slug LIKE 'Forrest%'` finds the `forrest` slug there but nothing on MariaDB. Slugs are always generated lowercase, so fold the pattern before comparing (`strings.ToLower`), as the album filter in `search.searchPhotos` does.

A value bound to `LIKE` is still a pattern: escape it with `clean.SqlLike` and use a condition that declares the escape character (`clean.SqlLikeCond`, `clean.SqlLikeAny`). A path prefix check needs `clean.SqlPrefixCond` with `clean.SqlPrefixArgs`, which adds a byte-exact comparison, because an escaped `LIKE` still folds ASCII case on SQLite. These helpers take the SQL dialect (`DbDialect()`) and whether the column is `VARBINARY`; `query.LikeCond` and the `search` builders take only the flag and resolve the dialect themselves. MySQL and SQLite produce the same condition either way; on PostgreSQL the flag selects `LIKE` for binary columns, which compare byte by byte, and `ILIKE` for text columns. Pass text and binary columns to separate `SqlLikeAny` calls.

The durable fix for an identity/path column is to make it `VARBINARY` — `album_path` is `VARBINARY(1024)` so it matches `photos.photo_path` and `album_path = ?` lookups are byte-exact at the database. Where a `utf8mb4` column must stay, keep the SQL but re-verify the match byte-exact in Go before accepting it (see `FindFolderAlbum` / `findFolderAlbumByPath`, whose Go re-check is retained as defense-in-depth even now that `album_path` is `VARBINARY`). For self-join SQL where a Go re-check is awkward, `HEX(col) = HEX(col)` compares byte-exact on both MariaDB and SQLite (PostgreSQL has no `HEX()`; its binary strings already compare byte by byte). Legacy folder slugs drop emoji entirely (`slug.Make("ins/🪞") == "ins"`) and long paths truncate to `ClipSlug` runes, so distinct folders can still collide on `album_slug`; folder albums are therefore deduplicated by `album_filter` (the byte-exact serialized path), not by slug (see `query.RemoveDuplicateMoments`).

### Per-Dialect SQL

Several helpers under `query/` write raw SQL because GORM cannot express the statement. Where one has to differ per driver — MariaDB's multi-table `UPDATE … JOIN` against SQLite's correlated subquery — keep the branches to what the dialects genuinely require and every selection rule identical, ordering included. A statement that can be written once for both should be: divergence between two hand-written branches is invisible to a SQLite-only run, and only `make test-mariadb` exercises the other.

Branch on `entity.DbDialect()` (or `query.DbDialect()`), which returns one of the `dsn.Dialect*` constants (`mysql`, `sqlite`, `postgres`) whatever name the ORM reports, and compare it with those constants rather than the configured driver names such as `sqlite3`. `entity.IsDialect` accepts either.

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

### Copying Between Databases

`Tables.Transfer` (used by `photoprism database transfer`) replaces the rows of every registered table except `migrations` and `versions` in one database with those of another, which may be of a different kind. Both schemas must be migrated by the same version; `Config.MigrateDbSchema` does that for a target without creating default records. Rows are read from one snapshot of the source, per table in batches ordered by primary key, and written with a plain `INSERT ... VALUES` built from the struct fields, not with `Create`: GORM would run hooks and replace zero values of fields with a `default` tag, e.g. store `zz` for an empty `album_country`. Afterwards PostgreSQL sequences are synced and the row counts are compared with the source, in the snapshot and after it; if the transfer does not complete, the target's tables are cleared again. `SameDatabase` tells whether two connections reach the same database, so a copy onto the source can be refused.

`TestTables_Transfer` copies the package test database into a SQLite file and, when their services are reachable, into MariaDB and PostgreSQL databases (`PHOTOPRISM_TEST_DSN_MARIADB` and `PHOTOPRISM_TEST_DSN_POSTGRES` override the development environment's), and compares every table row by row, so the three test drivers together cover each pair.

### Ignored Transfers

Excluded queued service shares use `FileShareIgnore`, with no retry error. Automatic sync
uses `FileSyncIgnore`; upload dispositions have a per-FileID internal key under the reserved
`.photoprism/sync` namespace so matching filenames in different roots cannot collide. Files whose
remote name exceeds `RemoteNameMaxLen` are recorded under `.photoprism/sync/long/<crc32 of the sync
path>/`, and `ResetFileSyncLongNames` removes those of other sync paths, so the files are checked
again after the sync path changes. Uploads that failed or stalled are recorded under `.photoprism/sync/upload/` (`FailFileSyncUpload`) with their error
count, and refresh marks them `FileSyncRetry` (`RetryFileSyncUploads`), so the files are tried once per sync cycle up to the retry limit. These keys are state identifiers,
never remote transfer destinations.
