#!/usr/bin/env bash
set -euo pipefail

# dav-probe.sh captures WebDAV PROPFIND responses for protocol/header inspection.
#
# Usage:
#   scripts/dav-probe.sh <collection-url>
#
# Environment variables:
#   DAV_USERNAME / DAV_PASSWORD  Basic auth credentials (optional)
#   DAV_INSECURE=1              Use curl -k for self-signed local TLS (optional)
#   DAV_OUTPUT_DIR              Output directory (default: ./.local/webdav)

usage() {
  cat <<'EOF'
Usage: scripts/dav-probe.sh <collection-url>

Example:
  DAV_USERNAME=admin DAV_PASSWORD=secret DAV_INSECURE=1 \
    scripts/dav-probe.sh "https://app.localssl.dev/i/acme/originals/"
EOF
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "error: required command not found: $1" >&2
    exit 1
  fi
}

header_value() {
  local file="$1"
  local name="$2"
  awk -F': *' -v wanted="$(printf '%s' "$name" | tr '[:upper:]' '[:lower:]')" '
    {
      key = tolower($1)
      sub(/\r$/, "", key)
      if (key == wanted) {
        val = $0
        sub(/^[^:]*:[[:space:]]*/, "", val)
        sub(/\r$/, "", val)
        print val
        exit 0
      }
    }
  ' "$file"
}

print_capture_summary() {
  local file="$1"
  local status_line

  status_line="$(awk 'NR==1 { sub(/\r$/, "", $0); print; exit }' "$file")"
  printf '\n== %s ==\n' "$(basename "$file")"
  printf 'Status: %s\n' "${status_line:-<missing>}"

  for h in \
    "Date" \
    "DAV" \
    "Allow" \
    "MS-Author-Via" \
    "Content-Type" \
    "Content-Length" \
    "Transfer-Encoding" \
    "ETag" \
    "Content-Security-Policy" \
    "Cross-Origin-Opener-Policy" \
    "Referrer-Policy" \
    "X-Content-Type-Options" \
    "X-Frame-Options" \
    "X-Robots-Tag" \
    "X-XSS-Protection"
  do
    value="$(header_value "$file" "$h" || true)"
    printf '%s: %s\n' "$h" "${value:-<absent>}"
  done

  # Show the first multistatus href entries for quick trailing-slash/encoding checks.
  printf 'Href sample:\n'
  awk '
    BEGIN { count = 0 }
    /<[^>]*href>/ {
      line = $0
      gsub(/\r/, "", line)
      while (match(line, /<[^>]*href>[^<]*<\/[^>]*href>/)) {
        token = substr(line, RSTART, RLENGTH)
        gsub(/<[^>]*href>/, "", token)
        gsub(/<\/[^>]*href>/, "", token)
        print "  " token
        line = substr(line, RSTART + RLENGTH)
        count++
        if (count >= 5) {
          exit 0
        }
      }
    }
  ' "$file" || true
}

main() {
  require_cmd curl
  require_cmd date
  require_cmd awk

  if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
    usage
    exit 0
  fi

  local url="${1:-}"
  if [ -z "$url" ]; then
    usage
    exit 1
  fi

  local out_dir="${DAV_OUTPUT_DIR:-./.local/webdav}"
  mkdir -p "$out_dir"

  local timestamp
  timestamp="$(date -u +%Y%m%dT%H%M%SZ)"

  local -a auth_args=()
  if [ -n "${DAV_USERNAME:-}" ] || [ -n "${DAV_PASSWORD:-}" ]; then
    if [ -z "${DAV_USERNAME:-}" ] || [ -z "${DAV_PASSWORD:-}" ]; then
      echo "error: set both DAV_USERNAME and DAV_PASSWORD, or neither" >&2
      exit 1
    fi
    auth_args=(-u "${DAV_USERNAME}:${DAV_PASSWORD}")
  fi

  local -a tls_args=()
  if [ "${DAV_INSECURE:-0}" = "1" ]; then
    tls_args=(-k)
  fi

  local body
  body='<?xml version="1.0" encoding="utf-8"?><D:propfind xmlns:D="DAV:"><D:allprop/></D:propfind>'

  local http_mode
  local depth
  for http_mode in "--http2" "--http1.1"; do
    for depth in 0 1 infinity; do
      local label mode_name file
      mode_name="${http_mode#--}"
      label="propfind_${mode_name}_depth${depth}"
      file="${out_dir}/${timestamp}_${label}.txt"

      echo "Capturing ${label} -> ${file}"

      curl \
        --silent --show-error \
        "${tls_args[@]}" \
        "${auth_args[@]}" \
        "${http_mode}" \
        -i \
        -X PROPFIND \
        -H "Depth: ${depth}" \
        -H "Content-Type: application/xml; charset=utf-8" \
        --data "${body}" \
        "$url" \
        >"$file"

      print_capture_summary "$file"
    done
  done

  printf '\nArtifacts written to: %s\n' "$out_dir"
}

main "$@"
