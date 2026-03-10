#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WINDOW_LINES=15
API_DIRS=(
  "internal/api"
  "plus/internal/api"
  "pro/internal/api"
  "portal/internal/api"
)

violations=()

check_file() {
  local file="$1"
  local rel="${file#"$ROOT_DIR"/}"
  local output

  output="$(
    awk -v window="$WINDOW_LINES" '
      /^[[:space:]]*\/\// { next }
      /^[[:space:]]*func[[:space:]]/ { last_limit = 0 }
      /LimitRequestBodyBytes[[:space:]]*\(/ { last_limit = NR }
      /c\.(BindJSON|ShouldBindJSON)\(/ {
        if (last_limit == 0 || NR - last_limit > window) {
          printf "%d:%s\n", NR, $0
        }
      }
    ' "$file"
  )"

  if [ -z "$output" ]; then
    return
  fi

  while IFS= read -r line; do
    violations+=("${rel}:${line}")
  done <<< "$output"
}

for dir in "${API_DIRS[@]}"; do
  if [ ! -d "$ROOT_DIR/$dir" ]; then
    continue
  fi

  while IFS= read -r -d '' file; do
    check_file "$file"
  done < <(find "$ROOT_DIR/$dir" -type f -name '*.go' ! -name '*_test.go' -print0)
done

if [ "${#violations[@]}" -gt 0 ]; then
  echo "ERROR: API JSON binding without nearby request-body limit detected:"
  printf '  %s\n' "${violations[@]}"
  echo
  echo "Add LimitRequestBodyBytes(...) before BindJSON(...) / ShouldBindJSON(...),"
  echo "then handle IsRequestBodyTooLarge(err) and AbortRequestTooLarge(...)."
  exit 1
fi

echo "OK: All reviewed API JSON binding sites have nearby request-body limits."
