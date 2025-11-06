#!/usr/bin/env bash

# Installs Kubernetes/Rancher tooling and supporting utilities for DevOps workflows.
# Intended for use inside the Docker image build where root privileges are available.

set -euo pipefail
IFS=$'\n\t'

PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH"

SUDO=""
if [[ $(id -u) -ne 0 ]]; then
  if command -v sudo >/dev/null 2>&1; then
    SUDO="sudo"
  else
    echo "This script requires root privileges or sudo access." >&2
    exit 1
  fi
fi

case "$(uname -m)" in
  x86_64 | amd64)
    LINUX_ARCH="amd64"
    ;;
  aarch64 | arm64)
    LINUX_ARCH="arm64"
    ;;
  *)
    echo "Unsupported architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

# Ensure installation runs non-interactively.
export DEBIAN_FRONTEND="noninteractive"

BIN_DIR="${BIN_DIR:-/usr/local/bin}"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}"' EXIT

install_apt_packages() {
  local packages=(
    bash-completion
    dnsutils
    jq
    mariadb-client
    mysql-shell
    netcat-openbsd
    nfs-common
    percona-toolkit
    socat
    yq
  )

  ${SUDO} env DEBIAN_FRONTEND="noninteractive" apt-get update
  ${SUDO} env DEBIAN_FRONTEND="noninteractive" apt-get install -y --no-install-recommends "${packages[@]}"
  ${SUDO} env DEBIAN_FRONTEND="noninteractive" apt-get clean
  ${SUDO} rm -rf /var/lib/apt/lists/*
}

fetch_latest_github_tag() {
  local repo="$1"
  curl -fsSL "https://api.github.com/repos/${repo}/releases/latest" | jq -r '.tag_name'
}

verify_with_checksums() {
  local checksum_file="$1"
  local artifact="$2"
  local pattern="$3"

  local sum
  sum="$(awk -v target="${pattern}" '$2 == target {print $1; exit}' "${checksum_file}")"
  if [[ -z "${sum}" ]]; then
    echo "Checksum for ${pattern} not found in ${checksum_file}" >&2
    exit 1
  fi
  echo "${sum}  ${artifact}" | sha256sum --check --status
}

install_kubectl() {
  local version="${KUBECTL_VERSION:-$(curl -fsSL https://dl.k8s.io/release/stable.txt)}"
  local artifact="${TMPDIR}/kubectl"

  curl -fsSLo "${artifact}" "https://dl.k8s.io/release/${version}/bin/linux/${LINUX_ARCH}/kubectl"
  curl -fsSLo "${artifact}.sha256" "https://dl.k8s.io/release/${version}/bin/linux/${LINUX_ARCH}/kubectl.sha256"
  (
    cd "${TMPDIR}"
    printf '%s  %s\n' "$(cat "$(basename "${artifact}.sha256")")" "$(basename "${artifact}")" | sha256sum --check --status -
  )
  ${SUDO} install -m 0755 "${artifact}" "${BIN_DIR}/kubectl"
}

install_helm() {
  local raw_tag="${HELM_VERSION:-$(fetch_latest_github_tag helm/helm)}"
  local base="helm-${raw_tag}-linux-${LINUX_ARCH}"

  curl -fsSLo "${TMPDIR}/${base}.tar.gz" "https://get.helm.sh/${base}.tar.gz"
  curl -fsSLo "${TMPDIR}/${base}.tar.gz.sha256sum" "https://get.helm.sh/${base}.tar.gz.sha256sum"
  (cd "${TMPDIR}" && sha256sum --check "${base}.tar.gz.sha256sum")
  tar -xzf "${TMPDIR}/${base}.tar.gz" -C "${TMPDIR}"
  ${SUDO} install -m 0755 "${TMPDIR}/linux-${LINUX_ARCH}/helm" "${BIN_DIR}/helm"
}

install_rancher_cli() {
  local version="${RANCHER_CLI_VERSION:-2.12.3}"
  local tarball="rancher-linux-${LINUX_ARCH}-v${version}.tar.gz"
  local base_url="https://github.com/rancher/cli/releases/download/v${version}"
  local checksum_file="${TMPDIR}/rancher-cli-sha256sum.txt"

  curl -fsSLo "${TMPDIR}/${tarball}" "${base_url}/${tarball}"
  if curl -fsSLo "${checksum_file}" "${base_url}/sha256sum.txt"; then
    local expected
    expected="$(awk -v name="${tarball}" '$2 == name {print $1}' "${checksum_file}")"
    if [[ -z "${expected}" ]]; then
      echo "Checksum entry for ${tarball} not found in sha256sum.txt; skipping verification." >&2
    else
      echo "${expected}  ${TMPDIR}/${tarball}" | sha256sum --check --status -
    fi
  else
    rm -f "${checksum_file}"
    echo "Checksum file not available for Rancher CLI ${version}; skipping verification." >&2
  fi
  tar -xzf "${TMPDIR}/${tarball}" -C "${TMPDIR}"
  if [[ -f "${TMPDIR}/rancher-v${version}/rancher" ]]; then
    ${SUDO} install -m 0755 "${TMPDIR}/rancher-v${version}/rancher" "${BIN_DIR}/rancher"
  else
    ${SUDO} install -m 0755 "${TMPDIR}/rancher-${version}/rancher" "${BIN_DIR}/rancher"
  fi
  if [[ -f "${TMPDIR}/rancher-v${version}/rancher-compose" ]]; then
    ${SUDO} install -m 0755 "${TMPDIR}/rancher-v${version}/rancher-compose" "${BIN_DIR}/rancher-compose"
  fi
}

install_kustomize() {
  local raw_tag="${KUSTOMIZE_VERSION:-$(fetch_latest_github_tag kubernetes-sigs/kustomize)}"
  local version="${raw_tag##*/}"
  local encoded_tag="${raw_tag//\//%2F}"
  local artifact="kustomize_${version}_linux_${LINUX_ARCH}.tar.gz"
  local checksum_file="${TMPDIR}/checksums.txt"

  curl -fsSLo "${TMPDIR}/${artifact}" "https://github.com/kubernetes-sigs/kustomize/releases/download/${encoded_tag}/${artifact}"
  curl -fsSLo "${checksum_file}" "https://github.com/kubernetes-sigs/kustomize/releases/download/${encoded_tag}/checksums.txt"
  verify_with_checksums "${checksum_file}" "${TMPDIR}/${artifact}" "${artifact}"
  tar -xzf "${TMPDIR}/${artifact}" -C "${TMPDIR}"
  ${SUDO} install -m 0755 "${TMPDIR}/kustomize" "${BIN_DIR}/kustomize"
}

