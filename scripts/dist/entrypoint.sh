#!/usr/bin/env bash

DOCKER_ARCH=${DOCKER_ARCH:-arch}
DOCKER_ENV=${DOCKER_ENV:-unknown}

echo "image: $DOCKER_ARCH-$DOCKER_ENV $DOCKER_TAG"

if [[ $(id -u) == "0" ]]; then
  echo "started as root"

  if [[ ! -e /opt/photoprism/.init ]] && [[ ${PHOTOPRISM_INIT} ]]; then
    for target in $PHOTOPRISM_INIT; do
      echo "init ${target}..."
      make -f /opt/photoprism/scripts/Makefile "${target}"
    done
    echo 1 >/opt/photoprism/.init
  fi
else
  echo "started as uid $(id -u)"
fi

STORAGE_PATH=${PHOTOPRISM_STORAGE_PATH:-/photoprism/storage}

re='^[0-9]+$'

# check for alternate umask variable
if [[ -z ${PHOTOPRISM_UMASK} ]] && [[ ${UMASK} =~ $re ]] && [[ ${#UMASK} == 4 ]]; then
  PHOTOPRISM_UMASK=${UMASK}
fi

# set file-creation mode (umask)
if [[ ${PHOTOPRISM_UMASK} =~ $re ]] && [[ ${#PHOTOPRISM_UMASK} == 4 ]]; then
  umask "${PHOTOPRISM_UMASK}"
  echo "custom file-creation mode ($(umask -p)): $(umask -S)"
else
  umask 0002
  echo "default file-creation mode ($(umask -p)): $(umask -S)"
fi

# script must run as root to perform changes
if [[ $(id -u) == "0" ]]; then
  # check for alternate user ID env variables
  if [[ -z ${PHOTOPRISM_UID} ]]; then
    if [[ ${UID} =~ $re ]] && [[ ${UID} != "0" ]]; then
      PHOTOPRISM_UID=${UID}
    elif [[ ${PUID} =~ $re ]] && [[ ${PUID} != "0" ]]; then
      PHOTOPRISM_UID=${PUID}
    fi
  fi

  # check for alternate group ID env variables
  if [[ -z ${PHOTOPRISM_GID} ]]; then
    if [[ ${GID} =~ $re ]] && [[ ${GID} != "0" ]]; then
      PHOTOPRISM_GID=${GID}
    elif [[ ${PGID} =~ $re ]] && [[ ${PGID} != "0" ]]; then
      PHOTOPRISM_GID=${PGID}
    fi
  fi

  # create missing user/group if needed
  if [[ ${PHOTOPRISM_UID} =~ $re ]] && [[ ${PHOTOPRISM_UID} != "0" ]] && [[ ${PHOTOPRISM_GID} =~ $re ]] && [[ ${PHOTOPRISM_GID} != "0" ]]; then
    groupadd -g "${PHOTOPRISM_GID}" "group_${PHOTOPRISM_GID}" 2>/dev/null
    useradd -o -u "${PHOTOPRISM_UID}" -g "${PHOTOPRISM_GID}" -d /photoprism "user_${PHOTOPRISM_UID}" 2>/dev/null
    usermod -g "${PHOTOPRISM_GID}" "user_${PHOTOPRISM_UID}" 2>/dev/null

    if [[ -z ${PHOTOPRISM_DISABLE_CHOWN} ]]; then
      echo "set PHOTOPRISM_DISABLE_CHOWN: \"true\" to disable storage permission updates"
      echo "updating storage permissions..."
      chown -Rf "${PHOTOPRISM_UID}:${PHOTOPRISM_GID}" "${STORAGE_PATH}" /photoprism/import /var/lib/photoprism
    fi

    echo "running as uid ${PHOTOPRISM_UID}:${PHOTOPRISM_GID}"
    echo "${@}"

    gosu "${PHOTOPRISM_UID}:${PHOTOPRISM_GID}" audit.sh && gosu "${PHOTOPRISM_UID}:${PHOTOPRISM_GID}" "$@" &
  elif [[ ${PHOTOPRISM_UID} =~ $re ]] && [[ ${PHOTOPRISM_UID} != "0" ]]; then
    # user ID only
    useradd -o -u "${PHOTOPRISM_UID}" -g 1000 -d /photoprism "user_${PHOTOPRISM_UID}" 2>/dev/null
    usermod -g 1000 "user_${PHOTOPRISM_UID}" 2>/dev/null

    if [[ -z ${PHOTOPRISM_DISABLE_CHOWN} ]]; then
      echo "set PHOTOPRISM_DISABLE_CHOWN: \"true\" to disable storage permission updates"
      echo "updating storage permissions..."
      chown -Rf "${PHOTOPRISM_UID}" "${STORAGE_PATH}" /photoprism/import /var/lib/photoprism
    fi

    echo "running as uid ${PHOTOPRISM_UID}"
    echo "${@}"

    gosu "${PHOTOPRISM_UID}" audit.sh && gosu "${PHOTOPRISM_UID}" "$@" &
  else
    # no user or group ID set via end variable
    echo "running as root"
    echo "${@}"

    audit.sh && "$@" &
  fi
else

  # running as root
  echo "running as uid $(id -u)"
  echo "${@}"

   audit.sh && "$@" &
fi

PID=$!

trap "kill $PID" INT TERM
wait
