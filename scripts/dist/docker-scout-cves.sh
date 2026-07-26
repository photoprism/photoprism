#!/usr/bin/env bash

# docker-scout-cves.sh runs Docker Scout CVE analysis on PhotoPrism published
# images and strips packages whose vulnerabilities are inert in a container
# (kernel packages, which a container never executes — it uses the host kernel).
#
# The raw docker scout output for a typical build contains ~17 CRITICAL and
# ~176 HIGH findings, almost all from the linux package. Filtering it out
# leaves a report that is actually actionable.
#
# Run this on ci, where docker scout is installed (v1.0.9+) and the photoprismci
# Docker Hub credentials are present in ~/.docker/config.json.
#
# Usage:
#   scripts/dist/docker-scout-cves.sh [--save] [--raw] [image ...]
#
# Flags:
#   --save   write per-image JSON to ~/logs/docker-scout-cves-YYYYMMDD.json
#   --raw    show the unfiltered docker scout output instead of the clean table
#
# With no image arguments the script scans the current published images.
#
# Examples:
#   scripts/dist/docker-scout-cves.sh
#   scripts/dist/docker-scout-cves.sh --save registry://photoprism/photoprism:preview
#   ssh ci 'bash -s' < scripts/dist/docker-scout-cves.sh

set -euo pipefail

# ---------------------------------------------------------------------------
# Packages whose CVEs are inert inside a container because the container uses
# the HOST kernel, not its own kernel package.  Any package whose name matches
# one of these extended-regex patterns is excluded from the output.
# ---------------------------------------------------------------------------
readonly SKIP_PACKAGES=(
  "^linux$"
  "^linux-"
)

# ---------------------------------------------------------------------------
# Default images to scan when none are given.
# ---------------------------------------------------------------------------
readonly DEFAULT_IMAGES=(
  "registry://photoprism/photoprism:ce"
  "registry://photoprism/photoprism:plus"
  "registry://photoprism/photoprism:preview"
)

# ---------------------------------------------------------------------------
# Parse arguments.
# ---------------------------------------------------------------------------
SAVE=false
RAW=false
IMAGES=()

for arg in "$@"; do
  case "$arg" in
    --save) SAVE=true ;;
    --raw)  RAW=true ;;
    -*)     echo "Unknown flag: $arg" >&2; exit 1 ;;
    *)      IMAGES+=("$arg") ;;
  esac
done

