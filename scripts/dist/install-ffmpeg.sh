#!/usr/bin/env bash

# Installs FFmpeg from distro packages or BtbN/FFmpeg-Builds static archives.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-ffmpeg.sh) [version] [destdir]

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"
set -euo pipefail

BTBN_REPO="BtbN/FFmpeg-Builds"
BTBN_BASE="https://github.com/${BTBN_REPO}/releases/download/latest"
BTBN_API="https://api.github.com/repos/${BTBN_REPO}/releases/latest"

# update_symlinks force-updates the ffmpeg and ffprobe symlinks in /usr/local/bin.
update_symlinks() {
  local ffmpeg_bin="$1"
  local ffprobe_bin="$2"

  "${SUDO[@]}" ln -sf "${ffmpeg_bin}" /usr/local/bin/ffmpeg
  "${SUDO[@]}" ln -sf "${ffprobe_bin}" /usr/local/bin/ffprobe

  echo "Symlinks:"
  echo "  /usr/local/bin/ffmpeg -> ${ffmpeg_bin}"
  echo "  /usr/local/bin/ffprobe -> ${ffprobe_bin}"
}

if [[ ${1:-} == "--help" ]]; then
  echo "Installs FFmpeg from distro packages or BtbN/FFmpeg-Builds static archives." 1>&2
  echo "Usage: ${0##*/} [version] [destdir]" 1>&2
  echo "" 1>&2
  echo "Arguments:" 1>&2
  echo "  version   FFmpeg version to install (default: release)" 1>&2
  echo "  destdir   Optional install directory for static builds (latest/master)" 1>&2
  echo "" 1>&2
  echo "Supported versions:" 1>&2
  echo "  release   Use distro package and repoint symlinks to /usr/bin/*" 1>&2
  echo "  latest    Latest stable static build from BtbN" 1>&2
  echo "  master    Latest nightly static build from BtbN" 1>&2
  exit 0
fi

# You can specify the version as the first argument.
FFMPEG_VERSION=${1:-release}

# For static builds, you can provide a custom installation directory as the second argument.
DESTDIR=${2:-/opt/ffmpeg}

if [[ $(id -u) == "0" ]]; then
  SUDO=()
else
  SUDO=(sudo)
fi

# Prefer distro packages by default in the Ubuntu-based development environment.
if [[ $FFMPEG_VERSION == "release" || $FFMPEG_VERSION == "system" ]]; then
  if [[ ! -x /usr/bin/ffmpeg || ! -x /usr/bin/ffprobe ]]; then
    echo "Installing FFmpeg from distro packages..."
    "${SUDO[@]}" apt-get update
    "${SUDO[@]}" apt-get -qq install ffmpeg
  else
    echo "FFmpeg distro package already installed."
  fi

  if [[ ! -x /usr/bin/ffmpeg || ! -x /usr/bin/ffprobe ]]; then
    echo "Error: FFmpeg distro binaries were not found in /usr/bin." 1>&2
    exit 1
  fi

  echo "Using FFmpeg from distro packages:"
  /usr/bin/ffmpeg -version | head -1
  update_symlinks "/usr/bin/ffmpeg" "/usr/bin/ffprobe"
  echo "Done."
  exit 0
fi

DESTDIR=$(realpath -m "${DESTDIR}")

# Determine target architecture for static builds.
if [[ -n ${PHOTOPRISM_ARCH:-} ]]; then
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

# Determine download URL for static builds.
# - "master" -> nightly build (ffmpeg-master-latest-*)
# - "latest" -> latest stable release (ffmpeg-nX.Y-latest-*)
case $FFMPEG_VERSION in
  master)
    ARCHIVE="ffmpeg-master-latest-${BTBN_ARCH}-gpl.tar.xz"
    URL="${BTBN_BASE}/${ARCHIVE}"
    ;;

  latest)
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
    echo "Use 'release', 'latest', or 'master'." 1>&2
    exit 1
    ;;
esac

echo "VERSION: $FFMPEG_VERSION"
echo "ARCHIVE: $ARCHIVE"
echo "DESTDIR: $DESTDIR"
echo "SOURCE:  BtbN/FFmpeg-Builds"
echo "Downloading from: $URL"

"${SUDO[@]}" mkdir -p "${DESTDIR}"

if ! curl -fsSL "$URL" | "${SUDO[@]}" tar --strip-components=1 --overwrite --mode=755 -x --xz -C "$DESTDIR"; then
  echo "Error: Failed to download or extract FFmpeg archive." 1>&2
  echo "Please check your network connection and try again." 1>&2
  exit 1
fi

"${SUDO[@]}" chown -R root:root "${DESTDIR}"

FFMPEG_BIN="${DESTDIR}/bin/ffmpeg"
FFPROBE_BIN="${DESTDIR}/bin/ffprobe"

if [[ ! -x "${FFMPEG_BIN}" || ! -x "${FFPROBE_BIN}" ]]; then
  echo "Error: Could not find ffmpeg/ffprobe binaries in ${DESTDIR}/bin" 1>&2
  exit 1
fi

# Force-update symbolic links for static builds.
update_symlinks "${FFMPEG_BIN}" "${FFPROBE_BIN}"

# Verify installation.
if [[ -x /usr/local/bin/ffmpeg ]]; then
  echo "FFmpeg installed successfully:"
  /usr/local/bin/ffmpeg -version | head -1
else
  echo "Warning: FFmpeg installation could not be verified." 1>&2
fi

echo "Done."
