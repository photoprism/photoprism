## Face Detection & Embedding Guidelines

**Last Updated:** August 24, 2026

### Overview

This document is the canonical reference for PhotoPrism's face detection and embedding pipeline: use it when assessing detection quality, tuning configuration, or integrating downstream tooling that depends on face embeddings.

Detection thresholds favor recall, and overlap handling keeps markers stable across re-detection. Every face embedding is L2-normalized where it is produced and where a centroid is computed, so cosine and Euclidean comparisons stay equivalent; see § Normalization for the one read path that does not, and the repair for it.

Embedding provenance is persisted: `faces.embed_model` and `markers.embed_model` record the model that produced each vector, `entity.Face.Match` refuses to compare clusters from a different model, and `photoprism faces audit` reports the cluster and marker counts per model.

Detector provenance is persisted alongside it: `markers.detect_model` records the detector whose landmarks produced the crop a vector was computed from, and `photoprism faces audit` reports the marker counts per detector beside the per-model ones. A blank value means the row was written before the column existed; a non-blank one attests the vector's crop rather than the stored landmarks. See § Detector Provenance.

### Detection Pipeline

Detection runs on ONNX Runtime. The detectors this build can run are registered in `detectors.go`, which carries each one's artifact, preprocessing contract and decode strategy:

| Detector | Artifact                            | Weights  | Installed By                     |
|:---------|:------------------------------------|:---------|:---------------------------------|
| `yunet`  | `face_detection_yunet_2026may.onnx` | MIT      | `make dep-models`                |
| `scrfd`  | `det_500m.onnx`                     | non-free | `scripts/dist/download-scrfd.sh` |

`FACE_DETECTOR` selects which one runs. Left unset it reads `auto` - derived again on every start, where `FACE_MODEL`'s `detect` resolves once and is written back - and the detector is **derived from the configured embedding model** — `EmbeddingModel.Detector` states the pairing, so adding a model states its detector. With no model configured, derivation falls back to the first installed detector whose weights may be redistributed, so **YuNet is what a build runs**. A detector that was named but cannot run disables detection rather than falling forward to another: a different detector places different landmarks and therefore produces a different crop.

**SCRFD is never bundled, mirrored, or redistributed.** Its weights are separately licensed, so `download-scrfd.sh` refuses without `INSIGHTFACE_ACCEPT_LICENSE=1`, prints the required notice, and fetches the publisher's own `det_500m.onnx` from an official InsightFace release at a pinned checksum. The registry names that artifact; `Detector.Legacy` lists the names an earlier install wrote, and `Detector.InstalledPath` falls back to them so a copy an operator already holds keeps working.

The detector consumes 720 px thumbnails (model input 640 px), schedules work on the meta/vision workers, and defaults to the available CPUs divided by the number of indexing workers (minimum 1 thread), because detection takes no lock and one session runs per worker.

**Preprocessing is per detector and cannot be crossed.** YuNet consumes raw 0-255 BGR, SCRFD normalized RGB, and neither is derivable from an ONNX graph. `DetectorForFile` keys it off the artifact the path names, because applying the other detector's channel order does not fail — it quietly detects worse.

The channel orders were verified against OpenCV's own source rather than inferred, because the inference goes the wrong way for two of them. `FaceDetectorYN` builds its blob with every `blobFromImage` default, so BGR; `FaceRecognizerSF` sets `swapRB`, so SFace takes **RGB** despite being fed an image that is BGR in memory. Every embedding model on the aligned crop takes RGB.

**Neither detector is rotation invariant.** Both are trained on upright faces, so recall falls away past roughly +/-30 degrees of in-plane roll and is effectively zero near 90 - a person lying down, or a camera held sideways where the orientation tag does not describe the subject. The face is not small or poor, it produces no candidate at all, so no size or score setting recovers it. `specs/intelligence/onnx-face-detection.md` records what recovering these would cost.

