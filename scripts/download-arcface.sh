#!/usr/bin/env bash

# Installs the InsightFace ArcFace recognition models for benchmarking.
#
# These weights are published for non-commercial research only, so they are NOT
# installed by "make dep" and MUST NOT be redistributed with PhotoPrism. Set
# ARCFACE_ACCEPT_LICENSE=1 to confirm that your use is covered before running this.
#
# See: https://github.com/deepinsight/insightface/tree/master/model_zoo

set -euo pipefail

TODAY=$(date -u +%Y%m%d)

MODELS_PATH="assets/models"
MODEL_DIR="$MODELS_PATH/arcface"
MODEL_VERSION="$MODEL_DIR/version.txt"
TMP_DIR="/tmp/photoprism/arcface"

if [[ "${ARCFACE_ACCEPT_LICENSE:-}" != "1" ]]; then
  cat >&2 <<'EOF'
The InsightFace pretrained recognition models are licensed for non-commercial
research only. PhotoPrism does not redistribute them.

Re-run with ARCFACE_ACCEPT_LICENSE=1 if your use is covered by that license:

  ARCFACE_ACCEPT_LICENSE=1 scripts/download-arcface.sh
EOF
  exit 1
fi

if ! command -v unzip >/dev/null 2>&1; then
  echo "unzip is required to extract the ArcFace models." >&2
  exit 1
fi

mkdir -p "${TMP_DIR}"
mkdir -p "${MODEL_DIR}"

hash_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

# install_model <pack> <entry> <target>
install_model() {
  local pack="$1"
  local entry="$2"
  local target="$3"
  local url="https://github.com/deepinsight/insightface/releases/download/v0.7/${pack}.zip"
  local archive="${TMP_DIR}/${pack}.zip"

  if [[ -f "${MODEL_DIR}/${target}" ]]; then
    echo "${target} already installed."
    return 0
  fi

  if [[ ! -f "${archive}" ]]; then
    echo "Downloading ${pack} from ${url}..."

    if ! curl -fL --no-progress-meter --retry 3 --retry-delay 2 -o "${archive}" "${url}"; then
      echo "Failed to download ${pack}." >&2
      return 1
    fi
  fi

  echo "Extracting ${entry}..."

  if ! unzip -j -o "${archive}" "${entry}" -d "${TMP_DIR}" >/dev/null; then
    echo "Failed to extract ${entry} from ${pack}." >&2
    return 1
  fi

  mv "${TMP_DIR}/$(basename "${entry}")" "${MODEL_DIR}/${target}"
  echo "ArcFace ${TODAY} $(hash_file "${MODEL_DIR}/${target}") (${target} from ${pack})" >> "${MODEL_VERSION}"
  echo "Installed ${target}."
}

install_model "buffalo_l" "w600k_r50.onnx" "w600k_r50.onnx"
install_model "buffalo_s" "w600k_mbf.onnx" "w600k_mbf.onnx"

echo "ArcFace benchmark models installed in ${MODEL_DIR} (non-commercial research use only)."
