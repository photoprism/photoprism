package commands

import (
	"fmt"
	"strconv"
	"strings"

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

// reportPerson returns the person a report command was narrowed to, which is a subject uid or a
// name fragment. Taken as an argument rather than a flag so the reports read like the other list
// commands, and so a person can be inspected without piping the output through grep.
func reportPerson(ctx *cli.Context) string {
	return clean.SearchString(strings.Join(ctx.Args().Slice(), " "))
}

// reportPaging returns the count and offset a report command was given, bounded like the API.
func reportPaging(ctx *cli.Context) (count, offset int) {
	count = int(ctx.Uint("count")) //nolint:gosec // CLI flag bounded here

	if count <= 0 || count > 1000 {
		count = 100
	}

	return count, max(ctx.Int("offset"), 0)
}

// printReport renders rows in the format the flags selected.
func printReport(ctx *cli.Context, rows [][]string, cols []string) error {
	result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

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

		cols := []string{"UID", "Name", "Src", "Markers", "Verified", "Hidden", "Files", "Photos"}
		rows := make([][]string, 0, len(people))

		for _, p := range people {
			rows = append(rows, []string{
				p.SubjUID, p.SubjName, entity.SrcString(p.SubjSrc), strconv.Itoa(p.Markers),
				reportBool(p.Verified), reportBool(p.SubjHidden),
				strconv.Itoa(p.FileCount), strconv.Itoa(p.PhotoCount),
			})
		}

		return printReport(ctx, rows, cols)
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

		cols := []string{"ID", "Name", "Subject", "Src", "Markers", "Samples", "Radius", "Collisions", "Collision Radius", "Kind", "Matched At"}
		rows := make([][]string, 0, len(faces))

		for _, f := range faces {
			rows = append(rows, []string{
				f.ID, f.SubjName, f.SubjUID, entity.SrcString(f.FaceSrc),
				strconv.Itoa(f.Markers), strconv.Itoa(f.Samples), report.Distance(f.SampleRadius),
				strconv.Itoa(f.Collisions), report.Distance(f.CollisionRadius),
				face.Kind(f.FaceKind).String(), report.DateTime(f.MatchedAt),
			})
		}

		return printReport(ctx, rows, cols)
	})
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

		cols := []string{"Marker", "Name", "Size", "Score", "Subject", "Src", "Face", "Dist", "Embedding", "Landmarks", "Invalid", "File", "Matched At"}
		rows := make([][]string, 0, len(markers))

		for _, m := range markers {
			rows = append(rows, []string{
				m.MarkerUID, m.MarkerName, strconv.Itoa(m.Size), strconv.Itoa(m.Score),
				m.SubjUID, entity.SrcString(m.SubjSrc), m.FaceID, report.Distance(m.FaceDist),
				reportVectors(m.EmbeddingDims), reportVectors(m.Landmarks),
				reportBool(m.MarkerInvalid), m.FileUID, report.DateTime(m.MatchedAt),
			})
		}

		return printReport(ctx, rows, cols)
	})
}
