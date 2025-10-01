#!/usr/bin/env bash

# Installs the yt-dlp binary, a vector search engine, on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-yt-dlp.sh)

set -euo pipefail

if ! command -v jq >/dev/null 2>&1; then
  echo "Error: jq is required but not installed." 1>&2
  exit 1
fi

# Show usage information if first argument is --help.
if [[ ${1:-} == "--help" ]]; then
  echo "Usage: ${0##*/} [destdir] [version]" 1>&2
  exit 0
fi

# You can provide a custom installation directory as the first argument.
DESTDIR=$(realpath "${1:-/usr/local}")

# Determine target architecture.
if [[ -n ${PHOTOPRISM_ARCH:-} ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

DESTARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

case $DESTARCH in
  amd64 | AMD64 | x86_64 | x86-64)
    DESTARCH=x86_64
    ;;

  arm64 | ARM64 | aarch64)
    DESTARCH=aarch64
    ;;

  arm | ARM | aarch | armv7l | armhf)
    DESTARCH=armv7l
    ;;

  *)
    echo "Unsupported Machine Architecture: \"$DESTARCH\"" 1>&2
    exit 1
    ;;
esac

. /etc/os-release

# Abort if not executed as root.
if [[ $(id -u) != "0" ]] && [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Error: Run ${0##*/} as root to install in a system directory!" 1>&2
  exit 1
fi

echo "Installing yt-dlp for ${DESTARCH^^}..."

# Determine the list of acceptable asset names for the requested architecture.
case $DESTARCH in
  x86_64)
    ASSET_CANDIDATES=("yt-dlp_linux" "yt-dlp_linux_x86_64" "yt-dlp")
    ;;
  aarch64)
    ASSET_CANDIDATES=("yt-dlp_linux_aarch64" "yt-dlp_linux_arm64")
    ;;
  armv7l)
    ASSET_CANDIDATES=("yt-dlp_linux_armv7l" "yt-dlp_linux_armv7" "yt-dlp_linux_armhf")
    ;;
  *)
    echo "Unsupported Machine Architecture: \"$DESTARCH\"" 1>&2
    exit 1
    ;;
esac

DEFAULT_RELEASES_URL="https://api.github.com/repos/yt-dlp/yt-dlp/releases?per_page=5"

if [[ -n ${2:-} ]]; then
  VERSION=${2}
  RELEASES_JSON=$(curl --fail --silent --show-error "https://api.github.com/repos/yt-dlp/yt-dlp/releases/tags/${VERSION}" || true)
  if [[ -z $RELEASES_JSON ]]; then
    echo "Error: Unable to fetch release metadata for tag ${VERSION}." 1>&2
    exit 1
  fi
  RELEASES_JSON="[${RELEASES_JSON}]"
else
  RELEASES_JSON=$(curl --fail --silent --show-error "$DEFAULT_RELEASES_URL" || true)
  if [[ -z $RELEASES_JSON ]]; then
    echo "Error: Unable to fetch release metadata from GitHub." 1>&2
    exit 1
  fi
fi

TAG_NAME=""
ASSET_NAME=""
ASSET_URL=""

while IFS= read -r release; do
  tag=$(echo "$release" | jq -r '.tag_name // empty')
  [[ -z $tag ]] && continue

  for candidate in "${ASSET_CANDIDATES[@]}"; do
    url=$(echo "$release" | jq -r --arg name "$candidate" '.assets[]? | select(.name == $name) | .browser_download_url' | head -n1)
    if [[ -n $url && $url != "null" ]]; then
      TAG_NAME=$tag
      ASSET_NAME=$candidate
      ASSET_URL=$url
      break 2
    fi
  done
done < <(echo "$RELEASES_JSON" | jq -c '.[]')

if [[ -z ${TAG_NAME} || -z ${ASSET_URL} ]]; then
  echo "Error: Could not resolve a downloadable asset for architecture ${DESTARCH}." 1>&2
  exit 1
fi

# Capture the most recent release tag for informational purposes.
LATEST_TAG=$(echo "$RELEASES_JSON" | jq -r '.[0].tag_name // empty')
VERSION=$TAG_NAME
GITHUB_URL=$ASSET_URL
DESTBIN="${DESTDIR}/bin/yt-dlp"

echo "--------------------------------------------------------------------------------"
echo "VERSION : ${VERSION}"
echo "LATEST  : ${LATEST_TAG:-unknown}"
echo "ASSET   : ${ASSET_NAME}"
echo "DOWNLOAD: ${GITHUB_URL}"
echo "DESTDIR : ${DESTDIR}"
echo "DESTBIN : ${DESTBIN}"
echo "--------------------------------------------------------------------------------"

echo "Downloading the yt-dlp binary to \"${DESTBIN}\"..."
mkdir -p "${DESTDIR}"
curl --fail --silent --show-error --location "${GITHUB_URL}" -o "${DESTBIN}"

echo "Changing permissions of \"${DESTBIN}\" to 755..."
chmod 755 "${DESTBIN}"

echo "Done."
