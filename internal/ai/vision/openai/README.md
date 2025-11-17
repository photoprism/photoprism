## PhotoPrism — OpenAI API Integration

**Last Updated:** November 14, 2025

### Overview

This package contains PhotoPrism’s adapter for the OpenAI Responses API. It enables existing caption and label workflows (`GenerateCaption`, `GenerateLabels`, and the `photoprism vision run` CLI) to call OpenAI models alongside TensorFlow and Ollama without changing worker or API code. The implementation focuses on predictable results, structured outputs, and clear observability so operators can opt in gradually.

#### Context & Constraints

- OpenAI requests flow through the existing vision client (`internal/ai/vision/api_client.go`) and must honour PhotoPrism’s timeout, logging, and ACL rules.
- Structured outputs are preferred but the adapter must gracefully handle free-form text; `output_text` responses are parsed both as JSON and as plain captions.
- Costs should remain predictable: requests are limited to a single 720 px thumbnail (`detail=low`) and capped token budgets (512 caption, 1024 labels).
- Secrets are supplied per model (`Service.Key`) with fallbacks to `OPENAI_API_KEY` / `_FILE`. Logs must redact sensitive data.

#### Goals

- Provide drop-in OpenAI support for captions and labels using `vision.yml`.
- Keep configuration ergonomic by auto-populating prompts, schema names, token limits, and sampling defaults.
- Expose enough logging and tests so operators can compare OpenAI output with existing engines before enabling it broadly.

#### Non-Goals

- Introducing a new `generate` model type or combined caption/label endpoint (reserved for a later phase).
- Replacing the default TensorFlow models; they remain active as fallbacks.
- Managing OpenAI billing or quota dashboards beyond surfacing token counts in logs and metrics.

### Prompt, Model, & Schema Guidance

- **Models:** The adapter targets GPT‑5 vision tiers (e.g. `gpt-5-nano`, `gpt-5-mini`). These models support image inputs, structured outputs, and deterministic settings. Set `Name` to the exact provider identifier so defaults are applied correctly. Caption models share the same configuration surface and run through the same adapter.
- **Prompts:** Defaults live in `defaults.go`. Captions use a single-sentence instruction; labels use `LabelPromptDefault` (or `LabelPromptNSFW` when PhotoPrism requests NSFW metadata). Custom prompts should retain schema reminders so structured outputs stay valid.
- **Schemas:** Labels use the JSON schema returned by `schema.LabelsJsonSchema(nsfw)`; the response format name is derived via `schema.JsonSchemaName` (e.g. `photoprism_vision_labels_v1`). Captions omit schemas unless operators explicitly request a structured format.
- **When to keep defaults:** For most deployments, leaving `System`, `Prompt`, `Schema`, and `Options` unset yields stable output with minimal configuration. Override them only when domain-specific language or custom scoring is necessary, and add regression tests alongside.

Budget-conscious operators can experiment with lighter prompts or lower-resolution thumbnails, but should keep token limits and determinism settings intact to avoid unexpected bills and UI churn.

#### Performance & Cost Estimates

- **Token budgets:** Captions request up to 512 output tokens; labels request up to 1024. Input tokens are typically ≤700 for a single 720 px thumbnail plus prompts.
- **Latency:** GPT‑5 nano/mini vision calls typically complete in 3–8 s, depending on OpenAI region. Including reasoning metadata (`reasoning.effort=low`) has negligible impact but improves traceability.
- **Costs:** Consult OpenAI’s pricing for the selected model. Multiply input/output tokens by the published rate. PhotoPrism currently sends one image per request to keep costs linear with photo count.

### Configuration

#### Environment Variables

