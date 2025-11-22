package commands

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/txt"
)

var (
	visionSourceNames []string
	visionSourcesOnce sync.Once
)

func initVisionSources() {
	visionSourcesOnce.Do(func() {
		namesSet := make(map[string]struct{}, len(entity.SrcVisionCommands))

		for alias := range entity.SrcVisionCommands {
			normalized := strings.TrimSpace(alias)

			if normalized == "" {
				continue
			}

			if _, ok := namesSet[normalized]; ok {
				continue
			}
			namesSet[normalized] = struct{}{}
			visionSourceNames = append(visionSourceNames, normalized)
		}

		sort.Strings(visionSourceNames)
	})
}

func sanitizeVisionSource(raw string) (entity.Src, error) {
	initVisionSources()

	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return entity.SrcAuto, nil
	}

	if src, ok := entity.SrcVisionCommands[value]; ok {
		return src, nil
	}

	allowed := append([]string(nil), visionSourceNames...)
	return "", fmt.Errorf("vision: unsupported source %q (allowed: %s)", raw, txt.JoinAnd(allowed))
}

func visionSourceUsage() string {
	initVisionSources()
	return strings.Join(visionSourceNames, ", ")
}
