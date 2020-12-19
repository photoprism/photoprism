package acl

var Permissions = ACL{
	ResourceDefault: Roles{
		RoleAdmin: Actions{ActionDefault: true},
	},
	ResourceConfig: Roles{
		RoleAdmin: Actions{ActionDefault: true},
		RoleGuest: Actions{ActionRead: true},
	},
	ResourceConfigOptions: Roles{
		RoleAdmin: Actions{ActionDefault: true},
	},
	ResourceAlbums: Roles{
		RoleAdmin: Actions{ActionDefault: true},
		RoleGuest: Actions{ActionSearch: true, ActionRead: true},
	},
	ResourcePhotos: Roles{
		RoleAdmin: Actions{ActionDefault: true},
		RoleGuest: Actions{ActionSearch: true, ActionRead: true, ActionDownload: true},
	},
}
