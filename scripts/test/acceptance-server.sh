#!/usr/bin/env bash

# Stops and checks the server of a TestCafe acceptance config: "stop" runs the given stop command and waits until
# the process from the PID file has exited and the HTTP port is free, "started" waits until that process alone listens.
set -euo pipefail

if (( $# < 2 )) || [[ "$1" != "stop" && "$1" != "started" ]] || [[ "$1" == "stop" && $# -lt 3 ]]; then
  echo "Usage: $0 stop <config path> <stop command...> | started <config path>" >&2
  exit 2
fi

mode=$1
options=${2%/}/options.yml
shift 2
timeout=${ACCEPTANCE_SERVER_TIMEOUT:-60}

if [[ ! "$timeout" =~ ^[1-9][0-9]{0,3}$ ]]; then
  echo "$0: ACCEPTANCE_SERVER_TIMEOUT must be a number of seconds" >&2
  exit 2
elif [[ ! -f "$options" ]]; then
  echo "$0: $options not found" >&2
  exit 1
elif ! command -v lsof >/dev/null; then
  echo "$0: lsof is required" >&2
  exit 1
fi

# option prints the first value of a top-level key in the options file, without quotes and comments.
option() {
  sed -n "s/^$1:[[:space:]]*//p" "$options" | head -n 1 | tr -d '\r' | sed -e 's/[[:space:]]#.*$//' \
    -e 's/[[:space:]]*$//' -e 's/^"\(.*\)"$/\1/' -e "s/^'\(.*\)'\$/\1/"
}

port=$(option HttpPort)
port=${port:-${PHOTOPRISM_HTTP_PORT:-2342}}
storage=$(option StoragePath)
storage=${storage:-${PHOTOPRISM_STORAGE_PATH:-storage}}
pidfile=$storage/photoprism.pid

if [[ ! "$port" =~ ^[1-9][0-9]{0,4}$ ]] || (( port > 65535 )); then
  echo "$0: invalid HttpPort \"$port\" in $options" >&2
  exit 1
fi

# listening reports whether a server accepts connections on the port.
listening() {
  (exec 3<>"/dev/tcp/localhost/$port") 2>/dev/null
}

# alive reports whether the process ID belongs to a running process whose program name starts with "photoprism".
alive() {
  local name

  [[ -n "$1" ]] && kill -0 "$1" 2>/dev/null || return 1
  name=$(tr '\0' '\n' 2>/dev/null < "/proc/$1/cmdline" | head -n 1)
  [[ "${name##*/}" == photoprism* ]]
}

# pid prints the process ID in the PID file if it belongs to a running PhotoPrism process. An empty PID file means that
# no server is running.
pid() {
  local p=""

  if [[ -f "$pidfile" && -s "$pidfile" ]]; then
    p=$(tr -dc '0-9' < "$pidfile")
  fi

  if alive "$p"; then
    echo "$p"
  fi
}

# listeners prints the IDs of the processes that listen on the port.
listeners() {
  lsof -t -nP -iTCP:"$port" -sTCP:LISTEN 2>/dev/null | sort -u || true
}

# holder describes the processes that listen on the port.
holder() {
  lsof -nP -iTCP:"$port" -sTCP:LISTEN 2>/dev/null | tail -n +2 | awk '{print $1 " (PID " $2 ")"}' | sort -u | paste -sd ',' - || true
}

if [[ "$mode" == "stop" ]]; then
  # The stop command usually empties the PID file once the process has exited, so its ID is read before stopping.
  old=$(pid)
  "$@"

  # Without a server process to wait for, a port that is in use belongs to another server.
  if [[ -z "$old" ]] && listening; then
    echo "$0: port $port is in use by $(holder), stop that server first" >&2
    exit 1
  fi

  for (( i = 0; i < timeout * 2; i++ )); do
    if ! alive "$old" && ! listening; then
      echo "Port $port is free."
      exit 0
    fi

    sleep 0.5
  done

  if alive "$old"; then
    echo "$0: server process $old has not exited" >&2
  else
    echo "$0: port $port is still in use by $(holder), stop that server first" >&2
  fi

  exit 1
fi

for (( i = 0; i < timeout * 2; i++ )); do
  p=$(pid)

  if [[ -z "$p" ]]; then
    echo "$0: the server in $pidfile is not running, see $storage/photoprism.log" >&2
    exit 1
  elif listening && [[ "$(listeners)" == "$p" ]]; then
    echo "Server $p listens on port $port."
    exit 0
  fi

  sleep 0.5
done

echo "$0: server $p does not listen on port $port alone (listening: $(holder))" >&2
exit 1
