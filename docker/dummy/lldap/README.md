# dummy-lldap

[lldap](https://github.com/lldap/lldap) as a second LDAP implementation for development and tests, next to
`dummy-ldap` (glauth). Unlike glauth, lldap returns `memberOf` and `displayName` only when they are requested by
name, and it binds users by DN only (`uid=<name>,ou=people,dc=localssl,dc=dev`), not by user principal name.

- LDAP: `ldap://dummy-lldap:3890` within the compose network; base DN `dc=localssl,dc=dev`, users in
  `ou=people`, groups in `ou=groups` (`cn=photoprism-<name>,ou=groups,dc=localssl,dc=dev`).
- Test users (password `photoprism`): `lladmin` (photoprism-admin), `lluser` (photoprism-user,
  photoprism-webdav), `llguest` (photoprism-guest, photoprism-nologin). Admin: `admin` / `photoprism`.
- `start.sh` seeds `bootstrap/` on every start; the data directory is a tmpfs.

PhotoPrism settings for this directory:

```yaml
PHOTOPRISM_LDAP_URI: "ldap://dummy-lldap:3890"
PHOTOPRISM_LDAP_BIND_DN: "uid"
PHOTOPRISM_LDAP_BASE_DN: "ou=people,dc=localssl,dc=dev"
PHOTOPRISM_LDAP_ROLE_DN: "cn=photoprism-*,ou=groups,dc=localssl,dc=dev"
PHOTOPRISM_LDAP_WEBDAV_DN: "cn=photoprism-webdav,ou=groups,dc=localssl,dc=dev"
PHOTOPRISM_LDAP_NOLOGIN_DN: "cn=photoprism-nologin,ou=groups,dc=localssl,dc=dev"
```
