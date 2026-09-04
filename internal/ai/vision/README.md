## PhotoPrism — Vision Package

**Last Updated:** September 2, 2026

### Overview

`internal/ai/vision` provides the shared model registry, request builders, and parsers that power PhotoPrism’s caption, label, face, NSFW, and future generate workflows. It reads `vision.yml`, normalizes models, and dispatches calls to local ONNX/TensorFlow engines or remote services:

- **ONNX Runtime (built-in labels)** — fixed-taxonomy ImageNet classifiers run locally with per-model preprocessing and checksum validation.
- **TensorFlow (transitional)** — built-in NSFW and FaceNet models remain local while their separate ONNX migrations are completed. Long-running TensorFlow inference can accumulate C-allocated tensor memory until GC finalizers run, so PhotoPrism periodically triggers garbage collection; tune with `PHOTOPRISM_TF_GC_EVERY` (default **200**, `0` disables).
- **Ollama** — local or proxied multimodal LLMs. See [`ollama/README.md`](ollama/README.md) for tuning and schema details. The engine defaults to `${OLLAMA_BASE_URL:-http://ollama:11434}/api/generate`, trimming any trailing slash on the base URL; set `OLLAMA_BASE_URL=https://ollama.com` to opt into cloud defaults. The default model is `gemma4:latest` (self-hosted) or `minimax-m3:cloud` (cloud), and reasoning is disabled by default (`Service.Think: "false"`) so thinking-capable models do not leak reasoning into results. That flag is a correctness guard rather than a performance one — a reasoning build still generates the reasoning and bills the tokens for it, so prefer a non-reasoning tag (for example `qwen3-vl:4b-instruct` over `qwen3-vl:4b`) where one exists.
- **OpenAI** — cloud Responses API. See [`openai/README.md`](openai/README.md) for prompts, schema variants, and header requirements.

Faces are the one type this registry does not own. A `face` entry in `vision.yml` schedules nothing - `FACE_RUN` decides when detection and embedding run, and a `Run` value on that entry is read and reported as ignored - and *which* model turns a crop into a vector is settled per instance by `FACE_MODEL`, which is detected once and recorded in `options.yml`. `Model.FaceModel()` returns that embedder before it looks at the `vision.yml` entry, and `nil` when embeddings are off (`FACE_MODEL=none`), when the configured weights are missing or license-refused, or while a library the model cannot read has embedding work paused. `MigrationFaceModel()` is the one caller exempt from the last gate, because `photoprism faces migrate` is what resolves that mismatch.

**A custom face model in `vision.yml` is therefore deprecated.** `FACE_MODEL` is authoritative; a custom entry is still loaded while no embedding model is active, logs a deprecation warning, and has its vectors recorded under the configured model's name rather than its own. Unlike a caption or label model, every face model needs code that knows its preprocessing contract — channel order, normalization, input geometry, alignment mode — so there is nothing useful to point at a different artifact here. The registry, thresholds, and provenance columns live in [`internal/ai/face`](../face/README.md).

### Configuration

#### Models

The `vision.yml` file is usually kept in the `storage/config` directory (override with `PHOTOPRISM_VISION_YAML`). It defines a list of models under `Models:`. Key fields are captured below. If a type is omitted entirely, PhotoPrism will auto-append the built-in defaults (labels, nsfw, face, caption) so you no longer need placeholder stanzas. The `Thresholds` block is optional; missing or out-of-range values fall back to defaults.

