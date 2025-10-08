#!/usr/bin/env bash

# Installs the TensorFlow libraries on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-tensorflow.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

set -e

TF_VERSION=${TF_VERSION:-2.18.0}

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

if [[ $1 == "auto" ]]; then
  TF_DRIVER="auto";
  DESTDIR="/usr";
elif [[ $1 == "gpu" ]]; then
  TF_DRIVER="gpu";
  DESTDIR="/usr";
else
  DESTDIR=$(realpath "${1:-/usr}")
fi

TMPDIR=${TMPDIR:-/tmp}

# Abort if not executed as root.
if [[ $(id -u) != "0" ]] && [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Error: Run ${0##*/} as root to install in a system directory!" 1>&2
  exit 1
fi

mkdir -p "$DESTDIR"

if [[ $DESTARCH != "amd64" ]]; then
  TF_DRIVER=""
fi

if [[ $TF_DRIVER == "auto" ]]; then
  echo "Detecting driver..."

  CPU_DETECTED=$(lshw -c processor -json 2>/dev/null)

  if [[ $(echo "${CPU_DETECTED}" | jq -r '.[0].capabilities.avx512_vnni') == "true" ]]; then
    TF_DRIVER="avx512"
    echo "✅ Your CPU supports AVX512-VNNI Vector Neural Network Instructions."
    if [[ $(lshw -c display -json 2>/dev/null | jq -r '.[0].configuration.driver') == "nvidia" ]]; then
      echo "✅ You can optionally install the tensorflow-gpu version as an Nvidia GPU was detected."
    fi
  elif [[ $(echo "${CPU_DETECTED}" | jq -r '.[0].capabilities.avx_vnni') == "true" ]]; then
    TF_DRIVER="vnni"
    echo "✅ Your CPU supports AVX-VNNI Vector Neural Network Instructions."
    if [[ $(lshw -c display -json 2>/dev/null | jq -r '.[0].configuration.driver') == "nvidia" ]]; then
      echo "✅ You can optionally install the tensorflow-gpu version as an Nvidia GPU was detected."
    fi
  elif [[ $(echo "${CPU_DETECTED}" | jq -r '.[0].capabilities.avx2') == "true" ]]; then
    TF_DRIVER="avx2"
    echo "✅ Your CPU supports AVX2 vector instructions."
  elif [[ $(echo "${CPU_DETECTED}" | jq -r '.[0].capabilities.avx') == "true" ]]; then
    TF_DRIVER="avx"
    echo "✅ Your CPU supports AVX vector instructions."
  elif [[ $(echo "${CPU_DETECTED}" | jq -r '.[0].capabilities.ssse3') == "true" ]]; then
    TF_DRIVER="ssse3"
    echo "✅ Your CPU supports SSSE3 instructions."
  else
    TF_DRIVER=""
    echo "❌ Your CPU does not support any AVX instructions."
  fi
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
  echo "Running \"ldconfig\"..."
  ldconfig
else
  echo "Running \"ldconfig -n $DESTDIR/lib\"..."
  ldconfig -n "$DESTDIR/lib"
fi

echo "Done."
