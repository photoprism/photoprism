#!/bin/bash

# abort if not executed as root
if [[ $(/usr/bin/id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

set -o errexit

if [[ ! -d /tmp ]]; then
  /bin/mkdir /tmp
fi

/bin/chmod 1777 /tmp

/usr/bin/apt-get -y autoremove
/usr/bin/apt-get -y autoclean
/bin/rm -rf /var/lib/apt/lists/*
/bin/rm -rf /tmp/* /var/tmp/*
history -c
/bin/cat /dev/null > /root/.bash_history
unset HISTFILE
/usr/bin/find /var/log -mtime -1 -type f -exec truncate -s 0 {} \;
/bin/rm -rf /var/log/*.gz /var/log/*.log /var/log/*.[0-9] /var/log/*-????????
/bin/rm -rf /var/lib/cloud/instances/*
/bin/rm -f /root/.ssh/* /etc/ssh/*key*

echo "Done."
