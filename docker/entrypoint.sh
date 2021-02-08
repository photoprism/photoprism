#!/usr/bin/env bash

if [[ ${UMASK} ]]; then
  echo "entrypoint: umask ${UMASK}"
  umask "${UMASK}"
fi

if [[ ${UID} ]] && [[ ${GID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  groupadd -f -g "${GID}" "${GID}"
  usermod -o -u "${UID}" -g "${GID}" photoprism

  if [[ -z ${PHOTOPRISM_DISABLE_CHOWN} ]] ; then
    echo "entrypoint: updating storage permissions"
    chown -Rf photoprism /photoprism/storage /photoprism/import /photoprism/assets /var/lib/photoprism /tmp/photoprism
  fi

  echo "entrypoint: uid ${UID}:${GID}"
  echo "entrypoint: starting ${@}"

  gosu "${UID}:${GID}" "$@" &
elif [[ ${UID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  usermod -o -u "${UID}" photoprism

  if [[ -z ${PHOTOPRISM_DISABLE_CHOWN} ]] ; then
    echo "entrypoint: updating storage permissions"
    chown -Rf photoprism /photoprism/storage /photoprism/import /photoprism/assets /var/lib/photoprism /tmp/photoprism
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