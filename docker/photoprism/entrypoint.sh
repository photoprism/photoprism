#!/usr/bin/env bash

if [[ ${UMASK} ]]; then
  umask ${UMASK}
fi

if [[ ${UID} ]] && [[ ${GID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  usermod -u ${UID} photoprism
  usermod -g ${GID} photoprism
  gosu ${UID}:${GID} "$@" &
elif [[ ${UID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  usermod -u ${UID} photoprism
  gosu ${UID} "$@" &
else
  "$@" &
fi

PHOTOPRISM_PID=$!

trap "kill $PHOTOPRISM_PID" INT TERM
wait