| Field                   | Default                                | Notes                                                                              |
|:------------------------|:---------------------------------------|:-----------------------------------------------------------------------------------|
| `Type` (required)       | —                                      | `labels`, `caption`, `face`, `nsfw`, `generate`. Drives routing & scheduling.      |
| `Name`                  | derived from type/version              | Display name; lower-cased by helpers.                                              |
| `Model`                 | `""`                                   | Raw identifier override; precedence: `Service.Model` → `Model` → `Name`.           |
| `Version`               | `latest` (non-OpenAI)                  | OpenAI payloads omit version.                                                      |
| `Engine`                | inferred from service/alias            | Aliases set formats, file scheme, resolution. Explicit `Service` values still win. |
| `Run`                   | `auto`                                 | See Run modes table below; ignored for `Type: face`, which follows `FACE_RUN`.     |
| `Default`               | `false`                                | Select the built-in model for a type.                                              |
| `Disabled`              | `false`                                | Registered but inactive.                                                           |
| `Resolution`            | model-specific / 720 (Ollama/OpenAI)  | Local ONNX geometry comes from its description or graph.                           |
| `System` / `Prompt`     | engine defaults                        | Override prompts per model.                                                        |
| `Format`                | `""`                                   | Response hint (`json`, `text`, `markdown`).                                        |
| `Normalize`             | engine default                         | Label name normalization; see the table below. Labels models only.                 |
| `Schema` / `SchemaFile` | engine defaults / empty                | Inline vs file JSON schema (labels).                                               |
| `TensorFlow`            | nil                                    | Local TF model info (paths, tags).                                                 |
| `ONNX`                  | nil                                    | Shared local ONNX artifact and preprocessing description.                         |
| `LabelFile`             | `labels.txt`                           | Vocabulary paired with a local labels model.                                      |
| `CanonicalOrder`        | `false`                                | Require canonical ImageNet-1k order and reject a background offset.               |
| `Options`               | nil                                    | Sampling/settings merged with engine defaults.                                     |
| `Service`               | nil                                    | Remote endpoint config (see below).                                                |

#### Label Name Normalization

Language models return label names in whatever shape their prompt encourages, so PhotoPrism canonicalizes them before they are stored. `Normalize` selects how:

| Value         | Result for `ferris wheel` | Behavior                                                                                                                                 |
|:--------------|:--------------------------|:-----------------------------------------------------------------------------------------------------------------------------------------|
| *(unset)*     | engine default            | `phrase` for hosted models, `single-word` otherwise.                                                                                     |
| `single-word` | `Ferris`                  | Collapse to the first token that resolves against the label vocabulary, or to the first token.                                           |
| `phrase`      | `Ferris Wheel`            | Keep the phrase, matching it — and its singular form — against the vocabulary as a whole first, so `sea lions` still becomes `Sea Lion`. |
| `false`       | `Ferris Wheel`            | Keep the name the model returned. No vocabulary name mapping at all, so `carousel` stays `Carousel` instead of becoming `Theme Park`.    |

`off`, `none`, `no`, and `disabled` are accepted as aliases of `false`.

Only the name depends on the mode. Confidence and topicality thresholds, categories, and priorities are applied identically in all three, including `false` — a label whose name matches a vocabulary rule still inherits that rule's threshold, so low-value names such as `background` are dropped in every mode. What does change is which rule is found: `ski-lift` inherits the stricter `ski` threshold when it collapses to `Ski`, and the global threshold when it is kept as `Ski Lift`.

The defaults differ because the failure modes do. A model counts as hosted when it carries the `cloud` version tag — which holds even when a local instance proxies the request — when it is one of OpenAI's own identifiers (`gpt-*`, `o1`/`o3`/`o4`), or when its endpoint is the Ollama Cloud host. Every signal is read from the model, so a configuration that reaches both a local instance and a hosted service classifies each entry on its own. An OpenAI-compatible local server such as vLLM, llama.cpp, or LM Studio runs open-weight models under their own names and is treated as self-hosted.

Hosted models only use a compound when the subject has one — across a 16-image benchmark the multi-word labels they returned were `ferris wheel`, `amusement park`, `roller coaster`, and `ski-lift`, every one of which the default mangles. Models small enough to run on an 8 GB GPU mix real compounds with filler such as `city_name` and `photo list`, which is what `single-word` keeps in check.

This matters most outside English, where a compound subject is usually two words. **A name written in a non-Latin script is therefore never collapsed, whatever the mode says.** The vocabulary is English, so splitting `حمار وحشي` (zebra) into tokens has nothing to resolve against and only changes the subject to `حمار` (donkey); the same holds for `גלגל ענק` (ferris wheel) and `גלגל` (wheel). The check is on the script rather than the language, because a Latin-script name can still resolve — Spanish `noria gigante` keeps the head noun `Noria` — and a name mixing scripts keeps normal handling, so `شاطئ beach` still resolves to `Beach` through the vocabulary.