install_k9s() {
  local raw_tag="${K9S_VERSION:-$(fetch_latest_github_tag derailed/k9s)}"
  local version="${raw_tag#v}"
  local artifact="k9s_Linux_${LINUX_ARCH}.tar.gz"
  local checksum_file
  local checksum_url
  local base_url="https://github.com/derailed/k9s/releases/download/${raw_tag}"

  curl -fsSLo "${TMPDIR}/${artifact}" "${base_url}/${artifact}"

  if curl -sfI "${base_url}/checksums.sha256" >/dev/null; then
    checksum_url="${base_url}/checksums.sha256"
    checksum_file="${TMPDIR}/checksums.sha256"
  elif curl -sfI "${base_url}/checksums.txt" >/dev/null; then
    checksum_url="${base_url}/checksums.txt"
    checksum_file="${TMPDIR}/checksums.txt"
  else
    echo "Checksum file not found for k9s ${raw_tag}; skipping verification." >&2
    checksum_file=""
  fi
  if [[ -n "${checksum_file}" ]]; then
    curl -fsSLo "${checksum_file}" "${checksum_url}"
    verify_with_checksums "${checksum_file}" "${TMPDIR}/${artifact}" "${artifact}"
  fi
  tar -xzf "${TMPDIR}/${artifact}" -C "${TMPDIR}"
  ${SUDO} install -m 0755 "${TMPDIR}/k9s" "${BIN_DIR}/k9s"
}

install_stern() {
  local raw_tag="${STERN_VERSION:-$(fetch_latest_github_tag stern/stern)}"
  local version="${raw_tag#v}"
  local artifact="stern_${version}_linux_${LINUX_ARCH}.tar.gz"
  local checksum_file="${TMPDIR}/checksums.txt"

  curl -fsSLo "${TMPDIR}/${artifact}" "https://github.com/stern/stern/releases/download/${raw_tag}/${artifact}"
  curl -fsSLo "${checksum_file}" "https://github.com/stern/stern/releases/download/${raw_tag}/checksums.txt"
  verify_with_checksums "${checksum_file}" "${TMPDIR}/${artifact}" "${artifact}"
  tar -xzf "${TMPDIR}/${artifact}" -C "${TMPDIR}"
  ${SUDO} install -m 0755 "${TMPDIR}/stern" "${BIN_DIR}/stern"
}

install_longhornctl() {
  local raw_tag="${LONGHORNCTL_VERSION:-$(fetch_latest_github_tag longhorn/cli)}"
  local version="${raw_tag#v}"
  local artifact="longhornctl-linux-${LINUX_ARCH}"

  curl -fsSLo "${TMPDIR}/${artifact}" "https://github.com/longhorn/cli/releases/download/${raw_tag}/${artifact}"
  curl -fsSLo "${TMPDIR}/${artifact}.sha256" "https://github.com/longhorn/cli/releases/download/${raw_tag}/${artifact}.sha256"
  (cd "${TMPDIR}" && sha256sum --check "$(basename "${artifact}.sha256")")
  ${SUDO} install -m 0755 "${TMPDIR}/${artifact}" "${BIN_DIR}/longhornctl"
}

install_kubectl_neat() {
  local raw_tag="${KUBECTL_NEAT_VERSION:-$(fetch_latest_github_tag itaysk/kubectl-neat)}"
  local version="${raw_tag#v}"
  local artifact="kubectl-neat_linux_${LINUX_ARCH}.tar.gz"
  local checksum_file="${TMPDIR}/checksums.txt"

  curl -fsSLo "${TMPDIR}/${artifact}" "https://github.com/itaysk/kubectl-neat/releases/download/${raw_tag}/${artifact}"
  curl -fsSLo "${checksum_file}" "https://github.com/itaysk/kubectl-neat/releases/download/${raw_tag}/checksums.txt"
  verify_with_checksums "${checksum_file}" "${TMPDIR}/${artifact}" "${artifact}"
  tar -xzf "${TMPDIR}/${artifact}" -C "${TMPDIR}"
  ${SUDO} install -m 0755 "${TMPDIR}/kubectl-neat" "${BIN_DIR}/kubectl-neat"
}

install_apt_packages
install_kubectl
install_helm
install_rancher_cli
install_kustomize
install_k9s
install_stern
install_longhornctl
install_kubectl_neat

cat <<'EOF'

DevOps tooling installation completed:
  - kubectl
  - helm
  - rancher CLI
  - kustomize
  - k9s
  - stern
  - longhornctl
  - kubectl-neat
  - Supporting utilities installed via apt

EOF
