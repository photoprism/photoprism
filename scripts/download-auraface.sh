#!/usr/bin/env bash

# Installs the fal AuraFace v1 model used to generate face embeddings when
# PHOTOPRISM_FACE_MODEL=auraface. The weights are Apache-2.0 licensed, so they
# may be redistributed with PhotoPrism. The graph is 249 MB and is therefore an
# opt-in download rather than a bundled model.
#
# The target file name differs from the upstream one on purpose: InsightFace's
# antelopev2 pack ships a different model under the same name "glintr100.onnx",
# and matching by name alone would apply one model's preprocessing to the
# other's weights without failing.

set -euo pipefail

TODAY=$(date -u +%Y%m%d)

MODEL_NAME="AuraFace"
MODEL_SOURCE="auraface_v1_glintr100.onnx"
PRIMARY_URL="https://dl.photoprism.app/onnx/models/${MODEL_SOURCE}?${TODAY}"
FALLBACK_URL="https://huggingface.co/fal/AuraFace-v1/resolve/main/glintr100.onnx?download=true"
MODEL_URL=${MODEL_URL:-"${PRIMARY_URL}"}
MODELS_PATH="assets/models"
MODEL_DIR="$MODELS_PATH/auraface"
MODEL_FILE="$MODEL_DIR/${MODEL_SOURCE}"
MODEL_TMP="/tmp/photoprism/${MODEL_SOURCE}"
MODEL_SHA256="a7933ea5330113b01c9b60351d8f4c33003f145d8470ac5f0e52ee2effe25c60"
MODEL_VERSION="$MODEL_DIR/version.txt"
MODEL_BACKUP="storage/backup/auraface-${TODAY}"

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
