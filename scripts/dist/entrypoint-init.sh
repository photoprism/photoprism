#!/usr/bin/env bash

# INITIALIZES CONTAINER PACKAGES AND PERMISSIONS

# abort if not executed as root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

# regular expressions
re='^[0-9]+$'

# detect environment
case $DOCKER_ENV in
  prod)
    INIT_SCRIPTS="/scripts"
    CHOWN_DIRS=("/photoprism" "/opt/photoprism")
    CHMOD_DIRS=("/opt/photoprism")
    ;;

  develop)
    INIT_SCRIPTS="/go/src/github.com/photoprism/photoprism/scripts/dist"
    CHOWN_DIRS=("/go /photoprism" "/opt/photoprism" "/tmp/photoprism")
    CHMOD_DIRS=("/photoprism" "/opt/photoprism" "/tmp/photoprism")
    ;;

  *)
    echo "unsupported init environment \"$DOCKER_ENV\"";
    exit
    ;;
esac

if [[ ${PHOTOPRISM_UID} =~ $re ]] && [[ ${PHOTOPRISM_UID} != "0" ]]; then
  if [[ ${PHOTOPRISM_GID} =~ $re ]] && [[ ${PHOTOPRISM_GID} != "0" ]]; then
    groupadd -g "${PHOTOPRISM_GID}" "group_${PHOTOPRISM_GID}" 2>/dev/null
    useradd -o -u "${PHOTOPRISM_UID}" -g "${PHOTOPRISM_GID}" -d "/photoprism" "user_${PHOTOPRISM_UID}" 2>/dev/null
    usermod -g "${PHOTOPRISM_GID}" "user_${PHOTOPRISM_UID}" 2>/dev/null

    if [[ -z ${PHOTOPRISM_DISABLE_CHOWN} ]]; then
      echo "updating filesystem permissions..."
      echo "PHOTOPRISM_DISABLE_CHOWN: \"true\" disables filesystem permission updates"
      chown --preserve-root -Rcf "${PHOTOPRISM_UID}:${PHOTOPRISM_GID}" "${CHOWN_DIRS[@]}"
      chmod --preserve-root -Rcf u+rwX "${CHMOD_DIRS[@]}"
    fi
  else
    useradd -o -u "${PHOTOPRISM_UID}" -g 1000 -d "/photoprism" "user_${PHOTOPRISM_UID}" 2>/dev/null
    usermod -g 1000 "user_${PHOTOPRISM_UID}" 2>/dev/null

    if [[ -z ${PHOTOPRISM_DISABLE_CHOWN} ]]; then
      echo "updating filesystem permissions..."
      echo "PHOTOPRISM_DISABLE_CHOWN: \"true\" disables filesystem permission updates"
      chown --preserve-root -Rcf "${PHOTOPRISM_UID}" "${CHOWN_DIRS[@]}"
      chmod --preserve-root -Rcf u+rwX "${CHMOD_DIRS[@]}"
    fi
  fi
fi

# do nothing if PHOTOPRISM_INIT was not set
if [[ -z ${PHOTOPRISM_INIT} ]]; then
  exit
fi

INIT_LOCK="/scripts/.init-lock"

# execute targets via make
if [[ ! -e ${INIT_LOCK} ]]; then
  for INIT_TARGET in $PHOTOPRISM_INIT; do
    echo "init $INIT_TARGET..."
    make -C "$INIT_SCRIPTS" "$INIT_TARGET"
  done

  echo 1 >${INIT_LOCK}
fi
