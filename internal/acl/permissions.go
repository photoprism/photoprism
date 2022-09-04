package acl

var Permissions = ACL{
	ResourceDefault: Roles{
		RoleAdmin: Actions{ActionDefault: true},
	},
	ResourceConfig: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionRead: true},
		RoleViewer: Actions{ActionRead: true},
		RoleGuest:  Actions{ActionRead: true},
	},
	ResourceConfigOptions: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionRead: true},
	},
	ResourceSettings: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionDefault: true},
	},
	ResourceLogs: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionRead: true},
		RoleViewer: Actions{ActionRead: true},
	},
	ResourceAccounts: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionRead: true},
	},
	ResourceSubjects: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionRead: true},
		RoleViewer: Actions{ActionRead: true},
	},
	ResourceAlbums: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionDefault: true},
		RoleViewer: Actions{ActionSearch: true, ActionRead: true, ActionComment: true},
		RoleGuest:  Actions{ActionSearch: true, ActionRead: true},
	},
	ResourceCameras: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionDefault: true},
		RoleViewer: Actions{ActionSearch: true, ActionRead: true},
		RoleGuest:  Actions{ActionSearch: true, ActionRead: true},
	},
	ResourceCategories: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionDefault: true},
		RoleViewer: Actions{ActionSearch: true, ActionRead: true},
		RoleGuest:  Actions{ActionSearch: true, ActionRead: true},
	},
	ResourceCountries: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionDefault: true},
		RoleViewer: Actions{ActionSearch: true, ActionRead: true},
		RoleGuest:  Actions{ActionSearch: true, ActionRead: true},
	},
	ResourceFiles: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionDefault: true},
		RoleViewer: Actions{ActionSearch: true, ActionRead: true},
		RoleGuest:  Actions{ActionSearch: true, ActionRead: true},
	},
	ResourceFolders: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionDefault: true},
		RoleViewer: Actions{ActionSearch: true, ActionRead: true},
		RoleGuest:  Actions{ActionSearch: true, ActionRead: true},
	},
	ResourceLabels: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionDefault: true},
		RoleViewer: Actions{ActionSearch: true, ActionRead: true},
		RoleGuest:  Actions{ActionSearch: true, ActionRead: true},
	},
	ResourceLenses: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionDefault: true},
		RoleViewer: Actions{ActionSearch: true, ActionRead: true},
		RoleGuest:  Actions{ActionSearch: true, ActionRead: true},
	},
	ResourceLinks: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionDefault: true},
	},
	ResourceGeo: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionDefault: true},
		RoleViewer: Actions{ActionSearch: true, ActionRead: true, ActionComment: true},
		RoleGuest:  Actions{ActionSearch: true, ActionRead: true},
	},
	ResourcePhotos: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionDefault: true},
		RoleViewer: Actions{ActionSearch: true, ActionRead: true, ActionDownload: true, ActionComment: true},
		RoleGuest:  Actions{ActionSearch: true, ActionRead: true, ActionDownload: true},
	},
	ResourcePrivate: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionDefault: true},
	},
	ResourcePlaces: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionDefault: true},
		RoleViewer: Actions{ActionSearch: true, ActionRead: true, ActionDownload: true},
	},
	ResourceUsers: Roles{
		RoleAdmin:   Actions{ActionDefault: true},
		RoleDefault: Actions{ActionUpdateSelf: true},
	},
	ResourcePasswords: Roles{
		RoleAdmin:  Actions{ActionDefault: true},
		RoleEditor: Actions{ActionUpdateSelf: true},
		RoleViewer: Actions{ActionUpdateSelf: true},
	},
}
