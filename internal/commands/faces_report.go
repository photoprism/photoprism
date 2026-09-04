package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// FacesSubjectsCommand reports the people a library holds, with the counts the index stores beside
// the counts its markers currently support.
var FacesSubjectsCommand = &cli.Command{
	Name:      "subjects",
	Usage:     "Lists people with the files and photos their markers support",
	ArgsUsage: "[name|uid]",
	Flags: append(report.CliFlags, CountFlag, OffsetFlag,
		&cli.BoolFlag{Name: "stored", Usage: "report the counts the index holds instead of counting now"},
	),
	Action: facesSubjectsAction,
}

// FacesListCommand reports the face clusters a library holds.
var FacesListCommand = &cli.Command{
	Name:      "ls",
	Aliases:   []string{"clusters"},
	Usage:     "Lists face clusters with their samples and current markers",
	ArgsUsage: "[name|uid]",
	Flags:     append(report.CliFlags, CountFlag, OffsetFlag),
	Action:    facesListAction,
}

// FacesMarkersCommand reports face markers, optionally narrowed to one person, one cluster, or one
// of the two shapes a diagnosis keeps returning to.
var FacesMarkersCommand = &cli.Command{
	Name:      "markers",
	Usage:     "Lists face markers and what they are assigned to",
	ArgsUsage: "[name|uid]",
	Flags: append(report.CliFlags, CountFlag, OffsetFlag,
		&cli.StringFlag{Name: "face", Aliases: []string{"f"}, Usage: "only markers of cluster `ID`"},
		&cli.BoolFlag{Name: "unassigned", Usage: "only markers that have a person but no cluster"},
		&cli.BoolFlag{Name: "dangling", Usage: "only markers whose cluster no longer exists"},
	),
	Action: facesMarkersAction,
}

// FacesConflictsCommand reports the cluster pairs that hold the same face while naming different
// people, which is the population collision resolution acts on.
var FacesConflictsCommand = &cli.Command{
	Name:      "conflicts",
	Usage:     "Lists face clusters that hold the same face but not the same person",
	ArgsUsage: "[name|uid]",
	Description: "Compares every eligible cluster pair, so --count limits the output and not the scan. " +
		"Pass a name or subject UID to compare one person's clusters instead.",
	Flags:  append(report.CliFlags, CountFlag, OffsetFlag),
	Action: facesConflictsAction,
}

// reportBool renders a flag with the shared Yes/No labels every other report column uses.
func reportBool(b bool) string {
	return report.Bool(b, report.Yes, report.No)
}

// reportVectors renders how much of a stored vector a marker holds. The width answers more than a
// yes would: a marker embedded by another model reports a different one, which is why the number is
// shown rather than a flag.
func reportVectors(n int) string {
	switch {
	case n == query.InvalidJSON:
		return "invalid"
	case n <= 0:
		return ""
	default:
		return strconv.Itoa(n)
	}
}

// reportUnrecorded marks a column the row carries no measurement for, which is not the same as a
// measured zero: the size bars fall back to another value where it appears.
const reportUnrecorded = "-"

// reportFrameShare renders how much of the frame's width a marker area covers, as a percentage
// without the sign, which the column heading carries.
//
// Relative because the alternative invites a misreading: the stored size is in pixels of the
// Fit720 detection thumbnail, and nothing in a table of numbers says so.
func reportFrameShare(w float32) string {
	if w <= 0 {
		return reportUnrecorded
	}

	return strconv.FormatFloat(float64(w)*100, 'f', 1, 64)
}

// reportThumbSize renders the extent an embedding was sampled at, in pixels of the image it was
// drawn from. Empty where none was recorded, which is what the clustering bar falls back for.
func reportThumbSize(px int) string {
	if px < 1 {
		return reportUnrecorded
	}

	return strconv.Itoa(px)
}

// reportThumbPixels renders the extent an embedding was sampled at with its unit, so the column is
// not read as the frame share beside it.
func reportThumbPixels(px int) string {
	if px < 1 {
		return reportUnrecorded
	}

	return strconv.Itoa(px) + "px"
}

// reportEmbedding renders the vector a marker holds as the model that produced it and its width.
//
// The model is named because a distance only means something within one embedding space, and a
// library holds markers from more than one.
func reportEmbedding(model string, dims int) string {
	switch {
	case dims == query.InvalidJSON:
		return "invalid"
	case dims <= 0:
		return ""
	case model == "":
		return strconv.Itoa(dims)
	default:
		return model + " " + strconv.Itoa(dims)
	}
}

