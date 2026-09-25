#!/usr/bin/env bash

# Installs the sudoers drop-in that lets the entrypoint run the init script as root, and records
# the image identity the scripts read at startup.
# Pass --develop in development images, where all targets and scripts may be run with sudo.
# Pass --image <name> to name the published image (develop | ce | plus | pro | portal).
# Run this as a file in the scripts directory, as the rules name the init script beside it.

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

set -e

# Resolve the scripts directory from this file, so the rules and the environment file refer
# to the scripts beside it wherever they are installed.
SCRIPTS_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" 2>/dev/null && pwd -P)

if [[ ! -f ${SCRIPTS_DIR}/entrypoint-init.sh ]]; then
  echo "Error: entrypoint-init.sh not found in '${SCRIPTS_DIR}'." 1>&2
  exit 1
fi

# require_root_owned fails unless the given path is owned by root and writable by nobody
# else. It rejects a symlink outright, since the mode that matters would be the target's.
require_root_owned() {
  local target=$1 owner mode

  if [[ -L ${target} ]]; then
    echo "Error: '${target}' must not be a symlink." 1>&2
    return 1
  fi

  owner=$(stat -c '%u' "${target}")
  mode=$(stat -c '%a' "${target}")

  if [[ ! ${owner} =~ ^[0-9]+$ ]] || [[ ! ${mode} =~ ^[0-7]+$ ]]; then
    echo "Error: cannot read the ownership of '${target}'." 1>&2
    return 1
  fi

  if [[ ${owner} != "0" ]]; then
    echo "Error: '${target}' must be owned by root." 1>&2
    return 1
  fi

  if (( 0${mode} & 0022 )); then
    echo "Error: '${target}' must not be writable by group or others." 1>&2
    return 1
  fi
}

# The command path is matched as a literal, so restrict the name to characters that cannot
# be read as a separator or as a pattern.
if [[ ${SCRIPTS_DIR} =~ [^A-Za-z0-9/_.-] ]]; then
  echo "Error: '${SCRIPTS_DIR}' contains a character that may not appear in a rule." 1>&2
  exit 1
fi

# The rules name this directory and the init script in it, so neither may be something
# another account can replace.
require_root_owned "${SCRIPTS_DIR}"
require_root_owned "${SCRIPTS_DIR}/entrypoint-init.sh"

INIT_SCRIPT="${SCRIPTS_DIR}/entrypoint-init.sh"
SUDOERS_FILE="/etc/sudoers.d/init"

# Records what the image is, so that the startup scripts do not have to take it from the caller.
# Written here because this is where the two image families differ.
IMAGE_ENV_FILE="${SCRIPTS_DIR}/.env"
DOCKER_ENV_NAME="prod"
DOCKER_IMG_NAME="develop"
DEVELOP=0

while [[ $# -gt 0 ]]; do
  case $1 in
    --develop) DEVELOP=1; shift ;;
    --image)
      if [[ -z $2 ]]; then
        echo "Error: --image needs a name." 1>&2
        exit 1
      fi
      DOCKER_IMG_NAME=$2
      shift 2
      ;;
    *)
      echo "Error: unknown argument '$1'." 1>&2
      exit 1
      ;;
  esac
done

# The names the startup scripts branch on, rejected here rather than at every reader.
case ${DOCKER_IMG_NAME} in
  develop | ce | plus | pro | portal) ;;
  *) echo "Error: unknown image name '${DOCKER_IMG_NAME}'." 1>&2; exit 1 ;;
esac

# Variables the init script and the targets it runs read from the environment. Only these are
# passed on, so that the caller cannot supply the ones that change how a command interprets
# its input, such as the variables GNU make accepts options and additional makefiles through.
# The proxy variables are included because the init targets download packages and models,
# and nothing else supplies them.
# DOCKER_ENV is inert on this list: every reader clears it before consulting the recorded
# file, so a caller's value is never read. It is kept for one rebuild cycle, and the probe
# list in the container init matrix has to be narrowed with it.
INIT_ENV="DOCKER_ENV TF_VERSION ONNX_GPU ONNX_VERSION \
http_proxy https_proxy ftp_proxy all_proxy no_proxy HTTP_PROXY HTTPS_PROXY FTP_PROXY ALL_PROXY NO_PROXY \
PHOTOPRISM_*"

SUDOERS_TMP=$(mktemp)

if [[ ${DEVELOP} == 1 ]]; then
  # Development images allow every target and script to be run with sudo.
  SUDOERS_FILE="/etc/sudoers.d/all"
  DOCKER_ENV_NAME="develop"
  printf '%s\n' \
    "Defaults env_keep += \"${INIT_ENV}\"" \
    'ALL ALL=(ALL) NOPASSWD:SETENV: ALL' \
    > "${SUDOERS_TMP}"
else
  printf '%s\n' \
    "Cmnd_Alias PHOTOPRISM_INIT_CMND = ${INIT_SCRIPT} \"\"" \
    "Defaults!PHOTOPRISM_INIT_CMND env_keep += \"${INIT_ENV}\"" \
    'ALL ALL=(root) NOPASSWD: PHOTOPRISM_INIT_CMND' \
    > "${SUDOERS_TMP}"
fi

# Check the file before it is in place, as sudo refuses to run anything at all while a
# malformed drop-in is present.
visudo -c -f "${SUDOERS_TMP}"

install -m 0440 -o root -g root "${SUDOERS_TMP}" "${SUDOERS_FILE}"
rm -f "${SUDOERS_TMP}"

# The readers follow this path at every start, so it must not be something else in disguise.
if [[ -L ${IMAGE_ENV_FILE} ]]; then
  echo "Error: '${IMAGE_ENV_FILE}' must not be a symlink." 1>&2
  exit 1
fi

printf '%s\n' \
  "DOCKER_ENV=${DOCKER_ENV_NAME}" \
  "DOCKER_IMG=${DOCKER_IMG_NAME}" \
  > "${IMAGE_ENV_FILE}"
chown root:root "${IMAGE_ENV_FILE}"
chmod 0444 "${IMAGE_ENV_FILE}"

echo "✅ Installed ${SUDOERS_FILE} for ${INIT_SCRIPT} (${DOCKER_ENV_NAME}, ${DOCKER_IMG_NAME})."
