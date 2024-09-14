package s2

import gs2 "github.com/golang/geo/s2"

// Range returns a token range to find nearby cells within the specified S2 level.
func Range(token string, level int) (start, end string) {
	token = NormalizeToken(token)

	cell := gs2.CellIDFromToken(token)

	if !cell.IsValid() {
		return start, end
	}

	// See https://s2geometry.io/resources/s2cell_statistics.html
	cellLevel := cell.Level()

	// Range level must not be greater than the cell level.
	if level > cellLevel {
		level = cellLevel
	}

	// Get parent cell ID for the given level.
	parentCell := cell.Parent(level)

	// Return computed S2 cell token range.
	return parentCell.Prev().ChildBeginAtLevel(cellLevel).ToToken(), parentCell.Next().ChildBeginAtLevel(cellLevel).ToToken()
}

// PrefixedRange returns a prefixed token range to find nearby cells within the specified S2 level.
func PrefixedRange(token string, level int) (start, end string) {
	start, end = Range(token, level)

	return Prefix(start), Prefix(end)
}
