#!/usr/bin/env bash

# Builds a libheif Debian package (photoprism-libheif) from source.
#
# Produces a single .deb that ships the libheif shared library, headers, CLI
# tools (heif-dec, heif-enc, heif-info, heif-convert, heif-thumbnailer), and
# all codec plugins (aomdec/aomenc/dav1d/ffmpegdec/j2kdec/j2kenc/jpegdec/
# jpegenc/libde265/rav1e/x265). Declares Provides:/Replaces:/Conflicts: for
# the distro-packaged libheif1, libheif-dev, libheif-examples, and
# libheif-plugin-* so apt's dependency solver treats the in-image install as
# a complete substitute and never pulls Canonical's older 1.21.2 alongside.
#
# Run inside a photoprism/develop:<codename> container that matches the
# target distro's libc / libstdc++ / codec ABI. Currently exercised against
# photoprism/develop:resolute (Ubuntu 26.04 LTS, libc 2.38, libstdc++ 13);
# other codenames need their own .deb because runtime soname pinning differs.
#
#   docker run --rm --platform=amd64 --pull=always \
#     -v ".:/go/src/github.com/photoprism/photoprism" \
#     -e BUILD_ARCH=amd64 -e SYSTEM_ARCH=amd64 \
#     photoprism/develop:resolute ./scripts/dist/build-libheif-deb.sh v1.22.2

set -e

if [[ ${1} == "--help" ]]; then
  echo "Usage: ${0##*/} [version]" 1>&2
  exit 0
fi

CURRENT_DIR=$(pwd)

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
    MULTIARCH_TRIPLE=x86_64-linux-gnu
    ;;

  arm64 | ARM64 | aarch64)
    DESTARCH=arm64
    MULTIARCH_TRIPLE=aarch64-linux-gnu
    ;;

  *)
    echo "Unsupported Machine Architecture: \"$DESTARCH\"" 1>&2
    echo "Only amd64 and arm64 are supported for .deb packaging." 1>&2
    exit 1
    ;;
esac

# shellcheck source=/dev/null
. /etc/os-release

# Only Resolute (Ubuntu 26.04) is currently in the .deb matrix; abort early
# on other codenames to avoid silently producing a .deb that would dpkg -i
# onto a host with mismatched libc/libstdc++ sonames.
if [[ $VERSION_CODENAME != "resolute" ]]; then
  echo "Error: ${0##*/} currently supports only VERSION_CODENAME=resolute," 1>&2
  echo "       this container reports VERSION_CODENAME=$VERSION_CODENAME." 1>&2
  exit 1
fi

