package mcp

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// getMediaAccess is the declaration of get_media.
var getMediaAccess = ToolAccess{Resource: acl.ResourcePhotos, Action: acl.ActionView}

// GetMediaInput defines the inputs of the get_media tool.
type GetMediaInput struct {
	UID string `json:"uid" jsonschema:"media UID from search_media"`
}

// MediaDetails is a media with all its details, as returned by get_media.
type MediaDetails struct {
	Media
	Width    int         `json:"width,omitempty"`
	Height   int         `json:"height,omitempty"`
	Duration float64     `json:"duration,omitempty"`
	Location *Location   `json:"location,omitempty"`
	Keywords []string    `json:"keywords,omitempty"`
	Files    []MediaFile `json:"files"`
	Quality  int         `json:"quality"`
	Private  bool        `json:"private"`
	Archived bool        `json:"archived"`
	Review   bool        `json:"review"`
}

// Location contains the coordinates of a media.
type Location struct {
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Altitude int     `json:"altitude,omitempty"`
}

// MediaFile is a file of a media, named without its path.
type MediaFile struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Size    int64  `json:"size"`
	Primary bool   `json:"primary"`
}

// errMediaNotFound is returned for a UID that does not exist or that the session cannot see.
func errMediaNotFound(uid string) error {
	return fmt.Errorf("media %s not found; call search_media to find media", uid)
}

// getMedia returns a media with its details, like GET /api/v1/photos/:uid for the same session.
func getMedia(_ context.Context, req *sdkmcp.CallToolRequest, input GetMediaInput) (*sdkmcp.CallToolResult, MediaDetails, error) {
	s := RequestSession(req)

	if s == nil {
		return nil, MediaDetails{}, errors.New("get_media requires an authenticated session")
	}

	uid := strings.ToLower(strings.TrimSpace(input.UID))

	if !rnd.IsUID(uid, entity.PhotoUID) {
		return nil, MediaDetails{}, fmt.Errorf("uid must be a media UID; call search_media to find media")
	}

	if visible, err := search.PhotoVisibleToSession(uid, s); err != nil || !visible {
		return nil, MediaDetails{}, errMediaNotFound(uid)
	}

	p, err := query.PhotoPreloadByUID(uid)

	if err != nil {
		return nil, MediaDetails{}, errMediaNotFound(uid)
	}

	p.RedactForSession(s)

	return nil, newMediaDetails(&p, s), nil
}

// newMediaDetails returns the details of a photo that has been preloaded and redacted for the session.
func newMediaDetails(p *entity.Photo, s *entity.Session) MediaDetails {
	row := search.Photo{
		ID:            p.ID,
		PhotoUID:      p.PhotoUID,
		PhotoType:     p.PhotoType,
		PhotoTitle:    p.PhotoTitle,
		PhotoCaption:  p.PhotoCaption,
		TakenAt:       p.TakenAt,
		TakenAtLocal:  p.TakenAtLocal,
		PhotoFavorite: p.PhotoFavorite,
		PhotoCountry:  p.PhotoCountry,
		CameraID:      p.CameraID,
		LensID:        p.LensID,
	}

	if p.Place != nil {
		row.PlaceID = p.Place.ID
		row.PlaceLabel = p.Place.PlaceLabel
		row.PlaceCountry = p.Place.PlaceCountry
	}

	items := []Media{newMedia(row)}
	addMediaDetails(items, search.PhotoResults{row}, s)

	result := MediaDetails{
		Media:    items[0],
		Files:    []MediaFile{},
		Quality:  p.PhotoQuality,
		Private:  p.PhotoPrivate,
		Archived: p.DeletedAt.Valid,
		Review:   p.PhotoQuality >= 0 && p.PhotoQuality < 3,
		Duration: p.PhotoDuration.Seconds(),
	}

	// Coordinates are location data, which the session must be allowed to view.
	if s.SeesAnyDetail(acl.ResourcePlaces) && (p.PhotoLat != 0 || p.PhotoLng != 0) {
		result.Location = &Location{Lat: p.PhotoLat, Lng: p.PhotoLng, Altitude: p.PhotoAltitude}
	}

	// The keywords of the picture details, which require full access to pictures.
	if p.Details != nil && s.SeesFullDetail(acl.ResourcePhotos) {
		for _, k := range strings.Split(p.Details.Keywords, ",") {
			if k = strings.TrimSpace(k); k != "" {
				result.Keywords = appendUnique(result.Keywords, k)
			}
		}
	}

	for _, f := range p.Files {
		if f.FileMissing {
			continue
		}

		if f.FilePrimary {
			result.Width = f.FileWidth
			result.Height = f.FileHeight
		}

		result.Files = append(result.Files, MediaFile{Name: path.Base(f.FileName), Type: f.FileType, Size: f.FileSize, Primary: f.FilePrimary})
	}

	return result
}

// registerGetMedia registers the get_media tool.
func registerGetMedia(server *sdkmcp.Server) {
	AddTool(server, getMediaAccess, &sdkmcp.Tool{
		Name: "get_media",
		Description: "Returns one picture or video with its details: the full fields of search_media plus dimensions, " +
			"duration, keywords, files (names without paths, types, and sizes), quality flags, and coordinates if the " +
			"account may view places. Takes a media UID from search_media.",
		Annotations: readOnly(),
	}, getMedia)
}
