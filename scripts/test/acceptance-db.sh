#!/usr/bin/env bash

# Copies the index of a TestCafe fixture into the "testcafe" database on MariaDB or PostgreSQL, and writes a
# "config-<db>-generated" folder next to the fixture's "config-sqlite" folder that uses it. With sqlite, it does nothing.
set -euo pipefail

if (( $# != 3 )); then
  echo "Usage: $0 <sqlite|mariadb|postgres> <photoprism binary> <config-sqlite path>" >&2
  exit 2
fi

db=$1
bin=$2
src=${3%/}
name=testcafe
user=testcafe
pass=testcafe

case "$db" in
  sqlite)
    exit 0
    ;;
  mariadb)
    driver=mysql
    server=mariadb:${MARIADB_PORT:-4001}
    ;;
  postgres)
    driver=postgres
    server=postgres:${POSTGRES_PORT:-4002}
    ;;
  *)
    echo "$0: unsupported database \"$db\", use sqlite, mariadb, or postgres" >&2
    exit 2
    ;;
esac

if [[ ! "${server##*:}" =~ ^[1-9][0-9]{0,4}$ ]] || (( ${server##*:} > 65535 )); then
  echo "$0: MARIADB_PORT and POSTGRES_PORT must be port numbers" >&2
  exit 2
fi

dir=$(dirname "$src")
root=$(realpath -- "$dir")
dst=$dir/config-$db-generated
marker=.testcafe-generated

if [[ "$(basename "$src")" != "config-sqlite" || ! -f "$src/options.yml" ]]; then
  echo "$0: $src/options.yml not found, the fixture config must be in a folder named config-sqlite" >&2
  exit 1
elif [[ -L "$src" || ! -d "$src" ]]; then
  echo "$0: $src must be a folder, not a symbolic link" >&2
  exit 1
elif [[ -e "$dst" || -L "$dst" ]] && [[ -L "$dst" || ! -f "$dst/$marker" ]]; then
  echo "$0: $dst was not created by this script, remove or rename it first" >&2
  exit 1
fi

# verify checks with the binary that the config uses the fixture storage, originals and import folders, and
# contains the given report lines. The node UUID is set, since the report would otherwise save a new one.
verify() {
  local cfg=$1 report storage line key folder
  shift
  report=$(PHOTOPRISM_NODE_UUID=019a0000-0000-7000-8000-000000000001 "$bin" -c "$cfg" show config --tsv 2>/dev/null) ||
    return 1
  storage=$(sed -n 's/^storage-path\t//p' <<< "$report")
  [[ -n "$storage" && "$storage" -ef "$dir" ]] || return 1

  for key in originals-path import-path; do
    folder=$(sed -n "s/^$key\t//p" <<< "$report")
    [[ -n "$root" && -n "$folder" && "$(realpath -m -- "$folder")/" == "$root/"* ]] || return 1
  done

  for line in "$@"; do
    grep -qxF -- "$line" <<< "$report" || return 1
  done
}

if ! verify "$src" $'database-driver\tsqlite3'; then
  echo "$0: $src does not use SQLite with the storage, originals and import folders in $dir" >&2
  exit 1
fi

# The client commands may consist of several words, e.g. to run them in a container.
read -ra mariadb <<< "${MARIADB:-mariadb}"
read -ra psql <<< "${PSQL:-psql}"

echo "Copying the index to the $db database \"$name\" on $server..."

if [[ "$db" == "mariadb" ]]; then
  "${mariadb[@]}" <<EOF
CREATE DATABASE IF NOT EXISTS \`$name\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS '$user'@'%' IDENTIFIED BY '$pass';
ALTER USER '$user'@'%' IDENTIFIED BY '$pass';
GRANT ALL PRIVILEGES ON \`$name\`.* TO '$user'@'%';
EOF
else
  "${psql[@]}" "${POSTGRES_ADMIN_URI:-postgres://photoprism:photoprism@postgres:${POSTGRES_PORT:-4002}/postgres}" \
    -v ON_ERROR_STOP=1 -q <<EOF
SELECT 'CREATE USER $user' WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '$user')\gexec
ALTER USER $user LOGIN PASSWORD '$pass';
SELECT 'CREATE DATABASE $name OWNER $user ENCODING ''UTF8'' TEMPLATE template0 LOCALE_PROVIDER icu ICU_LOCALE ''und'''
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$name')\gexec
EOF
fi

rm -rf "$dst"
PHOTOPRISM_TRANSFER_PASSWORD=$pass "$bin" -c "$src" database transfer --force \
  --target-driver "$db" --target-server "$server" --target-name "$name" --target-user "$user"

# Builds the folder under a temporary name, so that an interrupted run leaves no partial config behind.
tmp=$(mktemp -d "$dir/.config-$db-generated.XXXXXX")
trap 'rm -rf "$tmp"' EXIT
touch "$tmp/$marker"
cp -RLp "$src/." "$tmp/"

# Links the hub registration, so that both folders keep using the same one when it is renewed.
if [[ -f "$src/hub.yml" ]]; then
  ln -f "$src/hub.yml" "$tmp/hub.yml"
fi

# Sets all database options, since options.yml takes precedence over flags and environment variables.
rm -f "$tmp/options.yml"
(
  umask 077
  {
    grep -v '^Database' "$src/options.yml" || (( $? == 1 ))
    printf 'DatabaseDriver: %s\nDatabaseDSN: ""\nDatabaseServer: "%s"\nDatabaseName: "%s"\nDatabaseUser: "%s"\nDatabasePassword: "%s"\n' \
      "$driver" "$server" "$name" "$user" "$pass"
  } > "$tmp/options.yml"
)

if ! verify "$tmp" $'database-driver\t'"$driver" $'database-name\t'"$name" $'database-server\t'"$server" \
  $'database-user\t'"$user"; then
  echo "$0: the options in $dst would not take effect, check $src/options.yml" >&2
  exit 1
fi

mv -T "$tmp" "$dst"
echo "Created $dst."
