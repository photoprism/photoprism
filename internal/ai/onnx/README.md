## ONNX Model Description

**Last Updated:** September 2, 2026

### Overview

This package holds one description of an ONNX model's parameters, shared by the subsystems that run one. The field list is the description of a graph plus its preprocessing contract, and that does not vary by task, so labels, face detection, and face embeddings use the same structure rather than separate runtime registries. NSFW detection can adopt it unchanged when it moves from TensorFlow.

`ModelInfo` contains an `Input` describing geometry, layout, channel order, normalization, and the resize convention; an `Output` describing name, width, output count, and whether values are logits; plus artifact-level fields for file, immutable publisher source, checksum, license, and quantization. The provenance source is distinct from the operational download URL: the former identifies the weights from which an artifact was exported, while the install registry in `scripts/dist/download-models.sh` owns mirrors and fallbacks. The logits flag is optional so an explicit `false` for probability output is distinct from an omitted value that needs a documented default. Resize descriptions include the interpolation algorithm because two models with the same input size can require different short edges and filters.

### What Can Be Inferred and What Cannot

`Inspect` reads the structural parameters from the graph: tensor names, input width and height, axis order, whether a dimension is dynamic, output count, and output width. These are safe to fill silently because getting them wrong raises.

Channel order, normalization, and the resize convention are **not present in a graph**, and `Inspect` never fills them. That split is the reason the package exists: a wrong input shape errors immediately, while a wrong channel order or normalization produces a model that loads, runs, and returns plausible output that is quietly worse. A description without them falls back to a documented default that is applied explicitly and logged, never guessed.

`Inspect` and `Metadata` both create a temporary session internally, so inspection costs a model load. Call them when a model is about to be used; do not scan a directory of artifacts with them.

### A Known Model Means a Known Artifact

`VerifyChecksum` confirms a description against the file it was written for. Lookup selects a candidate by name, and names collide across publishers:

- InsightFace's `antelopev2` pack and fal's AuraFace both ship a file called `glintr100.onnx`, and they are different models. Our mirror renames ours to `auraface_v1_glintr100.onnx` for this reason.
- Published figures routinely describe a differently named export than the file a mirror actually serves, so a name match is not an artifact match.

A name-only match applies one model's preprocessing to another model's weights, and because the mismatched fields are exactly the ones that cannot be inferred, it fails quietly. Descriptions without a recorded checksum are accepted so that custom models supplied through `PHOTOPRISM_MODELS_PATH` keep working.

How a mismatch is handled is the consumer's decision:

| Consumer                             | On mismatch      | Why                                                                                                             |
|:-------------------------------------|:-----------------|:----------------------------------------------------------------------------------------------------------------|
| Face embeddings (`internal/ai/face`) | Refuse to load   | Vectors are persisted, and there is no safe fallback for alignment and normalization                            |
| Face detector (`internal/ai/face`)   | Warn and proceed | Costs recall on the next run rather than a library of incomparable vectors, and its layout comes from the graph |
| Labels (`internal/ai/classify`)       | Refuse to load   | A mismatched preprocessing contract silently changes every generated label                                      |

`VerifyGraph` is the cross-check that follows: a recorded value disagreeing with the graph means one of the two describes a different model, so it aborts rather than reconciling. Dimensions the graph leaves dynamic are not compared.

### Embedded Provenance

`Metadata` reads the `photoprism.` prefixed entries of a model's `metadata_props`. `InfoFromMetadata` converts the supported source, layout, channel-order, normalization, resize, quantization, and input/output fields into `ModelInfo`; `CompleteResizeMetadata` derives a short edge from a crop ratio once graph inspection has supplied the input size. Mean and standard-deviation metadata use the same explicit 0-255 input range as `Normalization`; there is no value-based unit inference. Models we export ourselves should carry their source revision and checksum, channel order, normalization, resize convention, and output semantics there.

Metadata inside the artifact survives mirroring, renaming, and being copied into an image in a way that a sibling `version.txt` does not, which is why a model we export records where it came from rather than relying on the script that installed it.

### Runtime

`EnsureRuntime` loads the ONNX Runtime shared library and initializes the global environment; it must succeed before any model is inspected or loaded. `SharedLibraryCandidates` lists the paths it tries, starting with an explicitly configured one.

The `github.com/yalue/onnxruntime_go` binding requests the exact C API version of the headers it vendors, so it fails to initialize against an older shared library. Bumping that module therefore requires a matching `ONNX_DEFAULT_VERSION` and checksum update in `scripts/dist/install-onnx.sh`, plus a rebuild of the base images that ship `libonnxruntime.so`.

### Consumers

- `internal/ai/face` — `EmbeddingModels` describes each embedding model and `Detectors` each detector. What stays in that package is what differs per task: alignment mode, embedding length, distance thresholds, and the detector's decode strategy, strides, and anchor count.
- `internal/ai/classify` — `Models` adds the label filename and canonical-class-order flag, while the shared description controls session validation and preprocessing.