// reportModelName renders the model a row records, marking an unrecorded one rather than leaving the
// cell blank: a blank reads as "none" where it means the row predates the column.
func reportModelName(model string) string {
	if model == "" {
		return reportUnrecorded
	}

	return model
}

// reportPerson returns the person a report command was narrowed to: a subject uid or a name fragment.
//
// Sanitized by clean.Name, matching what Subject.SetName writes. Not clean.SearchString, which
// turns "%" into "*" and would search for a character no name can hold.
func reportPerson(ctx *cli.Context) string {
	return clean.Name(strings.Join(ctx.Args().Slice(), " "))
}

// reportPaging returns the count and offset a report command was given, bounded like the API.
func reportPaging(ctx *cli.Context) (count, offset int) {
	count = int(ctx.Uint("count")) //nolint:gosec // CLI flag bounded here

	if count <= 0 || count > 1000 {
		count = 100
	}

	return count, max(ctx.Int("offset"), 0)
}

// rightAligned returns the column alignment for a report, right-aligning the columns named.
//
// By name rather than by position, so reordering the columns cannot leave the alignment pointing at
// whatever ends up at that index.
func rightAligned(cols []string, names ...string) []report.Align {
	want := make(map[string]bool, len(names))

	for _, n := range names {
		want[n] = true
	}

	out := make([]report.Align, len(cols))

	for i, c := range cols {
		if want[c] {
			out[i] = report.AlignRight
		}
	}

	return out
}

// printReport renders rows with the alignment and the JSON field names the caller asked for.
//
// The field names are given rather than derived, so retitling a column for readability cannot rename
// a key that something is already reading.
func printReport(ctx *cli.Context, rows [][]string, cols []string, align []report.Align, keys []string) error {
	result, err := report.RenderFormatOptions(rows, cols, report.CliFormat(ctx),
		report.Options{Align: align, Keys: keys})

	if err != nil {
		return err
	}

	fmt.Println(result)

	return nil
}

// facesSubjectsAction prints the people report.
func facesSubjectsAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		count, offset := reportPaging(ctx)

		people, err := query.SubjectReports(reportPerson(ctx), count, offset, !ctx.Bool("stored"))

		if err != nil {
			return err
		}

		// The counts run smallest to largest, which is also how they nest: a person holds few
		// clusters, a photo can hold several files, and one face can be marked more than once in a
		// file - a mirror puts the same person in it twice.
		cols := []string{"Subject", "Name", "Src", "Birth Date", "Favorite", "Verified", "Hidden", "Private", "Clusters", "Photos", "Files", "Markers", "Created At"}

		keys := []string{
			"subj_uid", "subj_name", "subj_src", "subj_birthday", "subj_favorite", "verified",
			"subj_hidden", "subj_private", "clusters", "photo_count", "file_count", "markers", "created_at",
		}

		align := rightAligned(cols, "Clusters", "Photos", "Files", "Markers")
		rows := make([][]string, 0, len(people))

		for _, p := range people {
			rows = append(rows, []string{
				p.SubjUID, p.SubjName, entity.SrcString(p.SubjSrc), report.Date(p.SubjBirthday),
				reportBool(p.SubjFavorite), reportBool(p.Verified), reportBool(p.SubjHidden),
				reportBool(p.SubjPrivate),
				strconv.Itoa(p.Clusters), strconv.Itoa(p.PhotoCount),
				strconv.Itoa(p.FileCount), strconv.Itoa(p.Markers),
				report.DateTime(&p.CreatedAt),
			})
		}

		return printReport(ctx, rows, cols, align, keys)
	})
}

