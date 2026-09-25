## PhotoPrism — pkg/fs

**Last Updated:** September 25, 2026

### Overview

`pkg/fs` provides safe, cross-platform filesystem helpers used across PhotoPrism. It supplies permission constants, copy/move utilities with force-aware semantics, safe path joins, archive extraction with size limits, bounded image decode helpers, MIME and extension lookups, hashing, canonical path casing, and fast directory walking with ignore lists.

#### Goals

- Offer reusable, side-effect-safe filesystem helpers that other packages can call without importing `internal/*`.
- Enforce shared permission defaults (`ModeDir`, `ModeFile`, `ModeConfigFile`, `ModeSecretFile`, `ModeBackupFile`).
- Protect against common filesystem attacks (path traversal, overwrite of non-empty files without `force`, unsafe zip extraction).
- Provide bounded image decode helpers that do not route TIFF through generic `image.Decode()` dispatch.
- Provide consistent file-type detection (extensions/MIME), hashing, and fast walkers with skip logic for caches and `.ppstorage` markers.

#### Non-Goals

- Database migrations or metadata parsing (handled elsewhere).
- Edition-specific behavior; all helpers are edition-agnostic.

### Package Layout (Code Map)

- Standard file/directory names: `const.go`; transfer reservations: `reserved.go`. `ConfigOptionsName`, `ConfigDefaultsName`, `ConfigSettingsName`, and `ConfigHubName` are extension-free basenames for `ConfigFilePath`. Constants can be reused without implying a transfer restriction.
- Permissions & paths: `mode.go`, `filepath.go`, `canonical.go`, `case.go`.
- Copy/Move & write helpers: `copy_move.go`, `write.go`, `cache.go`, `purge.go`.
- Archive extraction: `zip.go` (size limits, safe join), tests in `zip_test.go`.
- Bounded image decode helpers: `image_decode.go` (direct JPEG/PNG/GIF/BMP/TIFF/WEBP dispatch with TIFF offset validation).
- File info & types: `file_type*.go`, `mime.go`, `file_ext*.go`, `name.go`.
- Stacking names: `stack.go` (`StackPrefix`, `Insta360VideoPattern`, `Insta360PhotoPattern`).
- Hashing & IDs: `hash.go`, `id.go`.
- Walkers & ignore rules: `walk.go`, `ignore.go`, `done.go`.
- Utilities: `bytes.go`, `resolve.go`, `symlink.go`, `modtime.go`, `readlines.go`.

### Usage & Test Guidelines

