#!/usr/bin/env bash

# abort if the user is not root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run install-devtools.sh as root" 1>&2
  exit 1
fi

. /etc/os-release

if [[ $ID != "ubuntu" ]]; then
  echo "Dev-Tools only need to be installed on Ubuntu Linux."
  exit
fi

set -e

SYSTEM_ARCH=$("$(dirname "$0")/arch.sh")
INSTALL_ARCH=${2:-$SYSTEM_ARCH}

echo "Installing Dev-Tools for ${INSTALL_ARCH^^}..."

set -eux;
umask 0000

# Install Chrome or Chromium
if [[ $INSTALL_ARCH == "amd64" ]]; then
    wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -
    sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list'
    apt-get update
    apt-get -qq install google-chrome-stable
elif [[ $INSTALL_ARCH == "arm64" ]]; then
cat <<EOF > /etc/apt/preferences.d/chromium
Package: *
Pin: release o=LP-PPA-saiarcot895-chromium-dev
Pin-Priority: 1002
EOF
    add-apt-repository -y ppa:saiarcot895/chromium-dev
    apt-get update
    apt-get -qq install chromium-browser chromium-codecs-ffmpeg-extra
fi

# Remove package files
apt-get -y autoremove
apt-get -y autoclean
apt-get -y clean
rm -rf /var/lib/apt/lists/*

# Install Puppeteer, TestCafe & ChromeDriver
if [[ $INSTALL_ARCH == "amd64" ]]; then
    npm install --unsafe-perm=true --allow-root -g puppeteer testcafe testcafe-browser-provider-puppeteer chromedriver
elif [[ $INSTALL_ARCH == "arm64" ]]; then
    npm install --unsafe-perm=true --allow-root -g testcafe chromedriver
fi

echo "Done."