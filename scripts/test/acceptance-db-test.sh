#!/usr/bin/env bash

# Tests acceptance-db.sh with scratch fixtures and stand-ins for the database clients and the photoprism binary.
set -uo pipefail

script="$(cd "$(dirname "$0")" && pwd)/acceptance-db.sh"
root="$(cd "$(dirname "$0")/../.." && pwd)"
unset MARIADB_PORT POSTGRES_PORT POSTGRES_ADMIN_URI FAIL FAIL_CLIENT
tmp=$(mktemp -d) || exit 1
trap 'rm -rf "$tmp"' EXIT
failed=0

# The stand-ins log their arguments, standard input, and the transfer password, and fail if FAIL is set.
# The binary reports the options of a config folder up to the end of the first YAML document, and saves a
# node UUID in it if none is set.
cat > "$tmp/client" <<'EOF'
#!/usr/bin/env bash
{ echo "args: $*"; cat; } >> "$LOG"
[[ -z "${FAIL_CLIENT:-}" ]]
EOF
cat > "$tmp/photoprism" <<'EOF'
#!/usr/bin/env bash
if [[ "$1 $3 $4 $5" == "-c show config --tsv" ]]; then
  [[ -n "${PHOTOPRISM_NODE_UUID:-}" ]] || echo "NodeUUID: generated" >> "$2/options.yml"
  exec awk '/^\.\.\.$/ || (/^---$/ && NR > 1) { exit } { k = $0; sub(/:.*/, "", k); v = $0; sub(/^[^:]*: */, "", v); gsub(/"/, "", v) }
    k == "StoragePath" { print "storage-path\t" v } k == "DatabaseName" { print "database-name\t" v }
    k == "OriginalsPath" { print "originals-path\t" v } k == "ImportPath" { print "import-path\t" v }
    k == "DatabaseDriver" { print "database-driver\t" (v == "sqlite" ? "sqlite3" : v) }
    k == "DatabaseServer" { print "database-server\t" v } k == "DatabaseUser" { print "database-user\t" v }' "$2/options.yml"
fi
echo "args: $* password: ${PHOTOPRISM_TRANSFER_PASSWORD:-}" >> "$LOG"
[[ -z "${FAIL:-}" ]]
EOF
chmod +x "$tmp/client" "$tmp/photoprism"
export LOG=$tmp/log MARIADB="$tmp/client --defaults" PSQL="$tmp/client --pg"

# fixture creates a fixture folder with a config-sqlite folder and returns its path.
fixture() {
  mkdir -p "$tmp/$1/config-sqlite"
  printf 'DatabaseDriver: sqlite\nDatabaseDSN: ./index.db\nHttpPort: 2343\nImportPath: %s/import\nOriginalsPath: %s/originals\nStoragePath: %s\n' \
    "$tmp/$1" "$tmp/$1" "$tmp/$1" > "$tmp/$1/config-sqlite/options.yml"
  echo "hub" > "$tmp/$1/config-sqlite/hub.yml"
  echo "settings" > "$tmp/$1/config-sqlite/settings.yml"
  mkdir -p "$tmp/$1/config-sqlite/keys"
  echo "key" > "$tmp/$1/config-sqlite/keys/key.pem"
  echo "$tmp/$1"
}

# check reports a failed condition with its name.
check() {
  local name=$1
  shift

  if "$@"; then
    echo "ok   $name"
  else
    echo "FAIL $name"
    failed=1
  fi
}

# lacks reports whether the file does not contain the text.
# shellcheck disable=SC2329
lacks() {
  ! grep -q -- "$1" "$2"
}

# expect runs the script and checks its exit status and output.
expect() {
  local name=$1 want=$2 pattern=$3 out status
  shift 3
  : > "$LOG"
  out=$(bash "$script" "$@" 2>&1)
  status=$?

  if [[ "$status" != "$want" ]] || ! grep -q -- "$pattern" <<< "$out"; then
    echo "FAIL $name: exit $status, want $want and \"$pattern\": $out"
    failed=1
  else
    echo "ok   $name"
  fi
}

