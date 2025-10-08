package clean

import "strings"

// JSON attempts to extract a JSON object or array from raw text.
// It removes common wrappers such as Markdown code fences and trailing commentary.
// Returns an empty string when no JSON payload can be found.
func JSON(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)

		if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
			if idx := strings.Index(trimmed, "\n"); idx != -1 {
				trimmed = trimmed[idx+1:]
			} else {
				return ""
			}
		}

		if idx := strings.LastIndex(trimmed, "```"); idx != -1 {
			trimmed = trimmed[:idx]
		}
	}

	trimmed = strings.TrimSpace(trimmed)

	startObj := strings.Index(trimmed, "{")
	startArr := strings.Index(trimmed, "[")

	start := -1
	if startObj >= 0 && startArr >= 0 {
		if startObj < startArr {
			start = startObj
		} else {
			start = startArr
		}
	} else if startObj >= 0 {
		start = startObj
	} else if startArr >= 0 {
		start = startArr
	}

	endObj := strings.LastIndex(trimmed, "}")
	endArr := strings.LastIndex(trimmed, "]")

	end := -1
	if endObj >= 0 && endArr >= 0 {
		if endObj > endArr {
			end = endObj
		} else {
			end = endArr
		}
	} else if endObj >= 0 {
		end = endObj
	} else if endArr >= 0 {
		end = endArr
	}

	if start >= 0 && end > start {
		trimmed = trimmed[start : end+1]
	}

	return strings.TrimSpace(trimmed)
}
