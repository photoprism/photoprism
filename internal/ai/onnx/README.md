## ONNX Model Description

**Last Updated:** August 21, 2026

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

### Consumers

- `internal/ai/face` — `EmbeddingModels` describes each embedding model; `DetectorModel` describes the bundled SCRFD detector. What stays in that package is what differs per task: alignment mode, embedding length, distance thresholds, and the detector's decode strategy, strides, and anchor count.
