#!/usr/bin/env bash

if [[ ${UMASK} ]]; then
  echo "entrypoint: umask ${UMASK}"
  umask "${UMASK}"
fi

find /go -type d -print0 | xargs -0 chmod 777
chmod -Rf a+rw /var/lib/photoprism /tmp/photoprism /go

if [[ ${UID} ]] && [[ ${GID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  groupadd -f -g "${GID}" "${GID}"
  usermod -o -u "${UID}" -g "${GID}" photoprism

  if [[ -z ${PHOTOPRISM_DISABLE_CHOWN} ]] ; then
    echo "entrypoint: updating storage permissions"
    chown -Rf photoprism /photoprism /var/lib/photoprism /tmp/photoprism /go
  fi

  echo "entrypoint: uid ${UID}:${GID}"
  echo "entrypoint: starting ${@}"

  gosu "${UID}:${GID}" "$@" &
elif [[ ${UID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  usermod -o -u "${UID}" photoprism

  if [[ -z ${PHOTOPRISM_DISABLE_CHOWN} ]] ; then
    echo "entrypoint: updating storage permissions"
    chown -Rf photoprism /photoprism /var/lib/photoprism /tmp/photoprism /go
  fi

  echo "entrypoint: uid ${UID}"
  echo "entrypoint: starting ${@}"

  gosu "${UID}" "$@" &
else
  echo "entrypoint: uid $(id -u)"
  echo "entrypoint: starting ${@}"

  "$@" &
fi

PID=$!

trap "kill $PID" INT TERM
wait