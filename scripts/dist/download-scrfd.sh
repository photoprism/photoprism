#!/usr/bin/env bash

set -euo pipefail

TODAY=$(date -u +%Y%m%d)

MODEL_SOURCE="scrfd_500m_bnkps_shape640x640.onnx"
LOCAL_MODEL_NAME="scrfd.onnx"
PRIMARY_URL="https://dl.photoprism.app/onnx/models/${MODEL_SOURCE}?${TODAY}"
FALLBACK_URL="https://raw.githubusercontent.com/laolaolulu/FaceTrain/master/model/scrfd/${MODEL_SOURCE}"
MODEL_URL=${MODEL_URL:-"${PRIMARY_URL}"}
MODELS_PATH="assets/models"
MODEL_DIR="$MODELS_PATH/scrfd"
LEGACY_MODEL_DIR="$MODELS_PATH/scrfs"
MODEL_FILE="$MODEL_DIR/${LOCAL_MODEL_NAME}"
MODEL_TMP="/tmp/photoprism/${MODEL_SOURCE}"
MODEL_HASH="ae72185653e279aa2056b288662a19ec3519ced5426d2adeffbe058a86369a24  ${MODEL_TMP}"
MODEL_VERSION="$MODEL_DIR/version.txt"
MODEL_BACKUP="storage/backup/scrfd-${TODAY}"

mkdir -p /tmp/photoprism
mkdir -p storage/backup

if [[ -d "${LEGACY_MODEL_DIR}" && ! -d "${MODEL_DIR}" ]]; then
  echo "Migrating legacy SCRFD directory from ${LEGACY_MODEL_DIR} to ${MODEL_DIR}."
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

verify_hash() {
  local expected="$1"
  local file="$2"
  if command -v sha256sum >/dev/null 2>&1; then
    echo "${expected}  ${file}" | sha256sum -c - >/dev/null
  else
    echo "${expected}  ${file}" | shasum -a 256 -c - >/dev/null
  fi
}

if [[ -f "${MODEL_FILE}" ]]; then
  CURRENT_HASH=$(hash_file "${MODEL_FILE}")
  if [[ "${CURRENT_HASH}" == "${MODEL_HASH%% *}" ]]; then
    echo "SCRFD model already up to date."
    exit 0
  fi
fi

echo "Downloading SCRFD detector from ${MODEL_URL}..."
if ! curl -fsSL --retry 3 --retry-delay 2 -o "${MODEL_TMP}" "${MODEL_URL}"; then
  if [[ "${MODEL_URL}" != "${FALLBACK_URL}" ]]; then
    echo "Primary download failed, trying fallback..."
    MODEL_URL="${FALLBACK_URL}"
    MODEL_HASH="ae72185653e279aa2056b288662a19ec3519ced5426d2adeffbe058a86369a24  ${MODEL_TMP}"
    if ! curl -fsSL --retry 3 --retry-delay 2 -o "${MODEL_TMP}" "${MODEL_URL}"; then
      echo "Failed to download SCRFD detector." >&2
      exit 1
    fi
  else
    echo "Failed to download SCRFD detector." >&2
    exit 1
  fi
fi

echo "Verifying checksum..."
verify_hash "${MODEL_HASH%% *}" "${MODEL_TMP}"

if [[ -f "${MODEL_FILE}" ]]; then
  echo "Creating backup of existing detector at ${MODEL_BACKUP}"
  rm -rf "${MODEL_BACKUP}"
  mkdir -p "${MODEL_BACKUP}"
  mv "${MODEL_FILE}" "${MODEL_BACKUP}/"
  if [[ -f "${MODEL_VERSION}" ]]; then
    cp "${MODEL_VERSION}" "${MODEL_BACKUP}/"
  fi
fi

mv "${MODEL_TMP}" "${MODEL_FILE}"
echo "SCRFD ${TODAY} ${MODEL_HASH%% *} (${MODEL_SOURCE})" > "${MODEL_VERSION}"

echo "SCRFD detector installed in ${MODEL_DIR}."