The `github.com/yalue/onnxruntime_go` binding requests the exact C API version of the headers it vendors, so it fails to initialize against an older shared library. Bumping that module therefore requires a matching `ONNX_DEFAULT_VERSION` and checksum update in `scripts/dist/install-onnx.sh`, plus a rebuild of the base images that ship `libonnxruntime.so`. Tests that load the shared library — `TestNet` for the detector, and the ONNX embedder tests through `onnx.EnsureRuntime` — fail when the model is present but the runtime cannot be initialized, and skip only when the model itself is missing. A version mismatch must not pass as a skipped test.

Detector selection lives in `Config.FaceDetector()`, and `Config.FaceEngine()` reports the runtime that follows from it. Scheduling is controlled by the face model entry in `vision.yml`: `Config.FaceEngineRunType()` simply forwards to `vision.Config.RunType(ModelTypeFace)` and returns `never` when no detector is configured. This keeps face detection aligned with embedding generation so both always run together.

The detector also returns five facial landmarks, which `engine_onnx.go` decodes into `Face.Eyes` (both eyes) and `Face.Landmarks` (nose and mouth corners).

#### Detector Provenance

`face.Detect` stamps `Face.DetectModel` with the name of the **detector** that found each face - `yunet` or `scrfd`, not the engine. The engine is the runtime, every detector runs on ONNX, so the engine name would distinguish nothing; an engine that cannot name its detector falls back to the engine name, which is coarser provenance but still rules out the legacy Go detector and therefore its unalignable landmark vocabulary, and that value travels with the vector to `markers.detect_model` wherever one is written — `entity.Marker.SetEmbeddings`, the in-place upgrade in `entity.File.AddFace`, and `query.SaveFaceMigrationEmbeddings`, which takes it from the detection `Faces.detectMigrationEmbeddings` ran. Recording it at detection time is what keeps it truthful: reading it back later would ask global configuration, which by then names whatever engine is loaded — `none`, after a reconfiguration — rather than whatever produced the row.

The column attests **the crop a vector was computed from, and the landmarks beside it**. Every path that writes one writes the other from the same detection: `entity.File.AddFace`, the in-place upgrade next to it, and `query.SaveFaceMigrationEmbeddings`, which takes both from the detection `Faces.detectMigrationEmbeddings` ran. A migration that re-crops from stored geometry runs no detector and writes neither, which is correct - what is recorded still describes that crop. A known-good `detect_model` therefore answers "are these landmarks current" as well, which is what makes reusing them instead of re-detecting sound.

A blank value means the row predates the column. Clearing a vector clears both provenance columns, since a detector recorded next to no embedding would claim a crop that no longer exists. A migration that re-crops from stored geometry runs no detector and leaves the recorded one alone.

### Embedding Models

`FACE_MODEL` selects the model that turns a face crop into a vector, independently of the detector. Supported models are registered in `models.go`, so the CLI help and config report are generated from one source. Each entry carries the embedding length, alignment mode, and distance thresholds; its artifact and preprocessing contract — file, checksum, license, input geometry, channel order, normalization, resize convention, and output width — live in the `onnx.ModelInfo` that every subsystem running an ONNX model shares (see `internal/ai/onnx/README.md`).

| Model         | Runtime    | Dim | Input   | Alignment | Weights | License    | Installed By                               |
|:--------------|:-----------|----:|:--------|:----------|--------:|:-----------|:-------------------------------------------|
| `facenet`     | TensorFlow | 512 | 160×160 | box crop  |   92 MB | unknown    | `make dep-models`                          |
| `sface`       | ONNX       | 128 | 112×112 | ArcFace-5 |   39 MB | Apache-2.0 | `make dep-models`                          |
| `auraface`    | ONNX       | 512 | 112×112 | ArcFace-5 |  261 MB | Apache-2.0 | `scripts/dist/download-models.sh auraface` |
| `arcface_r50` | ONNX       | 512 | 112×112 | ArcFace-5 |  174 MB | non-free   | `scripts/dist/download-arcface.sh`         |
| `arcface_mbf` | ONNX       | 512 | 112×112 | ArcFace-5 |   14 MB | non-free   | `scripts/dist/download-arcface.sh`         |

