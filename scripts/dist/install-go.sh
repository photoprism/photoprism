#!/usr/bin/env bash

# Downloads and installs the Go compiler on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-go.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Show usage information?
if [[ ${1} == "--help" ]]; then
  echo "${0##*/} [version] [destdir] downloads and installs the Go compiler on Linux, for example:" 1>&2
  echo "${0##*/}" 1>&2
  echo "${0##*/} latest" 1>&2
  echo "${0##*/} 1.23.4 /usr/local" 1>&2
  exit 0
fi

set -e
set +x

# Check version to be installed:
GOLANG_VERSION=${1:-$GOLANG_VERSION}

if [[ -z $GOLANG_VERSION ]] || [[ $GOLANG_VERSION == "latest" ]]; then
  GOLANG_VERSION=$(curl -fsSL https://go.dev/VERSION?m=text | head -n 1)
fi

GOLANG_VERSION=${GOLANG_VERSION#"go"}

if [[ -z $GOLANG_VERSION ]]; then
  echo "Go compiler version must be passed as first argument, e.g. 1.23.4" 1>&2
  exit 1
fi

# Check destination directory:
DESTDIR=$(realpath "${2:-/usr/local}")

if [[ -z $DESTDIR ]] || [[ $DESTDIR == "default" ]]; then
  DESTDIR="/usr/local"
fi

# Determine the system architecture:
if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

DESTARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

# Start installation:
echo "Installing Go ${GOLANG_VERSION} for ${DESTARCH^^} in \"${DESTDIR}\". Please wait."

sudo mkdir -p "$DESTDIR"

case $DESTARCH in
  amd64 | AMD64 | x86_64 | x86-64)
    URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz"
    ;;

  arm64 | ARM64 | aarch64)
    URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-arm64.tar.gz"
    ;;

  arm | ARM | aarch | armv7l | armhf)
    URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-armv6l.tar.gz"
    ;;

  *)
    echo "Unsupported Machine Architecture: \"$DESTARCH\"" 1>&2
    exit 1
    ;;
esac

# Replace current installation in "$DESTDIR/go":
echo "Extracting \"${URL}\" to \"${DESTDIR}/go\"..."
set -eux;
sudo rm -rf "${DESTDIR}/go"
curl -fsSL "${URL}" | sudo tar -xz -C "${DESTDIR}"
set +x
echo "Done."

# Add symlink to go binary:
echo "Adding symbolic links for \"${DESTDIR}/go/bin/go\" and \"${DESTDIR}/go/bin/gofmt\"..."
set -eux;
sudo ln -sf "${DESTDIR}/go/bin/go" /usr/local/bin/go
sudo ln -sf "${DESTDIR}/go/bin/gofmt" /usr/local/bin/gofmt
set +x
echo "Done."

# Telemetry in Go >= 1.23 should be set to "off" in ~/.config/go/telemetry, see https://go.dev/doc/telemetry.
# You can do this by running "go telemetry off":
echo "Disabling Go telemetry..."
set +e -x;
mkdir -p ~/.config/go
go telemetry off
set +x

# Test go command by showing installed Go version:
echo "Installed Go version:"
go version

echo "Enjoy!"