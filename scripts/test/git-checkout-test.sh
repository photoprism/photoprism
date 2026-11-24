#!/usr/bin/env bash

# Tests git-checkout.sh against scratch repositories with local bare remotes.
# The assertions are single-quoted on purpose and evaluated by check, which also reads rc and out.
# shellcheck disable=SC2016,SC2034,SC2317,SC2329
set -uo pipefail

script="$(cd "$(dirname "$0")/.." && pwd)/git-checkout.sh"
tmp=$(mktemp -d) || exit 1
failed=0

cleanup() { rm -rf "$tmp"; }
trap cleanup EXIT

export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=Test GIT_AUTHOR_EMAIL=test@example.com GIT_COMMITTER_NAME=Test GIT_COMMITTER_EMAIL=test@example.com

# check reports a failed assertion without stopping the remaining cases.
check() {
  if eval "$2"; then echo "ok   $1"; else echo "FAIL $1"; failed=1; fi
}

# remote creates a bare origin at $tmp/remotes/<name> with a develop branch and the given extra branches.
remote() {
  local name=$1 seed="$tmp/seed-$1"
  shift
  git init -q -b develop "$seed"
  git -C "$seed" commit -q --allow-empty -m "init $name"
  for b in "$@"; do git -C "$seed" branch "$b"; done
  git clone -q --bare "$seed" "$tmp/remotes/$name.git"
  rm -rf "$seed"
}

# push adds a commit to a branch of a bare origin, so clones fall behind it.
push() {
  local w="$tmp/push-$1-$2"
  git clone -q -b "$2" "$tmp/remotes/$1.git" "$w"
  git -C "$w" commit -q --allow-empty -m "update $2"
  git -C "$w" push -q origin "$2"
  rm -rf "$w"
}

# workspace clones the main repository with a copy of the script and the given sub-repositories, all on branch feature.
workspace() {
  local w="$tmp/ws-$1"
  shift
  git clone -q "$tmp/remotes/main.git" "$w"
  git -C "$w" checkout -q -b feature
  mkdir -p "$w/scripts"
  cp "$script" "$w/scripts/git-checkout.sh"
  printf '/scripts/\n' > "$w/.git/info/exclude"
  for s in "$@"; do
    git clone -q "$tmp/remotes/$s.git" "$w/$s"
    git -C "$w/$s" checkout -q -b feature
    printf '/%s/\n' "$s" >> "$w/.git/info/exclude"
  done
  echo "$w"
}

# run calls the workspace copy of the script from the current directory.
run() { local w=$1; shift; bash "$w/scripts/git-checkout.sh" "$@"; }

# branch prints the current branch of a repository.
branch() { git -C "$1" rev-parse --abbrev-ref HEAD; }

# at reports whether a repository's HEAD equals origin/<branch>.
at() { [[ $(git -C "$1" rev-parse HEAD) == $(git -C "$1" rev-parse "origin/$2") ]]; }

remote main preview
remote pro preview
remote aaa
remote specs

# An existing local branch that is behind is fast-forwarded; a branch only on origin becomes a tracking branch.
w=$(workspace switch pro)
git -C "$w/pro" branch -q -D develop
push main develop
mkdir -p "$w/notarepo"
out=$(cd "$tmp" && run "$w" develop 2>&1); rc=$?
check "Switch: exit 0 when called from another folder" '[[ $rc -eq 0 ]]'
check "Switch: main on develop" '[[ $(branch "$w") == develop ]]'
check "Switch: local develop fast-forwarded" 'at "$w" develop'
check "Switch: pro on develop" '[[ $(branch "$w/pro") == develop ]] && at "$w/pro" develop'
check "Switch: pro tracks origin" '[[ $(git -C "$w/pro" rev-parse --abbrev-ref "develop@{upstream}") == origin/develop ]]'
check "Switch: folder without .git is skipped" '! grep -q notarepo <<<"$out"'
check "Switch: status table printed" 'grep -qE "^photoprism/pro +develop .* changes=0  ahead=0 behind=0" <<<"$out"'

