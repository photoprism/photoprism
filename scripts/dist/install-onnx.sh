#!/usr/bin/env bash

set -euo pipefail

ONNX_VERSION=${ONNX_VERSION:-1.22.0}
TODAY=$(date -u +%Y%m%d)
TMPDIR=${TMPDIR:-/tmp}
SYSTEM=$(uname -s)
ARCH=${PHOTOPRISM_ARCH:-$(uname -m)}
DESTDIR_ARG="${1:-/usr}"

if [[ ! -d "${DESTDIR_ARG}" ]]; then
  mkdir -p "${DESTDIR_ARG}"
fi

DESTDIR=$(realpath "${DESTDIR_ARG}")

if [[ $(id -u) != 0 ]] && { [[ "${DESTDIR}" == "/usr" ]] || [[ "${DESTDIR}" == "/usr/local" ]]; }; then
  echo "Error: Run ${0##*/} as root to install in '${DESTDIR}'." >&2
  exit 1
fi

mkdir -p "${DESTDIR}" "${TMPDIR}"

archive=""
sha=""

case "${SYSTEM}" in
  Linux)
    case "${ARCH}" in
      amd64|AMD64|x86_64|x86-64)
        archive="onnxruntime-linux-x64-${ONNX_VERSION}.tgz"
        sha="8344d55f93d5bc5021ce342db50f62079daf39aaafb5d311a451846228be49b3"
        ;;
      arm64|ARM64|aarch64)
        archive="onnxruntime-linux-aarch64-${ONNX_VERSION}.tgz"
        sha="bb76395092d150b52c7092dc6b8f2fe4d80f0f3bf0416d2f269193e347e24702"
        ;;
      *)
        echo "Warning: ONNX Runtime is not provided for Linux/${ARCH}; skipping install." >&2
        exit 0
        ;;
    esac
    ;;
  Darwin)
    case "${ARCH}" in
      arm64|ARM64|aarch64|x86_64|x86-64)
        archive="onnxruntime-osx-universal2-${ONNX_VERSION}.tgz"
        sha="cfa6f6584d87555ed9f6e7e8a000d3947554d589efe3723b8bfa358cd263d03c"
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
output_lib_dir="${DESTDIR}/lib"
mkdir -p "${output_lib_dir}"

for extracted in "${DESTDIR}/onnxruntime-linux-x64-${ONNX_VERSION}" "${DESTDIR}/onnxruntime-linux-aarch64-${ONNX_VERSION}" "${DESTDIR}/onnxruntime-osx-universal2-${ONNX_VERSION}"; do
  if [[ -d "${extracted}/lib" ]]; then
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
done

if [[ "${SYSTEM}" == "Linux" ]]; then
  if [[ "${DESTDIR}" == "/usr" || "${DESTDIR}" == "/usr/local" ]]; then
    ldconfig
  else
    ldconfig -n "${DESTDIR}/lib" >/dev/null 2>&1 || true
  fi
fi

echo "ONNX Runtime ${ONNX_VERSION} installed in '${DESTDIR}'."
