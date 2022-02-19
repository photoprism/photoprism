#!/usr/bin/env bash

set -e

TF_VERSION="1.15.2"

DESTDIR=$(realpath "${1:-/usr}")

echo "Installing TensorFlow in \"$DESTDIR\"..."

# abort if the user is not root
if [[ $(id -u) != "0" ]] && [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Error: Run install-tensorflow.sh as root to install in a system directory!" 1>&2
  exit 1
fi

mkdir -p "$DESTDIR"

SYSTEM_ARCH=$("$(dirname "$0")/arch.sh")
INSTALL_ARCH=${2:-$SYSTEM_ARCH}
TMPDIR=${TMPDIR:-/tmp}

if [[ -z $3 ]]; then
  INSTALL_FILE="${INSTALL_ARCH}/libtensorflow-${INSTALL_ARCH}-${TF_VERSION}.tar.gz"
else
  INSTALL_FILE="${INSTALL_ARCH}/libtensorflow-${INSTALL_ARCH}-${2}-${TF_VERSION}.tar.gz"
fi

if [ ! -f "$TMPDIR/$INSTALL_FILE" ]; then
  URL="https://dl.photoprism.app/tensorflow/${INSTALL_FILE}"
  echo "Downloading $INSTALL_ARCH libs from \"$URL\". Please wait."
  curl --create-dirs -fsSL -o "$TMPDIR/$INSTALL_FILE" "$URL"
fi

echo "Extracting \"$TMPDIR/$INSTALL_FILE\" to \"$DESTDIR\"..."

if [ -f "$TMPDIR/$INSTALL_FILE" ]; then
  tar --overwrite --mode=755 -C "$DESTDIR" -xzf "$TMPDIR/$INSTALL_FILE"
else
  echo "Fatal: \"$TMPDI/$INSTALL_FILE\" not found"
  exit 1
fi

if [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Running ldconfig..."
  ldconfig
else
  echo "Running \"ldconfig -n $DESTDIR/lib\"."
  ldconfig -n "$DESTDIR/lib"
fi

echo "Done."
