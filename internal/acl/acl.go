/*

Package acl contains PhotoPrism's access control lists for authorizing user actions.

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://photoprism.app/trademark>

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

*/
package acl

type Permission struct {
	Roles   Roles
	Actions Actions
}

type ACL map[Resource]Roles

func (l ACL) Deny(resource Resource, role Role, action Action) bool {
	return !l.Allow(resource, role, action)
}

func (l ACL) Allow(resource Resource, role Role, action Action) bool {
	if p, ok := l[resource]; ok {
		return p.Allow(role, action)
	} else if p, ok := l[ResourceDefault]; ok {
		return p.Allow(role, action)
	}

	return false
}

func (a Actions) Allow(action Action) bool {
	if result, ok := a[action]; ok {
		return result
	} else if result, ok := a[ActionDefault]; ok {
		return result
	}

	return false
}

func (r Roles) Allow(role Role, action Action) bool {
	if a, ok := r[role]; ok {
		return a.Allow(action)
	} else if a, ok := r[RoleDefault]; ok {
		return a.Allow(action)
	}

	return false
}
