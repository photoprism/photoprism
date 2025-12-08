#!/usr/bin/env bash

# Recursively dumps CLI usage information into [command]-cli-commands.txt"
# Usage: ./export-help.sh [command]

set -u -o pipefail

if [[ $# -lt 1 ]]; then
  echo "Usage: $0 <command>" >&2
  exit 1
fi

ROOT_CMD="$1"
OUTFILE="$(basename "$ROOT_CMD")-cli-commands.txt"

# Start with a clean file
: > "$OUTFILE"

# Bash ≥ 4 required for associative arrays.
declare -A VISITED
FIRST_BLOCK=1

crawl() {
  local full_cmd=("$@")
  local key="${full_cmd[*]}"

  # Skip if already processed (handles aliases)
  if [[ -n "${VISITED[$key]+x}" ]]; then
    return
  fi
  VISITED["$key"]=1

  # Capture --help (stdout+stderr)
  local out
  if ! out="$("${full_cmd[@]}" --help 2>&1)"; then
    : # continue even if non-zero exit
  fi

  # Append block
  if (( FIRST_BLOCK )); then
    FIRST_BLOCK=0
    printf "%s\n" "$out" >> "$OUTFILE"
  else
    printf "\n---\n\n%s\n" "$out" >> "$OUTFILE"
  fi

  # Extract subcommand names from COMMANDS: (or "Available Commands:") section(s)
  mapfile -t subs < <(printf "%s\n" "$out" | awk '
    BEGIN { section=0 }
    /^[[:space:]]*(COMMANDS|Available Commands):/ { section=1; next }
    section==1 && /^[A-Z][A-Z ]*:/ { section=0; next }  # next ALL-CAPS header ends the section
    section==1 {
      # Lines look like: "start, up        Starts the Web server"
      s=$0
      sub(/^[[:space:]]+/,"",s)
      if (s ~ /^[[:alnum:]][[:alnum:]-]*/) {
        # First token up to comma/space; skip help aliases
        split(s, a, /[,[:space:]]+/)
        if (a[1] != "" && a[1] != "help" && a[1] != "h") print a[1]
      }
    }
  ' | sort -u)

  # Recurse into each subcommand
  local sub
  for sub in "${subs[@]}"; do
    crawl "${full_cmd[@]}" "$sub"
  done
}

crawl "$ROOT_CMD"

echo "Wrote command help to: $OUTFILE"
