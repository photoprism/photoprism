#!/usr/bin/env bash

set -euo pipefail

# Installs the NVIDIA CUDA runtime libraries that the ONNX Runtime CUDA Execution Provider needs.
#
# The provider links cudart, cuBLAS, cuBLASLt and cuRAND, and loads cuDNN on demand. None of them
# are part of the ONNX Runtime archive, and the NVIDIA Container Toolkit injects the host driver
# only, so they have to be installed separately. Pair this with "install-onnx.sh --gpu", which
# installs the runtime build that carries libonnxruntime_providers_cuda.so.
#
# The packages are the vendor's own, downloaded from our mirror with the NVIDIA repository as a
# fallback, and verified against the checksums pinned below. Selecting another CUDA release means
# updating that table, because a version is a set of packages rather than a single number: their
# point versions move independently, so nothing can derive a file name or a checksum from it.
#
# One pinned set serves every base image we ship. The packages are built for one distribution
# release, but they are extracted rather than installed, and the vendor builds them against an
# old C library, so what decides is the glibc version rather than the distribution name. The
# script tests for that below instead of mapping a release to a repository.
#
# Requires the proprietary NVIDIA driver on the host, and for Docker the NVIDIA Container Toolkit.
# CUDA 13 needs driver 580 or later; an older one loads the libraries and then reports no usable
# device, which PhotoPrism treats as "no GPU" and falls back to the CPU.

TODAY=$(date -u +%Y%m%d)
TMPDIR=${TMPDIR:-/tmp}
SYSTEM=$(uname -s)
ARCH=${PHOTOPRISM_ARCH:-$(uname -m)}

# The highest glibc version the pinned libraries below require. Re-derive it whenever they
# change, rather than carrying it forward:
#
#   objdump -T <library> | grep -oE 'GLIBC_[0-9.]+' | sort -V | tail -1
#
# It currently sits below what the ONNX Runtime itself needs (2.28), so every image able to run
# the CPU runtime can load these too.
CUDA_GLIBC_MIN="2.27"

# Every library the pinned packages are expected to provide, named individually so that a set
# which is merely short is refused here rather than at the first inference. Re-derive when the
# packages change:
#
#   dpkg-deb -c <package> | grep -oE 'lib[a-z_]+\.so' | sort -u
CUDA_LIBRARIES=(
  "libcudart" "libcublas" "libcublasLt" "libcurand"
  "libcudnn" "libcudnn_adv" "libcudnn_cnn" "libcudnn_engines_precompiled"
  "libcudnn_engines_runtime_compiled" "libcudnn_engines_tensor_ir" "libcudnn_ext"
  "libcudnn_graph" "libcudnn_heuristic" "libcudnn_ops"
)

# Pinned packages: file name and SHA256, as published for the distribution release below.
CUDA_REPO="ubuntu2604"
CUDA_PACKAGES="\
cuda-cudart-13-3_13.3.29-1_amd64.deb 3bd63bf657aa5ceb5514d2e939d27a2b895f61cbfde5ba747346ccc1cbc02110
libcublas-13-3_13.6.0.2-1_amd64.deb 55866198b96d68b170073d6cfa102552813d333d7ae058ba0753df7624134738
libcurand-13-3_10.4.3.29-1_amd64.deb 90ca2bb670f6ffd05046aae0c367a237f654d568bf87b66d5e388b2448dd54e7
libcudnn9-cuda-13_9.26.0.51-1_amd64.deb ea37ae57cce911bfd1009f22758be96bce62c909975c6c8b321317039a3c2b6d"

DESTDIR_ARG=${1:-/usr}

if [[ ! -d "${DESTDIR_ARG}" ]]; then
  mkdir -p "${DESTDIR_ARG}"
fi

DESTDIR=$(realpath "${DESTDIR_ARG}")

