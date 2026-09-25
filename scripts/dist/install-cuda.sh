#!/usr/bin/env bash

set -euo pipefail

# Installs the NVIDIA CUDA runtime libraries that the ONNX Runtime CUDA Execution Provider needs.
#
# The provider links cudart, cuBLAS, cuBLASLt and cuRAND, and loads cuDNN on demand. None of them
# are part of the ONNX Runtime archive, and the NVIDIA Container Toolkit injects the host driver
# only, so they have to be installed separately. Pair this with "install-onnx.sh --gpu", which
# installs the runtime build that carries libonnxruntime_providers_cuda.so.
#
# Packages are downloaded directly from NVIDIA and verified against the pinned checksums.
# Package versions move independently, so selecting another CUDA release requires updating
# the filenames and checksums as a set.
#
# One pinned set serves every base image we ship. The packages are built for one distribution
# release, but they are extracted rather than installed, and the vendor builds them against an
# old C library, so what decides is the glibc version rather than the distribution name. The
# script tests for that below instead of mapping a release to a repository.
#
# Requires the proprietary NVIDIA driver on the host, and for Docker the NVIDIA Container Toolkit.
# CUDA 13 needs driver 580 or later; an older one loads the libraries and then reports no usable
# device, which PhotoPrism treats as "no GPU" and falls back to the CPU.

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
transaction_dir=""
install_complete=0
published=()

# cleanup restores an incomplete publication and removes temporary files.
# Failed recovery keeps the snapshots available for an operator to restore.
cleanup() {
  local result="$1" rollback_failed=0 name i

  trap - EXIT
  trap '' HUP INT TERM

  if [[ -n "${transaction_dir}" ]] && [[ "${install_complete}" == 0 ]]; then
    for ((i=${#published[@]}-1; i>=0; i--)); do
      name="${published[i]}"

      # A source still in staging means its atomic rename did not happen.
      if [[ -e "${transaction_dir}/new/${name}" || -L "${transaction_dir}/new/${name}" ]]; then
        continue
      fi

      if [[ -e "${transaction_dir}/previous/${name}" || -L "${transaction_dir}/previous/${name}" ]]; then
        if ! mv -fT -- "${transaction_dir}/previous/${name}" "${output_lib_dir}/${name}"; then
          rollback_failed=1
        fi
      elif ! rm -f -- "${output_lib_dir}/${name}"; then
        rollback_failed=1
      fi
    done
  fi

  if [[ "${rollback_failed}" != 0 ]]; then
    echo "Error: CUDA library recovery is incomplete; preserved files are in '${transaction_dir}'." >&2
    [[ "${result}" != 0 ]] || result=1
  elif [[ -n "${transaction_dir}" ]] && ! rm -rf -- "${transaction_dir}"; then
    echo "Error: failed to remove CUDA installation staging '${transaction_dir}'." >&2
    [[ "${result}" != 0 ]] || result=1
  fi

  if ! rm -rf -- "${work_dir}"; then
    echo "Error: failed to remove CUDA download staging '${work_dir}'." >&2
    [[ "${result}" != 0 ]] || result=1
  fi

  exit "${result}"
}

trap 'cleanup "$?"' EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

chmod 0700 "${work_dir}"
extract_dir="${work_dir}/extracted"
mkdir -p "${extract_dir}"

while read -r package sha; do
  [[ -z "${package}" ]] && continue

  package_path="${work_dir}/${package}"
  package_url="https://developer.download.nvidia.com/compute/cuda/repos/${CUDA_REPO}/x86_64/${package}"

  echo "Downloading ${package}..."

  # Each package is fetched whole, never resumed, and admitted only on an exact digest match.
  if ! curl -fsSL --retry 3 --retry-delay 2 --retry-all-errors -o "${package_path}" "${package_url}"; then
    echo "Failed to download ${package}." >&2
    exit 1
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

transaction_dir=$(mktemp -d "${output_lib_dir}/.cuda-XXXXXX")
mkdir -p "${transaction_dir}/new" "${transaction_dir}/previous"
install_names=()

# prepare_libraries stages the full set and preserves existing files and symlinks.
# Snapshots share old inodes; only staged replacements receive permission changes.
prepare_libraries() {
  local file name target staged

  for file in "${install_files[@]}"; do
    name=$(basename "${file}")
    target="${output_lib_dir}/${name}"
    staged="${transaction_dir}/new/${name}"

    if [[ -e "${staged}" || -L "${staged}" ]]; then
      echo "Error: the CUDA packages contain duplicate library '${name}'." >&2
      return 1
    fi

    if [[ -e "${target}" && ! -f "${target}" && ! -L "${target}" ]]; then
      echo "Error: CUDA library destination '${target}' is not a file or symlink." >&2
      return 1
    fi

    cp -a -- "${file}" "${staged}" || return 1

    if [[ -f "${staged}" && ! -L "${staged}" ]]; then
      chmod 0644 "${staged}" || return 1
    fi

    if [[ -e "${target}" || -L "${target}" ]]; then
      cp -a --link -- "${target}" "${transaction_dir}/previous/${name}" || return 1
    fi

    install_names+=("${name}")
  done
}

if ! prepare_libraries; then
  echo "Error: failed to prepare the CUDA libraries; the installed files are unchanged." >&2
  exit 1
fi

# publish_libraries replaces each entry atomically without changing old library inodes.
publish_libraries() {
  local name

  for name in "${install_names[@]}"; do
    published+=("${name}")
    mv -fT -- "${transaction_dir}/new/${name}" "${output_lib_dir}/${name}" || return 1
  done
}

if ! publish_libraries; then
  echo "Error: failed to publish the CUDA libraries; restoring the installed files." >&2
  exit 1
fi

# The complete set is committed before refreshing the loader cache.
install_complete=1
installed=${#install_names[@]}

if [[ "${DESTDIR}" == "/usr" || "${DESTDIR}" == "/usr/local" ]]; then
  ldconfig
else
  ldconfig -n "${output_lib_dir}" >/dev/null 2>&1 || true
fi

echo "CUDA runtime libraries installed in '${output_lib_dir}' (${installed} files)."
