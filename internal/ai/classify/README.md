## PhotoPrism — Classification Package

**Last Updated:** December 23, 2025

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

### Troubleshooting Tips

- **Labels are empty:** Verify the model labels file and that `Rules` thresholds are not too strict.
- **Model load failures:** Ensure `saved_model.pb` and `variables/` exist under the configured model path.
- **Unexpected outputs:** Check `TensorFlow.Input/Output` settings in `vision.yml` for custom models.

### Related Docs

- [`internal/ai/vision/README.md`](../vision/README.md) — model registry and `vision.yml` configuration
- [`internal/ai/tensorflow/README.md`](../tensorflow/README.md) — TensorFlow helpers, GC behavior, and model loading
