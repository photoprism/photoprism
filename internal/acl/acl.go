/*

Package acl contains PhotoPrism's access control lists for authorizing user actions.

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.org>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism® is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
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
