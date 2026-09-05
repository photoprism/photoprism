## PhotoPrism — Ollama Engine Integration

**Last Updated:** September 5, 2026

### Overview

This package provides PhotoPrism’s native adapter for Ollama-compatible multimodal models. It lets Caption, Labels, and future Generate workflows call locally hosted models without changing worker logic, reusing the shared API client (`internal/ai/vision/api_client.go`) and result types (`LabelResult`, `CaptionResult`). Requests stay inside your infrastructure, rely on base64 thumbnails, and honor the same ACL, timeout, and logging hooks as the default local engines. The adapter resolves `${OLLAMA_BASE_URL}/api/generate`, trimming trailing slashes and defaulting to `http://ollama:11434`; set `OLLAMA_BASE_URL=https://ollama.com` to opt into cloud defaults.

#### Constraints

- Engine defaults live in `internal/ai/vision/ollama` and are applied whenever a model sets `Engine: ollama`. Aliases map to `ApiFormatOllama`, `scheme.Base64`, and a default 720 px thumbnail. The default model is `gemma4:latest` for self-hosted instances and `minimax-m3:cloud` when `OLLAMA_BASE_URL` equals `https://ollama.com` (cloud defaults are only selected on that exact match). Label normalization follows the same split: a model tagged `:cloud`, or one whose endpoint is the cloud host, defaults to `Normalize: phrase`, because hosted models only return a compound label name when the subject has one. Self-hosted models stay on `single-word`.
- Reasoning is disabled by default (`DefaultThink = "false"`, applied to `Service.Think` when empty) so thinking-capable models do not leak their reasoning into captions or invalidate label JSON. Re-enable it explicitly with `Service.Think: "true"`.
- Responses may arrive as newline-delimited JSON chunks. `decodeOllamaResponse` keeps the most recent chunk, while the parser supports both `response` and `thinking` fallbacks for captions and labels and strips a leading, well-delimited `<think>...</think>` block from the response body as a defensive fallback.
- Structured JSON is optional for captions but enforced for labels when `Format: json` (default for label models targeting the Ollama engine).
- The adapter never overwrites local defaults. If an Ollama call fails, downstream code still has label, NSFW, and face models available.
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
2. **Request Build** — `ollamaBuilder.Build` wraps thumbnails with `NewApiRequestOllama`, which encodes them as base64 strings. `Model.GetModel()` resolves the exact Ollama tag (`gemma4:latest`, `qwen2.5vl:7b`, etc.).
3. **Transport** — `PerformApiRequest` uses a single HTTP POST (default timeout 10 min). Authentication is optional; provide `Service.Key` if you proxy through an API gateway.
4. **Parsing** — `ollamaParser.Parse` converts payloads into `ApiResponse`. It normalizes confidences (`LabelConfidenceDefault = 0.5` when missing), copies NSFW scores, and canonicalizes label names via `normalizeLabelResult`, in the mode the model's `Normalize` setting selects.
5. **Persistence** — `entity.SrcOllama` is stamped on labels/captions so UI badges and audits reflect the new source.

### Prompt, Schema, & Options Guidance

- **System Prompts**
  - Labels: `LabelSystem` enforces single-word nouns. Set `System` to override; assign `LabelSystemSimple` when you need descriptive phrases, and pair it with `Normalize: phrase` so the phrases survive.
  - Captions: no system prompt by default; rely on user prompt or set one explicitly for stylistic needs.
- **User Prompts**
  - Captions use `CaptionPrompt`, which requests one sentence in active voice.
  - Labels default to `LabelPromptDefault`; when the package-level `DetectNSFWLabels` global is true, the adapter swaps in `LabelPromptNSFW`. The global is set by `config.go` to `DetectNSFW() && Experimental()`, so both `PHOTOPRISM_DETECT_NSFW=true` and `PHOTOPRISM_EXPERIMENTAL=true` are required to enable the NSFW-aware prompt.
  - For stricter noun enforcement, set `Prompt` to `LabelPromptStrict`.
- **Schemas**
  - Labels rely on `schema.LabelsJson(nsfw)` (simple JSON template). Setting `Format: json` auto-attaches a reminder (`model.SchemaInstructions()`).
  - Override via `Schema` (inline YAML) or `SchemaFile`. `PHOTOPRISM_VISION_LABEL_SCHEMA_FILE` always wins if present.
