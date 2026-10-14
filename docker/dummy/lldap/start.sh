#!/bin/sh
# Starts lldap and seeds the test users and groups from /seed/bootstrap. The data directory is a tmpfs,
# so every container start begins with an empty directory. The container reports healthy once seeding is done.
/docker-entrypoint.sh run --config-file /data/lldap_config.toml &
pid=$!

i=0
until /app/lldap healthcheck --config-file /data/lldap_config.toml >/dev/null 2>&1; do
  i=$((i + 1))
  [ "$i" -gt 60 ] && break
  sleep 1
done

LLDAP_URL=http://localhost:17170 LLDAP_ADMIN_USERNAME=admin LLDAP_ADMIN_PASSWORD="$LLDAP_LDAP_USER_PASS" \
  USER_CONFIGS_DIR=/seed/bootstrap/users GROUP_CONFIGS_DIR=/seed/bootstrap/groups \
  /app/bootstrap.sh >/tmp/bootstrap.log 2>&1 && touch /tmp/seeded && echo "dummy-lldap: seeded test users and groups" || cat /tmp/bootstrap.log

wait "$pid"
