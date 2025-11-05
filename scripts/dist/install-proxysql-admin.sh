#!/usr/bin/env bash
set -euo pipefail

PROXYSQL_VERSION="3.0.1"
PROXYSQL_REVISION="1.1"
ARCH_RAW="$(uname -m)"

case "${ARCH_RAW}" in
  x86_64|amd64)
    DEB_DIR="x86_64"
    DEB_ARCH="amd64"
    RPM_ARCH="x86_64"
    TAR_ARCH="x86_64"
    ;;
  aarch64|arm64)
    DEB_DIR="aarch64"
    DEB_ARCH="arm64"
    RPM_ARCH="aarch64"
    TAR_ARCH="aarch64"
    ;;
  *)
    echo "proxysql-admin tooling is only provided for x86_64/arm64; detected ${ARCH_RAW}." >&2
    exit 1
    ;;
esac

if [[ ! -r /etc/os-release ]]; then
  echo "Cannot detect operating system (missing /etc/os-release)." >&2
  exit 1
fi

# shellcheck source=/dev/null
. /etc/os-release

SUDO=${SUDO:-sudo}
TMPDIR=$(mktemp -d)
trap 'rm -rf "${TMPDIR}"' EXIT

# Ensure installation runs non-interactively.
if [[ "${ID_LIKE:-}" =~ (debian|ubuntu) ]] || [[ "${ID:-}" =~ (debian|ubuntu) ]]; then
  export DEBIAN_FRONTEND="noninteractive"
fi

patch_mysql_client_version_parser() {
  local file="/usr/bin/proxysql-common"
  local pattern='[[:space:]]11\\.'
  if [[ ! -f "${file}" ]]; then
    return
  fi
  if grep -q "${pattern}" "${file}"; then
    return
  fi
  if ! command -v python3 >/dev/null 2>&1; then
    echo "python3 not found; skipping proxysql-common patch." >&2
    return
  fi
  ${SUDO} python3 - <<'PY'
from pathlib import Path
path = Path('/usr/bin/proxysql-common')
text = path.read_text()
target = '  elif echo "$version_string" | grep -qe "[[:space:]]10\\.11\\."; then\n    echo "10.11"\n'
if target not in text:
    raise SystemExit(0)
replacement = target + '  elif echo "$version_string" | grep -qe "[[:space:]]11\\."; then\n    echo "11.0"\n'
if '[[:space:]]11\\.' not in text:
    text = text.replace(target, replacement, 1)
    path.write_text(text)
PY
}

stop_disable_service() {
  command -v systemctl >/dev/null 2>&1 || return
  local svc
  for svc in proxysql proxysql-initial; do
    ${SUDO} systemctl stop "${svc}.service" >/dev/null 2>&1 || true
    ${SUDO} systemctl disable "${svc}.service" --now >/dev/null 2>&1 || true
    ${SUDO} systemctl mask "${svc}.service" >/dev/null 2>&1 || true
  done
}

install_from_deb() {
  local codename pkg url deb_arch
  codename=${VERSION_CODENAME:-}
  codename=${codename,,}
  if [[ ! ${codename} =~ ^[a-z]+$ ]]; then
    codename=""
  fi
  if [[ ${codename} != "jammy" && ${codename} != "noble" ]]; then
    codename="noble"
  fi
  case "${codename}" in
    jammy|noble) ;; # supported
    *) codename=noble ;;
  esac
  deb_arch=${DEB_ARCH}
  pkg="proxysql3_${PROXYSQL_VERSION}-${PROXYSQL_REVISION}.${codename}_${deb_arch}.deb"
  url="https://downloads.percona.com/downloads/proxysql3/proxysql3-${PROXYSQL_VERSION}/binary/debian/${codename}/${DEB_DIR}/${pkg}"
  echo "Downloading ${pkg}..."
  # Allow APT's sandbox user (_apt) to read the download directory & file
  ${SUDO} chmod 755 "${TMPDIR}"
  curl -fsSL "${url}" -o "${TMPDIR}/${pkg}"
  ${SUDO} chmod 644 "${TMPDIR}/${pkg}"
  echo "Installing ${pkg}..."
  ${SUDO} env DEBIAN_FRONTEND="noninteractive" apt-get update -y >/dev/null
  # Use absolute path so _apt can access it without relying on CWD permissions
  ${SUDO} env DEBIAN_FRONTEND="noninteractive" apt-get install -y "${TMPDIR}/${pkg}"
  stop_disable_service
}

