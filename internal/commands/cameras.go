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

// CamerasCommand registers the "cameras" CLI command.
var CamerasCommand = &cli.Command{
	Name:  "cameras",
	Usage: "Camera management subcommands",
	Subcommands: []*cli.Command{
		CamerasListCommand,
		CamerasAddCommand,
		CamerasUpdateCommand,
		CamerasRemoveCommand,
	},
}

// CamerasListCommand registers the list sub command.
var CamerasListCommand = &cli.Command{
	Name:      "ls",
	Usage:     "Lists discovered and added cameras",
	ArgsUsage: "[query]",
	Flags:     append(report.CliFlags, CountFlag, OffsetFlag, NoMakeFlag),
	Action:    camerasListAction,
}

// CamerasAddCommand registers the add sub command.
var CamerasAddCommand = &cli.Command{
	Name:  "add",
	Usage: "Adds a camera that can be assigned to pictures without camera metadata",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "make", Usage: "the make of the camera", Required: true},
		&cli.StringFlag{Name: "model", Usage: "the model of the camera", Required: true},
	},
	Action: camerasAddAction,
}

// CamerasRemoveCommand registers the rm sub command.
var CamerasRemoveCommand = &cli.Command{
	Name:  "rm",
	Usage: "Deletes a camera that is no longer needed",
	Flags: []cli.Flag{
		&cli.UintFlag{Name: "id", Usage: "camera id, alternatively pass --make and --model"},
		&cli.StringFlag{Name: "make", Usage: "the make of the camera"},
		&cli.StringFlag{Name: "model", Usage: "the model of the camera"},
		&cli.BoolFlag{Name: "reassign", Usage: "assigns pictures that use the camera to the unknown camera first"},
		YesFlag(),
	},
	Action: camerasRemoveAction,
}

// CamerasUpdateCommand registers the update sub command.
var CamerasUpdateCommand = &cli.Command{
	Name:  "update",
	Usage: "Updates a specific camera Make and Model",
	Flags: []cli.Flag{
		&cli.UintFlag{Name: "id", Usage: "camera id", Required: true},
		&cli.StringFlag{Name: "make", Usage: "the make of the camera", Required: true},
		&cli.StringFlag{Name: "model", Usage: "the model of the camera", Required: true},
	},
	Action: camerasUpdateAction,
}

// camerasListAction searches the database for cameras.
func camerasListAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {

		filter := strings.TrimSpace(strings.Join(ctx.Args().Slice(), " "))
		// Pagination identical to API defaults.
		count := int(ctx.Uint("count")) //nolint:gosec // CLI flag bounded by validation
		if count <= 0 || count > 1000 {
			count = 100
		}
		offset := max(ctx.Int("offset"), 0)

		frm := form.SearchCameras{
			Query:  filter,
			NoMake: ctx.Bool("nomake"),
			Count:  count,
			Offset: offset,
		}

		results, err := search.Cameras(frm)

		if err != nil {
			return err
		}

		format := report.CliFormat(ctx)

		cols := []string{"ID", "Camera Slug", "Camera Name", "Camera Make", "Camera Model", "Updated At"}
		rows := make([][]string, 0, len(results))

		for _, found := range results {
			v := []string{strconv.FormatUint(uint64(found.ID), 10), found.CameraSlug, found.CameraName, found.CameraMake, found.CameraModel, found.UpdatedAt.Format("2006-01-02 15:04:05")}
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

// camerasUpdateAction updates the make and model of a specific camera.
func camerasUpdateAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {

		cameraId := ctx.Uint("id")
		cameraMake := ctx.String("make")
		cameraModel := ctx.String("model")

		if strings.TrimSpace(cameraMake) == "" || strings.TrimSpace(cameraModel) == "" {
			return cli.Exit("make and model must not be empty", 2)
		}

		camera := query.FindCameraByID(cameraId)
		if camera == nil {
			return cli.Exit("camera not found", 3)
		} else if camera.Unknown() {
			return cli.Exit("unknown camera cannot be changed", 2)
		}
		if err := camera.UpdateMakeModel(cameraMake, cameraModel); err != nil {
			return cli.Exit(err, 1)
		}

		return printCamera(ctx, camera.ID)
	})
}

