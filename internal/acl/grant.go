package acl

// Standard grants provided to simplify configuration.
var (
	GrantFullAccess = Grant{
		FullAccess:      true,
		AccessAll:       true,
		AccessOwn:       true,
		AccessShared:    true,
		AccessLibrary:   true,
		ActionCreate:    true,
		ActionUpdate:    true,
		ActionDelete:    true,
		ActionDownload:  true,
		ActionShare:     true,
		ActionRate:      true,
		ActionReact:     true,
		ActionManage:    true,
		ActionSubscribe: true,
	}
	GrantSubscribeAll = Grant{
		AccessAll:       true,
		ActionSubscribe: true,
	}
	GrantSubscribeOwn = Grant{
		AccessOwn:       true,
		ActionSubscribe: true,
	}
	GrantViewAll = Grant{
		AccessAll:  true,
		ActionView: true,
	}
	GrantViewOwn = Grant{
		AccessOwn:  true,
		ActionView: true,
	}
	GrantViewShared = Grant{
		AccessShared:   true,
		ActionView:     true,
		ActionDownload: true,
	}
	GrantSearchShared = Grant{
		AccessShared:   true,
		ActionSearch:   true,
		ActionView:     true,
		ActionDownload: true,
	}
	GrantNone = Grant{}
)

// Grant represents permissions granted or denied.
type Grant map[Permission]bool

// Allow checks whether the permission is granted.
func (grant Grant) Allow(perm Permission) bool {
	if result, ok := grant[perm]; ok {
		return result
	} else if result, ok = grant[FullAccess]; ok {
		return result
	}

	return false
}

// GrantDefaults defines default grants for all supported roles.
var GrantDefaults = Roles{
	RoleAdmin:   GrantFullAccess,
	RoleVisitor: GrantViewShared,
	RoleClient:  GrantFullAccess,
}
