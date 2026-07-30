#!/usr/bin/env bash

# Installs the OpenCV Zoo SFace model used to generate face embeddings when
# PHOTOPRISM_FACE_MODEL=sface. The weights are Apache-2.0 licensed, so they may
# be redistributed with PhotoPrism.

set -euo pipefail

TODAY=$(date -u +%Y%m%d)

MODEL_NAME="SFace"
MODEL_SOURCE="face_recognition_sface_2021dec.onnx"
PRIMARY_URL="https://dl.photoprism.app/onnx/models/${MODEL_SOURCE}?${TODAY}"
FALLBACK_URL="https://media.githubusercontent.com/media/opencv/opencv_zoo/main/models/face_recognition_sface/${MODEL_SOURCE}"
MODEL_URL=${MODEL_URL:-"${PRIMARY_URL}"}
MODELS_PATH="assets/models"
MODEL_DIR="$MODELS_PATH/sface"
MODEL_FILE="$MODEL_DIR/${MODEL_SOURCE}"
MODEL_TMP="/tmp/photoprism/${MODEL_SOURCE}"
MODEL_SHA256="0ba9fbfa01b5270c96627c4ef784da859931e02f04419c829e83484087c34e79"
MODEL_VERSION="$MODEL_DIR/version.txt"
MODEL_BACKUP="storage/backup/sface-${TODAY}"

mkdir -p /tmp/photoprism
mkdir -p storage/backup
mkdir -p "${MODEL_DIR}"

hash_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

if [[ -f "${MODEL_FILE}" ]] && [[ "$(hash_file "${MODEL_FILE}")" == "${MODEL_SHA256}" ]]; then
  echo "${MODEL_NAME} model already up to date."
  exit 0
fi

echo "Downloading ${MODEL_NAME} model from ${MODEL_URL}..."

if ! curl -fsSL --retry 3 --retry-delay 2 -o "${MODEL_TMP}" "${MODEL_URL}"; then
  if [[ "${MODEL_URL}" == "${FALLBACK_URL}" ]]; then
    echo "Failed to download ${MODEL_NAME} model." >&2
    exit 1
  fi

  echo "Primary download failed, trying fallback..."

  if ! curl -fsSL --retry 3 --retry-delay 2 -o "${MODEL_TMP}" "${FALLBACK_URL}"; then
    echo "Failed to download ${MODEL_NAME} model." >&2
    exit 1
  fi
fi

echo "Verifying checksum..."

if [[ "$(hash_file "${MODEL_TMP}")" != "${MODEL_SHA256}" ]]; then
  echo "Checksum mismatch, refusing to install ${MODEL_NAME} model." >&2
  rm -f "${MODEL_TMP}"
  exit 1
fi

if [[ -f "${MODEL_FILE}" ]]; then
  echo "Creating backup of existing model at ${MODEL_BACKUP}"
  rm -rf "${MODEL_BACKUP}"
  mkdir -p "${MODEL_BACKUP}"
  mv "${MODEL_FILE}" "${MODEL_BACKUP}/"

  if [[ -f "${MODEL_VERSION}" ]]; then
    cp "${MODEL_VERSION}" "${MODEL_BACKUP}/"
  fi
fi

mv "${MODEL_TMP}" "${MODEL_FILE}"
echo "${MODEL_NAME} ${TODAY} ${MODEL_SHA256} (${MODEL_SOURCE})" > "${MODEL_VERSION}"

echo "${MODEL_NAME} model installed in ${MODEL_DIR}."
