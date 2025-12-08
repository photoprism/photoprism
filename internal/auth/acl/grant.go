package acl

// Grant represents permissions granted or denied.
type Grant map[Permission]bool

// Standard grants provided to simplify configuration.
var (
	GrantFullAccess = Grant{
		FullAccess:      true,
		AccessAll:       true,
		AccessOwn:       true,
		AccessShared:    true,
		AccessLibrary:   true,
		ActionUse:       true,
		ActionView:      true,
		ActionCreate:    true,
		ActionUpdate:    true,
		ActionDelete:    true,
		ActionDownload:  true,
		ActionShare:     true,
		ActionRate:      true,
		ActionReact:     true,
		ActionManage:    true,
		ActionPublish:   true,
		ActionSubscribe: true,
	}
	GrantUploadAccess = Grant{
		AccessOwn:      true,
		AccessShared:   true,
		ActionSearch:   true,
		ActionView:     true,
		ActionUpload:   true,
		ActionDownload: true,
		ActionReact:    true,
	}
	GrantOwn = Grant{
		AccessOwn:       true,
		ActionView:      true,
		ActionCreate:    true,
		ActionUpdate:    true,
		ActionDelete:    true,
		ActionSubscribe: true,
	}
	GrantAll = Grant{
		AccessAll:       true,
		AccessOwn:       true,
		ActionUse:       true,
		ActionView:      true,
		ActionCreate:    true,
		ActionUpdate:    true,
		ActionDelete:    true,
		ActionPublish:   true,
		ActionSubscribe: true,
	}
	GrantManageOwn = Grant{
		AccessOwn:       true,
		ActionView:      true,
		ActionCreate:    true,
		ActionUpdate:    true,
		ActionDelete:    true,
		ActionSubscribe: true,
		ActionManageOwn: true,
	}
	GrantConfigureOwn = Grant{
		AccessOwn:    true,
		ActionCreate: true,
		ActionUpdate: true,
		ActionDelete: true,
	}
	GrantUpdateOwn = Grant{
		AccessOwn:    true,
		ActionUpdate: true,
	}
	GrantViewOwn = Grant{
		AccessOwn:  true,
		ActionView: true,
	}
	GrantViewUpdateOwn = Grant{
		AccessOwn:    true,
		ActionView:   true,
		ActionUpdate: true,
	}
	GrantViewLibrary = Grant{
		AccessLibrary: true,
		ActionView:    true,
	}
	GrantViewAll = Grant{
		AccessAll:  true,
		AccessOwn:  true,
		ActionView: true,
	}
	GrantViewUpdateAll = Grant{
		AccessAll:    true,
		AccessOwn:    true,
		ActionView:   true,
		ActionUpdate: true,
	}
	GrantViewShared = Grant{
		AccessShared:   true,
		ActionView:     true,
		ActionDownload: true,
	}
	GrantReactShared = Grant{
		AccessShared:   true,
		ActionSearch:   true,
		ActionView:     true,
		ActionDownload: true,
		ActionReact:    true,
	}
	GrantSearchShared = Grant{
		AccessShared:   true,
		ActionSearch:   true,
		ActionView:     true,
		ActionDownload: true,
	}
	GrantSearchAll = Grant{
		AccessAll:    true,
		ActionView:   true,
		ActionSearch: true,
	}
	GrantSearchAllUpdateOwn = Grant{
		AccessAll:       true,
		AccessOwn:       true,
		ActionSearch:    true,
		ActionView:      true,
		ActionUpdateOwn: true,
	}
	GrantSearchDownloadUpdateOwn = Grant{
		AccessShared:    true,
		AccessOwn:       true,
		ActionSearch:    true,
		ActionView:      true,
		ActionDownload:  true,
		ActionUpdateOwn: true,
	}
	GrantSubscribeOwn = Grant{
		AccessOwn:       true,
		ActionSubscribe: true,
	}
	GrantSubscribeAll = Grant{
		AccessAll:       true,
		ActionSubscribe: true,
	}
	GrantPublishOwn = Grant{
		AccessOwn:     true,
		ActionPublish: true,
	}
	GrantUseOwn = Grant{
		AccessOwn: true,
		ActionUse: true,
	}
	GrantNone = Grant{}
)

// GrantDefaults defines default grants for all supported roles.
var GrantDefaults = Roles{
	RoleAdmin:   GrantFullAccess,
	RoleGuest:   GrantReactShared,
	RoleVisitor: GrantViewShared,
	RoleApp:     GrantSearchShared,
	RoleService: GrantSearchShared,
	RolePortal:  GrantFullAccess,
	RoleClient:  GrantFullAccess,
}

// Allow checks if this Grant includes the specified Permission.
func (grant Grant) Allow(perm Permission) bool {
	if result, ok := grant[perm]; ok {
		return result
	} else if result, ok = grant[FullAccess]; ok {
		return result
	}

	return false
}

// DenyAny checks if any of the Permissions are not covered by this Grant.
func (grant Grant) DenyAny(perms Permissions) bool {
	for i := range perms {
		if !grant.Allow(perms[i]) {
			return true
		}
	}

	return false
}