LATEST=$(curl --silent "https://api.github.com/repos/strukturag/libheif/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
LIBHEIF_VERSION=${1:-$LATEST}
UPSTREAM_VERSION=${LIBHEIF_VERSION#v}
DEB_VERSION="${UPSTREAM_VERSION}-photoprism1"

BUILD="libheif-${VERSION_CODENAME}-${DESTARCH}-${LIBHEIF_VERSION}"

# Build tree (cmake/make scratch space) and Debian staging root.
WORK_DIR="${CURRENT_DIR}/build/${BUILD}-deb"
PKG_ROOT="${WORK_DIR}/pkg"
DEB_FILE="${CURRENT_DIR}/build/${BUILD}.deb"

rm -rf "$WORK_DIR"
mkdir -p "$PKG_ROOT"

echo "--------------------------------------------------------------------------------"
echo "VERSION       : $LIBHEIF_VERSION"
echo "LATEST        : $LATEST"
echo "CODENAME      : $VERSION_CODENAME"
echo "ARCH          : $DESTARCH ($MULTIARCH_TRIPLE)"
echo "DEB PACKAGE   : photoprism-libheif $DEB_VERSION"
echo "DEB FILE      : $DEB_FILE"
echo "--------------------------------------------------------------------------------"

echo "Installing build dependencies..."

sudo apt-get -qq update
sudo apt-get -qq install build-essential gcc g++ gettext git autoconf automake cmake libtool \
  libjpeg-dev libpng-dev libwebp-dev libde265-dev libaom-dev aom-tools libyuv-dev libavcodec-dev \
  libsharpyuv-dev librav1e-dev libdav1d-dev libx265-dev libbrotli-dev zlib1g-dev \
  dpkg-dev fakeroot

cd "/tmp" || exit
rm -rf "/tmp/libheif"

echo "Cloning libheif source..."
git clone -c advice.detachedHead=false -b "$LIBHEIF_VERSION" --depth 1 https://github.com/strukturag/libheif.git libheif
cd libheif || exit

# Configure cmake with multiarch-aware install paths so the resulting .deb
# lands files where Debian/Ubuntu expect them and the embedded plugin loader
# search path (CMAKE_INSTALL_LIBDIR/libheif) matches the package layout.
mkdir build && cd build
cmake --preset=release \
  -DCMAKE_INSTALL_PREFIX=/usr \
  -DCMAKE_INSTALL_LIBDIR="lib/${MULTIARCH_TRIPLE}" \
  -DCMAKE_INSTALL_INCLUDEDIR=include \
  .. || exit 1
make -j"$(nproc)" || exit 1

echo "Staging install tree at $PKG_ROOT..."
DESTDIR="$PKG_ROOT" make install
cd "$CURRENT_DIR" || exit
rm -rf "/tmp/libheif"

# Replace any /usr/local prefix that may have leaked into .pc / .cmake files
# (CMAKE_INSTALL_PREFIX should have prevented this, but cheap defense).
find "$PKG_ROOT/usr/lib/${MULTIARCH_TRIPLE}/pkgconfig" "$PKG_ROOT/usr/lib/${MULTIARCH_TRIPLE}/cmake" \
  -type f \( -name "*.pc" -o -name "*.cmake" \) -exec \
  sed -i "s|/usr/local|/usr|g" {} \; 2>/dev/null || true

# Compute the runtime libc / libstdc++ minimums from the build container so
# the Depends line matches what the binaries actually need.
LIBC_MIN=$(dpkg-query -W -f='${Version}' libc6 | sed -E 's/^([0-9]+\.[0-9]+).*/\1/')
LIBSTDCPP_MIN=$(dpkg-query -W -f='${Version}' libstdc++6 | sed -E 's/^([0-9]+).*/\1/')

mkdir -p "$PKG_ROOT/DEBIAN"
mkdir -p "$PKG_ROOT/usr/share/doc/photoprism-libheif"

cat > "$PKG_ROOT/DEBIAN/control" <<EOF
Package: photoprism-libheif
Source: libheif
Version: $DEB_VERSION
Architecture: $DESTARCH
Multi-Arch: same
Maintainer: PhotoPrism UG <hello@photoprism.app>
Section: libs
Priority: optional
Depends: libc6 (>= ${LIBC_MIN}), libgcc-s1, libstdc++6 (>= ${LIBSTDCPP_MIN}), zlib1g, libbrotli1, libsharpyuv0
Recommends: libde265-0, libaom3, libdav1d7, libavcodec-extra
Provides: libheif1 (= ${UPSTREAM_VERSION}), libheif-dev (= ${UPSTREAM_VERSION}), libheif-examples (= ${UPSTREAM_VERSION}), libheif-plugin-aomdec (= ${UPSTREAM_VERSION}), libheif-plugin-aomenc (= ${UPSTREAM_VERSION}), libheif-plugin-libde265 (= ${UPSTREAM_VERSION}), libheif-plugin-x265 (= ${UPSTREAM_VERSION}), libheif-plugin-dav1d (= ${UPSTREAM_VERSION})
Replaces: libheif1, libheif-dev, libheif-examples, libheif-plugin-aomdec, libheif-plugin-aomenc, libheif-plugin-libde265, libheif-plugin-x265, libheif-plugin-dav1d
Conflicts: libheif1, libheif-dev, libheif-examples, libheif-plugin-aomdec, libheif-plugin-aomenc, libheif-plugin-libde265, libheif-plugin-x265, libheif-plugin-dav1d
Homepage: https://github.com/strukturag/libheif
Description: HEIF and AVIF file format decoder/encoder (PhotoPrism build)
 Built from upstream strukturag/libheif $LIBHEIF_VERSION by PhotoPrism's
 build-libheif-deb.sh script. Includes the heif-dec, heif-enc, heif-info,
 heif-convert, and heif-thumbnailer CLI tools, the libheif shared library,
 the development headers, the pkg-config and CMake integration files, and
 all codec plugins (aomdec, aomenc, dav1d, ffmpegdec, j2kdec, j2kenc,
 jpegdec, jpegenc, libde265, rav1e, x265).
 .
 Supersedes the distribution-packaged libheif1, libheif-dev, libheif-examples,
 and libheif-plugin-* packages. Provides those virtual package names so apt's
 dependency solver treats the in-image install as a complete substitute.
EOF

cat > "$PKG_ROOT/usr/share/doc/photoprism-libheif/copyright" <<EOF
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: libheif
Upstream-Contact: Dirk Farin <dirk.farin@gmail.com>
Source: https://github.com/strukturag/libheif

Files: *
Copyright: 2017-2026 struktur AG
License: LGPL-3.0+
 libheif is free software: you can redistribute it and/or modify it under the
 terms of the GNU Lesser General Public License as published by the Free
 Software Foundation, either version 3 of the License, or (at your option)
 any later version. See /usr/share/common-licenses/LGPL-3 for the full text.
EOF

# Minimal Debian changelog (lintian wants this; we keep it terse since the
# package is rebuilt 1:1 per upstream release rather than carrying its own
# revision history).
cat > "${WORK_DIR}/changelog.Debian" <<EOF
photoprism-libheif (${DEB_VERSION}) unstable; urgency=medium

  * Repackaged upstream libheif ${LIBHEIF_VERSION} for PhotoPrism's
    Resolute (Ubuntu 26.04) base image. Provides/Replaces/Conflicts
    against distro libheif1, libheif-dev, libheif-examples, and
    libheif-plugin-* so the apt solver picks this build as the single
    libheif provider.

 -- PhotoPrism UG <hello@photoprism.app>  $(date -R)
EOF
gzip -9n -c "${WORK_DIR}/changelog.Debian" > "$PKG_ROOT/usr/share/doc/photoprism-libheif/changelog.Debian.gz"

# Set conservative ownership and permissions on the staged tree before dpkg-deb
# walks it; fakeroot keeps dpkg-deb happy without needing real root.
chmod -R u+rwX,go+rX,go-w "$PKG_ROOT"
find "$PKG_ROOT" -type d -exec chmod 755 {} \;
find "$PKG_ROOT/usr/share/doc" -type f -exec chmod 644 {} \;
find "$PKG_ROOT/DEBIAN" -type f -exec chmod 644 {} \;

echo "Building $DEB_FILE..."
mkdir -p "${CURRENT_DIR}/build"
fakeroot dpkg-deb --build --root-owner-group "$PKG_ROOT" "$DEB_FILE"

echo "--------------------------------------------------------------------------------"
dpkg-deb --info "$DEB_FILE"
echo "--------------------------------------------------------------------------------"
dpkg-deb --contents "$DEB_FILE" | head -40
echo "..."
echo "--------------------------------------------------------------------------------"

echo "Done. .deb at $DEB_FILE"
