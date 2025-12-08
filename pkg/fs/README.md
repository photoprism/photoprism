## PhotoPrism — pkg/fs

**Last Updated:** November 25, 2025

### Overview

`pkg/fs` provides safe, cross-platform filesystem helpers used across PhotoPrism. It supplies permission constants, copy/move utilities with force-aware semantics, safe path joins, archive extraction with size limits, MIME and extension lookups, hashing, canonical path casing, and fast directory walking with ignore lists.

#### Goals

- Offer reusable, side-effect-safe filesystem helpers that other packages can call without importing `internal/*`.
- Enforce shared permission defaults (`ModeDir`, `ModeFile`, `ModeConfigFile`, `ModeSecretFile`, `ModeBackupFile`).
- Protect against common filesystem attacks (path traversal, overwrite of non-empty files without `force`, unsafe zip extraction).
- Provide consistent file-type detection (extensions/MIME), hashing, and fast walkers with skip logic for caches and `.ppstorage` markers.

#### Non-Goals

- Database migrations or metadata parsing (handled elsewhere).
- Edition-specific behavior; all helpers are edition-agnostic.

### Package Layout (Code Map)

- Permissions & paths: `mode.go`, `filepath.go`, `canonical.go`, `case.go`.
- Copy/Move & write helpers: `copy_move.go`, `write.go`, `cache.go`, `purge.go`.
- Archive extraction: `zip.go` (size limits, safe join), tests in `zip_test.go`.
- File info & types: `file_type*.go`, `mime.go`, `file_ext*.go`, `name.go`.
- Hashing & IDs: `hash.go`, `id.go`.
- Walkers & ignore rules: `walk.go`, `ignore.go`, `done.go`.
- Utilities: `bytes.go`, `resolve.go`, `symlink.go`, `modtime.go`, `readlines.go`.

### Usage & Test Guidelines

- Overwrite semantics: pass `force=true` only when the caller explicitly confirmed replacement; empty files may be replaced without `force`.
- Permissions: use provided mode constants; do not mix with stdlib `io/fs` bits.
- Zip extraction: always set `fileSizeLimit` / `totalSizeLimit` in `Unzip` for untrusted inputs; ensure tests cover path traversal and size caps (see `zip_test.go`).
- Focused tests: `go test ./pkg/fs -run 'Copy|Move|Unzip|Write' -count=1` keeps feedback quick; full package: `go test ./pkg/fs -count=1`.

### Recent Changes & Improvements

- Hardened `safeJoin`: normalize `\\`/`/`, use `filepath.Rel` to reject paths escaping `baseDir`, and keep volume/absolute checks.
- Added optional max-entries guard in `Unzip` and treat `totalSizeLimit=0` as “no limit” while documenting `-1` as unlimited.
- Added pool copy buffers (128–256 KiB) that use `io.CopyBuffer` in `Copy`, `Hash`, `Checksum`, `WriteFileFromReader` to cut allocations/GC.

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
