#!/usr/bin/env bash

# Installs the InsightFace SCRFD face detector.
#
# These weights are separately licensed, so they are NOT installed by "make dep" and
# MUST NOT be redistributed with PhotoPrism. Set INSIGHTFACE_ACCEPT_LICENSE=1 to confirm
# that your use is covered before running this.
#
# See: https://github.com/deepinsight/insightface/tree/master/model_zoo

set -euo pipefail

TODAY=$(date -u +%Y%m%d)

MODELS_PATH=${MODELS_PATH:-"${PHOTOPRISM_ASSETS_PATH:-assets}/models"}
MODEL_DIR="$MODELS_PATH/scrfd"
LEGACY_MODEL_DIR="$MODELS_PATH/scrfs"
MODEL_VERSION="$MODEL_DIR/version.txt"

# The detector ships inside a release pack rather than as a standalone asset. Its checksum is
# pinned so a replaced upstream file is rejected instead of silently installed.
MODEL_PACK="buffalo_s"
MODEL_ENTRY="det_500m.onnx"
MODEL_SHA256="5e4447f50245bbd7966bd6c0fa52938c61474a04ec7def48753668a9d8b4ea3a"
MODEL_URL="https://github.com/deepinsight/insightface/releases/download/v0.7/${MODEL_PACK}.zip"

if [[ "${INSIGHTFACE_ACCEPT_LICENSE:-}" != "1" ]]; then
  cat >&2 <<'EOF'
This optional PhotoPrism integration uses InsightFace model weights that are
separately licensed and are not covered by PhotoPrism's AGPL license. They are
permitted through PhotoPrism only for personal, noncommercial use by individual
users. The model weights may not be bundled, redistributed, sublicensed, used
outside PhotoPrism, or used for business, organizational, professional,
revenue-generating, SaaS, or managed-service purposes without a separate
commercial license from InsightFace. For licensing information, visit
https://www.insightface.ai or contact contact@insightface.ai.

Re-run with INSIGHTFACE_ACCEPT_LICENSE=1 if your use is covered:

  INSIGHTFACE_ACCEPT_LICENSE=1 scripts/dist/download-scrfd.sh
EOF
  exit 1
fi

if ! command -v unzip >/dev/null 2>&1; then
  echo "unzip is required to extract the SCRFD detector." >&2
  exit 1
fi

# A private directory the script owns, so nothing it downloads or extracts can be
# redirected through a name another user of the machine placed there first.
TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/photoprism-scrfd.XXXXXXXX")"

cleanup() {
  rm -rf "${TMP_DIR}"
}

trap cleanup EXIT

if [[ -d "${LEGACY_MODEL_DIR}" && ! -d "${MODEL_DIR}" ]]; then
  echo "Migrating legacy directory from ${LEGACY_MODEL_DIR} to ${MODEL_DIR}."
  mv "${LEGACY_MODEL_DIR}" "${MODEL_DIR}"
fi

mkdir -p "${MODEL_DIR}"

hash_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

if [[ -f "${MODEL_DIR}/${MODEL_ENTRY}" ]]; then
  if [[ "$(hash_file "${MODEL_DIR}/${MODEL_ENTRY}")" == "${MODEL_SHA256}" ]]; then
    echo "SCRFD detector already up to date."
    exit 0
  fi
fi

ARCHIVE="${TMP_DIR}/${MODEL_PACK}.zip"

echo "Downloading ${MODEL_PACK} from ${MODEL_URL}..."

if ! curl -fL --no-progress-meter --retry 3 --retry-delay 2 -o "${ARCHIVE}" "${MODEL_URL}"; then
  echo "Failed to download ${MODEL_PACK}." >&2
  exit 1
fi

echo "Extracting ${MODEL_ENTRY}..."

if ! unzip -j -o "${ARCHIVE}" "${MODEL_ENTRY}" -d "${TMP_DIR}" >/dev/null; then
  echo "Failed to extract ${MODEL_ENTRY} from ${MODEL_PACK}." >&2
  exit 1
fi

echo "Verifying checksum..."

if [[ "$(hash_file "${TMP_DIR}/${MODEL_ENTRY}")" != "${MODEL_SHA256}" ]]; then
  echo "Checksum mismatch, refusing to install ${MODEL_ENTRY}." >&2
  exit 1
fi

mv "${TMP_DIR}/${MODEL_ENTRY}" "${MODEL_DIR}/${MODEL_ENTRY}"
echo "SCRFD ${TODAY} ${MODEL_SHA256} (${MODEL_ENTRY} from ${MODEL_PACK})" > "${MODEL_VERSION}"

echo "SCRFD detector installed in ${MODEL_DIR} (personal, noncommercial use only)."
