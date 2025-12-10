## PhotoPrism — FFmpeg Integration

**Last Updated:** November 22, 2025

### Overview

`internal/ffmpeg` wraps the `ffmpeg` CLI to transcode videos to AVC/H.264, remux containers, and extract preview frames in a predictable, testable way. Command builders share option structs so CLI tools, workers, and tests can select software or hardware encoders without duplicating flag logic.

#### Context & Constraints

- Relies on the system `ffmpeg` binary; defaults to `FFmpegBin` but callers may override `Options.Bin`.
- Inputs are internal filenames and option structs (not user input); exec invocations are annotated with `#nosec G204`.
- Downstream jobs may run concurrently, so `TranscodeCmd` returns a `useMutex` hint to serialize expensive work.
- Remux and extract commands honor `Force` and reuse shared map flags; metadata copying is limited to safe defaults.

#### Goals

- Provide consistent command lines for software and hardware AVC encoders.
- Keep remuxing and preview extraction lightweight while preserving metadata where possible.
- Centralize quality and size clamping logic so UIs/CLI can pass user preferences safely.

#### Non-Goals

- Full coverage of every FFmpeg codec or container; the package focuses on MP4/H.264 paths required by PhotoPrism.
- Direct management of FFmpeg installation or GPU availability.

### Encoders, Containers, & Hardware

- **Software AVC:** `encode.TranscodeToAvcCmd` (x264 or default encoder).
- **Intel Quick Sync:** `internal/ffmpeg/intel` (`h264_qsv`) with optional `Options.Device`.
- **NVIDIA NVENC:** `internal/ffmpeg/nvidia` (`h264_nvenc`).
- **Apple VideoToolbox:** `internal/ffmpeg/apple` (`h264_videotoolbox`).
- **VA-API:** `internal/ffmpeg/vaapi` (`h264_vaapi`) supporting optional device paths.
- **V4L2 M2M:** `internal/ffmpeg/v4l` (`h264_v4l2m2m`) for ARM/embedded targets.
- **Containers:** MP4 is the primary target (`fs.VideoMp4`); `RemuxCmd` can handle other `fs.Type` values when provided.
- **Streaming flags:** `encode.MovFlags` defaults to `use_metadata_tags+faststart` to keep outputs stream-friendly.

### Package Layout (Code Map)

- `encode/` — shared option structs, quality helpers, default map/metadata flags, software AVC command builder.
- `apple/`, `intel/`, `nvidia/`, `vaapi/`, `v4l/` — hardware-specific AVC command builders.
- `remux.go` — container-only transfers with metadata copy and temp-file safety.
- `transcode_cmd.go` — selects encoder, handles animated image inputs, and signals mutex usage.
- `extract_image_cmd.go` — JPEG/PNG preview frame extraction with color-space presets.
- `test.go` & `*_test.go` — reusable command runner and smoke tests (use fixtures in `testdata/`).
- `ffmpeg.go` — package logger hook.

### Related Packages & Entry Points

- `internal/thumb` calls these builders for video previews and thumbnails.
- `internal/commands` and workers select encoders based on configuration options and reuse `encode.Options`.
- `pkg/fs` supplies path helpers, existence checks, and file-mode constants referenced by remux/extract logic.

### Configuration & Safety Notes

- Clamp size and quality via `NewVideoOptions` to `[1, 15360]` pixels and the defined quality bounds.
- Remuxing respects `Options.Force`; without it existing outputs are preserved.
- Metadata copying uses `-map_metadata` and `clean` sanitizers; only safe string fields (title, description, comment, author, creation_time) are added when set.
- Hardware helpers expect the matching FFmpeg build and devices; callers should gate selection via config or environment (see `PHOTOPRISM_FFMPEG_ENCODER` guidance in `AGENTS.md`).

### Testing

- Run unit tests: `go test ./internal/ffmpeg/...`
- Hardware-specific tests assume the encoder is available; keep runs gated via config when adding new cases.

### Operational Tips

- Prefer `TranscodeCmd` over manual `exec.Command` to keep logging, metadata, and mutex hints consistent.
- Use `RemuxFile` to convert containers without re-encoding; it creates a temp file and swaps atomically.
- For preview frames, pass `encode.Options` with `SeekOffset` and `TimeOffset` computed from video duration (see `NewPreviewImageOptions`).
