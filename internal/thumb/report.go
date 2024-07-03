package thumb

import (
	"fmt"
	"sort"

	"github.com/photoprism/photoprism/pkg/txt/report"
)

// Report returns a file format documentation table.
func Report(sizes SizeList, short bool) (rows [][]string, cols []string) {
	if short {
		cols = []string{"Size", "Usage"}
	} else {
		cols = []string{"Name", "Width", "Height", "Aspect Ratio", "Available", "Usage"}
	}

	sorted := append(SizeList{}, sizes...)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Width == sorted[j].Width {
			return sorted[i].Name < sorted[j].Name
		} else {
			return sorted[i].Width < sorted[j].Width
		}
	})

	rows = make([][]string, 0, len(sorted))

	for _, s := range sorted {
		if short {
			rows = append(rows, []string{fmt.Sprintf("%d", s.Width), s.Usage})
		} else {
			aspectRatio := report.Bool(s.Fit, "Preserved", "1:1")

			available := "On-Demand"

			if s.Required {
				available = "Always"
			} else if s.Optional {
				available = "Optional"
			}

			rows = append(rows, []string{s.Name.String(), fmt.Sprintf("%d", s.Width), fmt.Sprintf("%d", s.Height), aspectRatio, available, s.Usage})
		}
	}

	return rows, cols
}
