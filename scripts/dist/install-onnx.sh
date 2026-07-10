#!/usr/bin/env bash

set -euo pipefail

ONNX_VERSION=${ONNX_VERSION:-1.26.0}
TODAY=$(date -u +%Y%m%d)
TMPDIR=${TMPDIR:-/tmp}
SYSTEM=$(uname -s)
ARCH=${PHOTOPRISM_ARCH:-$(uname -m)}

# ONNX_GPU selects a CUDA (GPU) build instead of the default CPU build. It can be
# set via the ONNX_GPU environment variable or a --gpu[=cuda12|cuda13] argument:
#   unset / 0 / cpu  -> CPU build (default)
#   1 / gpu / cuda13 -> CUDA 13 build (recommended; CUDA 12 is deprecated upstream)
#   cuda12           -> CUDA 12 build (legacy)
# GPU builds are only published for Linux x64 and additionally require the NVIDIA
# driver plus a matching CUDA runtime and cuDNN to be present at runtime.
ONNX_GPU=${ONNX_GPU:-}

# Parse optional flags; the first non-flag argument is the install prefix.
DESTDIR_ARG="/usr"
destdir_set=0
for arg in "$@"; do
  case "${arg}" in
    --gpu)    ONNX_GPU="cuda13" ;;
    --gpu=*)  ONNX_GPU="${arg#--gpu=}" ;;
    --cpu)    ONNX_GPU="" ;;
    -*)       echo "Error: unknown option '${arg}'." >&2; exit 2 ;;
    *)        if [[ "${destdir_set}" == 0 ]]; then DESTDIR_ARG="${arg}"; destdir_set=1; fi ;;
  esac
done

# Normalize the GPU selection to an empty (CPU) or "cudaNN" variant.
gpu_variant=""
case "${ONNX_GPU,,}" in
  ""|0|cpu|false|no)    gpu_variant="" ;;
  1|gpu|cuda13|cuda-13) gpu_variant="cuda13" ;;
  cuda12|cuda-12)       gpu_variant="cuda12" ;;
  *) echo "Error: unsupported ONNX_GPU value '${ONNX_GPU}' (use cuda12 or cuda13)." >&2; exit 1 ;;
esac

if [[ ! -d "${DESTDIR_ARG}" ]]; then
  mkdir -p "${DESTDIR_ARG}"
fi

DESTDIR=$(realpath "${DESTDIR_ARG}")

if [[ $(id -u) != 0 ]] && { [[ "${DESTDIR}" == "/usr" ]] || [[ "${DESTDIR}" == "/usr/local" ]]; }; then
  echo "Error: Run ${0##*/} as root to install in '${DESTDIR}'." >&2
  exit 1
fi

mkdir -p "${DESTDIR}" "${TMPDIR}"

# version_lt returns success if $1 is a strictly lower semantic version than $2.
version_lt() {
  [[ "$1" != "$2" ]] && [[ "$(printf '%s\n%s\n' "$1" "$2" | sort -V | head -n1)" == "$1" ]]
}

archive=""
sha=""

case "${SYSTEM}" in
  Linux)
    case "${ARCH}" in
      amd64|AMD64|x86_64|x86-64)
        if [[ -n "${gpu_variant}" ]]; then
          # Upstream renamed the CUDA-12 archive from "-gpu-" to "-gpu_cuda12-" in v1.27.0.
          if [[ "${gpu_variant}" == "cuda12" ]] && version_lt "${ONNX_VERSION}" "1.27.0"; then
            archive="onnxruntime-linux-x64-gpu-${ONNX_VERSION}.tgz"
            sha="cb7df7ee2ca0f962c7ce7c839aeae36223d146a91fb4646d62fb0046f297479f"
          else
            archive="onnxruntime-linux-x64-gpu_${gpu_variant}-${ONNX_VERSION}.tgz"
            if [[ "${gpu_variant}" == "cuda13" ]]; then
              sha="aa619d5701bbe58046cc998b21e692d5b2aefac1479f375c4b988526cb80befa"
            fi
          fi
        else
          archive="onnxruntime-linux-x64-${ONNX_VERSION}.tgz"
          sha="1254da24fb389cf39dc0ff3451ab48301740ffbfcbaf646849df92f80ee92c57"
        fi
        ;;
      arm64|ARM64|aarch64)
        if [[ -n "${gpu_variant}" ]]; then
          echo "Error: ONNX Runtime GPU/CUDA builds are only available for Linux x64." >&2
          exit 1
        fi
        archive="onnxruntime-linux-aarch64-${ONNX_VERSION}.tgz"
        sha="34ff1c2d0f12e2cf3d33a0c5f82e39792e1d581fbd6968fd7c30d173654be01a"
        ;;
      *)
        echo "Warning: ONNX Runtime is not provided for Linux/${ARCH}; skipping install." >&2
        exit 0
        ;;
    esac
    ;;
  Darwin)
    if [[ -n "${gpu_variant}" ]]; then
      echo "Error: ONNX Runtime GPU/CUDA builds are only available for Linux x64." >&2
      exit 1
    fi
    case "${ARCH}" in
      arm64|ARM64|aarch64)
        archive="onnxruntime-osx-arm64-${ONNX_VERSION}.tgz"
        sha="7a1280bbb1701ea514f71828765237e7896e0f2e1cd332f1f70dbd5c3e33aca3"
        ;;
      x86_64|x86-64)
        echo "Warning: ONNX Runtime is not provided for macOS/${ARCH} in v${ONNX_VERSION}; skipping install." >&2
        exit 0
        ;;
      *)
        echo "Unsupported macOS architecture '${ARCH}'." >&2
        exit 1
        ;;
    esac
    ;;
  *)
    echo "Unsupported operating system '${SYSTEM}'." >&2
    exit 1
    ;;
 esac

