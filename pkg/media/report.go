package media

import (
	"sort"
	"strings"
	"unicode"

	"github.com/photoprism/photoprism/pkg/fs"
)

// Report returns a file format documentation table.
func Report(m fs.TypesExt, withDesc, withType, withExt bool) (rows [][]string, cols []string) {
	cols = make([]string, 0, 4)
	cols = append(cols, "Format")

	t := 0

	if withDesc {
		cols = append(cols, "Description")
	}

	if withType {
		if withDesc {
			t = 2
		} else {
			t = 1
		}

		cols = append(cols, "Type")
	}

	if withExt {
		cols = append(cols, "Extensions")
	}

	rows = make([][]string, 0, len(m))

	ucFirst := func(str string) string {
		for i, v := range str {
			return string(unicode.ToUpper(v)) + str[i+1:]
		}
		return ""
	}

	for f, ext := range m {
		sort.Slice(ext, func(i, j int) bool {
			return ext[i] < ext[j]
		})

		v := make([]string, 0, 4)
		v = append(v, strings.ToUpper(f.String()))

		if withDesc {
			v = append(v, fs.TypeInfo[f])
		}

		if withType {
			v = append(v, ucFirst(string(Formats[f])))
		}

		if withExt {
			v = append(v, strings.Join(ext, ", "))
		}

		rows = append(rows, v)
	}

	sort.Slice(rows, func(i, j int) bool {
		if t > 0 && rows[i][t] == rows[j][t] {
			return rows[i][0] < rows[j][0]
		} else {
			return rows[i][t] < rows[j][t]
		}
	})

	return rows, cols
}
