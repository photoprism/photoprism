## PhotoPrism — Classification Package

**Last Updated:** September 2, 2026

### Overview

`internal/ai/classify` runs fixed-taxonomy image classification through ONNX Runtime. It decodes an image, applies the preprocessing declared for the selected model, executes one output tensor, converts raw logits with stable softmax, and maps the resulting probabilities through the existing label rules.

The default and optional ImageNet-1k candidates share the existing 1000-entry vocabulary in `assets/models/nasnet/labels.txt`. No label index, rule, stored label, or `classify.Labels` consumer changes when the model changes.

### Registered Models

`models.go` is the classifier-specific registry. Every entry supplies a checksum-pinned `onnx.ModelInfo`, label filename, and canonical-order flag. The shared ONNX description records:

- artifact filename, SHA-256, license, and quantization;
- input/output tensor names, shape, count, and whether output values are logits;
- NCHW/NHWC layout, RGB/BGR order, per-channel mean and standard deviation;
- resize mode, short edge, crop ratio, and interpolation.

The graph is inspected at initialization and must agree with all recorded structural fields. A registered checksum mismatch, multiple outputs, a dynamic output width, a 1001-class background offset, non-finite output, or a label-count mismatch disables that labels model instead of substituting another named model.

### Configuration

`PHOTOPRISM_LABEL_MODEL` accepts `auto`, `none`, a registered name, or a custom model name. `auto` resolves to the bundled default, while `none` disables local classification. `photoprism config` reports `label-model`, `label-model-path`, and `label-model-runtime`.

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

The offline exporter downloads an immutable publisher checkpoint, records the source and artifact SHA-256 values, exports a fixed `[1, 3, 224, 224]` FP32 graph at opset 17, embeds preprocessing/provenance metadata, runs the ONNX checker and shape inference, and compares a normalized fixture through PyTorch and ONNX Runtime:

```bash
scripts/ai/export-label-models.py --model all \
  --exported-at 2026-09-02T04:30:00Z
```

Its manifest records the command, fixed metadata timestamp, and Python, PyTorch, timm, ONNX, ONNX Runtime, and NumPy versions. Supplying the same timestamp prevents wall-clock metadata from changing an otherwise identical graph checksum. RepViT is reparameterized before export, and distilled models must return one final combined logits tensor in eval mode. The embedded license defaults to `unverified`; pass `--license` only after the weight-specific review is complete.

### Benchmarking & Calibration

Before TensorFlow is removed from an environment, capture the incumbent NASNet output for a corpus with the opt-in build tag:

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

### Troubleshooting

- **The named model is disabled:** Check the reported model path, file SHA-256, ONNX Runtime installation, and warning log. A named selection never falls back to different weights.
- **The model output does not match labels:** Provide the exact `LabelFile`; shifted indices are rejected rather than accepted silently.
- **Confidence behavior changed:** Use the benchmark’s threshold maps and rule-activation drift. Do not reuse the NASNet threshold without corpus calibration.
- **A custom model produces poor labels:** Declare its color order, normalization, resize/crop convention, output type, and label file in `vision.yml` or embedded metadata.

### Related Docs

- [`internal/ai/onnx/README.md`](../onnx/README.md) — shared ONNX model descriptions and runtime setup
- [`internal/ai/vision/README.md`](../vision/README.md) — `vision.yml` model configuration
- [`specs/intelligence/onnx-label-generation.md`](../../../specs/intelligence/onnx-label-generation.md) — selection gates and candidate research