**The setting is normative: it states which model to use rather than being re-derived on every start.** With nothing configured it reads `auto`, and the first start that finds vectors in the library resolves it once and writes the answer to `options.yml`. Only an answer the library actually gave is recorded: a database that could not be read and one that holds no vectors both leave the setting alone, because writing a default down would outlive the moment it was true - a library restored afterwards would then be refused by a model nothing in it was produced with. Detection asks the library before it consults `face.AutoModelPreference`: it reads the recorded provenance of the stored face vectors and keeps whichever model produced most of them, because resolving to a different one would leave those clusters incomparable with everything indexed afterwards; only a library with no vectors follows the preference list, which starts with `sface`. Upgrading an existing installation therefore never changes its embedding space on its own, and once the name is written the answer can no longer move — the library's own answer is a majority vote that flips while a migration is only half done.

**Not recording a model is not the same as having none.** An empty library still resolves to the head of the preference list for that run, so a fresh install detects, embeds, clusters and names faces on SFace exactly as it would with the name written down - only the write is deferred, to the first start that finds vectors and can therefore let the library answer for itself. `photoprism config` reports `detect` until then, because it does not connect and will not guess. The states that really do leave an instance without a model say so instead: `FACE_MODEL=none`, weights that are not installed, and gated weights without acceptance each warn and disable embeddings.

Writing the file is not allowed to fail the start: a read-only `options.yml` produces a warning, the resolved value still applies to that process, and the next start resolves again. A **migration** is the exception - it changed the vectors, so a setting that could not follow them is a failed run with a non-zero exit that names the file to edit.

`options.yml` is loaded after the environment and the command line, so once the file names a model, `PHOTOPRISM_FACE_MODEL` no longer changes it. That is the safe outcome rather than only the consistent one: the embedding model is not a setting an existing instance changes by editing its environment, because changing it means regenerating every vector. An ignored value is reported once per start at info level. The variable keeps its real job — choosing the model for a new instance instead of detecting one.

The provenance count only sees markers that recorded a model. A library whose vectors all predate the column records none, so it resolves to `facenet`, which is the only model that could have produced them. A library that holds both, however, resolves to the recorded model even when the unrecorded vectors outnumber it — trying another model on a few photos and then unsetting `FACE_MODEL` therefore adopts that model for the whole library, and the legacy vectors are what then pauses embedding work until they are migrated.

**Every instance that shares a database must use the same face model.** Detection runs per instance: it reads the recorded provenance from the shared database, but checks `MODELS_PATH` on the local filesystem, so two instances with different models installed can resolve differently for the same library. Each then writes its own `options.yml`, which makes a divergence stable rather than fixing it. Install the same models everywhere, or set `FACE_MODEL` explicitly; either satisfies the rule. Instances sharing one database with different face model configurations are **not a supported deployment**.

The first start after an upgrade is answered the same way: the schema is migrated *after* the configuration is propagated, so the `markers` table has no `embed_model` column to read yet, and the face markers that hold a vector stand in for it. `facenet` is what gets written for such a library, which is correct for it. Commands that never connect to a database resolve nothing and report `detect`, so `photoprism config` stays usable when the database is what is broken.

**Configuring a model twice loads it once.** `ConfigureEmbedder` and `ConfigureEngine` keep the active session when the settings they are handed are the ones it was built from, because loading the same weights again from the same path produces an identical session while reading the file, verifying its checksum, and creating an inference session. An instance configures once, but a test binary builds many configurations. A model that failed to load is never kept, so an attempt made before its weights were installed is retried rather than remembered, and an embedder or engine installed through `UseEmbedder` / `UseEngine` is replaced rather than kept, because the recorded settings do not describe it. Callers that need a fresh session pass different settings or install `nil` first.

A configured model whose weights are missing disables embeddings with a warning rather than falling forward: another model's vectors would otherwise be produced and recorded under the name that was requested.

**A library the configured model cannot read pauses embedding work.** When stored vectors were produced by a model that is not comparable with the configured one, generation, clustering and matching stop after one warning naming both sides and the way out. Filtering the incomparable rows instead - which is what the model-aware queries do within a run - lets indexing keep writing a second vector space beside the first, and that is the state with no cheap way back. A model that could not be loaded at all - missing weights, or terms that were not accepted - pauses the same way while the library holds vectors, because clustering them under another model's distances would rewrite the library at thresholds it was never calibrated for. `FACE_MODEL=none` is not a mismatch: nothing is generated, and the vectors a library already holds stay comparable with each other. Detection keeps running and its markers are still written, because a marker without a vector is filled in on a later pass, so the faces stay recorded rather than having to be re-indexed. `photoprism faces audit` reports the counts, `photoprism faces migrate` resolves them and clears the block in the same run, and the migration itself is exempt by construction: it loads its own target embedder rather than the instance's.

