#!/usr/bin/env bash

# abort if the user is not root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run install-devtools.sh as root" 1>&2
  exit 1
fi

set -e

if [[ -z $1 ]]; then
    echo "Usage: install-devtools.sh [amd64|arm64|arm]" 1>&2
    exit 1
else
    set -eux;
    umask 0000

    # Install Chrome or Chromium
    if [[ $1 == "amd64" ]]; then
        wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -
        sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list'
        apt-get update
        apt-get -qq install google-chrome-stable
    elif [[ $1 == "arm64" ]]; then
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
    if [[ $1 == "amd64" ]]; then
        npm install --unsafe-perm=true --allow-root -g puppeteer testcafe testcafe-browser-provider-puppeteer chromedriver
    elif [[ $1 == "arm64" ]]; then
        npm install --unsafe-perm=true --allow-root -g testcafe chromedriver
    fi
fi

