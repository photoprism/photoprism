## PhotoPrism — NSFW Package

**Last Updated:** October 1, 2026

### Overview

`internal/ai/nsfw` runs local NSFW classifiers through ONNX Runtime and reduces model-specific outputs to one unsafe probability. Yahoo OpenNSFW is the bundled default; AdamCodd FP32/INT8, Falconsai, Freepik, custom ONNX graphs, and remote vision services use the same result contract.

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

1. **Upload handler — [`internal/api/users_upload.go`](../../api/users_upload.go).** When `PHOTOPRISM_UPLOAD_NSFW=false` (the default, since the flag has no value of its own), supported visual uploads are screened before indexing. If any file is flagged, the entire temporary batch is deleted and the request returns `403`. If an enabled detector cannot decide, the handler logs a warning and admits the upload. A detector explicitly disabled or not configured also admits the upload.

2. **Index + vision-worker pipelines — [`internal/photoprism/index_mediafile.go`](../../photoprism/index_mediafile.go), [`internal/workers/vision.go`](../../workers/vision.go).** When `PHOTOPRISM_DETECT_NSFW=true` (default `false`), the indexer marks new photos as `PhotoPrivate = true` if the model flags them. **Indexing fails neutral:** an undecided result logs a warning and writes nothing, because marking a whole library private on a missing model file would be the worse outcome and an operator could not tell those photos apart afterwards. **The vision worker fails closed:** an undecided result never changes the flag, so `vision run --force`, which re-examines photos that are already private, cannot un-private them when the detector is broken.

