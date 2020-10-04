#!/usr/bin/env bash

if [[ ${UMASK} ]]; then
  umask ${UMASK}
fi

if [[ ${UID} ]] && [[ ${GID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  usermod -u ${UID} photoprism
  usermod -g ${GID} photoprism
  exec gosu ${UID}:${GID} "$@"
elif [[ ${UID} ]] && [[ ${UID} != "0" ]] && [[ $(id -u) = "0" ]]; then
  usermod -u ${UID} photoprism
  exec gosu ${UID} "$@"
else
  exec "$@"
fi