#!/usr/bin/env bash

# Verifies that every target named in a Makefile "HELP_TEXT" block exists in that Makefile,
# so that renaming or removing a target does not silently leave "make help" advertising it.
#
# Usage: scripts/check-make-help.sh [makefile ...]
#
# Without arguments, all Makefiles that define a HELP_TEXT block are checked, including the
# ones in optional private subdirectories when they are present in this working copy.

set -euo pipefail

# Lists the Makefiles to check when none are passed as arguments.
default_files() {
  local file
  for file in Makefile frontend/Makefile scripts/dist/Makefile setup/Makefile \
    setup/docker/Makefile setup/podman/Makefile \
    specs/Makefile plus/Makefile pro/Makefile portal/Makefile; do
    [ -f "$file" ] && echo "$file"
  done
}

# Prints the names of all targets declared in a Makefile, ignoring define blocks.
declared_targets() {
  awk '
    /^define /  { skip = 1 }
    /^endef/    { skip = 0; next }
    skip        { next }
    /^[[:alnum:]][^[:space:]]*:/ { sub(/:.*/, "", $1); print $1 }
  ' "$1"
}

# Prints the target names advertised in the HELP_TEXT block of a Makefile.
# Indented "<name>  <description>" entries and comma-separated "Related:" lists are both read.
advertised_targets() {
  awk '
    /^define HELP_TEXT/ { inside = 1; next }
    /^endef/            { inside = 0 }
    !inside             { next }
    /^  [A-Za-z][A-Za-z0-9._\/-]*  +/ { print $1; next }
    /^  [A-Z][A-Za-z ]*:/ {
      sub(/^[^:]*:/, "")
      gsub(/,/, " ")
      for (i = 1; i <= NF; i++) print $i
    }
  ' "$1"
}

status=0
checked=0

if [ "$#" -gt 0 ]; then
  files=("$@")
else
  mapfile -t files < <(default_files)
fi

for file in "${files[@]}"; do
  if [ ! -f "$file" ]; then
    echo "ERROR: $file does not exist" >&2
    status=1
    continue
  fi

  if ! grep -q '^define HELP_TEXT' "$file"; then
    echo "ERROR: $file does not define a HELP_TEXT block" >&2
    status=1
    continue
  fi

  targets=$(declared_targets "$file")
  missing=""

  while read -r name; do
    [ -n "$name" ] || continue
    if ! printf '%s\n' "$targets" | grep -qxF "$name"; then
      missing="$missing $name"
    fi
  done < <(advertised_targets "$file")

  if [ -n "$missing" ]; then
    echo "ERROR: $file advertises targets that do not exist:$missing" >&2
    status=1
  fi

  checked=$((checked + 1))
done

if [ "$status" -eq 0 ]; then
  echo "OK: $checked Makefiles advertise only targets that exist."
fi

exit "$status"
