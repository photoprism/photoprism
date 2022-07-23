#!/usr/bin/env bash

# Installs GPU drivers on Linux
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-gpu.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# abort if not executed as root
if [[ $(id -u) != "0" ]]; then
  echo "Error: Run ${0##*/} as root" 1>&2
  exit 1
fi

set -e

if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

DESTARCH=${DESTARCH:-$SYSTEM_ARCH}
TMPDIR=${TMPDIR:-/tmp}

. /etc/os-release

apt-get update
apt-get -qq upgrade
apt-get -qq install lshw jq

# shellcheck disable=SC2207
GPU_DETECTED=($(lshw -c display -json 2>/dev/null | jq -r '.[].configuration.driver'))

echo "GPU detected: ${GPU_DETECTED[*]}"

case $DESTARCH in
  amd64 | AMD64 | x86_64 | x86-64)
    echo "Installing GPU drivers for ${DESTARCH^^}..."
    ;;

  *)
    echo "Installing GPU drivers for ${DESTARCH^^} not supported at this time."
    exit 0
    ;;
esac

# TODO: Install NVIDIA Drivers from https://developer.download.nvidia.com/compute/cuda/repos/
# curl -fsSL https://developer.download.nvidia.com/compute/cuda/repos/{DIST}/x86_64/7fa2af80.pub | gpg --dearmor -o /etc/apt/trusted.gpg.d/developer.download.nvidia.com.gpg
# add-apt-repository "deb https://developer.download.nvidia.com/compute/cuda/repos/{DIST}/x86_64/ /"
# curl -fsSL https://developer.download.nvidia.com/compute/cuda/repos/{DIST}/x86_64/cuda-{DIST}.pin > /etc/apt/preferences.d/cuda-repository-pin-600
# apt-get update
# apt-get install libglvnd-dev pkg-config dkms build-essential cuda nvidia-driver-510 nvidia-settings nvidia-utils-510 linux-headers-$(uname -r)

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
