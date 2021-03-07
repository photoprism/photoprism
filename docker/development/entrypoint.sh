#!/usr/bin/env bash

if [[ ${UMASK} ]]; then
  echo "umask ${UMASK}"
  umask "${UMASK}"
fi

find /go -type d -print0 | xargs -0 chmod 777
chmod -Rf a+rw /var/lib/photoprism /tmp/photoprism /go

if [[ ${UID} ]] && [[ ${GID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  groupadd -f -g "${GID}" "${GID}"
  usermod -o -u "${UID}" -g "${GID}" photoprism

  if [[ -z ${PHOTOPRISM_DISABLE_CHOWN} ]] ; then
    echo "updating storage permissions..."
    chown -Rf photoprism /photoprism /var/lib/photoprism /tmp/photoprism /go
  fi

  echo "running as uid ${UID}:${GID}"
  echo "${@}"

  gosu "${UID}:${GID}" "$@" &
elif [[ ${UID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  usermod -o -u "${UID}" photoprism

  if [[ -z ${PHOTOPRISM_DISABLE_CHOWN} ]] ; then
    echo "updating storage permissions..."
    chown -Rf photoprism /photoprism /var/lib/photoprism /tmp/photoprism /go
  fi

  echo "running as uid ${UID}"
  echo "${@}"

  gosu "${UID}" "$@" &
else
  echo "running as uid $(id -u)"
  echo "${@}"

  "$@" &
fi

PID=$!

trap "kill $PID" INT TERM
wait