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
		RoleVisitor: Grant{AccessShared: true, ActionView: true, ActionDownload: true},
	},
	ResourceFolders: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: Grant{AccessShared: true, ActionView: true, ActionDownload: true},
	},
	ResourcePlaces: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: Grant{AccessShared: true, ActionView: true, ActionDownload: true},
	},
	ResourceCalendar: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: Grant{AccessShared: true, ActionView: true, ActionDownload: true},
	},
	ResourceMoments: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleVisitor: Grant{AccessShared: true, ActionView: true, ActionDownload: true},
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
		RoleAdmin: GrantFullAccess,
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
	ResourceAccounts: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceUsers: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceConfig: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceDefault: Roles{
		RoleAdmin: GrantFullAccess,
	},
}
