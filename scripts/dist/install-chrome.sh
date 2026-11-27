#!/usr/bin/env bash

# Installs a Chromium-based browser on Linux:
#   - AMD64: Google Chrome stable (apt repo at dl.google.com).
#   - ARM64: Google Chrome stable from the same apt repo (default), on Ubuntu
#     and Debian, or Chromium as a fallback, see CHROMIUM_SOURCE below.
#
# Override the ARM64 browser source via env var (unset or empty selects
# google; the develop Dockerfiles accept it as a build argument, e.g.
# --build-arg CHROMIUM_SOURCE=internal):
#   CHROMIUM_SOURCE=google    (default) — Google Chrome stable, https://dl.google.com/linux/chrome/deb/
#   CHROMIUM_SOURCE=internal            — Ubuntu: Chromium from https://dl.photoprism.app/dist/chromium/
#   CHROMIUM_SOURCE=xtradeb             — Ubuntu: Chromium from https://ppa.launchpadcontent.net/xtradeb/apps/ubuntu
#   CHROMIUM_SOURCE=debian              — Debian: Chromium from the distribution's own repository
# The internal mirror is a small reprepro tree of the XtraDeb PPA's chromium /
# chromium-driver / chromium-sandbox .debs (BSD-1-Clause), kept so builds
# survive Launchpad outages. Ubuntu's own chromium-browser is a snap-shim
# transitional package and unusable in Docker.
# Upstream PPA: https://launchpad.net/~xtradeb/+archive/ubuntu/apps
#
# Chromium package set installed by the fallbacks (named
# explicitly on the apt-get install line so a missing piece fails fast and visibly):
#   - chromium                   — the browser binary
#   - chromium-common            — shared resources (strict-version Depends of chromium)
#   - chromium-driver            — chromedriver for headless / Selenium / TestCafe
#   - chromium-sandbox           — setuid sandbox helper
#
# Runtime library dependencies (transitively pulled in by apt). The exact
# package names drift between Ubuntu releases (jammy/noble/questing/resolute);
# the categories below stay stable. Confirm against
#   `dpkg-deb -f chromium_*.deb Depends`
# on the relevant codename before bumping a base image.
#   - GTK / GUI:        libgtk-3-0[t64], libxnvctrl0, libxrandr2, libxkbcommon0,
#                       libxcomposite1, libxdamage1, libxfixes3, libxext6,
#                       libxcb1, libx11-6, libpango-1.0-0, libcairo2
#   - Accessibility:    libatk1.0-0[t64], libatk-bridge2.0-0[t64], libatspi2.0-0[t64]
#   - GPU / EGL:        libgbm1 (>= 21.1.0), mesa-libgallium (transitive)
#   - Crypto / TLS:     libnss3 (>= 2:3.30), libnspr4
#   - Audio:            libasound2[t64], libpulse0, libopus0, libflac8|12|14
#   - Codecs / images:  libdav1d5|7, libopenh264-6|7|8, libopenjp2-7,
#                       liblcms2-2, libjpeg8, libfreetype6
#   - Compression:      libzstd1 (>= 1.4 / 1.5.5), zlib1g, libminizip1[t64]
#   - DBus / udev:      libdbus-1-3, libudev1
#   - System:           libc6, libgcc-s1, libcups2[t64], libfontconfig1,
#                       libdouble-conversion3, libexpat1, libgraphite2-3,
#                       libglib2.0-0[t64], libharfbuzz0b, libharfbuzz-subset0
#
# In our build images these come from the standard Ubuntu archive; the slim
# variants of `photoprism/develop:*` deliberately exclude many of them, which
# is why chromium installs there fail with "libbsd0:arm64 is selected for
# removal" or similar dep-resolver errors. Install chromium only into the
# full (non-slim) develop images.
#
# After installing the browser, the script adds a managed policy file that
# disables Chrome's generative AI features, including the on-device model
# download, so automated test runs see the same browser on every build.
# Set CHROME_AI_POLICY=skip to leave the policies out.
#
# This script must run as root. Use one of these invocations:
#
#   # Pipe via stdin (recommended one-liner — works everywhere, incl. SSH):
#   curl -fsSL https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-chrome.sh | sudo bash
#
#   # Or download first and run:
#   curl -fsSL https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-chrome.sh -o /tmp/install-chrome.sh
#   sudo bash /tmp/install-chrome.sh
#
# Do NOT use process substitution under sudo (`sudo bash <(curl …)`): the
# substitution opens /dev/fd/63 in the *parent* (unprivileged) shell, which
# the elevated bash cannot read once sudo re-execs — the script aborts
# immediately with `bash: /dev/fd/63: No such file or directory`.

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Usage: ${0##*/} must run as root. Try:" 1>&2
  echo "  curl -fsSL <url-to-this-script> | sudo bash" 1>&2
  exit 1
fi

# Determine target architecture.
if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

DESTARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

# shellcheck source=/dev/null
. /etc/os-release

# Adds PhotoPrism's internal chromium mirror as an apt source and installs
# chromium from it. The mirror tracks the XtraDeb PPA.
install_chromium_from_internal_mirror() {
  local keyring=/etc/apt/keyrings/photoprism-apt.gpg
  local src=/etc/apt/sources.list.d/photoprism-chromium.sources

  install -m 0755 -d /etc/apt/keyrings
  # PhotoPrism APT signing key (master): 99F2 9643 6E34 A5DD 1782  A5C9 7FE2 4EBF 235B EBF9
  curl -fsSL "https://dl.photoprism.app/dist/chromium/photoprism-apt.gpg.asc" \
    | gpg --no-tty --batch --yes --dearmor -o "$keyring"

  cat > "$src" <<EOF
Types: deb
URIs: https://dl.photoprism.app/dist/chromium
Suites: ${VERSION_CODENAME}
Components: main
Signed-By: ${keyring}
EOF

  apt-get update
  apt-get -qq install -y --no-install-recommends \
    chromium chromium-common chromium-driver chromium-sandbox
}

# Adds the XtraDeb PPA apt source and installs chromium from it. Fallback
# path used only when CHROMIUM_SOURCE=xtradeb.
install_chromium_from_xtradeb_ppa() {
  local keyring=/etc/apt/keyrings/xtradeb-apps.gpg
  local src=/etc/apt/sources.list.d/xtradeb-apps.sources

  install -m 0755 -d /etc/apt/keyrings
  # PPA signing key fingerprint: 5301FA4FD93244FBC6F6149982BB6851C64F6880
  curl -fsSL "https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x5301FA4FD93244FBC6F6149982BB6851C64F6880" \
    | gpg --no-tty --batch --yes --dearmor -o "$keyring"

  cat > "$src" <<EOF
Types: deb
URIs: https://ppa.launchpadcontent.net/xtradeb/apps/ubuntu
Suites: ${VERSION_CODENAME}
Components: main
Signed-By: ${keyring}
EOF

  apt-get update
  apt-get -qq install -y --no-install-recommends \
    chromium chromium-common chromium-driver chromium-sandbox
}

# Writes a managed policy file that disables generative AI features to the given policy directory.
install_ai_policies() {
  local dir="$1/managed"

  if [[ ${CHROME_AI_POLICY:-} == "skip" ]]; then
    echo "Skipping AI feature policies (CHROME_AI_POLICY=skip)."
    return 0
  fi

  install -m 0755 -d "$dir"
  cat > "$dir/disable-ai.json" <<EOF
{
  "GenAILocalFoundationalModelSettings": 1,
  "BuiltInAIAPIsEnabled": false,
  "AIModeSettings": 1,
  "AutofillPredictionSettings": 2,
  "ChromeSuggestionsSettings": 1,
  "CreateThemesSettings": 2,
  "DevToolsGenAiSettings": 2,
  "GeminiActOnWebSettings": 1,
  "GeminiSparkSettings": 1,
  "HelpMeWriteSettings": 2,
  "HistorySearchSettings": 2,
  "SearchContentSharingSettings": 1,
  "SmartTabSharingSettings": 1,
  "TabCompareSettings": 2,
  "ThirdPartyAiChatSettings": 1,
  "VoiceTypingSettings": 2
}
EOF
  chmod 0644 "$dir/disable-ai.json"
  echo "Installed AI feature policies in $dir."
}

# Adds Google's Chrome apt repo for the given Debian architecture (amd64 or arm64) and installs Chrome stable.
# The repo is bootstrapped just far enough to install the package; its postinst then replaces the bootstrap list
# with its own source and keyring.
install_google_chrome() {
  local arch="$1"
  local keyring="/etc/apt/keyrings/dl-ssl.google.com.gpg"

  echo "Installing Google Chrome (stable) on ${ID} for ${DESTARCH^^}..."
  # Once installed, the package keeps its own source in google-chrome.sources, so a re-run must not add a second one.
  if [[ ! -f /etc/apt/sources.list.d/google-chrome.sources ]]; then
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://dl-ssl.google.com/linux/linux_signing_key.pub | gpg --no-tty --batch --yes --dearmor -o "$keyring"
    echo "deb [arch=${arch} signed-by=${keyring}] https://dl.google.com/linux/chrome/deb/ stable main" > /etc/apt/sources.list.d/google-chrome.list
  fi
  apt-get update
  apt-get -qq install google-chrome-stable
  install_ai_policies /etc/opt/chrome/policies
}

case $DESTARCH in
  amd64 | AMD64 | x86_64 | x86-64)
    set -e
    install_google_chrome amd64
    ;;

  arm64 | ARM64 | aarch64)
    set -e
    case $ID in
      debian)
        case "${CHROMIUM_SOURCE:-google}" in
          google)
            apt-get update
            apt-get -qq install -y --no-install-recommends ca-certificates curl gnupg
            install_google_chrome arm64
            ;;
          debian)
            echo "Installing Chromium on ${ID} for ${DESTARCH^^}..."
            apt-get update
            apt-get -qq install chromium chromium-common chromium-driver chromium-sandbox
            install_ai_policies /etc/chromium/policies
            ;;
          *)
            echo "Unknown CHROMIUM_SOURCE='${CHROMIUM_SOURCE}' (expected 'google' or 'debian')" 1>&2
            exit 1
            ;;
        esac
        ;;

      ubuntu)
        apt-get -qq install -y --no-install-recommends ca-certificates curl gnupg
        case "${CHROMIUM_SOURCE:-google}" in
          google)
            install_google_chrome arm64
            ;;
          internal)
            echo "Installing Chromium (via PhotoPrism internal mirror) on ${ID} ${VERSION_CODENAME:-} for ${DESTARCH^^}..."
            install_chromium_from_internal_mirror
            install_ai_policies /etc/chromium/policies
            ;;
          xtradeb)
            echo "Installing Chromium (via XtraDeb PPA — fallback) on ${ID} ${VERSION_CODENAME:-} for ${DESTARCH^^}..."
            install_chromium_from_xtradeb_ppa
            install_ai_policies /etc/chromium/policies
            ;;
          *)
            echo "Unknown CHROMIUM_SOURCE='${CHROMIUM_SOURCE}' (expected 'google', 'internal' or 'xtradeb')" 1>&2
            exit 1
            ;;
        esac
        ;;

      *)
        echo "Unsupported distribution \"${ID}\" for ARM64 browser install" 1>&2
        exit 1
        ;;
    esac
    ;;

  *)
    echo "Unsupported Machine Architecture: \"$DESTARCH\"" 1>&2
    exit 0
    ;;
esac

echo "Done."
