## PhotoPrism — Vision Schema Reference

**Last Updated:** November 14, 2025

### Overview

This package contains the canonical label response specifications used by PhotoPrism’s external vision engines. It exposes two helpers:

- `LabelsJsonSchema(nsfw bool)` — returns a JSON **Schema** document tailored for OpenAI Responses requests, enabling strict validation of structured outputs.
- `LabelsJson(nsfw bool)` — returns a literal JSON **sample** that Ollama-style models can mirror when they only support prompt-enforced structures.

Both helpers build on the same field set (`name`, `confidence`, `topicality`, and optional NSFW flags) so downstream parsing logic (`LabelResult`) can remain engine-agnostic.

### Schema Types & Differences

| Helper                    | Target Engine            | Format                                                 | Validation Style                                                                    | When To Use                                                                                                     |
|---------------------------|--------------------------|--------------------------------------------------------|-------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------|
| `LabelsJsonSchema(false)` | OpenAI (standard labels) | JSON Schema Draft                                      | Strong: OpenAI enforces field types/ranges server-side before returning a response. | When calling GPT‑vision models via `ApiFormatOpenAI` to ensure PhotoPrism receives well-formed label arrays.    |
| `LabelsJsonSchema(true)`  | OpenAI (labels + NSFW)   | JSON Schema Draft with additional boolean/float fields | Strong: same enforcement plus required NSFW fields.                                 | When `DetectNSFWLabels` or NSFW-specific prompts are active and the model must emit `nsfw` + `nsfw_confidence`. |
| `LabelsJson(false)`       | Ollama (standard labels) | Plain JSON example                                     | Soft: model is nudged to mimic the structure through prompt instructions.           | When running self-hosted Ollama models that support “JSON mode” but do not consume JSON Schema definitions.     |
| `LabelsJson(true)`        | Ollama (labels + NSFW)   | Plain JSON example with NSFW keys                      | Soft: prompts describe the required keys; the adapter validates after parsing.      | When Ollama prompts mention NSFW scoring or PhotoPrism sets `DetectNSFWLabels=true`.                            |

**Key technical distinction:** OpenAI’s Responses API accepts a JSON Schema (see `LabelsJsonSchema*`) and guarantees compliance by rejecting invalid responses, while Ollama currently relies on prompt-directed output. For Ollama integrations we provide a representative JSON document (`LabelsJson*`) that models can imitate; PhotoPrism then normalizes and validates the results in Go.

### Field Definitions

- `name` — single-word noun describing the subject (string, required).
- `confidence` — normalized score between `0` and `1` (float, required).
- `topicality` — relative relevance score between `0` and `1` (float, required; defaults to `confidence` if omitted after parsing).
- `nsfw` — boolean flag indicating sensitive content (required only in NSFW variants).
- `nsfw_confidence` — normalized probability for the NSFW assessment (required only in NSFW variants).

OpenAI schemas enforce these ranges/types, while Ollama prompts remind the model to emit matching keys. After parsing, PhotoPrism applies `LabelConfidenceDefault` and `normalizeLabelResult` to fill gaps and enforce naming rules.

### Usage Guidance

1. **OpenAI models** (`Engine: openai`, `RequestFormat: openai`):
   - Leave `Schema` unset in `vision.yml`; the engine defaults call `LabelsJsonSchema(model.PromptContains("nsfw"))`.
   - Optionally override the schema via `Schema`/`SchemaFile` if you extend fields, but keep required keys so `LabelResult` parsing succeeds.
2. **Ollama models** (`Engine: ollama`, `RequestFormat: ollama`):
   - Rely on the built-in samples from `LabelsJson` or include them directly in prompts via `model.SchemaInstructions()`.
   - Because enforcement happens after the response arrives, keep `Format: json` (default) and `Options.ForceJson=true` for label models to make parsing stricter.
3. **Custom engines**:
   - Reuse these helpers to stay compatible with PhotoPrism’s label DTOs.
   - When adding new fields, update both schema/sample versions so OpenAI and Ollama adapters remain aligned.

### References

- JSON Schema primer: https://json-schema.org/learn/miscellaneous-examples  
- OpenAI structured outputs: https://platform.openai.com/docs/guides/structured-outputs  
- JSON mode background (Ollama-style prompts): https://www.alibabacloud.com/help/en/model-studio/json-mode  
- JSON syntax refresher: https://www.json.org/json-en.html
