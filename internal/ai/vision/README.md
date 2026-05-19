## PhotoPrism — Vision Package

**Last Updated:** March 3, 2026

### Overview

`internal/ai/vision` provides the shared model registry, request builders, and parsers that power PhotoPrism’s caption, label, face, NSFW, and future generate workflows. It reads `vision.yml`, normalizes models, and dispatches calls to one of three engines:

- **TensorFlow (built‑in)** — default Nasnet / NSFW / Facenet models, no remote service required. Long-running TensorFlow inference can accumulate C-allocated tensor memory until GC finalizers run, so PhotoPrism periodically triggers garbage collection to return that memory to the OS; tune with `PHOTOPRISM_TF_GC_EVERY` (default **200**, `0` disables). Lower values reduce peak RSS but increase GC overhead and can slow indexing, so keep the default unless memory pressure is severe.
- **Ollama** — local or proxied multimodal LLMs. See [`ollama/README.md`](ollama/README.md) for tuning and schema details. The engine defaults to `${OLLAMA_BASE_URL:-http://ollama:11434}/api/generate`, trimming any trailing slash on the base URL; set `OLLAMA_BASE_URL=https://ollama.com` to opt into cloud defaults.
- **OpenAI** — cloud Responses API. See [`openai/README.md`](openai/README.md) for prompts, schema variants, and header requirements.

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
| `Run`                   | `auto`                                 | See Run modes table below.                                                         |
| `Default`               | `false`                                | Keep one per type for TensorFlow fallbacks.                                        |
| `Disabled`              | `false`                                | Registered but inactive.                                                           |
| `Resolution`            | 224 (TensorFlow) / 720 (Ollama/OpenAI) | Thumbnail edge in px; TensorFlow models default to 224 unless you override.        |
| `System` / `Prompt`     | engine defaults                        | Override prompts per model.                                                        |
| `Format`                | `""`                                   | Response hint (`json`, `text`, `markdown`).                                        |
| `Schema` / `SchemaFile` | engine defaults / empty                | Inline vs file JSON schema (labels).                                               |
| `TensorFlow`            | nil                                    | Local TF model info (paths, tags).                                                 |
| `Options`               | nil                                    | Sampling/settings merged with engine defaults.                                     |
| `Service`               | nil                                    | Remote endpoint config (see below).                                                |

#### Run Modes

| Value           | When it runs                                                     | Recommended use                                |
|:----------------|:-----------------------------------------------------------------|:-----------------------------------------------|
| `auto`          | TensorFlow defaults during index; external via metadata/schedule | Leave as-is for most setups.                   |
| `manual`        | Only when explicitly invoked (CLI/API)                           | Experiments and diagnostics.                   |
| `on-index`      | During indexing + manual                                         | Fast built-in models only.                     |
| `newly-indexed` | Metadata worker after indexing + manual                          | External/Ollama/OpenAI without slowing import. |
| `on-demand`     | Manual, metadata worker, and scheduled jobs                      | Broad coverage without index path.             |
| `on-schedule`   | Scheduled jobs + manual                                          | Nightly/cron-style runs.                       |
| `always`        | Indexing, metadata, scheduled, manual                            | High-priority models; watch resource use.      |
| `never`         | Never executes                                                   | Keep definition without running it.            |

> **Note:** For performance reasons, `on-index` is only supported for the built-in TensorFlow models.

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

| Field                              | Default                                  | Notes                                                                                                                                                                                                    |
|:-----------------------------------|:-----------------------------------------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `Uri`                              | required for remote                      | Endpoint base. Empty keeps model local (TensorFlow). Ollama alias fills `${OLLAMA_BASE_URL}/api/generate`, defaulting to `http://ollama:11434`.                                                          |
| `Method`                           | `POST`                                   | Override verb if provider needs it.                                                                                                                                                                      |
| `Key`                              | `""`                                     | Bearer token; prefer env expansion (OpenAI: `OPENAI_API_KEY`, Ollama: `OLLAMA_API_KEY`).                                                                                                                 |
| `Username` / `Password`            | `""`                                     | Injected as basic auth when URI lacks userinfo.                                                                                                                                                          |
| `Model`                            | `""`                                     | Endpoint-specific override; wins over model/name.                                                                                                                                                        |
| `Org` / `Project`                  | `""`                                     | OpenAI headers (org/proj IDs).                                                                                                                                                                           |
| `Think`                            | `""`                                     | Optional reasoning hint passed as `think` in service requests. Supports levels like `low`, `medium`, `high`; string values `true`/`false` are normalized to JSON booleans on output. Omitted when empty. |
| `RequestFormat` / `ResponseFormat` | set by engine alias                      | Explicit values win over alias defaults.                                                                                                                                                                 |
| `FileScheme`                       | set by engine alias (`data` or `base64`) | Controls image transport.                                                                                                                                                                                |
| `Disabled`                         | `false`                                  | Disable the endpoint without removing the model.                                                                                                                                                         |

