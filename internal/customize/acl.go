package customize

import (
	"github.com/photoprism/photoprism/internal/acl"
)

// ApplyACL updates the current settings based on the access control list provided.
func (s *Settings) ApplyACL(list acl.ACL, role acl.Role) *Settings {
	m := *s

	// Features.
	m.Features.Search = s.Features.Search && list.AllowAny(acl.ResourcePhotos, role, acl.Permissions{acl.ActionSearch})
	m.Features.Ratings = s.Features.Ratings && list.AllowAny(acl.ResourcePhotos, role, acl.Permissions{acl.ActionRate})
	m.Features.Reactions = s.Features.Reactions && list.AllowAny(acl.ResourcePhotos, role, acl.Permissions{acl.ActionReact})
	m.Features.Videos = s.Features.Videos && list.AllowAny(acl.ResourceVideos, role, acl.Permissions{acl.ActionSearch})
	m.Features.Albums = s.Features.Albums && list.AllowAny(acl.ResourceAlbums, role, acl.Permissions{acl.ActionView})
	m.Features.Favorites = s.Features.Favorites && list.AllowAny(acl.ResourceFavorites, role, acl.Permissions{acl.ActionSearch})
	m.Features.Moments = s.Features.Moments && list.AllowAny(acl.ResourceMoments, role, acl.Permissions{acl.ActionSearch})
	m.Features.People = s.Features.People && list.AllowAny(acl.ResourcePeople, role, acl.Permissions{acl.ActionSearch})
	m.Features.Labels = s.Features.Labels && list.AllowAny(acl.ResourceLabels, role, acl.Permissions{acl.ActionSearch})
	m.Features.Places = s.Features.Places && list.AllowAny(acl.ResourcePlaces, role, acl.Permissions{acl.ActionSearch})
	m.Features.Folders = s.Features.Folders && list.AllowAny(acl.ResourceFolders, role, acl.Permissions{acl.ActionSearch})
	m.Features.Private = s.Features.Private && list.AllowAny(acl.ResourcePhotos, role, acl.Permissions{acl.AccessPrivate})

	// Permissions.
	m.Features.Edit = s.Features.Edit && list.AllowAny(acl.ResourcePhotos, role, acl.Permissions{acl.ActionUpdate})
	m.Features.Review = s.Features.Review && list.Allow(acl.ResourcePhotos, role, acl.ActionManage)
	m.Features.Archive = s.Features.Archive && list.AllowAny(acl.ResourcePhotos, role, acl.Permissions{acl.ActionDelete})
	m.Features.Delete = s.Features.Delete && list.AllowAny(acl.ResourcePhotos, role, acl.Permissions{acl.ActionDelete})
	m.Features.Share = s.Features.Share && list.AllowAny(acl.ResourceShares, role, acl.Permissions{acl.ActionManage})

	// Browse, upload and download files.
	m.Features.Files = s.Features.Files && list.AllowAny(acl.ResourceFiles, role, acl.Permissions{acl.ActionSearch})
	m.Features.Upload = s.Features.Upload && list.Allow(acl.ResourcePhotos, role, acl.ActionUpload)
	m.Features.Download = s.Features.Download && !s.Download.Disabled && list.Allow(acl.ResourcePhotos, role, acl.ActionDownload)

	// Library.
	m.Features.Library = s.Features.Library && list.Allow(acl.ResourceFiles, role, acl.ActionManage)
	m.Features.Import = s.Features.Import && list.Allow(acl.ResourceFiles, role, acl.ActionManage) && list.Allow(acl.ResourcePhotos, role, acl.ActionCreate)
	m.Features.Logs = s.Features.Logs && list.Allow(acl.ResourceLogs, role, acl.ActionView)

	// Settings.
	m.Features.Settings = s.Features.Settings && list.Allow(acl.ResourceSettings, role, acl.ActionUpdate)
	m.Features.Advanced = s.Features.Advanced && list.Allow(acl.ResourceConfig, role, acl.ActionManage)
	m.Features.Sync = s.Features.Sync && list.Allow(acl.ResourceAccounts, role, acl.ActionManage)
	m.Features.Account = s.Features.Account && list.Allow(acl.ResourcePassword, role, acl.ActionUpdate)

	return &m
}