# Running from inside a sub-repository still switches the whole set.
w=$(workspace inside pro)
out=$(cd "$w/pro" && run "$w" develop 2>&1); rc=$?
check "Inside: exit 0" '[[ $rc -eq 0 ]]'
check "Inside: main switched too" '[[ $(branch "$w") == develop ]]'

# A failing repository does not stop the ones after it.
w=$(workspace missing aaa pro specs)
out=$(run "$w" preview 2>&1); rc=$?
check "Missing: exit 1" '[[ $rc -eq 1 ]]'
check "Missing: reports aaa" 'grep -q "photoprism/aaa .*NO \"preview\" BRANCH ON ORIGIN" <<<"$out"'
check "Missing: pro after aaa still switched" '[[ $(branch "$w/pro") == preview ]]'
check "Missing: aaa unchanged" '[[ $(branch "$w/aaa") == feature ]]'

# A fetch failure is reported and the other repositories still switch.
w=$(workspace fetch pro)
git -C "$w/pro" remote set-url origin "$tmp/remotes/nonexistent.git"
out=$(run "$w" develop 2>&1); rc=$?
check "Fetch: exit 1" '[[ $rc -eq 1 ]]'
check "Fetch: reports pro" 'grep -q "photoprism/pro .*FETCH FAILED" <<<"$out"'
check "Fetch: main switched" '[[ $(branch "$w") == develop ]]'

# Uncommitted changes to tracked files block the switch in every repository.
w=$(workspace dirty pro)
echo change > "$w/pro/tracked.txt"
git -C "$w/pro" add tracked.txt && git -C "$w/pro" commit -q -m tracked
echo change2 >> "$w/pro/tracked.txt"
out=$(run "$w" develop 2>&1); rc=$?
check "Dirty: exit 1" '[[ $rc -eq 1 ]]'
check "Dirty: names the repository" 'grep -q "photoprism/pro: uncommitted changes" <<<"$out"'
check "Dirty: main not switched" '[[ $(branch "$w") == feature ]]'

# A repository whose status cannot be read counts as blocked, not as clean.
w=$(workspace broken pro)
echo junk > "$w/pro/.git/index"
out=$(run "$w" develop 2>&1); rc=$?
check "Broken: exit 1" '[[ $rc -eq 1 ]]'
check "Broken: reports status failure" 'grep -q "photoprism/pro: git status failed" <<<"$out"'
check "Broken: main not switched" '[[ $(branch "$w") == feature ]]'

# An operation in progress blocks the switch.
w=$(workspace bisect pro)
git -C "$w/pro" bisect start >/dev/null 2>&1
out=$(run "$w" develop 2>&1); rc=$?
check "Bisect: exit 1" '[[ $rc -eq 1 ]]'
check "Bisect: reports the operation" 'grep -q "photoprism/pro: an operation is in progress" <<<"$out"'
check "Bisect: main not switched" '[[ $(branch "$w") == feature ]]'

# Commits that only a detached HEAD holds block the switch; a detached HEAD on a known commit does not.
w=$(workspace detached pro)
git -C "$w/pro" checkout -q --detach
git -C "$w/pro" commit -q --allow-empty -m precious
precious=$(git -C "$w/pro" rev-parse HEAD)
out=$(run "$w" develop 2>&1); rc=$?
check "Detached: exit 1" '[[ $rc -eq 1 ]]'
check "Detached: reports the commits" 'grep -q "photoprism/pro: detached HEAD has commits" <<<"$out"'
check "Detached: commit still checked out" '[[ $(git -C "$w/pro" rev-parse HEAD) == "$precious" ]]'
git -C "$w/pro" checkout -q --detach origin/develop
out=$(run "$w" develop 2>&1); rc=$?
check "Detached: known commit switches" '[[ $rc -eq 0 && $(branch "$w/pro") == develop ]]'

# A local branch tracking another upstream still ends up at origin/<branch>.
w=$(workspace upstream)
git -C "$w" branch -q -D preview 2>/dev/null
git -C "$w" branch -q --no-track preview origin/preview
git -C "$w" branch -q --set-upstream-to=origin/develop preview
push main preview
out=$(run "$w" preview 2>&1); rc=$?
check "Upstream: exit 0" '[[ $rc -eq 0 ]]'
check "Upstream: at origin/preview" 'at "$w" preview'

