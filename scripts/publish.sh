#!/bin/bash

set -ex

LINK_ID=(lnk_1EY9_24JMe7 lnk_1EY9_24JIYa lnk_1EY9_217mHr lnk_1EY9_216bS8 lnk_1EY9_219th2 lnk_1EY9_2Z2L8m)
OS=(linux-arm64 darwin-arm64 linux-amd64 darwin-amd64 windows-amd64 windows-arm64)

for i in $(seq 0 5); do
	L=${LINK_ID[i]}
	O=${OS[i]}

	EXT="tar.gz"
	if [[ "$O" == "windows"* ]]; then
		EXT="zip"
	fi

	#Check that link we are about to publish is live
	curl -f -s -I "https://github.com/tigrisdata/tigris-cli/releases/download/${VERSION}/tigris-${VERSION}-$O.$EXT"

	#Push to Shorts.io
	curl -f -s -X POST \
		-H "authorization: ${API_KEY}" \
		-H "content-type: application/json" \
		-d '{"domain": "tigris.dev", "originalURL": "https://github.com/tigrisdata/tigris-cli/releases/download/'"${VERSION}"'/tigris-'"${VERSION}"'-'"$O"'.'"$EXT"'" }' \
		https://api.short.io/links/"$L"/
done
