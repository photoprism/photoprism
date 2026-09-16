#!/usr/bin/env bash

# Verifies by outcome which packaging path install-libheif.sh selects, and that it refuses an
# install that would shadow a newer library. Both decisions are valid shell either way round, so
# an inverted one does not fail a build, it changes which library the loader resolves to. Each
# case therefore asserts the artifact the script requested and the decision it announced, never
# the branch it took.
#
# Usage: scripts/check-libheif-install.sh
#
# Each case runs the real script as root in a user namespace, with the network stubbed, the
# distribution codename pinned, the target architecture fixed and a private /usr/local, so
# nothing outside the sandbox is read or written.

set -euo pipefail

cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."

# Every run_case call below, so that deleting one is a failure rather than a smaller report.
EXPECTED_CASES=42

INSTALLER="scripts/dist/install-libheif.sh"

if [[ ! -x ${INSTALLER} ]]; then
  echo "${INSTALLER}: not found or not executable, run this from a working copy" 1>&2
  exit 1
fi

INSTALLER=$(realpath "${INSTALLER}")

# The sandbox is what keeps the cases hermetic. In our own images it is always available, so its
# absence there is a broken environment; elsewhere - a host with unprivileged user namespaces
# restricted, which is the stock Ubuntu 24.04 AppArmor default - the cases cannot run at all, and
# saying so is better than failing a lint target that is otherwise portable.
if ! unshare -r -m true 2> /dev/null; then
  echo "Unprivileged user namespaces are unavailable, so the cases cannot be isolated." 1>&2

  if [[ ${PHOTOPRISM_CONTAINER:-} == "true" ]]; then
    echo "This is a container image, where they are expected to work." 1>&2
    exit 1
  fi

  echo "Skipping; run this in the development container to check the libheif installer." 1>&2
  exit 0
fi

# /usr/local is masked with an empty tmpfs so the prefix a case describes is the only one the
# installer can see, and so a case that ever gets past the stubbed download writes into the
# namespace rather than onto this host. Without the mount point there is nothing to mask.
if [[ ! -d /usr/local ]]; then
  echo "/usr/local does not exist, so the prefix cases cannot be isolated from this host." 1>&2
  exit 1
fi

WORK_DIR=$(mktemp -d "${TMPDIR:-/tmp}/photoprism-libheif-check.XXXXXXXX")
trap 'rm -rf "${WORK_DIR}"' EXIT

STUB_DIR="${WORK_DIR}/bin"
mkdir -p "${STUB_DIR}"

# curl records the artifact it was asked for and then fails, so every case stops at the download
# and the recorded URL says which path was taken. Checksum companions are not recorded, since
# they follow the artifact rather than identifying it.
cat > "${STUB_DIR}/curl" << 'STUB'
#!/usr/bin/env bash
url=""
for arg in "$@"; do
  case ${arg} in
    http://* | https://*) url=${arg} ;;
  esac
done
case ${url} in
  *.sha256) exit 1 ;;
esac
printf '%s\n' "${url}" >> "${LIBHEIF_CHECK_REQUESTS}"
exit 1
STUB

# dpkg answers only the "is our package installed" query the prefix path makes.
cat > "${STUB_DIR}/dpkg" << 'STUB'
#!/usr/bin/env bash
if [[ ${1:-} == "-s" ]] && [[ ${2:-} == "photoprism-libheif" ]]; then
  [[ ${LIBHEIF_CHECK_PP_INSTALLED:-0} == "1" ]] || exit 1
  echo "Package: photoprism-libheif"
  exit 0
fi
echo "amd64"
STUB

