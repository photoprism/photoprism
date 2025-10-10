package schema

// LabelsDefault provides the minimal JSON schema for label responses used across engines.
const (
	LabelsDefault = "{\n  \"labels\": [{\n    \"name\": \"\",\n    \"confidence\": 0,\n    \"topicality\": 0 }]\n}"
	LabelsNSFW    = "{\n  \"labels\": [{\n    \"name\": \"\",\n    \"confidence\": 0,\n    \"topicality\": 0,\n    \"nsfw\": false,\n    \"nsfw_confidence\": 0\n  }]\n}"
)

// Labels returns the canonical label schema string.
func Labels(nsfw bool) string {
	if nsfw {
		return LabelsNSFW
	} else {
		return LabelsDefault
	}
}
