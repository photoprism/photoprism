## PhotoPrism — Vision Package

**Last Updated:** November 25, 2025

### Overview

`internal/ai/vision` provides the shared model registry, request builders, and parsers that power PhotoPrism’s caption, label, face, NSFW, and future generate workflows. It reads `vision.yml`, normalizes models, and dispatches calls to one of three engines:

- **TensorFlow (built‑in)** — default Nasnet / NSFW / Facenet models, no remote service required.
- **Ollama** — local or proxied multimodal LLMs. See [`ollama/README.md`](ollama/README.md) for tuning and schema details.
- **OpenAI** — cloud Responses API. See [`openai/README.md`](openai/README.md) for prompts, schema variants, and header requirements.

### Configuration

#### Models

The `vision.yml` file is usually kept in the `storage/config` directory (override with `PHOTOPRISM_VISION_YAML`). It defines a list of models under `Models:`. Key fields are captured below. If a type is omitted entirely, PhotoPrism will auto-append the built-in defaults (labels, nsfw, face, caption) so you no longer need placeholder stanzas. The `Thresholds` block is optional; missing or out-of-range values fall back to defaults.

| Field                   | Default                                | Notes                                                                              |
|-------------------------|----------------------------------------|------------------------------------------------------------------------------------|
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
|-----------------|------------------------------------------------------------------|------------------------------------------------|
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

| Option            | Default                                                                                 | Description                                                                        |
|-------------------|-----------------------------------------------------------------------------------------|------------------------------------------------------------------------------------|
| `Temperature`     | engine default (`0.1` for Ollama; unset for OpenAI)                                     | Controls randomness; clamped to `[0,2]`. `gpt-5*` OpenAI models are forced to `0`. |
| `TopP`            | engine default (`0.9` for some Ollama label defaults; unset for OpenAI)                 | Nucleus sampling parameter.                                                        |
| `MaxOutputTokens` | engine default (OpenAI caption 512, labels 1024; Ollama label default 256)              | Upper bound on generated tokens; adapters raise low values to defaults.            |
| `ForceJson`       | engine-specific (`true` for OpenAI labels; `false` for Ollama labels; captions `false`) | Forces structured output when enabled.                                             |
| `SchemaVersion`   | derived from schema name                                                                | Override when coordinating schema migrations.                                      |
| `Stop`            | engine default                                                                          | Array of stop sequences (e.g., `["\\n\\n"]`).                                      |
| `NumThread`       | runtime auto                                                                            | Caps CPU threads for local engines.                                                |
| `NumCtx`          | engine default                                                                          | Context window length (tokens).                                                    |

#### Model Service

Used for Ollama/OpenAI (and any future HTTP engines). All credentials and identifiers support `${ENV_VAR}` expansion.

| Field                              | Default                                  | Notes                                                |
|------------------------------------|------------------------------------------|------------------------------------------------------|
| `Uri`                              | required for remote                      | Endpoint base. Empty keeps model local (TensorFlow). |
| `Method`                           | `POST`                                   | Override verb if provider needs it.                  |
| `Key`                              | `""`                                     | Bearer token; prefer env expansion.                  |
| `Username` / `Password`            | `""`                                     | Injected as basic auth when URI lacks userinfo.      |
| `Model`                            | `""`                                     | Endpoint-specific override; wins over model/name.    |
| `Org` / `Project`                  | `""`                                     | OpenAI headers (org/proj IDs)                        |
| `RequestFormat` / `ResponseFormat` | set by engine alias                      | Explicit values win over alias defaults.             |
| `FileScheme`                       | set by engine alias (`data` or `base64`) | Controls image transport.                            |
| `Disabled`                         | `false`                                  | Disable the endpoint without removing the model.     |

### Field Behavior & Precedence

- Model identifier resolution order: `Service.Model` → `Model` → `Name`. `Model.GetModel()` returns `(id, name, version)` where Ollama receives `name:version` and other engines receive `name` plus a separate `Version`.
- Env expansion runs for all `Service` credentials and `Model` overrides; empty or disabled models return empty identifiers.
- Options merging: engine defaults fill missing fields; explicit values always win. Temperature is capped at `MaxTemperature`.
- Authentication: `Service.Key` sets `Authorization: Bearer <token>`; `Username`/`Password` inject HTTP basic auth into the service URI when not already present.

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
      Uri: http://ollama:11434/api/generate
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

### Related Docs

- Ollama specifics: [`internal/ai/vision/ollama/README.md`](ollama/README.md)
- OpenAI specifics: [`internal/ai/vision/openai/README.md`](openai/README.md)
- REST API reference: https://docs.photoprism.dev/
- Developer guide (Vision): https://docs.photoprism.app/developer-guide/api/
