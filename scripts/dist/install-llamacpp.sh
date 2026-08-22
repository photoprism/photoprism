#!/usr/bin/env bash

# Installs llama.cpp prebuilt Linux binaries from GitHub on AMD64 and ARM64.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-llamacpp.sh)
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-llamacpp.sh) -- --vulkan /usr/local b9948
#
# The accelerator is auto-detected by default (the Vulkan build for any supported GPU,
# otherwise a plain CPU build). Vulkan covers NVIDIA, AMD, and Intel GPUs on Linux;
# upstream ships no CUDA build, and the ROCm build is intentionally not offered here
# because it is complex to provision and Vulkan matches or beats it in practice.

set -euo pipefail

if ! command -v jq >/dev/null 2>&1; then
  echo "Error: jq is required but not installed." 1>&2
  exit 1
fi

REPO="ggml-org/llama.cpp"

# usage prints the command-line synopsis and supported environment variables.
usage() {
  echo "Usage: ${0##*/} [--auto|--cpu|--vulkan|--sycl] [destdir] [version]" 1>&2
  echo "       ${0##*/} [--accel auto|cpu|vulkan|sycl] [destdir] [version]" 1>&2
  echo "" 1>&2
  echo "Arguments:" 1>&2
  echo "  destdir   Install prefix (default: /usr/local); binaries land in <destdir>/bin." 1>&2
  echo "  version   Release tag to install (e.g. b9948); defaults to the latest release." 1>&2
  echo "" 1>&2
  echo "Environment:" 1>&2
  echo "  PHOTOPRISM_LLAMA_ACCEL=auto|cpu|vulkan|sycl   Accelerator to install." 1>&2
  echo "  PHOTOPRISM_ARCH / BUILD_ARCH                  Override target architecture." 1>&2
  echo "  LLAMA_SHA256                                  Verify the download against this checksum." 1>&2
}

# ACCEL selects the hardware backend; "auto" resolves it from the detected GPU below.
ACCEL=${PHOTOPRISM_LLAMA_ACCEL:-auto}

if [[ ${1:-} == "--help" || ${1:-} == "-h" ]]; then
  usage
  exit 0
fi

while [[ $# -gt 0 ]]; do
  case $1 in
    --auto)   ACCEL=auto;   shift ;;
    --cpu)    ACCEL=cpu;    shift ;;
    --vulkan) ACCEL=vulkan; shift ;;
    --sycl)   ACCEL=sycl;   shift ;;
    --accel)
      ACCEL=${2:-}
      if [[ -z $ACCEL ]]; then
        echo "Error: --accel requires a value (auto, cpu, vulkan, or sycl)." 1>&2
        exit 1
      fi
      shift 2
      ;;
    --accel=*) ACCEL=${1#--accel=}; shift ;;
    -h | --help) usage; exit 0 ;;
    --) shift; break ;;
    -*) echo "Error: Unknown option: $1" 1>&2; exit 1 ;;
    *) break ;;
  esac
done

# Normalize the accelerator selection to a known keyword.
ACCEL=${ACCEL,,}
case $ACCEL in
  auto | cpu | vulkan | sycl) ;;
  gpu) ACCEL=auto ;;
  *) echo "Error: Unsupported accelerator \"$ACCEL\" (use auto, cpu, vulkan, or sycl)." 1>&2; exit 1 ;;
esac

# You can provide a custom installation directory as the first positional argument.
DESTDIR=$(realpath -m "${1:-/usr/local}")

# Determine target architecture.
if [[ -n ${PHOTOPRISM_ARCH:-} ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

DESTARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

# llama.cpp names its Linux assets "...-ubuntu-<accel>-x64" / "...-arm64".
case $DESTARCH in
  amd64 | AMD64 | x86_64 | x86-64) DESTARCH=x64 ;;
  arm64 | ARM64 | aarch64) DESTARCH=arm64 ;;
  *)
    echo "Error: Unsupported Machine Architecture: \"$DESTARCH\"" 1>&2
    exit 1
    ;;
esac