// facesListAction prints the cluster report.
func facesListAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		count, offset := reportPaging(ctx)

		faces, err := query.FaceReports(reportPerson(ctx), count, offset)

		if err != nil {
			return err
		}

		// Samples first and Markers last: the first is what the cluster was built from and the second
		// what it holds now, and the two drift.
		cols := []string{"Face", "Name", "Subject", "Src", "Kind", "Embedding", "Samples", "Radius", "Collisions", "Collision Radius", "Markers", "Matched At"}

		keys := []string{
			"face_id", "subj_name", "subj_uid", "face_src", "face_kind", "embedding",
			"samples", "sample_radius", "collisions", "collision_radius", "markers", "matched_at",
		}

		align := rightAligned(cols, "Samples", "Radius", "Collisions", "Collision Radius", "Markers")
		rows := make([][]string, 0, len(faces))

		for _, f := range faces {
			rows = append(rows, []string{
				f.ID, f.SubjName, f.SubjUID, entity.SrcString(f.FaceSrc), face.Kind(f.FaceKind).String(),
				reportEmbedding(f.EmbedModel, f.EmbeddingDims),
				strconv.Itoa(f.Samples), report.Distance(f.SampleRadius),
				strconv.Itoa(f.Collisions), report.Distance(f.CollisionRadius),
				strconv.Itoa(f.Markers), report.DateTime(f.MatchedAt),
			})
		}

		return printReport(ctx, rows, cols, align, keys)
	})
}

// reportResolution names what resolving a conflict would do to the first cluster, so a row an
// operator has to act on is legible without knowing where AmbiguityDist sits.
//
// ResolveCollision ignores a cluster that names nobody, so such a pair is reported but acted on by
// nothing - which the column has to say rather than assert an outcome that cannot happen.
func reportResolution(c query.FaceConflict) string {
	switch {
	case c.SubjUID == "":
		return "none"
	case c.Ambiguous():
		return "ambiguous"
	case c.Narrows():
		return "narrow"
	default:
		// Past AmbiguityDist but too close for the recorded radius to clear CollisionDist, so
		// resolution writes a number nothing enforces.
		return "inert"
	}
}

// faceConflictCols returns the conflict report columns.
func faceConflictCols() []string {
	return []string{"Face", "Name", "Subject", "Face 2", "Name 2", "Subject 2", "Resolution",
		"Dist", "Accept", "Accept 2", "Samples", "Samples 2"}
}

// faceConflictRows renders the reported pairs in the order faceConflictCols names.
func faceConflictRows(conflicts []query.FaceConflict) [][]string {
	rows := make([][]string, 0, len(conflicts))

	for _, c := range conflicts {
		rows = append(rows, []string{
			c.ID, c.SubjName, c.SubjUID,
			c.OtherID, c.OtherSubjName, c.OtherSubjUID,
			reportResolution(c), report.Distance(c.Dist),
			report.Distance(c.Accept), report.Distance(c.OtherAccept),
			strconv.Itoa(c.Samples), strconv.Itoa(c.OtherSamples),
		})
	}

	return rows
}

// facesConflictsAction prints the cluster conflict report.
func facesConflictsAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		format, err := report.CliFormatStrict(ctx)

		if err != nil {
			return err
		}

		count, offset := reportPaging(ctx)

		conflicts, scan, err := query.FaceConflicts(reportPerson(ctx), count, offset)

		if err != nil {
			return err
		}

		notes, err := query.FaceConflictReportNotes()

		if err != nil {
			return err
		}

		cols := faceConflictCols()
		rows := faceConflictRows(conflicts)
		lines := faceConflictNotes(scan, notes, unresolvedConflicts(conflicts))

		if format == report.JSON {
			return printFaceConflictsJSON(rows, cols, scan, lines)
		}

		result, renderErr := report.Render(rows, cols, report.Options{Format: format})

		if renderErr != nil {
			return renderErr
		}

		fmt.Println(result)

		notesOut := faceConflictNoteWriter(format)

		for _, line := range lines {
			fmt.Fprintf(notesOut, "%s\n", line)
		}

		fmt.Fprintln(notesOut)

		return nil
	})
}

