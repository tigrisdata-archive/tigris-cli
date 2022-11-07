#!/bin/bash

set -x

for i in linux_amd64 linux_arm64 darwin_amd64 darwin_arm64 windows_amd64 windows_arm64; do
	ext="tar.gz"
	if [ $i = "windows_amd64" ] || [ $i = "windows_arm64" ]; then
		ext="zip"
	fi

	out=$(grep $i.$ext ../../dist/checksums.txt|cut -f 1 -d ' ')

	sed -i "s/{{${i}_checksum}}/$out/" package.json
done

