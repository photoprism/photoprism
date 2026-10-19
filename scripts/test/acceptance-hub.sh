#!/usr/bin/env bash

# Gives a TestCafe fixture copy its own instance serial and keeps its hub registration. The restart rules run it
# before the binary, which reads the serial and would otherwise save a hub.yml without an API key.
set -euo pipefail

if (( $# < 1 || $# > 2 )); then
  echo "Usage: $0 <fixture folder> [hub backup file in config-sqlite]" >&2
  exit 2
fi

root=${1%/}
config=$root/config-sqlite
hub=$config/hub.yml
backup=${2:+$config/$2}
marker=$root/.testcafe-serial
bom=$'\xef\xbb\xbf'

if [[ -L "$config" || ! -d "$config" || -L "$root/backup" ]]; then
  echo "$0: $config must be a folder, and $root/backup must not be a symbolic link" >&2
  exit 1
elif [[ -e "$config/hub.yaml" || -L "$config/hub.yaml" ]]; then
  echo "$0: $config/hub.yaml is not supported, use hub.yml" >&2
  exit 1
elif [[ "${2:-}" == */* ]]; then
  echo "$0: the hub backup must be a file name in $config" >&2
  exit 2
fi

for f in "$hub" "$root/serial" "$root/backup/serial" "$marker" ${backup:+"$backup"}; do
  if [[ -L "$f" ]] || { [[ -e "$f" ]] && [[ ! -f "$f" || ! -r "$f" ]]; }; then
    echo "$0: $f must be a readable regular file" >&2
    exit 1
  fi
done

if [[ -n "$backup" && ! -f "$backup" ]]; then
  echo "$0: $backup not found" >&2
  exit 1
elif [[ -n "$backup" && "$backup" -ef "$hub" ]]; then
  echo "$0: $backup must not be $hub" >&2
  exit 1
fi

# identity prints the inode and birth time of the serial file and its value; a copied or unpacked fixture has new
# files, so the identity does not depend on the path through which the fixture is reached.
identity() {
  [[ -f "$root/serial" ]] || return 0
  printf '%s %s\n' "$(stat -c %i:%W -- "$root/serial")" "$(tr -d '[:space:]' < "$root/serial")"
}

# haskey reports whether hub.yml has a Key line with an API key, and stops the script if it cannot be read.
haskey() {
  local status=0
  [[ -f "$hub" ]] || return 1
  grep -aqE "^($bom)?Key:[[:space:]]+([0-9a-fA-F]{40}|\"[0-9a-fA-F]{40}\"|'[0-9a-fA-F]{40}')([[:space:]]+#.*)?[[:space:]]*\$" -- "$hub" ||
    status=$?

  if (( status > 1 )); then
    echo "$0: could not read $hub" >&2
    exit 1
  fi

  return "$status"
}

# replace writes the given file without Serial lines to hub.yml through a temporary file in the same folder, so that
# a failed run leaves the previous one. Config folders generated for other databases are linked to it again later.
replace() {
  hubtmp=$(mktemp "$config/.hub.yml.acceptance-hub.XXXXXX")
  trap 'rm -f -- "$hubtmp"' EXIT
  exec 3> "$hubtmp"
  grep -avE "^($bom)?Serial:" -- "$1" >&3 || (( $? == 1 ))
  exec 3>&-
  [[ ! -f "$hub" ]] || chmod --reference="$hub" -- "$hubtmp"
  mv -fT -- "$hubtmp" "$hub"
  trap - EXIT
}

# newserial prints a random serial in the format of the binary: "z", six base-36 digits of the time and nine random ones.
newserial() {
  local digits=0123456789abcdefghijklmnopqrstuvwxyz s="" t b
  t=$(date +%s)

  while (( t > 0 )); do
    s=${digits:t%36:1}$s
    t=$((t / 36))
  done

  s=z${s:0:6}

  for b in $(od -An -N9 -tu1 /dev/urandom); do
    s+=${digits:b%36:1}
  done

  printf '%s' "$s"
}

# The serial is replaced once per copy; a serial file with several links is shared with a copy made with hard links.
# Removing the backup copy lets the binary create a new one if it rejected the new serial.
serial=$(tr -d '[:space:]' 2>/dev/null < "$root/serial" || true)
rotated=0

if [[ ! "$serial" =~ ^z[0-9a-z]{15}$ || "$(stat -c %h -- "$root/serial")" != 1 ||
  "$(identity)" != "$(cat -- "$marker" 2>/dev/null)" ]]; then
  serial=$(newserial)

  if [[ ! "$serial" =~ ^z[0-9a-z]{15}$ ]]; then
    echo "$0: could not create a new instance serial" >&2
    exit 1
  fi

  rm -f -- "$root/serial" "$root/backup/serial" "$marker"
  (set -C; printf '%s' "$serial" > "$root/serial"; identity > "$marker")
  rotated=1
  echo "Created a new instance serial for $root."
fi

# Removes temporary files that a run stopped by a signal left behind.
for f in "$config"/.hub.yml.acceptance-hub.*; do
  if [[ -f "$f" && ! -L "$f" ]]; then
    rm -f -- "$f"
  fi
done

# hub.yml is restored from the backup unless it has an API key. Otherwise, Serial lines other than the copy's are
# removed, and a copy with a new serial gets a file of its own, since a copy made with hard links shares it.
if [[ -n "$backup" ]] && ! haskey; then
  replace "$backup"
  echo "Restored $hub from $backup."
elif [[ -f "$hub" ]] && { (( rotated )) ||
  [[ "$(grep -acE -- "^($bom)?Serial:" "$hub")" != "$(grep -acxE -- "($bom)?Serial: \"?$serial\"?" "$hub")" ]]; }; then
  replace "$hub"
fi
