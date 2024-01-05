package acl

// Resources specifies granted permissions by Resource and Role.
var Resources = ACL{
	ResourceFiles: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantFullAccess,
	},
	ResourceFolders: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: GrantSearchShared,
		RoleClient:  GrantFullAccess,
	},
	ResourceShares: Roles{
		RoleAdmin: GrantFullAccess,
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
		RoleVisitor: GrantSearchShared,
		RoleClient:  GrantFullAccess,
	},
	ResourceCalendar: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: GrantSearchShared,
		RoleClient:  GrantFullAccess,
	},
	ResourcePeople: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantFullAccess,
	},
	ResourcePlaces: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: GrantViewShared,
		RoleClient:  GrantFullAccess,
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
		RoleVisitor: Grant{AccessOwn: true, ActionView: true},
		RoleClient:  Grant{AccessOwn: true, ActionView: true, ActionUpdate: true},
	},
	ResourceServices: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourcePassword: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceUsers: Roles{
		RoleAdmin: Grant{AccessAll: true, AccessOwn: true, ActionView: true, ActionCreate: true, ActionUpdate: true, ActionDelete: true, ActionSubscribe: true},
	},
	ResourceLogs: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantFullAccess,
	},
	ResourceWebDAV: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantFullAccess,
	},
	ResourceMetrics: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantViewAll,
	},
	ResourceFeedback: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceDefault: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantNone,
	},
}
