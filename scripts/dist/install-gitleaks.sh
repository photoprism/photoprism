#!/usr/bin/env bash

# Installs the Gitleaks credential leak scanner on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-gitleaks.sh)
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-gitleaks.sh) -- --version v8.30.1
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-gitleaks.sh) -- /opt/photoprism

set -euo pipefail

REPO="gitleaks/gitleaks"
TOOL="gitleaks"

if ! command -v jq >/dev/null 2>&1; then
  echo "Error: jq is required but not installed." 1>&2
  exit 1
fi

if [[ ${1:-} == "--help" ]]; then
  echo "Usage: ${0##*/} [--version <tag>] [destdir]" 1>&2
  echo "" 1>&2
  echo "Environment:" 1>&2
  echo "  PHOTOPRISM_ARCH=amd64|arm64" 1>&2
  exit 0
fi

VERSION="latest"

while [[ $# -gt 0 ]]; do
  case $1 in
    --version)
      VERSION=${2:-}
      if [[ -z $VERSION ]]; then
        echo "Error: --version requires a tag (e.g. v8.30.1)." 1>&2
        exit 1
      fi
      shift 2
      ;;
    --help)
      echo "Usage: ${0##*/} [--version <tag>] [destdir]" 1>&2
      exit 0
      ;;
    --)
      shift
      break
      ;;
    -*)
      echo "Error: Unknown option: $1" 1>&2
      exit 1
      ;;
    *)
      break
      ;;
  esac
done

DESTDIR=$(realpath "${1:-/usr/local}")

if [[ -n ${PHOTOPRISM_ARCH:-} ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi
DESTARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

# Gitleaks tarball asset names embed the architecture but NOT (always) "linux";
# resolve them via the GitHub API so we don't have to hardcode version tags.
case $DESTARCH in
  amd64 | AMD64 | x86_64 | x86-64)
    ASSET_ARCH="x64"
    ;;
  arm64 | ARM64 | aarch64)
    ASSET_ARCH="arm64"
    ;;
  *)
    echo "Unsupported Machine Architecture: \"$DESTARCH\"" 1>&2
    exit 1
    ;;
esac

if [[ $(id -u) != "0" ]] && [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Error: Run ${0##*/} as root to install in a system directory!" 1>&2
  exit 1
fi

if [[ $VERSION == "latest" ]]; then
  RELEASE_API="https://api.github.com/repos/${REPO}/releases/latest"
else
  RELEASE_API="https://api.github.com/repos/${REPO}/releases/tags/${VERSION}"
fi

RELEASE_JSON=$(curl --fail --silent --show-error "${RELEASE_API}" || true)
if [[ -z $RELEASE_JSON ]]; then
  echo "Error: Unable to fetch release metadata from ${RELEASE_API}." 1>&2
  exit 1
fi

TAG_NAME=$(echo "${RELEASE_JSON}" | jq -r '.tag_name // empty')
ASSET_NAME=$(echo "${RELEASE_JSON}" | jq -r --arg arch "linux_${ASSET_ARCH}.tar.gz" \
  '.assets[]? | select(.name | endswith($arch)) | .name' | head -n1)
ASSET_URL=$(echo "${RELEASE_JSON}" | jq -r --arg name "${ASSET_NAME}" \
  '.assets[]? | select(.name == $name) | .browser_download_url' | head -n1)
CHECKSUM_NAME=$(echo "${RELEASE_JSON}" | jq -r \
  '.assets[]? | select(.name | test("checksums.txt$")) | .name' | head -n1)
CHECKSUM_URL=$(echo "${RELEASE_JSON}" | jq -r --arg name "${CHECKSUM_NAME}" \
  '.assets[]? | select(.name == $name) | .browser_download_url' | head -n1)

if [[ -z ${TAG_NAME} || -z ${ASSET_URL} ]]; then
  echo "Error: Could not resolve a downloadable asset for architecture ${DESTARCH}." 1>&2
  exit 1
fi

DESTBIN="${DESTDIR}/bin/${TOOL}"

echo "--------------------------------------------------------------------------------"
echo "VERSION  : ${TAG_NAME}"
echo "ASSET    : ${ASSET_NAME}"
echo "DOWNLOAD : ${ASSET_URL}"
echo "DESTBIN  : ${DESTBIN}"
echo "--------------------------------------------------------------------------------"

echo "Downloading ${TOOL} to \"${DESTBIN}\"..."
mkdir -p "${DESTDIR}/bin"
tmp_dir=$(mktemp -d)
trap 'rm -rf "${tmp_dir}"' EXIT

curl --fail --silent --show-error --location "${ASSET_URL}" -o "${tmp_dir}/${ASSET_NAME}"

if [[ -n "${CHECKSUM_URL}" ]]; then
  curl --fail --silent --show-error --location "${CHECKSUM_URL}" -o "${tmp_dir}/checksums.txt"
  expected=$(awk -v want="${ASSET_NAME}" '$2 == want {print $1; exit}' "${tmp_dir}/checksums.txt")
  if [[ -z "${expected}" ]]; then
    echo "Warning: ${ASSET_NAME} not listed in ${CHECKSUM_NAME}; skipping verification." 1>&2
  else
    actual=$(sha256sum "${tmp_dir}/${ASSET_NAME}" | awk '{print $1}')
    if [[ "${expected}" != "${actual}" ]]; then
      echo "Error: SHA-256 mismatch for ${TOOL} (${ASSET_NAME})." 1>&2
      echo "  expected: ${expected}" 1>&2
      echo "  actual:   ${actual}" 1>&2
      exit 1
    fi
    echo "Checksum OK (${actual})."
  fi
else
  echo "Warning: no published checksums file found for ${TAG_NAME}; skipping verification." 1>&2
fi

tar -xzf "${tmp_dir}/${ASSET_NAME}" -C "${tmp_dir}"

if [[ ! -f "${tmp_dir}/${TOOL}" ]]; then
  echo "Error: ${TOOL} binary not found inside ${ASSET_NAME}." 1>&2
  exit 1
fi

install -m 755 "${tmp_dir}/${TOOL}" "${DESTBIN}"

if command -v "${TOOL}" >/dev/null 2>&1; then
  "${TOOL}" version | head -n 1
fi

echo "Done."
