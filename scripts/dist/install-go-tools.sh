#!/bin/bash

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:/usr/local/go/bin:/go/bin"

# abort if not executed as root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

echo "Installing Go Tools..."

set -e

go install github.com/tianon/gosu@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/kyoh86/richgo@latest
go install github.com/psampaz/go-mod-outdated@latest
go install github.com/dsoprea/go-exif/v3/command/exif-read-tool@latest
go install github.com/mikefarah/yq/v4@latest