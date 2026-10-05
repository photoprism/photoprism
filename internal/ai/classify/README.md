## PhotoPrism — Classification Package

**Last Updated:** October 5, 2026

### Overview

`internal/ai/classify` runs fixed-taxonomy image classification through ONNX Runtime. It decodes an image, applies the preprocessing declared for the selected model, executes one output tensor, converts raw logits with stable softmax, and maps the resulting probabilities through the existing label rules.

The default and optional ImageNet-1k candidates share the 1000-entry vocabulary embedded from `internal/ai/classify/labels.txt`. It remains readable and diffable in the repository and does not depend on a model directory at runtime. No label index, rule, stored label, or `classify.Labels` consumer changes when the model changes.

### Photo Inputs

The bundled S2 classifier uses a `tile_224` center input and a whole-photo input with distortion capped at 4:3. Both become 224×224 pixels. Beyond 4:3, the whole-photo input crops the excess from both ends before resampling; within 4:3, it retains the whole photo. Its source is an existing whole-photo rendition with at least 224 pixels on the short side, or the original, never `fit_720`.

Exact square images use one whole-square input; squares no larger than 224 pixels use the original directly. Images with either decoded dimension below 224 pixels also use one input: an aspect-ratio-preserving center crop from the original, resized to the model dimensions. Rotation preserves the square and short-side decisions. Prepared inputs explicitly bypass S2's native resize/crop while retaining RGB/NCHW ImageNet normalization. Other models and remote services keep their own preprocessing.

The global confidence floor defaults to 20%; an explicit `Thresholds.Confidence` in `vision.yml` takes precedence. Each input retains at most five qualifying labels, then the existing merge combines them using maximum confidence without an additional per-photo cap.

### Label Rules

`rules.yml` maps lowercase raw class names to visible labels, confidence minimums, categories and priorities. Class meanings stay distinct: `cardigan` is clothing and uses the fashion/Portrait rule, while `cardigan dog` is the dog breed. The hardware class `nail` is ignored. Rules with priority -3 or lower are excluded even when a probability reaches exactly 1. Schipperke aliases require a raw probability of at least 0.995 (99.5%); class minimums compare float32 probabilities before visible-label confidence is rounded.

Regenerate `rules.go` with `go generate ./internal/ai/classify`. Generation rejects duplicate keys, unknown fields, uppercase names, missing alias targets and aliases that do not reference a direct rule. The package tests verify that the generated map matches the complete YAML source. The generator and its rejection tests are a normally compiled package; run them with `go test ./internal/ai/classify/gen`.

### Registered Models

`models.go` is the classifier-specific registry. Every entry supplies a checksum-pinned `onnx.ModelInfo`, label filename, and canonical-order flag. The shared ONNX description records:

- artifact filename, SHA-256, license, and quantization;
- input/output tensor names, shape, count, and whether output values are logits;
- NCHW/NHWC layout, RGB/BGR order, per-channel mean and standard deviation;
- resize mode, short edge, crop ratio, and interpolation.

The graph is inspected at initialization and must agree with all recorded structural fields. A registered checksum mismatch, multiple outputs, a dynamic output width, a 1001-class background offset, non-finite output, or a label-count mismatch prevents inference instead of substituting another model. Initialization errors are cached separately from operator disablement; restart PhotoPrism after installing or repairing an artifact.

### Configuration

`PHOTOPRISM_LABELS_MODEL` accepts `auto` and `none`. In `auto` mode, `vision.yml` chooses the registered, custom, or remote labels model. A `Default: true` entry, or no labels entry, selects the first installed registered model in preference order, starting with `efficientformerv2_s2`. If no artifact is installed, the entry stays enabled and a startup warning provides the download command; installing it requires a restart. `none` disables labels regardless of `vision.yml` without persisting that override. The deprecated `PHOTOPRISM_DISABLE_CLASSIFICATION` applies unless `PHOTOPRISM_LABELS_MODEL` is set to `auto` or `none`; explicit `auto` overrides it, while an unsupported value does not.

Select a registered alternative in `vision.yml`:

```yaml
Models:
  - Type: labels
    Name: repvit_m1_0
```

The default model `efficientformerv2_s2` runs during indexing, whether it is selected automatically or named in `vision.yml`. Alternative label models run after indexing by default; set `Run: on-index` on the entry to run inline. `photoprism vision ls` reports the selected models, effective enabled status, and whether their artifacts are installed; remote models show installation status `n/a`.

A custom model is resolved under `PHOTOPRISM_MODELS_PATH` as:

```text
<models path>/<name>/<name>.onnx
<models path>/<name>/labels.txt
```

Use `vision.yml` to override `Path`, `LabelFile`, `Resolution`, or `ONNX` fields. Output width is read from the graph and must exactly equal the selected label file, so ImageNet-21k and other vocabularies are supported without a hard-coded class count. Embedded `photoprism.*` metadata can carry the preprocessing contract; missing semantic fields use logged ImageNet defaults.

### Inference Safety

- Raw logits use maximum-subtracted softmax before exponentiation.
- Probabilities must be finite, in range, and sum to 1 within `1e-4`.
- Tensor input follows the model-specific channel order and memory layout.
- Inference is serialized on a model session so parallel indexing is deterministic.
- Session and tensor resources are destroyed explicitly.

### Exporting Candidates

The offline exporter downloads an immutable Apache-2.0 publisher checkpoint, records the source and artifact SHA-256 values, exports a fixed `[1, 3, 224, 224]` FP32 graph at opset 17, embeds preprocessing/provenance metadata, runs the ONNX checker and shape inference, and compares a normalized fixture through PyTorch and ONNX Runtime:

```bash
scripts/ai/export-label-models.py --model all \
  --license Apache-2.0 \
  --exported-at 2026-09-02T04:30:00Z
```

Its manifest records the command, fixed metadata timestamp, and Python, PyTorch, timm, ONNX, ONNX Runtime, and NumPy versions. Supplying the same timestamp prevents wall-clock metadata from changing an otherwise identical graph checksum. RepViT is reparameterized before export, and distilled models must return one final combined logits tensor in eval mode. Pass the reviewed license with `--license Apache-2.0` so newly exported metadata records the same publisher terms as the registry.

### Benchmarking & Calibration

Before TensorFlow is removed from an environment, capture the incumbent NASNet output for a corpus with the opt-in build tag. The baseline loads the TensorFlow model from `assets/models/nasnet`, which no install target provides, so extract `https://dl.photoprism.app/tensorflow/nasnet.zip` there first:

```bash
PHOTOPRISM_TEST_LABEL_BASELINE_DIR=/photos/corpus \
PHOTOPRISM_TEST_LABEL_BASELINE_REPORT=/reports/nasnet.json \
go test -tags labelbaseline ./internal/ai/classify \
  -run TestGenerateTensorFlowLabelBaseline -count=1
```

Compare installed ONNX candidates on each target architecture:

```bash
PHOTOPRISM_TEST_LABEL_CORPUS=/reports/nasnet.json \
PHOTOPRISM_TEST_LABEL_REPORT=/reports/onnx-arm64.json \
go test ./internal/ai/classify -run TestExternalLabelBenchmark -count=1
```

The report includes top-5 overlap, visible-label agreement, rule-activation drift, threshold crossings and calibration points, p50/p95 latency, model load time, Linux peak RSS, artifact size, and optional correct/false-positive counts when the manifest contains human annotations. The harness runs each candidate in a separate process so peak RSS is model-specific. Repeat the comparison on x86-64 and ARM64 with a representative photo corpus.

EfficientFormerV2 S2 is the default because the reviewed 402-image Wikimedia corpus reached 81.8% visible-label coverage, compared with 73.4% for S1, leaving 73 rather than 107 images unlabeled. Its higher ARM64 latency (80 ms p50 and 121 ms p95, versus 52 ms and 75 ms) and peak RSS (350 MB versus 319 MB) are accepted for the materially better indexing coverage and mean top-1 quality.

### Corpus Calibration

Raw-vector capture and crop/rule experiments are developer tooling, not ordinary package tests.
Calibration runs require an explicitly selected corpus; their images and reports stay outside the repository.
The regular suite uses repository fixtures and synthetic data, without private calibration inputs.

### Troubleshooting

- **The selected model cannot initialize:** Check the reported model path, file SHA-256, ONNX Runtime installation, and warning log. A named selection never falls back to different weights.
- **The model output does not match labels:** Provide the exact `LabelFile`; shifted indices are rejected rather than accepted silently.
- **Confidence behavior changed:** Use the benchmark’s threshold maps and rule-activation drift. Do not reuse the NASNet threshold without corpus calibration.
- **A custom model produces poor labels:** Declare its color order, normalization, resize/crop convention, output type, and label file in `vision.yml` or embedded metadata.

### Related Docs

- [`internal/ai/onnx/README.md`](../onnx/README.md) — shared ONNX model descriptions and runtime setup
- [`internal/ai/vision/README.md`](../vision/README.md) — `vision.yml` model configuration
