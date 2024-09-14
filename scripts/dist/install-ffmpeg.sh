#!/usr/bin/env bash

# This installs the static FFmpeg build available at https://johnvansickle.com/ffmpeg/.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-ffmpeg.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Show usage information if first argument is --help.
if [[ ${1} == "--help" ]]; then
  echo "This installs the static FFmpeg build available at https://johnvansickle.com/ffmpeg/." 1>&2
  echo "Usage: ${0##*/} [destdir] [version]" 1>&2
  exit 0
fi

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

# You can provide a custom installation directory as the first argument.
DESTDIR=$(realpath "${1:-/opt/ffmpeg}")

# In addition, you can specify a custom version as the second argument.
FFMPEG_VERSION=${2:-release}

# Determine the system architecture.
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

. /etc/os-release

echo "Installing FFmpeg..."

ARCHIVE="ffmpeg-${FFMPEG_VERSION}-${DESTARCH}-static.tar.xz"
URL="https://johnvansickle.com/ffmpeg/releases/${ARCHIVE}"

echo "------------------------------------------------"
echo "VERSION: $FFMPEG_VERSION"
echo "ARCHIVE: $ARCHIVE"
echo "DESTDIR: $DESTDIR"
echo "------------------------------------------------"

echo "Extracting \"$URL\" to \"$DESTDIR\"."
mkdir -p "${DESTDIR}"
wget --inet4-only -c "$URL" -O - | tar --strip-components=1 --overwrite --mode=755 -x --xz -C "$DESTDIR"
chown -R root:root "${DESTDIR}"

# Create a symbolic link to the static ffmpeg binary.
ln -sf "${DESTDIR}/ffmpeg" /usr/local/bin/ffmpeg

echo "Done."
