package acl

import "sync"

// EventsMutex guards concurrent updates to the Events ACL.
var EventsMutex = &sync.Mutex{}

// Events specifies granted permissions by event channel and Role.
var Events = ACL{
	ResourceDefault: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ChannelAudit: Roles{
		RoleAdmin:  GrantFullAccess,
		RolePortal: GrantFullAccess,
	},
	ChannelLog: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ChannelSystem: Roles{
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
