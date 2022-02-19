#!/usr/bin/env bash

# This script returns the normalized machine architecture (amd64, arm64, or arm).
# An error is returned if the architecture is currently not supported by PhotoPrism.

if [[ $TARGETARCH ]]; then
  SYSTEM_ARCH=$TARGETARCH
elif [[ $OS == "Windows_NT" ]]; then
  if [[ $PROCESSOR_ARCHITEW6432 == "AMD64" || $PROCESSOR_ARCHITECTURE == "AMD64" ]]; then
    SYSTEM_ARCH=amd64
  else
    echo "Unsupported Windows Architecture" 1>&2
    exit 1
  fi
else
  SYSTEM_ARCH=$(arch)
fi

BUILD_ARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

case $BUILD_ARCH in
  amd64 | AMD64 | x86_64)
    BUILD_ARCH=amd64
    ;;

  arm64 | ARM64 | aarch64)
    BUILD_ARCH=arm64
    ;;

  arm | ARM | aarch | armv7l | armhf | arm6l)
    BUILD_ARCH=arm
    ;;

  *)
    echo "Unsupported Machine Architecture: \"$BUILD_ARCH\"" 1>&2
    exit 1
    ;;
esac

export BUILD_ARCH=$BUILD_ARCH

echo $BUILD_ARCH