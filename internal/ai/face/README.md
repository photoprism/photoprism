## Face Detection and Embedding Guidelines

**Last Updated:** October 7, 2025

### Overview

This document captures the current state of PhotoPrism's face detection and embedding pipeline following the October 2025 optimizations. It should be used as the canonical reference when assessing detection quality, tuning configuration, or integrating downstream tooling that depends on FaceNet embeddings.

Key changes:

- Multi-angle scanning is enabled by default and can be tuned via configuration.
- Detection thresholds were relaxed to improve recall, while overlap handling was adjusted to preserve historical behaviour.
- All face embeddings are now L2-normalized at creation, midpoint calculation, and deserialization time to keep cosine and Euclidean comparisons consistent.
- Benchmarks were added to track the cost of hotspot routines (`Embedding.Dist` and `EmbeddingsMidpoint`).

> **TODO:** Persist detector provenance in `FaceSrc` (e.g., use `entity.SrcONNX` for SCRFD detections) so hybrid libraries can toggle background filtering per embedding source when upgrading from Pigo.

### Detection Pipeline

PhotoPrism now supports two interchangeable detection engines:

- **Pigo** (default) — CPU-only cascade classifier, retains historical behaviour.
- **ONNX SCRFD 0.5g** — optional ONNX Runtime-backed CNN that delivers higher recall on occluded or off-axis faces. The ONNX engine consumes 720 px thumbnails (model input 640 px), schedules work on the meta/vision workers, and defaults to half the available CPUs (minimum 1 thread). The engine is enabled automatically when `FACE_ENGINE=auto` and the bundled SCRFD model is present (the prebuilt runtime targets glibc ≥ 2.27 on x86_64/arm64). Operators can switch at runtime via `photoprism --face-engine=<auto|pigo|onnx>` or `photoprism faces reset --engine=<auto|pigo|onnx>` for a full re-index.

Runtime selection lives in `Config.FaceEngine()`; `auto` resolves to ONNX when the SCRFD assets are available, otherwise Pigo. `Config.FaceEngineRunType()` mirrors the vision run-type semantics: ONNX defaults to asynchronous `on-demand` mode when only a single inference thread is configured so the indexer remains responsive.

#### Angle Sweep

- The detector now evaluates the Pigo cascade at **-0.3, 0, and +0.3 radians**. These angles are exposed via the new `FACE_ANGLE` option.
- Configuration entry points:
  - CLI flag: `--face-angle=<rad>` (repeatable).
  - Environment variable: `FACE_ANGLE` (comma-separated list).
  - Options API: `Config.FaceAngles()`.
- At start-up the detector receives a copy of `face.DetectionAngles`, so runtime overrides do not mutate the global defaults.

#### Quality & Overlap Thresholds

- The dynamic quality curve in `face.QualityThreshold` was flattened for better small-face recall:
  - +12 for scales <26, +8 for <32, +6 for <40, +4 for <50, +2 for <80, +1 for <110.
- The face overlap floor remains **42 %** to preserve legacy marker behaviour (`OverlapThresholdFloor = 41`). Tests rely on that value (e.g., `Markers.Contains/SameFace`).

#### Landmark Handling

- Landmarks are only evaluated when both eyes are successfully detected for a given face. Eye candidates and cascades respect the configurable perturbation budget.
- The primary detection angles (`FACE_ANGLE`) do **not** affect landmark estimation, which continues to run at 0° to match the cascade assumptions.

### Embedding Handling

#### Normalization

All embeddings, regardless of origin, are normalized to unit length (‖x‖₂ = 1):

- `NewEmbedding` normalizes the raw float32 inference output.
- `EmbeddingsMidpoint` normalizes each contributor, averages component-wise, and renormalizes the centroid.
- `UnmarshalEmbedding` and `UnmarshalEmbeddings` normalize data when loading from persisted JSON.
- Static datasets (children/background samples) and random generators now normalize their entries after perturbation.
- `photoprism faces audit --fix` re-normalizes persisted embeddings, rekeys face IDs, and re-links markers (ID + `FaceDist`) so historical data adopts the canonical unit-length vectors.
- `Faces.Match` pre-filters matchable clusters, keeps an in-memory veto list for freshly cleared markers, and caches embeddings to avoid redundant distance checks; `BenchmarkSelectBestFace` (1024 faces) now reports a bucket size of ~16 candidates out of 1024 (≈98 % fewer distance evaluations) at ≈0.55 ms/op with zero allocations.
- Face clusters update their sample statistics (`Samples`, `SampleRadius`) from the latest matches via `Face.UpdateMatchStats`, avoiding stale radii during optimize loops.
- Cluster materialisation now pre-sizes buffers; `BenchmarkClusterMaterialize` reports ~14.8 µs/op with 64 allocations (≈56 KB) versus the legacy ~29.8 µs/op with 384 allocations (≈105 KB).

