#!/usr/bin/env bash

# Installs Go Tools on Linux
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-go-tools.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:/usr/local/go/bin:/go/bin:$PATH"

# abort if not executed as root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

DESTARCH=${2:-$SYSTEM_ARCH}

echo "Installing Go Tools for ${DESTARCH^^}..."

set -e

mkdir -p "$GOPATH/src" "$GOBIN"

# Install gosu in "/usr/local/sbin".
echo "Installing gosu in /usr/local/sbin..."
GOBIN="/usr/local/sbin" go install github.com/tianon/gosu@latest
chown root:root /usr/local/sbin/gosu
chmod 755 /usr/local/sbin/gosu

# Install remaining tools in "/usr/local/bin".
case $DESTARCH in
  arm | ARM | aarch | armv7l | armhf)
    # no additional tools on ARMv7 to reduce build time
    echo "Skipping installation of goimports, go-mod-outdated, exif-read-tool and richgo."
    ;;

  *)
    echo "Installing goimports, go-mod-outdated, exif-read-tool and richgo in /usr/local/bin..."
    GOBIN="/usr/local/bin" go install golang.org/x/tools/cmd/goimports@latest
    GOBIN="/usr/local/bin" go install github.com/psampaz/go-mod-outdated@latest
    GOBIN="/usr/local/bin" go install github.com/dsoprea/go-exif/v3/command/exif-read-tool@latest
    GOBIN="/usr/local/bin" go install github.com/mikefarah/yq/v4@latest
    GOBIN="/usr/local/bin" go install github.com/kyoh86/richgo@latest
    ;;
esac

chmod -R a+rwX "$GOPATH"

echo "Done."
