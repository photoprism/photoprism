#!/usr/bin/env bash

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# This script returns the normalized machine architecture (amd64, arm64, or arm).
# An error is returned if the architecture is currently not supported by PhotoPrism.

if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
elif [[ $OS == "Windows_NT" ]]; then
  if [[ $PROCESSOR_ARCHITEW6432 == "AMD64" || $PROCESSOR_ARCHITECTURE == "AMD64" ]]; then
    SYSTEM_ARCH=amd64
  else
    echo "Unsupported Windows Architecture" 1>&2
    exit 1
  fi
else
  SYSTEM_ARCH=$(uname -m)
fi

BUILD_ARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

#   Value    Normalized
#   aarch64  arm64      # the latest v8 arm architecture. Used on Apple M1, AWS Graviton, and Raspberry Pi 3's and 4's
#   armhf    arm        # 32-bit v7 architecture. Used in Raspberry Pi 3 and  Pi 4 when 32bit Raspbian Linux is used
#   armel    arm/v6     # 32-bit v6 architecture. Used in Raspberry Pi 1, 2, and Zero
#   i386     386        # older Intel 32-Bit architecture, originally used in the 386 processor
#   x86_64   amd64      # all modern Intel-compatible x84 64-Bit architectures
#   x86-64   amd64      # same

case $BUILD_ARCH in
  amd64 | AMD64 | x86_64 | x86-64)
    BUILD_ARCH=amd64
    ;;

  arm64 | ARM64 | aarch64)
    BUILD_ARCH=arm64
    ;;

  arm | ARM | aarch | armv7l | armhf)
    BUILD_ARCH=arm
    ;;

  *)
    echo "Unsupported Machine Architecture: \"$BUILD_ARCH\"" 1>&2
    exit 1
    ;;
esac

export BUILD_ARCH=$BUILD_ARCH

echo $BUILD_ARCH