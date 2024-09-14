package acl

// Events specifies granted permissions by event channel and Role.
var Events = ACL{
	ResourceDefault: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ChannelUser: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleGuest:   GrantSubscribeOwn,
		RoleVisitor: GrantSubscribeOwn,
	},
	ChannelSession: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleGuest:   GrantSubscribeOwn,
		RoleVisitor: GrantSubscribeOwn,
	},
}