Phrase mode pairs with a system prompt that does not demand single-word nouns — see `LabelSystemSimple` in the Ollama engine. It cannot repair a model that concatenates instead (`ferriswheel` stays `Ferriswheel`).

#### Run Modes

| Value           | When it runs                                                     | Recommended use                                |
|:----------------|:-----------------------------------------------------------------|:-----------------------------------------------|
| `auto`          | Built-in local defaults during index; external via metadata/schedule | Leave as-is for most setups.                 |
| `manual`        | Only when explicitly invoked (CLI/API)                           | Experiments and diagnostics.                   |
| `on-index`      | During indexing + manual                                         | Fast built-in models only.                     |
| `newly-indexed` | Metadata worker after indexing + manual                          | External/Ollama/OpenAI without slowing import. |
| `on-demand`     | Manual, metadata worker, and scheduled jobs                      | Broad coverage without index path.             |
| `on-schedule`   | Scheduled jobs + manual                                          | Nightly/cron-style runs.                       |
| `always`        | Indexing, metadata, scheduled, manual                            | High-priority models; watch resource use.      |
| `never`         | Never executes                                                   | Keep definition without running it.            |

> **Note:** For performance reasons, `on-index` is only supported for built-in local models.

#### Model Options

The model `Options` adjust model parameters such as temperature, top-p, and schema constraints when using [Ollama](ollama/README.md) or [OpenAI](openai/README.md). Rows are ordered exactly as defined in `vision/model_options.go`.

| Option             | Engines        | Default             | Description                                                                             |
|:-------------------|:---------------|:--------------------|:----------------------------------------------------------------------------------------|
| `Temperature`      | Ollama, OpenAI | engine default      | Controls randomness with a value between `0.01` and `2.0`; not used for OpenAI's GPT-5. |
| `TopK`             | Ollama         | engine default      | Limits sampling to the top K tokens to reduce rare or noisy outputs.                    |
| `TopP`             | Ollama, OpenAI | engine default      | Nucleus sampling; keeps the smallest token set whose cumulative probability ≥ `p`.      |
| `MinP`             | Ollama         | engine default      | Drops tokens whose probability mass is below `p`, trimming the long tail.               |
| `TypicalP`         | Ollama         | engine default      | Keeps tokens with typicality under the threshold; combine with TopP/MinP for flow.      |
| `TfsZ`             | Ollama         | engine default      | Tail free sampling parameter; lower values reduce repetition.                           |
| `Seed`             | Ollama         | random per run      | Fix for reproducible outputs; unset for more variety between runs.                      |
| `NumKeep`          | Ollama         | engine default      | How many tokens to keep from the prompt before sampling starts.                         |
| `RepeatLastN`      | Ollama         | engine default      | Number of recent tokens considered for repetition penalties.                            |
| `RepeatPenalty`    | Ollama         | engine default      | Multiplier >1 discourages repeating the same tokens or phrases.                         |
| `PresencePenalty`  | OpenAI         | engine default      | Increases the likelihood of introducing new tokens by penalizing existing ones.         |
| `FrequencyPenalty` | OpenAI         | engine default      | Penalizes tokens in proportion to their frequency so far.                               |
| `PenalizeNewline`  | Ollama         | engine default      | Whether to apply repetition penalties to newline tokens.                                |
| `Stop`             | Ollama, OpenAI | engine default      | Array of stop sequences (e.g., `["\\n\\n"]`).                                           |
| `Mirostat`         | Ollama         | engine default      | Enables Mirostat sampling (`0` off, `1/2` modes).                                       |
| `MirostatTau`      | Ollama         | engine default      | Controls surprise target for Mirostat sampling.                                         |
| `MirostatEta`      | Ollama         | engine default      | Learning rate for Mirostat adaptation.                                                  |
| `NumPredict`       | Ollama         | engine default      | Ollama-specific max output tokens; synonymous intent with `MaxOutputTokens`.            |
| `MaxOutputTokens`  | Ollama, OpenAI | engine default      | Upper bound on generated tokens; adapters raise low values to defaults.                 |
| `ForceJson`        | Ollama, OpenAI | engine default      | Forces structured output when enabled.                                                  |
| `SchemaVersion`    | Ollama, OpenAI | derived from schema | Override when coordinating schema migrations.                                           |
| `CombineOutputs`   | OpenAI         | engine default      | Controls whether multi-output models combine results automatically.                     |
| `Detail`           | OpenAI         | engine default      | Controls OpenAI vision detail level (`low`, `high`, `auto`).                            |
| `NumCtx`           | Ollama, OpenAI | engine default      | Context window length (tokens).                                                         |
| `NumThread`        | Ollama         | runtime auto        | Caps CPU threads for local engines.                                                     |
| `NumBatch`         | Ollama         | engine default      | Batch size for prompt processing.                                                       |
| `NumGpu`           | Ollama         | engine default      | Number of GPUs to distribute work across.                                               |
| `MainGpu`          | Ollama         | engine default      | Primary GPU index when multiple GPUs are present.                                       |
| `LowVram`          | Ollama         | engine default      | Enable VRAM-saving mode; may reduce performance.                                        |
| `VocabOnly`        | Ollama         | engine default      | Load vocabulary only for quick metadata inspection.                                     |
| `UseMmap`          | Ollama         | engine default      | Memory map model weights instead of fully loading them.                                 |
| `UseMlock`         | Ollama         | engine default      | Lock model weights in RAM to reduce paging.                                             |
| `Numa`             | Ollama         | engine default      | Enable NUMA-aware allocations when available.                                           |

