package vision

import (
	"strings"
)

// FilterModels takes a list of model type names and a scheduling context, and
// returns only the types that are allowed to run according to the supplied
// predicate. Empty or unknown names are ignored.
func FilterModels(models []string, when RunType, allow func(ModelType, RunType) bool) []string {
	if len(models) == 0 {
		return models
	}

	filtered := make([]string, 0, len(models))

	for _, name := range models {
		modelType := ModelType(strings.TrimSpace(name))
		if modelType == "" {
			continue
		}

		if allow == nil || allow(modelType, when) {
			filtered = append(filtered, string(modelType))
		}
	}

	return filtered
}
