## PhotoPrism — NVIDIA NVENC Transcoding

**Last Updated:** May 30, 2026

### Overview

`internal/ffmpeg/nvidia` builds the FFmpeg command line that transcodes videos to MPEG-4 AVC (H.264) through NVIDIA NVENC. The encoder accepts software frames and uploads them to the GPU internally, so the command uses a software filter chain and has no `hwupload` step — it is unaffected by the FFmpeg 8 filter-device requirement that the [VA-API](../vaapi/README.md) path must satisfy.

The single entry point is `TranscodeToAvcCmd(srcName, destName string, opt encode.Options) *exec.Cmd`. It emits one command regardless of `Options.Device` (NVENC selects the GPU via `-gpu any`, not a DRM path).

### Command Line

```
ffmpeg -hide_banner -y -strict -2 \
  -hwaccel auto \
  -i <src> \
  -pix_fmt yuv420p \
  -c:v h264_nvenc \
  -map 0:v:0 -map 0:a:0? -ignore_unknown \
  -c:a aac \
  -preset fast -pixel_format yuv420p -gpu any \
  -vf "scale='if(gte(iw,ih), min(<size>, iw), -2):if(gte(iw,ih), -2, min(<size>, ih))',format=yuv420p" \
  -rc:v constqp -cq 25 -tune 2 -profile:v 1 -level:v auto -coder:v 1 \
  -f mp4 -movflags use_metadata_tags+faststart -map_metadata 0 \
  <dest>
```

#### Pipeline

1. **Decode** — `-hwaccel auto` lets FFmpeg pick a hardware decoder (it selects `cuda` when an NVIDIA device is present) and falls back to software decoding otherwise.
2. **Filter** — `scale=…,format=yuv420p` scales and converts in software.
3. **Encode** — `h264_nvenc` uploads the YUV420P frames and encodes them on the GPU.

### Flags

| Flag                         | Value                                  | Purpose                                                     |
|------------------------------|----------------------------------------|-------------------------------------------------------------|
| `-hwaccel`                   | `auto`                                 | Best-effort hardware decode (resolves to `cuda` on NVIDIA). |
| `-c:v`                       | `h264_nvenc`                           | NVENC H.264 encoder.                                        |
| `-pix_fmt` / `-pixel_format` | `yuv420p`                              | Forces 8-bit 4:2:0 output for broad playback compatibility. |
| `-gpu`                       | `any`                                  | Lets the driver choose an NVENC-capable GPU.                |
| `-rc:v` / `-cq`              | `constqp` / `25` (`DefaultQuality` 50) | Constant-QP rate control, via `Options.CqQuality()`.        |
| `-preset`                    | `fast`                                 | Encoder speed/quality trade-off, via `Options.Preset`.      |
| `-tune`                      | `2`                                    | NVENC tuning info — low latency (`ll`).                     |
| `-profile:v`                 | `1`                                    | H.264 profile selector — main.                              |
| `-level:v`                   | `auto`                                 | Lets the encoder derive the H.264 level.                    |
| `-coder:v`                   | `1`                                    | CABAC entropy coding.                                       |

### Encoders & Decoders

- **Encoders**: `h264_nvenc`, `hevc_nvenc`, `av1_nvenc`. PhotoPrism uses `h264_nvenc`.
- **Decoders**: NVDEC/CUVID decoders such as `h264_cuvid`, `hevc_cuvid`, `vp9_cuvid`; reached here indirectly via `-hwaccel auto` rather than by naming a `*_cuvid` decoder.

### Device Paths

- NVENC uses the NVIDIA driver, not DRM render nodes. The kernel devices are `/dev/nvidia0` (plus `/dev/nvidiactl`, `/dev/nvidia-uvm`).
- `encode.DefaultAvcEncoder()` selects NVENC automatically when `/dev/nvidia0` exists, `NVIDIA_DRIVER_CAPABILITIES` is `video` or `all`, `NVIDIA_VISIBLE_DEVICES` is a number or `all`, and `PHOTOPRISM_INIT` does not contain `ffmpeg`.

### Supported Input & Output Formats

- **Input**: any container/codec FFmpeg can demux and decode; decode is opportunistic with a software fallback, so unsupported codecs still transcode (including 10-bit HEVC, which is converted to 8-bit on output).
- **Output**: H.264 (Main profile, 4:2:0 8-bit) in an MP4 container with `use_metadata_tags+faststart`.

### Required System Packages & Libraries

- The NVIDIA proprietary driver, which provides the runtime libraries FFmpeg loads on demand: `libnvidia-encode.so` (NVENC) and `libnvcuvid.so` (NVDEC/CUVID). These are dlopened at run time, so they do not appear in `ldd ffmpeg`.
- An NVENC-capable GPU (Kepler or newer).
- FFmpeg built with NVENC/NVDEC support (ffnvcodec headers) — confirm with `ffmpeg -hwaccels` (lists `cuda`) and `ffmpeg -encoders | grep nvenc`.
- In Docker, expose the GPU with the NVIDIA Container Toolkit and set `NVIDIA_DRIVER_CAPABILITIES=video` (or `all`) plus `NVIDIA_VISIBLE_DEVICES`.

### Verification

Confirmed on this environment with FFmpeg 8.0.1 (libavcodec 62) on an NVIDIA GeForce RTX 4060, driver 595.71.05, CUDA 13.2: the `30fps.mov` fixture (10-bit HEVC) transcodes to H.264 Main 1500×844 and `25fps.vp9` (VP9) to H.264 Main 320×240, with `-hwaccel auto` resolving to `cuda` and no deprecation warnings. Run the real hardware path with:

```
PHOTOPRISM_FFMPEG_TEST_ENCODER=nvidia go test ./internal/ffmpeg -run 'TestTranscodeCmd/Nvidia' -count=1 -v
```

Without the opt-in variable the test only asserts the generated command string.
