#!/usr/bin/env bash

if [[ ${UMASK} ]]; then
  umask "${UMASK}"
fi

if [[ ${UID} ]] && [[ ${GID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  groupadd -f -g "${GID}" "${GID}"
  usermod -o -u "${UID}" -g "${GID}" photoprism
  chown -Rf photoprism /photoprism/assets /photoprism/storage /photoprism/import /var/lib/photoprism /tmp/photoprism
  gosu "${UID}:${GID}" "$@" &
elif [[ ${UID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  usermod -o -u "${UID}" photoprism
  chown -Rf photoprism /photoprism/assets /photoprism/storage /photoprism/import /var/lib/photoprism /tmp/photoprism
  gosu "${UID}" "$@" &
else
  "$@" &
fi

PID=$!

trap "kill $PID" INT TERM
wait