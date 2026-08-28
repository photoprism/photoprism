package commands

import (
	"fmt"
	"strconv"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// FacesSubjectsCommand reports the people a library holds, with the counts the index stores beside
// the counts its markers currently support.
var FacesSubjectsCommand = &cli.Command{
	Name:   "subjects",
	Usage:  "Lists people with their stored and live counts",
	Flags:  append(report.CliFlags, CountFlag, OffsetFlag),
	Action: facesSubjectsAction,
}

// FacesListCommand reports the face clusters a library holds.
var FacesListCommand = &cli.Command{
	Name:    "ls",
	Aliases: []string{"clusters"},
	Usage:   "Lists face clusters with their samples and current markers",
	Flags:   append(report.CliFlags, CountFlag, OffsetFlag),
	Action:  facesListAction,
}

// FacesMarkersCommand reports face markers, optionally narrowed to one person, one cluster, or one
// of the two shapes a diagnosis keeps returning to.
var FacesMarkersCommand = &cli.Command{
	Name:  "markers",
	Usage: "Lists face markers and what they are assigned to",
	Flags: append(report.CliFlags, CountFlag, OffsetFlag,
		&cli.StringFlag{Name: "subject", Aliases: []string{"s"}, Usage: "only markers of person `UID`"},
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

		people, err := query.SubjectReports(count, offset)

		if err != nil {
			return err
		}

		cols := []string{"Name", "UID", "Src", "Verified", "Files", "Photos", "Live Files", "Live Photos", "Markers"}
		rows := make([][]string, 0, len(people))

		for _, p := range people {
			rows = append(rows, []string{
				p.SubjName, p.SubjUID, entity.SrcString(p.SubjSrc), reportBool(p.Verified),
				strconv.Itoa(p.FileCount), strconv.Itoa(p.PhotoCount),
				strconv.Itoa(p.LiveFiles), strconv.Itoa(p.LivePhotos), strconv.Itoa(p.Markers),
			})
		}

		return printReport(ctx, rows, cols)
	})
}

// facesListAction prints the cluster report.
func facesListAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		count, offset := reportPaging(ctx)

		faces, err := query.FaceReports(count, offset)

		if err != nil {
			return err
		}

		cols := []string{"ID", "Person", "Src", "Kind", "Samples", "Radius", "Collisions", "Collision Radius", "Markers", "Matched At"}
		rows := make([][]string, 0, len(faces))

		for _, f := range faces {
			rows = append(rows, []string{
				f.ID, f.SubjName, entity.SrcString(f.FaceSrc), strconv.Itoa(f.FaceKind),
				strconv.Itoa(f.Samples), report.Distance(f.SampleRadius),
				strconv.Itoa(f.Collisions), report.Distance(f.CollisionRadius),
				strconv.Itoa(f.Markers), report.DateTime(f.MatchedAt),
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
			SubjUID:    clean.UID(ctx.String("subject")),
			FaceID:     clean.Token(ctx.String("face")),
			Unassigned: ctx.Bool("unassigned"),
			Dangling:   ctx.Bool("dangling"),
			Count:      count,
			Offset:     offset,
		})

		if err != nil {
			return err
		}

		cols := []string{"Marker", "File", "Cluster", "Person", "Src", "Name", "Size", "Score", "Dist", "Invalid", "Matched At"}
		rows := make([][]string, 0, len(markers))

		for _, m := range markers {
			rows = append(rows, []string{
				m.MarkerUID, m.FileUID, m.FaceID, m.SubjUID, entity.SrcString(m.SubjSrc), m.MarkerName,
				strconv.Itoa(m.Size), strconv.Itoa(m.Score), report.Distance(m.FaceDist),
				reportBool(m.MarkerInvalid), report.DateTime(m.MatchedAt),
			})
		}

		return printReport(ctx, rows, cols)
	})
}
