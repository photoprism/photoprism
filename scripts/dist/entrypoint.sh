#!/usr/bin/env bash

# regular expressions
re='^[0-9]+$'

# set env defaults
export PHOTOPRISM_ARCH=${PHOTOPRISM_ARCH:-arch}
export DOCKER_ENV=${DOCKER_ENV:-unknown}
export DOCKER_TAG=${DOCKER_TAG:-unknown}

# detect environment
case $DOCKER_ENV in
  prod)
    INIT_SCRIPT="/scripts/entrypoint-init.sh"
    ;;

  develop)
    INIT_SCRIPT="/go/src/github.com/photoprism/photoprism/scripts/dist/entrypoint-init.sh"
    ;;

  *)
    INIT_SCRIPT=""
    echo "unknown environment \"$DOCKER_ENV\"";
    ;;
esac

# normalize user and group ID environment variables
if [[ -z ${PHOTOPRISM_UID} ]]; then
  if [[ ${UID} =~ $re ]] && [[ ${UID} != "0" ]]; then
    export PHOTOPRISM_UID=${UID}
  elif [[ ${PUID} =~ $re ]] && [[ ${PUID} != "0" ]]; then
    export PHOTOPRISM_UID=${PUID}
  fi
  if [[ -z ${PHOTOPRISM_GID} ]]; then
    if [[ ${GID} =~ $re ]] && [[ ${GID} != "0" ]]; then
      export PHOTOPRISM_GID=${GID}
    elif [[ ${PGID} =~ $re ]] && [[ ${PGID} != "0" ]]; then
      export PHOTOPRISM_GID=${PGID}
    fi
  fi
fi

# docker image info
DOCKER_IMAGE="$PHOTOPRISM_ARCH-$DOCKER_ENV/$DOCKER_TAG"

# initialize container packages and permissions
if [[ -f "${INIT_SCRIPT}" ]]; then
  if [[ $(id -u) == "0" ]]; then
    echo "init $DOCKER_IMAGE as root"
    bash -c "${INIT_SCRIPT}"
  else
    echo "init $DOCKER_IMAGE as uid $(id -u)"
    sudo -E "${INIT_SCRIPT}"
  fi
else
  echo "started $DOCKER_IMAGE as uid $(id -u)"
fi

# set explicit home directory
export HOME="/photoprism"

# check for alternate umask variable
if [[ -z ${PHOTOPRISM_UMASK} ]] && [[ ${UMASK} =~ $re ]] && [[ ${#UMASK} == 4 ]]; then
  export PHOTOPRISM_UMASK=${UMASK}
fi

# set file-creation mode (umask)
if [[ ${PHOTOPRISM_UMASK} =~ $re ]] && [[ ${#PHOTOPRISM_UMASK} == 4 ]]; then
  umask "${PHOTOPRISM_UMASK}"
else
  umask 0002
fi

# display additional container info for troubleshooting
echo "umask: \"$(umask)\" ($(umask -S))"
echo "home-directory: ${HOME}"
echo "storage-path: ${PHOTOPRISM_STORAGE_PATH}"
echo "originals-path: ${PHOTOPRISM_ORIGINALS_PATH}"
echo "import-path: ${PHOTOPRISM_IMPORT_PATH}"
echo "assets-path: ${PHOTOPRISM_ASSETS_PATH}"

# change to another user and group on request
if [[ $(id -u) == "0" ]] && [[ ${PHOTOPRISM_UID} =~ $re ]] && [[ ${PHOTOPRISM_UID} != "0" ]]; then
  # check uid and gid env variables
  if [[ ${PHOTOPRISM_GID} =~ $re ]] && [[ ${PHOTOPRISM_GID} != "0" ]]; then
    echo "switching to uid ${PHOTOPRISM_UID}:${PHOTOPRISM_GID}"
    echo "${@}"

    # run command as uid:gid
    ([[ ${DOCKER_ENV} != "prod" ]] || gosu "${PHOTOPRISM_UID}:${PHOTOPRISM_GID}" "/scripts/audit.sh") \
     && gosu "${PHOTOPRISM_UID}:${PHOTOPRISM_GID}" "$@" &
  else
    echo "switching to uid ${PHOTOPRISM_UID}"
    echo "${@}"

    # run command as uid
    ([[ ${DOCKER_ENV} != "prod" ]] || gosu "${PHOTOPRISM_UID}" "/scripts/audit.sh") \
     && gosu "${PHOTOPRISM_UID}" "$@" &
  fi
else
  echo "running as uid $(id -u)"
  echo "${@}"

  # run command
  ([[ ${DOCKER_ENV} != "prod" ]] || "/scripts/audit.sh") \
   && "$@" &
fi

PID=$!

trap "kill $PID" INT TERM
wait
