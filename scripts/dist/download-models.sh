#!/usr/bin/env bash

# Downloads and installs machine-learning models from a checksum-pinned registry.
#
# Models with their own installation requirements keep dedicated scripts instead.
#
# This lives under scripts/dist so it ships to /scripts in the production images, where
# it is how an operator installs a model that is not bundled.

set -euo pipefail

MODELS_PATH=${MODELS_PATH:-"${PHOTOPRISM_ASSETS_PATH:-assets}/models"}
TMP_PATH=${TMP_PATH:-"/tmp/photoprism"}
BACKUP_PATH=${BACKUP_PATH:-"${PHOTOPRISM_STORAGE_PATH:-storage}/backup"}
TODAY=$(date -u +%Y%m%d)
STAMP=$(date -u +%Y%m%d-%H%M%S)

MIRROR_URL="https://dl.photoprism.app"
ONNX_URL="${MIRROR_URL}/onnx/models"
TENSORFLOW_URL="${MIRROR_URL}/tensorflow"
OPENCV_ZOO_URL="https://media.githubusercontent.com/media/opencv/opencv_zoo/main/models"

# Registry fields: name|url|fallback|sha256|type|dir|file
#
# Type "zip" extracts an archive containing the model directory, and records the archive
# checksum in version.txt because a directory cannot be hashed cheaply. Type "file"
# installs a single file, whose own checksum is what the up-to-date check compares.
#
# The fallback is the publisher's own copy, used when the mirror is unreachable. It is
# empty for the models we export ourselves, which have no upstream equivalent.
#
# Several "file" entries may share one directory, as the two yunet exports do; each keeps
# its own line in version.txt.
MODELS="\
facenet|${TENSORFLOW_URL}/facenet.zip||bf9ae0945d2ac53ac3db27082162d2b9dda5ba2c564c0e4c4f539f31f8b670af|zip|facenet|
nasnet|${TENSORFLOW_URL}/nasnet.zip||a0e1ad8d5a5a0ff9efc4b3ed89898bf008563ee36cacd0c804a384f8fc661588|zip|nasnet|
nsfw|${TENSORFLOW_URL}/nsfw.zip||eb5e5d22e37961c3192a4757efff883f77bc989c0efceabb1395e0959d966f14|zip|nsfw|
sface|${ONNX_URL}/face_recognition_sface_2021dec.onnx|${OPENCV_ZOO_URL}/face_recognition_sface/face_recognition_sface_2021dec.onnx|0ba9fbfa01b5270c96627c4ef784da859931e02f04419c829e83484087c34e79|file|sface|face_recognition_sface_2021dec.onnx
auraface|${ONNX_URL}/auraface_v1_glintr100.onnx|https://huggingface.co/fal/AuraFace-v1/resolve/main/glintr100.onnx?download=true|a7933ea5330113b01c9b60351d8f4c33003f145d8470ac5f0e52ee2effe25c60|file|auraface|auraface_v1_glintr100.onnx
yunet|${ONNX_URL}/face_detection_yunet_2026may.onnx|${OPENCV_ZOO_URL}/face_detection_yunet/face_detection_yunet_2026may.onnx|ebafce4e3c118d6554634be5c27ab333b4c047a9a8c3faf1d7cf93101c22f0f0|file|yunet|face_detection_yunet_2026may.onnx
yunet-2023mar|${ONNX_URL}/face_detection_yunet_2023mar.onnx|${OPENCV_ZOO_URL}/face_detection_yunet/face_detection_yunet_2023mar.onnx|8f2383e4dd3cfbb4553ea8718107fc0423210dc964f9f4280604804ed2552fa4|file|yunet|face_detection_yunet_2023mar.onnx
centerface|${ONNX_URL}/centerface.onnx|https://raw.githubusercontent.com/Star-Clouds/CenterFace/master/models/onnx/centerface.onnx|77e394b51108381b4c4f7b4baf1c64ca9f4aba73e5e803b2636419578913b5fe|file|centerface|centerface.onnx"

FORCE=false
OVERRIDE_URL=""

# usage prints the command syntax and the models the registry knows.
usage() {
  cat <<EOF
Usage: $(basename "$0") [options] <model>...

Options:
  -l, --list        List the available models and exit.
  -f, --force       Reinstall even when the model is already up to date.
  -u, --url <url>   Download from this URL instead of the registry source.
                    Requires exactly one model; the checksum is still enforced.
  -h, --help        Show this help and exit.

Environment:
  MODELS_PATH       Install prefix (default "\$PHOTOPRISM_ASSETS_PATH/models").
  TMP_PATH          Download directory (default "/tmp/photoprism").
  BACKUP_PATH       Backup directory (default "\$PHOTOPRISM_STORAGE_PATH/backup").

EOF
  list_models
}

