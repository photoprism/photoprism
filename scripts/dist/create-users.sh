#!/bin/sh

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts"

# abort if not executed as root
if [ $(id -u) != "0" ]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

echo "Creating default users and groups..."

# create groups 'video' and 'render'
groupadd -f -r -g 44 video 1>&2
groupadd -f -r -g 109 render 1>&2
echo "✅ added groups 'video' (44) and 'render' (109)"

# create user/group 'videodriver'
groupdel -f 937 >/dev/null 2>&1
groupadd -f -r -g 937 videodriver  1>&2
userdel -r -f videodriver >/dev/null 2>&1
useradd -u 937 -r -N -g 937 -G photoprism,video,render -s /bin/bash -m -d "/home/videodriver" videodriver
echo "✅ added user/group 'videodriver' (937)"

# create group 'photoprism'
groupdel -f 1000 >/dev/null 2>&1
groupadd -f -g 1000 photoprism 1>&2
echo "✅ added group 'photoprism' (1000)"

# create user 'photoprism'
userdel -r -f photoprism >/dev/null 2>&1
userdel -r -f 1000 >/dev/null 2>&1
useradd -u 1000 -N -g 1000 -G video,render,videodriver -s /bin/bash -m -d "/home/photoprism" photoprism
echo "✅ added user 'photoprism' (1000)"

add_user()
{
  userdel -r -f "user-$1" >/dev/null 2>&1
  groupdel -f "group-$1" >/dev/null 2>&1
  groupadd -f -g "$1" "group-$1"
  useradd -u "$1" -g "$1" -G photoprism,video,render,videodriver -s /bin/bash -m -d "/home/user-$1" "user-$1" 2>/dev/null
  printf "."
}

printf "👥 adding user/group id ranges 50-99, 500-549, 900-936, 937-949, and 1001-1099"

for i in $(seq 50 99); do add_user "$i"; done
for i in $(seq 500 549); do add_user "$i"; done
for i in $(seq 900 936); do add_user "$i"; done
for i in $(seq 938 949); do add_user "$i"; done
for i in $(seq 1001 1099); do add_user "$i"; done

printf " ✔\n"

chgrp -f -R 1000 /home

echo "Done."
