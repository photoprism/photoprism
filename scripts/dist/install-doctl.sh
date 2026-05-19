#!/usr/bin/env bash

# Downloads and installs the DigitalOcean CLI (doctl) on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-doctl.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Show usage information?
if [[ ${1:-} == "--help" ]]; then
  echo "${0##*/} [version] [destdir] downloads and installs doctl on Linux, for example:" 1>&2
  echo "${0##*/}" 1>&2
  echo "${0##*/} latest" 1>&2
  echo "${0##*/} 1.158.0 /usr/local" 1>&2
  exit 0
fi

set -Eeuo pipefail

# Determine version to install (default: latest):
DOCTL_VERSION=${1:-latest}

if [[ $DOCTL_VERSION == "latest" ]]; then
  # Resolve "latest" via the releases/latest redirect (avoids GitHub API rate limit).
  LATEST_URL=$(curl -fsSL -o /dev/null -w '%{url_effective}' https://github.com/digitalocean/doctl/releases/latest)
  DOCTL_VERSION=${LATEST_URL##*/v}
fi

# Strip leading "v" if user passed e.g. "v1.158.0":
DOCTL_VERSION=${DOCTL_VERSION#v}

if [[ -z $DOCTL_VERSION ]]; then
  echo "doctl version must be passed as first argument, e.g. 1.158.0" 1>&2
  exit 1
fi

# Determine destination directory:
DESTDIR=$(realpath "${2:-/usr/local}")

# Determine the system architecture:
if [[ ${PHOTOPRISM_ARCH:-} ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

DESTARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

case $DESTARCH in
  amd64 | AMD64 | x86_64 | x86-64)
    DESTARCH=amd64
    ;;

  arm64 | ARM64 | aarch64)
    DESTARCH=arm64
    ;;

  arm | ARM | aarch | armv7l | armhf | armv6l | armel)
    echo "doctl is not currently published for 32-bit ARM (armv6/armv7/armhf)." 1>&2
    echo "Use a 64-bit OS (arm64/aarch64) instead." 1>&2
    exit 1
    ;;

  *)
    echo "Unsupported Machine Architecture: \"$DESTARCH\"" 1>&2
    exit 1
    ;;
esac

# Sudo only if not already root:
SUDO=""
if [[ $(id -u) -ne 0 ]]; then
  if command -v sudo >/dev/null 2>&1; then
    SUDO="sudo"
  else
    echo "Run ${0##*/} as root or install sudo!" 1>&2
    exit 1
  fi
fi

ARCHIVE="doctl-${DOCTL_VERSION}-linux-${DESTARCH}.tar.gz"
URL="https://github.com/digitalocean/doctl/releases/download/v${DOCTL_VERSION}/${ARCHIVE}"

echo "Installing doctl ${DOCTL_VERSION} for ${DESTARCH^^} in \"${DESTDIR}/bin\". Please wait."
echo "URL: $URL"

$SUDO mkdir -p "$DESTDIR/bin"

# doctl tarball contains the doctl binary at the archive root.
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL "$URL" | tar -xz -C "$TMPDIR"

if [[ ! -f "$TMPDIR/doctl" ]]; then
  echo "doctl binary not found in archive" 1>&2
  exit 1
fi

$SUDO install -m 755 "$TMPDIR/doctl" "$DESTDIR/bin/doctl"

# Test doctl by showing installed version:
echo "Installed doctl version:"
"$DESTDIR/bin/doctl" version | head -n 1

echo "Done."
