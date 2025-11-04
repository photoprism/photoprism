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

BIN_DIR="${BIN_DIR:-/usr/local/bin}"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}"' EXIT

export PIPX_HOME="${PIPX_HOME:-/opt/pipx}"
export PIPX_BIN_DIR="${PIPX_BIN_DIR:-/usr/local/bin}"
$SUDO install -d -m 0755 "${PIPX_HOME}" "${PIPX_BIN_DIR}"

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
    pipx
    python3-venv
    socat
    yq
  )

  $SUDO apt-get update
  $SUDO apt-get install -y --no-install-recommends "${packages[@]}"
  $SUDO apt-get clean
  $SUDO rm -rf /var/lib/apt/lists/*
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
  $SUDO install -m 0755 "${artifact}" "${BIN_DIR}/kubectl"
}

install_helm() {
  local raw_tag="${HELM_VERSION:-$(fetch_latest_github_tag helm/helm)}"
  local base="helm-${raw_tag}-linux-${LINUX_ARCH}"

  curl -fsSLo "${TMPDIR}/${base}.tar.gz" "https://get.helm.sh/${base}.tar.gz"
  curl -fsSLo "${TMPDIR}/${base}.tar.gz.sha256sum" "https://get.helm.sh/${base}.tar.gz.sha256sum"
  (cd "${TMPDIR}" && sha256sum --check "${base}.tar.gz.sha256sum")
  tar -xzf "${TMPDIR}/${base}.tar.gz" -C "${TMPDIR}"
  $SUDO install -m 0755 "${TMPDIR}/linux-${LINUX_ARCH}/helm" "${BIN_DIR}/helm"
}

install_rancher_cli() {
  local version="${RANCHER_CLI_VERSION:-2.12.3}"
  local tarball="rancher-linux-${LINUX_ARCH}-v${version}.tar.gz"
  local checksum_url="https://releases.rancher.com/cli2/v${version}/${tarball}.sha256sum"
  local checksum_file="${TMPDIR}/${tarball}.sha256sum"

  curl -fsSLo "${TMPDIR}/${tarball}" "https://releases.rancher.com/cli2/v${version}/${tarball}"
  if curl -fsSLo "${checksum_file}" "${checksum_url}"; then
    (cd "${TMPDIR}" && sha256sum --check "${tarball}.sha256sum")
  else
    rm -f "${checksum_file}"
    echo "Checksum file not available for Rancher CLI ${version}; skipping verification." >&2
  fi
  tar -xzf "${TMPDIR}/${tarball}" -C "${TMPDIR}"
  if [[ -f "${TMPDIR}/rancher-v${version}/rancher" ]]; then
    $SUDO install -m 0755 "${TMPDIR}/rancher-v${version}/rancher" "${BIN_DIR}/rancher"
  else
    $SUDO install -m 0755 "${TMPDIR}/rancher-${version}/rancher" "${BIN_DIR}/rancher"
  fi
  if [[ -f "${TMPDIR}/rancher-v${version}/rancher-compose" ]]; then
    $SUDO install -m 0755 "${TMPDIR}/rancher-v${version}/rancher-compose" "${BIN_DIR}/rancher-compose"
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
  $SUDO install -m 0755 "${TMPDIR}/kustomize" "${BIN_DIR}/kustomize"
}

install_k9s() {
  local raw_tag="${K9S_VERSION:-$(fetch_latest_github_tag derailed/k9s)}"
  local version="${raw_tag#v}"
  local artifact="k9s_Linux_${LINUX_ARCH}.tar.gz"
  local checksum_file="${TMPDIR}/checksums.txt"
  local checksum_url="https://github.com/derailed/k9s/releases/download/${raw_tag}/checksums.txt"

  curl -fsSLo "${TMPDIR}/${artifact}" "https://github.com/derailed/k9s/releases/download/${raw_tag}/${artifact}"
  if ! curl -fsSLo "${checksum_file}" "${checksum_url}"; then
    checksum_file="${TMPDIR}/checksums.sha256"
    checksum_url="https://github.com/derailed/k9s/releases/download/${raw_tag}/checksums.sha256"
    curl -fsSLo "${checksum_file}" "${checksum_url}"
  fi
  verify_with_checksums "${checksum_file}" "${TMPDIR}/${artifact}" "${artifact}"
  tar -xzf "${TMPDIR}/${artifact}" -C "${TMPDIR}"
  $SUDO install -m 0755 "${TMPDIR}/k9s" "${BIN_DIR}/k9s"
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
  $SUDO install -m 0755 "${TMPDIR}/stern" "${BIN_DIR}/stern"
}

install_longhornctl() {
  local raw_tag="${LONGHORNCTL_VERSION:-$(fetch_latest_github_tag longhorn/cli)}"
  local version="${raw_tag#v}"
  local artifact="longhornctl-linux-${LINUX_ARCH}"

  curl -fsSLo "${TMPDIR}/${artifact}" "https://github.com/longhorn/cli/releases/download/${raw_tag}/${artifact}"
  curl -fsSLo "${TMPDIR}/${artifact}.sha256" "https://github.com/longhorn/cli/releases/download/${raw_tag}/${artifact}.sha256"
  (cd "${TMPDIR}" && sha256sum --check "$(basename "${artifact}.sha256")")
  $SUDO install -m 0755 "${TMPDIR}/${artifact}" "${BIN_DIR}/longhornctl"
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
  $SUDO install -m 0755 "${TMPDIR}/kubectl-neat" "${BIN_DIR}/kubectl-neat"
}

install_proxysql_admin() {
  if command -v pipx >/dev/null 2>&1; then
    if ! pipx ensurepath --force >/dev/null 2>&1; then
      echo "pipx ensurepath failed; continuing with existing PATH." >&2
    fi
    if ! pipx list 2>/dev/null | grep -q "proxysql-admin-tool"; then
      if ! pipx install proxysql-admin-tool >/dev/null 2>&1; then
        echo "pipx install proxysql-admin-tool from PyPI failed; trying GitHub source." >&2
        if ! pipx install "git+https://github.com/sysown/proxysql-admin-tool.git@master" >/dev/null 2>&1; then
          echo "pipx install proxysql-admin-tool failed; skipping proxysql-admin-tool installation." >&2
        fi
      fi
    fi
  else
    echo "pipx not available; skipping proxysql-admin-tool installation." >&2
  fi
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
install_proxysql_admin

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
  - proxysql-admin-tool (via pipx)
  - Supporting utilities installed via apt

EOF
