package schema

import (
	"encoding/json"
)

// LabelsJsonSchemaDefault provides the minimal JSON schema for label responses used across engines.
const (
	LabelsJsonSchemaDefault = `{
  "type": "object",
  "properties": {
    "labels": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string",
            "minLength": 1
          },
          "confidence": {
            "type": "number",
            "minimum": 0,
            "maximum": 1
          },
          "topicality": {
            "type": "number",
            "minimum": 0,
            "maximum": 1
          }
        },
        "required": ["name", "confidence", "topicality"],
        "additionalProperties": false
      },
      "default": []
    }
  },
  "required": ["labels"],
  "additionalProperties": false
}`
	LabelsJsonDefault    = "{\n  \"labels\": [{\n    \"name\": \"\",\n    \"confidence\": 0,\n    \"topicality\": 0 }]\n}"
	LabelsJsonSchemaNSFW = `{
  "type": "object",
  "properties": {
    "labels": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string",
            "minLength": 1
          },
          "confidence": {
            "type": "number",
            "minimum": 0,
            "maximum": 1
          },
          "topicality": {
            "type": "number",
            "minimum": 0,
            "maximum": 1
          },
          "nsfw": {
            "type": "boolean"
          },
          "nsfw_confidence": {
            "type": "number",
            "minimum": 0,
            "maximum": 1
          }
        },
        "required": [
          "name",
          "confidence",
          "topicality",
          "nsfw",
          "nsfw_confidence"
        ],
        "additionalProperties": false
      },
      "default": []
    }
  },
  "required": ["labels"],
  "additionalProperties": false
}`
	LabelsJsonNSFW = "{\n  \"labels\": [{\n    \"name\": \"\",\n    \"confidence\": 0,\n    \"topicality\": 0,\n    \"nsfw\": false,\n    \"nsfw_confidence\": 0\n  }]\n}"
)

// LabelsJsonSchema returns the canonical label JSON Schema string for OpenAI API endpoints.
//
// Related documentation and references:
// - https://platform.openai.com/docs/guides/structured-outputs
// - https://json-schema.org/learn/miscellaneous-examples
func LabelsJsonSchema(nsfw bool) json.RawMessage {
	if nsfw {
		return json.RawMessage(LabelsJsonSchemaNSFW)
	} else {
		return json.RawMessage(LabelsJsonSchemaDefault)
	}
}

// LabelsJson returns the canonical label JSON string for Ollama vision models.
//
// Related documentation and references:
// - https://www.alibabacloud.com/help/en/model-studio/json-mode
// - https://www.json.org/json-en.html
func LabelsJson(nsfw bool) string {
	if nsfw {
		return LabelsJsonNSFW
	} else {
		return LabelsJsonDefault
	}
}
