package acl

type Role string
type Roles map[Role]Actions

const (
	RoleDefault Role = "*"
	RoleAdmin   Role = "admin"
	RoleChild   Role = "child"
	RoleFamily  Role = "family"
	RoleFriend  Role = "friend"
	RoleGuest   Role = "guest"
)