# Allow an explicit checksum override (e.g. when installing a non-default version
# or a GPU variant whose checksum is not pinned in this script).
sha="${ONNX_SHA256:-${sha}}"

verify_sha() {
  local expected="$1"
  local file="$2"
  if command -v sha256sum >/dev/null 2>&1; then
    echo "${expected}  ${file}" | sha256sum -c - >/dev/null
  else
    echo "${expected}  ${file}" | shasum -a 256 -c - >/dev/null
  fi
}

if [[ -z "${archive}" ]]; then
  echo "Could not determine ONNX Runtime archive." >&2
  exit 1
fi

if [[ -z "${sha}" ]]; then
  echo "Error: no checksum pinned for '${archive}'. Set ONNX_SHA256 to install it." >&2
  exit 1
fi

primary_url="https://dl.photoprism.app/onnx/runtime/v${ONNX_VERSION}/${archive}?${TODAY}"
fallback_url="https://github.com/microsoft/onnxruntime/releases/download/v${ONNX_VERSION}/${archive}"
package_path="${TMPDIR}/${archive}"

if [[ -f "${package_path}" ]]; then
  if verify_sha "${sha}" "${package_path}"; then
    echo "Using cached archive ${package_path}."
  else
    echo "Cached archive ${package_path} failed checksum, re-downloading..."
    rm -f "${package_path}"
  fi
fi

if [[ ! -f "${package_path}" ]]; then
  echo "Downloading ONNX Runtime ${ONNX_VERSION} (${archive})..."
  if ! curl -fsSL --retry 3 --retry-delay 2 -o "${package_path}" "${primary_url}"; then
    echo "Primary download failed, trying upstream release..."
    if ! curl -fsSL --retry 3 --retry-delay 2 -o "${package_path}" "${fallback_url}"; then
      echo "Failed to download ONNX Runtime archive." >&2
      exit 1
    fi
  fi
fi

echo "Verifying checksum..."
verify_sha "${sha}" "${package_path}"

echo "Extracting to ${DESTDIR}..."
tar --overwrite --mode=755 -C "${DESTDIR}" -xzf "${package_path}"

# Normalize layout: copy libraries into ${DESTDIR}/lib and remove extracted tree.
# The archive extracts to a top directory named after itself (minus ".tgz"),
# which also covers the GPU builds and their extra provider libraries.
output_lib_dir="${DESTDIR}/lib"
mkdir -p "${output_lib_dir}"

# Determine the extracted top-level directory from the archive itself: upstream's
# GPU archive filenames (e.g. "-gpu_cuda13-") do not match their internal
# directory name (e.g. "-gpu-"), so it cannot be derived from the file name.
# (|| true: head closes the pipe early, which SIGPIPEs tar under pipefail.)
extracted_name=$(tar tzf "${package_path}" 2>/dev/null | head -1 | cut -d/ -f1 || true)
extracted="${DESTDIR}/${extracted_name}"
if [[ -n "${extracted_name}" && -d "${extracted}/lib" ]]; then
  find "${extracted}/lib" -maxdepth 1 -type f -name "libonnxruntime*.so*" -print0 | while IFS= read -r -d '' file; do
    cp -af "${file}" "${output_lib_dir}/"
  done
  # copy any symlinks as well to preserve SONAME links
  find "${extracted}/lib" -maxdepth 1 -type l -name "libonnxruntime*.so*" -print0 | while IFS= read -r -d '' link; do
    target=$(readlink "${link}")
    ln -sf "${target}" "${output_lib_dir}/$(basename "${link}")"
  done
  rm -rf "${extracted}"
fi

if [[ "${SYSTEM}" == "Linux" ]]; then
  if [[ "${DESTDIR}" == "/usr" || "${DESTDIR}" == "/usr/local" ]]; then
    ldconfig
  else
    ldconfig -n "${DESTDIR}/lib" >/dev/null 2>&1 || true
  fi
fi

echo "ONNX Runtime ${ONNX_VERSION}${gpu_variant:+ (GPU/${gpu_variant})} installed in '${DESTDIR}'."
