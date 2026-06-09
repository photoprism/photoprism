#!/usr/bin/env bash

# Installs the heif-dec, heif-enc, and heif-info binaries on Linux.
# On libheif 1.21+, heif-convert is a symlink to heif-dec and heif-thumbnailer is no longer shipped.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-libheif.sh)
#
# Resolute installs photoprism-libheif via a real .deb that Provides/Replaces/Conflicts
# apt's libheif1, libheif-dev, libheif-examples, and libheif-plugin-* names. Other Debian
# distros get a contentless stub with the same metadata after the tarball extract. Either
# way, the package is apt-mark held so dist-upgrade leaves it alone — but an explicit
# `apt install libheif1` would still trigger the Conflicts: relation and replace
# photoprism-libheif, silently breaking the from-source HEIC pipeline. Don't do that;
# upgrade by bumping LIBHEIF_VERSION and rerunning this script instead.

set -e

# Show usage information if first argument is --help.
if [[ ${1} == "--help" ]]; then
  echo "Usage: ${0##*/} [destdir] [version]" 1>&2
  exit 0
fi

# You can provide a custom installation directory as the first argument.
DESTDIR=$(realpath "${1:-/usr/local}")

# In addition, you can specify a custom version to be installed as the second argument.
LIBHEIF_VERSION=${2:-v1.22.2}

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
    DESTARCH=arm
    ;;

  *)
    echo "Unsupported Machine Architecture: \"$DESTARCH\"" 1>&2
    exit 1
    ;;
esac

# shellcheck source=/dev/null
. /etc/os-release

# Abort if not executed as root.
if [[ $(id -u) != "0" ]] && [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Error: Run ${0##*/} as root to install in a system directory!" 1>&2
  exit 1
fi

mkdir -p "$DESTDIR"

# Map codenames to find and use a compatible version.
case $VERSION_CODENAME in
  vera | virginia)
    VERSION_CODENAME=jammy
    ;;
esac

echo "Installing libheif..."

# On Ubuntu 26.04 LTS (Resolute) we ship a real .deb (photoprism-libheif) so
# apt's libheif1 / libheif-dev / libheif-plugin-* don't coexist with the
# from-source build. The .deb only exists when DESTDIR is left at the default
# /usr or /usr/local — dpkg paths are absolute and ignore custom prefixes.
if [[ $VERSION_CODENAME == "resolute" ]] && [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]] && command -v apt-get > /dev/null; then
  ARCHIVE="libheif-${VERSION_CODENAME}-${DESTARCH}-${LIBHEIF_VERSION}.deb"
  URL="https://dl.photoprism.app/dist/libheif/${ARCHIVE}"
  TMPDEB="/tmp/${ARCHIVE}"

  echo "--------------------------------------------------------------------------------"
  echo "VERSION: $LIBHEIF_VERSION"
  echo "PACKAGE: $ARCHIVE (photoprism-libheif Debian package)"
  echo "--------------------------------------------------------------------------------"

  if ! curl -fsSL "$URL" -o "$TMPDEB"; then
    echo "❌ Failed to download \"$URL\"."
    exit 1
  fi

  export DEBIAN_FRONTEND=noninteractive
  if apt-get install -y --no-install-recommends "$TMPDEB"; then
    apt-mark hold photoprism-libheif > /dev/null
    rm -f "$TMPDEB"
    echo "✅ Installed photoprism-libheif from \"$URL\"."
    echo "Done."
    exit 0
  fi

  echo "⚠️ apt-get install of \"$TMPDEB\" failed; falling back to tarball path."
  rm -f "$TMPDEB"
fi

ARCHIVE="libheif-${VERSION_CODENAME}-${DESTARCH}-${LIBHEIF_VERSION}.tar.gz"
URL="https://dl.photoprism.app/dist/libheif/${ARCHIVE}"

echo "--------------------------------------------------------------------------------"
echo "VERSION: $LIBHEIF_VERSION"
echo "ARCHIVE: $ARCHIVE"
echo "DESTDIR: $DESTDIR"
echo "--------------------------------------------------------------------------------"

if curl -fsSL "$URL" | tar --overwrite --mode=755 -xz -C "$DESTDIR" 2> /dev/null; then
  echo "✅ Extracted \"$URL\" to \"$DESTDIR\""
else
  echo "❌ No libheif binaries are available for this architecture or distribution."
  exit 0
fi

