package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// NoMakeFlag represents a CLI flag to filter the lenses returned to only those without Make populated.
var NoMakeFlag = &cli.BoolFlag{
	Name:  "nomake",
	Usage: "return only records without Make",
}

// LensesCommand registers the "lenses" CL command.
var LensesCommand = &cli.Command{
	Name:  "lenses",
	Usage: "Lens management subcommands",
	Subcommands: []*cli.Command{
		LensesListCommand,
		LensesUpdateCommand,
	},
}

// LensesListCommand registers the list sub command
var LensesListCommand = &cli.Command{
	Name:      "ls",
	Usage:     "Lists discovered lenses",
	ArgsUsage: "[query]",
	Flags:     append(report.CliFlags, CountFlag, OffsetFlag, NoMakeFlag),
	Action:    lensesListAction,
}

// LensesUpdateCommand registers the update sub command
var LensesUpdateCommand = &cli.Command{
	Name:  "update",
	Usage: "Updates a specific lens Make and Model",
	Flags: []cli.Flag{
		&cli.UintFlag{Name: "id", Usage: "lens id", Required: true},
		&cli.StringFlag{Name: "make", Usage: "the make of the lens", Required: true},
		&cli.StringFlag{Name: "model", Usage: "the model of the lens", Required: true},
	},
	Action: lensesUpdateAction,
}

// lensesListAction searches the database for lenses.
func lensesListAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {

		filter := strings.TrimSpace(strings.Join(ctx.Args().Slice(), " "))
		// Pagination identical to API defaults.
		count := int(ctx.Uint("count")) //nolint:gosec // CLI flag bounded by validation
		if count <= 0 || count > 1000 {
			count = 100
		}
		offset := max(ctx.Int("offset"), 0)

		frm := form.SearchLenses{
			Query:  filter,
			NoMake: ctx.Bool("nomake"),
			Count:  count,
			Offset: offset,
		}

		results, err := search.Lenses(frm)

		if err != nil {
			return err
		}

		format := report.CliFormat(ctx)

		cols := []string{"ID", "Lens Slug", "Lens Name", "Lens Make", "Lens Model", "Updated At"}
		rows := make([][]string, 0, len(results))

		for _, found := range results {
			v := []string{strconv.FormatUint(uint64(found.ID), 10), found.LensSlug, found.LensName, found.LensMake, found.LensModel, found.UpdatedAt.Format("2006-01-02 15:04:05")}
			rows = append(rows, v)
		}

		result, err := report.RenderFormat(rows, cols, format)

		if err != nil {
			return err
		}

		fmt.Println(result)

		return nil
	})
}

// lensesUpdateAction searches the database for lenses.
func lensesUpdateAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {

		lensId := ctx.Uint("id")
		lensMake := ctx.String("make")
		lensModel := ctx.String("model")

		lens := query.FindLensByID(lensId)
		if lens == nil {
			return cli.Exit("lens not found", 1)
		}
		if err := lens.UpdateMakeModel(lensMake, lensModel); err != nil {
			return cli.Exit(err, 1)
		}

		frm := form.SearchLenses{
			ID:     strconv.FormatUint(uint64(lens.ID), 10),
			Count:  10,
			Offset: 0,
		}

		results, err := search.Lenses(frm)

		if err != nil {
			return err
		}

		format := report.CliFormat(ctx)

		cols := []string{"ID", "Lens Slug", "Lens Name", "Lens Make", "Lens Model", "Updated At"}
		rows := make([][]string, 0, len(results))

		for _, found := range results {
			v := []string{strconv.FormatUint(uint64(found.ID), 10), found.LensSlug, found.LensName, found.LensMake, found.LensModel, found.UpdatedAt.Format("2006-01-02 15:04:05")}
			rows = append(rows, v)
		}

		result, err := report.RenderFormat(rows, cols, format)

		if err != nil {
			return err
		}

		fmt.Println(result)

		return nil
	})
}
