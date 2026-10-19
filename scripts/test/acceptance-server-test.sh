#!/usr/bin/env bash

# Tests acceptance-server.sh with scratch configs, stand-in processes, and listeners on a free local port.
set -uo pipefail

if [[ ! -r /proc/self/cmdline ]] || ! command -v lsof python3 >/dev/null || (( BASH_VERSINFO[0] * 100 + BASH_VERSINFO[1] < 403 )); then
  echo "Skipping $0: requires /proc, lsof, python3, and bash 4.3 or later."
  exit 0
fi

script="$(cd "$(dirname "$0")" && pwd)/acceptance-server.sh"
tmp=$(mktemp -d) || exit 1
port=$(python3 -I -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1])')
pids=()
failed=0

# cleanup stops the stand-in processes and removes the scratch folder.
# shellcheck disable=SC2329
cleanup() {
  for p in "${pids[@]}"; do kill "$p" 2>/dev/null || true; done
  rm -rf "$tmp"
}
trap cleanup EXIT

# config writes an options file with the given HttpPort line and returns its folder.
config() {
  mkdir -p "$tmp/$1"
  printf '%s\nStoragePath: %s\n' "$2" "$tmp/$1" > "$tmp/$1/options.yml"
  echo "$tmp/$1"
}

# listener accepts connections on the loopback port given as argument and serves nothing.
listener='import socket, sys, time; s = socket.socket(); s.bind(("127.0.0.1", int(sys.argv[1]))); s.listen(); time.sleep(60)'

# listen starts a stand-in server on the port.
listen() {
  standin "$1" python3 -I -c "$listener" "$port"

  for _ in $(seq 50); do (exec 3<>"/dev/tcp/localhost/$port") 2>/dev/null && return; sleep 0.1; done
}

# standin starts a process named photoprism-test that runs the given command, and writes its PID file in the folder.
standin() {
  local dir=$1
  shift
  (exec -a photoprism-test "$@" >/dev/null 2>&1) &
  pids+=($!)
  echo "$!" > "$dir/photoprism.pid"

  for _ in $(seq 50); do grep -qa photoprism-test "/proc/$!/cmdline" 2>/dev/null && return; sleep 0.1; done
}

# expect runs the script and checks its exit status and output.
expect() {
  local name=$1 want=$2 pattern=$3 out status
  shift 3
  out=$(ACCEPTANCE_SERVER_TIMEOUT=${TIMEOUT:-3} bash "$script" "$@" 2>&1)
  status=$?

  if [[ "$status" != "$want" ]] || ! grep -q -- "$pattern" <<< "$out"; then
    echo "FAIL $name: exit $status, want $want and \"$pattern\": $out"
    failed=1
  else
    echo "ok   $name"
  fi
}

free=$(config free "HttpPort: $port")
expect "StopFree" 0 "Port $port is free" stop "$free" true
expect "StartedNotRunning" 1 "is not running" started "$free"

for spelling in "HttpPort: $port" "HttpPort: \"$port\"" "HttpPort: '$port' # comment" "HttpPort: $port   " $'HttpPort: '"$port"$'\r'; do
  dir=$(config "spelling$RANDOM" "$spelling")
  listen "$dir"
  expect "Started ($spelling)" 0 "listens on port $port" started "$dir"
  rm -f "$dir/photoprism.pid"
  expect "StopForeign ($spelling)" 1 "in use by" stop "$dir" true
  kill "${pids[-1]}"
  wait "${pids[-1]}" 2>/dev/null
done

# Without HttpPort, the port is taken from the environment.
fallback=$(config fallback "")
listen "$fallback"
PHOTOPRISM_HTTP_PORT=$port expect "MissingPortKey" 0 "listens on port $port" started "$fallback"
kill "${pids[-1]}"
wait "${pids[-1]}" 2>/dev/null

# A stand-in server is running, but another process holds the port.
foreign=$(config foreign "HttpPort: $port")
standin "$foreign" sleep 30
(exec python3 -I -c "$listener" "$port") &
pids+=($!)
for _ in $(seq 50); do (exec 3<>"/dev/tcp/localhost/$port") 2>/dev/null && break; sleep 0.1; done
expect "StartedForeign" 1 "alone" started "$foreign"
kill "${pids[-1]}" "${pids[-2]}"
wait "${pids[-1]}" "${pids[-2]}" 2>/dev/null

# A PID file that names a process other than PhotoPrism is ignored.
stale=$(config stale "HttpPort: $port")
sleep 30 &
pids+=($!)
echo "$!" > "$stale/photoprism.pid"
TIMEOUT=10 expect "StaleForeignPid" 0 "Port $port is free" stop "$stale" true

# An empty PID file, as the stop command leaves it, means that no server is running.
empty=$(config empty "HttpPort: $port")
: > "$empty/photoprism.pid"
expect "StartedEmptyPid" 1 "is not running" started "$empty"
expect "StopEmptyPid" 0 "Port $port is free" stop "$empty" true
listen "$empty"
: > "$empty/photoprism.pid"
expect "StopEmptyPidForeign" 1 "in use by" stop "$empty" true
kill "${pids[-1]}"
wait "${pids[-1]}" 2>/dev/null

for spelling in "HttpPort: 0" "HttpPort: 08" "HttpPort: 65536" "HttpPort: 80abc"; do
  expect "Invalid ($spelling)" 1 "invalid HttpPort" stop "$(config "invalid$RANDOM" "$spelling")" true
done

TIMEOUT=abc expect "InvalidTimeout" 2 "must be a number" stop "$free" true
expect "Usage" 2 "Usage" stop "$free"

# The stand-in stop command empties the PID file at once, and the process exits two seconds later.
slow=$(config slow "HttpPort: $port")
standin "$slow" sleep 2
expect "StopWaitsForExit" 0 "Port $port is free" stop "$slow" truncate -s 0 "$slow/photoprism.pid"

if kill -0 "${pids[-1]}" 2>/dev/null; then
  echo "FAIL StopWaitsForExit: returned before the process exited"
  failed=1
fi

# A process that does not exit is reported by its ID.
stuck=$(config stuck "HttpPort: $port")
standin "$stuck" sleep 30
TIMEOUT=1 expect "StopStuck" 1 "has not exited" stop "$stuck" true

exit "$failed"
