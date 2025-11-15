## PhotoPrism — Ollama Engine Integration

**Last Updated:** November 14, 2025

### Overview

This package provides PhotoPrism’s native adapter for Ollama-compatible multimodal models. It lets Caption, Labels, and future Generate workflows call locally hosted models without changing worker logic, reusing the shared API client (`internal/ai/vision/api_client.go`) and result types (`LabelResult`, `CaptionResult`). Requests stay inside your infrastructure, rely on base64 thumbnails, and honor the same ACL, timeout, and logging hooks as the default TensorFlow engines.

#### Context & Constraints

- Engine defaults live in `internal/ai/vision/ollama` and are applied whenever a model sets `Engine: ollama`. Aliases map to `ApiFormatOllama`, `scheme.Base64`, and a default 720 px thumbnail.  
- Responses may arrive as newline-delimited JSON chunks. `decodeOllamaResponse` keeps the most recent chunk, while `parseOllamaLabels` replays plain JSON strings found in `response`.  
- Structured JSON is optional for captions but enforced for labels when `Format: json` (default for label models targeting the Ollama engine).  
- The adapter never overwrites TensorFlow defaults. If an Ollama call fails, downstream code still has Nasnet, NSFW, and Face models available.  
- Workers assume a single-image payload per request. Run `photoprism vision run` to validate multi-image prompts before changing that invariant.

#### Goals

- Let operators opt into local, private LLMs for captions and labels via `vision.yml`.  
- Provide safe defaults (prompts, schema, sampling) so most deployments only need to specify `Name`, `Engine`, and `Service.Uri`.  
- Surface reproducible logs, metrics, and CLI commands that make it easy to compare Ollama output against TensorFlow/OpenAI engines.

#### Non-Goals

- Managing Ollama itself (model downloads, GPU scheduling, or authentication). Use the Compose profiles provided in the repository.  
- Adding new HTTP endpoints or bypassing the existing `photoprism vision` CLI.  
- Replacing TensorFlow workers—Ollama engines are additive and opt-in.

### Architecture & Request Flow

1. **Model Selection** — `Config.Model(ModelType)` returns the top-most enabled entry. When `Engine: ollama`, `ApplyEngineDefaults()` fills in the request/response format, base64 file scheme, and a 720 px resolution unless overridden.  
2. **Request Build** — `ollamaBuilder.Build` wraps thumbnails with `NewApiRequestOllama`, which encodes them as base64 strings. `Model.Model()` resolves the exact Ollama tag (`gemma3:4b`, `qwen2.5vl:7b`, etc.).  
3. **Transport** — `PerformApiRequest` uses a single HTTP POST (default timeout 10 min). Authentication is optional; provide `Service.Key` if you proxy through an API gateway.  
4. **Parsing** — `ollamaParser.Parse` converts payloads into `ApiResponse`. It normalizes confidences (`LabelConfidenceDefault = 0.5` when missing), copies NSFW scores, and canonicalizes label names via `normalizeLabelResult`.  
5. **Persistence** — `entity.SrcOllama` is stamped on labels/captions so UI badges and audits reflect the new source.

### Prompt, Schema, & Options Guidance

- **System Prompts**  
  - Labels: `LabelSystem` enforces single-word nouns. Set `System` to override; assign `LabelSystemSimple` when you need descriptive phrases.  
  - Captions: no system prompt by default; rely on user prompt or set one explicitly for stylistic needs.
- **User Prompts**  
  - Captions use `CaptionPrompt`, which requests one sentence in active voice.  
  - Labels default to `LabelPromptDefault`; when `DetectNSFWLabels` is true, the adapter swaps in `LabelPromptNSFW`.  
  - For stricter noun enforcement, set `Prompt` to `LabelPromptStrict`.  
- **Schemas**  
  - Labels rely on `schema.LabelsJson(nsfw)` (simple JSON template). Setting `Format: json` auto-attaches a reminder (`model.SchemaInstructions()`).  
  - Override via `Schema` (inline YAML) or `SchemaFile`. `PHOTOPRISM_VISION_LABEL_SCHEMA_FILE` always wins if present.  