**License-gated weights are refused at use, not only at download.** The InsightFace ArcFace weights are not published under an OSI-approved license and are therefore never bundled. Their installer ships in `scripts/dist/`, so a user of a published image can run it; it requires `INSIGHTFACE_ACCEPT_LICENSE=1` and verifies a pinned checksum. The application applies the same gate again when a model is selected: a gated model named by `FACE_MODEL` leaves embeddings unconfigured unless the acceptance variable is set **and** the edition is eligible, gated names are left out of the `--face-model` help text, and detection never resolves to one on its own. A library whose vectors were produced by gated weights still resolves to them, because a model that cannot read those vectors is no substitute — the gate then reports why nothing is being embedded, and `photoprism faces migrate` is the way out.

`make dep-models` installs every model a development build runs or ships, and `make dep` includes it, so `make all install` copies SFace into the published images — a model new libraries default to has to be there. The Go test targets depend on the same target, so the ONNX embedder tests cannot silently skip because only a subset of the models was installed.

AuraFace is installed by no target at all. Its Apache-2.0 weights could be redistributed, but a 261 MB graph in every published image is not worth it, so it stays an explicit `scripts/dist/download-models.sh auraface` download. `assets/.buildignore` excludes `models/auraface` and `models/arcface`, so a developer copy is never picked up by `make install` — which also keeps the non-free ArcFace weights out of any build. The file is deliberately renamed from the upstream `glintr100.onnx`: InsightFace's antelopev2 pack ships a different model under that name, and because channel order and normalization cannot be read from an ONNX graph, a name collision would apply one model's preprocessing to the other's weights silently.

That collision is why every ONNX entry records the artifact's SHA256 and why the embedder refuses to load a file whose checksum does not match. A name match with a different artifact has no safe fallback here: the fields that would differ are the ones a graph cannot supply, so the wrong preprocessing would be applied and every vector written under the requested model's name. The detector only warns on the same mismatch, because a different detector costs recall on the next indexing run rather than a library of vectors that cannot be compared with anything. The checksums are also verified on install — from the shared registry in `scripts/dist/download-models.sh` for the bundled models, and from `scripts/dist/download-arcface.sh` for the license-gated ones — and `TestEmbeddingModelChecksums` fails if a copy drifts from the registry.

Models marked `ArcFace-5` need landmark-aligned input. `align.go` fits a similarity transform from the detected landmarks onto the standard 112×112 template that both OpenCV and InsightFace use, and falls back to an unaligned bounding box crop when a face has no complete landmark set.

The transform reads the smallest cached rendition that can still fill the template (`crop.ImageFromIdealThumb`), which is usually larger than the 720 px thumbnail the detector measured the face in. Warping from the detection thumbnail would upscale a small face onto the template and blur exactly the detail the model relies on. An embedding is therefore not a pure function of the original and the model: it also depends on which renditions the thumbnail ladder holds, so raising `--thumb-size` after indexing changes the vectors that a later re-index produces.

**`photoprism faces migrate` is how the model is changed.** Vectors from two models are not comparable even when their lengths match, so the migration re-embeds every marker, replaces every cluster, and then records the target in `options.yml` — the setting follows the data, because a later start that read the previous model would hide the whole library from matching. It defaults to `face.DefaultModelName()`, the one model the product offers, so the ordinary case names no target; any other model has to be given to `--to`, and an instance with `FACE_MODEL=none` keeps that decision rather than being migrated onto one. Run it during a maintenance window:

```bash
photoprism faces migrate --dry-run
photoprism faces migrate --yes
```

