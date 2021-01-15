#!/usr/bin/env bash

if [[ ${UMASK} ]]; then
  umask "${UMASK}"
fi

find /go -type d -print0 | xargs -0 chmod 777
chmod -R a+rw /var/lib/photoprism /go

if [[ ${UID} ]] && [[ ${GID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  groupadd -f -g "${GID}" "${GID}"
  usermod -o -u "${UID}" -g "${GID}" photoprism
  chown -R photoprism:photoprism /photoprism /var/lib/photoprism /go
  gosu "${UID}:${GID}" "$@" &
elif [[ ${UID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  usermod -o -u "${UID}" photoprism
  chown -R photoprism /photoprism /var/lib/photoprism /go
  gosu "${UID}" "$@" &
else
  "$@" &
fi

PID=$!

trap "kill $PID" INT TERM
wait