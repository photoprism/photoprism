## PhotoPrism — NSFW Package

**Last Updated:** September 2, 2026

### Overview

`internal/ai/nsfw` runs the built-in TensorFlow NSFW classifier to score images for drawing, hentai, neutral, porn, and sexy content. It is the default backend that powers the `Type: nsfw` model entry in [`internal/ai/vision`](../vision/README.md) and is the only NSFW engine that ships with PhotoPrism out of the box; operators can override it through `vision.yml` with an Ollama or OpenAI endpoint when they prefer to run NSFW detection on a remote LLM.

### The Result Contract

`Result` carries a three-valued `Status` — `safe`, `unsafe`, or `unavailable` — alongside the `Score` it was decided from and the `Threshold` it was compared against. **The zero value is `unavailable`**, so a result that no detector filled in reads as "nothing was decided" rather than as a clearance.

That inversion is the point of the type. A safety signal whose zero value means safe cannot be told apart from a detector that never ran, and every path that could fail — a missing model, an unreadable thumbnail, a decode error, a disabled detector — produced exactly that value.

- `IsSafe()` is true **only** for an explicit clearance, never for a missing or failed check.
- `IsUnsafe()` is true only for an explicit detection.
- `IsUnavailable()` is true for everything else, including the zero value.
- `Decide(threshold)` reduces class scores to a decision. A result with no scores stays unavailable.
- `Unavailable(reason)` records why nothing was decided, which is what the callers log.

`ErrNotConfigured` and `ErrDetectorUnavailable` distinguish a detector the operator turned off from one that broke. The upload path treats those differently, so they must not be collapsed.

### Where It Gets Called

Two upstream callers wire the package into the runtime, and they answer an undecided result differently on purpose:

1. **Upload handler — [`internal/api/users_upload.go`](../../api/users_upload.go).** When `PHOTOPRISM_UPLOAD_NSFW=false` (the default, since the flag has no value of its own), every accepted upload is screened by `vision.DetectNSFW` before indexing. Files that are flagged are deleted on the spot — they never reach `originals/`. **This path fails closed:** an upload is admitted only on an explicit `IsSafe()`, because admitting a file the detector could not read leaves exactly the content the check exists to keep out. The one exception is `ErrNotConfigured`, which admits — that is the operator having turned screening off, and rejecting would delete every upload on those instances. The mismatch is reported once at startup instead.

2. **Index + vision-worker pipelines — [`internal/photoprism/index_mediafile.go`](../../photoprism/index_mediafile.go), [`internal/workers/vision.go`](../../workers/vision.go), [`internal/workers/meta.go`](../../workers/meta.go).** When `PHOTOPRISM_DETECT_NSFW=true` (default `false`), the indexer marks new photos as `PhotoPrivate = true` if the model flags them. **Indexing fails neutral:** an undecided result logs a warning and writes nothing, because marking a whole library private on a missing model file would be the worse outcome and an operator could not tell those photos apart afterwards. **The vision worker fails closed:** an undecided result never changes the flag, so `vision run --force`, which re-examines photos that are already private, cannot un-private them when the detector is broken.

Both flags are independent: you can reject uploads without flagging existing imports, flag existing imports without policing uploads, or both. The user-facing matrix lives at [docs.photoprism.app/user-guide/ai/nsfw/](https://docs.photoprism.app/user-guide/ai/nsfw/).

### Detection Through the Labels Model

When `Type: labels` is served by an Ollama or OpenAI engine and **both** `PHOTOPRISM_DETECT_NSFW=true` and `PHOTOPRISM_EXPERIMENTAL=true` are set, [`internal/config/config.go`](../../config/config.go) flips the package-level global `vision.DetectNSFWLabels` to `true`. The Ollama and OpenAI engine builders then swap their default label prompts for `LabelPromptNSFW` and the JSON schema generators add `nsfw` + `nsfw_confidence` fields, so NSFW classification piggybacks on the label-generation call instead of running as a separate inference pass.

When the shortcut is active, the labels-path check in `index_mediafile.go` (`labels.IsNSFW(threshold)`) can promote a photo to private without this package being touched. The dedicated TensorFlow model in `internal/ai/nsfw` is still used as a fallback whenever the labels path either does not run or does not return NSFW signals, and whenever `vision run --models nsfw` is invoked directly.

### How It Works

- **Model Loading** — Loads the NSFW SavedModel from `assets/models/` and resolves input/output ops (inferred if missing).
- **Input Preparation** — JPEG thumbnails (default size `Fit720`, see `MediaFile.DetectNSFW`) are decoded and transformed to the configured input resolution.
- **Inference & Output** — Produces five class probabilities, which `getScores` validates and maps positionally. A graph of another width is rejected rather than read under the wrong class names.
- **Decision** — `Decide` reduces the classes to `max(Porn, Sexy, Hentai)` and compares that to the threshold. `Neutral` is not consulted: it used to veto a detection whenever it exceeded `0.25`, which a five-way softmax makes unreachable above a threshold of `0.5` and, below that, only ever suppressed the detections a lower threshold was asking for.

### Threshold

`vision.yml` carries a `Thresholds.NSFW` value (range `0-100`) that controls how confident the model must be before a picture is flagged. Lower values are more aggressive; higher values more permissive. It governs the dedicated NSFW model and the NSFW fields returned via the label-generation shortcut alike.

```yaml
Thresholds:
  NSFW: 75
```

Left unset it resolves to `75` (`vision.DefaultNSFWThreshold`). Unset is a distinct state rather than a synonym for 75: it is what lets a model's own calibrated default apply, since a threshold tuned for one model's output distribution says nothing about where another puts its decision boundary. `Thresholds.NSFWIsSet` is the predicate that tells the two apart.

> ⚠ **Instances upgrading from a release before this threshold was honored were flagging at a hardcoded `0.98` while indexing**, not at the documented value. Detection is therefore more aggressive by default after the change, and photos scoring between the two are flagged on the next index or `vision run --force`. Set `Thresholds: {NSFW: 98}` to keep the previous behavior. There is no way to reverse this per photo — `photos.photo_private` is a single boolean with no record of what set it.

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
