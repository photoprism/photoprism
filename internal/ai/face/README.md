## Face Detection & Embedding Guidelines

**Last Updated:** August 17, 2026

### Overview

This document captures the current state of PhotoPrism's face detection and embedding pipeline following the October 2025 optimizations. It should be used as the canonical reference when assessing detection quality, tuning configuration, or integrating downstream tooling that depends on FaceNet embeddings.

Key changes:

- Detection thresholds were relaxed to improve recall, while overlap handling was adjusted to preserve historical behaviour.
- All face embeddings are now L2-normalized at creation, midpoint calculation, and deserialization time to keep cosine and Euclidean comparisons consistent.
- Benchmarks were added to track the cost of hotspot routines (`Embedding.Dist` and `EmbeddingsMidpoint`).

Embedding provenance is persisted: `faces.embed_model` and `markers.embed_model` record the model that produced each vector, `entity.Face.Match` refuses to compare clusters from a different model, and `photoprism faces audit` reports the cluster and marker counts per model.

### Detection Pipeline

PhotoPrism now uses a single detector:

- **ONNX SCRFD 0.5g** — ONNX Runtime-backed CNN that delivers higher recall on occluded or off-axis faces. The detector consumes 720 px thumbnails (model input 640 px), schedules work on the meta/vision workers, and defaults to half the available CPUs (minimum 1 thread). Operators can select `FACE_ENGINE=onnx` explicitly or leave `FACE_ENGINE=auto`, which resolves to ONNX when the bundled [SCRFD model](https://yakhyo.github.io/facial-analysis/) is present and otherwise disables detection rather than picking another engine.

The `github.com/yalue/onnxruntime_go` binding requests the exact C API version of the headers it vendors, so it fails to initialize against an older shared library. Bumping that module therefore requires a matching `ONNX_DEFAULT_VERSION` and checksum update in `scripts/dist/install-onnx.sh`, plus a rebuild of the base images that ship `libonnxruntime.so`. `TestNet` is the only test that loads the shared library, so it fails when the SCRFD model is present but the detector cannot be initialized, and skips only when the model itself is missing — a version mismatch must not pass as a skipped test.

Runtime selection lives in `Config.FaceEngine()`. Scheduling is controlled by the face model entry in `vision.yml`: `Config.FaceEngineRunType()` simply forwards to `vision.Config.RunType(ModelTypeFace)` and returns `never` when no detector is configured. This keeps face detection aligned with embedding generation so both always run together.

The detector also returns five facial landmarks, which `engine_onnx.go` decodes into `Face.Eyes` (both eyes) and `Face.Landmarks` (nose and mouth corners).

### Embedding Models

`FACE_MODEL` selects the model that turns a face crop into a vector, independently of the detector. Supported models are registered in `models.go`, so the CLI help and config report are generated from one source. Each entry carries the embedding length, alignment mode, and distance thresholds; its artifact and preprocessing contract — file, source, checksum, license, input geometry, channel order, normalization, resize convention, and output width — live in the `onnx.ModelInfo` that every subsystem running an ONNX model shares (see `internal/ai/onnx/README.md`).

| Model         | Runtime    | Dim | Input   | Alignment | Weights | License       | Installed By                   |
|:--------------|:-----------|----:|:--------|:----------|--------:|:--------------|:-------------------------------|
| `facenet`     | TensorFlow | 512 | 160×160 | box crop  |   92 MB | unknown       | `make dep-tensorflow`          |
| `sface`       | ONNX       | 128 | 112×112 | ArcFace-5 |   39 MB | Apache-2.0    | `make dep-sface`               |
| `auraface`    | ONNX       | 512 | 112×112 | ArcFace-5 |  261 MB | Apache-2.0    | `scripts/download-auraface.sh` |
| `arcface_r50` | ONNX       | 512 | 112×112 | ArcFace-5 |  174 MB | research-only | `scripts/download-arcface.sh`  |
| `arcface_mbf` | ONNX       | 512 | 112×112 | ArcFace-5 |   14 MB | research-only | `scripts/download-arcface.sh`  |

`auto` asks the library before it consults `face.AutoModelPreference`. A library that already holds face vectors keeps the model that produced them, because resolving to a different one would leave every stored cluster incomparable with everything indexed afterwards; only a library with no vectors follows the preference list, which starts with `sface`. Upgrading an existing installation therefore never changes its embedding space on its own — moving to another model is an explicit `FACE_MODEL` change followed by `photoprism faces migrate`.

Two details make that guarantee hold. Vectors written before the provenance column existed report no model at all, and since the schema is migrated *after* the configuration is propagated, the first start after an upgrade reads a `markers` table that has no `embed_model` column yet; both cases are read as FaceNet, because nothing else could have produced them. And a resolution found before the database is connected is not cached, so an early lookup cannot freeze the preference list into place.

An explicitly configured model whose weights are missing falls back with a warning: embeddings would otherwise be produced by the fallback model and recorded under the name that was requested. The InsightFace ArcFace weights are published for non-commercial research only and are therefore never bundled; their install script requires `ARCFACE_ACCEPT_LICENSE=1` and verifies a pinned checksum.

SFace is part of `make dep` through `dep-onnx`, so `make all install` copies it into the published images — a model new libraries default to has to be there. The Go test targets depend on `dep-sface` separately, so the ONNX embedder tests never silently skip when only a subset of the dependencies was installed.

AuraFace is installed by no target at all. Its Apache-2.0 weights could be redistributed, but a 261 MB graph in every published image is not worth it, so it stays an explicit `scripts/download-auraface.sh` download. `assets/.buildignore` excludes `models/auraface` and `models/arcface`, so a developer copy is never picked up by `make install` — which also keeps the research-only ArcFace weights out of any build. The file is deliberately renamed from the upstream `glintr100.onnx`: InsightFace's antelopev2 pack ships a different model under that name, and because channel order and normalization cannot be read from an ONNX graph, a name collision would apply one model's preprocessing to the other's weights silently.

That collision is why every ONNX entry records the artifact's SHA256 and why the embedder refuses to load a file whose checksum does not match. A name match with a different artifact has no safe fallback here: the fields that would differ are the ones a graph cannot supply, so the wrong preprocessing would be applied and every vector written under the requested model's name. The detector only warns on the same mismatch, because a different detector costs recall on the next indexing run rather than a library of vectors that cannot be compared with anything. The checksums are also verified by the install scripts, and `TestEmbeddingModelChecksums` fails if the two copies drift apart.

Models marked `ArcFace-5` need landmark-aligned input. `align.go` fits a similarity transform from the detected landmarks onto the standard 112×112 template that both OpenCV and InsightFace use, and falls back to an unaligned bounding box crop when a face has no complete landmark set.

**Switching models invalidates existing clusters.** Vectors from two models are not comparable even when their lengths match, so change `FACE_MODEL` first and then migrate the library during a maintenance window. The target must match the resolved configured model:

```bash
photoprism faces migrate --to sface --dry-run
photoprism faces migrate --to sface --yes
```

The migration preserves manual subject assignments, checkpoints regenerated marker embeddings so it can resume after interruption, and atomically replaces face clusters before rebuilding automatic matches. Box-crop models reuse marker geometry and cached thumbnails, falling back to the original; landmark-aligned models redetect each affected thumbnail once so legacy landmarks cannot be mistaken for the required five-point layout. Markers that cannot be regenerated have their stale embeddings cleared and cause a nonzero exit status. Until migration is complete, model-aware queries and `entity.Face.Match` exclude incompatible vectors. Clustering and matching thresholds follow the target model, so no manual retuning is required after a switch.

To compare installed models on a labeled dataset of identity subdirectories:

```bash
PHOTOPRISM_TEST_FACE_DATASET=/path/to/dataset \
  go test ./internal/ai/face -run TestBenchmarkEmbeddingModels -count=1 -v -timeout 120m
```

#### Model-Specific Thresholds

Each model entry carries its own `ClusterDist`, `ClusterRadius`, `MatchDist`, `CollisionDist`, and `Epsilon`. Distances are not comparable across models, so one global set of thresholds fits exactly one embedding space: measured on the benchmark datasets, SFace reaches a true accept rate of 0.7934 with the FaceNet-tuned values and 0.9603 with its own, at a tenth of FaceNet's false accept rate. `Config.FaceClusterDist()`, `FaceClusterRadius()`, and `FaceMatchDist()` resolve the configured model's value unless the operator sets `FACE_CLUSTER_DIST`, `FACE_CLUSTER_RADIUS`, or `FACE_MATCH_DIST` explicitly, and `Config.Propagate` publishes the result to the package variables the indexer reads.

`TestCalibrateFaceThresholds` derives these values. It measures what the shipped FaceNet configuration costs on a labeled dataset and then finds, for every other model, the thresholds that meet the same error budget in that model's own scale, so a model switch is never more permissive than the configuration PhotoPrism ships today:

```bash
PHOTOPRISM_TEST_FACE_DATASET=/path/to/dataset \
  go test ./internal/ai/face -run TestCalibrateFaceThresholds -count=1 -v -timeout 180m
```

Cluster centroids are built with `EmbeddingsMidpoint` and scored as `dist - min(radius, cap)`, which is the quantity `Face.Match` compares against `MatchDist`, so a threshold sweep yields the constant directly. Two operating points are reported: one that spends the baseline budget, and one at a tenth of it. The registry uses the stricter point, where every ONNX model still beats FaceNet's current true accept rate — fewer false merges *and* more correct matches. FaceNet keeps its shipped values, because changing them would alter matching for every existing library on upgrade.

**The operating point in absolute terms.** The derivation is relative to what PhotoPrism already ships rather than to a chosen error rate, so the rate the shipped values actually imply is worth stating. On the benchmark datasets, the FaceNet configuration in production costs **1.43 % false automatic matches** at a true accept rate of 0.8318. The registry's thresholds target a tenth of that, **0.14 %**, and the measured true accept rate at that point is:

| Model         | `ClusterDist` | `ClusterRadius` | `MatchDist` |    TAR |    FAR |
|:--------------|--------------:|----------------:|------------:|-------:|-------:|
| `facenet`     |          0.64 |            0.42 |        0.40 | 0.8318 | 1.43 % |
| `sface`       |          0.91 |            0.67 |        0.39 | 0.9603 | 0.14 % |
| `auraface`    |          0.98 |            0.76 |        0.35 | 0.9308 | 0.14 % |
| `arcface_r50` |          1.07 |            0.67 |        0.55 | 0.9943 | 0.14 % |
| `arcface_mbf` |          1.03 |            0.64 |        0.49 | 0.9648 | 0.14 % |

FaceNet is the odd row because it keeps what it ships rather than a calibrated point, so it sits at the 1.43 % baseline that defines the budget. Every ONNX model is an order of magnitude stricter *and* more accurate. Spending the full baseline instead would buy SFace 0.9852 rather than 0.9603; the stricter point is the deliberate choice, because a wrong automatic merge costs a user more than a match they have to make by hand.

Two caveats apply to the recommendations. The measured centroids are always pure because they are built from labeled identities, so a wider `ClusterRadius` is less safe in production, where an impure cluster has a large radius and would be given more slack. And `ClusterDist` is derived from pairwise distance equivalence rather than from a DBSCAN simulation, so cluster fragmentation and merge behavior are not measured. Validate against a real library before treating these values as final.

**`CollisionDist` and `Epsilon` are derived, not measured.** One is the floor below which two vectors count as indistinguishable and the other the slack added to that check, so neither corresponds to an error budget; what they follow is the width of the model's distance scale:

    value = FaceNet value * (model ClusterDist / ClusterDistDefault), rounded to three decimals

| Model         | Scale | `CollisionDist` | `Epsilon` |
|:--------------|------:|----------------:|----------:|
| `facenet`     | 1.000 |           0.050 |     0.010 |
| `sface`       | 1.422 |           0.071 |     0.014 |
| `auraface`    | 1.531 |           0.077 |     0.015 |
| `arcface_r50` | 1.672 |           0.084 |     0.017 |
| `arcface_mbf` | 1.609 |           0.080 |     0.016 |

Their practical effect is small, but leaving them fixed at the FaceNet values under a scale that is roughly 1.4x wider is the same trap the per-model thresholds exist to close. `FACE_COLLISION_DIST` and `FACE_EPSILON_DIST` override them the same way the three calibrated thresholds are overridden.

#### Quality & Overlap Thresholds

- The dynamic quality curve in `face.QualityThreshold` was flattened for better small-face recall:
  - +12 for scales <26, +8 for <32, +6 for <40, +4 for <50, +2 for <80, +1 for <110.
- The face overlap floor remains **42 %** to preserve legacy marker behaviour (`OverlapThresholdFloor = 41`). Tests rely on that value (e.g., `Markers.Contains/SameFace`).

### Embedding Handling

#### Memory Management

FaceNet embeddings are generated through TensorFlow bindings that allocate tensors in C memory. Those allocations are released by Go GC finalizers, so long-running indexing jobs can show steadily rising RSS even when the Go heap stays small. To keep memory bounded during extended face indexing runs, PhotoPrism now triggers periodic garbage collection and returns freed C-allocated tensor buffers to the OS. You can tune this behavior with `PHOTOPRISM_TF_GC_EVERY` (default **200**; set to `0` to disable). Lower values reduce peak RSS but increase GC overhead and can slow indexing, so keep the default unless memory pressure is severe.

#### Normalization

All embeddings, regardless of origin, are normalized to unit length (‖x‖₂ = 1):

- `NewEmbedding` normalizes the raw float32 inference output.
- `EmbeddingsMidpoint` normalizes each contributor, averages component-wise, and renormalizes the centroid.
- `UnmarshalEmbedding` and `UnmarshalEmbeddings` normalize data when loading from persisted JSON.
- Random generators normalize their entries after perturbation.
- `photoprism faces audit --fix` re-normalizes persisted embeddings, rekeys face IDs, and re-links markers (ID + `FaceDist`) so historical data adopts the canonical unit-length vectors.
- `Faces.Match` pre-filters matchable clusters, keeps an in-memory veto list for freshly cleared markers, and caches embeddings to avoid redundant distance checks; `BenchmarkSelectBestFace` (1024 faces) now reports a bucket size of ~16 candidates out of 1024 (≈98 % fewer distance evaluations) at ≈0.55 ms/op with zero allocations.
- Face clusters update their sample statistics (`Samples`, `SampleRadius`) from the latest matches via `Face.UpdateMatchStats`. `ClampSampleRadius` bounds the radius at `ClusterRadius` wherever it is written, and `AcceptDist` applies the same bound where it is read, so a recalibrated threshold reaches rows written under the previous one. With the shipped FaceNet values an automatic match therefore accepts embeddings up to **0.82** from the centroid.
- `AcceptDistMax` caps that cutoff at **1.4** whatever the configuration. Embeddings are unit vectors, so two independent ones sit at √2 ≈ 1.41 and anything at or above that accepts every pair. The bound follows from normalization alone and holds for every model; the widest calibrated model reaches 1.22.
- Cluster materialisation pre-sizes buffers; `BenchmarkClusterMaterialize` reports ~14.8 µs/op with 64 allocations (≈56 KB).

This guarantees that Euclidean distance comparisons are equivalent to cosine comparisons, aligning our thresholds with [FaceNet](https://maucher.pages.mi.hdm-stuttgart.de/orbook/face/faceRecognition.html) literature.

#### Face Kind Reference

| Kind            | Value | Source                                     | Matching Behavior                               | Notes                                                                                                           |
|:----------------|:-----:|:-------------------------------------------|:------------------------------------------------|:----------------------------------------------------------------------------------------------------------------|
| `RegularFace`   |   1   | Default embedding classification           | Eligible for matching and clustering            | Every cluster starts here.                                                                                      |
| *(reserved)*    |  2–3  | —                                          | —                                               | Held by the retired children and background classifications; never reused, because `faces.face_kind` is stored. |
| `AmbiguousFace` |   4   | `entity.Face.ResolveCollision()` heuristic | Excluded from matching and manual merge retries | Assigned when two subjects collide at very low distance (< 0.02); face remains until collision cleared.         |

### Manual Cluster Merging & Retained Markers

The `Faces.Optimize` loop still prefers the operator-curated clusters (`face_src = 'manual'`). When multiple manual clusters for the same subject can be merged, `query.MergeFaces` materialises a midpoint cluster and reassigns markers to it. If some markers remain attached to the original clusters (for example because their embeddings sit far from the midpoint), the old clusters cannot be purged and the optimiser now emits a **warning**:

```
faces: retained manual clusters after merge: kept 4 candidate cluster(s) [...] for subject <uid> because markers still reference them
```

This is informational—the optimiser skips that merge and progresses. To reduce noise, consider:

- Running `photoprism faces reset --engine=<auto|onnx>` to regenerate markers with consistent embeddings.
- Reviewing the subject’s manual clusters in the UI and trimming outliers or reassigning photos to other people.
- Confirming that the remaining clusters genuinely represent different appearances (lighting, age); in that case it is safe to ignore the warning.

No automatic data cleanup runs in this scenario, so operators remain in control of manual edits.

Additional safeguards limit how often stubborn clusters are retried:

- Every manual cluster stores a retry counter (`faces.merge_retry`) and optional note (`merge_notes`). The optimiser skips clusters once the retry count reaches `MergeMaxRetry` (default **1**). The limit may be raised or disabled with the environment variable `PHOTOPRISM_FACE_MERGE_MAX_RETRY` (`0` = unlimited retries).
- Warnings surface only when the retry counter is incremented. Subsequent optimise runs log at debug level until counters are reset.
- `photoprism faces optimize --retry` clears retry counters before running the optimiser, allowing administrators to reprocess clusters after manual cleanup.
- `photoprism faces audit --subject=<uid>` focuses the audit report on a specific person and prints retry counts, sample statistics, and outstanding clusters so operators know which photos still need attention.
- The warning text now includes the retry count and cluster IDs.

#### Midpoint Computation

- The midpoint routine now performs a single pass (with vector normalization) and uses an inlined L2 distance when computing the sample radius.
- Benchmarked at ~99 µs/op and 4 KB/op for 128 vectors @512 dims, down from ~194 µs/op and >500 KB/op.

#### Distance Function

- `Embedding.Dist` was hand-optimized with loop unrolling (4-way accumulation) and now runs at ~155 ns/op, down from ~242 ns/op (≈36 % faster).
- Euclidean distance remains the recommended metric; with unit vectors, cosine similarity would yield identical rankings, so no change is required to distance thresholds.

### FaceNet Integration Recommendations

- Face detection and recognition are disabled as a unit by `PHOTOPRISM_DISABLE_FACES`; `PHOTOPRISM_DISABLE_TENSORFLOW` also stops FaceNet and is deprecated. `FACE_MODEL=none` keeps detection and disables embedding generation.
- If you expose similarity scores, convert Euclidean distance to cosine using: `cos θ = 1 - (d² / 2)` (since embeddings are normalized).
- Keep distance thresholds (e.g., merge, clustering) expressed in the Euclidean domain unless downstream tooling mandates cosine values. The current merge tests expect distances around **0.040** for identical subjects.
- When updating pretrained models or embedding datasets, re-run the dedicated benchmarks and fixture-based tests:
  - `BenchmarkEmbeddingDist`
  - `BenchmarkEmbeddingsMidpoint`
  - `TestMergeFaces/SameSubjects`
  - `TestNet`

### Troubleshooting FaceNet Model Files

If FaceNet unit tests fail with `Read less bytes than requested`, the local model file is typically incomplete or corrupted (`assets/models/facenet/saved_model.pb`).

Recovery steps:

- `rm -f /tmp/photoprism/facenet.zip`
- `rm -rf assets/models/facenet`
- `make dep-tensorflow` (or `scripts/dist/download-models.sh facenet`)
- Re-run `go test ./internal/ai/face -run TestNet -count=1`

### Configuration Summary

| Setting               | Default                      | Description                                                                             |
|:----------------------|:-----------------------------|:----------------------------------------------------------------------------------------|
| `FACE_ENGINE`         | `auto`                       | Detection engine (`auto`, `onnx`). `auto` resolves to ONNX when the SCRFD model exists. |
| `FACE_ENGINE_THREADS` | `runtime.NumCPU()/2` (≥1)    | ONNX inference threads.                                                                 |
| `FACE_MODEL`          | `auto`                       | Embedding model (`auto`, `none`, `facenet`, `sface`, `auraface`, `arcface_r50`, `arcface_mbf`). |
| `FACE_SCORE`          | `9.0` (with dynamic offsets) | Base quality threshold before scale adjustments.                                        |
| `FACE_OVERLAP`        | `42`                         | Maximum allowed IoU when deduplicating markers.                                         |

Run scheduling is configured through the face model entry in `vision.yml`. Adjust the model’s `Run` value (for example `on-schedule`, `manual`, or `never`) to control when detection and embedding jobs execute—no separate `FACE_ENGINE_RUN` flag is required.
When the model is left on the default `auto` run mode, face detection participates in manual, auto, and on-demand workflows but skips scheduled cron runs so background jobs do not trigger unexpectedly; the same applies to an explicit `on-demand` run mode, which now skips cron executions by default. Set `Run` to `on-schedule` explicitly if you want faces processed during scheduled vision passes.

> Additional merge tuning: set `PHOTOPRISM_FACE_MERGE_MAX_RETRY` to control how often manual clusters are retried (default 1, `0` = unlimited). See the optimiser notes above.

### Benchmark Reference

| Benchmark                     | Before             | After           |
|:------------------------------|:-------------------|:----------------|
| `BenchmarkEmbeddingDist`      | ~242 ns/op         | ~155 ns/op      |
| `BenchmarkEmbeddingsMidpoint` | ~194 µs/op, 528 KB | ~99 µs/op, 4 KB |

Re-run these benchmarks after any detector or embedding adjustments to catch regressions early.
