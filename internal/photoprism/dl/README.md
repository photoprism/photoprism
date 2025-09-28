# PhotoPrism Download Helpers

This package provides thin wrappers around `yt-dlp`, which the `photoprism dl` command uses for metadata discovery and downloading.

It currently supports two invocation methods:
  - Pipe: stream to stdout, PhotoPrism writes a file and remuxes with ffmpeg to ensure MP4 + embedded metadata.
  - File: `yt-dlp` writes files to disk using `--output`; PhotoPrism captures final paths via `--print after_move:filepath` and may remux when needed.

## Auth & Headers

- Supports `--cookies`, `--cookies-from-browser BROWSER[:PROFILE]`, and repeatable `--add-header` for both metadata and download flows.  
  - Container note: The `photoprism dl` CLI runs in a container by default and therefore does not expose a `--cookies-from-browser` flag (no access to local browser profiles). Use `--cookies <path>` with a Netscape cookies.txt file.
- Secrets are never logged; header values are redacted in trace logs.

## Key APIs

- `NewMetadata(ctx, url, Options)` → discovers formats and info (via `--dump-single-json`).
- `Metadata.DownloadWithOptions(ctx, DownloadOptions)` → pipe method (`stdout`).
- `Metadata.DownloadToFileWithOptions(ctx, DownloadOptions)` → file method (`--output` + `--print`).
- `RemuxOptionsFromInfo(ffmpegBin, fs.VideoMp4, Info, sourceURL)` → builds ffmpeg options to embed title/description/author/comment/created.

## yt-dlp CLI

- **Format selection**
  - Default behavior is `bestvideo*+bestaudio/best`, which already prefers the highest quality muxable streams.
  - Force a specific selection with `--format <expr>`; filter examples:
    - `bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]` – prefer MP4-compatible DASH.
    - `bv*[height<=1080]+ba/best[height<=1080]` – cap resolution.
- **Sorting**
  - `--format-sort <keys>` reorders candidates before the selection step. Keys can include `res`, `fps`, `size`, `br`, `vcodec`, etc.
  - Our CLI exposes this as `photoprism dl --format-sort` (alias `-s`). When omitted, pipe mode uses `lang,quality,res,fps,codec:avc:m4a,channels,size,br,asr,proto,ext,hasaud,source,id`; file mode leaves sorting to yt-dlp defaults.
- **Metadata & Post-processing**
  - `--embed-metadata` writes tags (title, artist, comment) via ffmpeg/mutagen.
  - `--postprocessor-args "ffmpeg:-metadata creation_time=<RFC3339>"` lets us inject timestamps when skipping local remux; PhotoPrism sets this when `CreatedFromInfo` yields a value.
  - `--merge-output-format mp4` and `--remux-video mp4` keep downloads MP4-friendly when yt-dlp has to join streams.
- **Other frequently used knobs**
  - `--download-sections`, `--add-header`, `--cookies`, `--proxy`, `--impersonate` – passed through via `dl.Options` when callers need them.

## Testing

- Tests stub `yt-dlp` with a tiny shell script that echoes JSON or creates a dummy file and prints its path. This avoids external network calls and brittle extractor behavior.
- Logging redaction is covered; argument construction is verified for cookies/headers.

## Notes

- Prefer the file method for sources with separate audio/video streams; the pipe method cannot always merge in that case.
- When the CLI’s `--file-remux=auto` is used, the final ffmpeg remux is skipped for MP4 outputs that already include metadata.
- Keep `yt-dlp` updated. Releases older than `2025.09.23` are known to miss YouTube video formats (SABR gating); the CLI now logs a warning when it detects an outdated build.
- Users who favor one approach can set `PHOTOPRISM_DL_METHOD=file` (or `pipe`) in the environment to change the default without touching CLI flags.