This guarantees that Euclidean distance comparisons are equivalent to cosine comparisons, aligning our thresholds with FaceNet literature.

### Manual Cluster Merging & Retained Markers

The `Faces.Optimize` loop still prefers the operator-curated clusters (`face_src = 'manual'`). When multiple manual clusters for the same subject can be merged, `query.MergeFaces` materialises a midpoint cluster and reassigns markers to it. If some markers remain attached to the original clusters (for example because their embeddings sit far from the midpoint), the old clusters cannot be purged and the optimiser now emits a **warning**:

```
faces: retained manual clusters after merge: kept 4 candidate cluster(s) [...] for subject <uid> because markers still reference them
```

This is informational—the optimiser skips that merge and progresses. To reduce noise, consider:

- Running `photoprism faces reset --engine=<pigo|onnx>` to regenerate markers with consistent embeddings.
- Reviewing the subject’s manual clusters in the UI and trimming outliers or reassigning photos to other people.
- Confirming that the remaining clusters genuinely represent different appearances (lighting, age); in that case it is safe to ignore the warning.

No automatic data cleanup runs in this scenario, so operators remain in control of manual edits.

#### Midpoint Computation

- The midpoint routine now performs a single pass (with vector normalization) and uses an inlined L2 distance when computing the sample radius.
- Benchmarked at ~99 µs/op and 4 KB/op for 128 vectors @512 dims, down from ~194 µs/op and >500 KB/op.

#### Distance Function

- `Embedding.Dist` was hand-optimized with loop unrolling (4-way accumulation) and now runs at ~155 ns/op, down from ~242 ns/op (≈36 % faster).
- Euclidean distance remains the recommended metric; with unit vectors, cosine similarity would yield identical rankings, so no change is required to distance thresholds.

### FaceNet Integration Recommendations

- Ensure FaceNet inference remains disabled only when explicitly configured (`PHOTOPRISM_FACENET_DISABLED`).
- If you expose similarity scores, convert Euclidean distance to cosine using: `cos θ = 1 - (d² / 2)` (since embeddings are normalized).
- Keep distance thresholds (e.g., merge, clustering) expressed in the Euclidean domain unless downstream tooling mandates cosine values. The current merge tests expect distances around **0.040** for identical subjects.
- When updating pretrained models or embedding datasets, re-run the dedicated benchmarks and fixture-based tests:
  - `BenchmarkEmbeddingDist`
  - `BenchmarkEmbeddingsMidpoint`
  - `TestMergeFaces/SameSubjects`
  - `TestNet`

### Configuration Summary

| Setting                  | Default                      | Description                                                                                     |
|--------------------------|------------------------------|-------------------------------------------------------------------------------------------------|
| `FACE_ENGINE`            | `auto`                       | Detection engine (`auto`, `pigo`, `onnx`). `auto` resolves to ONNX when the SCRFD model exists. |
| `FACE_ENGINE_RUN`        | `auto`                       | Run schedule (`auto`, `on-demand`, `on-index`, ...). `auto` is stored as an empty string.       |
| `FACE_ENGINE_THREADS`    | `runtime.NumCPU()/2` (≥1)    | ONNX inference threads; ignored by Pigo.                                                        |
| `FACE_ANGLE`             | `-0.3,0,0.3`                 | Detection angles (radians) swept by Pigo.                                                       |
| `FACE_SCORE`             | `9.0` (with dynamic offsets) | Base quality threshold before scale adjustments.                                                |
| `FACE_OVERLAP`           | `42`                         | Maximum allowed IoU when deduplicating markers.                                                 |

### Benchmark Reference

| Benchmark                     | Before             | After           |
|-------------------------------|--------------------|-----------------|
| `BenchmarkEmbeddingDist`      | ~242 ns/op         | ~155 ns/op      |
| `BenchmarkEmbeddingsMidpoint` | ~194 µs/op, 528 KB | ~99 µs/op, 4 KB |

Re-run these benchmarks after any detector or embedding adjustments to catch regressions early.