if [[ ${#IMAGES[@]} -eq 0 ]]; then
  IMAGES=("${DEFAULT_IMAGES[@]}")
fi

# ---------------------------------------------------------------------------
# Dependency checks.
# ---------------------------------------------------------------------------
if ! command -v docker >/dev/null 2>&1; then
  echo "FATAL: docker CLI not found — run this on ci, not inside the dev container." >&2
  exit 1
fi

if ! docker scout version >/dev/null 2>&1; then
  echo "FATAL: docker scout plugin not found." >&2
  echo "Install with: curl -sSfL https://raw.githubusercontent.com/docker/scout-cli/main/install.sh | sh -s --" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "FATAL: jq not found — install with: apt-get install -y jq" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# Build the jq skip filter from SKIP_PACKAGES.
# ---------------------------------------------------------------------------
# The docker scout cves --format json output groups vulnerabilities by package:
#
#   { "packages": [
#       { "package": { "name": "...", "version": "...", "type": "deb" },
#         "vulnerabilities": [ { "id": "CVE-...", "severity": "HIGH", ... }, ... ]
#       }, ...
#   ] }
#
# The filter below excludes any package whose name matches a skip pattern,
# then counts C/H/M/L per remaining package and prints a sorted table.
# ---------------------------------------------------------------------------
JQ_SKIP_CONDITION=""
for pattern in "${SKIP_PACKAGES[@]}"; do
  if [[ -n "$JQ_SKIP_CONDITION" ]]; then
    JQ_SKIP_CONDITION+=" or "
  fi
  # Escape the pattern for use inside a jq string (only backslash needs escaping
  # because we do not use any other jq string metacharacters here).
  escaped="${pattern//\\/\\\\}"
  JQ_SKIP_CONDITION+="(.package.name // .name | test(\"${escaped}\"))"
done

# jq expression: filter packages, count severities, sort by descending severity.
# Handles both { "package": { "name": ... } } and { "name": ... } structures
# so it works across minor docker scout JSON schema variations.
JQ_FILTER=$(cat <<JQEOF
def pkg_name: (.package.name // .name // "unknown");
def pkg_version: (.package.version // .version // "");
def pkg_type: (.package.type // .type // "");

def severity_counts(pkgs):
  {
    critical: (pkgs | map(.vulnerabilities // [] | map(select(.severity == "CRITICAL"))) | flatten | length),
    high:     (pkgs | map(.vulnerabilities // [] | map(select(.severity == "HIGH")))     | flatten | length),
    medium:   (pkgs | map(.vulnerabilities // [] | map(select(.severity == "MEDIUM")))   | flatten | length),
    low:      (pkgs | map(.vulnerabilities // [] | map(select(.severity == "LOW")))      | flatten | length)
  };

(.packages // [])
| map(select( ${JQ_SKIP_CONDITION} | not ))
| group_by(pkg_name)
| map(
    . as \$group |
    {
      name:     (\$group[0] | pkg_name),
      version:  (\$group[0] | pkg_version),
      type:     (\$group[0] | pkg_type),
      critical: (\$group | map(.vulnerabilities // [] | map(select(.severity == "CRITICAL"))) | flatten | length),
      high:     (\$group | map(.vulnerabilities // [] | map(select(.severity == "HIGH")))     | flatten | length),
      medium:   (\$group | map(.vulnerabilities // [] | map(select(.severity == "MEDIUM")))   | flatten | length),
      low:      (\$group | map(.vulnerabilities // [] | map(select(.severity == "LOW")))      | flatten | length)
    }
  )
| map(select(.critical > 0 or .high > 0 or .medium > 0 or .low > 0))
| sort_by([-.critical, -.high, -.medium, -.low])
JQEOF
)

# ---------------------------------------------------------------------------
# Per-image totals jq expression.
# ---------------------------------------------------------------------------
JQ_TOTALS=$(cat <<JQEOF
{
  critical: (map(.critical) | add // 0),
  high:     (map(.high)     | add // 0),
  medium:   (map(.medium)   | add // 0),
  low:      (map(.low)      | add // 0)
}
JQEOF
)

# ---------------------------------------------------------------------------
# Helper: print a formatted table row.
# ---------------------------------------------------------------------------
print_table_header() {
  printf "%-45s  %-22s  %4s  %4s  %6s  %4s\n" \
    "Package" "Version" "C" "H" "M" "L"
  printf "%-45s  %-22s  %4s  %4s  %6s  %4s\n" \
    "$(printf '%0.s-' {1..45})" "$(printf '%0.s-' {1..22})" \
    "----" "----" "------" "----"
}

print_table_row() {
  local name="$1" version="$2" c="$3" h="$4" m="$5" l="$6"
  printf "%-45s  %-22s  %4d  %4d  %6d  %4d\n" \
    "${name:0:45}" "${version:0:22}" "$c" "$h" "$m" "$l"
}

# ---------------------------------------------------------------------------
# Log directory for --save.
# ---------------------------------------------------------------------------
LOG_DIR="$HOME/logs"
LOG_SUFFIX="$(date -u '+%Y%m%d')"

# ---------------------------------------------------------------------------
# Main loop.
# ---------------------------------------------------------------------------
GRAND_C=0
GRAND_H=0
GRAND_M=0
GRAND_L=0
SKIPPED_C=0
SKIPPED_H=0
SKIPPED_M=0
SKIPPED_L=0

for IMAGE in "${IMAGES[@]}"; do
  echo
  echo "=== ${IMAGE} ==="
  echo

  if [[ "$RAW" == "true" ]]; then
    # Show unfiltered output for debugging or reference.
    docker scout cves "$IMAGE" --only-fixed=false
    continue
  fi

  # Fetch the CVE report as JSON.
  RAW_JSON=""
  if ! RAW_JSON=$(docker scout cves "$IMAGE" --only-fixed=false --format json 2>&1); then
    echo "ERROR: docker scout cves failed for ${IMAGE}" >&2
    echo "$RAW_JSON" >&2
    continue
  fi

  # Sanity-check: if the output is not JSON (e.g. a login prompt), bail clearly.
  if ! echo "$RAW_JSON" | jq empty 2>/dev/null; then
    echo "ERROR: docker scout returned non-JSON output for ${IMAGE}" >&2
    echo "  (Are the photoprismci credentials present in ~/.docker/config.json?)" >&2
    echo "$RAW_JSON" | head -5 >&2
    continue
  fi

  # --save: write raw JSON to log file.
  if [[ "$SAVE" == "true" ]]; then
    mkdir -p "$LOG_DIR"
    log_file="${LOG_DIR}/docker-scout-cves-${LOG_SUFFIX}.json"
    # Append image + JSON as one object per line (JSONL).
    echo "$RAW_JSON" | jq --arg img "$IMAGE" '. + {scanned_image: $img}' >> "$log_file"
    echo "Raw JSON saved to ${log_file}"
    echo
  fi

  # Apply filters and build the package table.
  FILTERED_JSON=$(echo "$RAW_JSON" | jq "$JQ_FILTER" 2>/dev/null || true)

  if [[ -z "$FILTERED_JSON" ]] || [[ "$FILTERED_JSON" == "null" ]] || [[ "$FILTERED_JSON" == "[]" ]]; then
    echo "  No actionable CVEs found after filtering."

    # Still compute totals for skipped packages.
    RAW_TOTALS=$(echo "$RAW_JSON" | jq '
      (.packages // []) | {
        critical: (map(.vulnerabilities // [] | map(select(.severity == "CRITICAL"))) | flatten | length),
        high:     (map(.vulnerabilities // [] | map(select(.severity == "HIGH")))     | flatten | length),
        medium:   (map(.vulnerabilities // [] | map(select(.severity == "MEDIUM")))   | flatten | length),
        low:      (map(.vulnerabilities // [] | map(select(.severity == "LOW")))      | flatten | length)
      }' 2>/dev/null || echo '{"critical":0,"high":0,"medium":0,"low":0}')
    echo "  (Raw totals incl. filtered: $(echo "$RAW_TOTALS" | jq -r '"C:\(.critical) H:\(.high) M:\(.medium) L:\(.low)"'))"
    continue
  fi

  # Print the filtered table.
  print_table_header
  while IFS= read -r row; do
    name=$(echo "$row" | jq -r '.name')
    version=$(echo "$row" | jq -r '.version')
    c=$(echo "$row" | jq -r '.critical')
    h=$(echo "$row" | jq -r '.high')
    m=$(echo "$row" | jq -r '.medium')
    l=$(echo "$row" | jq -r '.low')
    print_table_row "$name" "$version" "$c" "$h" "$m" "$l"
  done < <(echo "$FILTERED_JSON" | jq -c '.[]')
  echo

  # Per-image totals (filtered).
  TOTALS=$(echo "$FILTERED_JSON" | jq "$JQ_TOTALS")
  img_c=$(echo "$TOTALS" | jq -r '.critical')
  img_h=$(echo "$TOTALS" | jq -r '.high')
  img_m=$(echo "$TOTALS" | jq -r '.medium')
  img_l=$(echo "$TOTALS" | jq -r '.low')
  printf "  Filtered total: %dC  %dH  %dM  %dL\n" "$img_c" "$img_h" "$img_m" "$img_l"

  # Raw totals for the skipped-packages note.
  RAW_TOTALS=$(echo "$RAW_JSON" | jq '
    (.packages // []) | {
      critical: (map(.vulnerabilities // [] | map(select(.severity == "CRITICAL"))) | flatten | length),
      high:     (map(.vulnerabilities // [] | map(select(.severity == "HIGH")))     | flatten | length),
      medium:   (map(.vulnerabilities // [] | map(select(.severity == "MEDIUM")))   | flatten | length),
      low:      (map(.vulnerabilities // [] | map(select(.severity == "LOW")))      | flatten | length)
    }' 2>/dev/null || echo '{"critical":0,"high":0,"medium":0,"low":0}')
  raw_c=$(echo "$RAW_TOTALS" | jq -r '.critical')
  raw_h=$(echo "$RAW_TOTALS" | jq -r '.high')
  raw_m=$(echo "$RAW_TOTALS" | jq -r '.medium')
  raw_l=$(echo "$RAW_TOTALS" | jq -r '.low')
  sk_c=$(( raw_c - img_c ))
  sk_h=$(( raw_h - img_h ))
  sk_m=$(( raw_m - img_m ))
  sk_l=$(( raw_l - img_l ))
  if (( sk_c + sk_h + sk_m + sk_l > 0 )); then
    printf "  Stripped (kernel): %dC  %dH  %dM  %dL\n" "$sk_c" "$sk_h" "$sk_m" "$sk_l"
  fi

  GRAND_C=$(( GRAND_C + img_c ))
  GRAND_H=$(( GRAND_H + img_h ))
  GRAND_M=$(( GRAND_M + img_m ))
  GRAND_L=$(( GRAND_L + img_l ))
  SKIPPED_C=$(( SKIPPED_C + sk_c ))
  SKIPPED_H=$(( SKIPPED_H + sk_h ))
  SKIPPED_M=$(( SKIPPED_M + sk_m ))
  SKIPPED_L=$(( SKIPPED_L + sk_l ))
done

# ---------------------------------------------------------------------------
# Grand total across all images.
# ---------------------------------------------------------------------------
if [[ "$RAW" != "true" ]] && (( ${#IMAGES[@]} > 1 )); then
  echo
  echo "=== Grand Total (${#IMAGES[@]} images) ==="
  printf "  Filtered total: %dC  %dH  %dM  %dL\n" \
    "$GRAND_C" "$GRAND_H" "$GRAND_M" "$GRAND_L"
  if (( SKIPPED_C + SKIPPED_H + SKIPPED_M + SKIPPED_L > 0 )); then
    printf "  Stripped (kernel): %dC  %dH  %dM  %dL\n" \
      "$SKIPPED_C" "$SKIPPED_H" "$SKIPPED_M" "$SKIPPED_L"
  fi
fi

# ---------------------------------------------------------------------------
# Severity note.
# ---------------------------------------------------------------------------
if [[ "$RAW" != "true" ]]; then
  echo
  echo "NOTE: Docker Scout severity ratings often diverge from the Ubuntu Security"
  echo "Tracker. For packages where HIGH or CRITICAL findings look surprising, verify"
  echo "the distro rating at https://ubuntu.com/security/cves/<CVE-ID> before acting."
fi
