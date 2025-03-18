#!/usr/bin/env bash

# Installs NodeJS, NPM and TestCafe on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-nodejs.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

set -e

. /etc/os-release

# NodeJS version to be installed.
NODE_MAJOR=22

# Check if NodeJS is installed.
if which node > /dev/null
then
  echo "NodeJS is already installed."
else
  echo "Installing NodeJS and NPM from deb.nodesource.com..."

  # Download the signature key to "/etc/apt/keyrings/nodesource.gpg".
  sudo mkdir -p /etc/apt/keyrings
  curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg

  # Add node repository source to "/etc/apt/sources.list.d/nodesource.list".
  echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_$NODE_MAJOR.x nodistro main" | sudo tee /etc/apt/sources.list.d/nodesource.list

  sudo apt-get update && sudo apt-get -qq install nodejs
fi

# Check if NPM is installed.
if which npm > /dev/null
then
    echo "NPM is already installed."
else
  echo "NPM is required to install these packages".
  exit 1
fi

# Upgrade NPM and install development dependencies.
echo "Configuring NPM..."
sudo npm config set cache ~/.cache/npm
echo "Updating NPM..."
sudo npm update --unsafe-perm=true --allow-root -g npm
echo "Installing TestCafe..."
sudo npm install --unsafe-perm=true --allow-root -g \
  npm@latest npm-check-updates@latest license-report@latest n@latest testcafe@3.7.2
echo "Installing ESLint..."
sudo npm install --unsafe-perm=true --allow-root -g \
  eslint@8 eslint-config-prettier eslint-config-standard eslint-formatter-pretty \
  eslint-plugin-html eslint-plugin-import eslint-plugin-node eslint-plugin-prettier \
  eslint-plugin-promise eslint-plugin-vue eslint-webpack-plugin prettier
echo "Installing Vue Language Tools..."
sudo npm install --unsafe-perm=true --allow-root -g \
  @vue/language-server @vue/typescript-plugin
echo "Done."