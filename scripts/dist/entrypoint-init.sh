#!/usr/bin/env bash

# INITIALIZES CONTAINER PACKAGES AND PERMISSIONS
export PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts"

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

# regular expressions
re='^[0-9]+$'

# detect environment
case $DOCKER_ENV in
  prod)
    export PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:/opt/photoprism/bin";
    INIT_SCRIPTS="/scripts"
    CHOWN_DIRS=("/photoprism/storage")
    CHMOD_DIRS=("/photoprism/storage")
    ;;

  develop)
    export PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:/usr/local/go/bin:/go/bin:/opt/photoprism/bin";
    INIT_SCRIPTS="/scripts"
    CHOWN_DIRS=("/photoprism" "/opt/photoprism" "/go" "/tmp/photoprism")
    CHMOD_DIRS=("/opt/photoprism" "/tmp/photoprism")
    ;;

  *)
    echo "init: unsupported environment $DOCKER_ENV";
    exit
    ;;
esac

if [[ ${PHOTOPRISM_UID} =~ $re ]] && [[ ${PHOTOPRISM_UID} != "0" ]]; then
  # Create user account if it does not exist yet (required by /usr/bin/setpriv).
  getent passwd "${PHOTOPRISM_UID}" > /dev/null
  if [ $? -eq 2 ] ; then
    userdel -r -f "user-${PHOTOPRISM_UID}" >/dev/null 2>&1
    groupdel -f "group-${PHOTOPRISM_UID}" >/dev/null 2>&1
    groupadd -f -g "${PHOTOPRISM_UID}" "group-${PHOTOPRISM_UID}"
    useradd -u "${PHOTOPRISM_UID}" -g "${PHOTOPRISM_UID}" -G photoprism,www-data,video,davfs2,renderd,render,ssl-cert,videodriver -s /bin/bash -m -d "/home/user-${PHOTOPRISM_UID}" "user-${PHOTOPRISM_UID}" 2>/dev/null
    echo "init: account with the user id ${PHOTOPRISM_UID} has been created"
  else
    echo "init: account with the user id ${PHOTOPRISM_UID} already exists"
  fi

  if [[ ${PHOTOPRISM_GID} =~ $re ]] && [[ ${PHOTOPRISM_GID} != "0" ]]; then
    CHOWN="${PHOTOPRISM_UID}:${PHOTOPRISM_GID}"
  else
    CHOWN="${PHOTOPRISM_UID}"
  fi

  if [[ -z ${PHOTOPRISM_DISABLE_CHOWN} ]] || [[ ${PHOTOPRISM_DISABLE_CHOWN} == "false" ]]; then
    echo "init: updating filesystem permissions"
    echo "PHOTOPRISM_DISABLE_CHOWN=\"true\" disables permission updates"
    chown --preserve-root --silent -R "${CHOWN}" "${CHOWN_DIRS[@]}"
    chmod --preserve-root --silent -R u+rwX "${CHMOD_DIRS[@]}"
  fi
fi

# do nothing if PHOTOPRISM_INIT was not set
if [[ -z ${PHOTOPRISM_INIT} ]]; then
  if [[ ${PHOTOPRISM_DEFAULT_TLS} = "true" ]]; then
    make --no-print-directory -C "$INIT_SCRIPTS" "https"
  fi
  exit
fi

INIT_LOCK="/scripts/.init-lock"

# execute targets via /usr/bin/make
if [[ ! -e ${INIT_LOCK} ]]; then
  for INIT_TARGET in $PHOTOPRISM_INIT; do
    echo "init: $INIT_TARGET"
    make --no-print-directory -C "$INIT_SCRIPTS" "$INIT_TARGET"
  done

  echo 1 >${INIT_LOCK}
fi
