#!/bin/sh

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts"

# abort if not executed as root
if [ $(id -u) != "0" ]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

echo "Creating default users and groups..."

groupadd -f -r -g 44 video 1>&2
groupadd -f -r -g 109 render 1>&2

groupdel -f 1000 >/dev/null 2>&1
userdel -f photoprism >/dev/null 2>&1
userdel -f 1000 >/dev/null 2>&1

groupadd -f -g 1000 photoprism 1>&2
useradd -N -o -u 1000 -g photoprism -G video,render -s /bin/bash -m -d "/home/photoprism" photoprism

add_user()
{
  userdel -f "$1" >/dev/null 2>&1
  groupdel -f "group-$1" >/dev/null 2>&1
  groupdel -f "$1" >/dev/null 2>&1
  groupadd -f -g "$1" "group-$1"
  useradd -u "$1" -g "$1" -G photoprism,video,render -s /bin/bash -m -d "/home/user-$1" "user-$1" 2>/dev/null
}

for i in $(seq 50 99); do add_user "$i"; done
for i in $(seq 500 549); do add_user "$i"; done
for i in $(seq 1001 1099); do add_user "$i"; done

chgrp -f -R 1000 /home

echo "Done."
