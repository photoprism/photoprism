## PhotoPrism — Intel Quick Sync Transcoding

**Last Updated:** May 30, 2026

### Overview

`internal/ffmpeg/intel` builds the FFmpeg command line that transcodes videos to MPEG-4 AVC (H.264) through Intel Quick Sync Video (QSV). On Linux, QSV runs on top of VA-API but keeps the whole pipeline on the GPU, which avoids the system-memory round trip the [VA-API](../vaapi/README.md) path performs. Use QSV on Intel GPUs that the oneVPL runtime supports (Broadwell / Gen8 and newer); older Intel GPUs must use the VA-API encoder instead.

The single entry point is `TranscodeToAvcCmd(srcName, destName string, opt encode.Options) *exec.Cmd`.

### Command Line

`-hwaccel_device` is added only when `Options.Device` is set; otherwise FFmpeg selects the default device:

```
ffmpeg -hide_banner -y -strict -2 \
  -hwaccel qsv [-hwaccel_device <device>] -hwaccel_output_format qsv \
  -i <src> \
  -c:a aac \
  -vf "scale_qsv=w='if(gte(iw,ih), min(<size>, iw), -1)':h='if(gte(iw,ih), -1, min(<size>, ih))':format=nv12" \
  -c:v h264_qsv \
  -map 0:v:0 -map 0:a:0? -ignore_unknown \
  -preset fast -global_quality 25 \
  -f mp4 -movflags use_metadata_tags+faststart -map_metadata 0 \
  <dest>
```

#### Pipeline

1. **Decode** — `-hwaccel qsv -hwaccel_output_format qsv` decodes into QSV surfaces that stay in GPU memory. The input codec must have a QSV decoder; if it does not, FFmpeg fails and the caller falls back to the software encoder.
2. **Filter** — `scale_qsv=…:format=nv12` scales and converts on the GPU, so there is no `hwupload` step (and none of the filter-device requirement the VA-API path has under FFmpeg 8).
3. **Encode** — `h264_qsv` encodes the QSV surfaces directly.

### Flags

| Flag                     | Value                      | Purpose                                                              |
|--------------------------|----------------------------|----------------------------------------------------------------------|
| `-hwaccel`               | `qsv`                      | Quick Sync hardware decode.                                          |
| `-hwaccel_device`        | `<device>` (when set)      | Selects the render node for decode and encode.                       |
| `-hwaccel_output_format` | `qsv`                      | Keeps decoded frames as on-GPU QSV surfaces.                         |
| `-vf scale_qsv=…`        | from `encode.FormatQSV`    | On-GPU scale and NV12 conversion (computes the auto axis with `-1`). |
| `-c:v`                   | `h264_qsv`                 | Quick Sync H.264 encoder.                                            |
| `-preset`                | `fast`                     | Encoder speed/quality trade-off, via `Options.Preset`.               |
| `-global_quality`        | `25` (`DefaultQuality` 50) | Quality-based rate-control target, via `Options.GlobalQuality()`.    |

### Encoders & Decoders

- **Encoders**: `h264_qsv`, `hevc_qsv`, `av1_qsv`, `vp9_qsv`, `mpeg2_qsv`, `mjpeg_qsv`. PhotoPrism uses `h264_qsv`.
- **Decoders**: `h264_qsv`, `hevc_qsv`, `vp9_qsv`, and the other `*_qsv` decoders the runtime exposes.

### Device Paths

- DRM render node, conventionally `/dev/dri/renderD128`.
- `Options.Device` accepts a full path or a numeric index; empty lets FFmpeg pick the default device.
- The process user needs render-node access (`render`/`video` group membership).

### Supported Input & Output Formats

- **Input**: containers whose video codec has a matching `*_qsv` hardware decoder (e.g. H.264, HEVC, VP9). Unlike VA-API there is no software-decode fallback inside this command because the pipeline requires QSV surfaces.
- **Output**: H.264 in an MP4 container with `use_metadata_tags+faststart`.

### Required System Packages & Libraries

- The oneVPL dispatcher `libvpl.so.2` (`libvpl2`); FFmpeg here is built with `--enable-libvpl --disable-libmfx`, so QSV is served by oneVPL, not the legacy Media SDK.
- The oneVPL GPU runtime for the device: `libmfx-gen` / `onevpl-intel-gpu` (`libmfx-gen.so`), which supports Broadwell (Gen8) and newer.
- The Intel media VA driver (`intel-media-va-driver`, `iHD_drv_video.so`) plus `libva` and `libdrm` — QSV sits on top of VA-API on Linux.
- Confirm support with `ffmpeg -hwaccels` (lists `qsv`) and `ffmpeg -encoders | grep qsv`.

> **Hardware caveat:** the oneVPL GPU runtime does not support pre-Broadwell GPUs (e.g. Haswell). On those systems `h264_qsv` will not initialize and the VA-API encoder (`h264_vaapi`) should be used instead.

### Verification

Confirmed on this environment with FFmpeg 8.0.1 (libavcodec 62), Intel iHD driver 26.1.2, oneVPL runtime `libmfx-gen` 1.2.16, encoding `/dev/dri/renderD128`: the `30fps.mov` fixture (HEVC) transcodes to H.264 1500×844 and `25fps.vp9` (VP9) to H.264 320×240. Run the real hardware path with:

```
PHOTOPRISM_FFMPEG_TEST_ENCODER=intel go test ./internal/ffmpeg -run 'TestTranscodeCmd/(IntelHvc|IntelVp9)' -count=1 -v
```

Without the opt-in variable the test only asserts the generated command string.
