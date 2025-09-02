#!/usr/bin/env bash

# Add go alias and custom prompt in /etc/bash.bashrc:
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-profile.sh)

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

set -e

echo "Add ll alias and custom prompt in /etc/bash.bashrc"

echo 'alias ll="ls -alh"' >> /etc/bash.bashrc
echo 'export PS1="\u@$DOCKER_TAG:\w\$ "' >> /etc/bash.bashrc

echo "Add alias of go to richgo if it's available"

# Source: https://stackoverflow.com/questions/592620/how-can-i-check-if-a-program-exists-from-a-bash-script/677212#677212
if command -v richgo >/dev/null 2>&1; then
  echo 'alias go=richgo' >> /etc/bash.bashrc
fi

echo "Remove the unnecessary custom .bashrc for the root user"

rm /root/.bashrc