**It detects at lower floors than indexing does**, and the reason is the direction of the mistake: a false positive costs an index a thumbnail to reject, while a missed detection costs a migration a curated marker's vector. Detection runs at `face.MinSizeThreshold` (10 px) rather than `FACE_SIZE`, because a marker's size is in the pixels of the thumbnail it was detected in and an earlier detector fell back to a larger one - so a legacy marker can describe a face well under the ordinary floor, which no score would recover. `FACE_SIZE_RETRY` does not cover that case: it only fires when a picture yields nothing, and one large face beside a small marker is not nothing. The score floor drops to `face.MigrationScoreThreshold` unless `FACE_SCORE` is set, since that is a decision rather than a calibration. Only the detections a stored marker claims are embedded, so the extra ones cost nothing beyond detection - and leaving them in would also hand the whole file the crop rendition its smallest detection needs.

**It is also how a detector change is repaired.** For a landmark-aligned model the crop is an axis of the embedding space, so a marker cropped by a detector other than the one now in force is stale even though its vector is the target's — leaving it puts one library in two crop spaces. `markers.detect_model` is what tells them apart, and a marker recording none counts as stale too, since its crop cannot be shown to be the current detector's. On the first run after that column was added that is every marker, which is why naming the model already in use is not a no-op; the plan reports the count before the prompt. Nothing is risked by it: a marker whose vector already matches the target keeps that vector when detection cannot find the face again, which is the ordinary outcome for a box a person drew by hand. Those are reported apart from failures, because they are neither work done nor a loss.

**A run holds a lock the rest of the instance can see.** The worker guards are process-local, so a CLI run and a server cannot see each other's; `mutex.AcquireFileLock` writes `storage/faces.lock` instead, which the indexing, vision and face workers consult and hold off on. The file carries its own expiry (`mutex.FileLockMaxAge`) and the holder renews it, so a run that is killed releases it rather than wedging the instance until somebody finds the file. A second migration refuses to start while a live lock is held and names the process holding it. **Stopping the server is still the safe way to run it**: the lock covers the background workers, not a person editing people in the app, and running against a live instance can leave markers pointing at clusters that no longer exist - which `photoprism faces audit --fix` repairs but nothing runs automatically. The dry run is read-only and safe at any time; it reports the markers whose file the index already knows it cannot offer for re-embedding, and warns when the originals path is empty or unreadable, which is what an unmounted volume looks like from the outside.

The migration preserves every subject assignment, whether a person set it or the matcher did, and seeds each replacement cluster from the assignments that agree with their own midpoint and stay within the widest distance that cluster can accept: which photos show a person is library knowledge rather than something the old vector space encoded, and a cluster rebuilt from the hand-named markers alone would be too narrow to accept the faces it already held. Assignments further from that midpoint than the target model's cluster distance are left out of the seed so a single mismatched face cannot widen the replacement, and a group that disagrees with itself keeps all of its samples rather than guessing which ones are the outliers. Clusters a person had hidden are hidden again once the replacements exist, because their markers are the only record that the decision was made. It checkpoints regenerated marker embeddings so it can resume after interruption, and atomically replaces face clusters before rebuilding automatic matches. Box-crop models reuse marker geometry and cached thumbnails, falling back to the original; landmark-aligned models redetect each affected thumbnail once so legacy landmarks cannot be mistaken for the required five-point layout. Markers that cannot be regenerated have their stale embeddings cleared and cause a nonzero exit status, which still counts as a completed run: the clusters were replaced, so the setting follows them. The destructive finalize is refused when too many of them **cleared both clustering bars** (`facesMigrateMaxFailureRatio`) - one below them seeds no cluster and joins none, so its vector changes nothing either way, and counting it made a detector that re-finds fewer weak faces than the one before it look like a storage outage. The refusal names which of the two causes it was, because a file that cannot be read and a face the detector declines to find again ask for opposite responses. Within a run, model-aware queries and `entity.Face.Match` exclude incompatible vectors. Clustering and matching thresholds follow the target model, so no manual retuning is required after a switch.

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
| `sface`       |          0.85 |            0.60 |        0.35 |    n/a |    n/a |
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
| `sface`       | 1.219 |           0.061 |     0.012 |
| `auraface`    | 1.531 |           0.077 |     0.015 |
| `arcface_r50` | 1.672 |           0.084 |     0.017 |
| `arcface_mbf` | 1.609 |           0.080 |     0.016 |

