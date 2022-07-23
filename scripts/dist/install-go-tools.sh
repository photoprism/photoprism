#!/usr/bin/env bash

# Installs Go Tools on Linux
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-go-tools.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:/usr/local/go/bin:/go/bin:$PATH"

# abort if not executed as root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

if [[ -z "$GOPATH" ]] || [[ -z "$GOBIN" ]]; then
  echo "\$GOPATH and \$GOBIN must be set" 1>&2
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

go install github.com/tianon/gosu@latest

case $DESTARCH in
  arm | ARM | aarch | armv7l | armhf)
    # no additional tools on ARMv7 to reduce build time
    echo "Skipping installation of goimports, go-mod-outdated, exif-read-tool and richgo."
    ;;

  *)
    go install golang.org/x/tools/cmd/goimports@latest
    go install github.com/psampaz/go-mod-outdated@latest
    go install github.com/dsoprea/go-exif/v3/command/exif-read-tool@latest
    go install github.com/mikefarah/yq/v4@latest

    go install github.com/kyoh86/richgo@latest
    cp "$GOBIN/richgo" /usr/local/bin/richgo
    ;;
esac

chmod -R a+rwX "$GOPATH"

# install gosu in /usr/local/sbin
cp "$GOBIN/gosu" /usr/local/sbin/gosu
chown root:root /usr/local/sbin/gosu
chmod 755 /usr/local/sbin/gosu

echo "Done."
