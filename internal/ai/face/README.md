## Face Detection and Embedding Guidelines

**Last Updated:** October 2, 2025

### Overview

This document captures the current state of PhotoPrism's face detection and embedding pipeline following the October 2025 optimizations. It should be used as the canonical reference when assessing detection quality, tuning configuration, or integrating downstream tooling that depends on FaceNet embeddings.

Key changes:

- Multi-angle scanning is enabled by default and can be tuned via configuration.
- Detection thresholds were relaxed to improve recall, while overlap handling was adjusted to preserve historical behaviour.
- All face embeddings are now L2-normalized at creation, midpoint calculation, and deserialization time to keep cosine and Euclidean comparisons consistent.
- Benchmarks were added to track the cost of hotspot routines (`Embedding.Dist` and `EmbeddingsMidpoint`).

### Detection Pipeline

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
- Static datasets (`KidsEmbeddings`, `IgnoredEmbeddings`) and random generators now normalize their entries after perturbation.
- `photoprism faces audit --fix` re-normalizes persisted embeddings, rekeys face IDs, and re-links markers (ID + `FaceDist`) so historical data adopts the canonical unit-length vectors.
- `Faces.Match` pre-filters matchable clusters, keeps an in-memory veto list for freshly cleared markers, and caches embeddings to avoid redundant distance checks; `BenchmarkSelectBestFace` (1024 faces) now reports a bucket size of ~16 candidates out of 1024 (≈98 % fewer distance evaluations) at ≈0.55 ms/op with zero allocations.
- Face clusters update their sample statistics (`Samples`, `SampleRadius`) from the latest matches via `Face.UpdateMatchStats`, avoiding stale radii during optimize loops.
- Cluster materialisation now pre-sizes buffers; `BenchmarkClusterMaterialize` reports ~14.8 µs/op with 64 allocations (≈56 KB) versus the legacy ~29.8 µs/op with 384 allocations (≈105 KB).

This guarantees that Euclidean distance comparisons are equivalent to cosine comparisons, aligning our thresholds with FaceNet literature.

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

| Setting             | Default                      | Description                                                                     |
|---------------------|------------------------------|---------------------------------------------------------------------------------|
| `FACE_ANGLE`        | `-0.3,0,0.3`                 | Detection angles (radians) swept by Pigo.                                       |
| `FACE_SCORE`        | `9.0` (with dynamic offsets) | Base quality threshold before scale adjustments.                                |
| `FACE_OVERLAP`      | `42`                         | Maximum allowed IoU when deduplicating markers.                                 |
| `FACE_KIDS_DIST`    | `0.695`                      | Distance cutoff for kids-face detection (still interpreted in Euclidean space). |
| `FACE_IGNORED_DIST` | `0.86`                       | Distance cutoff for ignoring generic embeddings.                                |

### Benchmark Reference

| Benchmark                     | Before             | After           |
|-------------------------------|--------------------|-----------------|
| `BenchmarkEmbeddingDist`      | ~242 ns/op         | ~155 ns/op      |
| `BenchmarkEmbeddingsMidpoint` | ~194 µs/op, 528 KB | ~99 µs/op, 4 KB |

Re-run these benchmarks after any detector or embedding adjustments to catch regressions early.
