package acl

type Role string
type Roles map[Role]Actions

const (
	RoleDefault     Role = "*"
	RoleAdmin       Role = "admin"
	RolePartner     Role = "partner"
	RoleFamily      Role = "family"
	RoleSibling     Role = "sibling"
	RoleParent      Role = "parent"
	RoleGrandparent Role = "grandparent"
	RoleChild       Role = "child"
	RoleFriend      Role = "friend"
	RoleBestFriend  Role = "best-friend"
	RoleClassmate   Role = "classmate"
	RoleWorkmate    Role = "workmate"
	RoleGuest       Role = "guest"
)
