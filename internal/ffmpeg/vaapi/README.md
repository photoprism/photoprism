## PhotoPrism — VA-API Hardware Transcoding

**Last Updated:** May 30, 2026

### Overview

`internal/ffmpeg/vaapi` builds the FFmpeg command line that transcodes videos to MPEG-4 AVC (H.264) through the Video Acceleration API (VA-API). VA-API is the generic Linux hardware-acceleration interface and works with Intel (iHD/i965) and AMD (Mesa) GPUs, which makes it the most portable hardware path and the recommended choice for Intel GPUs that are too old for Quick Sync via oneVPL (see [Intel QSV](../intel/README.md)).

The single entry point is `TranscodeToAvcCmd(srcName, destName string, opt encode.Options) *exec.Cmd`.

### Command Line

The builder emits one command; the initialized device gets an explicit path only when `Options.Device` is set, otherwise FFmpeg auto-detects the default render node:

```
ffmpeg -hide_banner -y -strict -2 \
  -init_hw_device vaapi=va[:<device>] \
  -hwaccel vaapi -hwaccel_device va -filter_hw_device va \
  -i <src> \
  -c:a aac \
  -vf "scale='if(gte(iw,ih), min(<size>, iw), -2):if(gte(iw,ih), -2, min(<size>, ih))',format=nv12,hwupload" \
  -c:v h264_vaapi \
  -map 0:v:0 -map 0:a:0? -ignore_unknown \
  -qp 25 \
  -f mp4 -movflags use_metadata_tags+faststart -map_metadata 0 \
  <dest>
```

#### Pipeline

1. **Decode** — `-hwaccel vaapi` decodes in hardware when the driver supports the input codec and silently falls back to software decoding otherwise (it is best-effort). Because no `-hwaccel_output_format` is set, frames are transferred back to system memory after decoding.
2. **Filter** — `scale=…` downscales in software, `format=nv12` converts the pixel layout, and `hwupload` uploads the frames to VA-API surfaces on the device referenced by `-filter_hw_device`.
3. **Encode** — `h264_vaapi` encodes the uploaded surfaces.

#### FFmpeg 8 Requirement

FFmpeg 8 no longer derives a filter device from `-hwaccel vaapi` alone, so the `hwupload` filter aborts with `A hardware device reference is required to upload frames to.` unless a filter device is provided explicitly. The builder therefore creates a named device with `-init_hw_device vaapi=va[:<device>]` and points both the decoder (`-hwaccel_device va`) and the filter graph (`-filter_hw_device va`) at it. The legacy `-vaapi_device <path>` shorthand also works but is decoder-agnostic; the named-device form keeps hardware decode and filtering on the same device.

### Flags

| Flag                          | Value                             | Purpose                                                              |
|-------------------------------|-----------------------------------|----------------------------------------------------------------------|
| `-init_hw_device`             | `vaapi=va` or `vaapi=va:<device>` | Creates the named VA-API device `va`; auto-detects when no path set. |
| `-hwaccel`                    | `vaapi`                           | Best-effort hardware decode.                                         |
| `-hwaccel_device`             | `va`                              | Binds the decoder to the named device.                               |
| `-filter_hw_device`           | `va`                              | Supplies the device that `hwupload` uploads to.                      |
| `-vf … ,format=nv12,hwupload` | from `encode.FormatNV12`          | Software scale, NV12 conversion, then upload to a VA-API surface.    |
| `-c:v`                        | `h264_vaapi`                      | VA-API H.264 encoder.                                                |
| `-qp`                         | `25` (`DefaultQuality` 50)        | Constant-QP quality, via `Options.QpQuality()`.                      |

### Encoders & Decoders

- **Encoders** (driver-dependent): `h264_vaapi`, `hevc_vaapi`, `av1_vaapi`, `vp8_vaapi`, `vp9_vaapi`, `mpeg2_vaapi`, `mjpeg_vaapi`. PhotoPrism uses `h264_vaapi`.
- **Decoders**: the `vaapi` hwaccel decodes whatever the VA driver advertises (commonly MPEG-2, H.264, HEVC, VP8/VP9, AV1). Use `vainfo` to list the entry points for a given device.

### Device Paths

- DRM render node, conventionally `/dev/dri/renderD128` (a second GPU is `renderD129`, and so on).
- `Options.Device` accepts either a full path or a numeric index; an empty value lets FFmpeg pick the default render node.
- The process user needs access to the node (typically membership in the `render` and/or `video` groups).

### Supported Input & Output Formats

- **Input**: any container/codec FFmpeg can demux and decode; hardware decode is opportunistic with a software fallback, so unsupported codecs still transcode.
- **Output**: H.264 in an MP4 container with `use_metadata_tags+faststart` for streaming.

### Required System Packages & Libraries

- `libva` with the DRM backend (`libva.so.2`, `libva-drm.so.2`) and `libdrm.so.2`.
- A VA driver that matches the GPU, installed under `…/dri/`:
  - Intel Gen8+ → `intel-media-va-driver` (`iHD_drv_video.so`).
  - Older Intel (pre-Gen8) → `i965-va-driver` (`i965_drv_video.so`).
  - AMD → `mesa-va-drivers` (`radeonsi_drv_video.so`, `r600_drv_video.so`).
- FFmpeg built with VA-API support — confirm with `ffmpeg -hwaccels` (lists `vaapi`).
- Optional: `vainfo` (`libva-utils`) for diagnostics; `LIBVA_DRIVER_NAME` to force a driver.

### Verification

Confirmed on this environment with FFmpeg 8.0.1 (libavcodec 62), Intel iHD driver 26.1.2, libva 2.22 / VA-API 1.23, encoding `/dev/dri/renderD128`: the `25fps.vp9` fixture (VP9) transcodes to H.264 320×240. Run the real hardware path with:

```
PHOTOPRISM_FFMPEG_TEST_ENCODER=vaapi go test ./internal/ffmpeg -run 'TestTranscodeCmd/Vaapi' -count=1 -v
```

Without the opt-in variable the test only asserts the generated command string.
