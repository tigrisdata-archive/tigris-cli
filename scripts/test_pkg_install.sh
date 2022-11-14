#!/bin/bash

set -ex

if [ -z "$VERSION" ]; then
	echo "Set VERSION variable to test installation of"
fi

tigris version && exit 1
sudo snap install tigris
which tigris
tigris version | grep -e "v1.0.0-beta"
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

npx tigris version && exit 1
npm install @tigrisdata/tigris-cli
npx tigris version || true
npx tigris version | grep -e "$VERSION" || true
npm uninstall @tigrisdata/tigris-cli

npx tigris version && exit 1
npm install -g @tigrisdata/tigris-cli
npx tigris version || true
npx tigris version | grep -e "$VERSION" || true
npm uninstall -g @tigrisdata/tigris-cli

npx tigris version && exit 1
tigris version || true
tigris version && exit 1

