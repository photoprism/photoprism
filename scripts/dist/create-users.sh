#!/bin/sh

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts"

# abort if not executed as root
if [ $(id -u) != "0" ]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

echo "Creating default users and groups..."

groupadd -f -r -g 44 video
groupadd -f -r -g 109 render
groupadd -f -g 1000 photoprism

add_user()
{
  useradd -u "$1" -g photoprism -G video,render -s /bin/bash -m -d "/home/user-$1" "user-$1" 2>/dev/null
}

for i in $(seq 50 99); do add_user "$i"; done
for i in $(seq 500 549); do add_user "$i"; done
for i in $(seq 1000 1099); do add_user "$i"; done

echo "Done."
