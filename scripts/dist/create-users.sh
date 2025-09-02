#!/bin/sh

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Abort if not executed as root.
if [ "$(id -u)" != "0" ]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

echo "Creating default users and groups..."

# create groups 'www-data', 'video', 'davfs2', and 'render'
groupadd -f -r -g 33 www-data 1>&2
echo "✅ Added group www-data (33)."
groupadd -f -r -g 44 video 1>&2
echo "✅ Added group video (44)."
groupadd -f -r -g 105 davfs2 1>&2
echo "✅ Added group davfs2 (105)."
groupadd -f -r -g 109 renderd 1>&2
echo "✅ Added group renderd (109)."
groupadd -f -r -g 115 render 1>&2
echo "✅ Added group render (115)."
groupadd -f -r -g 116 ssl-cert 1>&2
echo "✅ Added group ssl-cert (116)."

# create group 'videodriver'
groupdel -f 937 >/dev/null 2>&1
groupadd -f -r -g 937 videodriver  1>&2
echo "✅ Added group videodriver (937)."

# create group 'photoprism'
groupdel -f ubuntu >/dev/null 2>&1
groupdel -f photoprism >/dev/null 2>&1
groupdel -f 1000 >/dev/null 2>&1
groupadd -f -g 1000 photoprism 1>&2
echo "✅ Added group photoprism (1000)."

# add existing www-data user to groups
usermod -a -G photoprism,video,davfs2,renderd,render,ssl-cert,videodriver www-data

# a directory that will be used as a custom skeleton when creating users
SKELETON="/tmp/custom-skeleton"
mkdir --parents "${SKELETON}"

# copy the .config/ directory from the default skeleton
cp -r /etc/skel/.config/ ${SKELETON}/.config/

# create user 'videodriver'
userdel -r -f videodriver >/dev/null 2>&1
useradd -u 937 -r -N -g 937 -G photoprism,www-data,video,davfs2,renderd,render,ssl-cert -s /bin/bash -m -d "/home/videodriver" --skel "${SKELETON}" videodriver
echo "✅ Added user videodriver (937)."

# create user 'photoprism'
userdel -r -f ubuntu >/dev/null 2>&1
userdel -r -f photoprism >/dev/null 2>&1
userdel -r -f 1000 >/dev/null 2>&1
useradd -u 1000 -N -g photoprism -G www-data,video,davfs2,renderd,render,ssl-cert,videodriver -s /bin/bash -m -d "/home/photoprism" --skel "${SKELETON}" photoprism
echo "✅ Added user photoprism (1000)."

add_user()
{
  userdel -r -f "user-$1" >/dev/null 2>&1
  groupdel -f "group-$1" >/dev/null 2>&1
  groupadd -f -g "$1" "group-$1"
  useradd -u "$1" -g "$1" -G photoprism,www-data,video,davfs2,renderd,render,ssl-cert,videodriver -s /bin/bash -m -d "/home/user-$1" --skel "${SKELETON}" "user-$1" 2>/dev/null
  printf "."
}

printf "👥 adding user/group id ranges 50-99, 500-600, 900-936, 938-999, 1001-1250, and 2000-2100"

for i in $(seq 50 99); do add_user "$i"; done
for i in $(seq 500 600); do add_user "$i"; done
for i in $(seq 900 936); do add_user "$i"; done
for i in $(seq 938 999); do add_user "$i"; done
for i in $(seq 1001 1250); do add_user "$i"; done
for i in $(seq 2000 2100); do add_user "$i"; done

printf " ✔\n"

chgrp -f -R 1000 /home

# remove the custom skeleton
rm -rf "${SKELETON}"

echo "Done."
