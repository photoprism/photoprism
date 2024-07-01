package acl

// Permission scopes to Grant multiple Permissions for a Resource.
const (
	ScopeRead  Permission = "read"
	ScopeWrite Permission = "write"
)

var (
	GrantScopeRead = Grant{
		AccessShared:    true,
		AccessLibrary:   true,
		AccessPrivate:   true,
		AccessOwn:       true,
		AccessAll:       true,
		ActionSearch:    true,
		ActionView:      true,
		ActionDownload:  true,
		ActionSubscribe: true,
	}
	GrantScopeWrite = Grant{
		AccessShared:    true,
		AccessLibrary:   true,
		AccessPrivate:   true,
		AccessOwn:       true,
		AccessAll:       true,
		ActionUpload:    true,
		ActionCreate:    true,
		ActionUpdate:    true,
		ActionShare:     true,
		ActionDelete:    true,
		ActionRate:      true,
		ActionReact:     true,
		ActionManage:    true,
		ActionManageOwn: true,
	}
)