# A local branch that diverged from origin is reported instead of merged.
w=$(workspace diverged)
git -C "$w" checkout -q develop
git -C "$w" commit -q --allow-empty -m "local only"
git -C "$w" checkout -q feature
push main develop
out=$(run "$w" develop 2>&1); rc=$?
check "Diverged: exit 1" '[[ $rc -eq 1 ]]'
check "Diverged: reports it" 'grep -q "NOT UPDATED" <<<"$out"'
check "Diverged: no merge commit" '[[ $(git -C "$w" rev-list --merges --count HEAD) -eq 0 ]]'

# A branch named like a tracked folder is switched to, not restored as a path.
seed="$tmp/seed-paths"
git init -q -b develop "$seed"
mkdir -p "$seed/docs" && echo a > "$seed/docs/a.txt"
git -C "$seed" add docs && git -C "$seed" commit -q -m docs
git -C "$seed" branch docs
git clone -q --bare "$seed" "$tmp/remotes/paths.git"
w="$tmp/ws-paths"
git clone -q "$tmp/remotes/paths.git" "$w"
git -C "$w" checkout -q -b feature
git -C "$w" config checkout.guess false
mkdir -p "$w/scripts" && cp "$script" "$w/scripts/git-checkout.sh" && echo /scripts/ > "$w/.git/info/exclude"
out=$(run "$w" docs 2>&1); rc=$?
check "Paths: exit 0" '[[ $rc -eq 0 ]]'
check "Paths: on branch docs" '[[ $(branch "$w") == docs ]]'

# Symlinked folders are skipped, so another repository is never switched.
remote other
git clone -q "$tmp/remotes/other.git" "$tmp/other"
git -C "$tmp/other" checkout -q -b feature
w=$(workspace symlink)
ln -s "$tmp/other" "$w/linked"
echo /linked >> "$w/.git/info/exclude"
out=$(run "$w" develop 2>&1); rc=$?
check "Symlink: exit 1" '[[ $rc -eq 1 ]]'
check "Symlink: reported as skipped" 'grep -q "photoprism/linked .*SKIPPED (symlink" <<<"$out"'
check "Symlink: linked repository untouched" '[[ $(branch "$tmp/other") == feature ]]'
check "Symlink: main still switched" '[[ $(branch "$w") == develop ]]'

# A symlink to the script resolves to the repository the script belongs to.
mkdir -p "$tmp/bin"
w=$(workspace scriptlink)
ln -s "$w/scripts/git-checkout.sh" "$tmp/bin/gco"
(cd "$tmp" && bash "$tmp/bin/gco" develop >/dev/null 2>&1); rc=$?
check "Scriptlink: exit 0" '[[ $rc -eq 0 && $(branch "$w") == develop ]]'

# A branch deleted on origin is not checked out from a stale tracking ref.
w=$(workspace pruned)
git -C "$tmp/remotes/main.git" branch -q temp develop
git -C "$w" fetch -q origin
git -C "$tmp/remotes/main.git" branch -q -D temp
out=$(run "$w" temp 2>&1); rc=$?
check "Pruned: exit 1" '[[ $rc -eq 1 && $(branch "$w") == feature ]]'
check "Pruned: reports no branch" 'grep -q "NO \"temp\" BRANCH ON ORIGIN" <<<"$out"'

# Untracked files do not block the switch.
w=$(workspace untracked pro)
echo scratch > "$w/pro/untracked.txt"
(run "$w" develop >/dev/null 2>&1); rc=$?
check "Untracked: exit 0" '[[ $rc -eq 0 ]]'
check "Untracked: pro on develop" '[[ $(branch "$w/pro") == develop ]]'

# Status mode changes nothing, lists every repository, and shows unpushed commits.
w=$(workspace status pro)
git -C "$w/pro" checkout -q develop
git -C "$w/pro" commit -q --allow-empty -m unpushed
refs() { git -C "$w" for-each-ref; git -C "$w/pro" for-each-ref; }
before=$(refs | sha1sum)
out=$(run "$w" --status 2>&1); rc=$?
check "Status: exit 0" '[[ $rc -eq 0 ]]'
check "Status: lists main" 'grep -qE "^photoprism +feature " <<<"$out"'
check "Status: shows ahead count" 'grep -qE "^photoprism/pro +develop .* ahead=1 behind=0" <<<"$out"'
check "Status: no switch or fetch" '[[ $(branch "$w") == feature && $(refs | sha1sum) == "$before" ]]'

