## PhotoPrism — Video Package

**Last Updated:** June 1, 2026

### Codecs & Containers

For maximum browser compatibility, PhotoPrism can transcode video codecs and containers [supported by FFmpeg](https://www.ffmpeg.org/documentation.html) to [MPEG-4 AVC](https://en.wikipedia.org/wiki/MPEG-4).

Running the following command in a terminal displays a list of supported codecs:

```
ffmpeg -decoders
```

See our advanced setup guide to learn how to [configure hardware video transcoding](https://docs.photoprism.app/getting-started/advanced/transcoding/).

Please Note:

1. Not all [video and audio formats](https://caniuse.com/?search=video%20format) can be [played with every browser](https://docs.photoprism.app/getting-started/troubleshooting/browsers/). For example, [AAC](https://caniuse.com/aac "Advanced Audio Coding") - the default audio codec for [MPEG-4 AVC / H.264](https://caniuse.com/avc "Advanced Video Coding") - is supported natively in Chrome, Safari, and Edge, while it is only optionally supported by the OS in Firefox and Opera.
2. HEVC/H.265 video files can have a `.mp4` file extension too, which is often associated with AVC only. This is because MP4 is a *container* format, meaning that the actual video content may be compressed with H.264, H.265, or something else. The file extension doesn't really tell you anything other than that it's probably a video file.
3. In case [FFmpeg is disabled](https://docs.photoprism.app/user-guide/settings/advanced/#disable-ffmpeg) or not installed, videos cannot be indexed because still images cannot be created. You should also have [ExifTool enabled](https://docs.photoprism.app/getting-started/config-options/#feature-flags) to extract metadata such as duration, resolution, and codec.

### Hybrid Photo/Video Formats

For more information on hybrid photo/video file formats, e.g. Apple Live Photos and Samsung/Google Motion Photos, see [github.com/photoprism/photoprism/tree/develop/pkg/media](https://github.com/photoprism/photoprism/tree/develop/pkg/media) and [docs.photoprism.app/developer-guide/media/live](https://docs.photoprism.app/developer-guide/media/live/).

### Codec Detection

PhotoPrism probes MP4 and QuickTime (ISO base media) files with [`go-mp4`](https://github.com/abema/go-mp4) to read the container brand, track layout, and codec. A chunk scan first locates the `ftyp` box and a compatible brand (`Chunks.FileTypeOffset` in `chunks.go`); in hybrid photo/video files this also yields the embedded-video offset, so the leading still image is skipped before parsing. MPEG-4 AVC / H.264 is then reported directly by the parser and needs no further work.

HEVC / H.265 and [MagicYUV](https://en.wikipedia.org/wiki/MagicYUV) sample entries are not yet surfaced by `go-mp4`, so PhotoPrism falls back to a chunk scan over the file head when the parser does not return a codec:

- The scan looks for the four-byte ISO BMFF sample-entry codes that identify these codecs — `hvc1`, `hev1`, `dvh1`, … for HEVC and `M8RG`, `M8Y2`, … for MagicYUV — in a single buffered pass (see `Chunks.SampleEntryOffset` in `chunks.go`).
- Each candidate is validated as a genuine visual sample entry rather than a raw byte match: the four bytes that precede it must be a plausible big-endian box size, and the bytes that follow must be the six reserved zero bytes and the nonzero `data_reference_index` mandated by ISO/IEC 14496-12 (`isVisualSampleEntry`). This rejects four-byte codes that merely collide with payload bytes, which is common in raw elementary streams such as the DV data inside some QuickTime files.
- The scan reads at most `HeadScanLimit` (16 MiB) and stops as soon as a valid entry is found. Buffered reads carry a small overlap so a code straddling a block boundary, and its validation window, stay visible.

The detected codec, together with the codec and container reported by ExifTool, drives the FFmpeg exclude list (default `magy`, `vfw`) that keeps known-problematic formats out of transcoding and thumbnail extraction.

#### Possible Improvements

- **Read the codec from the parsed `stsd` box.** Once `go-mp4` reports HEVC (a pending upstream contribution), the chunk scan and `HeadScanLimit` could be retired for parseable MP4/QuickTime files. Reading the sample entry directly would also detect codecs in large non-faststart files, where `moov` sits past the 16 MiB cap and the head scan cannot reach it today.
- **Anchor the scan at the video offset.** Hybrid photo/video files embed the video after the still image, so the scan currently reads through several megabytes of image data before reaching the video's sample entry. Starting at the already-known video offset would skip the image prefix without affecting plain video files.

> `HeadScanLimit` cannot simply be lowered for speed: motion photos place the video's sample entry several megabytes into the file, and non-faststart MP4/QuickTime files keep `moov` near the end, so a smaller cap would silently miss valid HEVC and MagicYUV streams.

### Standard Resolutions

The [`PHOTOPRISM_FFMPEG_SIZE`](https://docs.photoprism.app/getting-started/config-options/#file-converters) config option allows to limit the resolution of [transcoded videos](https://docs.photoprism.app/getting-started/advanced/transcoding/). It accepts the following standard sizes, while other values are automatically adjusted to the next supported size:

| Size | Usage              |
|:-----|:-------------------|
| 720  | SD TV, Mobile      |
| 1280 | HD TV, SXGA        |
| 1920 | Full HD            |
| 2048 | DCI 2K, Tablets    |
| 2560 | Quad HD, Notebooks |
| 3840 | 4K Ultra HD        |
| 4096 | DCI 4K, Retina 4K  |
| 7680 | 8K Ultra HD 2      |

### Technical References & Tutorials

| Title                               | URL                                                                                |
|:------------------------------------|:-----------------------------------------------------------------------------------|
| Web Video Codec Guide               | https://developer.mozilla.org/en-US/docs/Web/Media/Guides/Formats/Video_codecs     |
| Web Video Content-Type Headers      | https://developer.mozilla.org/en-US/docs/Web/Media/Guides/Formats/codecs_parameter |
| Media Container Formats             | https://developer.mozilla.org/en-US/docs/Web/Media/Guides/Formats/Containers       |
| MP4 Signature Format                | https://www.file-recovery.com/mp4-signature-format.htm                             |
| List of file signatures (Wikipedia) | https://en.wikipedia.org/wiki/List_of_file_signatures                              |
| How to use the io.Reader interface  | https://yourbasic.org/golang/io-reader-interface-explained/                        |
| AV1 Codec ISO Media File Format     | https://aomediacodec.github.io/av1-isobmff                                         |

----

*PhotoPrism® is a [registered trademark](https://www.photoprism.app/trademark/). By using the software and services we provide, you agree to our [Terms of Service](https://www.photoprism.app/terms/), [Privacy Policy](https://www.photoprism.app/privacy/), and [Code of Conduct](https://www.photoprism.app/code-of-conduct/). Docs are [available](https://link.photoprism.app/github-docs) under the [CC BY-NC-SA 4.0 License](https://creativecommons.org/licenses/by-nc-sa/4.0/); [additional terms](https://github.com/photoprism/photoprism/blob/develop/assets/README.md) may apply.*
