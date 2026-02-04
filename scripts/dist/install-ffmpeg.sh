#!/usr/bin/env bash

# Installs a static FFmpeg build from BtbN/FFmpeg-Builds or johnvansickle.com.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-ffmpeg.sh) [destdir] [version]

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

if [[ ${1} == "--help" ]]; then
  echo "Installs a static FFmpeg build." 1>&2
  echo "Usage: ${0##*/} [destdir] [version]" 1>&2
  echo "" 1>&2
  echo "Arguments:" 1>&2
  echo "  destdir   Installation directory (default: /opt/ffmpeg)" 1>&2
  echo "  version   FFmpeg version to install (default: release)" 1>&2
  echo "" 1>&2
  echo "Supported versions:" 1>&2
  echo "  latest    Latest git master from BtbN (amd64/arm64 only, recommended)" 1>&2
  echo "  release   Latest stable release from johnvansickle.com (default)" 1>&2
  echo "  6.0.1     Specific version from johnvansickle.com old-releases" 1>&2
  exit 0
fi

DESTDIR=$(realpath "${1:-/opt/ffmpeg}")
FFMPEG_VERSION=${2:-release}

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
    ;;
  arm64 | ARM64 | aarch64)
    DESTARCH=arm64
    ;;
  arm | ARM | aarch | armv7l | armhf)
    DESTARCH=armhf
    ;;
  *)
    echo "Unsupported Machine Architecture: \"$DESTARCH\"" 1>&2
    exit 1
    ;;
esac

echo "Installing FFmpeg..."

# Determine download URL and source.
# - "latest" → BtbN/FFmpeg-Builds (amd64/arm64 only)
# - "release" → johnvansickle.com/ffmpeg/releases/
# - specific version → johnvansickle.com/ffmpeg/old-releases/
USE_BTBN=false

if [[ $FFMPEG_VERSION == "latest" ]]; then
  case $DESTARCH in
    amd64)
      USE_BTBN=true
      ARCHIVE="ffmpeg-master-latest-linux64-gpl.tar.xz"
      URL="https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/${ARCHIVE}"
      ;;
    arm64)
      USE_BTBN=true
      ARCHIVE="ffmpeg-master-latest-linuxarm64-gpl.tar.xz"
      URL="https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/${ARCHIVE}"
      ;;
    *)
      echo "BtbN builds not available for ${DESTARCH}, using johnvansickle.com release instead." 1>&2
      FFMPEG_VERSION="release"
      ;;
  esac
fi

if [[ $USE_BTBN == false ]]; then
  if [[ $FFMPEG_VERSION == "release" ]]; then
    ARCHIVE="ffmpeg-release-${DESTARCH}-static.tar.xz"
    URL="https://johnvansickle.com/ffmpeg/releases/${ARCHIVE}"
  else
    ARCHIVE="ffmpeg-${FFMPEG_VERSION}-${DESTARCH}-static.tar.xz"
    URL="https://johnvansickle.com/ffmpeg/old-releases/${ARCHIVE}"
  fi
fi

DESTDIR="${DESTDIR}/bin"

echo "VERSION: $FFMPEG_VERSION"
echo "ARCHIVE: $ARCHIVE"
echo "DESTDIR: $DESTDIR"
if [[ $USE_BTBN == true ]]; then
  echo "SOURCE:  BtbN/FFmpeg-Builds"
else
  echo "SOURCE:  johnvansickle.com"
fi

echo ""
echo "Downloading from: $URL"
sudo mkdir -p "${DESTDIR}"

if ! curl -fsSL "$URL" | sudo tar --strip-components=1 --overwrite --mode=755 -x --xz -C "$DESTDIR"; then
  echo "Error: Failed to download or extract FFmpeg archive." 1>&2
  echo "Please check that the version '${FFMPEG_VERSION}' exists and try again." 1>&2
  exit 1
fi

sudo chown -R root:root "${DESTDIR}"

# Locate ffmpeg binary (BtbN: bin/, JVS: root).
if [[ -x "${DESTDIR}/bin/ffmpeg" ]]; then
  FFMPEG_BIN="${DESTDIR}/bin/ffmpeg"
  FFPROBE_BIN="${DESTDIR}/bin/ffprobe"
else
  FFMPEG_BIN="${DESTDIR}/ffmpeg"
  FFPROBE_BIN="${DESTDIR}/ffprobe"
fi

if [[ ! -x "${FFMPEG_BIN}" ]]; then
  echo "Error: Could not find ffmpeg binary in ${DESTDIR}" 1>&2
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
