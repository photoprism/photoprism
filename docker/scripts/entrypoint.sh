#!/usr/bin/env bash

if [[ ${UMASK} ]]; then
  echo "umask ${UMASK}"
  umask "${UMASK}"
fi

if [[ ${UID} ]] && [[ ${GID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  groupadd -f -g "${GID}" "${GID}"
  usermod -o -u "${UID}" -g "${GID}" photoprism

  if [[ -z ${PHOTOPRISM_DISABLE_CHOWN} ]] ; then
    echo "updating storage permissions..."
    chown -Rf photoprism /photoprism/storage /photoprism/import /photoprism/assets /var/lib/photoprism /tmp/photoprism
  fi

  echo "running as uid ${UID}:${GID}"
  echo "${@}"

  gosu "${UID}:${GID}" "$@" &
elif [[ ${UID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  usermod -o -u "${UID}" photoprism

  if [[ -z ${PHOTOPRISM_DISABLE_CHOWN} ]] ; then
    echo "updating storage permissions..."
    chown -Rf photoprism /photoprism/storage /photoprism/import /photoprism/assets /var/lib/photoprism /tmp/photoprism
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