if [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Running \"ldconfig\"..."
  ldconfig
else
  echo "Running \"ldconfig -n $DESTDIR/lib\"..."
  ldconfig -n "$DESTDIR/lib"
fi

# Replace any distro-shipped libheif* packages with a contentless photoprism-libheif
# stub that Provides/Replaces/Conflicts them. Apt's solver then accepts the in-image
# install as a substitute and won't pull the older distro library back in on the next
# dist-upgrade. The actual binaries live under $DESTDIR (typically /usr/local/lib),
# and ldconfig already orders /usr/local/lib ahead of /usr/lib so consumers resolve
# to the from-source build. Resolute exits earlier through the real .deb path above.
if [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]] && command -v apt-get > /dev/null && command -v dpkg-deb > /dev/null; then
  INSTALLED_LIBHEIF_PKGS=$(dpkg-query -W -f='${Package}\n' 'libheif*' 2>/dev/null | grep -v '^photoprism-libheif$' | sort -u || true)
  if [[ -n $INSTALLED_LIBHEIF_PKGS ]]; then
    UPSTREAM_VERSION=${LIBHEIF_VERSION#v}
    STUB_VERSION="${UPSTREAM_VERSION}-photoprism1"
    STUB_ARCH=$(dpkg --print-architecture)
    STUB_DIR=$(mktemp -d)
    trap 'rm -rf "$STUB_DIR"' EXIT

    # Build comma-separated Provides:/Replaces:/Conflicts: from the actually-installed
    # set, so per-distro naming (libheif1 vs libheif1t64, plugin variants) is handled
    # without hard-coded per-codename branches.
    STUB_PROVIDES=$(printf '%s\n' "$INSTALLED_LIBHEIF_PKGS" | sed "s/\$/ (= $UPSTREAM_VERSION)/" | paste -sd ',' | sed 's/,/, /g')
    STUB_REPLACES=$(printf '%s\n' "$INSTALLED_LIBHEIF_PKGS" | paste -sd ',' | sed 's/,/, /g')

    echo "--------------------------------------------------------------------------------"
    echo "Generating photoprism-libheif stub for apt-installed packages:"
    # shellcheck disable=SC2001
    echo "$INSTALLED_LIBHEIF_PKGS" | sed 's/^/  /'
    echo "--------------------------------------------------------------------------------"

    mkdir -p "$STUB_DIR/DEBIAN" "$STUB_DIR/usr/share/doc/photoprism-libheif"
    cat > "$STUB_DIR/DEBIAN/control" <<STUBEOF
Package: photoprism-libheif
Version: $STUB_VERSION
Architecture: $STUB_ARCH
Multi-Arch: foreign
Maintainer: PhotoPrism UG <hello@photoprism.app>
Section: libs
Priority: optional
Provides: $STUB_PROVIDES
Replaces: $STUB_REPLACES
Conflicts: $STUB_REPLACES
Description: libheif ${LIBHEIF_VERSION} stub for PhotoPrism's from-source install
 Contentless package that satisfies apt's libheif* dependencies after
 install-libheif.sh has placed the real binaries under $DESTDIR. Apt's
 solver treats this as the libheif provider so the distro libheif1 and
 libheif-plugin-* packages are not pulled back in on the next dist-upgrade.
 .
 Real binaries: $DESTDIR/bin/heif-* and $DESTDIR/lib/libheif.so.*
STUBEOF
    cat > "$STUB_DIR/usr/share/doc/photoprism-libheif/copyright" <<STUBEOF
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: libheif
Source: https://github.com/strukturag/libheif

Files: *
Copyright: 2017-2026 struktur AG
License: LGPL-3.0+
 See /usr/share/common-licenses/LGPL-3 for the full text.
STUBEOF

    STUB_DEB="$STUB_DIR/photoprism-libheif_${STUB_VERSION}_${STUB_ARCH}.deb"
    dpkg-deb --build --root-owner-group "$STUB_DIR" "$STUB_DEB" > /dev/null

    export DEBIAN_FRONTEND=noninteractive
    if apt-get install -y --no-install-recommends "$STUB_DEB"; then
      apt-mark hold photoprism-libheif > /dev/null
      echo "✅ Installed photoprism-libheif stub (apt's libheif1/libheif-plugin-* superseded)."
    else
      echo "⚠️ Failed to install photoprism-libheif stub; apt's libheif packages remain alongside the from-source binaries."
    fi

    rm -rf "$STUB_DIR"
    trap - EXIT
  fi
fi

echo "Done."