fx=$(fixture fx)
expect "Usage" 2 "Usage" mariadb "$tmp/photoprism"
expect "Unsupported" 2 "unsupported database" mysql "$tmp/photoprism" "$fx/config-sqlite"
expect "SQLite" 0 "" sqlite "$tmp/photoprism" "$fx/config-sqlite"
check "SQLiteNoChanges" test "$(ls "$fx")" == config-sqlite -a ! -s "$LOG"
mkdir -p "$fx/other"
cp "$fx/config-sqlite/options.yml" "$fx/other/"
expect "SourceName" 1 "config-sqlite" mariadb "$tmp/photoprism" "$fx/other"
check "SourceNameNoTarget" test ! -e "$fx/config-mariadb-generated"
expect "SourceMissing" 1 "not found" mariadb "$tmp/photoprism" "$tmp/missing/config-sqlite"

cp "$fx/config-sqlite/options.yml" "$tmp/fx-options.yml"
mkdir "$fx/config-mariadb"
touch "$fx/config-mariadb/keep"
expect "MariaDB" 0 "Created $fx/config-mariadb-generated" mariadb "$tmp/photoprism" "$fx/config-sqlite/"
check "MariaDBFixtureFolderKept" test -f "$fx/config-mariadb/keep"
check "MariaDBSourceIdentical" cmp -s "$tmp/fx-options.yml" "$fx/config-sqlite/options.yml"
opt=$fx/config-mariadb-generated/options.yml
check "MariaDBMarker" test -f "$fx/config-mariadb-generated/.testcafe-generated"
check "MariaDBDriver" grep -qx 'DatabaseDriver: mysql' "$opt"
check "MariaDBServer" grep -qx 'DatabaseServer: "mariadb:4001"' "$opt"
check "MariaDBName" grep -qx 'DatabaseName: "testcafe"' "$opt"
check "MariaDBEmptyDSN" grep -qx 'DatabaseDSN: ""' "$opt"
check "MariaDBUser" grep -qx 'DatabaseUser: "testcafe"' "$opt"
check "MariaDBPassword" grep -qx 'DatabasePassword: "testcafe"' "$opt"
check "MariaDBDatabaseKeys" test "$(grep -c '^Database' "$opt")" -eq 6
check "MariaDBPrivate" test "$(stat -c %a "$opt")" == 600
check "MariaDBSettings" grep -qx settings "$fx/config-mariadb-generated/settings.yml"
check "MariaDBKeys" grep -qx key "$fx/config-mariadb-generated/keys/key.pem"
check "MariaDBNoTempFolder" test -z "$(find "$fx" -maxdepth 1 -name '.config-*')"
check "MariaDBOtherOptions" grep -qx 'HttpPort: 2343' "$opt"
check "MariaDBHubLinked" test "$fx/config-mariadb-generated/hub.yml" -ef "$fx/config-sqlite/hub.yml"
check "MariaDBSourceUnchanged" grep -qx 'DatabaseDriver: sqlite' "$fx/config-sqlite/options.yml"
check "MariaDBClient" grep -qx "args: --defaults" "$LOG"
check "MariaDBCreatesDatabase" grep -q "CREATE DATABASE IF NOT EXISTS \`testcafe\`" "$LOG"
check "MariaDBUserPassword" grep -q "ALTER USER 'testcafe'@'%' IDENTIFIED BY 'testcafe';" "$LOG"
check "MariaDBGrant" grep -q "GRANT ALL PRIVILEGES ON \`testcafe\`.\* TO 'testcafe'@'%';" "$LOG"
check "MariaDBTransfer" grep -q -- "-c $fx/config-sqlite database transfer --force --target-driver mariadb --target-server mariadb:4001 --target-name testcafe --target-user testcafe password: testcafe" "$LOG"
check "MariaDBNoPasswordFlag" lacks "--target-password" "$LOG"

# A folder created by the script is replaced, including files added to it.
touch "$fx/config-mariadb-generated/stale"
MARIADB_PORT=3306 expect "Replace" 0 "mariadb:3306" mariadb "$tmp/photoprism" "$fx/config-sqlite"
check "ReplaceRemovesFiles" test ! -e "$fx/config-mariadb-generated/stale"
check "ReplacePort" grep -qx 'DatabaseServer: "mariadb:3306"' "$opt"

