#!/usr/bin/env bash

# This builds the heif-convert, heif-enc, heif-info and heif-thumbnailer binaries from source.
#
# To create ARMv7 binaries with Docker on Ubuntu 22.04 LTS, you can e.g. run the following:
#
#   docker run --rm --platform=arm --pull=always -v ".:/go/src/github.com/photoprism/photoprism" \
#   -e BUILD_ARCH=arm -e SYSTEM_ARCH=arm --entrypoint "" photoprism/develop:jammy ./scripts/dist/build-libheif.sh

# Show usage information if first argument is --help.
if [[ ${1} == "--help" ]]; then
  echo "Usage: ${0##*/} [version]" 1>&2
  exit 0
fi

CURRENT_DIR=$(pwd)

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
    DESTARCH=arm
    ;;

  *)
    echo "Unsupported Machine Architecture: \"$DESTARCH\"" 1>&2
    exit 1
    ;;
esac

. /etc/os-release

LATEST=$(curl --silent "https://api.github.com/repos/strukturag/libheif/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
LIBHEIF_VERSION=${1:-$LATEST}

BUILD="libheif-$VERSION_CODENAME-$DESTARCH-$LIBHEIF_VERSION"

DESTDIR="${CURRENT_DIR}/build/$BUILD"

mkdir -p "$DESTDIR"

ARCHIVE="${CURRENT_DIR}/build/$BUILD.tar.gz"

echo "------------------------------------------------"
echo "VERSION: $LIBHEIF_VERSION"
echo "LATEST : $LATEST"
echo "ARCHIVE: $ARCHIVE"
echo "------------------------------------------------"

echo "Installing build dependencies..."

sudo apt-get -qq update
sudo apt-get -qq install build-essential gcc g++ gettext git autoconf automake cmake libtool libjpeg-dev libpng-dev libwebp-dev libde265-dev libaom-dev libavcodec-dev

if [[ $VERSION_CODENAME == "noble" ]]; then
  sudo apt-get -qq install libsharpyuv-dev
fi

cd "/tmp" || exit
rm -rf "/tmp/libheif"

echo "Cloning git repository..."
git clone -c advice.detachedHead=false -b "$LIBHEIF_VERSION" --depth 1 https://github.com/strukturag/libheif.git libheif
cd libheif || exit
(mkdir build && cd build && cmake --preset=release ..)
make -C build

# Install heif-convert, heif-enc, heif-info, and heif-thumbnailer in "/usr/local".
echo "Installing binaries..."
DESTDIR=$DESTDIR make -C build install
cd "$CURRENT_DIR" || exit
rm -rf "/tmp/libheif"

# Create a tar archive to distribute the binaries.
echo "Creating $ARCHIVE..."
tar -czf "$ARCHIVE" -C "$DESTDIR/usr/local" bin lib

echo "Done."