# Abort if not executed as root when installing into a system directory.
if [[ $(id -u) != "0" ]] && [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Error: Run ${0##*/} as root to install in a system directory!" 1>&2
  exit 1
fi

# detect_gpu_vendors populates GPU_VENDORS with the GPU makers found on this host.
# It reads PCI vendor IDs from sysfs first (no dependencies) and falls back to lspci.
GPU_VENDORS=""
detect_gpu_vendors() {
  local f id out
  for f in /sys/class/drm/card[0-9]*/device/vendor; do
    [[ -r $f ]] || continue
    id=$(cat "$f" 2>/dev/null || true)
    case $id in
      0x10de) GPU_VENDORS+=" nvidia" ;;
      0x1002) GPU_VENDORS+=" amd" ;;
      0x8086) GPU_VENDORS+=" intel" ;;
    esac
  done
  if [[ -z ${GPU_VENDORS// /} ]] && command -v lspci >/dev/null 2>&1; then
    out=$(lspci 2>/dev/null | grep -iE 'vga|3d controller|display' || true)
    grep -qi 'nvidia' <<<"$out" && GPU_VENDORS+=" nvidia"
    grep -qiE 'amd|advanced micro|ati' <<<"$out" && GPU_VENDORS+=" amd"
    grep -qi 'intel' <<<"$out" && GPU_VENDORS+=" intel"
  fi
  command -v nvidia-smi >/dev/null 2>&1 && GPU_VENDORS+=" nvidia"
  # Deduplicate and trim.
  GPU_VENDORS=$(tr ' ' '\n' <<<"$GPU_VENDORS" | sort -u | tr '\n' ' ' | sed 's/^ *//;s/ *$//')
}

# has_gpu returns success if the given GPU vendor was detected.
has_gpu() { [[ " $GPU_VENDORS " == *" $1 "* ]]; }

# vulkan_runtime_available returns success if the Vulkan loader is present on the host.
vulkan_runtime_available() {
  command -v vulkaninfo >/dev/null 2>&1 && return 0
  ldconfig -p 2>/dev/null | grep -q 'libvulkan\.so\.1' && return 0
  return 1
}

# Resolve "auto" to a concrete accelerator based on the detected GPU: the Vulkan build
# for any supported GPU (NVIDIA, AMD, or Intel), otherwise a plain CPU build.
if [[ $ACCEL == auto ]]; then
  detect_gpu_vendors
  echo "GPU detected: ${GPU_VENDORS:-none}"
  if [[ -n $GPU_VENDORS ]] && { has_gpu nvidia || has_gpu amd || has_gpu intel; }; then
    ACCEL=vulkan
  else
    ACCEL=cpu
  fi
  echo "Selected accelerator: ${ACCEL} (override with --cpu/--vulkan/--sycl)."
fi

echo "Installing llama.cpp (${ACCEL}) for ${DESTARCH^^}..."

# Fetch release metadata: a specific tag when requested, otherwise the latest release.
if [[ -n ${2:-} ]]; then
  VERSION=${2}
  RELEASE_JSON=$(curl --fail --silent --show-error "https://api.github.com/repos/${REPO}/releases/tags/${VERSION}" || true)
  if [[ -z $RELEASE_JSON ]]; then
    echo "Error: Unable to fetch release metadata for tag ${VERSION}." 1>&2
    exit 1
  fi
else
  RELEASE_JSON=$(curl --fail --silent --show-error "https://api.github.com/repos/${REPO}/releases/latest" || true)
  if [[ -z $RELEASE_JSON ]]; then
    echo "Error: Unable to fetch release metadata from GitHub." 1>&2
    exit 1
  fi
fi

TAG_NAME=$(jq -r '.tag_name // empty' <<<"$RELEASE_JSON")
if [[ -z $TAG_NAME ]]; then
  echo "Error: Release metadata did not include a tag name." 1>&2
  exit 1
fi

# asset_pattern returns the regex matching the ubuntu asset for the given accelerator.
asset_pattern() {
  case $1 in
    cpu)    echo "bin-ubuntu-${DESTARCH}\\.tar\\.gz$" ;;
    vulkan) echo "bin-ubuntu-vulkan-${DESTARCH}\\.tar\\.gz$" ;;
    sycl)   echo "bin-ubuntu-sycl-fp16-${DESTARCH}\\.tar\\.gz$" ;;
  esac
}

# asset_url prints the download URL of the first asset matching the given accelerator.
asset_url() {
  local re
  re=$(asset_pattern "$1")
  [[ -z $re ]] && return 0
  jq -r --arg re "$re" '.assets[]? | select(.name | test($re)) | .browser_download_url' <<<"$RELEASE_JSON" | head -n1
}

# Build a fallback chain so an unavailable accelerator degrades to the plain CPU build
# (e.g. Vulkan or SYCL may be missing for a given arch in some releases).
case $ACCEL in
  vulkan) CANDIDATES=(vulkan cpu) ;;
  sycl)   CANDIDATES=(sycl cpu) ;;
  cpu)    CANDIDATES=(cpu) ;;
esac