# pushfile commits a file with the given content to a branch of a bare origin.
pushfile() {
  local w="$tmp/pushfile-$1-$2"
  git clone -q -b "$2" "$tmp/remotes/$1.git" "$w"
  echo "$4" > "$w/$3"
  git -C "$w" add "$3" && git -C "$w" commit -q -m "upstream $3"
  git -C "$w" push -q origin "$2"
  rm -rf "$w"
}

# Rebase mode puts unpushed local commits on top of origin instead of failing.
w=$(workspace rebase pro)
git -C "$w" checkout -q develop
echo local > "$w/local.txt" && git -C "$w" add local.txt && git -C "$w" commit -q -m "local work"
git -C "$w" checkout -q feature
pushfile main develop upstream.txt one
out=$(run "$w" develop 2>&1); rc=$?
check "Rebase: plain mode refuses and points to git-rebase" '[[ $rc -eq 1 ]] && grep -q "try git-rebase" <<<"$out"'
out=$(run "$w" --rebase develop 2>&1); rc=$?
check "Rebase: exit 0" '[[ $rc -eq 0 ]]'
check "Rebase: reports the count" 'grep -q "photoprism .*OK (rebased, 1 local commits on top)" <<<"$out"'
check "Rebase: local commit on top of origin" '[[ $(git -C "$w" rev-parse HEAD^) == $(git -C "$w" rev-parse origin/develop) && $(git -C "$w" log -1 --format=%s) == "local work" ]]'
check "Rebase: no merge commit" '[[ $(git -C "$w" rev-list --merges --count origin/develop..HEAD) -eq 0 ]]'
check "Rebase: pro switched too" '[[ $(branch "$w/pro") == develop ]] && at "$w/pro" develop'

# A conflict aborts the rebase in that repository, leaves it as it was, and the others are still processed.
w=$(workspace conflict pro)
git -C "$w" checkout -q develop
echo mine > "$w/clash.txt" && git -C "$w" add clash.txt && git -C "$w" commit -q -m "mine"
mine=$(git -C "$w" rev-parse HEAD)
git -C "$w" checkout -q feature
git -C "$w/pro" checkout -q develop
echo pro > "$w/pro/pro.txt" && git -C "$w/pro" add pro.txt && git -C "$w/pro" commit -q -m "pro work"
pushfile main develop clash.txt theirs
pushfile pro develop other.txt theirs
out=$(run "$w" --rebase develop 2>&1); rc=$?
check "Conflict: exit 1" '[[ $rc -eq 1 ]]'
check "Conflict: reported" 'grep -q "photoprism .*REBASE FAILED, left unchanged" <<<"$out"'
check "Conflict: commit and branch unchanged" '[[ $(branch "$w") == develop && $(git -C "$w" rev-parse HEAD) == "$mine" ]]'
check "Conflict: no rebase left in progress" '[[ ! -e "$w/.git/rebase-merge" && ! -e "$w/.git/rebase-apply" && -z $(git -C "$w" status --porcelain --untracked-files=no) ]]'
check "Conflict: pro still rebased" 'grep -q "photoprism/pro .*OK (rebased, 1 local commits on top)" <<<"$out" && [[ $(git -C "$w/pro" rev-parse HEAD^) == $(git -C "$w/pro" rev-parse origin/develop) ]]'