# A failed transfer leaves no config folder behind.
FAIL=1 expect "TransferFails" 1 "" mariadb "$tmp/photoprism" "$fx/config-sqlite"
check "TransferFailsNoTarget" test ! -e "$fx/config-mariadb-generated"

expect "PostgreSQL" 0 "Created $fx/config-postgres-generated" postgres "$tmp/photoprism" "$fx/config-sqlite"
check "PostgreSQLDriver" grep -qx 'DatabaseDriver: postgres' "$fx/config-postgres-generated/options.yml"
check "PostgreSQLServer" grep -qx 'DatabaseServer: "postgres:4002"' "$fx/config-postgres-generated/options.yml"
check "PostgreSQLLocale" grep -q "ENCODING ''UTF8'' TEMPLATE template0 LOCALE_PROVIDER icu ICU_LOCALE ''und''" "$LOG"
check "PostgreSQLClient" grep -qx "args: --pg postgres://photoprism:photoprism@postgres:4002/postgres -v ON_ERROR_STOP=1 -q" "$LOG"
check "PostgreSQLUserPassword" grep -q "ALTER USER testcafe LOGIN PASSWORD 'testcafe';" "$LOG"
check "PostgreSQLTransfer" grep -q -- "--target-driver postgres --target-server postgres:4002 --target-name testcafe --target-user testcafe password: testcafe" "$LOG"
POSTGRES_ADMIN_URI=postgres://admin@db/postgres POSTGRES_PORT=5432 expect "PostgreSQLEnv" 0 "postgres:5432" postgres "$tmp/photoprism" "$fx/config-sqlite"
check "PostgreSQLEnvClient" grep -qx "args: --pg postgres://admin@db/postgres -v ON_ERROR_STOP=1 -q" "$LOG"

# Without a hub.yml in the fixture, none is created.
nohub=$(fixture nohub)
rm "$nohub/config-sqlite/hub.yml"
expect "NoHub" 0 "Created" mariadb "$tmp/photoprism" "$nohub/config-sqlite"
check "NoHubFile" test ! -e "$nohub/config-mariadb-generated/hub.yml"

# A failed database client stops the script before the transfer.
FAIL_CLIENT=1 expect "ClientFails" 1 "" postgres "$tmp/photoprism" "$fx/config-sqlite"
check "ClientFailsNoTransfer" lacks "database transfer" "$LOG"

# Folders the script did not create are left alone.
other=$(fixture other)
mkdir "$other/config-mariadb-generated"
touch "$other/config-mariadb-generated/keep"
expect "Foreign" 1 "not created by this script" mariadb "$tmp/photoprism" "$other/config-sqlite"
check "ForeignKept" test -f "$other/config-mariadb-generated/keep"
check "ForeignNoClient" test ! -s "$LOG"
mkdir "$tmp/target"
touch "$tmp/target/.testcafe-generated" "$tmp/target/keep"
ln -s "$tmp/target" "$other/config-postgres-generated"
expect "Symlink" 1 "not created by this script" postgres "$tmp/photoprism" "$other/config-sqlite"
check "SymlinkKept" test -f "$tmp/target/keep" -a -L "$other/config-postgres-generated"

# A symlinked fixture config, or options file in it, is neither written nor replaced.
real=$(fixture real)
linked=$(fixture linked)
rm -rf "$linked/config-sqlite"
ln -s "$real/config-sqlite" "$linked/config-sqlite"
expect "SourceSymlink" 1 "not a symbolic link" mariadb "$tmp/photoprism" "$linked/config-sqlite"
check "SourceSymlinkKept" test ! -e "$linked/config-mariadb-generated" -a ! -e "$real/config-sqlite/.testcafe-generated"
shared=$(fixture shared)
mv "$shared/config-sqlite/options.yml" "$tmp/shared-options.yml"
ln -s "$tmp/shared-options.yml" "$shared/config-sqlite/options.yml"
expect "OptionsSymlink" 0 "Created" mariadb "$tmp/photoprism" "$shared/config-sqlite"
check "OptionsSymlinkKept" grep -qx 'DatabaseDriver: sqlite' "$tmp/shared-options.yml"
check "OptionsSymlinkCopied" test ! -L "$shared/config-mariadb-generated/options.yml"