# Test the destination rather than compare it against a list of system directories, so that an
# install into a root-owned prefix such as the package layout reports the reason it cannot write
# instead of failing later in the extract.
if [[ $(id -u) != 0 ]] && [[ ! -w "${DESTDIR}" ]]; then
  echo "Error: Run ${0##*/} as root to install in '${DESTDIR}'." >&2
  exit 1
fi

if [[ "${SYSTEM}" != "Linux" ]]; then
  echo "Warning: CUDA is supported on Linux only; skipping the install on ${SYSTEM}." >&2
  exit 0
fi

# Skip rather than fail, because this target is usually paired with the runtime install and
# make stops at the first failed prerequisite: an error here would leave a host with no ONNX
# Runtime at all, where it should simply have the CPU one.
case "${ARCH}" in
  amd64|AMD64|x86_64|x86-64) ;;
  *)
    echo "Warning: ONNX Runtime GPU builds exist for Linux x64 only; skipping the CUDA install on ${ARCH}." >&2
    exit 0
    ;;
esac

if ! command -v dpkg-deb >/dev/null 2>&1; then
  echo "Error: dpkg-deb is required to extract the CUDA packages." >&2
  exit 1
fi

# version_lt returns success if $1 is a strictly lower semantic version than $2.
version_lt() {
  [[ "$1" != "$2" ]] && [[ "$(printf '%s\n%s\n' "$1" "$2" | sort -V | head -n1)" == "$1" ]]
}

# Refuse a C library too old to load the libraries, rather than installing 1.5 GB that no
# process can open. An unreadable version is refused for the same reason: it is not glibc.
libc_version=$(getconf GNU_LIBC_VERSION 2>/dev/null | awk '{print $NF}')

if [[ ! ${libc_version} =~ ^[0-9]+\.[0-9]+ ]]; then
  libc_version=$(ldd --version 2>/dev/null | head -1 | awk '{print $NF}')
fi

if [[ ! ${libc_version} =~ ^[0-9]+\.[0-9]+ ]]; then
  echo "Error: could not determine the glibc version; the CUDA libraries need ${CUDA_GLIBC_MIN} or later." >&2
  exit 1
fi

if version_lt "${libc_version}" "${CUDA_GLIBC_MIN}"; then
  echo "Error: the CUDA libraries need glibc ${CUDA_GLIBC_MIN} or later, and this system has ${libc_version}." >&2
  exit 1
fi

# Report a missing device rather than refuse: the container toolkit may attach one later, and
# an image build legitimately installs these libraries with no GPU present.
if [[ ! -e /dev/nvidiactl ]] && ! command -v nvidia-smi >/dev/null 2>&1; then
  echo "Warning: no NVIDIA device is visible here, so these libraries will install but stay unused."
fi

mkdir -p "${DESTDIR}" "${TMPDIR}"

verify_sha() {
  local expected="$1"
  local file="$2"
  if command -v sha256sum >/dev/null 2>&1; then
    echo "${expected}  ${file}" | sha256sum -c - >/dev/null
  else
    echo "${expected}  ${file}" | shasum -a 256 -c - >/dev/null
  fi
}

# Everything is downloaded and extracted inside one private directory that mktemp creates with
# a random name, owned by this process and readable by no one else. Nothing is written to a
# predictable path in a shared temporary directory, where another account could preempt the name
# or swap the file between its verification and its use. The cost is that a repeated run fetches
# the packages again rather than reusing them.
work_dir=$(mktemp -d "${TMPDIR}/cuda-XXXXXX")
chmod 0700 "${work_dir}"

extract_dir="${work_dir}/extracted"
mkdir -p "${extract_dir}"

# Remove the whole working directory on every way out, including a failure: it holds the
# downloaded packages as well as the extracted tree, and neither is wanted afterwards.
cleanup() {
  rm -rf "${work_dir}"
}

trap cleanup EXIT