- **Options**
  - Labels: default `Temperature` equals `DefaultTemperature` (0.1 unless configured), `TopP=0.9`, `Stop=["\n\n"]`.
  - Captions: only `Temperature` is set; other parameters inherit global defaults.
  - Custom `Options` merge with engine defaults. Leave `ForceJson=true` for labels so PhotoPrism can reject malformed payloads early.

### Supported Ollama Vision Models

> Recommended defaults: `gemma4:latest` for self-hosted instances and `minimax-m3:cloud` for Ollama Cloud. Both return reliable JSON on stock settings, which is what a default has to do. `minimax-m3:cloud` also stays in the requested language for labels and captions alike; `gemma4:latest` does so for German but only two thirds of the time for Arabic and Hebrew, and it is weaker than similarly sized models at naming uncommon subjects — see the caveats below before converting a large library. The table below lists additional models that work well for specific use cases. Latency and label counts come from a fixed 16-image benchmark on an RTX 4060 (8 GB VRAM); expect different absolute numbers on other hardware, and re-measure before switching a large library over.

| Model (Ollama Tag)          | Size & Footprint                                                                                                                                                                                   | Strengths                                                                                                    | JSON & Language Notes                                                                                                                                                                                                 | When To Use                                                                                                                                            |
|:----------------------------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:-------------------------------------------------------------------------------------------------------------|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:-------------------------------------------------------------------------------------------------------------------------------------------------------|
| `gemma4:e2b / e4b (latest)` | ~7.2 GB (e2b) / ~9.6 GB (e4b) downloads; both fit 8 GB VRAM — the nested architecture loads only the active parameters, measured at 1.7 GB (e2b) and 3.2 GB (e4b) resident at a 4096-token context | Fastest captions of any self-hosted model tested (~0.7 s); the only family that never emitted a phrase label | Emits clean JSON (any code fences stripped by `clean.JSON`); strong multilingual output                                                                                                                               | **Recommended self-hosted default** when JSON hygiene matters most. Weaker at identifying uncommon subjects than Qwen3-VL at a comparable size         |
| `qwen3-vl:2b / 4b / 8b`     | Dense 2B/4B/8B tiers (~1.9 GB, ~3.3 GB, ~6.1 GB downloads) with native 256 K context extendable to 1 M                                                                                             | Best subject accuracy per gigabyte in the benchmark; 32-language OCR, spatial/video reasoning, GUI grounding | Emits JSON reliably when prompts specify schema; prefer the `-instruct` tags — the reasoning builds cost ~17x the output tokens and 5x the caption latency for no visible gain, and return far more multi-word labels | General-purpose captions/labels; `4b-instruct` is the best all-round self-hosted pick, `8b-instruct` only if you can absorb roughly double the latency |
| `gemma3:4b / 12b / 27b`     | 4B/12B/27B parameters, ~3.3 GB to 17 GB downloads, 128 K context                                                                                                                                   | Multimodal text+image reasoning with SigLIP encoder, handles OCR/long documents, tool/function calling       | Emits structured JSON reliably; >140 languages with strong default English output                                                                                                                                     | Still competitive at 4B; prefer `gemma4:e2b` for lower latency or `qwen3-vl:4b-instruct` for accuracy                                                  |
| `qwen3.5:4b / 9b`           | ~3.4 GB / ~6.6 GB downloads, both fit 8 GB VRAM                                                                                                                                                    | Fast, clean single-pass `response`; solid captions                                                           | Thinking-capable, so keep `Service.Think: "false"` (the default)                                                                                                                                                      | Smaller Qwen-family alternative; the 4B tier outperforms the 9B tier on labels, so do not assume bigger is better here                                 |
| `minicpm-v4.5:8b`           | 8 B params, ~6.1 GB download, 40 K context                                                                                                                                                         | Strong OCR, multi-image/video support, detailed captions (~28 words)                                         | Multilingual. Emits `snake_case` compounds more often than other models                                                                                                                                               | Memory-constrained deployments that need OCR-heavy captions                                                                                            |
| `minicpm-v4.6:1b`           | ~752 M params, ~1.6 GB download                                                                                                                                                                    | Runs almost anywhere; fastest labeler tested (~0.9 s) and identifies subjects better than its size suggests  | Highest phrase rate of any model measured (~23%), so labels need close review                                                                                                                                         | CPU-only or heavily shared GPUs where latency matters more than label hygiene                                                                          |
| `qwen2.5vl:7b`              | 8.29 B params (Q4_K_M) ~6 GB download, 125 K context                                                                                                                                               | Charts, GUI grounding, DocVQA, multi-image reasoning                                                         | JSON mode tuned for schema compliance; 20+ languages                                                                                                                                                                  | Captions and document analysis. It returned the fewest usable labels of any general model tested, so prefer Qwen3-VL for label generation              |
| `llama3.2-vision:11b`       | 11 B params, ~7.8 GB download, requires >=8 GB VRAM; 90 B variant needs >=64 GB                                                                                                                    | Strong general reasoning, captioning, OCR, supported by Meta ecosystem tooling                               | Vision tasks officially supported in English; text-only tasks cover eight major languages                                                                                                                             | Keep captions consistent with Meta-compatible prompts or when teams already standardize on Llama 3.x                                                   |