- **Options**  
  - Labels: default `Temperature` equals `DefaultTemperature` (0.1 unless configured), `TopP=0.9`, `Stop=["\n\n"]`.  
  - Captions: only `Temperature` is set; other parameters inherit global defaults.  
  - Custom `Options` merge with engine defaults. Leave `ForceJson=true` for labels so PhotoPrism can reject malformed payloads early.

### Supported Ollama Vision Models

| Model (Ollama Tag)      | Size & Footprint                                                                                                                                    | Strengths                                                                                                                   | JSON & Language Notes                                                                                                        | When To Use                                                                                                                                                                  |
|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `gemma3:4b / 12b / 27b` | 4B/12B/27B parameters, ~3.3 GB → 17 GB downloads, 128 K context                                                                                     | Multimodal text+image reasoning with SigLIP encoder, handles OCR/long documents, supports tool/function calling             | Emits structured JSON reliably; >140 languages with strong default English output                                            | High-quality captions + multilingual labels when you have ≥12 GB VRAM (4B works on 8 GB with Q4_K_M)                                                                         |
| `qwen2.5vl:7b`          | 8.29 B params (Q4_K_M) ≈6 GB download, 125 K context                                                                                                | Excellent charts, GUI grounding, DocVQA, multi-image reasoning, agentic tool use                                            | JSON mode tuned for schema compliance; supports 20+ languages with strong Chinese/English parity                             | Label extraction for mixed-language archives or UI/diagram analysis                                                                                                          |
| `qwen3-vl:2b / 4b / 8b` | Dense 2B/4B/8B tiers (~3 GB, ~3.5 GB, ~6 GB downloads) with native 256 K context extendable to 1 M; fits single 12–24 GB GPUs or high-end CPUs (2B) | Spatial + video reasoning upgrades (Interleaved-MRoPE, DeepStack), 32-language OCR, GUI/agent control, long-document ingest | Emits JSON reliably when prompts specify schema; multilingual captions/labels with Thinking variants boosting STEM reasoning | General-purpose captions/labels when you need long-context doc/video support without cloud APIs; 2B for CPU/edge, 4B as balanced default, 8B when accuracy outweighs latency |
| `llama3.2-vision:11b`   | 11 B params, ~7.8 GB download, requires ≥8 GB VRAM; 90 B variant needs ≥64 GB                                                                       | Strong general reasoning, captioning, OCR, supported by Meta ecosystem tooling                                              | Vision tasks officially supported in English; text-only tasks cover eight major languages                                    | Keep captions consistent with Meta-compatible prompts or when teams already standardize on Llama 3.x                                                                         |
| `minicpm-v:8b-2.6`      | 8 B params, ~5.5 GB download, 32 K context                                                                                                          | Optimized for edge GPUs, high OCR accuracy, multi-image/video support, low token count (≈640 tokens for 1.8 MP)             | Multilingual (EN/ZH/DE/FR/IT/KR). Emits concise JSON but may need stricter stopping sequences                                | Memory-constrained deployments that still require NSFW/OCR-aware label output                                                                                                |

> Tip: pull models inside the dev container with `docker compose --profile ollama up -d` and then `docker compose exec ollama ollama pull gemma3:4b`. Keep the profile stopped when you do not need extra GPU/CPU load.

> Qwen3-VL models stream their JSON payload via the `thinking` field. PhotoPrism v2025.11+ captures this automatically; if you run older builds, upgrade before enabling these models or responses will appear empty.

### Configuration

#### Environment Variables

- `PHOTOPRISM_VISION_LABEL_SCHEMA_FILE` — Absolute path to a JSON snippet that overrides the default label schema (applies to every Ollama label model).  
- `PHOTOPRISM_VISION_YAML` — Custom `vision.yml` path. Keep it synced in Git if you automate deployments.  
- `OLLAMA_HOST`, `OLLAMA_MODELS`, `OLLAMA_MAX_QUEUE`, `OLLAMA_NUM_PARALLEL`, etc. — Provided in `compose*.yaml` to tune the Ollama daemon. Adjust `OLLAMA_KEEP_ALIVE` if you want models to stay loaded between worker batches.  
- `PHOTOPRISM_LOG_LEVEL=trace` — Enables verbose request/response previews (truncated to avoid leaking images). Use temporarily when debugging parsing issues.

#### `vision.yml` Example

