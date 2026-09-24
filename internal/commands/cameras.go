package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

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
