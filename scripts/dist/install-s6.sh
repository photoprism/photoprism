#!/usr/bin/env bash

# This downloads and installs the s6-overlay binaries and noarch files from GitHub:
# - https://github.com/just-containers/s6-overlay
#
# s6 is a suite of utilities for UNIX that provides process supervision, e.g. for use with Docker:
# - https://github.com/skarnet/s6
# - https://skarnet.org/software/s6/
# - https://ahmet.im/blog/minimal-init-process-for-containers/
# - https://labex.io/tutorials/docker-how-to-gracefully-shut-down-a-long-running-docker-container-417742

# Show usage information if first argument is --help.
if [[ ${1} == "--help" ]]; then
  echo "Usage: ${0##*/} [version] [dir]" 1>&2
  exit 0
fi

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "${0##*/} [version] [dir] must be run as root" 1>&2
  exit 1
fi

# You can provide a custom installation directory as the first argument.
S6_OVERLAY_DESTDIR=$(realpath "${2:-/}")

# Determine target architecture.
if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

. /etc/os-release

S6_OVERLAY_ARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

case $S6_OVERLAY_ARCH in
  amd64 | AMD64 | x86_64 | x86-64)
    S6_OVERLAY_ARCH=x86_64
    ;;

  arm64 | ARM64 | aarch64)
    S6_OVERLAY_ARCH=aarch64
    ;;

  arm | ARM | aarch | armv7l | armhf)
    S6_OVERLAY_ARCH=armhf
    ;;

  *)
    echo "Unsupported Machine Architecture: \"$S6_OVERLAY_ARCH\"" 1>&2
    exit 1
    ;;
esac

set -eu

S6_OVERLAY_LATEST=$(curl --silent "https://api.github.com/repos/just-containers/s6-overlay/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
S6_OVERLAY_VERSION=${1:-$S6_OVERLAY_LATEST}
S6_ARCH_URL="https://github.com/just-containers/s6-overlay/releases/download/${S6_OVERLAY_VERSION}/s6-overlay-${S6_OVERLAY_ARCH}.tar.xz"
S6_NOARCH_URL="https://github.com/just-containers/s6-overlay/releases/download/${S6_OVERLAY_VERSION}/s6-overlay-noarch.tar.xz"

echo "Installing S6 Overlay..."

echo "------------------------------------------------"
echo "VERSION   : ${S6_OVERLAY_VERSION}"
echo "LATEST    : ${S6_OVERLAY_LATEST}"
echo "DESTDIR   : ${S6_OVERLAY_DESTDIR}"
echo "BINARY URL: ${S6_ARCH_URL}"
echo "NOARCH URL: ${S6_NOARCH_URL}"
echo "------------------------------------------------"

# Create the destination directory if it does not already exist.
mkdir -p "${S6_OVERLAY_DESTDIR}"

# Download and install the s6-overlay release from GitHub.
echo "Extracting \"$S6_ARCH_URL\" to \"$S6_OVERLAY_DESTDIR\"."
curl -fsSL "$S6_ARCH_URL" | tar -C "${S6_OVERLAY_DESTDIR}" -Jxp

echo "Extracting \"$S6_NOARCH_URL\" to \"$S6_OVERLAY_DESTDIR\"."
curl -fsSL "$S6_NOARCH_URL" | tar -C "${S6_OVERLAY_DESTDIR}" -Jxp

echo "Done."
