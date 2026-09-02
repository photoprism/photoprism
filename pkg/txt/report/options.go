package report

// Align names how a column's cells are positioned. Declared here rather than reusing the table
// writer's own type, so callers do not take a dependency on it to ask for a right-aligned column.
type Align string

const (
	// AlignDefault leaves the format's own default in place.
	AlignDefault Align = ""
	// AlignLeft positions cells at the start of the column.
	AlignLeft Align = "left"
	// AlignRight positions cells at the end of the column, which is what makes a column of numbers
	// comparable down the page.
	AlignRight Align = "right"
	// AlignCenter positions cells in the middle of the column.
	AlignCenter Align = "center"
)

// Options represents render options.
type Options struct {
	Format  Format
	Caption string
	Valid   bool
	NoWrap  bool
	// Align sets the alignment per column, indexed as the columns are. Shorter than the columns is
	// allowed and leaves the rest at the default; it is ignored by the formats that carry no layout.
	Align []Align
	// Keys names the JSON fields, indexed as the columns are. Without it the column headings are
	// canonicalized into keys, which ties a machine-readable contract to text written to be read:
	// retitling a column renames its key, and a heading like "in %" canonicalizes to "in".
	Keys []string
}
