#!/usr/bin/env bash

# Installs a static FFmpeg build from BtbN/FFmpeg-Builds.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-ffmpeg.sh) [destdir] [version]

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

BTBN_REPO="BtbN/FFmpeg-Builds"
BTBN_BASE="https://github.com/${BTBN_REPO}/releases/download/latest"
BTBN_API="https://api.github.com/repos/${BTBN_REPO}/releases/latest"

if [[ ${1} == "--help" ]]; then
  echo "Installs a static FFmpeg build from BtbN/FFmpeg-Builds." 1>&2
  echo "Usage: ${0##*/} [destdir] [version]" 1>&2
  echo "" 1>&2
  echo "Arguments:" 1>&2
  echo "  destdir   Installation directory (default: /opt/ffmpeg)" 1>&2
  echo "  version   FFmpeg version to install (default: latest)" 1>&2
  echo "" 1>&2
  echo "Supported versions:" 1>&2
  echo "  latest    Latest stable release from BtbN (default)" 1>&2
  echo "  master    Latest nightly build from BtbN" 1>&2
  exit 0
fi

DESTDIR=$(realpath "${1:-/opt/ffmpeg}")
FFMPEG_VERSION=${2:-latest}

# Determine target architecture.
if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

DESTARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

case $DESTARCH in
  amd64 | AMD64 | x86_64 | x86-64)
    DESTARCH=amd64
    BTBN_ARCH="linux64"
    ;;
  arm64 | ARM64 | aarch64)
    DESTARCH=arm64
    BTBN_ARCH="linuxarm64"
    ;;
  *)
    echo "Unsupported architecture: \"$DESTARCH\" (BtbN builds are available for amd64 and arm64 only)" 1>&2
    exit 1
    ;;
esac

echo "Installing FFmpeg..."

# Determine download URL.
# - "master" → nightly build (ffmpeg-master-latest-*)
# - "latest" → latest stable release (ffmpeg-nX.Y-latest-*)
case $FFMPEG_VERSION in
  master)
    ARCHIVE="ffmpeg-master-latest-${BTBN_ARCH}-gpl.tar.xz"
    URL="${BTBN_BASE}/${ARCHIVE}"
    ;;
  latest)
    # Discover the highest stable version from BtbN release assets.
    # Asset names follow the pattern: ffmpeg-n8.0-latest-linux64-gpl-8.0.tar.xz
    ARCHIVE=$(curl -sSf "$BTBN_API" \
      | grep -oE "ffmpeg-n[0-9]+\.[0-9]+-latest-${BTBN_ARCH}-gpl-[0-9]+\.[0-9]+\.tar\.xz" \
      | sort -rV \
      | head -1)

    if [[ -z $ARCHIVE ]]; then
      echo "Error: Could not determine latest stable FFmpeg version from BtbN." 1>&2
      echo "Please check your network connection and try again." 1>&2
      exit 1
    fi

    URL="${BTBN_BASE}/${ARCHIVE}"
    ;;
  *)
    echo "Error: Unsupported version '${FFMPEG_VERSION}'." 1>&2
    echo "Use 'latest' for the latest stable release or 'master' for nightly builds." 1>&2
    exit 1
    ;;
esac

echo "VERSION: $FFMPEG_VERSION"
echo "ARCHIVE: $ARCHIVE"
echo "DESTDIR: $DESTDIR"
echo "SOURCE:  BtbN/FFmpeg-Builds"
echo ""
echo "Downloading from: $URL"
sudo mkdir -p "${DESTDIR}"

if ! curl -fsSL "$URL" | sudo tar --strip-components=1 --overwrite --mode=755 -x --xz -C "$DESTDIR"; then
  echo "Error: Failed to download or extract FFmpeg archive." 1>&2
  echo "Please check your network connection and try again." 1>&2
  exit 1
fi

sudo chown -R root:root "${DESTDIR}"

# Locate ffmpeg binary (BtbN archives contain a bin/ subdirectory).
FFMPEG_BIN="${DESTDIR}/bin/ffmpeg"
FFPROBE_BIN="${DESTDIR}/bin/ffprobe"

if [[ ! -x "${FFMPEG_BIN}" ]]; then
  echo "Error: Could not find ffmpeg binary in ${DESTDIR}/bin" 1>&2
  exit 1
fi

# Create symbolic links.
sudo ln -sf "${FFMPEG_BIN}" /usr/local/bin/ffmpeg
sudo ln -sf "${FFPROBE_BIN}" /usr/local/bin/ffprobe

# Verify installation.
if command -v /usr/local/bin/ffmpeg &> /dev/null; then
  echo ""
  echo "FFmpeg installed successfully:"
  /usr/local/bin/ffmpeg -version | head -1
  echo ""
  echo "Symlinks:"
  echo "  /usr/local/bin/ffmpeg -> ${FFMPEG_BIN}"
  echo "  /usr/local/bin/ffprobe -> ${FFPROBE_BIN}"
else
  echo "Warning: FFmpeg installation could not be verified." 1>&2
fi

echo ""
echo "Done."