#### Model Service

Configures the endpoint URL, method, format, and authentication for [Ollama](ollama/README.md), [OpenAI](openai/README.md), and other engines that perform remote HTTP requests:

| Field                              | Default                                  | Notes                                                                                                                                                                                                                                                                                         |
|:-----------------------------------|:-----------------------------------------|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `Uri`                              | required for remote                      | Endpoint base. Empty keeps a configured ONNX or TensorFlow model local. Ollama alias fills `${OLLAMA_BASE_URL}/api/generate`, defaulting to `http://ollama:11434`.                                                                                                                           |
| `Method`                           | `POST`                                   | Override verb if provider needs it.                                                                                                                                                                                                                                                           |
| `Key`                              | `""`                                     | Bearer token; prefer env expansion (OpenAI: `OPENAI_API_KEY`, Ollama: `OLLAMA_API_KEY`).                                                                                                                                                                                                      |
| `Username` / `Password`            | `""`                                     | Injected as basic auth when URI lacks userinfo.                                                                                                                                                                                                                                               |
| `Model`                            | `""`                                     | Endpoint-specific override; wins over model/name.                                                                                                                                                                                                                                             |
| `Org` / `Project`                  | `""`                                     | OpenAI headers (org/proj IDs).                                                                                                                                                                                                                                                                |
| `Tier`                             | `""`                                     | OpenAI service tier sent as top-level `service_tier` in the request body (e.g. `flex` for cheaper, slower processing). OpenAI-only; supports `${ENV}` expansion. Omitted when empty (OpenAI default `auto`).                                                                                  |
| `Think`                            | `""` (Ollama engine: `"false"`)          | Optional reasoning hint passed as `think` in service requests. The Ollama engine defaults it to `"false"` (reasoning off); re-enable with `"true"`. Supports levels like `low`, `medium`, `high`; string values `true`/`false` are normalized to JSON booleans on output. Omitted when empty. |
| `RequestFormat` / `ResponseFormat` | set by engine alias                      | Explicit values win over alias defaults.                                                                                                                                                                                                                                                      |
| `FileScheme`                       | set by engine alias (`data` or `base64`) | Controls image transport.                                                                                                                                                                                                                                                                     |
| `Disabled`                         | `false`                                  | Disable the endpoint without removing the model.                                                                                                                                                                                                                                              |

> **Authentication:** All credentials and identifiers support `${ENV_VAR}` expansion. `Service.Key` sets `Authorization: Bearer <token>`; `Username`/`Password` injects HTTP basic authentication into the service URI when it is not already present. When `Service.Key` is empty, PhotoPrism defaults to `OPENAI_API_KEY` (OpenAI engine) or `OLLAMA_API_KEY` (Ollama engine), also honoring their `_FILE` counterparts. Key and schema file paths must reference readable regular files (directories are ignored/rejected).

