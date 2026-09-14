#!/usr/bin/env bash

# Installs the heif-dec, heif-enc, and heif-info binaries on Linux.
# On libheif 1.21+, heif-convert is a symlink to heif-dec and heif-thumbnailer is no longer shipped.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-libheif.sh)
#
# By default the binaries are installed under /usr/local alongside the distribution packages,
# which is what that prefix is for: no file conflicts anything, and the loader prefers them.
#
#   --deb        install through apt, replacing the distribution libheif packages and anything
#                that depends on them (default: install under /usr/local alongside them)
#   --downgrade  allow installing a version older than the one already in the prefix, or
#                older than the distribution's own libheif packages
#
# --deb is what our own images use, and they select it through PHOTOPRISM_CONTAINER rather
# than by passing the flag. It exists so an image carries one libheif rather than two: Resolute
# installs photoprism-libheif via a real .deb that Provides/Replaces/Conflicts apt's libheif1,
# libheif-dev, libheif-examples and libheif-plugin-* names, and other Debian distros get a
# contentless stub with the same metadata after the tarball extract. Either way the package is
# apt-mark held so dist-upgrade leaves it alone - but an explicit `apt install libheif1` would
# still trigger the Conflicts: relation and replace photoprism-libheif, silently breaking the
# from-source HEIC pipeline. Don't do that; upgrade by bumping LIBHEIF_VERSION and rerunning
# this script with the same path selection it was installed with.

set -e

# USE_DEB selects the packaging path; DOWNGRADE overrides the version check. Neither implies
# the other. PHOTOPRISM_CONTAINER marks our own images, which are the images the packaging
# path exists for.
USE_DEB=0
DOWNGRADE=0
SHOW_HELP=0
ARGS=()

if [[ ${PHOTOPRISM_CONTAINER:-} == "true" ]]; then
  USE_DEB=1
fi

for arg in "$@"; do
  case $arg in
    --deb) USE_DEB=1 ;;
    --downgrade) DOWNGRADE=1 ;;
    --help) SHOW_HELP=1 ;;
    -*)
      echo "Error: unknown option '${arg}'. Run with --help for the available ones." 1>&2
      exit 1
      ;;
    *) ARGS+=("$arg") ;;
  esac
done

set -- "${ARGS[@]+"${ARGS[@]}"}"

