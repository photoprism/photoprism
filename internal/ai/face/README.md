## Face Detection & Embedding Guidelines

**Last Updated:** August 23, 2026

### Overview

This document is the canonical reference for PhotoPrism's face detection and embedding pipeline: use it when assessing detection quality, tuning configuration, or integrating downstream tooling that depends on face embeddings.

Detection thresholds favor recall, and overlap handling keeps markers stable across re-detection. Every face embedding is L2-normalized where it is produced and where a centroid is computed, so cosine and Euclidean comparisons stay equivalent; see § Normalization for the one read path that does not, and the repair for it.

Embedding provenance is persisted: `faces.embed_model` and `markers.embed_model` record the model that produced each vector, `entity.Face.Match` refuses to compare clusters from a different model, and `photoprism faces audit` reports the cluster and marker counts per model.

### Detection Pipeline

PhotoPrism uses a single detector:

- **ONNX SCRFD 0.5g** — ONNX Runtime-backed CNN that delivers higher recall on occluded or off-axis faces. The detector consumes 720 px thumbnails (model input 640 px), schedules work on the meta/vision workers, and defaults to the available CPUs divided by the number of indexing workers (minimum 1 thread), because detection takes no lock and one session runs per worker. Operators can select `FACE_ENGINE=onnx` explicitly or leave `FACE_ENGINE=auto`, which resolves to ONNX when the bundled [SCRFD model](https://yakhyo.github.io/facial-analysis/) is present and otherwise disables detection rather than picking another engine.

The `github.com/yalue/onnxruntime_go` binding requests the exact C API version of the headers it vendors, so it fails to initialize against an older shared library. Bumping that module therefore requires a matching `ONNX_DEFAULT_VERSION` and checksum update in `scripts/dist/install-onnx.sh`, plus a rebuild of the base images that ship `libonnxruntime.so`. Tests that load the shared library — `TestNet` for the detector, and the ONNX embedder tests through `onnx.EnsureRuntime` — fail when the model is present but the runtime cannot be initialized, and skip only when the model itself is missing. A version mismatch must not pass as a skipped test.

Runtime selection lives in `Config.FaceEngine()`. Scheduling is controlled by the face model entry in `vision.yml`: `Config.FaceEngineRunType()` simply forwards to `vision.Config.RunType(ModelTypeFace)` and returns `never` when no detector is configured. This keeps face detection aligned with embedding generation so both always run together.

The detector also returns five facial landmarks, which `engine_onnx.go` decodes into `Face.Eyes` (both eyes) and `Face.Landmarks` (nose and mouth corners).

### Embedding Models

`FACE_MODEL` selects the model that turns a face crop into a vector, independently of the detector. Supported models are registered in `models.go`, so the CLI help and config report are generated from one source. Each entry carries the embedding length, alignment mode, and distance thresholds; its artifact and preprocessing contract — file, checksum, license, input geometry, channel order, normalization, resize convention, and output width — live in the `onnx.ModelInfo` that every subsystem running an ONNX model shares (see `internal/ai/onnx/README.md`).

| Model         | Runtime    | Dim | Input   | Alignment | Weights | License       | Installed By                               |
|:--------------|:-----------|----:|:--------|:----------|--------:|:--------------|:-------------------------------------------|
| `facenet`     | TensorFlow | 512 | 160×160 | box crop  |   92 MB | unknown       | `make dep-models`                          |
| `sface`       | ONNX       | 128 | 112×112 | ArcFace-5 |   39 MB | Apache-2.0    | `make dep-models`                          |
| `auraface`    | ONNX       | 512 | 112×112 | ArcFace-5 |  261 MB | Apache-2.0    | `scripts/dist/download-models.sh auraface` |
| `arcface_r50` | ONNX       | 512 | 112×112 | ArcFace-5 |  174 MB | research-only | `scripts/dist/download-arcface.sh`         |
| `arcface_mbf` | ONNX       | 512 | 112×112 | ArcFace-5 |   14 MB | research-only | `scripts/dist/download-arcface.sh`         |

**The setting is normative: it states which model to use rather than being re-derived on every start.** With nothing configured it reads `detect`, and the first start that finds vectors in the library resolves it once and writes the answer to `options.yml`. Only an answer the library actually gave is recorded: a database that could not be read and one that holds no vectors both leave the setting alone, because writing a default down would outlive the moment it was true - a library restored afterwards would then be refused by a model nothing in it was produced with. Detection asks the library before it consults `face.AutoModelPreference`: it reads the recorded provenance of the stored face vectors and keeps whichever model produced most of them, because resolving to a different one would leave those clusters incomparable with everything indexed afterwards; only a library with no vectors follows the preference list, which starts with `sface`. Upgrading an existing installation therefore never changes its embedding space on its own, and once the name is written the answer can no longer move — the library's own answer is a majority vote that flips while a migration is only half done.

**Not recording a model is not the same as having none.** An empty library still resolves to the head of the preference list for that run, so a fresh install detects, embeds, clusters and names faces on SFace exactly as it would with the name written down - only the write is deferred, to the first start that finds vectors and can therefore let the library answer for itself. `photoprism config` reports `detect` until then, because it does not connect and will not guess. The states that really do leave an instance without a model say so instead: `FACE_MODEL=none`, weights that are not installed, and gated weights without acceptance each warn and disable embeddings.

Writing the file is not allowed to fail the start: a read-only `options.yml` produces a warning, the resolved value still applies to that process, and the next start resolves again. A **migration** is the exception - it changed the vectors, so a setting that could not follow them is a failed run with a non-zero exit that names the file to edit.

`options.yml` is loaded after the environment and the command line, so once the file names a model, `PHOTOPRISM_FACE_MODEL` no longer changes it. That is the safe outcome rather than only the consistent one: the embedding model is not a setting an existing instance changes by editing its environment, because changing it means regenerating every vector. An ignored value is reported once per start at info level. The variable keeps its real job — choosing the model for a new instance instead of detecting one.

The provenance count only sees markers that recorded a model. A library whose vectors all predate the column records none, so it resolves to `facenet`, which is the only model that could have produced them. A library that holds both, however, resolves to the recorded model even when the unrecorded vectors outnumber it — trying another model on a few photos and then unsetting `FACE_MODEL` therefore adopts that model for the whole library, and the legacy vectors are what then pauses embedding work until they are migrated.

**Every instance that shares a database must use the same face model.** Detection runs per instance: it reads the recorded provenance from the shared database, but checks `MODELS_PATH` on the local filesystem, so two instances with different models installed can resolve differently for the same library. Each then writes its own `options.yml`, which makes a divergence stable rather than fixing it. Install the same models everywhere, or set `FACE_MODEL` explicitly; either satisfies the rule. Instances sharing one database with different face model configurations are **not a supported deployment**.

The first start after an upgrade is answered the same way: the schema is migrated *after* the configuration is propagated, so the `markers` table has no `embed_model` column to read yet, and the face markers that hold a vector stand in for it. `facenet` is what gets written for such a library, which is correct for it. Commands that never connect to a database resolve nothing and report `detect`, so `photoprism config` stays usable when the database is what is broken.

**Configuring a model twice loads it once.** `ConfigureEmbedder` and `ConfigureEngine` keep the active session when the settings they are handed are the ones it was built from, because loading the same weights again from the same path produces an identical session while reading the file, verifying its checksum, and creating an inference session. An instance configures once, but a test binary builds many configurations. A model that failed to load is never kept, so an attempt made before its weights were installed is retried rather than remembered, and an embedder or engine installed through `UseEmbedder` / `UseEngine` is replaced rather than kept, because the recorded settings do not describe it. Callers that need a fresh session pass different settings or install `nil` first.

A configured model whose weights are missing disables embeddings with a warning rather than falling forward: another model's vectors would otherwise be produced and recorded under the name that was requested.

**A library the configured model cannot read pauses embedding work.** When stored vectors were produced by a model that is not comparable with the configured one, generation, clustering and matching stop after one warning naming both sides and the way out. Filtering the incomparable rows instead - which is what the model-aware queries do within a run - lets indexing keep writing a second vector space beside the first, and that is the state with no cheap way back. A model that could not be loaded at all - missing weights, or terms that were not accepted - pauses the same way while the library holds vectors, because clustering them under another model's distances would rewrite the library at thresholds it was never calibrated for. `FACE_MODEL=none` is not a mismatch: nothing is generated, and the vectors a library already holds stay comparable with each other. Detection keeps running and its markers are still written, because a marker without a vector is filled in on a later pass, so the faces stay recorded rather than having to be re-indexed. `photoprism faces audit` reports the counts, `photoprism faces migrate` resolves them and clears the block in the same run, and the migration itself is exempt by construction: it loads its own target embedder rather than the instance's.

**License-gated weights are refused at use, not only at download.** The InsightFace ArcFace weights are published for non-commercial research only and are therefore never bundled. Their installer ships in `scripts/dist/`, so a user of a published image can run it; it requires `INSIGHTFACE_ACCEPT_LICENSE=1` and verifies a pinned checksum. The application applies the same gate again when a model is selected: a gated model named by `FACE_MODEL` leaves embeddings unconfigured unless the acceptance variable is set **and** the edition is eligible, gated names are left out of the `--face-model` help text, and detection never resolves to one on its own. A library whose vectors were produced by gated weights still resolves to them, because a model that cannot read those vectors is no substitute — the gate then reports why nothing is being embedded, and `photoprism faces migrate` is the way out.

`make dep-models` installs every model a development build runs or ships, and `make dep` includes it, so `make all install` copies SFace into the published images — a model new libraries default to has to be there. The Go test targets depend on the same target, so the ONNX embedder tests cannot silently skip because only a subset of the models was installed.

AuraFace is installed by no target at all. Its Apache-2.0 weights could be redistributed, but a 261 MB graph in every published image is not worth it, so it stays an explicit `scripts/dist/download-models.sh auraface` download. `assets/.buildignore` excludes `models/auraface` and `models/arcface`, so a developer copy is never picked up by `make install` — which also keeps the research-only ArcFace weights out of any build. The file is deliberately renamed from the upstream `glintr100.onnx`: InsightFace's antelopev2 pack ships a different model under that name, and because channel order and normalization cannot be read from an ONNX graph, a name collision would apply one model's preprocessing to the other's weights silently.

That collision is why every ONNX entry records the artifact's SHA256 and why the embedder refuses to load a file whose checksum does not match. A name match with a different artifact has no safe fallback here: the fields that would differ are the ones a graph cannot supply, so the wrong preprocessing would be applied and every vector written under the requested model's name. The detector only warns on the same mismatch, because a different detector costs recall on the next indexing run rather than a library of vectors that cannot be compared with anything. The checksums are also verified on install — from the shared registry in `scripts/dist/download-models.sh` for the bundled models, and from `scripts/dist/download-arcface.sh` for the license-gated ones — and `TestEmbeddingModelChecksums` fails if a copy drifts from the registry.

Models marked `ArcFace-5` need landmark-aligned input. `align.go` fits a similarity transform from the detected landmarks onto the standard 112×112 template that both OpenCV and InsightFace use, and falls back to an unaligned bounding box crop when a face has no complete landmark set.

The transform reads the smallest cached rendition that can still fill the template (`crop.ImageFromIdealThumb`), which is usually larger than the 720 px thumbnail the detector measured the face in. Warping from the detection thumbnail would upscale a small face onto the template and blur exactly the detail the model relies on. An embedding is therefore not a pure function of the original and the model: it also depends on which renditions the thumbnail ladder holds, so raising `--thumb-size` after indexing changes the vectors that a later re-index produces.

**`photoprism faces migrate --to <model>` is how the model is changed.** Vectors from two models are not comparable even when their lengths match, so the migration re-embeds every marker, replaces every cluster, and then records the target in `options.yml` — the setting follows the data, because a later start that read the previous model would hide the whole library from matching. Run it during a maintenance window; the target does not have to match what is configured:

```bash
photoprism faces migrate --to sface --dry-run
photoprism faces migrate --to sface --yes
```

**Stop the server first.** The migration replaces every face cluster in one transaction, and the worker guards it checks are process-local: a CLI run cannot see the indexing and matching a running instance performs on the same rows, and two concurrent migrations cannot see each other. Running it against a live instance can leave markers pointing at clusters that no longer exist, which `photoprism faces audit --fix` repairs but nothing runs automatically. The dry run is read-only and safe at any time; it reports the markers whose file the index already knows it cannot offer for re-embedding, and warns when the originals path is empty or unreadable, which is what an unmounted volume looks like from the outside.

The migration preserves every subject assignment, whether a person set it or the matcher did, and seeds each replacement cluster from the assignments that agree with their own midpoint and stay within the widest distance that cluster can accept: which photos show a person is library knowledge rather than something the old vector space encoded, and a cluster rebuilt from the hand-named markers alone would be too narrow to accept the faces it already held. Assignments further from that midpoint than the target model's cluster distance are left out of the seed so a single mismatched face cannot widen the replacement, and a group that disagrees with itself keeps all of its samples rather than guessing which ones are the outliers. Clusters a person had hidden are hidden again once the replacements exist, because their markers are the only record that the decision was made. It checkpoints regenerated marker embeddings so it can resume after interruption, and atomically replaces face clusters before rebuilding automatic matches. Box-crop models reuse marker geometry and cached thumbnails, falling back to the original; landmark-aligned models redetect each affected thumbnail once so legacy landmarks cannot be mistaken for the required five-point layout. Markers that cannot be regenerated have their stale embeddings cleared and cause a nonzero exit status, which still counts as a completed run: the clusters were replaced, so the setting follows them. Within a run, model-aware queries and `entity.Face.Match` exclude incompatible vectors. Clustering and matching thresholds follow the target model, so no manual retuning is required after a switch.

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
| `sface`       |          0.78 |            0.60 |        0.35 |    n/a |    n/a |
| `auraface`    |          0.98 |            0.76 |        0.35 | 0.9308 | 0.14 % |
| `arcface_r50` |          1.07 |            0.67 |        0.55 | 0.9943 | 0.14 % |
| `arcface_mbf` |          1.03 |            0.64 |        0.49 | 0.9648 | 0.14 % |

FaceNet is the odd row because it keeps what it ships rather than a calibrated point, so it sits at the 1.43 % baseline that defines the budget. The remaining ONNX models are an order of magnitude stricter *and* more accurate on that benchmark, because a wrong automatic merge costs a user more than a match they have to make by hand.

**SFace is the second odd row, and its TAR/FAR are marked n/a deliberately.** Its values no longer come from that benchmark run: the tenth-budget point it produced (0.91 / 0.67 / 0.39, accept distance 1.06) was measured against hand-named lookalike siblings and admitted roughly a quarter of cross-sibling comparisons. The accept distance sat where the false accept rate climbs steeply - between 0.97 and 1.06 it rises fivefold for nine hundredths of a distance - so recall was being bought at a price the error budget was never meant to cover. SFace now sits at the budget-matched point (accept distance 0.95). Re-running `TestCalibrateFaceThresholds` on a broad dataset is what should fill this row back in; the sibling set is a hard-case check, not a calibration set.

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

- `ScoreThreshold` (`FACE_SCORE`, default 9.0) is the base minimum detector score, and `ClusterScoreThreshold` (`FACE_CLUSTER_SCORE`, default 20) is the higher bar a face must clear to contribute to automatic clustering. Both live in `internal/ai/face/config.go`.
- Two detections count as the same face when their area overlap exceeds `OverlapThresholdFloor` (41 %), which is `OverlapThreshold` (42 %) relaxed by one point to absorb rounding. Tests rely on that value (e.g., `Markers.Contains/SameFace`).

### Embedding Handling

#### Memory Management

FaceNet embeddings are generated through TensorFlow bindings that allocate tensors in C memory. Those allocations are released by Go GC finalizers, so long-running indexing jobs can show steadily rising RSS even when the Go heap stays small. To keep memory bounded during extended face indexing runs, PhotoPrism triggers periodic garbage collection and returns freed C-allocated tensor buffers to the OS. You can tune this behavior with `PHOTOPRISM_TF_GC_EVERY` (default **200**; set to `0` to disable). Lower values reduce peak RSS but increase GC overhead and can slow indexing, so keep the default unless memory pressure is severe.

#### Normalization

All embeddings, regardless of origin, are normalized to unit length (‖x‖₂ = 1):

- `NewEmbedding` normalizes the raw float32 inference output.
- `EmbeddingsMidpoint` normalizes each contributor, averages component-wise, and renormalizes the centroid.
- `UnmarshalEmbeddings` normalizes when loading from persisted JSON, which is the path `query.Embeddings` uses to feed clustering. **`entity.Face.Embedding()` and `entity.Marker.Embeddings()` do not**: they call `json.Unmarshal` directly, so the matching path compares a stored vector as it was written. Everything written since normalization existed is already unit length, and `EmbeddingsMidpoint` normalizes its own inputs, so this is reachable only by a row predating that contract.
- Random generators normalize their entries after perturbation.
- `photoprism faces audit --fix` re-normalizes persisted embeddings, rekeys face IDs, and re-links markers (ID + `FaceDist`) so historical data adopts the canonical unit-length vectors. It is the repair for the read path above, and it reads the stored JSON as written in order to detect what needs repairing.
- `Faces.Match` pre-filters matchable clusters, keeps an in-memory veto list for freshly cleared markers, and decodes each cluster embedding and its thresholds once per run rather than once per marker.
- `selectBestFace` compares the marker against every candidate and returns the closest one that accepts it. Each comparison is bounded by the tighter of what that candidate accepts and what the current best already achieves, and `Embedding.DistWithin` abandons the sum once it passes that bound, so a candidate that cannot win costs a fraction of a full compare. The answer is the one an exhaustive scan returns, which `TestSelectBestFaceReturnsClosest` pins against an unbounded reference.
- Face clusters update their sample statistics (`Samples`, `SampleRadius`) from the latest matches via `Face.UpdateMatchStats`, which may **widen** the recorded extent but never narrow it: a run visits only the markers that were unmatched when it started, so narrowing from that subset would drop the accept distance to whatever the newest marker happened to be and refuse the members beyond it. `Face.SetEmbeddings` recomputes both from actual membership, which is the path that may shrink a cluster. `ClampSampleRadius` bounds the radius at `ClusterRadius` wherever it is written, and `AcceptDist` applies the same bound where it is read, so a recalibrated threshold reaches rows written under the previous one. With the shipped FaceNet values an automatic match therefore accepts embeddings up to **0.82** from the centroid.
- **Two ceilings, one above the other.** `AcceptDistMax` caps the cutoff at **1.4** wherever it is read, including for a radius stored under an earlier calibration. `ConfigDistMax` (**1.25**) is the highest value an operator may configure, and `Config.faceAcceptThresholds` refuses a cluster radius and match distance whose sum reaches past it, so the range that can be set is the range the models are calibrated in - the widest of them reaches 1.22. A value above is refused and the calibrated one used, rather than accepted, reported, and then clipped where it is read. Embeddings are unit vectors, so two independent ones average √2 ≈ 1.41 - an average, not a floor, so a noticeable share of unrelated pairs already falls below 1.4, which is why nothing near it is configurable.
- Cluster materialization pre-sizes buffers; `BenchmarkClusterMaterialize` reports ~14.8 µs/op with 64 allocations (≈56 KB).

This guarantees that Euclidean distance comparisons are equivalent to cosine comparisons, aligning our thresholds with [FaceNet](https://maucher.pages.mi.hdm-stuttgart.de/orbook/face/faceRecognition.html) literature.

#### Face Kind Reference

| Kind            | Value | Source                                     | Matching Behavior                               | Notes                                                                                                           |
|:----------------|:-----:|:-------------------------------------------|:------------------------------------------------|:----------------------------------------------------------------------------------------------------------------|
| `RegularFace`   |   1   | Default embedding classification           | Eligible for matching and clustering            | Every cluster starts here.                                                                                      |
| *(reserved)*    |  2–3  | —                                          | —                                               | Held by the retired children and background classifications; never reused, because `faces.face_kind` is stored. |
| `AmbiguousFace` |   4   | `entity.Face.ResolveCollision()` heuristic | Excluded from matching and manual merge retries | Assigned when two subjects collide at very low distance (< 0.02); face remains until collision cleared.         |

### Manual Cluster Merging & Retained Markers

The `Faces.Optimize` loop still prefers the operator-curated clusters (`face_src = 'manual'`). When multiple manual clusters for the same subject can be merged, `query.MergeFaces` materializes a midpoint cluster and reassigns markers to it. If some markers remain attached to the original clusters (for example because their embeddings sit far from the midpoint), the old clusters cannot be purged and the optimizer emits a **warning**:

```
faces: retained manual clusters after merge: kept 4 candidate cluster(s) [...] for subject <uid> because markers still reference them
```

This is informational—the optimizer skips that merge and progresses. To reduce noise, consider:

- Running `photoprism faces reset --engine=<auto|onnx>` to regenerate markers with consistent embeddings.
- Reviewing the subject’s manual clusters in the UI and trimming outliers or reassigning photos to other people.
- Confirming that the remaining clusters genuinely represent different appearances (lighting, age); in that case it is safe to ignore the warning.

No automatic data cleanup runs in this scenario, so operators remain in control of manual edits.

Additional safeguards limit how often stubborn clusters are retried:

- Every manual cluster stores a retry counter (`faces.merge_retry`) and optional note (`merge_notes`). The optimizer skips clusters once the retry count reaches `MergeMaxRetry` (default **1**). The limit may be raised or disabled with the environment variable `PHOTOPRISM_FACE_MERGE_MAX_RETRY` (`0` = unlimited retries).
- Warnings surface only when the retry counter is incremented. Subsequent optimize runs log at debug level until counters are reset.
- `photoprism faces optimize --retry` clears retry counters before running the optimizer, allowing administrators to reprocess clusters after manual cleanup.
- `photoprism faces audit --subject=<uid>` focuses the audit report on a specific person and prints retry counts, sample statistics, and outstanding clusters so operators know which photos still need attention.
- The warning text includes the retry count and cluster IDs.

#### Midpoint Computation

- The midpoint routine performs a single pass (with vector normalization) and uses an inlined L2 distance when computing the sample radius.
- Benchmarked at ~99 µs/op and 4 KB/op for 128 vectors @512 dims.

#### Distance Function

- `Embedding.Dist` is hand-optimized with loop unrolling (4-way accumulation) and runs at ~155 ns/op.
- `Embedding.DistWithin` is the matching hot path. It accumulates the same squared distance but abandons as soon as the running sum passes the caller's limit, testing every 16th component so the branch stays off the critical path of the vectors that survive. It returns -1 rather than a distance above the limit, and both functions return -1 for a non-finite component: a NaN distance compares below every threshold it is fed to, so it would be accepted as a match and could never be displaced by a real one. `Embeddings.DistWithin` tightens the limit to each hit, so it reports the same minimum `Dist` would whenever that minimum is within the limit.
- **How much the bound saves depends on the model.** For a random unit pair at 512 dims the mean abandon depth is ~34 % of the vector at FaceNet's 0.82 accept distance, ~63 % at AuraFace's 1.11 and ~75 % at ArcFace-R50's 1.22, and close to the whole vector as the limit approaches the √2 ≈ 1.41 that two independent unit vectors average, where almost nothing can be abandoned. The saving degrades smoothly rather than falling off a cliff: measured over 2048 candidates that match nothing, with SFace's 128-dimension vectors, selection costs ~95 µs at its 0.95 accept distance, ~137 µs at the `ConfigDistMax` limit of 1.25, and ~140 µs at 1.4. The configurable limit therefore keeps a gate inside the calibrated range; it is not what makes matching fast. Cost is measured indirectly, by `BenchmarkSelectBestFace` and `BenchmarkSelectBestFaceUnmatched` in `internal/photoprism`.
- Euclidean distance remains the recommended metric; with unit vectors, cosine similarity would yield identical rankings, so no change is required to distance thresholds.

### Fixture Vectors

`RandomEmbedding` and `RandomEmbeddings` produce vectors of whatever width the configured model expects, which is what makes them usable as fixtures at all. `FixtureEmbedding` and `FixtureEmbeddingAt` add the two properties a seeded library needs: the same seed always yields the same vector, and a vector can be placed at a chosen distance from another one. Independent seeds are near-orthogonal, so they stand for different people; a chosen distance expressed against a cluster's accept distance keeps a marker inside or outside that cluster whatever the model. `entity.GenerateFaceFixtureVectors` builds the face and marker fixtures from both, immediately before they are written.

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
- `make dep-models` (or `scripts/dist/download-models.sh facenet`)
- Re-run `go test ./internal/ai/face -run TestNet -count=1`

### Configuration Summary

| Setting               | Default                                                                          | Description                                                                                                                                                                              |
|:----------------------|:---------------------------------------------------------------------------------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `FACE_ENGINE`         | `auto`                                                                           | Detection engine (`auto`, `onnx`). `auto` resolves to ONNX when the SCRFD model exists.                                                                                                  |
| `FACE_ENGINE_THREADS` | detection `runtime.NumCPU()/IndexWorkers()`, embedding `runtime.NumCPU()/2` (≥1) | ONNX inference threads. Detection runs one session per indexing worker, embedding one in total.                                                                                          |
| `FACE_MODEL`          | *(unset)*                                                                        | Embedding model (`detect`, `none`, `facenet`, `sface`, `auraface`). Unset detects it once and writes the name to `options.yml`; the gated InsightFace names are accepted but not listed. |
| `FACE_SCORE`          | `9.0` (with dynamic offsets)                                                     | Base quality threshold before scale adjustments.                                                                                                                                         |
| `FACE_OVERLAP`        | `42`                                                                             | Maximum allowed IoU when deduplicating markers.                                                                                                                                          |

`FACE_MODEL` is authoritative for which model produces embeddings. A `face` entry in `vision.yml` schedules detection and embedding through its `Run` value, but a **custom face model configured there is deprecated**: it is still loaded while no embedding model is active, it logs a warning, and its vectors are recorded under the configured model's name rather than its own. Every supported face model needs code that knows its preprocessing contract, so there is nothing useful to configure per installation the way a caption or label model can be.

Run scheduling is configured through the face model entry in `vision.yml`. Adjust the model’s `Run` value (for example `on-schedule`, `manual`, or `never`) to control when detection and embedding jobs execute—no separate `FACE_ENGINE_RUN` flag is required.
When the model is left on the default `auto` run mode, face detection participates in manual, auto, and on-demand workflows but skips scheduled cron runs so background jobs do not trigger unexpectedly; the same applies to an explicit `on-demand` run mode, which skips cron executions by default. Set `Run` to `on-schedule` explicitly if you want faces processed during scheduled vision passes.

> Additional merge tuning: set `PHOTOPRISM_FACE_MERGE_MAX_RETRY` to control how often manual clusters are retried (default 1, `0` = unlimited). See the optimizer notes above.

### Breaking Changes

Collected here so they can be turned into release notes rather than rediscovered.

- **`--face-skip-children` and `--face-allow-background` are removed**, together with `PHOTOPRISM_FACE_SKIP_CHILDREN` and `PHOTOPRISM_FACE_ALLOW_BACKGROUND`. The environment variables are ignored silently, and both options were `yaml:"-"`, so an `options.yml` carrying them is unaffected. A Compose `command:` line that still passes either **flag** fails to start, because unknown flags are rejected. The flags existed for development rather than for tuning a library.
- **The out-of-distribution background filter is gone**, and it was enabled by default (`IgnoreBackground` defaulted to true) for every existing FaceNet library. Removing it is required rather than optional: it compared each embedding against bundled FaceNet-space reference vectors, so under any model of a different width every face would have been classified as background and matching would have stopped library-wide. The child filter it is paired with was already inert. Measured under FaceNet with both forced on, neither fired.
- **A custom face model in `vision.yml` is deprecated** in favor of `FACE_MODEL` — see § Configuration Summary. It still works and warns.
- **`FACE_MODEL` defaults to unset and is written once it has been detected.** `detect` is the explicit spelling, `auto` is accepted as a synonym, and an unsupported value is reported and applies as if nothing were set, without its detected name being written down. Changing the model afterwards is `photoprism faces migrate --to <model>`, which updates `options.yml` itself; an environment variable no longer changes the model of a library that already has one.
- **A library the configured model cannot read pauses embedding work instead of filtering it.** Generation, clustering and matching stop after one warning until `photoprism faces migrate` reconciles them. Detection keeps running, so the faces stay recorded and their vectors are filled in afterwards.
- **The InsightFace models are absent from the `--face-model` help and refused unless their terms are accepted.** Their installers moved to `scripts/dist/`, so they ship in published images; the weights are still never bundled.

### Benchmark Reference

| Benchmark                     | Current         |
|:------------------------------|:----------------|
| `BenchmarkEmbeddingDist`      | ~155 ns/op      |
| `BenchmarkEmbeddingsMidpoint` | ~99 µs/op, 4 KB |

Re-run these benchmarks after any detector or embedding adjustments to catch regressions early.