RESOLVED_ACCEL=""
ASSET_URL=""
for candidate in "${CANDIDATES[@]}"; do
  url=$(asset_url "$candidate")
  if [[ -n $url && $url != "null" ]]; then
    RESOLVED_ACCEL=$candidate
    ASSET_URL=$url
    break
  fi
done

if [[ -z $ASSET_URL ]]; then
  echo "Error: No llama.cpp Linux asset found for ${ACCEL}/${DESTARCH} in release ${TAG_NAME}." 1>&2
  exit 1
fi

if [[ $RESOLVED_ACCEL != "$ACCEL" ]]; then
  echo "Warning: No ${ACCEL} build for ${DESTARCH} in ${TAG_NAME}; falling back to ${RESOLVED_ACCEL}." 1>&2
fi

ASSET_NAME=${ASSET_URL##*/}
LLAMA_DIR="${DESTDIR}/lib/llama.cpp"

echo "--------------------------------------------------------------------------------"
echo "VERSION : ${TAG_NAME}"
echo "ACCEL   : ${RESOLVED_ACCEL}"
echo "ARCH    : ${DESTARCH}"
echo "ASSET   : ${ASSET_NAME}"
echo "DOWNLOAD: ${ASSET_URL}"
echo "DESTDIR : ${DESTDIR}"
echo "LIBDIR  : ${LLAMA_DIR}"
echo "--------------------------------------------------------------------------------"

tmp_tar=$(mktemp)
workdir=$(mktemp -d)
trap 'rm -rf "${tmp_tar}" "${workdir}"' EXIT

echo "Downloading ${ASSET_NAME}..."
curl --fail --silent --show-error --location --retry 3 --retry-delay 2 --retry-all-errors "${ASSET_URL}" -o "${tmp_tar}"

# Verify the download when a checksum was provided; upstream ships no checksum manifest.
if [[ -n ${LLAMA_SHA256:-} ]]; then
  if command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "${tmp_tar}" | awk '{print $1}')
  else
    actual=$(shasum -a 256 "${tmp_tar}" | awk '{print $1}')
  fi
  if [[ ${LLAMA_SHA256} != "${actual}" ]]; then
    echo "Error: SHA-256 mismatch for ${ASSET_NAME}." 1>&2
    echo "  expected: ${LLAMA_SHA256}" 1>&2
    echo "  actual:   ${actual}" 1>&2
    exit 1
  fi
  echo "Checksum OK (${actual})."
fi

echo "Extracting ${ASSET_NAME}..."
tar xzf "${tmp_tar}" -C "${workdir}"

# The archive extracts to a single top-level "llama-<tag>" directory.
srcdir=$(find "${workdir}" -mindepth 1 -maxdepth 1 -type d | head -n1)
if [[ -z ${srcdir} ]]; then
  echo "Error: Unexpected archive layout in ${ASSET_NAME}." 1>&2
  exit 1
fi

# Replace any previous install so stale libraries don't linger next to the binaries.
echo "Installing binaries and libraries to ${LLAMA_DIR}..."
rm -rf "${LLAMA_DIR}"
mkdir -p "${LLAMA_DIR}"
cp -a "${srcdir}/." "${LLAMA_DIR}/"

# Symlink the executables into <destdir>/bin. The binaries carry RUNPATH "$ORIGIN",
# so they resolve their co-located shared libraries through the symlink.
mkdir -p "${DESTDIR}/bin"
linked=0
for f in "${LLAMA_DIR}"/llama* "${LLAMA_DIR}"/ggml-rpc-server; do
  [[ -f $f ]] || continue
  case $f in *.so | *.so.*) continue ;; esac
  [[ -x $f ]] || continue
  ln -sf "$f" "${DESTDIR}/bin/$(basename "$f")"
  linked=$((linked + 1))
done

if [[ $linked -eq 0 ]]; then
  echo "Error: No llama.cpp executables found in ${LLAMA_DIR}." 1>&2
  exit 1
fi

echo "Linked ${linked} executables into ${DESTDIR}/bin."

# Warn when the selected backend needs a runtime that is not present on this host.
case $RESOLVED_ACCEL in
  vulkan)
    if ! vulkan_runtime_available; then
      echo "Note: The Vulkan build needs the Vulkan loader (libvulkan.so.1) and a GPU driver ICD at runtime." 1>&2
      echo "      On Debian/Ubuntu: apt-get install libvulkan1 mesa-vulkan-drivers (plus the NVIDIA driver for NVIDIA GPUs)." 1>&2
    fi
    ;;
esac

echo "Try: ${DESTDIR}/bin/llama-cli --version"
echo "Done."
