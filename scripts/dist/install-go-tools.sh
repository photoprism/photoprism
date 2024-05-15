#!/usr/bin/env bash

# This installs Go tools on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-go-tools.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:/usr/local/go/bin:/go/bin:$PATH"

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

if ! command -v go &> /dev/null
then
    echo "Go must be installed to run this."
    exit 1
fi

# Determine the system architecture.
if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

DESTARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

if [ -d "/go" ]; then
  GOPATH="/go"
elif [[ -z $GOPATH ]]; then
  GOPATH=$(go env GOPATH)
fi

set -e

mkdir -p "$GOPATH/src"

# Install remaining tools in "/usr/local/bin".
case $DESTARCH in
  arm | ARM | aarch | armv7l | armhf)
    echo "Installing Go tools for ${DESTARCH^^} in /usr/local/bin..."
    GOBIN="/usr/local/bin" go install golang.org/x/tools/cmd/goimports@latest
    GOBIN="/usr/local/bin" go install github.com/psampaz/go-mod-outdated@latest
    GOBIN="/usr/local/bin" go install github.com/kyoh86/richgo@latest
    GOBIN="/usr/local/bin" go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
    ;;

  *)
    echo "Installing Go tools for ${DESTARCH^^} in /usr/local/bin..."
    GOBIN="/usr/local/bin" go install golang.org/x/tools/cmd/goimports@latest
    GOBIN="/usr/local/bin" go install golang.org/x/tools/cmd/godoc@latest
    GOBIN="/usr/local/bin" go install github.com/psampaz/go-mod-outdated@latest
    GOBIN="/usr/local/bin" go install github.com/mikefarah/yq/v4@latest
    GOBIN="/usr/local/bin" go install github.com/kyoh86/richgo@latest
    GOBIN="/usr/local/bin" go install github.com/muesli/duf@latest
    GOBIN="/usr/local/bin" go install github.com/go-delve/delve/cmd/dlv@latest
    GOBIN="/usr/local/bin" go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
    GOBIN="/usr/local/bin" go install github.com/google/go-licenses@latest
    ;;
esac

chmod -R a+rwX "$GOPATH"

echo "Done."
