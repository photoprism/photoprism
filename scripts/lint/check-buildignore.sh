#!/usr/bin/env bash

# Checks that the rsync filter rules in assets/.buildignore let "make install" copy exactly the
# files of the given models plus models/NOTICE, by applying them to a scratch tree with every
# registry model and decoys.
#
# Usage: check-buildignore.sh <buildignore> <registry> <model>...

set -euo pipefail

if [[ $# -lt 3 ]]; then
  echo "Usage: ${0##*/} <buildignore> <registry> <model>..." >&2
  exit 2
fi

BUILDIGNORE="$1"
REGISTRY="$2"
shift 2

fail() {
  echo "${BUILDIGNORE}: $*" >&2
  exit 1
}

[[ -f "$BUILDIGNORE" ]] || fail "file not found"
[[ -f "$REGISTRY" ]] || fail "model registry ${REGISTRY} not found"

# rsync clears the whole filter list on a "!" line, so it would silently ship every model.
if grep -q '^!' "$BUILDIGNORE"; then
  fail "'!' is not supported, list each model as '+ /models/<name>/' before '- /models/*'"
fi

# Registry entries have the fields name|url|fallback|sha256|type|dir|file; the last one ends the quoted list.
entries="$(awk -F'|' 'NF == 7 && $1 ~ /^[a-z0-9_-]+$/ { sub(/"$/, "", $7); print $1 "|" $5 "|" $6 "|" $7 }' "$REGISTRY")"
[[ -n "$entries" ]] || fail "no models found in ${REGISTRY}"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/src/models" "$tmp/dst"

# entryFiles prints the files a registry entry installs, relative to the assets directory.
entryFiles() {
  local type="$1" dir="$2" file="$3"

  if [[ "$type" == "zip" ]]; then
    printf 'models/%s/saved_model.pb\nmodels/%s/variables/variables.index\n' "$dir" "$dir"
  else
    printf 'models/%s/%s\n' "$dir" "$file"
  fi
}

# Creates the files of every registry model, so that rules naming them are applied as in a build.
while IFS='|' read -r _ type dir file; do
  for part in "$dir" "$file"; do
    if [[ -n "$part" && ( ! "$part" =~ ^[A-Za-z0-9._-]+$ || "$part" == "." || "$part" == ".." ) ]]; then
      fail "invalid model path ${part} in ${REGISTRY}"
    fi
  done
  while IFS= read -r path; do
    mkdir -p "$tmp/src/$(dirname "$path")"
    touch "$tmp/src/$path"
  done < <(entryFiles "$type" "$dir" "$file")
done <<< "$entries"

# Adds retired, license-gated, and unknown model directories as well as a loose file.
for decoy in nasnet nsfw arcface scrfd "decoy-$RANDOM"; do
  mkdir -p "$tmp/src/models/$decoy"
  touch "$tmp/src/models/$decoy/model.onnx"
done
touch "$tmp/src/models/loose.onnx"

# Adds the notice that every build ships with the bundled models.
touch "$tmp/src/models/NOTICE"

# Lists the files the bundled models must provide.
expected=""
for model in "$@"; do
  entry="$(awk -F'|' -v name="$model" '$1 == name { print; exit }' <<< "$entries")"
  [[ -n "$entry" ]] || fail "model ${model} is not in ${REGISTRY}"
  IFS='|' read -r _ type dir file <<< "$entry"
  expected+="$(entryFiles "$type" "$dir" "$file")"$'\n'
done
expected+="models/NOTICE"$'\n'
expected="$(printf '%s' "$expected" | sort -u)"

copied="$(rsync -r -l --safe-links --dry-run --out-format='%n' --exclude-from="$BUILDIGNORE" "$tmp/src/" "$tmp/dst")"
actual="$(grep '^models/.*[^/]$' <<< "$copied" | sort -u || true)"

if [[ "$expected" != "$actual" ]]; then
  echo "expected files: ${expected//$'\n'/ }" >&2
  echo "copied files: ${actual//$'\n'/ }" >&2
  fail "rules do not copy exactly the bundled models and NOTICE, list each as '+ /models/<name>/' and '+ /models/NOTICE' before '- /models/*'"
fi

echo "OK: ${BUILDIGNORE} copies exactly the bundled models and NOTICE."
