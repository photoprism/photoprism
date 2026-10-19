#!/usr/bin/env bash

# Tests acceptance-hub.sh with scratch fixtures.
set -uo pipefail

script="$(cd "$(dirname "$0")" && pwd)/acceptance-hub.sh"
tmp=$(mktemp -d) || exit 1
trap 'rm -rf "$tmp"' EXIT
failed=0
archive=zqnqt973pwbcfm7b
key=0123456789abcdef0123456789abcdef01234567
other=fedcba9876543210fedcba9876543210fedcba98

# fixture creates a fixture folder with the archive serial, a hub.yml with a key, and a hub backup with another key,
# and returns its path.
fixture() {
  local dir=$tmp/$1
  mkdir -p "$dir/config-sqlite" "$dir/backup"
  printf '%s' "$archive" > "$dir/serial"
  printf '%s' "$archive" > "$dir/backup/serial"
  printf 'Key: %s\nSecret: secret\nSession: session\nStatus: ultimate\nSerial: %s\nEnv: test\n' "$key" "$archive" \
    > "$dir/config-sqlite/hub.yml"
  printf 'Key: %s\nSecret: backup\nStatus: test\nSerial: %s\nEnv: test\n' "$other" "$archive" \
    > "$dir/config-sqlite/hubBackup.yml"
  echo "$dir"
}

# check reports a failed condition with its name.
check() {
  local name=$1
  shift

  if "$@"; then
    echo "ok   $name"
  else
    echo "FAIL $name"
    failed=1
  fi
}

# expect runs the script and checks its exit status and output.
expect() {
  local name=$1 want=$2 pattern=$3 out status
  shift 3
  out=$(bash "$script" "$@" 2>&1)
  status=$?

  if [[ "$status" != "$want" ]] || ! grep -q -- "$pattern" <<< "$out"; then
    echo "FAIL $name: exit $status, want $want and \"$pattern\": $out"
    failed=1
  else
    echo "ok   $name"
  fi
}

# quiet runs the script and checks that it succeeds without output.
quiet() {
  local name=$1 out
  shift
  out=$(bash "$script" "$@" 2>&1)
  check "$name" test $? -eq 0 -a -z "$out"
}

# serial prints the serial of a fixture.
serial() {
  cat "$1/serial"
}

