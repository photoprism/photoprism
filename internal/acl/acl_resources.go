package acl

// Resources specifies granted permissions by Resource and Role.
var Resources = ACL{
	ResourceFiles: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourcePhotos: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: Grant{AccessShared: true, ActionView: true, ActionDownload: true},
	},
	ResourceVideos: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: Grant{AccessShared: true, ActionView: true, ActionDownload: true},
	},
	ResourceAlbums: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: GrantSearchShared,
	},
	ResourceFolders: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: GrantSearchShared,
	},
	ResourcePlaces: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: Grant{AccessShared: true, ActionView: true, ActionDownload: true},
	},
	ResourceCalendar: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: GrantSearchShared,
	},
	ResourceMoments: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: GrantSearchShared,
	},
	ResourcePeople: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceFavorites: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceLabels: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceLogs: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceSettings: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: Grant{AccessOwn: true, ActionView: true},
	},
	ResourceFeedback: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourcePassword: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceShares: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceServices: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceUsers: Roles{
		RoleAdmin: Grant{AccessAll: true, AccessOwn: true, ActionView: true, ActionCreate: true, ActionUpdate: true, ActionDelete: true, ActionSubscribe: true},
	},
	ResourceConfig: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceDefault: Roles{
		RoleAdmin: GrantFullAccess,
	},
}
