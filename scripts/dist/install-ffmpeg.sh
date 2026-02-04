#!/usr/bin/env bash

# Installs a static FFmpeg build from johnvansickle.com.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-ffmpeg.sh)


PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

if [[ ${1} == "--help" ]]; then
  echo "Installs a static FFmpeg build from johnvansickle.com." 1>&2
  echo "Usage: ${0##*/} [destdir] [version]" 1>&2
  echo "" 1>&2
  echo "Arguments:" 1>&2
  echo "  destdir   Installation directory (default: /opt/ffmpeg)" 1>&2
  echo "  version   FFmpeg version to install (default: release)" 1>&2
  echo "" 1>&2
  echo "Supported versions:" 1>&2
  echo "  release   Latest stable release, currently 7.0.2 (default)" 1>&2
  echo "  latest    Latest git master build (recommended)" 1>&2
  echo "  6.0.1     Specific version from old-releases (6.0.1, 6.0, 5.1.1, ...)" 1>&2
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

# shellcheck source=/dev/null
. /etc/os-release

echo "Installing FFmpeg..."

# Determine download URL.
# - "latest" → https://johnvansickle.com/ffmpeg/builds/ (git master)
# - "release" → https://johnvansickle.com/ffmpeg/releases/
# - specific version → https://johnvansickle.com/ffmpeg/old-releases/
if [[ $FFMPEG_VERSION == "latest" ]]; then
  ARCHIVE="ffmpeg-git-${DESTARCH}-static.tar.xz"
  URL="https://johnvansickle.com/ffmpeg/builds/${ARCHIVE}"
elif [[ $FFMPEG_VERSION == "release" ]]; then
  ARCHIVE="ffmpeg-release-${DESTARCH}-static.tar.xz"
  URL="https://johnvansickle.com/ffmpeg/releases/${ARCHIVE}"
else
  ARCHIVE="ffmpeg-${FFMPEG_VERSION}-${DESTARCH}-static.tar.xz"
  URL="https://johnvansickle.com/ffmpeg/old-releases/${ARCHIVE}"
fi

DESTDIR="${DESTDIR}/bin"

echo "VERSION: $FFMPEG_VERSION"
echo "ARCHIVE: $ARCHIVE"
echo "DESTDIR: $DESTDIR"

echo ""
echo "Extracting \"$URL\" to \"$DESTDIR\"..."
sudo mkdir -p "${DESTDIR}"

if ! curl -fsSL "$URL" | sudo tar --strip-components=1 --overwrite --mode=755 -x --xz -C "$DESTDIR"; then
  echo "Error: Failed to download or extract FFmpeg archive." 1>&2
  echo "Please check that the version '${FFMPEG_VERSION}' exists and try again." 1>&2
  exit 1
fi

sudo chown -R root:root "${DESTDIR}"

FFMPEG_BIN="${DESTDIR}/ffmpeg"
FFPROBE_BIN="${DESTDIR}/ffprobe"

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
