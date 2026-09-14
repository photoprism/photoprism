#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API_DIRS=(
  "internal/api"
  "internal/mcp"
  "internal/server"
  "plus/internal/api"
  "pro/internal/api"
  "portal/internal/api"
)

# Helpers whose caller sets the bound before invoking them, so a function-local
# check cannot see it.
EXCLUDE_PATTERNS=(
  "portal/internal/api/oidc_logout.go"
  "portal/internal/api/oidc_token.go"
)

excluded() {
  local rel="$1"
  local pattern

  for pattern in "${EXCLUDE_PATTERNS[@]}"; do
    case "$rel" in
      *"$pattern") return 0 ;;
    esac
  done

  return 1
}

violations=()

# check_file flags request-body sinks in a function that never calls
# LimitRequestBodyBytes. Form parsing and the content-type-driven binds reach the
# body only on a write method, so those are flagged per enclosing route.
check_file() {
  local file="$1"
  local rel="${file#"$ROOT_DIR"/}"
  local output

  if excluded "$rel"; then
    return
  fi

  output="$(
    awk '
      /^[[:space:]]*\/\// { next }
      /^[[:space:]]*func[[:space:]]/ { last_limit = 0 }
      /router\.(GET|HEAD|OPTIONS)[[:space:]]*\(/ { reads_body = 0 }
      /router\.(POST|PUT|PATCH|DELETE)[[:space:]]*\(/ { reads_body = 1 }
      /LimitRequestBodyBytes[[:space:]]*\(/ { last_limit = NR }
      /c\.(BindJSON|BindXML|BindYAML|ShouldBindJSON|ShouldBindXML|ShouldBindYAML|ShouldBindBodyWith)\(/ {
        if (last_limit == 0) {
          printf "%d:%s\n", NR, $0
        }
      }
      /c\.(Bind|BindWith|MustBindWith|ShouldBind|ShouldBindWith)\(/ {
        if (last_limit == 0 && reads_body) {
          printf "%d:%s\n", NR, $0
        }
      }
      /(c\.(PostForm|PostFormArray|PostFormMap|FormFile|MultipartForm|SaveUploadedFile)\(|c\.Request\.(ParseForm|ParseMultipartForm)\()/ {
        if (last_limit == 0 && reads_body) {
          printf "%d:%s\n", NR, $0
        }
      }
      /[A-Za-z_][A-Za-z0-9_]*\.ServeHTTP[[:space:]]*\([^,]*,[[:space:]]*c\.Request[[:space:]]*\)/ {
        if (last_limit == 0) {
          printf "%d:%s\n", NR, $0
        }
      }
      /(io\.ReadAll|io\.LimitReader|json\.NewDecoder|xml\.NewDecoder|yaml\.NewDecoder)[[:space:]]*\([[:space:]]*c\.Request\.Body[[:space:]]*\)/ {
        if (last_limit == 0) {
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
