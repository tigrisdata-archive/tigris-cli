#!/bin/bash
# Copyright 2022 Tigris Data, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -ex

LINK_ID=(lnk_1EY9_24JMe7 lnk_1EY9_24JIYa lnk_1EY9_217mHr lnk_1EY9_216bS8 lnk_1EY9_219th2 lnk_1EY9_2Z2L8m)
OS=(linux_arm64 darwin_arm64 linux_amd64 darwin_amd64 windows_amd64 windows_arm64)

for i in $(seq 0 5); do
	L=${LINK_ID[i]}
	O=${OS[i]}

	EXT="tar.gz"
	if [[ "$O" == "windows"* ]]; then
		EXT="zip"
	fi

	#Check that link we are about to publish is live
	curl -f -s -I "https://github.com/tigrisdata/tigris-cli/releases/download/${VERSION}/tigris_${VERSION#v}_$O.$EXT"

	#Push to Shorts.io
	curl -f -s -X POST \
		-H "authorization: ${API_KEY}" \
		-H "content-type: application/json" \
		-d '{"domain": "tigris.dev", "originalURL": "https://github.com/tigrisdata/tigris-cli/releases/download/'"${VERSION}"'/tigris_'"${VERSION#v}"'_'"$O"'.'"$EXT"'" }' \
		https://api.short.io/links/"$L"/
done
