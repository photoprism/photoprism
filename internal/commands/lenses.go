package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize/english"
	"github.com/jinzhu/gorm"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
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
		LensesAddCommand,
		LensesUpdateCommand,
		LensesRemoveCommand,
	},
}

// LensesListCommand registers the list sub command
var LensesListCommand = &cli.Command{
	Name:      "ls",
	Usage:     "Lists discovered and added lenses",
	ArgsUsage: "[query]",
	Flags:     append(report.CliFlags, CountFlag, OffsetFlag, NoMakeFlag),
	Action:    lensesListAction,
}

// LensesAddCommand registers the add sub command.
var LensesAddCommand = &cli.Command{
	Name:  "add",
	Usage: "Adds a lens that can be assigned to pictures without lens metadata",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "make", Usage: "the make of the lens", Required: true},
		&cli.StringFlag{Name: "model", Usage: "the model of the lens", Required: true},
	},
	Action: lensesAddAction,
}

// LensesRemoveCommand registers the rm sub command.
var LensesRemoveCommand = &cli.Command{
	Name:  "rm",
	Usage: "Deletes a lens that is no longer needed",
	Flags: []cli.Flag{
		&cli.UintFlag{Name: "id", Usage: "lens id, alternatively pass --make and --model"},
		&cli.StringFlag{Name: "make", Usage: "the make of the lens"},
		&cli.StringFlag{Name: "model", Usage: "the model of the lens"},
		&cli.BoolFlag{Name: "reassign", Usage: "assigns pictures that use the lens to the unknown lens first"},
		YesFlag(),
	},
	Action: lensesRemoveAction,
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

		if strings.TrimSpace(lensMake) == "" || strings.TrimSpace(lensModel) == "" {
			return cli.Exit("make and model must not be empty", 2)
		}

		lens := query.FindLensByID(lensId)
		if lens == nil {
			return cli.Exit("lens not found", 3)
		} else if lens.Unknown() {
			return cli.Exit("unknown lens cannot be changed", 2)
		}
		if err := lens.UpdateMakeModel(lensMake, lensModel); err != nil {
			return cli.Exit(err, 1)
		}

		return printLens(ctx, lens.ID)
	})
}

// lensesAddAction adds a lens with the specified make and model, unless it already exists.
func lensesAddAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		lensMake := ctx.String("make")
		lensModel := ctx.String("model")

		if strings.TrimSpace(lensMake) == "" || strings.TrimSpace(lensModel) == "" {
			return cli.Exit("make and model must not be empty", 2)
		}

		lens, created, err := entity.AddLens(lensMake, lensModel)

		switch {
		case errors.Is(err, entity.ErrInvalidValue):
			return cli.Exit(err, 2)
		case err != nil:
			return cli.Exit(err, 1)
		case created:
			log.Infof("lens %s has been added", lens.String())
		default:
			log.Infof("lens %s already exists", lens.String())
		}

		return printLens(ctx, lens.ID)
	})
}

// printLens renders the lens with the specified ID in the format selected by the report flags.
func printLens(ctx *cli.Context, id uint) error {
	frm := form.SearchLenses{
		ID:     strconv.FormatUint(uint64(id), 10),
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
}

// lensesRemoveAction deletes a lens, provided no picture uses it or the pictures may be reassigned.
func lensesRemoveAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		reassign := ctx.Bool("reassign")

		id := ctx.Uint("id")
		lensMake := strings.TrimSpace(ctx.String("make"))
		lensModel := strings.TrimSpace(ctx.String("model"))

		// Find the lens either by ID or by make and model, since the ID is not shown anywhere else.
		var lens *entity.Lens

		switch {
		case id > 0 && (lensMake != "" || lensModel != ""):
			return cli.Exit("pass either --id or --make and --model, not both", 2)
		case id > 0:
			lens = query.FindLensByID(id)
		case lensMake != "" && lensModel != "":
			found := entity.FindLensesByMakeModel(lensMake, lensModel)

			if len(found) > 1 {
				ids := make([]string, len(found))

				for i := range found {
					ids[i] = strconv.FormatUint(uint64(found[i].ID), 10)
				}

				return cli.Exit(fmt.Errorf("found %d lenses with this make and model (IDs %s), pass --id to select one",
					len(found), strings.Join(ids, ", ")), 2)
			} else if len(found) == 1 {
				lens = &found[0]
			}
		default:
			return cli.Exit("pass either --id or --make and --model", 2)
		}

		if lens == nil {
			return cli.Exit("lens not found", 3)
		} else if lens.Unknown() {
			return cli.Exit("unknown lens cannot be deleted", 2)
		}

		count, err := lens.PhotoCount()

		if err != nil {
			return cli.Exit(err, 1)
		} else if count > 0 && !reassign {
			return cli.Exit(fmt.Errorf("lens %s is used by %s, pass --reassign to assign them to the unknown lens",
				lens.String(), english.Plural(count, "picture", "pictures")), 2)
		}

		label := fmt.Sprintf("Delete lens %s with ID %d?", lens.String(), lens.ID)

		if count > 0 {
			label = fmt.Sprintf("Assign %s to the unknown lens and delete lens %s with ID %d?", english.Plural(count, "picture", "pictures"), lens.String(), lens.ID)
		}

		if proceed, confirmErr := ConfirmAction(ctx.Bool("yes"), label); confirmErr != nil {
			return confirmErr
		} else if !proceed {
			log.Infof("lens %s with ID %d was not deleted", lens.String(), lens.ID)
			return nil
		}

		reassigned, err := lens.Delete(reassign)

		if reassigned > 0 {
			log.Infof("assigned %s to the unknown lens", english.Plural(int(reassigned), "picture", "pictures"))
		}

		switch {
		case errors.Is(err, entity.ErrInUse):
			// Pictures may have been assigned to the lens concurrently, e.g. by an index run.
			return cli.Exit(fmt.Errorf("lens %s is still used by pictures, please try again", lens.String()), 2)
		case errors.Is(err, gorm.ErrRecordNotFound):
			return cli.Exit("lens not found", 3)
		case err != nil:
			return cli.Exit(err, 1)
		}

		log.Infof("lens %s with ID %d has been deleted", lens.String(), lens.ID)

		return nil
	})
}
