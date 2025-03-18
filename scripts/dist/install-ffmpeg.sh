#!/usr/bin/env bash

# Installs the static FFmpeg build available at https://johnvansickle.com/ffmpeg/.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-ffmpeg.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Show usage information if first argument is --help.
if [[ ${1} == "--help" ]]; then
  echo "Installs the release or latest static build of FFmpeg from the internet." 1>&2
  echo "Usage: ${0##*/} [destdir] [version]" 1>&2
  exit 0
fi

# You can provide a custom installation directory as the first argument.
DESTDIR=$(realpath "${1:-/opt/ffmpeg}")

# In addition, you can specify a custom version as the second argument.
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

. /etc/os-release

echo "Installing FFmpeg..."

if [[ $FFMPEG_VERSION == "latest" ]] && [[ $DESTARCH == "amd64" ]]; then
  # https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz
  ARCHIVE="ffmpeg-master-${FFMPEG_VERSION}-linux64-gpl.tar.xz"
  URL="https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/${ARCHIVE}"
elif [[ $FFMPEG_VERSION == "latest" ]] && [[ $DESTARCH == "arm64" ]]; then
  # https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linuxarm64-gpl.tar.xz
  ARCHIVE="ffmpeg-master-${FFMPEG_VERSION}-linuxarm64-gpl.tar.xz"
  URL="https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/${ARCHIVE}"
else
  ARCHIVE="ffmpeg-${FFMPEG_VERSION}-${DESTARCH}-static.tar.xz"
  URL="https://johnvansickle.com/ffmpeg/releases/${ARCHIVE}"
  DESTDIR="${DESTDIR}/bin"
fi

echo "------------------------------------------------"
echo "VERSION: $FFMPEG_VERSION"
echo "ARCHIVE: $ARCHIVE"
echo "DESTDIR: $DESTDIR"
echo "------------------------------------------------"

echo "Extracting \"$URL\" to \"$DESTDIR\"."
sudo mkdir -p "${DESTDIR}"
curl -fsSL "$URL" | sudo tar --strip-components=1 --overwrite --mode=755 -x --xz -C "$DESTDIR"
sudo chown -R root:root "${DESTDIR}"

# Create a symbolic link to the static ffmpeg binary.
sudo ln -sf "${DESTDIR}/bin/ffmpeg" /usr/local/bin/ffmpeg

echo "Done."