> **Retries:** The shared service client retries transient `HTTP 429` responses (rate limiting, `flex`-tier capacity pressure) with bounded exponential backoff — `ServiceMaxRetries` attempts, `ServiceRetryDelay` base delay, capped at `ServiceRetryMaxDelay` — honoring a `Retry-After` header when present (also capped at `ServiceRetryMaxDelay`, so a provider asking for a longer pause is retried sooner and may fail through to the next worker pass) and keeping the total within `ServiceTimeout`. Other error statuses stay terminal, so the item is only reattempted on the next worker pass.

### Field Behavior & Precedence

- Model identifier resolution order: `Service.Model` → `Model` → `Name`. `Model.GetModel()` returns `(id, name, version)` where Ollama receives `name:version` and other engines receive `name` plus a separate `Version`.
- Env expansion runs for all `Service` credentials and `Model` overrides; empty or disabled models return empty identifiers.
- Options merging: engine defaults fill missing fields; explicit values always win. Temperature is capped at `MaxTemperature`.
- Authentication: `Service.Key` sets `Authorization: Bearer <token>`; `Username`/`Password` inject HTTP basic auth into the service URI when not already present. `Username`, `Password`, and `Key` are never serialized to JSON, and `photoprism vision ls` prints the endpoint with the password redacted, so a shared terminal transcript or report does not carry it.
- Reasoning control: `Service.Think` maps to `ApiRequest.Think` and is serialized only when non-empty (`omitempty`). The Ollama engine defaults it to `"false"` via its engine alias (applied when `Service.Think` is empty), so reasoning is off out of the box; other engines leave it empty. During JSON encoding, `"true"` / `"false"` are converted to boolean `true` / `false`; other non-empty values are sent as strings.
- Label name normalization: `Normalize` resolves as explicit value → `phrase` when `Model.IsCloud()` → `EngineInfo.DefaultNormalize` → `single-word`, at read time rather than at load, so a changed engine default reaches configurations that never set the field. It is applied to the response and never sent to the service. An unrecognized value is reported once when `vision.yml` is loaded and then treated as unset.

### Minimal Examples

#### Built-in Local Defaults

```yaml
Models:
  - Type: labels
    Default: true
    Run: auto

  - Type: nsfw
    Default: true
    Run: auto

  - Type: face
    Default: true
```

#### Ollama Labels

```yaml
Models:
  - Type: labels
    Model: gemma4:latest
    Engine: ollama
    Run: newly-indexed
    Service:
      Uri: ${OLLAMA_BASE_URL}/api/generate
```

To keep compound names such as `ferris wheel` instead of collapsing them, relax the system prompt and switch the normalization together — one without the other has no effect:

```yaml
Models:
  - Type: labels
    Model: gemma4:latest
    Engine: ollama
    Run: newly-indexed
    Normalize: phrase
    System: |
      You are a PhotoPrism vision model. Output concise JSON that matches the schema.
    Service:
      Uri: ${OLLAMA_BASE_URL}/api/generate
```

More Ollama guidance: [`internal/ai/vision/ollama/README.md`](ollama/README.md).

#### OpenAI Captions

```yaml
Models:
  - Type: caption
    Model: gpt-5-mini
    Engine: openai
    Run: newly-indexed
    Service:
      Uri: https://api.openai.com/v1/responses
      Org: ${OPENAI_ORG}
      Project: ${OPENAI_PROJECT}
      Key: ${OPENAI_API_KEY}
```

More OpenAI guidance: [`internal/ai/vision/openai/README.md`](openai/README.md).

#### Custom ONNX Labels

```yaml
Models:
  - Type: labels
    Name: custom_21k
    Engine: onnx
    Path: custom_21k
    LabelFile: labels-imagenet21k.txt
    ONNX:
      File: custom_21k.onnx
      Output:
        Logits: true
```

### Custom ONNX Label Models — What’s Supported

- Scope: Fixed-taxonomy local classification (`labels`). Use Ollama or OpenAI for captions and open-vocabulary labels.
- Location & paths: If `Path` is empty, the model is loaded from `assets/models/<name>` (lowercased, underscores). If `Path` is set, it is still searched under `assets/models`; absolute paths are not supported.
- Expected files: One `.onnx` graph and the exact label file declared by `LabelFile`. The output width must equal the number of labels.
- Preprocessing: Declare geometry, layout, color order, mean/std, resize/crop convention, and interpolation in `ONNX.Input` or embedded `photoprism.*` metadata. `Resolution` remains an explicit override for graphs with dynamic spatial axes.
- Output: One tensor is required. Declare `ONNX.Output.Logits`; omitted output semantics default to raw logits with a warning.
- Sources: Labels produced by local ONNX models are recorded with source `image`; overriding the source isn’t supported yet.
- Config file: `vision.yml` is the conventional name; in the latest version, `.yaml` is also supported by the loader.

