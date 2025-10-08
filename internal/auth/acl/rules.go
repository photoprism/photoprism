package acl

import "sync"

var RulesMutex = &sync.Mutex{}

// Rules specifies granted permissions by Resource and Role.
var Rules = ACL{
	ResourceFiles: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantFullAccess,
	},
	ResourceFolders: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleGuest:   GrantSearchShared,
		RoleVisitor: GrantSearchShared,
		RoleClient:  GrantFullAccess,
	},
	ResourceShares: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantFullAccess,
	},
	ResourcePhotos: GrantDefaults,
	ResourceVideos: GrantDefaults,
	ResourceFavorites: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantFullAccess,
	},
	ResourceAlbums: GrantDefaults,
	ResourceMoments: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleGuest:   GrantSearchShared,
		RoleVisitor: GrantSearchShared,
		RoleClient:  GrantFullAccess,
	},
	ResourceCalendar: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleGuest:   GrantSearchShared,
		RoleVisitor: GrantSearchShared,
		RoleClient:  GrantFullAccess,
	},
	ResourcePeople: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantFullAccess,
	},
	ResourcePlaces: Roles{
		RoleAdmin:    GrantFullAccess,
		RoleGuest:    GrantReactShared,
		RoleVisitor:  GrantViewShared,
		RoleInstance: GrantUseOwn,
		RoleService:  GrantUseOwn,
		RolePortal:   GrantUseOwn,
		RoleClient:   GrantFullAccess,
	},
	ResourceLabels: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantFullAccess,
	},
	ResourceConfig: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleClient:  GrantViewOwn,
		RoleDefault: GrantViewOwn,
	},
	ResourceSettings: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleGuest:   GrantViewUpdateOwn,
		RoleVisitor: GrantViewOwn,
		RolePortal:  GrantFullAccess,
		RoleClient:  GrantViewUpdateOwn,
	},
	ResourceServices: Roles{
		RoleAdmin:  GrantFullAccess,
		RolePortal: GrantFullAccess,
	},
	ResourcePasscode: Roles{
		RoleAdmin:  GrantFullAccess,
		RolePortal: GrantFullAccess,
		RoleGuest:  GrantConfigureOwn,
	},
	ResourcePassword: Roles{
		RoleAdmin:  GrantFullAccess,
		RolePortal: GrantFullAccess,
		RoleGuest:  GrantUpdateOwn,
	},
	ResourceUsers: Roles{
		RoleAdmin:    GrantManageOwn,
		RoleGuest:    GrantViewUpdateOwn,
		RoleInstance: GrantViewOwn,
		RoleService:  GrantViewOwn,
		RolePortal:   GrantFullAccess,
		RoleClient:   GrantViewOwn,
	},
	ResourceSessions: Roles{
		RoleAdmin:   GrantManageOwn,
		RolePortal:  GrantFullAccess,
		RoleDefault: GrantOwn,
	},
	ResourceLogs: Roles{
		RoleAdmin:  GrantFullAccess,
		RolePortal: GrantFullAccess,
		RoleClient: GrantFullAccess,
	},
	ResourceApi: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantPublishOwn,
	},
	ResourceWebDAV: Roles{
		RoleAdmin:  GrantFullAccess,
		RolePortal: GrantFullAccess,
		RoleClient: GrantFullAccess,
	},
	ResourceWebhooks: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantPublishOwn,
	},
	ResourceMetrics: Roles{
		RoleAdmin:    GrantFullAccess,
		RoleInstance: GrantNone,
		RoleService:  GrantViewAll,
		RolePortal:   GrantViewAll,
		RoleClient:   GrantViewAll,
	},
	ResourceVision: Roles{
		RoleAdmin:    GrantFullAccess,
		RoleInstance: GrantUseOwn,
		RoleService:  GrantUseOwn,
		RolePortal:   GrantUseOwn,
		RoleClient:   GrantUseOwn,
	},
	ResourceCluster: Roles{
		RoleAdmin:    GrantFullAccess,
		RoleInstance: GrantSearchDownloadUpdateOwn,
		RoleService:  GrantSearchDownloadUpdateOwn,
		RolePortal:   GrantFullAccess,
		RoleClient:   GrantSearchDownloadUpdateOwn,
	},
	ResourceFeedback: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceDefault: Roles{
		RoleAdmin:    GrantFullAccess,
		RoleInstance: GrantNone,
		RoleService:  GrantNone,
		RolePortal:   GrantNone,
		RoleClient:   GrantNone,
	},
}
