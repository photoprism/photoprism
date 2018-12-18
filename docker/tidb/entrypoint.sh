#!/bin/bash
set -eo pipefail
shopt -s nullglob

# if command starts with an option, prepend tidb-server
if [ "${1:0:1}" = '-' ]; then
	set -- tidb-server "$@"
fi

# skip setup if they want an option that stops tidb-server
wantHelp=
for arg; do
	case "$arg" in
		-'?'|--help|--print-defaults|-V|--version)
			wantHelp=1
			break
			;;
	esac
done

# usage: file_env VAR [DEFAULT]
#    ie: file_env 'XYZ_DB_PASSWORD' 'example'
# (will allow for "$XYZ_DB_PASSWORD_FILE" to fill in the value of
#  "$XYZ_DB_PASSWORD" from a file, especially for Docker's secrets feature)
file_env() {
	local var="$1"
	local fileVar="${var}_FILE"
	local def="${2:-}"
	if [ "${!var:-}" ] && [ "${!fileVar:-}" ]; then
		echo >&2 "error: both $var and $fileVar are set (but are exclusive)"
		exit 1
	fi
	local val="$def"
	if [ "${!var:-}" ]; then
		val="${!var}"
	elif [ "${!fileVar:-}" ]; then
		val="$(< "${!fileVar}")"
	fi
	export "$var"="$val"
	unset "$fileVar"
}

# usage: process_init_file FILENAME MYSQLCOMMAND...
#    ie: process_init_file foo.sh mysql -uroot
# (process a single initializer file, based on its extension. we define this
# function here, so that initializer scripts (*.sh) can use the same logic,
# potentially recursively, or override the logic used in subsequent calls)
process_init_file() {
	local f="$1"; shift
	local mysql=( "$@" )

	case "$f" in
		*.sh)     echo "$0: running $f"; . "$f" ;;
		*.sql)    echo "$0: running $f"; "${mysql[@]}" < "$f"; echo ;;
		*.sql.gz) echo "$0: running $f"; gunzip -c "$f" | "${mysql[@]}"; echo ;;
		*)        echo "$0: ignoring $f" ;;
	esac
	echo
}

_check_config() {
	toRun=( "$@" --verbose --help )
	if ! errors="$("${toRun[@]}" 2>&1 >/dev/null)"; then
		cat >&2 <<-EOM

			ERROR: tidb-server failed while attempting to check config
			command was: "${toRun[*]}"

			$errors
		EOM
		exit 1
	fi
}

# Fetch value from server config
# We use tidb-server --verbose --help instead of my_print_defaults because the
# latter only show values present in config files, and not server defaults
_get_config() {
	local conf="$1"; shift
	"$@" --verbose --help --log-bin-index="$(mktemp -u)" 2>/dev/null \
		| awk '$1 == "'"$conf"'" && /^[^ \t]/ { sub(/^[^ \t]+[ \t]+/, ""); print; exit }'
	# match "datadir      /some/path with/spaces in/it here" but not "--xyz=abc\n     datadir (xyz)"
}

# allow the container to be started with `--user`
if [ "$1" = 'tidb-server' -a -z "$wantHelp" -a "$(id -u)" = '0' ]; then
	_check_config "$@"
	DATADIR="$(_get_config 'datadir' "$@")"
	mkdir -p "$DATADIR"
	chown -R mysql:mysql "$DATADIR"
	exec mysql "$BASH_SOURCE" "$@"
fi