# Commits that are also on another remote branch are not rewritten.
w=$(workspace shared)
git -C "$w" checkout -q develop
git -C "$w" checkout -q -b topic
git -C "$w" commit -q --allow-empty -m "pushed topic"
git -C "$w" push -q origin topic
git -C "$w" checkout -q develop
git -C "$w" merge -q --ff-only topic
git -C "$w" commit -q --allow-empty -m "local on develop"
shared=$(git -C "$w" rev-parse HEAD)
git -C "$w" branch -q other HEAD~1
git -C "$w" config rebase.updateRefs true
git -C "$w" checkout -q feature
push main develop
out=$(run "$w" --rebase develop 2>&1); rc=$?
check "Shared: exit 1" '[[ $rc -eq 1 ]]'
check "Shared: reported" 'grep -q "NOT REBASED (local commits are also on another remote branch" <<<"$out"'
check "Shared: develop unchanged" '[[ $(git -C "$w" rev-parse develop) == "$shared" ]]'
git -C "$w" push -q origin --delete topic
git -C "$w" fetch -q --prune origin
otherbefore=$(git -C "$w" rev-parse other)
out=$(run "$w" --rebase develop 2>&1); rc=$?
check "UpdateRefs: rebased" '[[ $rc -eq 0 && $(git -C "$w" rev-parse HEAD~2) == $(git -C "$w" rev-parse origin/develop) ]]'
check "UpdateRefs: other local branch not moved" '[[ $(git -C "$w" rev-parse other) == "$otherbefore" ]]'

# A failing hook stops the rebase, which is aborted and reported, not left in progress.
w=$(workspace hook)
git -C "$w" checkout -q develop
git -C "$w" commit -q --allow-empty -m "local work"
hooked=$(git -C "$w" rev-parse HEAD)
git -C "$w" checkout -q feature
printf '#!/bin/sh\nexit 1\n' > "$w/.git/hooks/pre-rebase" && chmod +x "$w/.git/hooks/pre-rebase"
push main develop
out=$(run "$w" --rebase develop 2>&1); rc=$?
check "Hook: exit 1 and unchanged" '[[ $rc -eq 1 && $(git -C "$w" rev-parse develop) == "$hooked" ]]'
check "Hook: reported as failed" 'grep -q "REBASE FAILED, left unchanged" <<<"$out"'

# Unpushed merge commits are not flattened.
w=$(workspace merges)
git -C "$w" checkout -q develop
git -C "$w" checkout -q -b side
git -C "$w" commit -q --allow-empty -m side
git -C "$w" checkout -q develop
git -C "$w" commit -q --allow-empty -m main
git -C "$w" merge -q --no-ff --no-edit side
merged=$(git -C "$w" rev-parse HEAD)
git -C "$w" checkout -q feature
push main develop
out=$(run "$w" --rebase develop 2>&1); rc=$?
check "Merges: exit 1" '[[ $rc -eq 1 ]]'
check "Merges: reported" 'grep -q "NOT REBASED (unpushed merge commits" <<<"$out"'
check "Merges: branch unchanged" '[[ $(git -C "$w" rev-parse develop) == "$merged" ]]'

# A branch that is only ahead of origin is left as it is.
w=$(workspace ahead)
git -C "$w" checkout -q develop
git -C "$w" commit -q --allow-empty -m unpushed
ahead=$(git -C "$w" rev-parse HEAD)
git -C "$w" checkout -q feature
out=$(run "$w" --rebase develop 2>&1); rc=$?
check "Ahead: exit 0 and unchanged" '[[ $rc -eq 0 && $(git -C "$w" rev-parse HEAD) == "$ahead" ]]'
check "Ahead: status shows it" 'grep -qE "^photoprism +develop .* ahead=1 behind=0" <<<"$out"'
(run "$w" --rebase >/dev/null 2>&1); rc=$?
check "Ahead: --rebase without a branch exits 2" '[[ $rc -eq 2 ]]'

# Invalid arguments are rejected before anything runs.
for name in "bad..name" "--force" "-x" "@{-1}" "a b"; do
  (run "$w" "$name" >/dev/null 2>&1); rc=$?
  check "Invalid: \"$name\" exits 2" '[[ $rc -eq 2 ]]'
done
(run "$w" >/dev/null 2>&1); rc=$?
check "Invalid: no argument exits 2" '[[ $rc -eq 2 ]]'
mkdir -p "$tmp/empty/scripts" && cp "$script" "$tmp/empty/scripts/git-checkout.sh"
(run "$tmp/empty" develop >/dev/null 2>&1); rc=$?
check "Invalid: not a repository exits 1" '[[ $rc -eq 1 ]]'

exit $failed