Their practical effect is small, but leaving them fixed at the FaceNet values under a scale that is roughly 1.4x wider is the same trap the per-model thresholds exist to close. `FACE_COLLISION_DIST` and `FACE_EPSILON_DIST` override them the same way the three calibrated thresholds are overridden.

**SFace is the exception, and deliberately so.** Its scale stays pinned at the `ClusterDist` these two were calibrated at rather than following it to the 0.85 that ships. A collision floor states how close two vectors can be and still be told apart, which is a property of the space; `ClusterDist` also carries a recall-against-merging choice, and a choice is not a re-measurement.

#### Quality & Overlap Thresholds

- **Scores are detector confidence in percent, 0-100, and neither bar is shared.** `FACE_SCORE` is the base minimum; unset, each detector's own `MinScore` decides. It reaches the detection session itself (`ONNXOptions.ScoreThreshold`, on the 0-1 scale the detectors report internally) as well as the filter `face.Detect` applies afterwards, so a value **below** a detector's calibrated cutoff lowers it rather than being ignored. `FACE_CLUSTER_SCORE` is the higher bar for automatic clustering; unset, `face.ClusterScore` looks it up from **the detector that produced each marker** (`Detector.ClusterMinScore`: YuNet 20, SCRFD 60) rather than from the one in force, and a value set here applies to every marker because it is a choice rather than a calibration a marker was never scored against. `-1` removes it. A library holds markers from more than one detector and nothing recomputes a score, so a shared bar would exclude old markers for a calibration they were never scored against - permanently. Rows written before `markers.detect_model` existed keep `ClusterScoreThresholdDefault`, so an upgrade strands nothing.

  **Confidence is a weak quality signal.** It saturates above a detector's own cutoff, so it separates faces from non-faces rather than good crops from poor ones. `FACE_CLUSTER_SIZE` is what actually gates recognition quality, being the point at which a crop stops being invented by interpolation. `face.ScoreUncertainty` maps the same 0-100 scale onto the uncertainty a people or portrait label carries.

  **Each detector also carries its own cutoff** (`Detector.MinScore`), because they do not score alike: SCRFD emits one calibrated sigmoid and sits at 0.50, while YuNet scores as `sqrt(cls x obj)`. YuNet ships at 0.09 with a clustering bar of 20, the pair the last stable release ran, rather than at the 0.65 and 70 its own corpus measured. That measurement found where YuNet stops accepting flowers and statues, which weighs false positives and not **re-detection** - and re-detection is what decides whether a migration keeps a curated marker or discards it. Reproducing a configuration whose production behavior is known is what makes the preview's data points attributable; both values are open until then. `TestDetectorRecall` pins the measured recall at the calibrated cutoff explicitly, so a threshold decision cannot silently rewrite it. A cutoff copied from another detector is not calibration, which `TestDetectorMinScore` states.
