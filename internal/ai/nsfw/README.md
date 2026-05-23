## PhotoPrism — NSFW Package

**Last Updated:** May 21, 2026

### Overview

`internal/ai/nsfw` runs the built-in TensorFlow NSFW classifier to score images for drawing, hentai, neutral, porn, and sexy content. It is the default backend that powers the `Type: nsfw` model entry in [`internal/ai/vision`](../vision/README.md) and is the only NSFW engine that ships with PhotoPrism out of the box; operators can override it through `vision.yml` with an Ollama or OpenAI endpoint when they prefer to run NSFW detection on a remote LLM.

### Where It Gets Called

The package itself only exposes the model loader and a thin scoring API (`Result.IsSafe`, `Result.IsNsfw(threshold)`). Two upstream callers wire it into the runtime:

1. **Upload handler — [`internal/api/users_upload.go`](../../api/users_upload.go).** When `PHOTOPRISM_UPLOAD_NSFW=false` (default `true`), every accepted upload is screened by `vision.DetectNSFW` before indexing. Files that score above the threshold are deleted on the spot — they never reach `originals/`. When `UPLOAD_NSFW=true`, the upload path skips the check entirely.

2. **Index + vision-worker pipelines — [`internal/photoprism/index_mediafile.go`](../../photoprism/index_mediafile.go), [`internal/workers/vision.go`](../../workers/vision.go), [`internal/workers/meta.go`](../../workers/meta.go).** When `PHOTOPRISM_DETECT_NSFW=true` (default `false`), the indexer marks new photos as `PhotoPrivate = true` if the NSFW model flags them. Both code paths short-circuit when `DetectNSFW()` is false — the model is then neither loaded nor invoked.

Both flags are independent: you can reject uploads without flagging existing imports, flag existing imports without policing uploads, or both. The user-facing matrix lives at [docs.photoprism.app/user-guide/ai/nsfw/](https://docs.photoprism.app/user-guide/ai/nsfw/).

### Detection Through the Labels Model

When `Type: labels` is served by an Ollama or OpenAI engine and **both** `PHOTOPRISM_DETECT_NSFW=true` and `PHOTOPRISM_EXPERIMENTAL=true` are set, [`internal/config/config.go`](../../config/config.go) flips the package-level global `vision.DetectNSFWLabels` to `true`. The Ollama and OpenAI engine builders then swap their default label prompts for `LabelPromptNSFW` and the JSON schema generators add `nsfw` + `nsfw_confidence` fields, so NSFW classification piggybacks on the label-generation call instead of running as a separate inference pass.

When the shortcut is active, the labels-path check in `index_mediafile.go` (`labels.IsNSFW(threshold)`) can promote a photo to private without this package being touched. The dedicated TensorFlow model in `internal/ai/nsfw` is still used as a fallback whenever the labels path either does not run or does not return NSFW signals, and whenever `vision run --models nsfw` is invoked directly.

### How It Works

- **Model Loading** — Loads the NSFW SavedModel from `assets/models/` and resolves input/output ops (inferred if missing).
- **Input Preparation** — JPEG thumbnails (default size `Fit720`, see `MediaFile.DetectNSFW`) are decoded and transformed to the configured input resolution.
- **Inference & Output** — Produces five class probabilities mapped into a `Result` struct for downstream thresholds and UI badges.

### Threshold

`vision.yml` carries a `Thresholds.NSFW` value (default `75`, range `0-100`) that controls how confident the model must be before a picture is flagged. Lower values are more aggressive; higher values more permissive. The threshold applies to both the dedicated NSFW model and the NSFW fields returned via the label-generation shortcut.

```yaml
Thresholds:
  NSFW: 75
```

### Memory & Performance

TensorFlow tensors allocate C memory and are freed by Go GC finalizers. To keep RSS bounded during long runs, PhotoPrism periodically triggers garbage collection to return freed tensor memory to the OS. Tune with:

- `PHOTOPRISM_TF_GC_EVERY` (default **200**, `0` disables).  
  Lower values reduce peak RSS but increase GC overhead and can slow indexing.

### Troubleshooting Tips

- **Model fails to load:** Verify `saved_model.pb` and `variables/` exist under the model path.
- **Unexpected scores:** Confirm the input resolution matches the model and that logits are handled correctly.
- **High memory usage:** Adjust `PHOTOPRISM_TF_GC_EVERY` or reduce concurrent indexing load.
- **NSFW detection appears to stop working after switching labels to an LLM:** Confirm both `PHOTOPRISM_DETECT_NSFW=true` and `PHOTOPRISM_EXPERIMENTAL=true` are set. Without both, the labels-path shortcut is disabled and only an explicit `vision run --models nsfw` (or another caller that goes through this package directly) will produce NSFW flags.

### Related Docs

- [`internal/ai/vision/README.md`](../vision/README.md) — model registry, run scheduling, and the `DetectNSFWLabels` global
- [`internal/ai/vision/ollama/README.md`](../vision/ollama/README.md) — Ollama engine: `LabelPromptNSFW` swap-in
- [`internal/ai/vision/openai/README.md`](../vision/openai/README.md) — OpenAI engine: NSFW-aware prompt and schema
- [`internal/ai/vision/schema/README.md`](../vision/schema/README.md) — JSON schema variants used when NSFW is enabled
- [`internal/ai/tensorflow/README.md`](../tensorflow/README.md) — TensorFlow helpers, GC behavior, and model loading
- [docs.photoprism.app/user-guide/ai/nsfw/](https://docs.photoprism.app/user-guide/ai/nsfw/) — user-facing reference + flag matrix