# list_models prints the registry as a table.
list_models() {
  echo "Available models:"
  local name type dir file
  while IFS='|' read -r name _ _ _ type dir file; do
    [[ -z "${name}" ]] && continue
    if [[ "${type}" == "zip" ]]; then
      printf '  %-14s %s\n' "${name}" "${MODELS_PATH}/${dir}"
    else
      printf '  %-14s %s\n' "${name}" "${MODELS_PATH}/${dir}/${file}"
    fi
  done <<<"${MODELS}"
}

# lookup returns the registry entry for a model name, or nothing when it is unknown.
lookup() {
  local want="$1" name rest
  while IFS='|' read -r name rest; do
    if [[ "${name}" == "${want}" ]]; then
      echo "${name}|${rest}"
      return 0
    fi
  done <<<"${MODELS}"
  return 1
}

# hash_file prints the SHA-256 checksum of a file, or nothing when it does not exist.
hash_file() {
  [[ -f "$1" ]] || return 0

  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

# fetch downloads a URL to a destination path.
fetch() {
  local url="$1" dest="$2"

  # Our own host is behind a CDN, so a dated query defeats a stale cached copy that
  # would otherwise fail the checksum on every retry.
  if [[ "${url}" == "${MIRROR_URL}/"* ]]; then
    url="${url}?${TODAY}"
  fi

  if command -v curl >/dev/null 2>&1; then
    curl -fL --progress-bar --retry 3 --retry-delay 2 -o "${dest}" "${url}"
  elif command -v wget >/dev/null 2>&1; then
    wget --inet4-only "${url}" -O "${dest}"
  else
    echo "Error: neither curl nor wget is available." >&2
    return 1
  fi
}

# download_verified fetches a model and verifies its checksum, falling back to the
# publisher when the mirror fails. An existing verified download is reused as-is.
download_verified() {
  local url="$1" fallback="$2" sha256="$3" tmp="$4" actual

  if [[ "$(hash_file "${tmp}")" == "${sha256}" ]]; then
    return 0
  fi

  rm -f "${tmp}"

  echo "Downloading ${url}"

  if ! fetch "${url}" "${tmp}"; then
    if [[ -z "${fallback}" ]]; then
      echo "Error: download failed and no fallback is available." >&2
      return 1
    fi

    echo "Retrying with ${fallback}"

    if ! fetch "${fallback}" "${tmp}"; then
      echo "Error: download failed from both sources." >&2
      return 1
    fi
  fi

  actual="$(hash_file "${tmp}")"

  if [[ "${actual}" != "${sha256}" ]]; then
    echo "Error: checksum mismatch, refusing to install." >&2
    echo "  expected ${sha256}" >&2
    echo "  actual   ${actual:-none}" >&2
    rm -f "${tmp}"
    return 1
  fi
}

# backup_path moves an existing model file or directory aside.
#
# It refuses to replace an earlier backup: a retry after a failed install would otherwise
# overwrite the last good copy with whatever the failure left behind.
backup_path() {
  local source="$1" backup="$2"

  [[ -e "${source}" ]] || return 0

  if [[ -e "${backup}" ]]; then
    echo "Error: backup ${backup} already exists." >&2
    return 1
  fi

  echo "Backing up ${source} to ${backup}"

  mkdir -p "${BACKUP_PATH}" || return 1
  mv "${source}" "${backup}"
}

# record_version notes what was installed, replacing any earlier line for the same file
# so a directory holding several exports stays accurate.
record_version() {
  local version="$1" name="$2" sha256="$3" file="$4" existing=""

  if [[ -f "${version}" && -n "${file}" ]]; then
    existing="$(awk -v f="${file}" '$NF != f' "${version}")"
  fi

  {
    [[ -n "${existing}" ]] && echo "${existing}"
    echo "${name} ${TODAY} ${sha256} ${file}"
  } >"${version}.tmp" || return 1

  mv "${version}.tmp" "${version}"
}

# install_zip extracts into a staging directory and swaps it into place once the contents
# are known good, so a failed extraction can neither remove the installed model nor leave
# behind a version.txt recording something that was never installed.
install_zip() {
  local name="$1" sha256="$2" dir="$3" tmp="$4"
  local target="${MODELS_PATH}/${dir}"
  local staging="${MODELS_PATH}/.staging-${dir}-$$"

  rm -rf "${staging}"
  mkdir -p "${staging}" || return 1

  if ! unzip -q -o "${tmp}" -d "${staging}"; then
    echo "Error: cannot extract ${tmp}." >&2
    rm -rf "${staging}"
    return 1
  fi

  # The archive has to contain the directory the registry names, or the swap below would
  # install nothing where the application looks for it.
  if [[ ! -d "${staging}/${dir}" ]]; then
    echo "Error: ${name} archive does not contain \"${dir}\"." >&2
    rm -rf "${staging}"
    return 1
  fi

  if ! echo "${name} ${TODAY} ${sha256}" >"${staging}/${dir}/version.txt"; then
    rm -rf "${staging}"
    return 1
  fi

  if ! backup_path "${target}" "${BACKUP_PATH}/${dir}-${STAMP}"; then
    rm -rf "${staging}"
    return 1
  fi

  if ! mv "${staging}/${dir}" "${target}"; then
    echo "Error: cannot install ${name} in ${target}." >&2
    rm -rf "${staging}"
    return 1
  fi

  rm -rf "${staging}"
}

# install_file stages the download inside the target directory, confirms it survived the
# move intact, and only then swaps it over the installed copy.
install_file() {
  local name="$1" sha256="$2" dir="$3" file="$4" tmp="$5"
  local target="${MODELS_PATH}/${dir}"
  local staged="${target}/.staging-${file}-$$"

  mkdir -p "${target}" || return 1

  if ! mv "${tmp}" "${staged}"; then
    echo "Error: cannot stage ${file} in ${target}." >&2
    rm -f "${staged}"
    return 1
  fi

  # Moving onto another filesystem copies rather than renames, which a full disk can
  # truncate, so what matters is the checksum of the staged copy rather than the source.
  if [[ "$(hash_file "${staged}")" != "${sha256}" ]]; then
    echo "Error: ${file} did not survive the move to ${target}." >&2
    rm -f "${staged}"
    return 1
  fi

  if ! backup_path "${target}/${file}" "${BACKUP_PATH}/${dir}-${STAMP}-${file}"; then
    rm -f "${staged}"
    return 1
  fi

  if ! mv "${staged}" "${target}/${file}"; then
    echo "Error: cannot install ${name} in ${target}." >&2
    rm -f "${staged}"
    return 1
  fi

  record_version "${target}/version.txt" "${name}" "${sha256}" "${file}"
}

# install_model installs one registry entry and prints what it did.
install_model() {
  local entry name url fallback sha256 type dir file target tmp

  if ! entry="$(lookup "$1")"; then
    echo "Error: unknown model \"$1\"." >&2
    return 1
  fi

  IFS='|' read -r name url fallback sha256 type dir file <<<"${entry}"

  if [[ -n "${OVERRIDE_URL}" ]]; then
    url="${OVERRIDE_URL}"
    fallback=""
  fi

  target="${MODELS_PATH}/${dir}"

  if [[ "${FORCE}" != true ]] && up_to_date "${sha256}" "${type}" "${dir}" "${file}"; then
    echo "${name}: already up to date."
    return 0
  fi

  tmp="${TMP_PATH}/$(basename "${url%%\?*}")"

  download_verified "${url}" "${fallback}" "${sha256}" "${tmp}" || return 1

  # Each step is checked explicitly because errexit is suspended for the whole dynamic
  # extent of a function whose status the caller tests, so nothing here can abort on its
  # own and an unchecked failure would be reported as a successful install.
  if [[ "${type}" == "zip" ]]; then
    install_zip "${name}" "${sha256}" "${dir}" "${tmp}" || return 1
  else
    install_file "${name}" "${sha256}" "${dir}" "${file}" "${tmp}" || return 1
  fi

  echo "${name}: installed in ${target}."
}

# up_to_date reports whether the installed model already matches the registry checksum.
up_to_date() {
  local sha256="$1" type="$2" dir="$3" file="$4" version

  if [[ "${type}" == "file" ]]; then
    [[ "$(hash_file "${MODELS_PATH}/${dir}/${file}")" == "${sha256}" ]]
    return
  fi

  version="${MODELS_PATH}/${dir}/version.txt"

  [[ -f "${version}" ]] && grep -qF "${sha256}" "${version}"
}

ARGS=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    -l | --list)
      list_models
      exit 0
      ;;
    -f | --force)
      FORCE=true
      shift
      ;;
    -u | --url)
      if [[ $# -lt 2 ]]; then
        echo "Error: --url requires a value." >&2
        exit 1
      fi
      OVERRIDE_URL="$2"
      shift 2
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    -*)
      echo "Error: unknown option \"$1\"." >&2
      exit 1
      ;;
    *)
      ARGS+=("$1")
      shift
      ;;
  esac
done

# Downloading every model by default would pull hundreds of megabytes that most
# environments never use, so a bare invocation asks for names instead.
if [[ ${#ARGS[@]} -eq 0 ]]; then
  echo "Error: no model specified." >&2
  echo >&2
  usage >&2
  exit 1
fi

# A single URL cannot serve several different models.
if [[ -n "${OVERRIDE_URL}" && ${#ARGS[@]} -ne 1 ]]; then
  echo "Error: --url requires exactly one model." >&2
  exit 1
fi

if ! command -v sha256sum >/dev/null 2>&1 && ! command -v shasum >/dev/null 2>&1; then
  echo "Error: neither sha256sum nor shasum is available." >&2
  exit 1
fi

mkdir -p "${TMP_PATH}" "${MODELS_PATH}"

FAILED=()

for model in "${ARGS[@]}"; do
  if ! install_model "${model}"; then
    FAILED+=("${model}")
  fi
done

if [[ ${#FAILED[@]} -gt 0 ]]; then
  echo >&2
  echo "Failed: ${FAILED[*]}" >&2
  exit 1
fi
