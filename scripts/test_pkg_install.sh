#!/bin/bash

set -ex

if [ -z "$VERSION" ]; then
	echo "Set VERSION variable to test installation of"
fi

#Delete v prefix from v1.0.0-beta.X
VERSION=${VERSION/#v}

tigris version && exit 1
sudo snap install tigris
which tigris
tigris version | grep -e "$VERSION"
sudo snap remove tigris

hash -r

curl -fsSL -o install.sh https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh
CI=1 /bin/bash install.sh
eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"

tigris version && exit 1
/home/linuxbrew/.linuxbrew/bin/brew install tigrisdata/tigris/tigris-cli
which tigris
tigris version
tigris version | grep -e "$VERSION"
/home/linuxbrew/.linuxbrew/bin/brew uninstall tigrisdata/tigris/tigris-cli

hash -r

# Test local NPM install
npx tigris version && exit 1
npm install @tigrisdata/tigris-cli
npx tigris version
npx tigris version | grep -e "$VERSION"
npm uninstall @tigrisdata/tigris-cli

# Test global NPM install
npx tigris version && exit 1
tigris version && exit 1
npm install -g @tigrisdata/tigris-cli
tigris version
tigris version | grep -e "$VERSION"
npm uninstall -g @tigrisdata/tigris-cli

hash -r

npx tigris version && exit 1
tigris version && exit 1

