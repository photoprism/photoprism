## PhotoPrism — NVIDIA NVENC Transcoding

**Last Updated:** October 3, 2026

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
  -rc:v vbr -cq 31 -b:v 0 [-maxrate <n>M] \
  -tune hq -profile:v high -level:v auto -coder:v 1 \
  -f mp4 -movflags use_metadata_tags+faststart -map_metadata 0 \
  <dest>
```

#### Pipeline

1. **Decode** — `-hwaccel auto` lets FFmpeg pick a hardware decoder (it selects `cuda` when an NVIDIA device is present) and falls back to software decoding otherwise.
2. **Filter** — `scale=…,format=yuv420p` scales and converts in software.
3. **Encode** — `h264_nvenc` uploads the YUV420P frames and encodes them on the GPU.

#### Rate Control

The encoder runs in VBR mode with a constant quality target (`-cq`) and no average bitrate (`-b:v 0`), so the size follows the content. `encode.CqQuality()` maps `PHOTOPRISM_FFMPEG_QUALITY` linearly to `-cq` as `1 + (100 - quality) * 3 / 5`: 100 gives 1, 80 gives 13, the default 50 gives 31, and 30 gives 43. The scale is steeper than the CRF scale of the software encoder because NVENC needs a higher `-cq` for the same visual quality: at 31 it comes close to `libx264 -preset fast -crf 25` in SSIM and XPSNR, and at 43 it matches `-crf 35`. At a `-cq` below about 15, i.e. a quality above about 78, the driver's own VBR ceiling is reached, so higher settings no longer differ.

`Options.MaxBitrate` limits the peak bitrate with `-maxrate`. `Convert.AvcBitrate()` sets it from the output resolution, i.e. the source size scaled down to `PHOTOPRISM_FFMPEG_SIZE`, at 12 bits per second for each pixel of the frame (25 Mbit/s for 1080p), limited by `PHOTOPRISM_FFMPEG_BITRATE` (60 Mbit/s by default), which also applies when the size is unknown. With `PHOTOPRISM_FFMPEG_BITRATE=-1`, no `-maxrate` is emitted. The limit is a rate control target rather than a hard cap: at common frame rates, noisy content can exceed it by about a third. NVENC applies it per frame at the nominal frame rate, so it is stricter for videos whose actual frame rate is lower, such as phone videos recorded with a variable frame rate. A `-bufsize` is not set, since it has no effect in constant-quality mode.

### Flags

| Flag                         | Value                                    | Purpose                                                     |
|------------------------------|------------------------------------------|-------------------------------------------------------------|
| `-hwaccel`                   | `auto`                                   | Best-effort hardware decode (resolves to `cuda` on NVIDIA). |
| `-c:v`                       | `h264_nvenc`                             | NVENC H.264 encoder.                                        |
| `-pix_fmt` / `-pixel_format` | `yuv420p`                                | Forces 8-bit 4:2:0 output for broad playback compatibility. |
| `-gpu`                       | `any`                                    | Lets the driver choose an NVENC-capable GPU.                |
| `-rc:v` / `-cq` / `-b:v`     | `vbr` / `31` / `0` (`DefaultQuality` 50) | VBR with a quality target, via `Options.CqQuality()`.       |
| `-maxrate`                   | `Options.MaxRate()`                      | Peak bitrate target, omitted without a limit.               |
| `-preset`                    | `fast`                                   | Encoder speed/quality trade-off, via `Options.Preset`.      |
| `-tune`                      | `hq`                                     | NVENC tuning info — high quality.                           |
| `-profile:v`                 | `high`                                   | H.264 High profile, as written by the software encoder.     |
| `-level:v`                   | `auto`                                   | Lets the encoder derive the H.264 level.                    |
| `-coder:v`                   | `1`                                      | CABAC entropy coding.                                       |

### Encoders & Decoders

- **Encoders**: `h264_nvenc`, `hevc_nvenc`, `av1_nvenc`. PhotoPrism uses `h264_nvenc`.
- **Decoders**: NVDEC/CUVID decoders such as `h264_cuvid`, `hevc_cuvid`, `vp9_cuvid`; reached here indirectly via `-hwaccel auto` rather than by naming a `*_cuvid` decoder.

### Device Paths

- NVENC uses the NVIDIA driver, not DRM render nodes. The kernel devices are `/dev/nvidia0` (plus `/dev/nvidiactl`, `/dev/nvidia-uvm`).
- `encode.DefaultAvcEncoder()` selects NVENC automatically when `/dev/nvidia0` exists, `NVIDIA_DRIVER_CAPABILITIES` is `video` or `all`, `NVIDIA_VISIBLE_DEVICES` is a number or `all`, and `PHOTOPRISM_INIT` does not contain `ffmpeg`.

### Supported Input & Output Formats

- **Input**: any container/codec FFmpeg can demux and decode; decode is opportunistic with a software fallback, so unsupported codecs still transcode (including 10-bit HEVC, which is converted to 8-bit on output).
- **Output**: H.264 (High profile, 4:2:0 8-bit) in an MP4 container with `use_metadata_tags+faststart`.

### Required System Packages & Libraries

- The NVIDIA proprietary driver, which provides the runtime libraries FFmpeg loads on demand: `libnvidia-encode.so` (NVENC) and `libnvcuvid.so` (NVDEC/CUVID). These are dlopened at run time, so they do not appear in `ldd ffmpeg`.
- An NVENC-capable GPU and a driver that supports the NVENC API version of the FFmpeg build; FFmpeg logs the minimum driver version if it is older. The command targets Turing and newer GPUs; it is verified on Ada (see below).
- FFmpeg built with NVENC/NVDEC support (ffnvcodec headers) — confirm with `ffmpeg -hwaccels` (lists `cuda`) and `ffmpeg -encoders | grep nvenc`.
- In Docker, expose the GPU with the NVIDIA Container Toolkit and set `NVIDIA_DRIVER_CAPABILITIES=video` (or `all`) plus `NVIDIA_VISIBLE_DEVICES`.

### Verification

Confirmed on this environment with FFmpeg 8.0.1 (libavcodec 62) on an NVIDIA GeForce RTX 4060, driver 595.91.07, CUDA 13.2: the `30fps.mov` fixture (10-bit HEVC) transcodes to H.264 High 1500×844 and `25fps.vp9` (VP9) to H.264 High 320×240, with `-hwaccel auto` resolving to `cuda`. Run the real hardware path with:

```
PHOTOPRISM_FFMPEG_TEST_ENCODER=nvidia go test ./internal/ffmpeg -run 'TestTranscodeCmd/Nvidia' -count=1 -v
```

Without the opt-in variable the test only asserts the generated command string.