# valid reports whether a serial is accepted by the binary: "z", a base-36 time since 2017 that is not in the
# future, and nine more base-36 digits.
# shellcheck disable=SC2329
valid() {
  local s=$1 t
  [[ "$s" =~ ^z[0-9a-z]{15}$ ]] || return 1
  t=$((36#${s:1:6}))
  (( t >= 1483228800 && t <= $(date +%s) ))
}

# edit applies a sed expression to a file.
edit() {
  local content
  content=$(sed "$1" "$2") && printf '%s\n' "$content" > "$2"
}

# lacks reports whether the file does not contain the text.
# shellcheck disable=SC2329
lacks() {
  ! grep -q -- "$1" "$2"
}

fx=$(fixture fx)
hub=$fx/config-sqlite/hub.yml
expect "Usage" 2 "Usage"
expect "UsageArgs" 2 "Usage" "$fx" hubBackup.yml extra
expect "BackupName" 2 "file name" "$fx" ../hubBackup.yml
check "UsageUnchanged" test "$(serial "$fx")" == "$archive"

# The first run replaces the archive serial and keeps the key.
chmod 640 "$hub"
expect "First" 0 "new instance serial" "$fx" hubBackup.yml
first=$(serial "$fx")
check "FirstNewSerial" test "$first" != "$archive"
check "FirstValid" valid "$first"
check "FirstBackupSerialRemoved" test ! -e "$fx/backup/serial"
check "FirstMarker" test "$(cat "$fx/.testcafe-serial")" == "$(stat -c %i:%W "$fx/serial") $first"
check "FirstKey" grep -qx "Key: $key" "$hub"
check "FirstStatus" grep -qx "Status: ultimate" "$hub"
check "FirstNoSerial" lacks "Serial:" "$hub"
check "FirstMode" test "$(stat -c %a "$hub")" == 640
check "FirstNoTempFile" test -z "$(find "$fx/config-sqlite" -name '.hub.yml.*')"
check "FirstBackupUnchanged" grep -qx "Serial: $archive" "$fx/config-sqlite/hubBackup.yml"

# Later runs keep the serial, also after the binary restored its backup copy and saved it in hub.yml.
printf '%s' "$first" > "$fx/backup/serial"
echo "Serial: $first" >> "$hub"
cp "$hub" "$tmp/hub-saved.yml"
quiet "Second" "$fx" hubBackup.yml
check "SecondSerial" test "$(serial "$fx")" == "$first"
check "SecondBackupSerial" test "$(cat "$fx/backup/serial")" == "$first"
check "SecondHubUnchanged" cmp -s "$hub" "$tmp/hub-saved.yml"
quiet "Third" "$fx"
check "ThirdSerial" test "$(serial "$fx")" == "$first"

# A Serial line that differs from the serial of the copy is removed, and so are several lines.
sed "s/^Serial: .*/Serial: $archive/" "$tmp/hub-saved.yml" > "$hub"
quiet "OtherSerialLine" "$fx" hubBackup.yml
check "OtherSerialLineRemoved" lacks "Serial:" "$hub"
check "OtherSerialLineKey" grep -qx "Key: $key" "$hub"
edit "s/^Serial: .*/Serial: \"$first\"/" "$tmp/hub-saved.yml"
cp "$tmp/hub-saved.yml" "$hub"
quiet "QuotedSerialLine" "$fx" hubBackup.yml
check "QuotedSerialLineKept" cmp -s "$hub" "$tmp/hub-saved.yml"
check "OtherSerialLineSerial" test "$(serial "$fx")" == "$first"
printf 'Serial: %s\nSerial: %s\n' "$first" "$archive" >> "$hub"
quiet "SerialLines" "$fx" hubBackup.yml
check "SerialLinesRemoved" lacks "Serial:" "$hub"

# A copied fixture folder gets its own serial, although the marker was copied with it.
cp -a "$fx" "$tmp/copy"
expect "Copy" 0 "new instance serial" "$tmp/copy" hubBackup.yml
check "CopySerial" test "$(serial "$tmp/copy")" != "$first" -a "$(serial "$tmp/copy")" != "$archive"
check "CopyValid" valid "$(serial "$tmp/copy")"
check "CopyOriginalKept" test "$(serial "$fx")" == "$first"
check "CopyKey" grep -qx "Key: $key" "$tmp/copy/config-sqlite/hub.yml"
quiet "CopyOriginal" "$fx" hubBackup.yml
check "CopyOriginalSerial" test "$(serial "$fx")" == "$first"
cp -al "$fx" "$tmp/linkcopy"
expect "HardLinkCopy" 0 "new instance serial" "$tmp/linkcopy" hubBackup.yml
check "HardLinkCopySerial" test "$(serial "$tmp/linkcopy")" != "$first" -a "$(serial "$fx")" == "$first"
check "HardLinkCopyHub" test ! "$tmp/linkcopy/config-sqlite/hub.yml" -ef "$hub"
check "HardLinkCopyHubKey" grep -qx "Key: $key" "$tmp/linkcopy/config-sqlite/hub.yml"
quiet "HardLinkCopyOriginal" "$fx" hubBackup.yml

# A moved fixture, or one reached through another path, keeps its serial.
mv "$tmp/copy" "$tmp/moved"
moved=$(serial "$tmp/moved")
quiet "Moved" "$tmp/moved" hubBackup.yml
check "MovedSerial" test "$(serial "$tmp/moved")" == "$moved"
ln -s "$tmp/moved" "$tmp/linked-path"
quiet "OtherPath" "$tmp/linked-path" hubBackup.yml
check "OtherPathSerial" test "$(serial "$tmp/moved")" == "$moved"

# When the original of a copy made with hard links restarts first, it gets the new serial and the copy keeps its own.
cp -al "$tmp/moved" "$tmp/linkfirst"
expect "LinkOriginalFirst" 0 "new instance serial" "$tmp/moved" hubBackup.yml
quiet "LinkCopyAfter" "$tmp/linkfirst" hubBackup.yml
check "LinkPairDistinct" test "$(serial "$tmp/moved")" != "$(serial "$tmp/linkfirst")"
check "LinkPairHub" test ! "$tmp/moved/config-sqlite/hub.yml" -ef "$tmp/linkfirst/config-sqlite/hub.yml"

# A copy made with hard links that took the place of the original is recognized by the links of its serial file.
cp -a "$fx" "$tmp/samepath"
expect "SamePathSetup" 0 "new instance serial" "$tmp/samepath" hubBackup.yml
mv "$tmp/samepath" "$tmp/samepath-orig"
cp -al "$tmp/samepath-orig" "$tmp/samepath"
samepath=$(serial "$tmp/samepath")
expect "SamePathCopy" 0 "new instance serial" "$tmp/samepath" hubBackup.yml
check "SamePathCopySerial" test "$(serial "$tmp/samepath")" != "$samepath" -a "$(serial "$tmp/samepath-orig")" == "$samepath"
check "SamePathCopyHub" test ! "$tmp/samepath/config-sqlite/hub.yml" -ef "$tmp/samepath-orig/config-sqlite/hub.yml"

# A serial that was replaced in place, removed, or is not valid is replaced.
printf '%s' "$archive" > "$fx/serial"
expect "Overwritten" 0 "new instance serial" "$fx" hubBackup.yml
check "OverwrittenSerial" test "$(serial "$fx")" != "$archive" -a "$(serial "$fx")" != "$first"
second=$(serial "$fx")
rm "$fx/serial"
expect "Missing" 0 "new instance serial" "$fx" hubBackup.yml
check "MissingValid" valid "$(serial "$fx")"
check "MissingSerial" test "$(serial "$fx")" != "$second"
printf 'invalid' > "$fx/serial"
expect "Invalid" 0 "new instance serial" "$fx" hubBackup.yml
check "InvalidValid" valid "$(serial "$fx")"
rm "$fx/serial" "$fx/.testcafe-serial"
expect "MissingNoMarker" 0 "new instance serial" "$fx" hubBackup.yml
check "MissingNoMarkerValid" valid "$(serial "$fx")"
rm "$fx/.testcafe-serial"
expect "NoMarker" 0 "new instance serial" "$fx" hubBackup.yml
check "NoMarkerKey" grep -qx "Key: $key" "$hub"

# A hub.yml without a key is restored from the backup without its Serial line.
nokey=$(fixture nokey)
edit '/^Key:/d' "$nokey/config-sqlite/hub.yml"
expect "NoKey" 0 "Restored" "$nokey" hubBackup.yml
check "NoKeyBackupKey" grep -qx "Key: $other" "$nokey/config-sqlite/hub.yml"
check "NoKeyNoSerial" lacks "Serial:" "$nokey/config-sqlite/hub.yml"
check "NoKeyBackupUnchanged" grep -qx "Serial: $archive" "$nokey/config-sqlite/hubBackup.yml"
quiet "NoKeyAgain" "$nokey" hubBackup.yml
check "NoKeyAgainKey" grep -qx "Key: $other" "$nokey/config-sqlite/hub.yml"
edit 's/^Key: .*/Key: ""/' "$nokey/config-sqlite/hub.yml"
expect "EmptyKey" 0 "Restored" "$nokey" hubBackup.yml
check "EmptyKeyBackupKey" grep -qx "Key: $other" "$nokey/config-sqlite/hub.yml"

# Key lines that the binary does not read as a key are restored from the backup.
for line in "Key: $key#x" "Key: \"$key'" "Key: ${key}0" "  Key: $key" "Key:$key"; do
  printf 'Secret: secret\n%s\nStatus: ultimate\n' "$line" > "$nokey/config-sqlite/hub.yml"
  expect "KeyInvalid ($line)" 0 "Restored" "$nokey" hubBackup.yml
  check "KeyInvalidBackupKey ($line)" grep -qx "Key: $other" "$nokey/config-sqlite/hub.yml"
done

# Key lines with quotes, another case, trailing blanks or comments, CRLF line endings or a BOM are kept.
for line in "Key: \"$key\"" "Key: '$key'" "Key: ${key^^}" "Key: $key " $'Key: '"$key"$'\r' $'Key:\t'"$key # key" \
  $'\xef\xbb\xbfKey: '"$key"; do
  printf 'Secret: secret\n%s\nStatus: ultimate\n' "$line" > "$nokey/config-sqlite/hub.yml"
  cp "$nokey/config-sqlite/hub.yml" "$tmp/hub-key.yml"
  quiet "KeyForm (${line//$'\r'/\\r})" "$nokey" hubBackup.yml
  check "KeyFormKept (${line//$'\r'/\\r})" cmp -s "$nokey/config-sqlite/hub.yml" "$tmp/hub-key.yml"
done

nohub=$(fixture nohub)
rm "$nohub/config-sqlite/hub.yml"
expect "NoHub" 0 "Restored" "$nohub" hubBackup.yml
check "NoHubBackupKey" grep -qx "Key: $other" "$nohub/config-sqlite/hub.yml"
check "NoHubNoSerial" lacks "Serial:" "$nohub/config-sqlite/hub.yml"

# Without a backup, as in CE, a hub.yml keeps its key, and one without a key is not restored.
ce=$(fixture ce)
rm "$ce/config-sqlite/hubBackup.yml"
expect "CE" 0 "new instance serial" "$ce"
check "CEKey" grep -qx "Key: $key" "$ce/config-sqlite/hub.yml"
check "CENoSerial" lacks "Serial:" "$ce/config-sqlite/hub.yml"
edit '/^Key:/d' "$ce/config-sqlite/hub.yml"
quiet "CENoKey" "$ce"
check "CENoKeyNotRestored" lacks "Key:" "$ce/config-sqlite/hub.yml"
printf 'Serial: %s\n' "$archive" > "$ce/config-sqlite/hub.yml"
quiet "CEOnlySerial" "$ce"
check "CEOnlySerialRemoved" lacks "Serial:" "$ce/config-sqlite/hub.yml"
rm "$ce/config-sqlite/hub.yml"
quiet "CENoHub" "$ce"
check "CENoHubNotCreated" test ! -e "$ce/config-sqlite/hub.yml"

# A missing backup stops the script before it changes anything.
missing=$(fixture missing)
edit '/^Key:/d' "$missing/config-sqlite/hub.yml"
expect "BackupMissing" 1 "not found" "$missing" missing.yml
check "BackupMissingSerial" test "$(serial "$missing")" == "$archive"
check "BackupMissingHub" grep -qx "Serial: $archive" "$missing/config-sqlite/hub.yml"

# A backup that is hub.yml itself is refused, and one without other lines than Serial gives an empty hub.yml.
expect "BackupIsHub" 1 "must not be" "$missing" hub.yml
ln -f "$missing/config-sqlite/hub.yml" "$missing/config-sqlite/linked.yml"
expect "BackupLinkedToHub" 1 "must not be" "$missing" linked.yml
check "BackupIsHubSerial" test "$(serial "$missing")" == "$archive"
printf 'Serial: %s\n' "$archive" > "$missing/config-sqlite/serial-only.yml"
expect "BackupSerialOnly" 0 "Restored" "$missing" serial-only.yml
check "BackupSerialOnlyHub" lacks "Serial:" "$missing/config-sqlite/hub.yml"
printf '\xef\xbb\xbfSerial: %s\nKey: \nEnv: test\n' "$archive" > "$missing/config-sqlite/bom.yml"
expect "BackupBom" 0 "Restored" "$missing" bom.yml
check "BackupBomNoSerial" lacks "Serial:" "$missing/config-sqlite/hub.yml"
printf '\xef\xbb\xbfSerial: %s\nKey: %s\n' "$archive" "$key" > "$missing/config-sqlite/hub.yml"
quiet "HubBom" "$missing" bom.yml
check "HubBomNoSerial" lacks "Serial:" "$missing/config-sqlite/hub.yml"
check "HubBomKey" grep -qx "Key: $key" "$missing/config-sqlite/hub.yml"

# An unreadable hub.yml or a hub.yaml stops the script before it changes anything.
locked=$(fixture locked)

if (( $(id -u) != 0 )); then
  chmod 200 "$locked/config-sqlite/hub.yml"
  expect "Unreadable" 1 "readable regular file" "$locked" hubBackup.yml
  chmod 600 "$locked/config-sqlite/hub.yml"
  check "UnreadableKept" grep -qx "Key: $key" "$locked/config-sqlite/hub.yml"
  check "UnreadableSerial" test "$(serial "$locked")" == "$archive"
fi

touch "$locked/config-sqlite/hub.yaml"
expect "Yaml" 1 "not supported" "$locked" hubBackup.yml
check "YamlSerial" test "$(serial "$locked")" == "$archive"

# An interrupted write leaves the previous hub.yml.
broken=$(fixture broken)
expect "InterruptedSetup" 0 "new instance serial" "$broken" hubBackup.yml
echo "Serial: $archive" >> "$broken/config-sqlite/hub.yml"
cp "$broken/config-sqlite/hub.yml" "$tmp/hub-broken.yml"
(ulimit -f 0; bash "$script" "$broken" hubBackup.yml) >/dev/null 2>&1
check "InterruptedFails" test $? -ne 0
check "InterruptedHubKept" cmp -s "$broken/config-sqlite/hub.yml" "$tmp/hub-broken.yml"
quiet "InterruptedNext" "$broken" hubBackup.yml
check "InterruptedNoTempFile" test -z "$(find "$broken/config-sqlite" -name '.hub.yml.*')"
check "InterruptedNextKey" grep -qx "Key: $key" "$broken/config-sqlite/hub.yml"
check "InterruptedNextSerial" lacks "Serial:" "$broken/config-sqlite/hub.yml"

# Lines after binary content are kept.
printf 'Key: %s\nNote: \001\000\377\nStatus: ultimate\nSerial: %s\n' "$key" "$archive" > "$broken/config-sqlite/hub.yml"
quiet "Binary" "$broken" hubBackup.yml
check "BinaryKept" grep -aqx "Status: ultimate" "$broken/config-sqlite/hub.yml"
check "BinaryKey" grep -qx "Key: $key" "$broken/config-sqlite/hub.yml"
check "BinaryNoSerial" lacks "Serial:" "$broken/config-sqlite/hub.yml"
printf 'Key: %s\nSecret: secret\nStatus: ultimate\n' "$key" > "$broken/config-sqlite/hub.yml"

# A failed read leaves the previous hub.yml, and a failed serial generation leaves the previous serial.
mkdir "$tmp/bin"
real_grep=$(command -v grep)
# shellcheck disable=SC2016
printf '#!/usr/bin/env bash\nif [[ "$1" == -avE ]]; then echo partial; exit 2; fi\nexec %q "$@"\n' "$real_grep" > "$tmp/bin/grep"
printf '#!/usr/bin/env bash\nexit 1\n' > "$tmp/bin/od"
chmod +x "$tmp/bin/grep" "$tmp/bin/od"
echo "Serial: $archive" >> "$broken/config-sqlite/hub.yml"
cp "$broken/config-sqlite/hub.yml" "$tmp/hub-broken.yml"
PATH="$tmp/bin:$PATH" expect "ReadFails" 1 "" "$broken" hubBackup.yml
check "ReadFailsHubKept" cmp -s "$broken/config-sqlite/hub.yml" "$tmp/hub-broken.yml"
check "ReadFailsNoTempFile" test -z "$(find "$broken/config-sqlite" -name '.hub.yml.*')"
odfails=$(fixture odfails)
PATH="$tmp/bin:$PATH" expect "SerialFails" 1 "could not create" "$odfails" hubBackup.yml
check "SerialFailsKept" test "$(serial "$odfails")" == "$archive"

# Temporary files left by a stopped run are removed, other files and folders are kept.
leftover=$(fixture leftover)
touch "$leftover/config-sqlite/.hub.yml.acceptance-hub.AbC123" "$leftover/config-sqlite/.hub.yml.backup"
mkdir "$leftover/config-sqlite/.hub.yml.acceptance-hub.folder"
expect "Leftover" 0 "" "$leftover" hubBackup.yml
check "LeftoverRemoved" test ! -e "$leftover/config-sqlite/.hub.yml.acceptance-hub.AbC123"
check "LeftoverOthersKept" test -f "$leftover/config-sqlite/.hub.yml.backup" -a -d "$leftover/config-sqlite/.hub.yml.acceptance-hub.folder"

# Symbolic links and other files that are not regular files stop the script before it changes anything.
linked=$(fixture linked)
mv "$linked/config-sqlite/hub.yml" "$tmp/linked-hub.yml"
ln -s "$tmp/linked-hub.yml" "$linked/config-sqlite/hub.yml"
expect "HubSymlink" 1 "regular file" "$linked" hubBackup.yml
check "HubSymlinkKept" test -L "$linked/config-sqlite/hub.yml" -a "$(serial "$linked")" == "$archive"
check "HubSymlinkTarget" grep -qx "Serial: $archive" "$tmp/linked-hub.yml"
folder=$(fixture folder)
rm "$folder/config-sqlite/hub.yml"
mkdir "$folder/config-sqlite/hub.yml"
expect "HubFolder" 1 "regular file" "$folder" hubBackup.yml
check "HubFolderSerial" test "$(serial "$folder")" == "$archive"

for name in serial backup/serial .testcafe-serial config-sqlite/hubBackup.yml; do
  dir=$(fixture "link-${name//\//-}")
  printf '%s' "$archive" > "$tmp/target-${name//\//-}"
  rm -f "$dir/$name"
  ln -s "$tmp/target-${name//\//-}" "$dir/$name"
  expect "Symlink ($name)" 1 "regular file" "$dir" hubBackup.yml
  check "SymlinkTarget ($name)" test "$(cat "$tmp/target-${name//\//-}")" == "$archive" -a -L "$dir/$name"
  check "SymlinkHub ($name)" grep -qx "Serial: $archive" "$dir/config-sqlite/hub.yml"
done

for name in config-sqlite backup; do
  dir=$(fixture "folder-$name")
  mv "$dir/$name" "$tmp/real-$name"
  ln -s "$tmp/real-$name" "$dir/$name"
  expect "FolderSymlink ($name)" 1 "symbolic link" "$dir" hubBackup.yml
  check "FolderSymlinkSerial ($name)" test "$(serial "$dir")" == "$archive"
done

exit "$failed"
