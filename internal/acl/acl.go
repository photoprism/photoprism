/*
Package acl provides access control lists for authorization checks.

Copyright (c) 2018 - 2024 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package acl

// ACL represents an access control list based on Resource, Roles, and Permissions.
type ACL map[Resource]Roles

// Deny checks whether the role must be denied access to the specified resource.
func (acl ACL) Deny(resource Resource, role Role, perm Permission) bool {
	return !acl.Allow(resource, role, perm)
}

// DenyAll checks whether the role is granted none of the permissions for the specified resource.
func (acl ACL) DenyAll(resource Resource, role Role, perms Permissions) bool {
	return !acl.AllowAny(resource, role, perms)
}

// Allow checks whether the role is granted permission for the specified resource.
func (acl ACL) Allow(resource Resource, role Role, perm Permission) bool {
	if p, ok := acl[resource]; ok {
		return p.Allow(role, perm)
	} else if p, ok = acl[ResourceDefault]; ok {
		return p.Allow(role, perm)
	}

	return false
}

// AllowAny checks whether the role is granted any of the permissions for the specified resource.
func (acl ACL) AllowAny(resource Resource, role Role, perms Permissions) bool {
	if len(perms) == 0 {
		return false
	}

	for i := range perms {
		if acl.Allow(resource, role, perms[i]) {
			return true
		}
	}

	return false
}

// AllowAll checks whether the role is granted all of the permissions for the specified resource.
func (acl ACL) AllowAll(resource Resource, role Role, perms Permissions) bool {
	if len(perms) == 0 {
		return false
	}

	for i := range perms {
		if acl.Deny(resource, role, perms[i]) {
			return false
		}
	}

	return true
}

// Resources returns the resources specified in the ACL.
func (acl ACL) Resources() (result []string) {
	if len(acl) == 0 {
		return []string{}
	}

	result = make([]string, 0, len(acl))

	for resource := range acl {
		if resource != ResourceDefault && resource.String() != "" {
			result = append(result, resource.String())
		}
	}

	return result
}