install_from_rpm() {
  local major pkg url rpm_arch
  major=${VERSION_ID%%.*}
  [[ -z "${major}" ]] && major=${PROXYSQL_VERSION%%.*}
  rpm_arch=${RPM_ARCH}
  pkg="proxysql3-${PROXYSQL_VERSION}-${PROXYSQL_REVISION}.el${major}.${rpm_arch}.rpm"
  url="https://downloads.percona.com/downloads/proxysql3/proxysql3-${PROXYSQL_VERSION}/binary/redhat/${major}/${rpm_arch}/${pkg}"
  echo "Downloading ${pkg}..."
  curl -fsSL "${url}" -o "${TMPDIR}/${pkg}"
  echo "Installing ${pkg}..."
  if command -v dnf >/dev/null 2>&1; then
    ${SUDO} dnf install -y "${TMPDIR}/${pkg}"
  else
    ${SUDO} yum install -y "${TMPDIR}/${pkg}"
  fi
  stop_disable_service
}

install_from_tarball() {
  local ver tarball url target_dir arch
  ver=$(ldd --version | head -n1 | awk '{print $NF}')
  arch=${TAR_ARCH}
  tarball="proxysql-${PROXYSQL_VERSION}-Linux-${arch}.glibc${ver}.tar.gz"
  url="https://downloads.percona.com/downloads/proxysql3/proxysql3-${PROXYSQL_VERSION}/binary/tarball/${tarball}"
  echo "Downloading ${tarball}..."
  curl -fsSL "${url}" -o "${TMPDIR}/${tarball}" || {
    echo "Falling back to glibc2.34 build." >&2
    tarball="proxysql-${PROXYSQL_VERSION}-Linux-${arch}.glibc2.34.tar.gz"
    url="https://downloads.percona.com/downloads/proxysql3/proxysql3-${PROXYSQL_VERSION}/binary/tarball/${tarball}"
    curl -fsSL "${url}" -o "${TMPDIR}/${tarball}"
  }
  target_dir="/usr/local/proxysql-${PROXYSQL_VERSION}"
  echo "Extracting ${tarball} to ${target_dir}..."
  ${SUDO} rm -rf "${target_dir}"
  ${SUDO} mkdir -p "${target_dir}"
  ${SUDO} tar -xzf "${TMPDIR}/${tarball}" -C "${target_dir}" --strip-components=1
  for bin in proxysql-admin proxysql-admin-common percona-scheduler-admin; do
    if [[ -f "${target_dir}/usr/bin/${bin}" ]]; then
      ${SUDO} install -m 0755 "${target_dir}/usr/bin/${bin}" "/usr/local/bin/${bin}"
    fi
  done
  if [[ -f "${target_dir}/etc/proxysql-admin.cnf" ]]; then
    ${SUDO} install -d /usr/local/share/proxysql-admin
    ${SUDO} install -m 0644 "${target_dir}/etc/proxysql-admin.cnf" /usr/local/share/proxysql-admin/proxysql-admin.cnf.sample
  fi
}

detect_installed_package_version() {
  if command -v dpkg-query >/dev/null 2>&1; then
    dpkg-query -W -f='${Version}' proxysql3 2>/dev/null || true
    return
  fi
  if command -v rpm >/dev/null 2>&1; then
    rpm -q --queryformat '%{VERSION}-%{RELEASE}' proxysql3 2>/dev/null || true
    return
  fi
  printf ''
}

existing_path="$(command -v proxysql-admin || true)"
if [[ -n "${existing_path}" ]]; then
  install_required="yes"
  installed_pkg_version="$(detect_installed_package_version)"
  resolved_path="$(readlink -f "${existing_path}" 2>/dev/null || true)"
  if [[ -n "${installed_pkg_version}" ]]; then
    if [[ "${installed_pkg_version}" == *"${PROXYSQL_VERSION}"* ]]; then
      install_required="no"
    fi
  elif [[ -n "${resolved_path}" && "${resolved_path}" == *"/usr/local/proxysql-${PROXYSQL_VERSION}/"* ]]; then
    install_required="no"
  fi

  if [[ "${install_required}" == "no" ]]; then
    echo "proxysql-admin already detected at ${existing_path}; refreshing proxysql-common version parser patch."
    patch_mysql_client_version_parser
    exit 0
  fi

  echo "proxysql-admin detected at ${existing_path} (package version: ${installed_pkg_version:-unknown}); reinstalling ProxySQL tooling ${PROXYSQL_VERSION}-${PROXYSQL_REVISION}."
fi

if [[ "${ID_LIKE:-}" =~ (debian|ubuntu) ]] || [[ "${ID:-}" =~ (debian|ubuntu) ]]; then
  install_from_deb
elif [[ "${ID_LIKE:-}" =~ (rhel|centos|rocky|fedora) ]] || [[ "${ID:-}" =~ (rhel|centos|rocky|fedora) ]]; then
  install_from_rpm
else
  echo "Unknown distribution (${ID}); installing from generic tarball." >&2
  install_from_tarball
fi

patch_mysql_client_version_parser

if ! command -v proxysql-admin >/dev/null 2>&1; then
  echo "proxysql-admin installation did not place the binary on PATH." >&2
  exit 1
fi

echo "proxysql-admin installed at $(command -v proxysql-admin)."
