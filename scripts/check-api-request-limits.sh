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

# check_file scans the given Go file for request-body sinks that are not
# preceded within WINDOW_LINES by a LimitRequestBodyBytes call. Three
# sink patterns are covered:
#
#   (1) c.BindJSON / c.ShouldBindJSON — the common Gin binding path.
#   (2) <name>.ServeHTTP(<writer>, c.Request) — delegation to a third
#       party handler (e.g. the MCP SDK) that reads the body internally
#       (via io.ReadAll or a decoder) without a size cap of its own.
#       Handlers that forward raw requests MUST still wrap c.Request.Body
#       with http.MaxBytesReader via api.LimitRequestBodyBytes beforehand
#       so the cap applies before the third-party code reads the body.
#   (3) io.ReadAll(c.Request.Body) / json.NewDecoder(c.Request.Body) —
#       direct unbounded reads that bypass both Gin's binding helpers
#       and the SDK-delegation pattern. Same remediation applies: wrap
#       c.Request.Body via LimitRequestBodyBytes before the call.
#
# The three sink patterns share the same fifteen-line window rule so the
# regression guard stays uniform regardless of how a handler ingests
# its body.
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
      /[A-Za-z_][A-Za-z0-9_]*\.ServeHTTP[[:space:]]*\([^,]*,[[:space:]]*c\.Request[[:space:]]*\)/ {
        if (last_limit == 0 || NR - last_limit > window) {
          printf "%d:%s\n", NR, $0
        }
      }
      /(io\.ReadAll|io\.LimitReader|json\.NewDecoder|xml\.NewDecoder|yaml\.NewDecoder)[[:space:]]*\([[:space:]]*c\.Request\.Body[[:space:]]*\)/ {
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
  echo "ERROR: API request-body sink without nearby LimitRequestBodyBytes detected:"
  printf '  %s\n' "${violations[@]}"
  echo
  echo "Add LimitRequestBodyBytes(...) before one of:"
  echo "  * c.BindJSON(...) / c.ShouldBindJSON(...)"
  echo "  * <handler>.ServeHTTP(<writer>, c.Request)"
  echo "  * io.ReadAll(c.Request.Body) / json.NewDecoder(c.Request.Body) / ..."
  echo
  echo "On the BindJSON path, handle IsRequestBodyTooLarge(err) and"
  echo "AbortRequestTooLarge(...). For SDK-delegated or direct-read paths,"
  echo "wrap the response writer to rewrite the upstream 400 -> 413 so the"
  echo "response stays consistent with the rest of the JSON API."
  exit 1
fi

echo "OK: All reviewed API request-body sinks have nearby LimitRequestBodyBytes calls."
