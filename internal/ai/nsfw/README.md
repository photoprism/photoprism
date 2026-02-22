## PhotoPrism — NSFW Package

**Last Updated:** December 23, 2025

### Overview

`internal/ai/nsfw` runs the built-in TensorFlow NSFW classifier to score images for drawing, hentai, neutral, porn, and sexy content. It is used during indexing and metadata workflows when the NSFW model is enabled.

### How It Works

- **Model Loading** — Loads the NSFW SavedModel from `assets/models/` and resolves input/output ops (inferred if missing).
- **Input Preparation** — JPEG images are decoded and transformed to the configured input resolution.
- **Inference & Output** — Produces five class probabilities mapped into a `Result` struct for downstream thresholds and UI badges.

### Memory & Performance

TensorFlow tensors allocate C memory and are freed by Go GC finalizers. To keep RSS bounded during long runs, PhotoPrism periodically triggers garbage collection to return freed tensor memory to the OS. Tune with:

- `PHOTOPRISM_TF_GC_EVERY` (default **200**, `0` disables).  
  Lower values reduce peak RSS but increase GC overhead and can slow indexing.

### Troubleshooting Tips

- **Model fails to load:** Verify `saved_model.pb` and `variables/` exist under the model path.
- **Unexpected scores:** Confirm the input resolution matches the model and that logits are handled correctly.
- **High memory usage:** Adjust `PHOTOPRISM_TF_GC_EVERY` or reduce concurrent indexing load.

### Related Docs

- [`internal/ai/vision/README.md`](../vision/README.md) — model registry and run scheduling
- [`internal/ai/tensorflow/README.md`](../tensorflow/README.md) — TensorFlow helpers, GC behavior, and model loading
