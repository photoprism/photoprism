#!/usr/bin/env bash

if [[ $(id -u) == "0" ]]; then
  echo "develop: base image started as root"

  if [ -e /root/.init ]; then
    echo "develop: initialized"
  elif [[ ${PHOTOPRISM_INIT} ]]; then
    for target in $PHOTOPRISM_INIT; do
      echo "init ${target}..."
      make -f /go/src/github.com/photoprism/photoprism/scripts/dist/Makefile "${target}"
    done
    echo 1 >/root/.init
  fi
else
  echo "develop: base image started as uid $(id -u)"
fi

re='^[0-9]+$'

# check for legacy umask env variables
if [[ -z ${PHOTOPRISM_UMASK} ]] && [[ ${UMASK} =~ $re ]]; then
  PHOTOPRISM_UMASK=${UMASK}
  echo "WARNING: UMASK without PHOTOPRISM_ prefix is deprecated, use PHOTOPRISM_UMASK: \"${PHOTOPRISM_UMASK}\" instead"
fi

# set file permission mask
if [[ ${PHOTOPRISM_UMASK} =~ $re ]]; then
  echo "develop: umask ${PHOTOPRISM_UMASK}"
  umask "${PHOTOPRISM_UMASK}"
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
      echo "develop: set PHOTOPRISM_DISABLE_CHOWN: \"true\" to disable storage permission updates"
      echo "develop: updating storage permissions..."
      chown -Rf "${PHOTOPRISM_UID}:${PHOTOPRISM_GID}" /photoprism /tmp/photoprism /go
    fi

    echo "develop: running as uid ${PHOTOPRISM_UID}:${PHOTOPRISM_GID}"
    echo "${@}"

    gosu "${PHOTOPRISM_UID}:${PHOTOPRISM_GID}" "$@" &
  elif [[ ${PHOTOPRISM_UID} =~ $re ]] && [[ ${PHOTOPRISM_UID} != "0" ]]; then
    # user ID only
    useradd -o -u "${PHOTOPRISM_UID}" -g 1000 -d /photoprism "user_${PHOTOPRISM_UID}" 2>/dev/null
    usermod -g 1000 "user_${PHOTOPRISM_UID}" 2>/dev/null

    if [[ -z ${PHOTOPRISM_DISABLE_CHOWN} ]]; then
      echo "develop: set PHOTOPRISM_DISABLE_CHOWN: \"true\" to disable storage permission updates"
      echo "develop: updating storage permissions..."
      chown -Rf "${PHOTOPRISM_UID}" /photoprism /var/lib/photoprism /tmp/photoprism /go
    fi

    echo "develop: running as uid ${PHOTOPRISM_UID}"
    echo "${@}"

    gosu "${PHOTOPRISM_UID}" "$@" &
  else
    # run as root
    echo "develop: running as root"
    echo "${@}"

    "$@" &
  fi
else
  # running as user
  echo "develop: running as uid $(id -u)"
  echo "${@}"

  "$@" &
fi

PID=$!

trap "kill $PID" INT TERM
wait
