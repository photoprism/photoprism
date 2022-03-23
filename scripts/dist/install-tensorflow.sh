#!/bin/bash

PATH="/usr/local/sbin/:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts"

set -e

TF_VERSION=${TF_VERSION:-1.15.2}

SYSTEM_ARCH=$("$(dirname "$0")/arch.sh")
DESTARCH=${DESTARCH:-$SYSTEM_ARCH}

if [[ $1 == "auto" ]]; then
  TF_DRIVER="auto";
  DESTDIR="/usr";
else
  DESTDIR=$(realpath "${1:-/usr}")
fi

TMPDIR=${TMPDIR:-/tmp}

# abort if not executed as root
if [[ $(id -u) != "0" ]] && [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Error: Run ${0##*/} as root to install in a system directory!" 1>&2
  exit 1
fi

/bin/mkdir -p "$DESTDIR"

if [[ $TF_DRIVER == "auto" ]]; then
  echo "Detecting driver..."
  TF_DRIVER=$("$(dirname "$0")/tensorflow-driver.sh")
fi

if [[ -z $TF_DRIVER ]]; then
  echo "Installing TensorFlow ${TF_VERSION} for ${DESTARCH^^} in \"$DESTDIR\"..."
  INSTALL_FILE="${DESTARCH}/libtensorflow-${DESTARCH}-${TF_VERSION}.tar.gz"
else
  echo "Installing TensorFlow ${TF_VERSION} for ${DESTARCH^^}-${TF_DRIVER^^} in \"$DESTDIR\"..."
  INSTALL_FILE="${DESTARCH}/libtensorflow-${DESTARCH}-${TF_DRIVER}-${TF_VERSION}.tar.gz"
fi

if [ ! -f "$TMPDIR/$INSTALL_FILE" ]; then
  URL="https://dl.photoprism.app/tensorflow/${INSTALL_FILE}"
  echo "Downloading ${DESTARCH} libs from \"$URL\". Please wait."
  curl --create-dirs -fsSL -o "$TMPDIR/$INSTALL_FILE" "$URL"
fi

echo "Extracting \"$TMPDIR/$INSTALL_FILE\" to \"$DESTDIR\"."

if [ -f "$TMPDIR/$INSTALL_FILE" ]; then
  tar --overwrite --mode=755 -C "$DESTDIR" -xzf "$TMPDIR/$INSTALL_FILE"
else
  echo "Fatal: \"$TMPDIR/$INSTALL_FILE\" not found"
  exit 1
fi

if [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Running \"ldconfig\"."
  ldconfig
else
  echo "Running \"ldconfig -n $DESTDIR/lib\"."
  ldconfig -n "$DESTDIR/lib"
fi

echo "Done."
