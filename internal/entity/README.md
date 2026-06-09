## PhotoPrism — Database Entities

**Last Updated:** June 1, 2026

### Overview

`internal/entity` holds the GORM models (Photo, File, Album, Label, Face, User, Client, Session, Service, Marker, …), their query and create/update helpers, the test fixtures (`*_fixtures.go`), and the migration helpers under `migrate/`. Models map to the database via GORM v1 (`github.com/jinzhu/gorm`) and are shared by the API, workers, and CLI.

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
mysql < scripts/sql/reset-acceptance.sql
PHOTOPRISM_TEST_DRIVER="mysql" \
PHOTOPRISM_TEST_DSN="root:photoprism@tcp(mariadb:4001)/acceptance?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true" \
go test ./internal/entity/... -count=1 -tags="slow,develop"
```

MariaDB strict mode rejects inserts that SQLite quietly accepts, so a test that only ran on SQLite can fail here:

- **Primary keys must be set.** An empty PK (`""` UID, zero ID) triggers `Error 1364: Field '<col>' doesn't have a default value`. Use a valid ID/UID, not a placeholder like `"1234"`.
- **Values must fit the column.** Oversized strings give `Error 1406: Data too long`; out-of-range integers give `Error 1264: Out of range value` (e.g. `photo_id` is `INT UNSIGNED`, max 4294967295).
- UID format (see `pkg/rnd/uid.go`): a one-byte prefix + 6 base36 time chars + 9 base36 random, 16 chars total (`p…` photo, `a…` album, `c…` client, `u…` user, `l…` label). Reuse existing fixtures for foreign-key safety; use a throwaway but in-range value only where a real reference would overwrite seeded data (e.g. a synthetic `photo_id` so a Details row does not attach to a real photo).
- Fixtures live in `*_fixtures.go`, but some join rows are created **indirectly** from a parent fixture's embedded slice (e.g. a `photos_labels` row from a `Photo` fixture's `Labels`). Verify a combination is free against the **seeded database**, not just the fixtures file.
- `List`-style global queries (`WHERE … <> ''` with no per-test scope) are not isolated on the shared `acceptance` database: rows from other tests in the same run leak in, so a `len(list) == N` assertion that holds on a per-test SQLite file can fail on MariaDB.

### Collation & Emoji

MariaDB's `utf8mb4_unicode_ci` assigns most emoji the **same collation weight**, so an SQL `=`, `<>`, or `LIKE` on a `utf8mb4` column treats distinct emoji as equal (e.g. `test/🪞` matches `test/🎃`). SQLite compares text byte-exact, so this only reproduces on MariaDB.

- `utf8mb4` columns that collapse: `albums.album_title`, display/name text (`*_name`, `*_title`).
- `VARBINARY` columns that stay byte-exact: `albums.album_slug`, `albums.album_filter`, `albums.album_path`, `photos.photo_path`, and every `*_uid`. A `utf8mb4` column compared against a `VARBINARY` column is byte-exact (the binary operand wins).

The durable fix for an identity/path column is to make it `VARBINARY` — `album_path` is `VARBINARY(1024)` so it matches `photos.photo_path` and `album_path = ?` lookups are byte-exact at the database. Where a `utf8mb4` column must stay, keep the SQL but re-verify the match byte-exact in Go before accepting it (see `FindFolderAlbum` / `findFolderAlbumByPath`, whose Go re-check is retained as defense-in-depth even now that `album_path` is `VARBINARY`). For self-join SQL where a Go re-check is awkward, `HEX(col) = HEX(col)` compares byte-exact on both MariaDB and SQLite. Legacy folder slugs drop emoji entirely (`slug.Make("ins/🪞") == "ins"`) and long paths truncate to `ClipSlug` runes, so distinct folders can still collide on `album_slug`; folder albums are therefore deduplicated by `album_filter` (the byte-exact serialized path), not by slug (see `query.RemoveDuplicateMoments`).

### VARBINARY Index Prefix Limit

InnoDB caps an index key prefix at **767 bytes** on the `COMPACT`/`REDUNDANT` row formats, and only allows up to 3072 bytes on `DYNAMIC`/`COMPRESSED`. On a `VARBINARY` column the prefix is counted in **bytes** (on `utf8mb4` it is counted in characters, i.e. up to 4 bytes each), so converting a long text column to `VARBINARY` can push an existing prefix index over the limit on older or non-`DYNAMIC` installs. Keep prefix indexes on long `VARBINARY` path/filter columns at **≤ 767 bytes**; the project convention is **512** (`albums.album_filter(512)`, `albums.album_path(512)`). A prefix index only narrows candidate rows — the full-column comparison stays exact — so a shorter prefix costs nothing for correctness.