> **Label count is a deliberate default.** `LabelPromptDefault` does not state how many labels to return, because a short list of high-confidence labels is more useful and cheaper than a long one — count multiplies through the database, the API response, and the UI — and models that read an image poorly mostly add noise when pushed for more. Models differ widely in what they volunteer: hosted models return 7-12 per image, models that fit in 8 GB return 1-4. Adding an explicit count range to a model's `Prompt` multiplies the set by 1.9-3.5x at roughly double the latency and more names that break the single-word contract, so treat it as per-model tuning to be measured, not as a fix.

> **A model's phrase rate matters under the default normalization.** `normalizeLabelResult` reduces a phrase to one token and usually keeps the wrong one: `ferris wheel` is stored as `Ferris`, `amusement park` as `Park`, and `lifeguard tower` as `Tower`. That is the `single-word` default, not a fixed rule — set `Normalize: phrase` on the model, together with a `System` prompt that permits phrases (`LabelSystemSimple`), to store the compound name instead. Models that concatenate (`ferriswheel`) store an unsearchable token that no mode can repair, so the phrase rate stays a selection criterion for anyone running the default.

> **`Service.Think: "false"` keeps reasoning out of the result; it does not stop the model reasoning.** On `qwen3-vl:4b` the flag moved caption latency from 7.6 s to 6.6 s and output from 528 to 414 tokens — the work still happens, it is just not shown. Its `-instruct` sibling produced a longer caption in 1.2 s using 24 tokens, and returned 0 % multi-word labels against 21 % for the reasoning build. Treat the flag as a correctness guard and pick the `-instruct` tag for throughput.

> **PhotoPrism never asks for a language.** Neither the shipped system prompt nor the user prompt names one, so a model answers in whatever language it defaults to — English for every model measured. A non-English library only gets non-English output if the model's `Prompt` or `System` says so:
>
> ```yaml
>     Prompt: |
>       Analyze the image and return label objects with name, confidence (0-1), and topicality (0-1).
>       Respond in German.
> ```

> **A model that follows a language instruction in prose may drop it for labels.** Requesting German, Arabic, or Hebrew does not affect the two tasks equally, so verify both on a sample before converting a library. Measured on the 16-image benchmark, share of requests whose **labels** came back in the requested language: `gemma4:latest` 100 % German, 69 % Arabic, 62 % Hebrew; `gemma4:e2b` 62 % German and **0 %** for both Arabic and Hebrew while captioning them correctly; `qwen3-vl:4b-instruct` 100 % / 31 % / 75 %; `qwen3.5:4b` 81 % / 94 % / 94 %. Several cloud models show a milder version of the same split. Nothing surfaces it — the response is valid JSON with plausible labels, so no error is logged and no fallback fires.

> **Fluent output is not correct output.** `qwen3.5:4b` has the best in-language rate of any self-hosted model measured and the worst accuracy: its Hebrew is grammatical but names the wrong things, captioning a photo of an elephant as "the lion crushes the birds". Checking that the output is in the right alphabet proves nothing about its meaning, so review a sample of the actual content rather than trusting that a model "supports" a language.