// faceConflictNotes returns what the table itself cannot show: what the pass compared, the
// threshold the Resolution column turns on, and the clusters a reader would otherwise assume had
// been compared.
func faceConflictNotes(scan query.FaceConflictScan, notes query.FaceConflictNotes, unresolved int) []string {
	lines := []string{
		fmt.Sprintf("Compared %s across %s.",
			english.Plural(scan.Compared, "pair", "pairs"), english.Plural(scan.Clusters, "cluster", "clusters")),
		fmt.Sprintf("Resolving below %s retires a cluster as ambiguous and above %s narrows it; in between it records a radius the matcher ignores.",
			report.Distance(face.AmbiguityDist()), report.Distance(face.CollisionDist+face.Epsilon)),
	}

	if unresolved > 0 {
		lines = append(lines, fmt.Sprintf("%s resolved as none, pairing a named cluster with an unnamed one, "+
			"which nothing resolves: an unnamed cluster is not evidence of a second person. Name it if it is the same person.",
			english.Plural(unresolved, "row is", "rows are")))
	}

	if notes.Ambiguous > 0 {
		lines = append(lines, fmt.Sprintf("%s already retired as ambiguous and not compared.",
			english.Plural(notes.Ambiguous, "cluster in the index is", "clusters in the index are")))
	}

	if notes.Hidden > 0 {
		lines = append(lines, fmt.Sprintf("%s hidden and not compared.",
			english.Plural(notes.Hidden, "cluster in the index is", "clusters in the index are")))
	}

	if notes.InertRadius > 0 {
		lines = append(lines, fmt.Sprintf("%s a collision radius at or below the collision distance of %s, which the matcher does not enforce.",
			english.Plural(notes.InertRadius, "compared cluster records", "compared clusters record"), report.Distance(face.CollisionDist)))
	}

	if notes.BelowOwnSpread > 0 {
		lines = append(lines, fmt.Sprintf("%s a collision radius inside the sample radius, so they now accept fewer faces than they were built from.",
			english.Plural(notes.BelowOwnSpread, "compared cluster records", "compared clusters record")))
	}

	return lines
}

// unresolvedConflicts counts reported pairs the resolver will not act on, which is every row whose
// reported side names nobody. The other side always names somebody, since two anonymous clusters
// never pair.
func unresolvedConflicts(conflicts []query.FaceConflict) (n int) {
	for _, c := range conflicts {
		if c.SubjUID == "" {
			n++
		}
	}

	return n
}

// faceConflictNoteWriter returns where the notes belong for a format: stderr for the delimited ones,
// so redirecting stdout to a file leaves a parseable table and an interactive run still shows them.
func faceConflictNoteWriter(format report.Format) io.Writer {
	switch format {
	case report.CSV, report.TSV:
		return os.Stderr
	default:
		return os.Stdout
	}
}

// printFaceConflictsJSON writes the report as a single object, so a saved report carries the notes
// instead of losing everything the table could not show.
func printFaceConflictsJSON(rows [][]string, cols []string, scan query.FaceConflictScan, notes []string) error {
	b, err := json.Marshal(map[string]any{
		"conflicts": report.RowsToObjects(rows, cols),
		"scan":      map[string]int{"clusters": scan.Clusters, "compared": scan.Compared},
		"notes":     notes,
	})

	if err != nil {
		return err
	}

	fmt.Println(string(b))

	return nil
}

// facesMarkersAction prints the marker report.
func facesMarkersAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		count, offset := reportPaging(ctx)

		markers, err := query.MarkerReports(query.MarkerReportFilter{
			Person:     reportPerson(ctx),
			FaceID:     clean.Token(ctx.String("face")),
			Unassigned: ctx.Bool("unassigned"),
			Dangling:   ctx.Bool("dangling"),
			Count:      count,
			Offset:     offset,
		})

		if err != nil {
			return err
		}

		// Embedding and Landmarks sit together as the two vectors a marker stores, with the detector
		// that produced the crop after them.
		cols := []string{"Marker", "Src", "Size", "in %", "Score", "Name", "Subject", "Src", "Face", "Dist", "Invalid", "Embedding", "Landmarks", "Detector", "File", "Matched At"}

		align := rightAligned(cols, "Size", "in %", "Score", "Dist", "Landmarks")

		keys := []string{
			"marker_uid", "marker_src", "thumb_size", "frame_share", "score",
			"marker_name", "subj_uid", "subj_src", "face_id", "face_dist", "marker_invalid",
			"embedding", "landmarks", "detect_model", "file_uid", "matched_at",
		}

		rows := make([][]string, 0, len(markers))

		for _, m := range markers {
			rows = append(rows, []string{
				m.MarkerUID, entity.SrcString(m.MarkerSrc),
				reportThumbPixels(m.ThumbSize), reportFrameShare(m.W), strconv.Itoa(m.Score),
				m.MarkerName, m.SubjUID, entity.SrcString(m.SubjSrc),
				m.FaceID, report.Distance(m.FaceDist), reportBool(m.MarkerInvalid),
				reportEmbedding(m.EmbedModel, m.EmbeddingDims), reportVectors(m.Landmarks),
				reportModelName(m.DetectModel), m.FileUID, report.DateTime(m.MatchedAt),
			})
		}

		return printReport(ctx, rows, cols, align, keys)
	})
}
