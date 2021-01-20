#!/usr/bin/env bash

if [[ ${UMASK} ]]; then
  umask "${UMASK}"
fi

find /go -type d -print0 | xargs -0 chmod 777
chmod -Rf a+rw /var/lib/photoprism /tmp/photoprism /go

if [[ ${UID} ]] && [[ ${GID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  groupadd -f -g "${GID}" "${GID}"
  usermod -o -u "${UID}" -g "${GID}" photoprism
  chown -Rf photoprism /photoprism /var/lib/photoprism /tmp/photoprism /go
  gosu "${UID}:${GID}" "$@" &
elif [[ ${UID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  usermod -o -u "${UID}" photoprism
  chown -Rf photoprism /photoprism /var/lib/photoprism /tmp/photoprism /go
  gosu "${UID}" "$@" &
else
  "$@" &
fi

PID=$!

trap "kill $PID" INT TERM
wait