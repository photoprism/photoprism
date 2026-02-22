## PhotoPrism — TensorFlow Package

**Last Updated:** December 23, 2025

### Overview

`internal/ai/tensorflow` provides the shared TensorFlow helpers used by PhotoPrism’s built-in AI features (labels, NSFW, and FaceNet embeddings). It wraps SavedModel loading, input/output discovery, image tensor preparation, and label handling so higher-level packages can focus on domain logic.

### Key Components

- **Model Loading** — `SavedModel`, `GetModelTagsInfo`, and `GetInputAndOutputFromSavedModel` discover and load SavedModel graphs with appropriate tags.
- **Input Preparation** — `Image`, `ImageTransform`, and `ImageTensorBuilder` convert JPEG images to tensors with the configured resolution, color order, and resize strategy.
- **Output Handling** — `AddSoftmax` can insert a softmax op when a model exports logits.
- **Labels** — `LoadLabels` loads label lists for classification models.

### Model Loading Notes

- Built-in models live under `assets/models/` and are accessed via helpers in `internal/ai/vision` and `internal/ai/classify`.
- When a model lacks explicit tags or signatures, the helpers attempt to infer input/output operations. Logs will show when inference kicks in.
- Classification models may emit logits; if `ModelInfo.Output.Logits` is true, a softmax op is injected at load time.

### Memory & Garbage Collection

TensorFlow tensors are allocated in C memory and freed by Go GC finalizers in the TensorFlow bindings. Long-running inference can therefore show increasing RSS even when the Go heap is small. PhotoPrism periodically triggers garbage collection to return freed C-allocated tensor buffers to the OS. Control this behavior with:

- `PHOTOPRISM_TF_GC_EVERY` (default **200**, `0` disables).  
  Lower values reduce peak RSS but increase GC overhead and can slow indexing.

### Troubleshooting Tips

- **Model fails to load:** Verify the SavedModel path, tags, and that `saved_model.pb` plus `variables/` exist under `assets/models/<name>`.
- **Input/output mismatch:** Check logs for inferred inputs/outputs and confirm `vision.yml` overrides (name, resolution, and `TensorFlow.Input/Output`).
- **Unexpected probabilities:** Ensure logits are handled correctly and labels match output indices.
- **High memory usage:** Confirm `PHOTOPRISM_TF_GC_EVERY` is set appropriately; model weights remain resident for the life of the process by design.

### Related Docs

- [`internal/ai/vision/README.md`](../vision/README.md) — model registry, `vision.yml` configuration, and run scheduling
- [`internal/ai/face/README.md`](../face/README.md) — FaceNet embeddings and face-specific tuning
- [`internal/ai/classify/README.md`](../classify/README.md) — classification workflow using TensorFlow helpers
- [`internal/ai/nsfw/README.md`](../nsfw/README.md) — NSFW model usage and result mapping