```yaml
Models:
  - Type: labels
    Name: qwen2.5vl:7b
    Engine: ollama
    Run: newly-indexed
    Resolution: 720
    Format: json
    Options:
      Temperature: 0.05
      Stop: ["\n\n"]
      ForceJson: true
    Service:
      Uri: http://ollama:11434/api/generate
      RequestFormat: ollama
      ResponseFormat: ollama
      FileScheme: base64

  - Type: caption
    Name: gemma3:4b
    Engine: ollama
    Disabled: false
    Options:
      Temperature: 0.2
    Service:
      Uri: http://ollama:11434/api/generate
```

Guidelines:

- Place new entries after the default TensorFlow models so they take precedence while Nasnet/NSFW remain as fallbacks.  
- Always specify the exact Ollama tag (`model:version`) so upgrades are deliberate.  
- Keep option flags before positional arguments in CLI snippets (`photoprism vision run -m labels --count 1`).  
- If you proxy requests (e.g., through Traefik), set `Service.Key` to `Bearer <token>` and configure the proxy to inject/validate it.

### Operational Checklist

- **Scheduling** — Use `Run: newly-indexed` for incremental runs, `Run: manual` for ad-hoc CLI calls, or `Run: on-schedule` when paired with the scheduler. Leave `Run: auto` if you want the worker to decide based on other model states.  
- **Timeouts & Retries** — Default timeout is 10 minutes (`ServiceTimeout`). Ollama streaming responses complete faster in practice; if you need stricter SLAs, wrap `photoprism vision run` in a job runner and retry failed batches manually.  
- **Fallbacks** — Keep Nasnet configured even when Ollama labels are primary. `labels.go` stops at the first successful engine, so duplicates are avoided.  
- **Security** — When exposing Ollama beyond localhost, terminate TLS at Traefik and enable API keys. Never return full JSON payloads in logs; rely on trace mode only for debugging and sanitize before sharing.  
- **Model Storage** — Bind-mount `./storage/services/ollama:/root/.ollama` (see Compose) so pulled models survive container restarts. Run `docker compose exec ollama ollama list` during deployments to verify availability.

### Observability & Testing

- **CLI Smoke Tests**  
  - Captions: `photoprism vision run -m caption --count 5 --force`.  
  - Labels: `photoprism vision run -m labels --count 5 --force`.  
  - After each run, check `photoprism vision ls` for `source=ollama`.  
- **Unit Tests**  
  - `go test ./internal/ai/vision/ollama ./internal/ai/vision -run Ollama -count=1` covers transport parsing and model defaults.  
  - Add fixtures under `internal/ai/vision/testdata` when capturing new response shapes; keep files small and anonymized.  
- **Logging**  
  - Set `PHOTOPRISM_LOG_LEVEL=debug` to watch summary lines (“processed labels/caption via ollama”).  
  - Use `log.Trace` sparingly; it prints truncated JSON blobs for troubleshooting.  
- **Metrics**  
  - `/api/v1/metrics` exposes counts per label source; scrape after a batch to compare throughput with TensorFlow/OpenAI runs.

### Code Map

- `internal/ai/vision/ollama/*.go` — Engine defaults, schema helpers, transport structs.  
- `internal/ai/vision/engine_ollama.go` — Builder/parser glue plus label/caption normalization.  
- `internal/ai/vision/api_ollama.go` — Base64 payload builder.  
- `internal/ai/vision/api_client.go` — Streaming decoder shared among engines.  
- `internal/ai/vision/models.go` — Default caption model definition (`gemma3`).  
- `compose*.yaml` — Ollama service profile, Traefik labels, and persistent volume wiring.  
- `frontend/src/common/util.js` — Maps `src="ollama"` to the correct badge; keep it updated when adding new source strings.

### Next Steps

- [ ] Add formal schema validation (JSON Schema or JTD) so malformed label responses fail fast before normalization.  
- [ ] Support multiple thumbnails per request once core workflows confirm the API contract (requires worker + UI changes).  
- [ ] Emit per-model latency and success metrics from the vision worker to simplify tuning when several Ollama engines run side-by-side.  
- [ ] Mirror any loader changes into PhotoPrism Plus/Pro templates to keep splash + browser checks consistent after enabling external engines.
