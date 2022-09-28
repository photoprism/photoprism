package acl

// Resources specifies granted permissions by Resource and Role.
var Resources = ACL{
	ResourceDefault: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceConfig: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceSettings: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceCalendar: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceMoments: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceFiles: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourcePeople: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceFavorites: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceFeedback: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceShares: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourcePassword: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceAccounts: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceLogs: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceLabels: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourcePhotos: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: Grant{AccessShared: true, ActionView: true, ActionDownload: true},
	},
	ResourceAlbums: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: Grant{AccessShared: true, ActionView: true, ActionDownload: true},
	},
	ResourceVideos: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: Grant{AccessShared: true, ActionView: true, ActionDownload: true},
	},
	ResourcePlaces: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: Grant{AccessShared: true, ActionView: true, ActionDownload: true},
	},
	ResourceFolders: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: Grant{AccessShared: true, ActionView: true, ActionDownload: true},
	},
}
