#!/usr/bin/env bash

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

set -o errexit

if [[ ! -d /tmp ]]; then
  mkdir /tmp
fi

chmod 1777 /tmp

apt-get -y autoremove
apt-get -y autoclean
rm -rf /var/lib/apt/lists/*
rm -rf /tmp/* /var/tmp/*
history -c
cat /dev/null > /root/.bash_history
unset HISTFILE
find /var/log -mtime -1 -type f -exec truncate -s 0 {} \;
rm -rf /var/log/*.gz /var/log/*.log /var/log/*.[0-9] /var/log/*-????????
rm -rf /var/lib/cloud/instances/*
rm -f /root/.ssh/* /etc/ssh/*key*

echo "Done."