- Overwrite semantics: pass `force=true` only when the caller explicitly confirmed replacement; empty files may be replaced without `force`.
- Permissions: use provided mode constants; do not mix with stdlib `io/fs` bits. `ModeFile` (`0666`) and `ModeDir` (`0777`) are creation defaults filtered by the process umask, not final modes for `Chmod`. Explicit permission changes use the intended final mode.
- Staging: `OpenStageFile` returns an exclusively created temporary file beside its destination, with `ModeFile` filtered by the process umask. The caller closes the handle and removes or publishes its pathname. `CreateStageFile` closes the reservation and returns its path for subprocess writers. Both preserve the destination extension and require an existing parent directory; neither uses a global temporary directory or copies data across mounts.
- Reserved names: `ReservedPathNames` returns a sorted copy of the exact administrative-name set, including PhotoPrism storage markers/credential files and `.gitconfig`. `ReservedPathPatterns` returns the additional component patterns: `.env.*`, `.*ignore`, `.*_history`, `.bash_history-*.tmp`, and `.*.cnf`. `ReservedPathSuffixes` returns the reserved component suffixes, currently `.rclonelink`, a storage-driver link representation. The default transfer policy excludes every matching ignore name. Local DAV uses `ReservedPathPolicy{AllowIgnoreNames: true}` for visibility and separately requires managed-file write authority. Use `HasReservedComponent` for policy decisions, not the catalogs alone.
- Reserved path components: `HasReservedComponent` matches the reserved names, patterns, and suffixes case-insensitively, using slash or backslash separators. Callers supply paths relative to their own root; the component helper neither resolves links nor decodes URLs. `HasReservedTarget` is a standalone resolved-target inspection utility, not used by upload, archive, WebDAV, or sync name-policy enforcement. Operator-created filesystem links are trusted configuration.
- ZIP entry filters: optional `Unzip` name/directory predicates are combined as requirements and run before extracting entries. A rejected entry is reported in `skipped` without writing or replacing its destination. Reserved components and symbolic-link entries are always skipped. Direct `UnzipFile` refuses symbolic-link entries before any destination write. Predicates receive the directory flag even when its entry name has no trailing slash; nil or omitted filters add no constraints beyond the built-in extraction policy.
- Stacking vs. file names: `StackPrefix` returns the name under which a file is stacked with a photo. It maps the lens, proxy, and derived sidecar files of an Insta360 separate-lens capture (`VID|LRV_…insv`, and `IMG_…insp`, whose naming is assumed) to the name of the `_00` file, whether or not sequences are stripped; for any other name it equals `BasePrefix`. Use `BasePrefix`, `AbsPrefix`, and `RelPrefix` wherever a file's own name is needed, e.g. to find or rename its sidecars.
- Zip extraction: always set `fileSizeLimit` / `totalSizeLimit` in `Unzip` for untrusted inputs; ensure tests cover path traversal and size caps (see `zip_test.go`).
- Image decode helpers: use `DecodeImageFile`, `DecodeImageConfigFile`, `DecodeImageData`, or `DecodeImageConfigData` instead of generic `image.Decode()` / `image.DecodeConfig()` for user media. TIFF headers are validated against the bounded reader size before decode.
- Focused tests: `go test ./pkg/fs -run 'Copy|Move|Unzip|Write' -count=1` keeps feedback quick; full package: `go test ./pkg/fs -count=1`.

### Recent Changes & Improvements

- Shared `SafeJoin` (`join.go`): normalize `\\`/`/`, use `filepath.Rel` to reject paths escaping `baseDir`, and keep volume/absolute checks. Reused by `Unzip`, the WebDAV sync client, and the server upload handler.
- Added optional max-entries guard in `Unzip` and treat `totalSizeLimit=0` as “no limit” while documenting `-1` as unlimited.
- Added pool copy buffers (128–256 KiB) that use `io.CopyBuffer` in `Copy`, `Hash`, `Checksum`, `Sha256`, `WriteFileFromReader` to cut allocations/GC.

#### Pool Copy Buffers

- Read/write iterations per 4 GiB file:
  - Before: ~131,072 iterations (4 GiB / 32 KiB).
  - After: 16,384 iterations (4 GiB / 256 KiB).
  - ~8× fewer syscalls and loop bookkeeping.
- Latency saved (order-of-magnitude):
  - If each read+write pair costs ~2 µs of syscall/loop overhead, skipping ~115k iterations saves ≈0.23 s on a 4 GiB stream.
  - On SSD/NVMe where disk I/O dominates, expect ~5–10% throughput gain; on spinning disks or network mounts with higher syscall cost, closer to 10–20% is realistic.
  - CPU-bound hashing (SHA-1) sees mostly overhead reduction; the hash itself stays the dominant cost, but you still avoid ~8× buffer boundary checks and syscalls, so a few percent improvement is typical.
- Allocation/GC savings:
  - Before: each call allocated a fresh 32 KiB buffer; hashing and copy both did this per invocation.
  - After: pooled 256 KiB buffer reused; effectively zero steady-state allocations for these paths, which is most noticeable when hashing or copying many files in a batch (less GC pressure, fewer pauses).
- Net effect on large video files (several GB):
  - Wall-clock improvement: modest but measurable (sub‑second on SSDs; up to a couple of seconds on slower media per 4 GiB).
  - CPU usage: a few percentage points lower due to fewer syscalls and eliminated buffer allocations.
  - GC: reduced minor-GC churn during bulk imports/hashes.
