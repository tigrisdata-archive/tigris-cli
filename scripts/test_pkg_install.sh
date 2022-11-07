#!/bin/bash

set -ex

if [ -z "$VERSION" ]; then
	echo "Set VERSION variable to test installation of"
fi

! tigris version
sudo snap install tigris
which tigris
tigris version | grep $VERSION
sudo snap remove tigris

hash -r

curl -fsSL -o install.sh https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh
CI=1 /bin/bash install.sh
eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"

! tigris version
/home/linuxbrew/.linuxbrew/bin/brew install tigrisdata/tigris/tigris-cli
which tigris
tigris version
tigris version | grep $VERSION
/home/linuxbrew/.linuxbrew/bin/brew uninstall tigrisdata/tigris/tigris-cli

! npx tigris version
npm install @tigrisdata/tigris-cli
npx tigris version
npx tigris version | grep $VERSION
npm uninstall @tigrisdata/tigris-cli

! tigris version
npm install -g @tigrisdata/tigris-cli
npx tigris version
npx tigris version | grep $VERSION
npm uninstall -g @tigrisdata/tigris-cli

! tigris version

