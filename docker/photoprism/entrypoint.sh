#!/usr/bin/env bash

if [[ $(id -u) == "0" ]]; then
  echo "started as root"

  if [ -e /root/.init ]; then
    echo "initialized"
  elif [[ ${PHOTOPRISM_INIT} ]]; then
    for target in $PHOTOPRISM_INIT; do
      echo "init ${target}..."
      make -f /root/Makefile "${target}"
    done
    echo 1 >/root/.init
  fi
fi

re='^[0-9]+$'

# Legacy umask env variable in use?
if [[ -z ${PHOTOPRISM_UMASK} ]] && [[ ${UMASK} =~ $re ]]; then
  PHOTOPRISM_UMASK=${UMASK}
  echo "WARNING: UMASK without PHOTOPRISM_ prefix is deprecated, use PHOTOPRISM_UMASK: \"${PHOTOPRISM_UMASK}\" instead"
fi

# Set file permission mask
if [[ ${PHOTOPRISM_UMASK} =~ $re ]]; then
  echo "umask ${PHOTOPRISM_UMASK}"
  umask "${PHOTOPRISM_UMASK}"
fi

# Script runs as root?
if [[ $(id -u) == "0" ]]; then
  # Alternate USER ID environment variable check
  if [[ -z ${PHOTOPRISM_UID} ]]; then
    if [[ ${UID} =~ $re ]] && [[ ${UID} != "0" ]]; then
      PHOTOPRISM_UID=${UID}
    elif [[ ${PUID} =~ $re ]] && [[ ${PUID} != "0" ]]; then
      PHOTOPRISM_UID=${PUID}
    fi
  fi

  # Alternate GROUP ID environment variable check
  if [[ -z ${PHOTOPRISM_GID} ]]; then
    if [[ ${GID} =~ $re ]] && [[ ${GID} != "0" ]]; then
      PHOTOPRISM_GID=${GID}
    elif [[ ${PGID} =~ $re ]] && [[ ${PGID} != "0" ]]; then
      PHOTOPRISM_GID=${PGID}
    fi
  fi

  # User and group ID set?
  if [[ ${PHOTOPRISM_UID} =~ $re ]] && [[ ${PHOTOPRISM_UID} != "0" ]] && [[ ${PHOTOPRISM_GID} =~ $re ]] && [[ ${PHOTOPRISM_GID} != "0" ]]; then
    groupadd -g "${PHOTOPRISM_GID}" "group_${PHOTOPRISM_GID}" 2>/dev/null
    useradd -o -u "${PHOTOPRISM_UID}" -g "${PHOTOPRISM_GID}" -d /photoprism "user_${PHOTOPRISM_UID}" 2>/dev/null
    usermod -g "${PHOTOPRISM_GID}" "user_${PHOTOPRISM_UID}" 2>/dev/null

    if [[ -z ${PHOTOPRISM_DISABLE_CHOWN} ]]; then
      echo "set PHOTOPRISM_DISABLE_CHOWN: \"true\" to disable storage permission updates"
      echo "updating storage permissions..."
      chown -Rf "${PHOTOPRISM_UID}:${PHOTOPRISM_GID}" /photoprism/storage /photoprism/import /photoprism/assets /var/lib/photoprism /tmp/photoprism
    fi

    echo "running as uid ${PHOTOPRISM_UID}:${PHOTOPRISM_GID}"
    echo "${@}"

    gosu "${PHOTOPRISM_UID}:${PHOTOPRISM_GID}" "$@" &
  elif [[ ${PHOTOPRISM_UID} =~ $re ]] && [[ ${PHOTOPRISM_UID} != "0" ]]; then
    # User ID only
    useradd -o -u "${PHOTOPRISM_UID}" -g 1000 -d /photoprism "user_${PHOTOPRISM_UID}" 2>/dev/null
    usermod -g 1000 "user_${PHOTOPRISM_UID}" 2>/dev/null

    if [[ -z ${PHOTOPRISM_DISABLE_CHOWN} ]]; then
      echo "set PHOTOPRISM_DISABLE_CHOWN: \"true\" to disable storage permission updates"
      echo "updating storage permissions..."
      chown -Rf "${PHOTOPRISM_UID}" /photoprism/storage /photoprism/import /photoprism/assets /var/lib/photoprism /tmp/photoprism
    fi

    echo "running as uid ${PHOTOPRISM_UID}"
    echo "${@}"

    gosu "${PHOTOPRISM_UID}" "$@" &
  else
    # No user or group ID set via end variable
    echo "running as root"
    echo "${@}"

    "$@" &
  fi
else
  # Running as root
  echo "running as uid $(id -u)"
  echo "${@}"

  "$@" &
fi

PID=$!

trap "kill $PID" INT TERM
wait