> **Medical vision models are not general photo models.** `medgemma1.5:4b` returns `{"labels": []}` for ordinary photos because it is trained to emit grounded detections (`box_2d` plus a label), and it leaks `<start_of_image>` markers into captions. `medgemma:4b` fills the schema but hallucinates outside its domain. Neither is usable for a general library without a dedicated prompt and schema.

> Tip: pull models inside the dev container with `docker compose --profile ollama up -d` and then `docker compose exec ollama ollama pull gemma4:latest`. Keep the profile stopped when you do not need extra GPU/CPU load.

> Do not trust the `capabilities` list from `GET /api/tags` to decide whether a pulled model is multimodal or reasoning-capable. Models that answer image prompts correctly are sometimes reported without a `vision` flag, and the same model can be listed with and without `thinking` by `/api/tags` and `/api/show`. Send a request and check the response instead.

> Qwen3-VL models can stream structured output via `thinking` while leaving `response` empty. The parser checks `response` first and falls back to `thinking`, so captions/labels continue to work with either field.

#### Ollama Cloud Models

Set `OLLAMA_BASE_URL=https://ollama.com` and provide `OLLAMA_API_KEY` to use hosted models (no local download or GPU required). The default cloud model is `minimax-m3:cloud`. The cloud catalog changes over time and models are occasionally retired without notice, so treat this as a snapshot and consult <https://ollama.com/search?c=cloud> for the current list; PhotoPrism logs a clear warning when a configured cloud model returns HTTP 404/410.

The table below reports median single-image latency over a fixed 16-image benchmark, and how reliably each model honors a requested output language. All six returned well-formed JSON for every English request, so the differences are in verbosity, speed, and language handling rather than reliability.

| Model (Cloud Tag)      | Latency (caption / labels) | Labels/Image | Non-English Labels | Notes                                                                                                                                                                                              |
|:-----------------------|:---------------------------|-------------:|:-------------------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `minimax-m3:cloud`     | ~3 s / ~3-5 s              |         13.4 | Reliable           | **Recommended default.** The only model that stayed in the requested language for both labels and captions in German, Arabic, and Hebrew. Largest label sets                                       |
| `kimi-k2.7-code:cloud` | ~1 s / ~2 s                |          8.6 | Unreliable         | Consistently the fastest labeler at equal English coverage, but drops the language instruction on roughly two thirds of non-English label requests. English-only libraries                         |
| `gemma4:31b-cloud`     | ~1-4 s / ~2-11 s           |          7.0 | Reliable           | The most conservative labeler; misreads some subjects (skier as snowboarder, omits `dome`). Its latency varied by more than 6x between two runs on the same day                                    |
| `qwen3.5:397b-cloud`   | ~4 s / ~5-6 s              |          8.4 | Mixed              | Strong captions; about 70 % of non-English label sets stay in the requested language                                                                                                               |
| `kimi-k2.6:cloud`      | ~2 s / ~4 s                |          8.6 | Not measured       | Superseded by newer models on both speed and accuracy                                                                                                                                              |
| `kimi-k3:cloud`        | ~1 s / ~4 s                |          9.1 | Not measured       | Best English quality measured, but **outside both the free and paid plans** — it needs a Pro or Max subscription plus metered usage credits, so it is not a candidate for library-scale generation |

> **Cloud latency swings between runs; self-hosted latency does not.** Repeating the benchmark hours apart on the same day moved `gemma4:31b-cloud` from ~2 s to ~11 s on labels while its output barely changed, and self-hosted models reproduced to within a tenth of a second. Size a cloud deployment on output quality and plan coverage, and measure latency against your own instance at the time it matters.

> **Prompt tokens per image depend on the model, not just the thumbnail size.** The same 720 px image cost 208 prompt tokens on `gemma4` and 1,182 on `qwen3-vl` — a 5.7x spread for identical input. Where tokens are billed per request, that difference multiplies by the size of the library, so check it alongside latency when comparing models.

> **Check plan coverage before choosing a cloud model.** Ollama Cloud bills by usage tier, and some models sit outside the plans entirely and consume extra credits per token. That is a larger cost factor than latency when captioning or labeling a whole library, so confirm a model's terms on its page before enabling it.

