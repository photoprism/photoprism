#!/usr/bin/env bash

# Reports elapsed wall time and preserves the wrapped command's exit status.
set -u

if (( $# < 2 )); then
  echo "Usage: $0 <label> <command> [args...]" >&2
  exit 2
fi

label=$1
shift
SECONDS=0
"$@"
status=$?
elapsed=$SECONDS
printf '\n%s: elapsed %dh %02dm %02ds (%ds)' \
  "$label" "$((elapsed / 3600))" "$((elapsed / 60 % 60))" \
  "$((elapsed % 60))" "$elapsed" >&2
if (( status != 0 )); then
  printf ', exit status %d' "$status" >&2
fi
printf '\n' >&2
exit "$status"
