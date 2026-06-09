## PhotoPrism — Vulkan Video Transcoding

**Last Updated:** May 30, 2026

### Overview

`internal/ffmpeg/vulkan` builds the FFmpeg command line that transcodes videos to MPEG-4 AVC (H.264) through the Vulkan video extensions. The `h264_vulkan` encoder is vendor-neutral: it runs on any GPU whose Vulkan driver advertises the video-encode extensions, which currently includes AMD (Mesa RADV), NVIDIA (proprietary driver), and — for decode more than encode — Intel (Mesa ANV). It requires FFmpeg 8 or later.

The single entry point is `TranscodeToAvcCmd(srcName, destName string, opt encode.Options) *exec.Cmd`.

### Command Line

The initialized device gets an explicit physical-device selector only when `Options.Device` is set; otherwise FFmpeg picks the default device:

```
ffmpeg -hide_banner -y -strict -2 \
  -init_hw_device vulkan=vk[:<device>] \
  -filter_hw_device vk \
  -i <src> \
  -c:a aac \
  -vf "scale='if(gte(iw,ih), min(<size>, iw), -2):if(gte(iw,ih), -2, min(<size>, ih))',format=nv12,hwupload" \
  -c:v h264_vulkan \
  -map 0:v:0 -map 0:a:0? -ignore_unknown \
  -qp 25 \
  -f mp4 -movflags use_metadata_tags+faststart -map_metadata 0 \
  <dest>
```

#### Pipeline

1. **Decode** — performed in software (the command sets no `-hwaccel`), so frames start in system memory.
2. **Filter** — `scale=…` and `format=nv12` run in software, then a single `hwupload` moves the frames onto the Vulkan device referenced by `-filter_hw_device`. The `hwupload` step is part of `encode.FormatNV12`, so the builder must not append it a second time — `…,hwupload,hwupload` fails because the frames are already on the GPU.
3. **Encode** — `h264_vulkan` encodes the uploaded Vulkan frames.

### Flags

| Flag                          | Value                             | Purpose                                                                |
|-------------------------------|-----------------------------------|------------------------------------------------------------------------|
| `-init_hw_device`             | `vulkan=vk` or `vulkan=vk:<index>` | Creates the named Vulkan device `vk`; `<index>` is a physical-device index, not a DRM path. |
| `-filter_hw_device`           | `vk`                              | Supplies the device that `hwupload` and the encoder attach to.         |
| `-vf … ,format=nv12,hwupload` | from `encode.FormatNV12`          | Software scale, NV12 conversion, then a single upload to a Vulkan frame. |
| `-c:v`                        | `h264_vulkan`                     | Vulkan video H.264 encoder (FFmpeg 8+).                                |
| `-qp`                         | `25` (`DefaultQuality` 50)        | Constant-QP quality, via `Options.QpQuality()`.                       |

### Encoders & Decoders

- **Encoders**: `h264_vulkan`, `hevc_vulkan`, `av1_vulkan`, `ffv1_vulkan` (driver-dependent). PhotoPrism uses `h264_vulkan`.
- **Decoders**: decode is done in software here; the Vulkan device is used only for filtering and encoding.

### Device Paths

- The selector is a Vulkan physical-device index (e.g. `0`), not a DRM render-node path. On a multi-GPU host the first device may not be the encode-capable one, so set `Options.Device` (`PHOTOPRISM_FFMPEG_DEVICE`) to the right index.
- The Vulkan driver still reaches the GPU through the kernel: Mesa drivers use `/dev/dri/renderD128`; the NVIDIA driver uses `/dev/nvidia*`.

### Supported Input & Output Formats

- **Input**: any container/codec FFmpeg can demux and decode in software.
- **Output**: H.264 in an MP4 container with `use_metadata_tags+faststart`.

### Required System Packages & Libraries

- FFmpeg built with Vulkan support — confirm with `ffmpeg -hwaccels` (lists `vulkan`) and `ffmpeg -encoders | grep vulkan`.
- The Vulkan loader `libvulkan.so.1` (`libvulkan1`).
- A Vulkan driver (ICD) that advertises `VK_KHR_video_encode_queue` and `VK_KHR_video_encode_h264`:
  - AMD (RDNA 2+) → `mesa-vulkan-drivers` (RADV).
  - Intel → `mesa-vulkan-drivers` (ANV); note that many integrated GPUs expose only decode, not encode.
  - NVIDIA (Turing+) → proprietary driver via the NVIDIA Container Toolkit with the `graphics` capability (the driver libraries are mounted from the host, not installed via `apt`).
- Optional: `vulkan-tools` (`vulkaninfo`) for diagnostics.

### Verification

Check that a device advertises the encode extensions before selecting this encoder:

```
vulkaninfo | grep VK_KHR_video_encode
```

If `VK_KHR_video_encode_h264` is listed, run the real hardware path with:

```
PHOTOPRISM_FFMPEG_TEST_ENCODER=vulkan go test ./internal/ffmpeg -run 'TestTranscodeCmd/Vulkan' -count=1 -v
```

Without the opt-in variable the test only asserts the generated command string. Note that Vulkan video encode is the least widely available of the hardware encoders: at the time of writing, the Intel ANV driver does not advertise the encode extensions on integrated Raptor Lake graphics, and an encode-capable AMD, NVIDIA, or discrete Intel device is required to exercise this path end to end.
