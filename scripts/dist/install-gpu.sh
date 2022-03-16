#!/usr/bin/env bash

# abort if not executed as root
if [[ $(id -u) != "0" ]]; then
  echo "Error: Run ${0##*/} as root" 1>&2
  exit 1
fi

set -e

SYSTEM_ARCH=$("$(/usr/bin/dirname "$0")/arch.sh")
DESTARCH=${DESTARCH:-$SYSTEM_ARCH}
TMPDIR=${TMPDIR:-/tmp}
. /etc/os-release

if [[ $DESTARCH != "amd64" ]]; then
  echo "Installing GPU drivers for ${DESTARCH} is not supported yet."
  exit
fi

/usr/bin/apt-get update
/usr/bin/apt-get -qq upgrade
/usr/bin/apt-get -qq install lshw jq

# shellcheck disable=SC2207
GPU_DETECTED=($(/usr/bin/lshw -c display -json 2>/dev/null | /usr/bin/jq -r '.[].configuration.driver'))

# shellcheck disable=SC2068
for t in ${GPU_DETECTED[@]}; do
  case $t in
    i915)
      /usr/bin/apt-get -qq install intel-opencl-icd intel-media-va-driver-non-free i965-va-driver-shaders libmfx1 libva2 vainfo libva-wayland2
      ;;

    nvidia)
      /usr/bin/apt-get -qq install nvidia-opencl-icd nvidia-vdpau-driver nvidia-driver-libs nvidia-kernel-dkms libva2 vainfo libva-wayland2
      ;;

    *)
      echo "Unsupported GPU: \"$t\"";
      ;;
  esac
done

echo "Done."