### CLI Quick Reference

- List models: `photoprism vision ls` (shows resolved IDs, engines, options, run mode, disabled flag).
- Run a model: `photoprism vision run -m labels --count 5` (use `--force` to bypass `Run` rules).
- Validate config: `photoprism vision ls --json` to confirm env-expanded values without triggering calls.

### When to Choose Each Engine

- **ONNX Runtime**: fast, offline fixed-taxonomy labels and face models with one shared native runtime.
- **TensorFlow**: transitional local FaceNet and NSFW support until their ONNX migrations land.
- **Ollama**: private, GPU/CPU-hosted multimodal LLMs; best for richer captions/labels without cloud traffic.
- **OpenAI**: highest quality reasoning and multimodal support; requires API key and network access.

### NSFW Detection

NSFW is wired through the same model registry as labels, captions, and faces — `Type: nsfw` resolves to the built-in TensorFlow classifier by default, and can be overridden in `vision.yml` to point at an Ollama or OpenAI endpoint.

There is also a fast-path: when `Type: labels` is served by an LLM, PhotoPrism can ask the labels call to include `nsfw` + `nsfw_confidence` in the same response. This is gated by the package-level global `DetectNSFWLabels`, set from `config.go` as `DetectNSFW() && Experimental()` — both `PHOTOPRISM_DETECT_NSFW=true` **and** `PHOTOPRISM_EXPERIMENTAL=true` are required. When either flag is off, the labels prompt stays on `LabelPromptDefault` (no NSFW fields), and `labels.IsNSFW()` cannot trigger.

The runtime guards in `internal/photoprism/index_mediafile.go` and `internal/workers/vision.go` additionally short-circuit any NSFW promotion on `conf.DetectNSFW()`. The dedicated `Type: nsfw` model is filtered out of scheduled runs by `VisionModelShouldRun` whenever `DetectNSFW()` is false.

`DetectNSFW` returns one `nsfw.Result` per image, and a result that no detector decided is `unavailable` rather than safe — including when the batch never ran, when a remote service returns fewer results than images, and when a single local file could not be read. Callers must act on `Status`, never on the class scores alone. `Thresholds.NSFW` is the operating point for both the dedicated model and the labels fast-path; left unset it resolves to `DefaultNSFWThreshold`. See [`internal/ai/nsfw/README.md`](../nsfw/README.md) for the result contract, the full call-graph, and the user-facing matrix at [docs.photoprism.app/user-guide/ai/nsfw/](https://docs.photoprism.app/user-guide/ai/nsfw/).

### Model Unload on Idle

PhotoPrism currently keeps local ONNX and TensorFlow models resident for the lifetime of the process to avoid repeated load costs. A future “model unload on idle” mode would track last-use timestamps and close the session after a configurable idle period. The trade-off is higher latency and CPU overhead on the next request, so it is not enabled today.

### Troubleshooting

- If face model initialization fails with `Read less bytes than requested` (often followed by `invalid face model configuration` in `GenerateFaceEmbeddings` tests), reinstall the local FaceNet assets:
  - `rm -f /tmp/photoprism/facenet.zip`
  - `rm -rf assets/models/facenet`
  - `make dep-models` (or `scripts/dist/download-models.sh facenet`)
  - Re-run: `go test ./internal/ai/face -run TestNet -count=1` and `go test ./internal/ai/vision -run TestGenerateFaceEmbeddings -count=1`

### Related Docs

- Ollama specifics: [`internal/ai/vision/ollama/README.md`](ollama/README.md)
- OpenAI specifics: [`internal/ai/vision/openai/README.md`](openai/README.md)
- REST API reference: https://docs.photoprism.dev/
- Developer guide (Vision): https://docs.photoprism.app/developer-guide/api/