- `OPENAI_API_KEY` / `OPENAI_API_KEY_FILE` — fallback credentials when a model’s `Service.Key` is unset.
- Existing `PHOTOPRISM_VISION_*` variables remain authoritative (see the [Getting Started Guide](https://docs.photoprism.app/getting-started/config-options/#computer-vision) for full lists).

#### `vision.yml` Examples

```yaml
Models:
  - Type: caption
    Name: gpt-5-nano
    Engine: openai
    Disabled: false    # opt in manually
    Resolution: 720    # optional; default is 720
    Options:
      Detail: low      # optional; defaults to low
      MaxOutputTokens: 512
    Service:
      Uri: https://api.openai.com/v1/responses
      FileScheme: data
      Key: ${OPENAI_API_KEY}

  - Type: labels
    Name: gpt-5-mini
    Engine: openai
    Disabled: false
    Resolution: 720
    Options:
      Detail: low
      MaxOutputTokens: 1024
      ForceJson: true  # redundant but explicit
    Service:
      Uri: https://api.openai.com/v1/responses
      FileScheme: data
      Key: ${OPENAI_API_KEY}
```

Keep TensorFlow entries in place so PhotoPrism falls back when the external service is unavailable.

#### Defaults

- File scheme: `data:` URLs (base64) for all OpenAI models.
- Resolution: 720 px thumbnails (`vision.Thumb(ModelTypeCaption|Labels)`).
- Options: `MaxOutputTokens` raised to 512 (caption) / 1024 (labels); `ForceJson=false` for captions, `true` for labels; `reasoning.effort="low"`.
- Sampling: `Temperature` and `TopP` set to `0` for `gpt-5*` models; inherited values (0.1/0.9) remain for other engines. `openaiBuilder.Build` performs this override while preserving the struct defaults for non-OpenAI adapters.
- Schema naming: Automatically derived via `schema.JsonSchemaName`, so operators may omit `SchemaVersion`.

### Documentation

- Label Generation: <https://docs.photoprism.app/developer-guide/vision/label-generation/>
- Caption Generation: <https://docs.photoprism.app/developer-guide/vision/caption-generation/>
- Vision CLI Commands: <https://docs.photoprism.app/developer-guide/vision/cli/>

### Implementation Details

#### Core Concepts

- **Structured outputs:** PhotoPrism leverages OpenAI’s structured output capability as documented at <https://platform.openai.com/docs/guides/structured-outputs>. When a JSON schema is supplied, the adapter emits `text.format` with `type: "json_schema"` and a schema name derived from the content. The parser then prefers `output_json`, but also attempts to decode `output_text` payloads that contain JSON objects.
- **Deterministic sampling:** GPT‑5 models are run with `temperature=0` and `top_p=0` to minimise variance, while still allowing developers to override values in `vision.yml` if needed.
- **Reasoning metadata:** Requests include `reasoning.effort="low"` so OpenAI returns structured reasoning usage counters, helping operators track token consumption.
- **Worker summaries:** The vision worker now logs either “updated …” or “processed … (no metadata changes detected)”, making reruns easy to audit.

#### Rate Limiting

OpenAI calls respect the existing `limiter.Auth` configuration used by the vision service. Failed requests surface standard HTTP errors and are not automatically retried; operators should ensure they have adequate account limits and consider external rate limiting when sharing credentials.

#### Testing & Validation

1. Unit tests: `go test ./internal/ai/vision/openai ./internal/ai/vision -run OpenAI -count=1`. Fixtures under `internal/ai/vision/openai/testdata/` replay real Responses payloads (captions and labels).
2. CLI smoke test: `photoprism vision run -m labels --count 1 --force` with trace logging enabled to inspect sanitised Responses.
3. Compare worker summaries and label sources (`openai`) in the UI or via `photoprism vision ls`.

#### Code Map

- **Adapter & defaults:** `internal/ai/vision/openai` (defaults, schema helpers, transport, tests).
- **Request/response plumbing:** `internal/ai/vision/api_request.go`, `api_client.go`, `engine_openai.go`, `engine_openai_test.go`.
- **Workers & CLI:** `internal/workers/vision.go`, `internal/commands/vision_run.go`.
- **Shared utilities:** `internal/ai/vision/schema`, `pkg/clean`, `pkg/media`.

#### Next Steps

- [ ] Introduce the future `generate` model type that combines captions, labels, and optional markers.
- [ ] Evaluate additional OpenAI models as pricing and capabilities evolve.
- [ ] Expose token usage metrics (input/output/reasoning) via Prometheus once the schema stabilises.