- `SizeThreshold` (`FACE_SIZE`, default 25 px) and `ClusterSizeThreshold` (`FACE_CLUSTER_SIZE`, default 60 px) are the size pair, and they do different jobs: the first decides whether a marker is created at all, the second whether that face may contribute to automatic clustering. A face below 60 px therefore never seeds a person even though it is detected and shown.
- **Both are measured in pixels of the detection thumbnail (`Fit720`), not of the original and not of the crop the embedder receives.** The crop comes from `crop.ImageFromIdealThumb`, which opens the smallest cached rendition wide enough to fill the 112 px template and falls back to the largest one cached, so with the default `THUMB_SIZE` of 1920 it is drawn at 2.67x the resolution the threshold was compared at. A 60 px face arrives as a 150 px crop and is downscaled onto the template; a 25 px face arrives as 63 px and is stretched 1.8x. `FACE_CLUSTER_SIZE` is therefore close to the point at which a face crop stops being upscaled at all, which is what makes it a sensible bar for clustering.
- **This is the mechanism behind the `THUMB_SIZE` warning in [Advanced Settings](https://docs.photoprism.app/user-guide/settings/advanced/#static-and-dynamic-size-limits).** Lowering the static size limit does not change any face threshold, but it lowers the rendition every crop is drawn from, so each face reaches the embedder with fewer real pixels. The size thresholds keep comparing the same numbers while the crops behind them get worse, which is why the effect is easy to miss.
- Two detections count as the same face when their area overlap exceeds `OverlapThresholdFloor` (41 %), which is `OverlapThreshold` (42 %) relaxed by one point to absorb rounding. Tests rely on that value (e.g., `Markers.Contains/SameFace`).

##### `FACE_SIZE` Decides Whether Crowds Are Seen at All

Detection runs on a 720 px thumbnail, so `FACE_SIZE` is compared against a share of 720 px rather than of the original. In a crowd photograph a person's face is often 10-15 px at that size, which is below the default minimum — so the faces are detected and then discarded, and the photo is indexed as containing nobody.

Measured over twelve crowd photographs that yield no face at the default, varying only the minimum size:

|    `FACE_SIZE` | Faces detected |
|---------------:|---------------:|
| 25 *(default)* |              0 |
|             20 |            117 |
|             15 |            879 |
|             12 |           1140 |
|             10 |           1149 |

The count is steepest below 20, which is why the option accepts values down to **10 px** - the bottom of the detectors' trained range, so a smaller setting would ask for faces no model can find. Anything below it is out of range and falls back to the 25 px default rather than being applied.

Lowering it globally is a trade rather than a fix. Photographs that already yield a face gain roughly three more each at 10 px, and those are the people in the background that most libraries would rather leave unmarked; photographs that yield nothing gain around nine. Raising recall on group shots without marking bystanders everywhere else therefore wants a per-photo decision rather than a smaller default.

**That decision is `FACE_SIZE_RETRY`, and the pipeline makes it per picture.** `face.DetectWithRetry` runs the detector once at `FACE_SIZE`, and only when that finds **no face at all** does it try again at `FACE_SIZE_RETRY` (default 10 px). A photograph whose subject was found never reaches the second pass, so bystanders stay unmarked where somebody is already recognized; a crowd, which yields nothing at the ordinary minimum, is searched again. Set it to `-1` to disable the fallback; zero selects the default, and it is never larger than `FACE_SIZE`.

Measured over the development library, the fallback changed the outcome for 19 pictures out of 861 and added 1163 faces, of which 1149 came from twelve crowd photographs that had yielded none. The cost is one extra inference pass on the pictures that would otherwise be indexed as containing nobody.

The migration does not use this call at all: it detects at `face.MinSizeThreshold` unconditionally, which is strictly more permissive, because a marker the fallback created has to find a partner there or be dropped.

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

- Running `photoprism faces reset --detector=<auto|none|yunet>` to regenerate markers with consistent embeddings.
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

| Setting                 | Default                                | Description                                                                                                                                                                                |
|:------------------------|:---------------------------------------|:-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `FACE_RUN`              | `auto`                                 | When detection and recognition run (`auto`, `always`, `on-index`, `newly-indexed`, `on-schedule`, `on-demand`, `manual`, `never`). Replaces the `Run` value in `vision.yml`.               |
| `FACE_DETECTOR`         | *(unset)*                              | Detection model (`auto`, `yunet`, `none`). Unset derives it from the face model; the gated InsightFace name is accepted but not offered.                                                   |
| `FACE_ENGINE`           | `auto`                                 | Detection runtime (`auto`, `onnx`, `none`) *deprecated*. Only `none` still has an effect, and `FACE_DETECTOR` overrides it.                                                                |
| `FACE_DETECTOR_THREADS` | `runtime.NumCPU()/IndexWorkers()` (≥1) | ONNX threads per detection session. Detection takes no lock, so one session runs per indexing worker.                                                                                      |
| `FACE_MODEL_THREADS`    | `runtime.NumCPU()/2` (≥1)              | ONNX threads for embedding, which runs one session in total behind the model lock.                                                                                                         |
| `FACE_ENGINE_THREADS`   | *(unset)*                              | Sets both of the above at once *deprecated*. The two derive different defaults, so a value that is right for one is not right for the other.                                               |
| `FACE_MODEL`            | *(unset)*                              | Embedding model (`auto`, `sface`, `none`). Unset detects it once and writes the name to `options.yml`; `facenet`, `auraface` and the gated InsightFace names are accepted but not offered. |
| `FACE_SCORE`            | `9` *(detector)*                       | Minimum detection quality, on the 0-100 scale. Unset, each detector's own calibrated cutoff decides; a value set here replaces it, in either direction, and `-1` removes it.               |
| `FACE_OVERLAP`          | `42`                                   | Maximum allowed IoU when deduplicating markers.                                                                                                                                            |

**`vision.yml` no longer configures faces.** `FACE_MODEL` is authoritative for which model produces embeddings and `FACE_RUN` for when it runs; a `face` entry in that file decides neither. A **custom face model configured there is deprecated**: it is still loaded while no embedding model is active, it logs a warning, and its vectors are recorded under the configured model's name rather than its own. Every supported face model needs code that knows its preprocessing contract, so there is nothing useful to configure per installation the way a caption or label model can be.

Run scheduling is `FACE_RUN` (`auto`, `always`, `on-index`, `newly-indexed`, `on-schedule`, `on-demand`, `manual`, `never`). Detection and embedding always run together, so one schedule covers both, and `DISABLE_FACES` turns the whole subsystem off. A `Run` value on a face entry in `vision.yml` is read and **ignored**, with one info-level line naming `FACE_RUN`: two ways to set one schedule raise a precedence question an operator cannot answer from the outside.
When `FACE_RUN` is left on `auto`, face detection participates in manual, auto, and on-demand workflows but skips scheduled cron runs so background jobs do not trigger unexpectedly; an explicit `on-demand` behaves the same way. Set `FACE_RUN=on-schedule` if you want faces processed during scheduled vision passes.

> Additional merge tuning: set `PHOTOPRISM_FACE_MERGE_MAX_RETRY` to control how often manual clusters are retried (default 1, `0` = unlimited). See the optimizer notes above.

### Breaking Changes

Collected here so they can be turned into release notes rather than rediscovered.

- **`--face-skip-children` and `--face-allow-background` are removed**, together with `PHOTOPRISM_FACE_SKIP_CHILDREN` and `PHOTOPRISM_FACE_ALLOW_BACKGROUND`. The environment variables are ignored silently, and both options were `yaml:"-"`, so an `options.yml` carrying them is unaffected. A Compose `command:` line that still passes either **flag** fails to start, because unknown flags are rejected. The flags existed for development rather than for tuning a library.
- **The out-of-distribution background filter is gone**, and it was enabled by default (`IgnoreBackground` defaulted to true) for every existing FaceNet library. Removing it is required rather than optional: it compared each embedding against bundled FaceNet-space reference vectors, so under any model of a different width every face would have been classified as background and matching would have stopped library-wide. The child filter it is paired with was already inert. Measured under FaceNet with both forced on, neither fired.
- **A custom face model in `vision.yml` is deprecated** in favor of `FACE_MODEL` — see § Configuration Summary. It still works and warns.
- **`FACE_MODEL` defaults to unset and is written once it has been detected.** `auto` is the spelling the help offers, `detect` is accepted as a synonym, and an unsupported value is reported and applies as if nothing were set, without its detected name being written down. Changing the model afterwards is `photoprism faces migrate`, which updates `options.yml` itself; an environment variable no longer changes the model of a library that already has one. `photoprism reset` clears the pin, because a reset leaves no vectors for it to keep comparable.
- **A library the configured model cannot read pauses embedding work instead of filtering it.** Generation, clustering and matching stop after one warning until `photoprism faces migrate` reconciles them. Detection keeps running, so the faces stay recorded and their vectors are filled in afterwards.
- **The InsightFace models are absent from the `--face-model` help and refused unless their terms are accepted.** Their installers moved to `scripts/dist/`, so they ship in published images; the weights are still never bundled.

### Benchmark Reference

| Benchmark                     | Current         |
|:------------------------------|:----------------|
| `BenchmarkEmbeddingDist`      | ~155 ns/op      |
| `BenchmarkEmbeddingsMidpoint` | ~99 µs/op, 4 KB |

Re-run these benchmarks after any detector or embedding adjustments to catch regressions early.