if [[ $# -gt 2 ]]; then
  echo "Error: expected at most [destdir] [version]. Run with --help." 1>&2
  exit 1
fi

# Show usage information when --help was given.
if [[ ${SHOW_HELP} == 1 ]]; then
  cat 1>&2 <<'USAGE'
Usage: install-libheif.sh [--deb] [--downgrade] [destdir] [version]

  --deb        install through apt, replacing the distribution libheif packages and
               anything that depends on them (default: install under /usr/local
               alongside them)
  --downgrade  allow installing a version older than the one already in the prefix,
               or older than the distribution's own libheif packages
USAGE
  exit 0
fi

# You can provide a custom installation directory as the first argument.
DESTDIR=$(realpath "${1:-/usr/local}")

# In addition, you can specify a custom version to be installed as the second argument.
LIBHEIF_VERSION=${2:-v1.23.4}

# A private staging directory this script owns.
if ! STAGING_DIR="$(mktemp -d "${TMPDIR:-/tmp}/photoprism-libheif.XXXXXXXX")"; then
  echo "❌ Cannot create a staging directory in \"${TMPDIR:-/tmp}\"."
  exit 1
fi

# STUB_DIR is set later, by the branch that builds a replacement package. One handler owns both,
# because an EXIT trap replaces the previous one rather than stacking with it.
STUB_DIR=""

cleanup() {
  [[ -n ${STAGING_DIR} ]] && rm -rf "${STAGING_DIR}"
  [[ -n ${STUB_DIR} ]] && rm -rf "${STUB_DIR}"
  return 0
}

trap cleanup EXIT

# verify_artifact checks a staged download against its published SHA-256 checksum.
#
# Artifacts are served with a ".sha256" companion where one has been published; a checksum that
# exists must match, and one that is served but malformed is refused rather than ignored.
verify_artifact() {
  local file="$1" url="$2" sums="$1.sha256" expected actual

  if ! curl -fsSL "${url}.sha256" -o "${sums}" 2> /dev/null; then
    echo "⚠️ No published checksum for \"${url}\"; installing unverified."
    return 0
  fi

  expected=$(awk '{print $1; exit}' "${sums}")

  if [[ ! ${expected} =~ ^[0-9a-fA-F]{64}$ ]]; then
    echo "❌ Published checksum for \"${url}\" is not a SHA-256 digest."
    return 1
  fi

  if command -v sha256sum > /dev/null 2>&1; then
    actual=$(sha256sum "${file}" | awk '{print $1}')
  else
    actual=$(shasum -a 256 "${file}" | awk '{print $1}')
  fi

  if [[ ${expected} != "${actual}" ]]; then
    echo "❌ Checksum mismatch for \"${url}\"."
    return 1
  fi

  echo "✅ Checksum OK (${actual})."
}

# newest prints the higher of two versions, which is the one "sort -V" puts last.
newest() {
  printf '%s\n%s\n' "$1" "$2" | sort -V | tail -1
}

# distro_libheif_version prints the highest upstream version among the distribution libheif
# packages whose files are on disk, and nothing when dpkg is absent or none of them is.
distro_libheif_version() {
  local highest="" name version state

  command -v dpkg-query > /dev/null 2>&1 || return 0

  while read -r name version state; do
    # Every state but these two leaves the shared libraries unpacked, so the loader finds them
    # whether or not the package is configured. Naming the two that do not is therefore the
    # reliable direction: an allow-list of "installed" misses a half-configured package that is
    # every bit as loadable.
    [[ ${state} != "not-installed" && ${state} != "config-files" ]] || continue

    # Defensive: our package is named so that dpkg's "libheif*" does not select it.
    [[ ${name} != "photoprism-libheif" ]] || continue

    # A dpkg version carries an epoch, a Debian revision, and sometimes a repacking suffix
    # around the upstream one. A "+dfsg" repack is the same upstream release, so leaving it on
    # would read as newer and refuse an install of the very version already packaged.
    version=${version#*:}
    version=${version%-*}
    version=${version%%+*}

    [[ ${version} =~ ^[0-9]+\.[0-9] ]] || continue

    if [[ -z ${highest} ]] || [[ $(newest "${highest}" "${version}") == "${version}" ]]; then
      highest=${version}
    fi
  done < <(dpkg-query -W -f='${Package} ${Version} ${db:Status-Status}\n' 'libheif*' 2> /dev/null || true)

  printf '%s' "${highest}"
}

# Determine target architecture.
if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

DESTARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

case $DESTARCH in
  amd64 | AMD64 | x86_64 | x86-64)
    DESTARCH=amd64
    ;;

  arm64 | ARM64 | aarch64)
    DESTARCH=arm64
    ;;

  arm | ARM | aarch | armv7l | armhf)
    DESTARCH=arm
    ;;

  *)
    echo "Unsupported Machine Architecture: \"$DESTARCH\"" 1>&2
    exit 1
    ;;
esac

# shellcheck source=/dev/null
. /etc/os-release

