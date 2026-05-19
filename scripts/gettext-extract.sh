#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
FRONTEND_DIR="${ROOT_DIR}/frontend"

if [[ ! -d "${FRONTEND_DIR}" ]]; then
  echo "ERROR: frontend directory not found: ${FRONTEND_DIR}" 1>&2
  exit 1
fi

declare -a src_dirs=("src")
declare -a overlay_dirs=("plus/frontend" "pro/frontend" "portal/frontend")

for rel_dir in "${overlay_dirs[@]}"; do
  if [[ -d "${ROOT_DIR}/${rel_dir}" ]]; then
    src_dirs+=("../${rel_dir}")
  fi
done

if [[ -n "${GETTEXT_EXTRA_SRC:-}" ]]; then
  # Split optional extra source directories on spaces.
  read -r -a extra_dirs <<< "${GETTEXT_EXTRA_SRC}"

  for extra_dir in "${extra_dirs[@]}"; do
    if [[ -n "${extra_dir}" ]]; then
      src_dirs+=("${extra_dir}")
    fi
  done
fi

echo "Extracting frontend translations from: ${src_dirs[*]}"

(
  cd "${FRONTEND_DIR}"

  # Skip msgmerge auto-filling while extracting POT.
  env SRC="${src_dirs[*]}" GETTEXT_MERGE=0 npm run gettext-extract

  # Keep source references stable across environments and private overlays.
  sed -i \
    -e 's#\.\./plus/frontend#src#g' \
    -e 's#\.\./pro/frontend#src#g' \
    -e 's#\.\./portal/frontend#src#g' \
    src/locales/translations.pot
)

echo "Merging gettext catalogs..."
"${ROOT_DIR}/scripts/gettext-merge.sh"