# Database options that would not take effect fail the script and leave no config folder behind.
ended=$(fixture ended)
echo "..." >> "$ended/config-sqlite/options.yml"
expect "SecondDocument" 1 "would not take effect" mariadb "$tmp/photoprism" "$ended/config-sqlite"
started=$(fixture started)
echo "---" >> "$started/config-sqlite/options.yml"
expect "NewDocument" 1 "would not take effect" postgres "$tmp/photoprism" "$started/config-sqlite"
check "NewDocumentNoTarget" test ! -e "$started/config-postgres-generated" -a -z "$(find "$started" -maxdepth 1 -name '.config-*')"
check "SecondDocumentNoTarget" test ! -e "$ended/config-mariadb-generated" -a -z "$(find "$ended" -maxdepth 1 -name '.config-*')"
unloaded=$(fixture unloaded)
sed -i '1i ...' "$unloaded/config-sqlite/options.yml"
expect "SourceNotLoaded" 1 "does not use SQLite" mariadb "$tmp/photoprism" "$unloaded/config-sqlite"
check "SourceNotLoadedNoTransfer" test ! -s "$LOG"
moved=$(fixture moved)
sed -i "s|^StoragePath: .*|StoragePath: $tmp|" "$moved/config-sqlite/options.yml"
expect "SourceStorage" 1 "does not use SQLite" postgres "$tmp/photoprism" "$moved/config-sqlite"

for key in OriginalsPath ImportPath; do
  outside=$(fixture "outside$key")
  sed -i "s|^$key: .*|$key: $tmp/outside|" "$outside/config-sqlite/options.yml"
  expect "Outside ($key)" 1 "originals and import folders" mariadb "$tmp/photoprism" "$outside/config-sqlite"
  sed -i "s|^$key: .*|$key: $outside/../outside|" "$outside/config-sqlite/options.yml"
  expect "Escapes ($key)" 1 "originals and import folders" mariadb "$tmp/photoprism" "$outside/config-sqlite"
  sed -i "s|^$key: .*|$key: $outside-old/folder|" "$outside/config-sqlite/options.yml"
  expect "Sibling ($key)" 1 "originals and import folders" mariadb "$tmp/photoprism" "$outside/config-sqlite"
  sed -i "/^$key: /d" "$outside/config-sqlite/options.yml"
  expect "Missing ($key)" 1 "originals and import folders" mariadb "$tmp/photoprism" "$outside/config-sqlite"
done

for port in 0 65536 "4001 -h x" '5432"'; do
  MARIADB_PORT=$port POSTGRES_PORT=$port expect "Port ($port)" 2 "port numbers" postgres "$tmp/photoprism" "$fx/config-sqlite"
done

# A dangling link in the fixture config fails the copy without creating its target.
dangling=$(fixture dangling)
ln -s "$tmp/created-by-link" "$dangling/config-sqlite/.testcafe-generated"
expect "DanglingLink" 1 "" mariadb "$tmp/photoprism" "$dangling/config-sqlite"
check "DanglingLinkNoTarget" test ! -e "$tmp/created-by-link" -a ! -e "$dangling/config-mariadb-generated"

