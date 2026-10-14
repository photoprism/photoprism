#!/usr/bin/env bash

# Installs the Deno JavaScript runtime on Linux, which yt-dlp needs to extract YouTube video formats.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-deno.sh)
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-deno.sh) -- --version v2.9.7
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-deno.sh) -- /opt/photoprism

set -euo pipefail

REPO="denoland/deno"

usage() {
  echo "Usage: ${0##*/} [--version <tag>] [destdir]" 1>&2
  echo "" 1>&2
  echo "Environment:" 1>&2
  echo "  PHOTOPRISM_ARCH=amd64|arm64" 1>&2
}

VERSION="latest"

while [[ $# -gt 0 ]]; do
  case $1 in
    --version)
      VERSION=${2:-}
      if [[ -z $VERSION ]]; then
        echo "Error: --version requires a tag (e.g. v2.9.7)." 1>&2
        exit 1
      fi
      shift 2
      ;;
    --help)
      usage
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

for cmd in curl jq unzip; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "Error: $cmd is required but not installed." 1>&2
    exit 1
  fi
done

# You can provide a custom installation directory as the first argument.
DESTDIR=$(realpath "${1:-/usr/local}")
DESTBIN="${DESTDIR}/bin/deno"

# Determine target architecture.
if [[ -n ${PHOTOPRISM_ARCH:-} ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi
DESTARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

case $DESTARCH in
  amd64 | AMD64 | x86_64 | x86-64)
    ASSET="deno-x86_64-unknown-linux-gnu.zip"
    ;;
  arm64 | ARM64 | aarch64)
    ASSET="deno-aarch64-unknown-linux-gnu.zip"
    ;;
  *)
    echo "Unsupported Machine Architecture: \"$DESTARCH\"" 1>&2
    exit 1
    ;;
esac

# Abort if not executed as root when installing into a system directory.
if [[ $(id -u) != "0" ]] && [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Error: Run ${0##*/} as root to install in a system directory!" 1>&2
  exit 1
fi

if [[ $VERSION == "latest" ]]; then
  RELEASE_URL="https://api.github.com/repos/${REPO}/releases/latest"
else
  [[ $VERSION == v* ]] || VERSION="v${VERSION}"
  RELEASE_URL="https://api.github.com/repos/${REPO}/releases/tags/${VERSION}"
fi

if ! RELEASE_JSON=$(curl --fail --silent --show-error --location --retry 3 --retry-delay 2 --retry-all-errors "$RELEASE_URL"); then
  echo "Error: Unable to fetch Deno release metadata from GitHub." 1>&2
  exit 1
fi

TAG_NAME=$(echo "$RELEASE_JSON" | jq -r '.tag_name // empty')
ASSET_URL=$(echo "$RELEASE_JSON" | jq -r --arg name "$ASSET" '.assets[]? | select(.name == $name) | .browser_download_url' | head -n1)
CHECKSUM_URL=$(echo "$RELEASE_JSON" | jq -r --arg name "${ASSET}.sha256sum" '.assets[]? | select(.name == $name) | .browser_download_url' | head -n1)

if [[ -z $TAG_NAME || -z $ASSET_URL || $ASSET_URL == "null" ]]; then
  echo "Error: Could not resolve ${ASSET} in Deno release ${VERSION}." 1>&2
  exit 1
fi

# Skip the download if the requested version is already installed.
if [[ -x $DESTBIN ]] && [[ $("$DESTBIN" --version 2>/dev/null | awk 'NR==1 {print "v"$2}') == "$TAG_NAME" ]]; then
  echo "Deno ${TAG_NAME} is already installed in \"${DESTBIN}\"."
  exit 0
fi

echo "--------------------------------------------------------------------------------"
echo "VERSION : ${TAG_NAME}"
echo "ASSET   : ${ASSET}"
echo "DOWNLOAD: ${ASSET_URL}"
echo "DESTBIN : ${DESTBIN}"
echo "--------------------------------------------------------------------------------"

tmp_dir=$(mktemp -d)
trap 'rm -rf "${tmp_dir}"' EXIT

echo "Downloading ${ASSET}..."
curl --fail --silent --show-error --location --retry 3 --retry-delay 2 --retry-all-errors "$ASSET_URL" -o "${tmp_dir}/${ASSET}"

# Verify the archive against the checksum file published with the release before installing.
if [[ -z $CHECKSUM_URL || $CHECKSUM_URL == "null" ]]; then
  echo "Error: Deno ${TAG_NAME} publishes no checksum for ${ASSET}; refusing to install." 1>&2
  exit 1
fi
curl --fail --silent --show-error --location --retry 3 --retry-delay 2 --retry-all-errors "$CHECKSUM_URL" -o "${tmp_dir}/${ASSET}.sha256sum"
expected=$(awk 'NR==1 {print $1}' "${tmp_dir}/${ASSET}.sha256sum")
if [[ ! $expected =~ ^[0-9a-fA-F]{64}$ ]]; then
  echo "Error: ${ASSET}.sha256sum contains no valid SHA-256 checksum; refusing to install." 1>&2
  exit 1
fi
actual=$(sha256sum "${tmp_dir}/${ASSET}" | awk '{print $1}')
if [[ ${expected,,} != "${actual,,}" ]]; then
  echo "Error: SHA-256 mismatch for ${ASSET}." 1>&2
  echo "  expected: ${expected}" 1>&2
  echo "  actual:   ${actual}" 1>&2
  exit 1
fi
echo "Checksum OK (${actual})."

unzip -q -o "${tmp_dir}/${ASSET}" deno -d "${tmp_dir}"

echo "Installing Deno to \"${DESTBIN}\" with permissions 755..."
mkdir -p "${DESTDIR}/bin"
install -m 755 "${tmp_dir}/deno" "${DESTBIN}"

"${DESTBIN}" --version | head -n1
echo "Done."