> **Authentication:** All credentials and identifiers support `${ENV_VAR}` expansion. `Service.Key` sets `Authorization: Bearer <token>`; `Username`/`Password` injects HTTP basic authentication into the service URI when it is not already present. When `Service.Key` is empty, PhotoPrism defaults to `OPENAI_API_KEY` (OpenAI engine) or `OLLAMA_API_KEY` (Ollama engine), also honoring their `_FILE` counterparts. Key and schema file paths must reference readable regular files (directories are ignored/rejected).

### Field Behavior & Precedence

- Model identifier resolution order: `Service.Model` → `Model` → `Name`. `Model.GetModel()` returns `(id, name, version)` where Ollama receives `name:version` and other engines receive `name` plus a separate `Version`.
- Env expansion runs for all `Service` credentials and `Model` overrides; empty or disabled models return empty identifiers.
- Options merging: engine defaults fill missing fields; explicit values always win. Temperature is capped at `MaxTemperature`.
- Authentication: `Service.Key` sets `Authorization: Bearer <token>`; `Username`/`Password` inject HTTP basic auth into the service URI when not already present.
- Reasoning control: `Service.Think` maps to `ApiRequest.Think` and is serialized only when non-empty (`omitempty`). During JSON encoding, `"true"` / `"false"` are converted to boolean `true` / `false`; other non-empty values are sent as strings.

### Minimal Examples

#### TensorFlow (built‑in defaults)

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
    Run: auto
```

#### Ollama Labels

```yaml
Models:
  - Type: labels
    Model: gemma3:latest
    Engine: ollama
    Run: newly-indexed
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

#### Custom TensorFlow Labels (SavedModel)

```yaml
Models:
  - Type: labels
    Name: transformer
    Engine: tensorflow
    Path: transformer   # resolved under assets/models
    Resolution: 224     # keep standard TF input size unless your model differs
    TensorFlow:
      Output:
        Logits: true    # set true for most TF2 SavedModel classifiers
```

### Custom TensorFlow Models — What’s Supported

- Scope: Classification tasks only (`labels`). TensorFlow models cannot generate captions today; use Ollama or OpenAI for captions.
- Location & paths: If `Path` is empty, the model is loaded from `assets/models/<name>` (lowercased, underscores). If `Path` is set, it is still searched under `assets/models`; absolute paths are not supported.
- Expected files: `saved_model.pb`, a `variables/` directory, and a `labels.txt` alongside the model; use TF2 SavedModel classifiers.
- Resolution: Stays at 224px unless your model requires a different input size; adjust `Resolution` and the `TensorFlow.Input` block if needed.
- Sources: Labels produced by TensorFlow models are recorded with source `image`; overriding the source isn’t supported yet.
- Config file: `vision.yml` is the conventional name; in the latest version, `.yaml` is also supported by the loader.

### CLI Quick Reference

- List models: `photoprism vision ls` (shows resolved IDs, engines, options, run mode, disabled flag).
- Run a model: `photoprism vision run -m labels --count 5` (use `--force` to bypass `Run` rules).
- Validate config: `photoprism vision ls --json` to confirm env-expanded values without triggering calls.

### When to Choose Each Engine

- **TensorFlow**: fast, offline defaults for core features (labels, faces, NSFW). Zero external deps.
- **Ollama**: private, GPU/CPU-hosted multimodal LLMs; best for richer captions/labels without cloud traffic.
- **OpenAI**: highest quality reasoning and multimodal support; requires API key and network access.

### Model Unload on Idle

PhotoPrism currently keeps TensorFlow models resident for the lifetime of the process to avoid repeated load costs. A future “model unload on idle” mode would track last-use timestamps and close the TensorFlow session/graph after a configurable idle period, releasing the model’s memory footprint back to the OS. The trade-off is higher latency and CPU overhead when a model is used again, plus extra I/O to reload weights. This may be attractive for low-frequency or memory-constrained deployments but would slow continuous indexing jobs, so it is not enabled today.

### Troubleshooting

- If face model initialization fails with `Read less bytes than requested` (often followed by `invalid face model configuration` in `GenerateFaceEmbeddings` tests), reinstall the local FaceNet assets:
  - `rm -f /tmp/photoprism/facenet.zip`
  - `rm -rf assets/models/facenet`
  - `make dep-tensorflow` (or `scripts/download-facenet.sh`)
  - Re-run: `go test ./internal/ai/face -run TestNet -count=1` and `go test ./internal/ai/vision -run TestGenerateFaceEmbeddings -count=1`

### Related Docs

- Ollama specifics: [`internal/ai/vision/ollama/README.md`](ollama/README.md)
- OpenAI specifics: [`internal/ai/vision/openai/README.md`](openai/README.md)
- REST API reference: https://docs.photoprism.dev/
- Developer guide (Vision): https://docs.photoprism.app/developer-guide/api/
