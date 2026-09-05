## PhotoPrism — NSFW Package

**Last Updated:** September 5, 2026

### Overview

`internal/ai/nsfw` runs local NSFW classifiers through ONNX Runtime and reduces model-specific outputs to one unsafe probability. AdamCodd ViT-Base INT8 is the bundled default; FP32, Falconsai, Freepik, Yahoo OpenNSFW, custom ONNX graphs, and remote vision services use the same result contract.

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

1. **Upload handler — [`internal/api/users_upload.go`](../../api/users_upload.go).** When `PHOTOPRISM_UPLOAD_NSFW=false` (the default, since the flag has no value of its own), supported visual uploads are screened before indexing. Files that are flagged are deleted on the spot — they never reach `originals/`. An upload is admitted only after the configured detector returns an explicit safe decision.

2. **Index + vision-worker pipelines — [`internal/photoprism/index_mediafile.go`](../../photoprism/index_mediafile.go), [`internal/workers/vision.go`](../../workers/vision.go), [`internal/workers/meta.go`](../../workers/meta.go).** When `PHOTOPRISM_DETECT_NSFW=true` (default `false`), the indexer marks new photos as `PhotoPrivate = true` if the model flags them. **Indexing fails neutral:** an undecided result logs a warning and writes nothing, because marking a whole library private on a missing model file would be the worse outcome and an operator could not tell those photos apart afterwards. **The vision worker fails closed:** an undecided result never changes the flag, so `vision run --force`, which re-examines photos that are already private, cannot un-private them when the detector is broken.

Both flags are independent: you can reject uploads without flagging existing imports, flag existing imports without policing uploads, or both. The user-facing matrix lives at [docs.photoprism.app/user-guide/ai/nsfw/](https://docs.photoprism.app/user-guide/ai/nsfw/).

### Detection Through the Labels Model

When `Type: labels` is served by an Ollama or OpenAI engine and **both** `PHOTOPRISM_DETECT_NSFW=true` and `PHOTOPRISM_EXPERIMENTAL=true` are set, [`internal/config/config.go`](../../config/config.go) flips the package-level global `vision.DetectNSFWLabels` to `true`. The Ollama and OpenAI engine builders then swap their default label prompts for `LabelPromptNSFW` and the JSON schema generators add `nsfw` + `nsfw_confidence` fields, so NSFW classification piggybacks on the label-generation call instead of running as a separate inference pass.

When the shortcut is active, the labels-path check in `index_mediafile.go` (`labels.IsNSFW(threshold)`) can promote a photo to private without this package being touched. The dedicated ONNX model is still used whenever the labels path does not return NSFW signals and whenever `vision run --models nsfw` is invoked directly.

### How It Works

- **Model Loading** — Verifies the artifact checksum and graph shape before creating one shared ONNX Runtime session.
- **Input Preparation** — Applies the registered input size, RGB/BGR order, normalization, interpolation, and resize convention. JPEG, PNG, GIF, BMP, TIFF, and WebP decode directly; upload screening creates a request-scoped JPEG preview for other visual media.
- **Inference & Output** — Validates the output width and rejects non-finite values. Binary logits use softmax or sigmoid, while Freepik uses the complement of its neutral class.
- **Decision** — Compares the reduced unsafe probability with the explicit operator threshold or the selected model's default. The comparison never uses argmax.

### Model Selection

`PHOTOPRISM_NSFW_MODEL` accepts `auto`, `none`, or a registered model name. `auto` resolves to the bundled default, while `none` disables the local detector. `photoprism config` reports the resolved name, artifact path, and runtime.

### Threshold

`vision.yml` carries a `Thresholds.NSFW` value (range `0-100`) that controls how confident the model must be before a picture is flagged. Lower values are more aggressive; higher values more permissive. It governs the dedicated NSFW model and the NSFW fields returned via the label-generation shortcut alike.

```yaml
Thresholds:
  NSFW: 98
```

Left unset, local ONNX indexing uses the conservative provisional model default of `98` until the representative corpus review is complete. Upload screening keeps its established `75` operating point. An explicit `Thresholds.NSFW` value overrides both. Unset is a distinct state because a threshold tuned for one model's output distribution does not transfer to another model.

### Calibration & Benchmarking

The immutable publisher revisions and export verification live in `scripts/ai/export-nsfw-models.py`. It copies the published AdamCodd graphs, exports Falconsai and Freepik from their publisher checkpoints, converts Yahoo from its publisher Caffe graph, and records source hashes, artifact hashes, environment versions, preprocessing, and framework-to-ONNX differences. For example:

```sh
python scripts/ai/export-nsfw-models.py --model all \
  --output-dir assets/models
```

The opt-in benchmark reads an annotated JSON corpus, runs every requested model in a separate process, and writes hardware-specific quality and performance metrics:

```sh
PHOTOPRISM_TEST_NSFW_CORPUS=/path/to/corpus.json \
PHOTOPRISM_TEST_NSFW_REPORT=/path/to/report.json \
go test ./internal/ai/nsfw -run '^TestExternalNSFWBenchmark$' -count=1 -v
```

Run it on both x86-64 and ARM64. The report includes load time, p50/p95 latency, peak RSS, artifact size, unsafe recall, benign false-positive rate, average precision, AUROC, Brier score, expected calibration error, operating points, and newly-safe/newly-unsafe identities when incumbent scores are present. `Example_nsfwBenchmarkCorpus` in `benchmark_external_test.go` documents the manifest shape. The checked-in unit corpus is only a smoke test; selecting the default threshold requires the representative reviewed corpus described in the intelligence specification.

Generate the incumbent TensorFlow scores before running the ONNX comparison:

```sh
PHOTOPRISM_TEST_NSFW_BASELINE_DIR=/path/to/jpeg-corpus \
PHOTOPRISM_TEST_NSFW_BASELINE_REPORT=/path/to/corpus.json \
go test -tags nsfwbaseline ./internal/ai/nsfw \
  -run '^TestGenerateTensorFlowNSFWBaseline$' -count=1 -v
```

### Troubleshooting Tips

- **Model fails to load:** Run `scripts/dist/download-models.sh <model-name>` and verify the reported checksum.
- **Unexpected scores:** Confirm the input resolution matches the model and that logits are handled correctly.
- **High memory usage:** Select an INT8 model or reduce concurrent indexing load.
- **NSFW detection appears to stop working after switching labels to an LLM:** Confirm both `PHOTOPRISM_DETECT_NSFW=true` and `PHOTOPRISM_EXPERIMENTAL=true` are set. Without both, the labels-path shortcut is disabled and only an explicit `vision run --models nsfw` (or another caller that goes through this package directly) will produce NSFW flags.

### Related Docs

- [`internal/ai/vision/README.md`](../vision/README.md) — model registry, run scheduling, and the `DetectNSFWLabels` global
- [`internal/ai/vision/ollama/README.md`](../vision/ollama/README.md) — Ollama engine: `LabelPromptNSFW` swap-in
- [`internal/ai/vision/openai/README.md`](../vision/openai/README.md) — OpenAI engine: NSFW-aware prompt and schema
- [`internal/ai/vision/schema/README.md`](../vision/schema/README.md) — JSON schema variants used when NSFW is enabled
- [docs.photoprism.app/user-guide/ai/nsfw/](https://docs.photoprism.app/user-guide/ai/nsfw/) — user-facing reference + flag matrix
