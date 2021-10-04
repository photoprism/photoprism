package acl

var Permissions = ACL{
	ResourceDefault: Roles{
		RoleAdmin: Actions{ActionDefault: true},
	},
	ResourceConfig: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleGuest:  Actions{ActionRead: true},
		RoleMember: Actions{ActionRead: true},
	},
	ResourceConfigOptions: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleMember: Actions{ActionRead: true},
	},
	ResourceSubjects: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleMember: Actions{ActionRead: true},
	},
	ResourceAlbums: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleGuest:  Actions{ActionSearch: true, ActionRead: true},
		RoleMember: Actions{ActionSearch: true, ActionRead: true},
	},
	ResourcePhotos: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleGuest:  Actions{ActionSearch: true, ActionRead: true, ActionDownload: true},
		RoleMember: Actions{ActionSearch: true, ActionRead: true, ActionDownload: true},
	},
	ResourceUsers: Roles{
		RoleDefault: Actions{ActionUpdateSelf: true},
		RoleAdmin:   Actions{ActionUpdateSelf: true},
		RoleMember:  Actions{ActionUpdateSelf: true},
	},
}
