#!/usr/bin/env bash

# Switches the main repository and every cloned sub-repository to the same branch, or prints their status.
#
# Usage: git-checkout.sh <branch>            fetch, switch to <branch>, and fast-forward it to origin/<branch>
#        git-checkout.sh --rebase <branch>   same, but rebase unpushed local commits onto origin/<branch>
#        git-checkout.sh --status            show branch, revision, changes, and ahead/behind counts
#
# It works on the repository it belongs to, wherever it is called from. It never merges or stashes, and rebases
# only with --rebase: then only commits that are on no remote are moved, merge commits are refused, and a
# rebase that stops (conflict, hook, signing) is aborted. Commits already upstream are dropped as empty.
# It refuses to start while any repository has uncommitted changes, an operation in progress, or commits
# that only a detached HEAD holds. The branch must exist on origin; exits 1 if any repository fails to switch.
set -uo pipefail

src=$0
while [[ -L $src ]]; do
  dir=$(cd -P "$(dirname "$src")" && pwd) || exit 1
  src=$(readlink "$src")
  [[ $src == /* ]] || src="$dir/$src"
done
cd -P "$(dirname "$src")/.." || exit 1

if [[ ! -e .git ]]; then
  echo "Not a Git repository: $(pwd)" >&2
  exit 1
fi

# repos lists the main repository and the cloned sub-repositories, one per line; symlinks are skipped.
repos() {
  echo "."
  local d
  for d in */; do
    d="${d%/}"
    [[ -L $d ]] && continue
    [[ -e "$d/.git" ]] && echo "$d"
  done
  return 0
}

# linked lists the symlinked folders that point to a Git repository, one per line.
linked() {
  local d
  for d in */; do
    d="${d%/}"
    [[ -L $d && -e "$d/.git" ]] && echo "$d"
  done
  return 0
}

# label prints the name a repository is reported under.
label() {
  if [[ $1 == "." ]]; then echo "photoprism"; else echo "photoprism/$1"; fi
}

# indent prints its input with each line indented by two spaces.
indent() {
  local line
  [[ -n $1 ]] || return 0
  while IFS= read -r line; do printf '  %s\n' "$line"; done <<<"$1"
}

# g runs git in a repository without taking optional locks or opening an editor.
g() {
  local r=$1
  shift
  GIT_EDITOR=true git --no-optional-locks -C "$r" "$@"
}

# status prints the branch, short revision, changed tracked files, and ahead/behind counts of each repository.
status() {
  local r head rev changes counts
  while IFS= read -r r; do
    head=$(g "$r" rev-parse --abbrev-ref HEAD 2>/dev/null) || head="?"
    rev=$(g "$r" rev-parse --short=9 HEAD 2>/dev/null) || rev="?"
    if changes=$(g "$r" status --porcelain --untracked-files=no 2>/dev/null); then
      changes=$(grep -c . <<<"$changes")
    else
      changes="?"
    fi
    counts=""
    if [[ $head != "HEAD" && $head != "?" ]] && g "$r" rev-parse --verify --quiet "refs/remotes/origin/$head" >/dev/null; then
      counts=$(g "$r" rev-list --left-right --count "HEAD...refs/remotes/origin/$head" 2>/dev/null | awk '{print "  ahead=" $1 " behind=" $2}')
    fi
    printf '%-22s %-12s %s  changes=%s%s\n' "$(label "$r")" "$head" "$rev" "$changes" "$counts"
  done < <(repos)
}

# rpath prints the path of a file in a repository's Git directory, relative to the main repository.
rpath() {
  local p
  p=$(g "$1" rev-parse --git-path "$2") || return 1
  [[ $p == /* ]] || p="$1/$p"
  echo "$p"
}

# blocked prints why a repository must not be switched now, or nothing if it may be.
blocked() {
  local r=$1 out p
  if ! out=$(g "$r" status --porcelain --untracked-files=no 2>&1); then
    echo "git status failed:"
    indent "$out"
    return
  fi
  if [[ -n $out ]]; then
    echo "uncommitted changes:"
    indent "$out"
    return
  fi
  for p in rebase-merge rebase-apply MERGE_HEAD CHERRY_PICK_HEAD REVERT_HEAD BISECT_START BISECT_LOG sequencer; do
    p=$(rpath "$r" "$p") || continue
    if [[ -e $p ]]; then
      echo "an operation is in progress (${p##*/}); finish or abort it first"
      return
    fi
  done
  if ! g "$r" symbolic-ref --quiet HEAD >/dev/null && [[ -n $(g "$r" rev-list -n 1 HEAD --not --branches --remotes --tags) ]]; then
    echo "detached HEAD has commits that no branch or tag contains; create a branch for them first"
  fi
}

if [[ $# -eq 1 && $1 == "--status" ]]; then
  status
  exit 0
fi

rebase=0
if [[ $# -eq 2 && $1 == "--rebase" ]]; then
  rebase=1
  shift
fi

if [[ $# -ne 1 || -z $1 ]]; then
  echo "Usage: $0 [--rebase] <branch> | --status" >&2
  exit 2
fi

branch=$1

if ! git check-ref-format "refs/heads/$branch" >/dev/null 2>&1 || [[ $branch == -* || $branch == *"@{"* ]]; then
  echo "Invalid branch name: $branch" >&2
  exit 2
fi

list=()
while IFS= read -r r; do list+=("$r"); done < <(repos)

# Refuse to switch any repository while one of them is not in a state that can be switched safely.
refused=0
for r in "${list[@]}"; do
  reason=$(blocked "$r")
  if [[ -n $reason ]]; then
    echo "$(label "$r"): $reason"
    refused=1
  fi
done

if [[ $refused -eq 1 ]]; then
  echo "Refusing to switch branches; nothing was changed."
  exit 1
fi

echo "Switching to \"$branch\"..."
failed=0
while IFS= read -r r; do
  printf '%-22s %s\n' "$(label "$r")" "SKIPPED (symlink, switch it in its own folder)"
  failed=1
done < <(linked)
for r in "${list[@]}"; do
  name=$(label "$r")
  remote="refs/remotes/origin/$branch"
  if ! out=$(g "$r" fetch --quiet --prune origin 2>&1); then
    printf '%-22s %s\n' "$name" "FETCH FAILED"
    indent "$out"
    failed=1
    continue
  fi
  if ! g "$r" rev-parse --verify --quiet "$remote" >/dev/null; then
    printf '%-22s %s\n' "$name" "NO \"$branch\" BRANCH ON ORIGIN"
    failed=1
    continue
  fi
  if g "$r" rev-parse --verify --quiet "refs/heads/$branch" >/dev/null; then
    out=$(g "$r" switch --quiet "$branch" 2>&1)
  else
    out=$(g "$r" switch --quiet --create "$branch" --track "origin/$branch" 2>&1)
  fi
  # shellcheck disable=SC2181
  if [[ $? -ne 0 ]]; then
    printf '%-22s %s\n' "$name" "SWITCH FAILED"
    indent "$out"
    failed=1
    continue
  fi
  if g "$r" merge-base --is-ancestor HEAD "$remote" || g "$r" merge-base --is-ancestor "$remote" HEAD; then
    if ! out=$(g "$r" merge --ff-only --quiet "$remote" 2>&1); then
      printf '%-22s %s\n' "$name" "NOT UPDATED"
      indent "$out"
      failed=1
      continue
    fi
  elif [[ $rebase -eq 0 ]]; then
    printf '%-22s %s\n' "$name" "NOT UPDATED (local branch has diverged from origin/$branch; try git-rebase)"
    failed=1
    continue
  elif [[ -n $(g "$r" rev-list -n 1 --merges "$remote..HEAD") ]]; then
    printf '%-22s %s\n' "$name" "NOT REBASED (unpushed merge commits; rebase them by hand)"
    failed=1
    continue
  elif [[ $(g "$r" rev-list --count "$remote..HEAD") != "$(g "$r" rev-list --count "$remote..HEAD" --not --remotes)" ]]; then
    printf '%-22s %s\n' "$name" "NOT REBASED (local commits are also on another remote branch; rebase them by hand)"
    failed=1
    continue
  else
    before=$(g "$r" rev-parse HEAD)
    if ! out=$(g "$r" rebase --quiet --no-autostash --no-autosquash --no-update-refs "$remote" 2>&1); then
      abort=$(g "$r" rebase --abort 2>&1)
      if [[ -e $(rpath "$r" rebase-merge) || -e $(rpath "$r" rebase-apply) || $(g "$r" rev-parse HEAD) != "$before" ||
        $(g "$r" symbolic-ref --quiet --short HEAD) != "$branch" ]]; then
        printf '%-22s %s\n' "$name" "LEFT MID-REBASE (the abort failed; run git rebase --abort there)"
        indent "$abort"
      else
        printf '%-22s %s\n' "$name" "REBASE FAILED, left unchanged (conflict or hook; rebase it by hand)"
      fi
      indent "$out"
      failed=1
      continue
    fi
    printf '%-22s %s\n' "$name" "OK (rebased, $(g "$r" rev-list --count "$remote..HEAD") local commits on top)"
    continue
  fi
  if ! g "$r" merge-base --is-ancestor "$remote" HEAD; then
    printf '%-22s %s\n' "$name" "NOT UPDATED"
    failed=1
    continue
  fi
  printf '%-22s %s\n' "$name" "OK"
done

echo
status

if [[ $failed -ne 0 ]]; then
  echo "One or more repositories did not switch to \"$branch\"; see above." >&2
  exit 1
fi
