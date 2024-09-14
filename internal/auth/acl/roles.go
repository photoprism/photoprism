package acl

// RoleStrings represents user role names mapped to roles.
type RoleStrings = map[string]Role

// UserRoles maps valid user account roles.
var UserRoles = RoleStrings{
	string(RoleAdmin):   RoleAdmin,
	string(RoleGuest):   RoleGuest,
	string(RoleVisitor): RoleVisitor,
	string(RoleNone):    RoleNone,
}

// ClientRoles maps valid API client roles.
var ClientRoles = RoleStrings{
	string(RoleAdmin):  RoleAdmin,
	string(RoleClient): RoleClient,
	string(RoleNone):   RoleNone,
}

// Roles grants permissions to roles.
type Roles map[Role]Grant

// Allow checks whether the permission is granted based on the role.
func (roles Roles) Allow(role Role, grant Permission) bool {
	if a, ok := roles[role]; ok {
		return a.Allow(grant)
	} else if a, ok = roles[RoleDefault]; ok {
		return a.Allow(grant)
	}

	return false
}