# dpkg-query records how it was called and reports the packages the case describes. Recording the
# arguments is what lets a case pin the query itself: a pattern that stops matching "libheif1",
# or a format whose fields move, produces an empty listing that every "no refusal" case would
# otherwise accept. Each entry is "<name>/<version>" with an optional "/<state>".
cat > "${STUB_DIR}/dpkg-query" << 'STUB'
#!/usr/bin/env bash
printf '%s\n' "$@" >> "${LIBHEIF_CHECK_QUERY}"
[[ -n ${LIBHEIF_CHECK_DISTRO:-} ]] || exit 1
for entry in ${LIBHEIF_CHECK_DISTRO}; do
  name=${entry%%/*}
  rest=${entry#*/}
  version=${rest%%/*}
  state=installed
  if [[ ${rest} == */* ]]; then
    state=${rest#*/}
  fi
  printf '%s %s %s\n' "${name}" "${version}" "${state}"
done
STUB

# apt-get and its neighbors only need to exist for the packaging path to be selected; the cases
# stop at the download, so a run that reaches one of them is itself the finding.
for stub in apt-get apt-mark dpkg-deb ldconfig; do
  cat > "${STUB_DIR}/${stub}" << STUB
#!/usr/bin/env bash
echo "${stub} was called, which no case should reach" 1>&2
exit 1
STUB
done

chmod 0755 "${STUB_DIR}"/*

FAILED=0
CHECKED=0
CASE_NAME=""
LAST_EXIT=0
LAST_OUTPUT=""
LAST_REQUESTS=""
LAST_QUERY=""

# fail reports a case that did not hold, naming what was expected and what happened.
fail() {
  echo "FAIL: ${CASE_NAME}" 1>&2
  echo "      ${1}" 1>&2
  echo "      exit ${LAST_EXIT}, requested: ${LAST_REQUESTS:-nothing}" 1>&2
  FAILED=$((FAILED + 1))
}

# run_case executes the installer once under the described conditions. Settings are given as
# key=value pairs, and anything after "--" is passed to the installer as an argument.
#
#   marker    PHOTOPRISM_CONTAINER value, or "unset" to leave it out of the environment
#   codename  the distribution the sandbox reports through /etc/os-release
#   existing  version to plant in the prefix; also "plain" or "alien" for an unversioned one
#   distro    distribution packages, as "<name>/<version>[/<state>]" separated by spaces
#   pp        whether dpkg reports photoprism-libheif as installed
#   destdir   the prefix passed to the installer
run_case() {
  CASE_NAME="$1"
  shift

  local marker="unset" codename="resolute" existing="" distro="" pp=0 destdir="/usr/local"
  local -a args=()

  while [[ $# -gt 0 ]]; do
    case "$1" in
      marker=*) marker="${1#*=}" ;;
      codename=*) codename="${1#*=}" ;;
      existing=*) existing="${1#*=}" ;;
      distro=*) distro="${1#*=}" ;;
      pp=*) pp="${1#*=}" ;;
      destdir=*) destdir="${1#*=}" ;;
      --)
        shift
        args=("$@")
        break
        ;;
      *)
        echo "run_case: unknown setting '$1'" 1>&2
        exit 1
        ;;
    esac
    shift
  done

  CHECKED=$((CHECKED + 1))

  local os_release="${WORK_DIR}/os-release"
  printf 'ID=ubuntu\nVERSION_CODENAME=%s\n' "${codename}" > "${os_release}"

  local requests="${WORK_DIR}/requests" query="${WORK_DIR}/query"
  : > "${requests}"
  : > "${query}"

  export LIBHEIF_CHECK_REQUESTS="${requests}"
  export LIBHEIF_CHECK_QUERY="${query}"
  export LIBHEIF_CHECK_PP_INSTALLED="${pp}"
  export LIBHEIF_CHECK_DISTRO="${distro}"
  export LIBHEIF_CHECK_DESTDIR="${destdir}"
  export LIBHEIF_CHECK_EXISTING="${existing}"
  export LIBHEIF_CHECK_OSRELEASE="${os_release}"
  export LIBHEIF_CHECK_STUBS="${STUB_DIR}"
  export LIBHEIF_CHECK_INSTALLER="${INSTALLER}"

  # The installer reads the target architecture from the environment before falling back to
  # "uname -m", so pinning it keeps every case asking for the same artifact name on any machine.
  export PHOTOPRISM_ARCH="amd64"
  unset BUILD_ARCH

  if [[ ${marker} == "unset" ]]; then
    unset PHOTOPRISM_CONTAINER
  else
    export PHOTOPRISM_CONTAINER="${marker}"
  fi

  # The inner script is quoted so it expands inside the namespace, where the mounts exist.
  # /etc/os-release is a symlink to /usr/lib/os-release on our base images, so the bind lands
  # there; a future case that masks /usr/lib would silently lose the codename pin.
  set +e
  # shellcheck disable=SC2016
  LAST_OUTPUT=$(unshare -r -m bash -c '
    set -eu
    mount --bind "${LIBHEIF_CHECK_OSRELEASE}" /etc/os-release
    mount -t tmpfs none /usr/local
    mkdir -p /usr/local/lib

    case ${LIBHEIF_CHECK_EXISTING} in
      "") ;;
      plain)
        mkdir -p "${LIBHEIF_CHECK_DESTDIR}/lib"
        : > "${LIBHEIF_CHECK_DESTDIR}/lib/libheif.so.1"
        ;;
      alien)
        mkdir -p "${LIBHEIF_CHECK_DESTDIR}/lib"
        : > "${LIBHEIF_CHECK_DESTDIR}/lib/libvendored.so"
        ln -sf libvendored.so "${LIBHEIF_CHECK_DESTDIR}/lib/libheif.so.1"
        ;;
      *)
        mkdir -p "${LIBHEIF_CHECK_DESTDIR}/lib"
        : > "${LIBHEIF_CHECK_DESTDIR}/lib/libheif.so.${LIBHEIF_CHECK_EXISTING}"
        ln -sf "libheif.so.${LIBHEIF_CHECK_EXISTING}" "${LIBHEIF_CHECK_DESTDIR}/lib/libheif.so.1"
        ;;
    esac

    export PATH="${LIBHEIF_CHECK_STUBS}:${PATH}"
    exec bash "${LIBHEIF_CHECK_INSTALLER}" "$@"
  ' _ "${LIBHEIF_CHECK_DESTDIR}" "${args[@]+"${args[@]}"}" 2>&1)
  LAST_EXIT=$?
  set -e

  LAST_REQUESTS=$(tr '\n' ' ' < "${requests}" | sed 's/ *$//')
  LAST_QUERY=$(cat "${query}")
}

# expect_request asserts that the run asked for exactly one artifact, ending in the given
# extension. The count matters: a run that tries the package and then falls back to the tarball
# would satisfy a test on the last request alone, and that degraded path is worth telling apart.
expect_request() {
  if [[ -z ${LAST_REQUESTS} ]]; then
    fail "expected one request for a ${1} artifact, but nothing was downloaded"
    return
  fi

  if [[ ${LAST_REQUESTS} == *" "* ]]; then
    fail "expected exactly one request, ending in ${1}"
    return
  fi

  if [[ ${LAST_REQUESTS} != *"${1}" ]]; then
    fail "expected the request to end in ${1}"
  fi
}

# expect_no_request asserts that the run stopped before asking for anything.
expect_no_request() {
  if [[ -n ${LAST_REQUESTS} ]]; then
    fail "expected no download, so the run should have stopped before it"
  fi
}

# expect_exit asserts the exit status of the run.
expect_exit() {
  if [[ ${LAST_EXIT} != "$1" ]]; then
    fail "expected exit ${1}"
  fi
}

# expect_output asserts that the run said something the decision is recognizable by.
expect_output() {
  if [[ ${LAST_OUTPUT} != *"$1"* ]]; then
    fail "expected the output to contain \"${1}\""
  fi
}

# expect_not_output asserts that the run did not say something.
expect_not_output() {
  if [[ ${LAST_OUTPUT} == *"$1"* ]]; then
    fail "expected the output not to contain \"${1}\""
  fi
}

# expect_query_arg asserts that the package listing was requested with the given argument, so
# that the pattern and the field order the parser depends on are pinned rather than assumed.
expect_query_arg() {
  if ! grep -qxF -- "$1" <<< "${LAST_QUERY}"; then
    fail "expected the package listing to be requested with \"${1}\""
  fi
}

# Which packaging path is selected. The marker is what our own images set, and it has to be the
# only thing besides an explicit flag that arms the dpkg paths.
run_case "no marker installs the tarball into the prefix"
expect_request ".tar.gz"
run_case "the container marker installs the Debian package" marker=true
expect_request ".deb"
run_case "--deb installs the Debian package without the marker" -- --deb
expect_request ".deb"
run_case "a marker other than true does not arm the packaging path" marker=false
expect_request ".tar.gz"
run_case "the Debian package is built for Resolute only" marker=true codename=jammy
expect_request ".tar.gz"
run_case "a custom prefix installs the tarball, since dpkg paths are absolute" \
  marker=true destdir="${WORK_DIR}/prefix"
expect_request ".tar.gz"

# The marker also decides whether the prefix path warns about the packaged copy it will shadow,
# which is a second, independent reading of the same gate.
run_case "the prefix path says the reported version will stop tracking the one in use" pp=1
expect_request ".tar.gz"
expect_output "photoprism-libheif is installed through apt"
run_case "the packaging path does not, because dpkg keeps describing what is loaded" \
  marker=true codename=jammy pp=1
expect_request ".tar.gz"
expect_not_output "photoprism-libheif is installed through apt"

# Installing into dpkg's own prefix without the packaging path. The second case passes
# --downgrade because /usr/lib cannot be masked - the sandbox's own binaries live there - so
# without it the case would read whatever libheif this host carries. It must also never reach
# the tarball path, which would unpack over /usr; the stubbed download is what keeps that
# unreachable, so do not teach the stub to succeed without masking the prefix first.
run_case "installing into /usr without --deb is refused" destdir=/usr
expect_exit 1
expect_no_request
expect_output "would leave files no package owns"
run_case "installing into /usr with the marker is allowed" marker=true destdir=/usr -- --downgrade
expect_request ".deb"

# Going backwards in the prefix. The loader resolves to the highest version it finds there, so a
# copy left behind keeps winning for every consumer.
run_case "an empty prefix installs" -- v1.23.4
expect_request ".tar.gz"
expect_not_output "Replacing libheif"
run_case "a newer version replaces an older one" existing=1.19.7 -- v1.23.4
expect_request ".tar.gz"
expect_output "Replacing libheif 1.19.7"
run_case "the same version is not reported as a replacement" existing=1.23.4 -- v1.23.4
expect_request ".tar.gz"
expect_not_output "Replacing libheif"
run_case "an older version over a newer one is refused" existing=1.24.1 -- v1.23.4
expect_exit 1
expect_no_request
expect_output "which is newer than 1.23.4"
run_case "--downgrade installs it anyway" existing=1.24.1 -- --downgrade v1.23.4
expect_request ".tar.gz"
expect_output "Replacing libheif 1.24.1"
run_case "versions are compared numerically, not as text" existing=1.23.10 -- v1.23.9
expect_exit 1
expect_no_request
expect_output "which is newer than 1.23.9"
run_case "a higher minor is not read as a lower one" existing=1.9.0 -- v1.10.0
expect_request ".tar.gz"
expect_output "Replacing libheif 1.9.0"
run_case "a higher major is refused as well" existing=2.0.0 -- v1.23.4
expect_exit 1
expect_no_request
expect_output "which is newer than 1.23.4"
run_case "an unversioned library is installed over, and said so" existing=plain -- v1.23.4
expect_request ".tar.gz"
expect_output "carries no recognizable version"
run_case "a link into an unrelated library is treated the same way" existing=alien -- v1.23.4
expect_request ".tar.gz"
expect_output "carries no recognizable version"

# Shadowing the distribution. The prefix goes ahead of the distribution's library directory for
# everything that loads libheif, so the version written there is the one consumers get.
run_case "an older build than the distribution's is refused" \
  distro="libheif1/1.24.1-1build2" -- v1.23.4
expect_exit 1
expect_no_request
expect_output "newer than the 1.23.4"
run_case "--downgrade shadows it anyway" distro="libheif1/1.24.1-1build2" -- --downgrade v1.23.4
expect_request ".tar.gz"
expect_output "Shadowing the distribution"
run_case "a newer build than the distribution's installs" distro="libheif1/1.21.2-1" -- v1.23.4
expect_request ".tar.gz"
expect_not_output "newer than the 1.23.4"

# The listing itself, so that a pattern which stops matching the library package, or a format
# whose fields move, cannot pass as "nothing is installed".
run_case "the packaged libheif is looked up by name and state" distro="libheif1/1.21.2-1" -- v1.23.4
expect_query_arg 'libheif*'
# The dpkg template is data, not shell: it must reach dpkg-query unexpanded.
# shellcheck disable=SC2016
expect_query_arg '-f=${Package} ${Version} ${db:Status-Status}\n'

# What a dpkg version carries around the upstream one.
run_case "an epoch is not read as part of the version" distro="libheif1/1:1.24.1-1" -- v1.23.4
expect_exit 1
expect_no_request
expect_output "newer than the 1.23.4"
run_case "a Debian revision is not read as a newer upstream version" \
  distro="libheif1/1.23.4-1build2" -- v1.23.4
expect_request ".tar.gz"
expect_not_output "newer than the 1.23.4"
run_case "a repack of the same release is not read as newer" \
  distro="libheif1/1.23.4+dfsg-1" -- v1.23.4
expect_request ".tar.gz"
expect_not_output "newer than the 1.23.4"
run_case "a +really repack is not read as newer either" \
  distro="libheif1/1.23.4+really1.23.4-1" -- v1.23.4
expect_request ".tar.gz"
expect_not_output "newer than the 1.23.4"
run_case "a tilde pre-release sorts below the release" distro="libheif1/1.23.4~rc1-1" -- v1.23.4
expect_request ".tar.gz"
expect_not_output "newer than the 1.23.4"
run_case "a hyphen in the upstream version is kept" distro="libheif1/1.23.4-rc1-1" -- v1.23.4
expect_exit 1
expect_no_request
expect_output "newer than the 1.23.4"

# Which packages count. Anything whose files are unpacked is loadable, whatever dpkg calls it.
run_case "a package awaiting configuration still counts" \
  distro="libheif1/1.24.1-1/unpacked" -- v1.23.4
expect_exit 1
expect_no_request
expect_output "newer than the 1.23.4"
run_case "a package awaiting triggers still counts" \
  distro="libheif1/1.24.1-1/triggers-awaited" -- v1.23.4
expect_exit 1
expect_no_request
expect_output "newer than the 1.23.4"
run_case "a half-configured package still counts" \
  distro="libheif1/1.24.1-1/half-configured" -- v1.23.4
expect_exit 1
expect_no_request
expect_output "newer than the 1.23.4"
run_case "a removed package whose files are gone does not count" \
  distro="libheif1/1.24.1-1/config-files" -- v1.23.4
expect_request ".tar.gz"
expect_not_output "newer than the 1.23.4"
run_case "a package that was never installed does not count" \
  distro="libheif1/1.24.1-1/not-installed" -- v1.23.4
expect_request ".tar.gz"
expect_not_output "newer than the 1.23.4"
run_case "a version that cannot be compared is ignored" distro="libheif1/20240101-1" -- v1.23.4
expect_request ".tar.gz"
expect_not_output "newer than the 1.23.4"
run_case "our own package is not read as the distribution's" \
  distro="photoprism-libheif/1.24.1-photoprism1" -- v1.23.4
expect_request ".tar.gz"
expect_not_output "newer than the 1.23.4"

# The highest counts, in either listing order, so that "the last one wins" cannot pass for it.
run_case "the highest of several packages decides, listed last" \
  distro="libheif-plugin-aomenc/1.21.2-1 libheif1t64/1.24.1-1" -- v1.23.4
expect_exit 1
expect_no_request
expect_output "newer than the 1.23.4"
run_case "the highest of several packages decides, listed first" \
  distro="libheif1t64/1.24.1-1 libheif-plugin-aomenc/1.21.2-1" -- v1.23.4
expect_exit 1
expect_no_request
expect_output "newer than the 1.23.4"

# Where the comparison does not apply, and where it found nothing - both said out loud, since a
# comparison that did not run reads exactly like one that ran and approved.
run_case "the packaging path replaces the distribution package instead of shadowing it" \
  marker=true codename=jammy distro="libheif1/1.24.1-1build2" -- v1.23.4
expect_request ".tar.gz"
expect_not_output "newer than the 1.23.4"
run_case "a private prefix is not in the loader's path, so it shadows nothing" \
  destdir="${WORK_DIR}/prefix" distro="libheif1/1.24.1-1build2" -- v1.23.4
expect_request ".tar.gz"
expect_not_output "newer than the 1.23.4"
run_case "a system with no packaged libheif is told so" -- v1.23.4
expect_request ".tar.gz"
expect_output "shadows nothing"

if [[ ${CHECKED} -ne ${EXPECTED_CASES} ]]; then
  echo "Expected ${EXPECTED_CASES} cases but ran ${CHECKED}; update EXPECTED_CASES deliberately." 1>&2
  FAILED=$((FAILED + 1))
fi

if [[ ${FAILED} -gt 0 ]]; then
  echo "Found ${FAILED} problem(s) in how install-libheif.sh decides." 1>&2
  exit 1
fi

echo "Libheif install decisions checked (${CHECKED} case(s))."