# rules checks that the restart and stop rules in a Makefile run every command that starts the binary or the scripts
# with the database, folders and port of the fixture in the environment, and whether it unsets the given variables.
# The public restart and stop targets are given, followed by the auth targets they are checked with.
rules() {
  local name=$1 dir=$2 fixture=$3 bin=$4 unset=$5 restart=$6 stop=$7 out v
  out=$(make -s -n -C "$dir" "$restart" "$stop" acceptance-auth-restart acceptance-auth-stop DB=mariadb 2>/dev/null)
  check "$name Rules" test "$(grep -cE "${bin//./\\.}|acceptance-(server|db)\\.sh" <<< "$out")" -eq 10

  for v in $unset; do
    check "$name Unset ($v)" test "$(grep -cF -- " -u PHOTOPRISM_$v " <<< "$out")" -eq 10
  done

  for v in DATABASE_DRIVER=sqlite "DATABASE_DSN=\"$fixture/index.db\"" "STORAGE_PATH=\"$fixture\"" \
    "ORIGINALS_PATH=\"$fixture/originals\"" "IMPORT_PATH=\"$fixture/import\"" HTTP_PORT=2343; do
    check "$name Set ($v)" test "$(grep -cF -- " PHOTOPRISM_$v " <<< "$out")" -eq 10
  done

  check "$name Start" test "$(grep -cE -- " ${bin//./\\.} -.* start -d\$" <<< "$out")" -eq 2
  check "$name StartConfig" test "$(grep -cF -- "-c \"$fixture/config-mariadb-generated\" start -d" <<< "$out")" -eq 2
}

common="DATABASE_SERVER DATABASE_NAME DATABASE_USER DATABASE_PASSWORD SIDECAR_PATH CACHE_PATH USERS_PATH TEMP_PATH"
common+=" BACKUP_PATH PID_FILENAME LOG_FILENAME DEFAULTS_YAML PORTAL_URL JOIN_TOKEN NODE_CLIENT_ID NODE_CLIENT_SECRET"

# pinned checks that a restart rule runs the binary and the scripts with the fixture environment, and starts the
# server with the given flag.
pinned() {
  local name=$1 dir=$2 target=$3 bin=$4 flag=$5 out runs
  out=$(make -s -n -C "$dir" "$target" DB=postgres 2>/dev/null)
  runs=$(grep -E "${bin//./\\.}|acceptance-(server|db)\\.sh" <<< "$out")
  check "$name Pinned ($target)" test -n "$runs" -a "$(grep -cv '^cd \.\. && env -u PHOTOPRISM_DATABASE_SERVER ' <<< "$runs")" -eq 0
  check "$name Flag ($target)" grep -qE -- " ${bin//./\\.} .*$flag.* start -d\$" <<< "$out"
}

# hubfirst checks that a restart rule restores the hub registration before it runs the binary or the scripts.
hubfirst() {
  local name=$1 dir=$2 target=$3 bin=$4 out hub run
  out=$(make -s -n -C "$dir" "$target" DB=postgres 2>/dev/null)
  hub=$(grep -n -m 1 '^hub=' <<< "$out" | cut -d: -f1)
  run=$(grep -nE -m 1 "$bin|acceptance-(server|db)\\.sh" <<< "$out" | cut -d: -f1)
  check "$name HubFirst ($target)" test -n "$hub" -a -n "$run" -a "${hub:-0}" -lt "${run:-0}"
}

if command -v make >/dev/null; then
  [[ ! -f "$root/Makefile" ]] || rules Makefile "$root" ./storage/acceptance ./photoprism "$common CONNECT" \
    acceptance-restart-1 acceptance-stop-1
  if [[ -f "$root/plus/Makefile" ]]; then
    rules PlusMakefile "$root/plus" ./plus/storage/acceptance-plus ./photoprism-plus "$common" \
      acceptance-restart-1 acceptance-stop-1
    check "PlusMakefile KeepsConnect" lacks "-u PHOTOPRISM_CONNECT" "$root/plus/Makefile"
    hubfirst PlusMakefile "$root/plus" acceptance-restart-1 ./photoprism-plus
    hubfirst PlusMakefile "$root/plus" acceptance-auth-restart ./photoprism-plus
  fi

  if [[ -f "$root/pro/Makefile" ]]; then
    rules ProMakefile "$root/pro" ./pro/storage/acceptance-pro ./photoprism-pro "$common" acceptance-restart acceptance-stop
    check "ProMakefile KeepsConnect" lacks "-u PHOTOPRISM_CONNECT" "$root/pro/Makefile"

    for t in acceptance-restart acceptance-auth-restart acceptance-auth-sync-restart acceptance-auth-ldap-restart; do
      hubfirst ProMakefile "$root/pro" "$t" ./photoprism-pro
    done

    pinned ProMakefile "$root/pro" acceptance-auth-sync-restart ./photoprism-pro "--ldap-sync --auth-mode"
    pinned ProMakefile "$root/pro" acceptance-auth-ldap-restart ./photoprism-pro '--ldap-uri "ldap://dummy-ldap:389"'
  fi
fi

exit "$failed"
