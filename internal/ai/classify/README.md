## PhotoPrism — Classification Package

**Last Updated:** March 6, 2026

### Overview

`internal/ai/classify` wraps PhotoPrism’s TensorFlow-based image classification (labels). It loads SavedModel classifiers (Nasnet by default), prepares inputs, runs inference, and maps output probabilities to label rules.

### How It Works

- **Model Loading** — The classifier loads a SavedModel under `assets/models/<name>` and resolves model tags and input/output ops (see `vision.yml` overrides for custom models).
- **Input Preparation** — JPEGs are decoded and resized/cropped to the model’s expected input resolution.
- **Inference** — The model outputs probabilities; `Rules` apply thresholds and priority to produce final labels.

### Memory & Performance

TensorFlow tensors allocate C memory and are freed by Go GC finalizers. To keep RSS bounded during long runs, PhotoPrism periodically triggers garbage collection to return freed tensor memory to the OS. Tune with:

- `PHOTOPRISM_TF_GC_EVERY` (default **200**, `0` disables).  
  Lower values reduce peak RSS but increase GC overhead and can slow indexing.

### Go 1.26 JPEG Decoder Impact

After the base image and toolchain upgrade on February 20, 2026, we observed measurable drift in TensorFlow label uncertainty values caused by changes in Go's `image/jpeg` implementation:

- **Direct Evidence** — The `ChameleonLimeJpg` fixture shifted from uncertainty `7` with Go `1.25.4` to `8` with Go `1.26.0` for the same model and inputs.
- **Pipeline Relevance** — Classification input decoding goes through `imaging.Decode(...)` in `model.go`, which uses Go's registered JPEG decoder.
- **Fixture Scan Result** — 55/55 JPEG fixtures in `assets/samples` decoded successfully on both versions (no compatibility failures), but all produced different decoded pixel hashes between Go `1.25.4` and `1.26.0`.
- **Output Stability** — In sampled tests, top labels remained stable (`chameleon`, `cat`, etc.), while confidence and uncertainty values moved slightly.

Operational notes:

- Prefer tolerance-based assertions (`assert.InDelta`) for JPEG-derived uncertainty/confidence tests instead of exact integer equality.
- Avoid bit-for-bit JPEG expectations in tests unless the codec/toolchain is pinned and intentionally version-locked.

### Troubleshooting Tips

- **Labels are empty:** Verify the model labels file and that `Rules` thresholds are not too strict.
- **Model load failures:** Ensure `saved_model.pb` and `variables/` exist under the configured model path.
- **Unexpected outputs:** Check `TensorFlow.Input/Output` settings in `vision.yml` for custom models.

### Related Docs

- [`internal/ai/vision/README.md`](../vision/README.md) — model registry and `vision.yml` configuration
- [`internal/ai/tensorflow/README.md`](../tensorflow/README.md) — TensorFlow helpers, GC behavior, and model loading
