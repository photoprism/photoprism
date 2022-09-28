package acl

// Roles that can be assigned to users.
const (
	RoleAdmin        Role = "admin"
	RoleVisitor      Role = "visitor"
	RoleUnauthorized Role = "unauthorized"
	RoleDefault      Role = "default"
	RoleUnknown      Role = ""
)

// ValidRoles specifies the valid user roles.
var ValidRoles = map[string]Role{
	string(RoleAdmin):        RoleAdmin,
	string(RoleVisitor):      RoleVisitor,
	string(RoleUnauthorized): RoleUnauthorized,
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
