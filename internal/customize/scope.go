package customize

import (
	"strings"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/pkg/list"
)

// ApplyScope updates the current settings based on the authorization scope passed.
func (s *Settings) ApplyScope(scope string) *Settings {
	if scope == "" || scope == list.All {
		return s
	}

	scopes := list.ParseAttr(strings.ToLower(scope))

	if scopes.Contains(acl.ResourceSettings.String()) {
		return s
	}

	m := *s

	// Features.
	m.Features.Albums = s.Features.Albums && scopes.Contains(acl.ResourceAlbums.String())
	m.Features.Favorites = s.Features.Favorites && scopes.Contains(acl.ResourceFavorites.String())
	m.Features.Folders = s.Features.Folders && scopes.Contains(acl.ResourceFolders.String())
	m.Features.Labels = s.Features.Labels && scopes.Contains(acl.ResourceLabels.String())
	m.Features.Moments = s.Features.Moments && scopes.Contains(acl.ResourceMoments.String())
	m.Features.People = s.Features.People && scopes.Contains(acl.ResourcePeople.String())
	m.Features.Places = s.Features.Places && scopes.Contains(acl.ResourcePlaces.String())
	m.Features.Private = s.Features.Private && scopes.Contains(acl.ResourcePhotos.String())
	m.Features.Ratings = s.Features.Ratings && scopes.Contains(acl.ResourcePhotos.String())
	m.Features.Reactions = s.Features.Reactions && scopes.Contains(acl.ResourcePhotos.String())
	m.Features.Search = s.Features.Search && scopes.Contains(acl.ResourcePhotos.String())
	m.Features.Videos = s.Features.Videos && scopes.Contains(acl.ResourceVideos.String())

	// Permissions.
	m.Features.Archive = s.Features.Archive && scopes.Contains(acl.ResourcePhotos.String())
	m.Features.Delete = s.Features.Delete && scopes.Contains(acl.ResourcePhotos.String())
	m.Features.Edit = s.Features.Edit && scopes.Contains(acl.ResourcePhotos.String())
	m.Features.Share = s.Features.Share && scopes.Contains(acl.ResourceShares.String())

	// Browse, upload and download files.
	m.Features.Upload = s.Features.Upload && scopes.Contains(acl.ResourcePhotos.String())
	m.Features.Download = s.Features.Download && scopes.Contains(acl.ResourcePhotos.String())

	// Library.
	m.Features.Import = s.Features.Import && scopes.Contains(acl.ResourceFiles.String())
	m.Features.Library = s.Features.Library && scopes.Contains(acl.ResourceFiles.String())
	m.Features.Files = s.Features.Files && scopes.Contains(acl.ResourceFiles.String())
	m.Features.Logs = s.Features.Logs && scopes.Contains(acl.ResourceLogs.String())

	// Settings.
	m.Features.Account = s.Features.Account && scopes.Contains(acl.ResourcePassword.String())
	m.Features.Settings = s.Features.Settings && scopes.Contains(acl.ResourceSettings.String())
	m.Features.Services = s.Features.Services && scopes.Contains(acl.ResourceServices.String())

	return &m
}