> **Cloud tag naming.** Against `https://ollama.com` the plain tag works (`gemma4:31b`, `qwen3.5:397b`). A self-hosted instance proxying to Ollama Cloud needs the explicit cloud tag (`gemma4:31b-cloud`, `qwen3.5:397b-cloud`) — the plain form is resolved as a local pull and fails with "model not found". Cloud-only models such as `kimi-k3` and `minimax-m3` accept the `:cloud` suffix either way.

### Configuration

#### Environment Variables

- `PHOTOPRISM_VISION_LABEL_SCHEMA_FILE` — Absolute path to a JSON snippet that overrides the default label schema (applies to every Ollama label model).
- `PHOTOPRISM_VISION_YAML` — Custom `vision.yml` path. Keep it synced in Git if you automate deployments.
- `OLLAMA_HOST`, `OLLAMA_MODELS`, `OLLAMA_MAX_QUEUE`, `OLLAMA_NUM_PARALLEL`, etc. — Provided in `compose*.yaml` to tune the Ollama daemon. Adjust `OLLAMA_KEEP_ALIVE` if you want models to stay loaded between worker batches.
- `OLLAMA_API_KEY` / `OLLAMA_API_KEY_FILE` — Default bearer token picked up when `Service.Key` is empty; useful for hosted Ollama services (e.g., Ollama Cloud).
- `OLLAMA_BASE_URL` — Base URL for the Ollama API; defaults to `http://ollama:11434`, trailing slashes are trimmed. Set to `https://ollama.com` to enable cloud defaults.
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
      Uri: ${OLLAMA_BASE_URL}/api/generate
      RequestFormat: ollama
      ResponseFormat: ollama
      FileScheme: base64
      Think: "false"

  - Type: caption
    Name: gemma4:latest
    Engine: ollama
    Disabled: false
    Options:
      Temperature: 0.2
    Service:
      Uri: ${OLLAMA_BASE_URL}/api/generate
      Think: "false"
```

Guidelines:

- Place new entries after the default local models so they take precedence while local label and NSFW models remain as fallbacks.
- Always specify the exact Ollama tag (`model:version`) so upgrades are deliberate.
- `Service.Think` defaults to `"false"` for the Ollama engine (reasoning off) and is sent whenever non-empty. Keep it quoted (for example `"false"`, `"true"`, or `"low"`) so YAML preserves it as a string; PhotoPrism serializes `"true"` / `"false"` as JSON booleans for Ollama compatibility. Set `Service.Think: "true"` to re-enable reasoning for a model that benefits from it.
- Model support is not universal: `think:true` may fail on models that do not implement reasoning, and `think:false` can still yield empty `response` fields on some reasoning-capable models (which then stream their JSON via the `thinking` field — the parser handles this).
- Keep option flags before positional arguments in CLI snippets (`photoprism vision run -m labels --count 1`).
- If you proxy requests (e.g., through Traefik), set `Service.Key` to `Bearer <token>` and configure the proxy to inject/validate it.

### Operational Checklist

- **Scheduling** — Use `Run: newly-indexed` for incremental runs, `Run: manual` for ad-hoc CLI calls, or `Run: on-schedule` when paired with the scheduler. Leave `Run: auto` if you want the worker to decide based on other model states.
- **Timeouts & Retries** — Default timeout is 10 minutes (`ServiceTimeout`). Transient `HTTP 429` responses are retried with bounded exponential backoff (within `ServiceTimeout`, honoring `Retry-After` up to `ServiceRetryMaxDelay`); other errors are terminal. Ollama streaming responses complete faster in practice; if you need stricter SLAs, wrap `photoprism vision run` in a job runner and retry failed batches manually.
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
- `internal/ai/vision/models.go` — Default caption model definition (`gemma4:latest`, `minimax-m3:cloud` for Ollama Cloud).
- `compose*.yaml` — Ollama service profile, Traefik labels, and persistent volume wiring.
- `frontend/src/common/util.js` — Maps `src="ollama"` to the correct badge; keep it updated when adding new source strings.

### Next Steps

- [ ] Add formal schema validation (JSON Schema or JTD) so malformed label responses fail fast before normalization.
- [ ] Support multiple thumbnails per request once core workflows confirm the API contract (requires worker + UI changes).
- [ ] Emit per-model latency and success metrics from the vision worker to simplify tuning when several Ollama engines run side-by-side.
- [ ] Mirror any loader changes into PhotoPrism Plus/Pro templates to keep splash + browser checks consistent after enabling external engines.
