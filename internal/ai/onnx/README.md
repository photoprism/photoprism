## ONNX Model Description

**Last Updated:** September 21, 2026

### Overview

This package holds one description of an ONNX model's parameters, shared by the subsystems that run one. The field list is the description of a graph plus its preprocessing contract, and that does not vary by task, so face detection and face embeddings use the same structure rather than a registry each. Label generation and NSFW detection run on TensorFlow today and can adopt it unchanged when they move.

`ModelInfo` is deliberately shaped after `tensorflow.ModelInfo` in `internal/ai/tensorflow`: an `Input` describing geometry, layout, channel order, normalization, and the resize convention; an `Output` describing name, width, and whether it carries logits; plus artifact-level fields for file, checksum, license, and quantization. Where an artifact is downloaded from is not among them: the install registry in `scripts/dist/download-models.sh` owns the URLs and verifies the same checksums, so a source that moves upstream changes in one place. `internal/ai/tensorflow` is scheduled for deletion once its three consumers migrate, so the two structures sitting side by side is a transitional state rather than an abstraction to unify.

### What Can Be Inferred and What Cannot

`Inspect` reads the structural parameters from the graph: tensor names, input width and height, axis order, whether a dimension is dynamic, and output width. These are safe to fill silently because getting them wrong raises.

Channel order, normalization, and the resize convention are **not present in a graph**, and `Inspect` never fills them. That split is the reason the package exists: a wrong input shape errors immediately, while a wrong channel order or normalization produces a model that loads, runs, and returns plausible output that is quietly worse. A description without them falls back to a documented default that is applied explicitly and logged, never guessed.

`Inspect` and `Metadata` both create a temporary session internally, so inspection costs a model load. Call them when a model is about to be used; do not scan a directory of artifacts with them.

### A Known Model Means a Known Artifact

`VerifyChecksum` confirms a description against the file it was written for. Lookup selects a candidate by name, and names collide across publishers:

- InsightFace's `antelopev2` pack and fal's AuraFace both ship a file called `glintr100.onnx`, and they are different models. Our mirror renames ours to `auraface_v1_glintr100.onnx` for this reason.
- Published figures routinely describe a differently named export than the file a mirror actually serves, so a name match is not an artifact match.

A name-only match applies one model's preprocessing to another model's weights, and because the mismatched fields are exactly the ones that cannot be inferred, it fails quietly. Descriptions without a recorded checksum are accepted so that custom models supplied through `PHOTOPRISM_MODELS_PATH` keep working.

How a mismatch is handled is the consumer's decision, and the two current consumers answer it differently:

| Consumer                             | On mismatch      | Why                                                                                                             |
|:-------------------------------------|:-----------------|:----------------------------------------------------------------------------------------------------------------|
| Face embeddings (`internal/ai/face`) | Refuse to load   | Vectors are persisted, and there is no safe fallback for alignment and normalization                            |
| Face detector (`internal/ai/face`)   | Warn and proceed | Costs recall on the next run rather than a library of incomparable vectors, and its layout comes from the graph |

`VerifyGraph` is the cross-check that follows: a recorded value disagreeing with the graph means one of the two describes a different model, so it aborts rather than reconciling. Dimensions the graph leaves dynamic are not compared.

### Embedded Provenance

`Metadata` reads the `photoprism.` prefixed entries of a model's `metadata_props`. Models we export ourselves should carry their own source, checksum, license, channel order, normalization, and resize convention there.

Metadata inside the artifact survives mirroring, renaming, and being copied into an image in a way that a sibling `version.txt` does not, which is why a model we export records where it came from rather than relying on the script that installed it.

### Runtime

`EnsureRuntime` loads the ONNX Runtime shared library and initializes the global environment; it must succeed before any model is inspected or loaded. `SharedLibraryCandidates` lists the paths it tries, starting with an explicitly configured one.

The `github.com/yalue/onnxruntime_go` binding requests the exact C API version of the headers it vendors, so it fails to initialize against an older shared library. Bumping that module therefore requires a matching `ONNX_DEFAULT_VERSION` and checksum update in `scripts/dist/install-onnx.sh`, plus a rebuild of the base images that ship `libonnxruntime.so`.

#### CUDA Library Installation

Run `make cuda` (alias: `make install-cuda`) from the repository root to install the NVIDIA CUDA runtime libraries and cuDNN through `sudo`. This installs the CUDA dependencies only; selecting the GPU ONNX Runtime build and enabling the CUDA provider are separate steps.

The installer downloads pinned packages directly from NVIDIA and verifies their SHA256 checksums before extraction. It does not use a PhotoPrism package mirror or add an apt repository.

`scripts/dist/install-cuda.sh` stages the complete library set on the destination filesystem before publishing it. Existing files and symlinks are preserved with hardlink snapshots, and each replacement uses an atomic rename. A failed publication or a catchable stop signal restores the prior entries and removes newly introduced ones. If recovery itself fails, the installer reports and retains its private recovery directory. Once the complete set is published, it remains installed while the loader cache is refreshed.

Run `make check-cuda-install` to verify installation and recovery using synthetic packages in a private prefix. The check needs Python 3 and standard Linux tools, but no GPU, downloads, root privileges, or system-library changes. Real CUDA compatibility and inference still require validation in a GPU-enabled environment.

### Execution Provider

`Provider` names the execution provider a session runs on: `cpu`, the default, or `cuda`. It is selected once from `PHOTOPRISM_ONNX_PROVIDER` / `--onnx-provider`, read through `Config.OnnxProvider`, and passed to each model loader. `ParseProvider` resolves an empty value as the default and reports an unknown one rather than failing, so an unusable setting cannot stop inference.

Build a session through `NewSessionConfig`, whose `Options` also serve the metadata call a loader makes first, and create the session with `SessionConfig.NewSession`. A provider that cannot be applied falls back to the CPU with one warning, and `SessionConfig.Provider` reports what is actually in force so a loader can log it.

Three properties are worth knowing before changing this:

- **Every step that opens a session goes through `SessionConfig.WithFallback`,** not just the inference session. Reading the graph opens one too, and it is where the provider's cost is paid first, so a GPU that is momentarily too busy would otherwise fail the whole model load — and a caller that drops the model it already had is worse off than one that keeps running on the CPU.
- **A CUDA session is verified with one warm-up inference** on a zero tensor of the model's input geometry, because provider options and the session itself are built successfully by an installation that only fails once a node runs. A failed warm-up reloads the model CPU-only, which costs one inference per model rather than one per image. Geometry the caller could not resolve reports `ErrGeometryUnknown`, which skips verification and keeps the session: an unverifiable model is not evidence against the GPU.
- **TF32 is switched off** in the CUDA provider options. The runtime enables it by default on Ampere and later, and the reduced mantissa moves an embedding by roughly `3e-4` per dimension, against `5e-7` for ordinary kernel-order differences. Embeddings are persisted and compared by distance, so both providers must compute the same graph in FP32; the measured cost of disabling it is a few percent of GPU inference time. `cudaProviderOptions` is separate from the code that applies it so the guard can be asserted without a device.

### Consumers

- `internal/ai/face` — `EmbeddingModels` describes each embedding model and `Detectors` each detector. What stays in that package is what differs per task: alignment mode, embedding length, distance thresholds, and the detector's decode strategy, strides, and anchor count.