Both flags are independent: you can reject uploads without flagging existing imports, flag existing imports without policing uploads, or both. The user-facing matrix lives at [docs.photoprism.app/user-guide/ai/nsfw/](https://docs.photoprism.app/user-guide/ai/nsfw/).

### Detection Through the Labels Model

When `Type: labels` is served by an Ollama or OpenAI engine, `PHOTOPRISM_DETECT_NSFW=true`, and `PHOTOPRISM_NSFW_MODEL=labels`, [`internal/config/config.go`](../../config/config.go) flips the package-level global `vision.DetectNSFWLabels` to `true`. The Ollama and OpenAI engine builders then swap their default label prompts for `LabelPromptNSFW` and the JSON schema generators add `nsfw` + `nsfw_confidence` fields, so NSFW classification piggybacks on the label-generation call instead of running as a separate inference pass.

When the shortcut is active, the labels-path check in `index_mediafile.go` (`labels.IsNSFW(threshold)`) can promote a photo to private without this package being touched. In `labels` mode no dedicated detector is loaded, and uploads are admitted without screening. If `UPLOAD_NSFW=false`, a startup system warning explains the missing screening. A local classifier or disabled labels model cannot provide NSFW fields, so a second warning reports that no NSFW detection takes place. In `auto` and `none` modes label responses never change the private flag, including responses to custom prompts.

### How It Works

- **Model Loading** — Verifies the artifact checksum and graph shape before creating one shared ONNX Runtime session.
- **Input Preparation** — Applies the registered input size, RGB/BGR order, normalization, interpolation, and resize convention. JPEG, PNG, GIF, BMP, TIFF, and WebP decode directly; upload screening creates a request-scoped JPEG preview for other visual media.
- **Inference & Output** — Validates the output width and rejects non-finite values. Binary logits use softmax or sigmoid, while Freepik uses the complement of its neutral class.
- **Decision** — Compares the reduced unsafe probability with the explicit operator threshold or the selected model's default. The comparison never uses argmax.

### Model Selection

`PHOTOPRISM_NSFW_MODEL` accepts `auto`, `none`, and `labels`. In `auto` mode `vision.yml` chooses the registered, custom, or remote detector. A `Default: true` entry, or no entry, selects the first installed registered model in preference order, starting with Yahoo OpenNSFW. When no artifact is installed, the default entry remains enabled; startup logs provide the download command and restart instruction. `none` disables dedicated and label-derived detection. `labels` uses only Ollama/OpenAI labels responses when `DETECT_NSFW=true`. Mode overrides are not persisted as user disablement by `vision save`.

```yaml
Models:
  - Type: nsfw
    Name: falconsai_nsfw_image_detection_224
```

All local ONNX detectors, including custom models, run during indexing with `Run: auto`; explicit `Run` settings take precedence. Remote detectors retain their background scheduling. `photoprism vision ls` reports effective enabled status and artifact presence without loading inference sessions.

### Threshold

`vision.yml` can configure the dedicated detector separately for upload screening and indexing, as well as the threshold at or above which a labels model flags NSFW content. Thresholds are whole percentages from `1` through `100`; values above `100` are treated as `100`. Lower values are more aggressive and higher values more permissive.

```yaml
Thresholds:
  NSFW: 75
  NSFWUpload: -1
  NSFWIndex: -1
```

`NSFWUpload` controls the dedicated detector in the upload handler, and `NSFWIndex` controls the dedicated detector in indexing and vision-worker runs. When either is omitted, `0`, or `-1`, that path uses the selected detector's calibrated threshold. `NSFW` applies only to labels models, i.e. the NSFW confidence returned by Ollama or OpenAI labels models with `PHOTOPRISM_NSFW_MODEL=labels`; it never changes the dedicated detector. When `NSFW` is omitted or not positive, it is `75`.

In automatic mode, the local dedicated ONNX detector uses the selected model's calibrated fallback threshold: AdamCodd FP32 uses `71.8`, AdamCodd INT8 uses `76.0`, Falconsai uses `52.9`, Freepik uses `99.2`, and Yahoo OpenNSFW uses `32.7`. Remote detector results use the fallback of `75` because they do not expose a local detector calibration. Automatic selection is a distinct state because a threshold tuned for one model's output distribution does not transfer to another model.

Custom detectors must declare `Reduction`. `softmax-unsafe` additionally requires `UnsafeClassIndex`, while `neutral-complement` requires `NeutralClassIndex`; zero is accepted only when it is explicitly present. `sigmoid-unsafe` reduces its single output without a class index. `DefaultThreshold` is the custom detector's automatic fallback probability from 0 to 1; when omitted, it falls back to `0.98`.

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

Run it on both x86-64 and ARM64. The report includes load time, p50/p95 latency, peak RSS, artifact size, unsafe recall, benign false-positive rate, average precision, AUROC, Brier score, expected calibration error, operating points, and newly-safe/newly-unsafe identities when incumbent scores are present. Threshold selection minimizes false positives among points that meet the requested recall; if none meet it, the highest-accuracy point is returned instead. `Example_nsfwBenchmarkCorpus` in `benchmark_external_test.go` documents the manifest shape. The checked-in unit corpus is only a smoke test; selecting the default threshold requires the representative reviewed corpus described in the intelligence specification.

The registered fallbacks were calibrated on 2,500 images: the manually reviewed [SIMAS sexual-content subset](https://zenodo.org/records/15423637) contributes 500 unsafe and 500 matched-safe photographs, while source-labeled adult illustration, hentai, and neutral sets contribute 500 images each. Because the expanded corpus contains 1,500 unsafe and 1,000 safe images, selection maximizes balanced accuracy instead of raw accuracy. The selected millesimal points stay within 0.3 percentage points of maximum balanced accuracy while preferring recall. Five-fold group-stratified validation checks that the operating points are not artifacts of a single split.

| Model | Threshold | Recall | False-Positive Rate | Balanced Accuracy | ARM64 p50 |
|:------|----------:|-------:|--------------------:|------------------:|----------:|
| AdamCodd FP32 | 0.718 | 93.4% | 15.7% | 88.9% | 465 ms |
| AdamCodd INT8 | 0.760 | 93.0% | 15.4% | 88.8% | 260 ms |
| Falconsai | 0.529 | 91.8% | **4.4%** | **93.7%** | 156 ms |
| Freepik | 0.992 | 88.7% | 5.4% | 91.7% | 979 ms |
| Yahoo OpenNSFW | 0.327 | **93.2%** | 6.8% | 93.2% | **20 ms** |

The latency figures describe the four-core ARM64 calibration host and are comparative rather than universal. At its historical `0.98` threshold, the previous TensorFlow detector reached 26.4% recall and a 1.6% false-positive rate. At its recalibrated comparison point of `0.468`, it reaches 94.2% recall with a 34.4% false-positive rate. Yahoo remains the default because it stays within 0.5 percentage points of Falconsai's balanced accuracy while running nearly eight times faster and using a much smaller artifact.

Generate the incumbent TensorFlow scores before running the ONNX comparison:

```sh
PHOTOPRISM_TEST_NSFW_BASELINE_DIR=/path/to/jpeg-corpus \
PHOTOPRISM_TEST_NSFW_BASELINE_REPORT=/path/to/corpus.json \
go test -tags nsfwbaseline ./internal/ai/nsfw \
  -run '^TestGenerateTensorFlowNSFWBaseline$' -count=1 -v
```

### Troubleshooting Tips

- **Model fails to load:** Run `scripts/dist/download-models.sh <model-name>` and verify the reported checksum.
- **Model was repaired or installed after an initialization failure:** Restart PhotoPrism. Initialization errors are cached for the process lifetime to avoid repeatedly loading a broken artifact.
- **Unexpected scores:** Confirm the input resolution matches the model and that logits are handled correctly.
- **High memory usage:** Select an INT8 model or reduce concurrent indexing load.
- **NSFW detection appears to stop working after switching labels to an LLM:** Confirm both `PHOTOPRISM_DETECT_NSFW=true` and `PHOTOPRISM_EXPERIMENTAL=true` are set. Without both, the labels-path shortcut is disabled and only an explicit `vision run --models nsfw` (or another caller that goes through this package directly) will produce NSFW flags.

### Related Docs

- [`internal/ai/vision/README.md`](../vision/README.md) — model registry, run scheduling, and the `DetectNSFWLabels` global
- [`internal/ai/vision/ollama/README.md`](../vision/ollama/README.md) — Ollama engine: `LabelPromptNSFW` swap-in
- [`internal/ai/vision/openai/README.md`](../vision/openai/README.md) — OpenAI engine: NSFW-aware prompt and schema
- [`internal/ai/vision/schema/README.md`](../vision/schema/README.md) — JSON schema variants used when NSFW is enabled
- [docs.photoprism.app/user-guide/ai/nsfw/](https://docs.photoprism.app/user-guide/ai/nsfw/) — user-facing reference + flag matrix
