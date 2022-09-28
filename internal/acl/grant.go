package acl

// Predefined grants to simplify configuration.
var (
	GrantFullAccess   = Grant{FullAccess: true, AccessAll: true, ActionCreate: true, ActionUpdate: true, ActionDelete: true, ActionDownload: true, ActionShare: true, ActionRate: true, ActionReact: true, ActionManage: true, ActionSubscribe: true}
	GrantSubscribeAll = Grant{AccessAll: true, ActionSubscribe: true}
	GrantSubscribeOwn = Grant{AccessOwn: true, ActionSubscribe: true}
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