if [ "$1" = 'tidb-server' -a -z "$wantHelp" ]; then
	# still need to check config, container may have started with --user
	_check_config "$@"
	# Get config
	DATADIR="$(_get_config 'datadir' "$@")"

	if [ ! -d "$DATADIR/mysql" ]; then
		file_env 'TIDB_ROOT_PASSWORD'
		if [ -z "$TIDB_ROOT_PASSWORD" -a -z "$TIDB_ALLOW_EMPTY_PASSWORD" -a -z "$TIDB_RANDOM_ROOT_PASSWORD" ]; then
			echo >&2 'error: database is uninitialized and password option is not specified '
			echo >&2 '  You need to specify one of TIDB_ROOT_PASSWORD, TIDB_ALLOW_EMPTY_PASSWORD and TIDB_RANDOM_ROOT_PASSWORD'
			exit 1
		fi

		mkdir -p "$DATADIR"

		echo 'Initializing database'
		"$@" --initialize-insecure
		echo 'Database initialized'

		if command -v mysql_ssl_rsa_setup > /dev/null && [ ! -e "$DATADIR/server-key.pem" ]; then
			# https://github.com/mysql/mysql-server/blob/23032807537d8dd8ee4ec1c4d40f0633cd4e12f9/packaging/deb-in/extra/mysql-systemd-start#L81-L84
			echo 'Initializing certificates'
			mysql_ssl_rsa_setup --datadir="$DATADIR"
			echo 'Certificates initialized'
		fi

		SOCKET="$(_get_config 'socket' "$@")"
		"$@" --socket="${SOCKET}" &
		pid="$!"

		mysql=( mysql --protocol=socket -uroot -hlocalhost --socket="${SOCKET}" )

		for i in {30..0}; do
			if echo 'SELECT 1' | "${mysql[@]}" &> /dev/null; then
				break
			fi
			echo 'MySQL init process in progress...'
			sleep 1
		done
		if [ "$i" = 0 ]; then
			echo >&2 'MySQL init process failed.'
			exit 1
		fi

		if [ -z "$TIDB_INITDB_SKIP_TZINFO" ]; then
			# sed is for https://bugs.mysql.com/bug.php?id=20545
			mysql_tzinfo_to_sql /usr/share/zoneinfo | sed 's/Local time zone must be set--see zic manual page/FCTY/' | "${mysql[@]}" mysql
		fi

		if [ ! -z "$TIDB_RANDOM_ROOT_PASSWORD" ]; then
			export TIDB_ROOT_PASSWORD="$(pwgen -1 32)"
			echo "GENERATED ROOT PASSWORD: $TIDB_ROOT_PASSWORD"
		fi

		rootCreate=
		# default root to listen for connections from anywhere
		file_env 'TIDB_ROOT_HOST' '%'
		if [ ! -z "$TIDB_ROOT_HOST" -a "$TIDB_ROOT_HOST" != 'localhost' ]; then
			# no, we don't care if read finds a terminating character in this heredoc
			# https://unix.stackexchange.com/questions/265149/why-is-set-o-errexit-breaking-this-read-heredoc-expression/265151#265151
			read -r -d '' rootCreate <<-EOSQL || true
				CREATE USER 'root'@'${TIDB_ROOT_HOST}' IDENTIFIED BY '${TIDB_ROOT_PASSWORD}' ;
				GRANT ALL ON *.* TO 'root'@'${TIDB_ROOT_HOST}' WITH GRANT OPTION ;
			EOSQL
		fi

		"${mysql[@]}" <<-EOSQL
			-- What's done in this file shouldn't be replicated
			--  or products like mysql-fabric won't work
			SET @@SESSION.SQL_LOG_BIN=0;

			ALTER USER 'root'@'localhost' IDENTIFIED BY '${TIDB_ROOT_PASSWORD}' ;
			GRANT ALL ON *.* TO 'root'@'localhost' WITH GRANT OPTION ;
			${rootCreate}
			DROP DATABASE IF EXISTS test ;
			FLUSH PRIVILEGES ;
		EOSQL

		if [ ! -z "$TIDB_ROOT_PASSWORD" ]; then
			mysql+=( -p"${TIDB_ROOT_PASSWORD}" )
		fi

		file_env 'TIDB_DATABASE'
		if [ "$TIDB_DATABASE" ]; then
			echo "CREATE DATABASE IF NOT EXISTS \`$TIDB_DATABASE\` ;" | "${mysql[@]}"
			mysql+=( "$TIDB_DATABASE" )
		fi

		file_env 'TIDB_USER'
		file_env 'TIDB_PASSWORD'
		if [ "$TIDB_USER" -a "$TIDB_PASSWORD" ]; then
			echo "CREATE USER '$TIDB_USER'@'%' IDENTIFIED BY '$TIDB_PASSWORD' ;" | "${mysql[@]}"

			if [ "$TIDB_DATABASE" ]; then
				echo "GRANT ALL ON \`$TIDB_DATABASE\`.* TO '$TIDB_USER'@'%' ;" | "${mysql[@]}"
			fi

			echo 'FLUSH PRIVILEGES ;' | "${mysql[@]}"
		fi

		echo
		ls /docker-entrypoint-initdb.d/ > /dev/null
		for f in /docker-entrypoint-initdb.d/*; do
			process_init_file "$f" "${mysql[@]}"
		done

		if [ ! -z "$TIDB_ONETIME_PASSWORD" ]; then
			"${mysql[@]}" <<-EOSQL
				ALTER USER 'root'@'%' PASSWORD EXPIRE;
			EOSQL
		fi
		if ! kill -s TERM "$pid" || ! wait "$pid"; then
			echo >&2 'MySQL init process failed.'
			exit 1
		fi

		echo
		echo 'MySQL init process done. Ready for start up.'
		echo
	fi
fi

exec "$@"