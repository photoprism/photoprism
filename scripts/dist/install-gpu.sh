#!/bin/bash

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# abort if not executed as root
if [[ $(id -u) != "0" ]]; then
  echo "Error: Run ${0##*/} as root" 1>&2
  exit 1
fi

set -e

SYSTEM_ARCH=$("$(dirname "$0")/arch.sh")
DESTARCH=${DESTARCH:-$SYSTEM_ARCH}
TMPDIR=${TMPDIR:-/tmp}
. /etc/os-release

if [[ $DESTARCH != "amd64" ]]; then
  echo "Installing GPU drivers for ${DESTARCH} is not supported yet."
  exit
fi

apt-get update
apt-get -qq upgrade
apt-get -qq install lshw jq

# shellcheck disable=SC2207
GPU_DETECTED=($(lshw -c display -json 2>/dev/null | jq -r '.[].configuration.driver'))

echo "GPU detected: ${GPU_DETECTED[*]}"

# shellcheck disable=SC2068
for t in ${GPU_DETECTED[@]}; do
  case $t in
    i915)
      echo "Installing Intel Drivers..."
      apt-get -qq install intel-opencl-icd intel-media-va-driver-non-free i965-va-driver-shaders mesa-va-drivers libmfx1 libva2 vainfo libva-wayland2
      ;;

    nvidia)
      echo "Installing Nvidia Drivers..."
      apt-get -qq install libcuda1 libnvcuvid1 libnvidia-encode1 nvidia-opencl-icd nvidia-vdpau-driver nvidia-driver-libs nvidia-kernel-dkms libva2 vainfo libva-wayland2
      ;;

    "null")
      # ignore
      ;;

    *)
      echo "Unsupported GPU: \"$t\"";
      ;;
  esac
done

echo "Done."
