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

## Testing

- Tests stub `yt-dlp` with a tiny shell script that echoes JSON or creates a dummy file and prints its path. This avoids external network calls and brittle extractor behavior.
- Logging redaction is covered; argument construction is verified for cookies/headers.

## Notes

- Prefer the file method for sources with separate audio/video streams; the pipe method cannot always merge in that case.
- When the CLI’s `--file-remux=auto` is used, the final ffmpeg remux is skipped for MP4 outputs that already include metadata.