while read -r package sha; do
  [[ -z "${package}" ]] && continue

  package_path="${work_dir}/${package}"
  primary_url="https://dl.photoprism.app/cuda/${CUDA_REPO}/${package}?${TODAY}"
  fallback_url="https://developer.download.nvidia.com/compute/cuda/repos/${CUDA_REPO}/x86_64/${package}"

  echo "Downloading ${package}..."

  # Each package is fetched whole, never resumed, and admitted only on an exact digest match.
  if ! curl -fsSL --retry 3 --retry-delay 2 -o "${package_path}" "${primary_url}"; then
    echo "Primary download failed, trying the NVIDIA repository..."
    if ! curl -fsSL --retry 3 --retry-delay 2 --retry-all-errors -o "${package_path}" "${fallback_url}"; then
      echo "Failed to download ${package}." >&2
      exit 1
    fi
  fi

  echo "Verifying checksum..."
  verify_sha "${sha}" "${package_path}"

  dpkg-deb -x "${package_path}" "${extract_dir}"
done <<<"${CUDA_PACKAGES}"

# Install the shared libraries the execution provider needs. The explicit name list is the guard:
# it admits these five families and nothing else the packages happen to carry, such as the nvblas
# library beside cuBLAS. The stubs directory is pruned for the development packages, which are not
# pinned here and would otherwise offer a link stub under the same name as the real library.
# One flat directory beside the ONNX Runtime, rather than the multiarch path the packages
# target. It keeps both halves of the install in one place and matches install-onnx.sh; a
# distribution CUDA package would land in the multiarch directory and win, so treat one
# appearing there as a conflict rather than an upgrade.
output_lib_dir="${DESTDIR}/lib"
mkdir -p "${output_lib_dir}"

# Collect the whole set before installing any of it, and require every library by name. A check
# per family would pass a cuDNN set that is merely short, and a missing cuDNN library is the one
# fault that surfaces at the first inference rather than at load. Checking first also means a set
# that falls short leaves the destination as it was.
install_files=()

for library in "${CUDA_LIBRARIES[@]}"; do
  matched=0

  while IFS= read -r -d '' file; do
    install_files+=("${file}")
    matched=$((matched + 1))
  done < <(find "${extract_dir}" -path "*/stubs/*" -prune -o -name "${library}.so*" \( -type f -o -type l \) -print0)

  if [[ "${matched}" == 0 ]]; then
    echo "Error: the downloaded packages contain no ${library}." >&2
    exit 1
  fi
done

copied=()

# install_libraries copies the collected set, recording each destination before writing it so
# that a partial file counts as this run's work.
install_libraries() {
  local file target

  for file in "${install_files[@]}"; do
    target="${output_lib_dir}/$(basename "${file}")"
    copied+=("${target}")

    cp -af "${file}" "${output_lib_dir}/" || return 1

    # Set the mode rather than preserve it. A shared object needs neither the execute bit nor
    # any of the special bits, and preserving what a package carried would put a setuid file on
    # the default library path if one ever appeared there.
    if [[ -f ${target} ]] && [[ ! -L ${target} ]]; then
      chmod 0644 "${target}" || return 1
    fi
  done
}

# A set that is short loads and then fails at the first inference, so an interrupted install
# removes what it wrote rather than leaving one behind. Absent libraries are the better
# outcome: the execution provider then reports itself unavailable before a session exists.
if ! install_libraries; then
  echo "Error: failed to install the CUDA libraries, removing what this run wrote." >&2

  for target in "${copied[@]}"; do
    if [[ -f ${target} ]] || [[ -L ${target} ]]; then
      rm -f "${target}"
    fi
  done

  exit 1
fi

installed=${#copied[@]}

if [[ "${DESTDIR}" == "/usr" || "${DESTDIR}" == "/usr/local" ]]; then
  ldconfig
else
  ldconfig -n "${output_lib_dir}" >/dev/null 2>&1 || true
fi

echo "CUDA runtime libraries installed in '${output_lib_dir}' (${installed} files)."
