#!/usr/bin/env bash

# Installs GPU drivers on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-gpu.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Error: Run ${0##*/} as root" 1>&2
  exit 1
fi

set -e

# Determine target architecture.
if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

DESTARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

TMPDIR=${TMPDIR:-/tmp}

# shellcheck source=/dev/null
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

# shellcheck disable=SC2068
for t in ${GPU_DETECTED[@]}; do
  case $t in
    i915 | i965 | intel | opencl | icd)
      echo "Installing Intel GPU Drivers..."
      # VA-API/QSV (h264_vaapi, h264_qsv) plus the Mesa ANV Vulkan driver and tools (h264_vulkan).
      apt-get -qq install intel-opencl-icd intel-media-va-driver-non-free i965-va-driver-shaders mesa-va-drivers libmfx-gen1.2 va-driver-all vainfo libva2 libvpl2 mesa-vulkan-drivers vulkan-tools
      ;;

    nvidia)
      # The NVIDIA driver and its Vulkan ICD are provided by the NVIDIA Container Toolkit, not apt;
      # we only add the Vulkan loader and tools so h264_vulkan can run and be verified with vulkaninfo.
      echo "Installing NVIDIA Vulkan loader and tools..."
      apt-get -qq install libvulkan1 vulkan-tools
      echo "NVIDIA Container Toolkit must be installed on the host: https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html"
      echo "For Vulkan transcoding (h264_vulkan), the container's NVIDIA_DRIVER_CAPABILITIES must include \"graphics\" (or \"all\") in addition to \"video\"/\"compute\"."
      ;;

    amdgpu)
      echo "Installing AMD VA-API and Vulkan GPU Drivers..."
      # VA-API (h264_vaapi) plus the Mesa RADV Vulkan driver and tools (h264_vulkan).
      apt-get -qq install mesa-va-drivers mesa-vulkan-drivers vainfo vulkan-tools libva2
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
