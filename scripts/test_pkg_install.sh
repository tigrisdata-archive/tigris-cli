#!/bin/bash

set -ex

VERSION=${GITHUB_REF_NAME}

if [ -z "$VERSION" ]; then
	echo "Set VERSION variable to test installation of"
fi

#Delete v prefix from v1.0.0-beta.X
VERSION=${VERSION/#v}

if [ -n "$UBUNTU" ]; then

tigris version && exit 1 # expected to fail
sudo snap install tigris
which tigris
tigris version | grep -e "$VERSION"
sudo snap remove tigris

hash -r

fi

if [ -n "$UBUNTU" ] || [ -n "$MACOS" ]; then

brew=/home/linuxbrew/.linuxbrew/bin/brew
if [ -n "$MACOS" ]; then
  brew=/usr/local/bin/brew
fi

curl -fsSL -o install.sh https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh
CI=1 /bin/bash install.sh
eval "$($brew shellenv)"

tigris version && exit 1 # expected to fail
$brew install tigrisdata/tigris/tigris-cli
which tigris
tigris version
tigris version | grep -e "$VERSION"
$brew uninstall tigrisdata/tigris/tigris-cli

hash -r

fi

if [ -n "$UBUNTU" ] || [ -n "$MACOS" ]; then
  SUDO=sudo
fi

# Test local NPM install
npx tigris version && exit 1 # expected to fail
npm install @tigrisdata/tigris-cli
npx tigris version
npx tigris version | grep -e "$VERSION"
npm uninstall @tigrisdata/tigris-cli

# Test global NPM install
npx tigris version && exit 1 # expected to fail
tigris version && exit 1 # expected to fail
$SUDO npm install -g @tigrisdata/tigris-cli
tigris version
tigris version | grep -e "$VERSION"
$SUDO npm uninstall -g @tigrisdata/tigris-cli

hash -r

npx tigris version && exit 1 # expected to fail
tigris version && exit 1 # expected to fail

exit 0
