package acl

// Grants represents Permission Grant by Resource.
type Grants map[Resource]Grant

// Grants returns the permissions granted to the specified Role by Resource.
func (acl ACL) Grants(role Role) Grants {
	result := make(map[Resource]Grant, len(acl))

	for resource := range acl {
		result[resource] = acl[resource][role]
	}

	return result
}