# Abort if not executed as root.
if [[ $(id -u) != "0" ]] && [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Error: Run ${0##*/} as root to install in a system directory!" 1>&2
  exit 1
fi

# /usr is dpkg's own prefix, so an archive unpacked there is an unowned library in a directory
# the package manager believes it manages. /usr/local is the prefix meant for exactly this.
if [[ $DESTDIR == "/usr" ]] && [[ $USE_DEB == 0 ]]; then
  echo "Error: installing into /usr without --deb would leave files no package owns." 1>&2
  echo "       Use the default /usr/local prefix, or pass --deb to install through apt." 1>&2
  exit 1
fi

# A host already carrying our package, installing without the packaging path, ends up with the
# library coming from the prefix while dpkg still reports the packaged version. Say so: only the
# operator can decide between reinstalling the distribution package and rerunning with --deb.
if [[ $USE_DEB == 0 ]] && command -v dpkg > /dev/null && dpkg -s photoprism-libheif > /dev/null 2>&1; then
  echo "ℹ️ photoprism-libheif is installed through apt on this host, and this run installs into" 1>&2
  echo "   \"${DESTDIR}\" instead. The version dpkg reports will not track the one in use." 1>&2
fi

mkdir -p "$DESTDIR"

# Map codenames to find and use a compatible version.
case $VERSION_CODENAME in
  vera | virginia)
    VERSION_CODENAME=jammy
    ;;
esac

# Installing under /usr/local puts our build ahead of the distribution one for everything that
# loads libheif, which is what that prefix is for and is correct while ours is newer. Going
# backwards is the failure worth catching: a stale copy left in the prefix keeps shadowing a
# newer distribution library for every consumer, and neither apt-get check nor dpkg -C reports
# anything, because the package manager never owned the file.
if [[ -e ${DESTDIR}/lib/libheif.so.1 || -L ${DESTDIR}/lib/libheif.so.1 ]]; then
  EXISTING_SO=$(readlink -f "${DESTDIR}/lib/libheif.so.1" 2>/dev/null || true)
  EXISTING_VERSION=${EXISTING_SO##*/libheif.so.}
  TARGET_VERSION=${LIBHEIF_VERSION#v}

  # Only a dotted version is comparable; anything else is a layout we did not write, and
  # guessing at it would be worse than saying so.
  if [[ -n ${EXISTING_VERSION} ]] && [[ ${EXISTING_VERSION} == "${EXISTING_SO}" || ! ${EXISTING_VERSION} =~ ^[0-9]+\.[0-9][0-9A-Za-z.+~-]*$ ]]; then
    echo "ℹ️ ${DESTDIR}/lib/libheif.so.1 carries no recognizable version; installing ${LIBHEIF_VERSION#v} over it." 1>&2
  elif [[ -n ${EXISTING_VERSION} ]] && [[ ${EXISTING_VERSION} != "${TARGET_VERSION}" ]]; then
    NEWEST=$(newest "${EXISTING_VERSION}" "${TARGET_VERSION}")

    if [[ ${NEWEST} == "${EXISTING_VERSION}" ]] && [[ ${DOWNGRADE} == 0 ]]; then
      echo "❌ ${DESTDIR}/lib already has libheif ${EXISTING_VERSION}, which is newer than ${TARGET_VERSION}." 1>&2
      echo "   Installing it would leave everything on this system resolving to the older build," 1>&2
      echo "   which no package manager reports. Pass --downgrade to install it anyway." 1>&2
      exit 1
    fi

    echo "ℹ️ Replacing libheif ${EXISTING_VERSION} in \"${DESTDIR}/lib\" with ${TARGET_VERSION}." 1>&2
  fi
fi

# The same comparison, one step out: the prefix is searched ahead of the distribution's library
# directory, so whatever is written here becomes the libheif every consumer loads. Only the
# prefix path reaches this - the packaging path replaces the distribution package instead.
# The "/usr" arm repeats the system prefixes rather than leaning on the earlier guard that
# already refuses it here, so this condition states which prefixes the loader searches.
if [[ $USE_DEB == 0 ]] && [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  DISTRO_VERSION=$(distro_libheif_version)
  TARGET_VERSION=${LIBHEIF_VERSION#v}

  # A comparison that did not happen reads exactly like one that approved, so say which it was.
  if ! command -v dpkg-query > /dev/null 2>&1; then
    echo "ℹ️ Without dpkg, the packaged libheif version is unknown and was not compared." 1>&2
  elif [[ -z ${DISTRO_VERSION} ]]; then
    echo "ℹ️ No packaged libheif is installed, so ${TARGET_VERSION} shadows nothing." 1>&2
  elif [[ ${DISTRO_VERSION} != "${TARGET_VERSION}" ]] &&
    [[ $(newest "${DISTRO_VERSION}" "${TARGET_VERSION}") == "${DISTRO_VERSION}" ]]; then
    if [[ ${DOWNGRADE} == 0 ]]; then
      echo "❌ The distribution's libheif is at ${DISTRO_VERSION}, newer than the ${TARGET_VERSION} this installs." 1>&2
      echo "   \"${DESTDIR}\" is searched ahead of it, so this would become the libheif every consumer loads." 1>&2
      echo "   In a container image, set PHOTOPRISM_CONTAINER=true or pass --deb to install through apt" 1>&2
      echo "   instead. On a host, bump the version, or pass --downgrade to install this one anyway." 1>&2
      exit 1
    fi

    echo "ℹ️ Shadowing the distribution's libheif ${DISTRO_VERSION} with ${TARGET_VERSION}." 1>&2
  fi
fi

echo "Installing libheif..."

# On Ubuntu 26.04 LTS (Resolute) we ship a real .deb (photoprism-libheif) so
# apt's libheif1 / libheif-dev / libheif-plugin-* don't coexist with the
# from-source build. The .deb only exists when DESTDIR is left at the default
# /usr or /usr/local - dpkg paths are absolute and ignore custom prefixes.
if [[ $USE_DEB == 1 ]] && [[ $VERSION_CODENAME == "resolute" ]] && [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]] && command -v apt-get > /dev/null; then
  ARCHIVE="libheif-${VERSION_CODENAME}-${DESTARCH}-${LIBHEIF_VERSION}.deb"
  URL="https://dl.photoprism.app/dist/libheif/${ARCHIVE}"
  TMPDEB="${STAGING_DIR}/${ARCHIVE}"

  echo "--------------------------------------------------------------------------------"
  echo "VERSION: $LIBHEIF_VERSION"
  echo "PACKAGE: $ARCHIVE (photoprism-libheif Debian package)"
  echo "--------------------------------------------------------------------------------"

  if ! curl -fsSL "$URL" -o "$TMPDEB"; then
    echo "❌ Failed to download \"$URL\"."
    exit 1
  fi

  if ! verify_artifact "$TMPDEB" "$URL"; then
    exit 1
  fi

  # apt reads the package as an unprivileged user and warns when it cannot. The staging
  # directory's name is unpredictable, which is what keeps it from being pre-planted; the
  # contents are a signed-by-checksum package rather than a secret.
  chmod 0755 "$STAGING_DIR"
  chmod 0644 "$TMPDEB"

  # photoprism-libheif is apt-mark held after install, so an explicit version bump
  # must pass --allow-change-held-packages or apt refuses to upgrade the held package.
  export DEBIAN_FRONTEND=noninteractive
  if apt-get install -y --allow-change-held-packages --no-install-recommends "$TMPDEB"; then
    apt-mark hold photoprism-libheif > /dev/null
    rm -f "$TMPDEB"
    echo "✅ Installed photoprism-libheif from \"$URL\"."
    echo "Done."
    exit 0
  fi

  echo "⚠️ apt-get install of \"$TMPDEB\" failed; falling back to tarball path."
  rm -f "$TMPDEB"
fi

ARCHIVE="libheif-${VERSION_CODENAME}-${DESTARCH}-${LIBHEIF_VERSION}.tar.gz"
URL="https://dl.photoprism.app/dist/libheif/${ARCHIVE}"

echo "--------------------------------------------------------------------------------"
echo "VERSION: $LIBHEIF_VERSION"
echo "ARCHIVE: $ARCHIVE"
echo "DESTDIR: $DESTDIR"
echo "--------------------------------------------------------------------------------"

TMPTAR="${STAGING_DIR}/${ARCHIVE}"

# Staged before it is unpacked, so that a transfer which stops partway cannot leave half an
# archive written over DESTDIR, and so the archive can be checked while it is still a file.
#
# A 404 means this platform has no published build, which is not an error: the image goes on
# without libheif. Every other outcome - a name that does not resolve, a refused connection, a
# TLS failure, a proxy error, a truncated transfer, a full disk - is a failure, and reporting it
# as "not available here" would publish an image that silently lacks HEIC support.
HTTP_STATUS=$(curl -sSL -w '%{http_code}' -o "$TMPTAR" "$URL" 2> /dev/null || echo "000")

if [[ $HTTP_STATUS == "404" ]]; then
  echo "❌ No libheif binaries are available for this architecture or distribution."
  exit 0
elif [[ $HTTP_STATUS != "200" ]]; then
  echo "❌ Failed to download \"$URL\" (HTTP $HTTP_STATUS)."
  exit 1
fi

if ! verify_artifact "$TMPTAR" "$URL"; then
  exit 1
fi

# Readable before writable: a listing failure means the archive is damaged, which is a failure
# rather than the "nothing published for this platform" case handled above.
if ! tar -tzf "$TMPTAR" > /dev/null 2>&1; then
  echo "❌ Downloaded \"$URL\" is not a readable archive."
  exit 1
fi

# ldconfig resolves a soname to the highest version it finds in the directory, so a previous
# install left beside this one would keep winning and the extracted build would never take
# effect - silently, since the link it repoints looks correct. The archive carries the whole
# set, so removing the versioned files first leaves exactly one candidate.
rm -f "${DESTDIR}"/lib/libheif.so.1.*

if ! tar --overwrite --mode=755 -xzf "$TMPTAR" -C "$DESTDIR"; then
  echo "❌ Failed to extract \"$URL\" to \"$DESTDIR\"."
  exit 1
fi

echo "✅ Extracted \"$URL\" to \"$DESTDIR\""

if [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Running \"ldconfig\"..."
  ldconfig
else
  echo "Running \"ldconfig -n $DESTDIR/lib\"..."
  ldconfig -n "$DESTDIR/lib"
fi

# Replace any distro-shipped libheif* packages with a contentless photoprism-libheif
# stub that Provides/Replaces/Conflicts them. Apt's solver then accepts the in-image
# install as a substitute and won't pull the older distro library back in on the next
# dist-upgrade. The actual binaries live under $DESTDIR (typically /usr/local/lib),
# and ldconfig already orders /usr/local/lib ahead of /usr/lib so consumers resolve
# to the from-source build. Resolute exits earlier through the real .deb path above.
if [[ $USE_DEB == 1 ]] && [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]] && command -v apt-get > /dev/null && command -v dpkg-deb > /dev/null; then
  INSTALLED_LIBHEIF_PKGS=$(dpkg-query -W -f='${Package}\n' 'libheif*' 2>/dev/null | grep -v '^photoprism-libheif$' | sort -u || true)
  if [[ -n $INSTALLED_LIBHEIF_PKGS ]]; then
    UPSTREAM_VERSION=${LIBHEIF_VERSION#v}
    STUB_VERSION="${UPSTREAM_VERSION}-photoprism1"
    STUB_ARCH=$(dpkg --print-architecture)
    STUB_DIR=$(mktemp -d)

    # Build comma-separated Provides:/Replaces:/Conflicts: from the actually-installed
    # set, so per-distro naming (libheif1 vs libheif1t64, plugin variants) is handled
    # without hard-coded per-codename branches.
    STUB_PROVIDES=$(printf '%s\n' "$INSTALLED_LIBHEIF_PKGS" | sed "s/\$/ (= $UPSTREAM_VERSION)/" | paste -sd ',' | sed 's/,/, /g')
    STUB_REPLACES=$(printf '%s\n' "$INSTALLED_LIBHEIF_PKGS" | paste -sd ',' | sed 's/,/, /g')

    echo "--------------------------------------------------------------------------------"
    echo "Generating photoprism-libheif stub for apt-installed packages:"
    # shellcheck disable=SC2001
    echo "$INSTALLED_LIBHEIF_PKGS" | sed 's/^/  /'
    echo "--------------------------------------------------------------------------------"

    mkdir -p "$STUB_DIR/DEBIAN" "$STUB_DIR/usr/share/doc/photoprism-libheif"
    cat > "$STUB_DIR/DEBIAN/control" <<STUBEOF
Package: photoprism-libheif
Version: $STUB_VERSION
Architecture: $STUB_ARCH
Multi-Arch: foreign
Maintainer: PhotoPrism UG <hello@photoprism.app>
Section: libs
Priority: optional
Provides: $STUB_PROVIDES
Replaces: $STUB_REPLACES
Conflicts: $STUB_REPLACES
Description: libheif ${LIBHEIF_VERSION} stub for PhotoPrism's from-source install
 Contentless package that satisfies apt's libheif* dependencies after
 install-libheif.sh has placed the real binaries under $DESTDIR. Apt's
 solver treats this as the libheif provider so the distro libheif1 and
 libheif-plugin-* packages are not pulled back in on the next dist-upgrade.
 .
 Real binaries: $DESTDIR/bin/heif-* and $DESTDIR/lib/libheif.so.*
STUBEOF
    cat > "$STUB_DIR/usr/share/doc/photoprism-libheif/copyright" <<STUBEOF
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: libheif
Source: https://github.com/strukturag/libheif

Files: *
Copyright: 2017-2026 struktur AG
License: LGPL-3.0+
 See /usr/share/common-licenses/LGPL-3 for the full text.
STUBEOF

    STUB_DEB="$STUB_DIR/photoprism-libheif_${STUB_VERSION}_${STUB_ARCH}.deb"
    dpkg-deb --build --root-owner-group "$STUB_DIR" "$STUB_DEB" > /dev/null

    export DEBIAN_FRONTEND=noninteractive
    if apt-get install -y --allow-change-held-packages --no-install-recommends "$STUB_DEB"; then
      apt-mark hold photoprism-libheif > /dev/null
      echo "✅ Installed photoprism-libheif stub (apt's libheif1/libheif-plugin-* superseded)."
    else
      echo "⚠️ Failed to install photoprism-libheif stub; apt's libheif packages remain alongside the from-source binaries."
    fi

    rm -rf "$STUB_DIR"
  fi
fi

# Assert the outcome rather than trusting the branch: a marker that did not arm the packaging
# path, or a copy left behind in the prefix, both produce a successful-looking run that resolves
# to a different library than intended. Nothing else in a build would report either.
if command -v ldconfig > /dev/null; then
  RESOLVED=$(ldconfig -p 2>/dev/null | grep -c "libheif\.so\.1" || true)

  if [[ $USE_DEB == 1 ]] && [[ ${RESOLVED} -gt 1 ]]; then
    echo "⚠️ ${RESOLVED} copies of libheif.so.1 are visible to the loader; this path installs one." 1>&2
    ldconfig -p 2>/dev/null | grep "libheif\.so\.1" 1>&2 || true
  elif [[ ${RESOLVED} -eq 0 ]]; then
    echo "⚠️ No libheif.so.1 is visible to the loader after installing." 1>&2
  fi
fi

echo "Done."