// camerasAddAction adds a camera with the specified make and model, unless it already exists.
func camerasAddAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		cameraMake := ctx.String("make")
		cameraModel := ctx.String("model")

		if strings.TrimSpace(cameraMake) == "" || strings.TrimSpace(cameraModel) == "" {
			return cli.Exit("make and model must not be empty", 2)
		}

		camera, created, err := entity.AddCamera(cameraMake, cameraModel)

		switch {
		case errors.Is(err, entity.ErrInvalidValue):
			return cli.Exit(err, 2)
		case err != nil:
			return cli.Exit(err, 1)
		case created:
			log.Infof("camera %s has been added", camera.String())
		default:
			log.Infof("camera %s already exists", camera.String())
		}

		return printCamera(ctx, camera.ID)
	})
}

// printCamera renders the camera with the specified ID in the format selected by the report flags.
func printCamera(ctx *cli.Context, id uint) error {
	frm := form.SearchCameras{
		ID:     strconv.FormatUint(uint64(id), 10),
		Count:  10,
		Offset: 0,
	}

	results, err := search.Cameras(frm)

	if err != nil {
		return err
	}

	format := report.CliFormat(ctx)

	cols := []string{"ID", "Camera Slug", "Camera Name", "Camera Make", "Camera Model", "Updated At"}
	rows := make([][]string, 0, len(results))

	for _, found := range results {
		v := []string{strconv.FormatUint(uint64(found.ID), 10), found.CameraSlug, found.CameraName, found.CameraMake, found.CameraModel, found.UpdatedAt.Format("2006-01-02 15:04:05")}
		rows = append(rows, v)
	}

	result, err := report.RenderFormat(rows, cols, format)

	if err != nil {
		return err
	}

	fmt.Println(result)

	return nil
}

// camerasRemoveAction deletes a camera, provided no picture uses it or the pictures may be reassigned.
func camerasRemoveAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		reassign := ctx.Bool("reassign")

		id := ctx.Uint("id")
		cameraMake := strings.TrimSpace(ctx.String("make"))
		cameraModel := strings.TrimSpace(ctx.String("model"))

		// Find the camera either by ID or by make and model, since the ID is not shown anywhere else.
		var camera *entity.Camera

		switch {
		case id > 0 && (cameraMake != "" || cameraModel != ""):
			return cli.Exit("pass either --id or --make and --model, not both", 2)
		case id > 0:
			camera = query.FindCameraByID(id)
		case cameraMake != "" && cameraModel != "":
			found := entity.FindCamerasByMakeModel(cameraMake, cameraModel)

			if len(found) > 1 {
				ids := make([]string, len(found))

				for i := range found {
					ids[i] = strconv.FormatUint(uint64(found[i].ID), 10)
				}

				return cli.Exit(fmt.Errorf("found %d cameras with this make and model (IDs %s), pass --id to select one",
					len(found), strings.Join(ids, ", ")), 2)
			} else if len(found) == 1 {
				camera = &found[0]
			}
		default:
			return cli.Exit("pass either --id or --make and --model", 2)
		}

		if camera == nil {
			return cli.Exit("camera not found", 3)
		} else if camera.Unknown() {
			return cli.Exit("unknown camera cannot be deleted", 2)
		}

		count, err := camera.PhotoCount()

		if err != nil {
			return cli.Exit(err, 1)
		} else if count > 0 && !reassign {
			return cli.Exit(fmt.Errorf("camera %s is used by %s, pass --reassign to assign them to the unknown camera",
				camera.String(), english.Plural(count, "picture", "pictures")), 2)
		}

		label := fmt.Sprintf("Delete camera %s with ID %d?", camera.String(), camera.ID)

		if count > 0 {
			label = fmt.Sprintf("Assign %s to the unknown camera and delete camera %s with ID %d?", english.Plural(count, "picture", "pictures"), camera.String(), camera.ID)
		}

		if proceed, confirmErr := ConfirmAction(ctx.Bool("yes"), label); confirmErr != nil {
			return confirmErr
		} else if !proceed {
			log.Infof("camera %s with ID %d was not deleted", camera.String(), camera.ID)
			return nil
		}

		reassigned, err := camera.Delete(reassign)

		if reassigned > 0 {
			log.Infof("assigned %s to the unknown camera", english.Plural(int(reassigned), "picture", "pictures"))
		}

		switch {
		case errors.Is(err, entity.ErrInUse):
			// Pictures may have been assigned to the camera concurrently, e.g. by an index run.
			return cli.Exit(fmt.Errorf("camera %s is still used by pictures, please try again", camera.String()), 2)
		case errors.Is(err, gorm.ErrRecordNotFound):
			return cli.Exit("camera not found", 3)
		case err != nil:
			return cli.Exit(err, 1)
		}

		log.Infof("camera %s with ID %d has been deleted", camera.String(), camera.ID)

		return nil
	})
}
