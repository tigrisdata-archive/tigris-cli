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

cli="./tigris"

make

$cli version

$cli local up
$cli local logs >/dev/null 2>&1

$cli server info
$cli server version

db_tests() {
	$cli ping

	$cli drop database db1 || true

	$cli create database db1

	coll1='{"title":"coll1","properties":{"Key1":{"type":"string"},"Field1":{"type":"integer"},"Field2":{"type":"integer"}},"primary_key":["Key1"]}'
	coll111='{"title":"coll111","properties":{"Key1":{"type":"string"},"Field1":{"type":"integer"}},"primary_key":["Key1"]}'

	#reading schemas from command line parameters
	$cli create collection db1 "$coll1" "$coll111"

	out=$($cli describe collection db1 coll1)
	diff -w -u <(echo '{"collection":"coll1","schema":'"$coll1"'}') <(echo "$out")

	out=$($cli describe database db1)
	# The output order is not-deterministic, try both combinations:
	# BUG: Http doesn't fill database name
	diff -u <(echo '{"db":"db1","collections":[{"collection":"coll1","schema":'"$coll1"'},{"collection":"coll111","schema":'"$coll111"'}]}') <(echo "$out") ||
	diff -u <(echo '{"db":"db1","collections":[{"collection":"coll111","schema":'"$coll111"'},{"collection":"coll1","schema":'"$coll1"'}]}') <(echo "$out") ||
	diff -u <(echo '{"collections":[{"collection":"coll111","schema":'"$coll111"'},{"collection":"coll1","schema":'"$coll1"'}]}') <(echo "$out") ||
	diff -u <(echo '{"collections":[{"collection":"coll1","schema":'"$coll1"'},{"collection":"coll111","schema":'"$coll111"'}]}') <(echo "$out")

	#reading schemas from stream
	# \n at the end to test empty line skipping
	# this also test multi-line streams
	echo -e '{ "title" : "coll2",
	"properties": { "Key1": { "type": "string" },
	"Field1": { "type": "integer" }, "Field2": { "type": "integer" } }, "primary_key": ["Key1"] }\n        \n\n' | $cli create collection db1 -
	#reading array of schemas
	echo '[{ "title" : "coll3", "properties": { "Key1": { "type": "string" }, "Field1": { "type": "integer" } }, "primary_key": ["Key1"] }, { "title" : "coll4", "properties": { "Key1": { "type": "string" }, "Field1": { "type": "integer" } }, "primary_key": ["Key1"] }]' | $cli create collection db1 -
	#reading schemas from command line array
	$cli create collection db1 '[{ "title" : "coll5", "properties": { "Key1": { "type": "string" }, "Field1": { "type": "integer" } }, "primary_key": ["Key1"] }, { "title" : "coll6", "properties": { "Key1": { "type": "string" }, "Field1": { "type": "integer" } }, "primary_key": ["Key1"] }]' '{ "title" : "coll7", "properties": { "Key1": { "type": "string" }, "Field1": { "type": "integer" } }, "primary_key": ["Key1"] }'
	# allow to skip - in non interactive input
	$cli create collection db1 <<< '[{ "title" : "coll8", "properties": { "Key1": { "type": "string" }, "Field1": { "type": "integer" } }, "primary_key": ["Key1"] }, { "title" : "coll9", "properties": { "Key1": { "type": "string" }, "Field1": { "type": "integer" } }, "primary_key": ["Key1"] }]'

	$cli list databases
	$cli list collections db1

	#insert from command line parameters
	$cli insert db1 coll1 '{"Key1": "vK1", "Field1": 1}' \
		'{"Key1": "vK2", "Field1": 10}'

	#duplicate key
	$cli insert db1 coll1 '{"Key1": "vK1", "Field1": 1}' && exit 1

	#insert from array
	$cli insert db1 coll1 '[{"Key1": "vK7", "Field1": 1},
		{"Key1": "vK8", "Field1": 10}]'

	$cli replace db1 coll1 '{"Key1": "vK1", "Field1": 1111}' \
		'{"Key1": "vK211", "Field1": 10111}'

	$cli replace db1 coll1 '[{"Key1": "vK7", "Field1": 22222}]' \
		'[{"Key1": "vK2", "Field1": 10}]'

	#insert from standard input stream
	cat <<EOF | $cli insert "db1" "coll1"
{"Key1": "vK10", "Field1": 10}
{"Key1": "vK20", "Field1": 20}
{"Key1": "vK30", "Field1": 30}
EOF

	cat <<EOF | $cli replace "db1" "coll1"
{"Key1": "vK100", "Field1": 100}
{"Key1": "vK200", "Field1": 200}
{"Key1": "vK300", "Field1": 300}
EOF

	#insert from standard input array
	#NOTE: space and tabs are intentional. to test trim functionality
	cat <<EOF | $cli insert "db1" "coll1" -
  	 [
{"Key1": "vK1011", "Field1": 1044},
{"Key1": "vK2011", "Field1": 2055},
{"Key1": "vK3011", "Field1": 3066}
]
EOF

	cat <<EOF | $cli replace db1 coll1 -
[{"Key1": "vK101", "Field1": 104},
{"Key1": "vK202", "Field1": 205},
{"Key1": "vK303", "Field1": 306}]
EOF

	#copy collection content
	$cli read db1 coll1 | $cli insert db1 coll2 -

	exp_out='{"Key1": "vK1", "Field1": 1111}
{"Key1": "vK10", "Field1": 10}
{"Key1": "vK100", "Field1": 100}
{"Key1": "vK101", "Field1": 104}
{"Key1": "vK1011", "Field1": 1044}
{"Key1": "vK2", "Field1": 10}
{"Key1": "vK20", "Field1": 20}
{"Key1": "vK200", "Field1": 200}
{"Key1": "vK2011", "Field1": 2055}
{"Key1": "vK202", "Field1": 205}
{"Key1": "vK211", "Field1": 10111}
{"Key1": "vK30", "Field1": 30}
{"Key1": "vK300", "Field1": 300}
{"Key1": "vK3011", "Field1": 3066}
{"Key1": "vK303", "Field1": 306}
{"Key1": "vK7", "Field1": 22222}
{"Key1": "vK8", "Field1": 10}'

	out=$($cli read db1 coll1 '{}')
	diff -w -u <(echo "$exp_out") <(echo "$out")

	out=$($cli read db1 coll2 '{}')
	diff -w -u <(echo "$exp_out") <(echo "$out")

	# shellcheck disable=SC2016
	$cli update db1 coll1 '{"Key1": "vK1"}' '{"$set" : {"Field1": 1000}}'

	out=$($cli read db1 coll1 '{"Key1": "vK1"}' '{"Field1":true}')
	diff -w -u <(echo '{"Field1":1000}') <(echo "$out")

	$cli delete db1 coll1 '{"Key1": "vK1"}'

	out=$($cli read "db1" "coll1" '{"Key1": "vK1"}')
	[[ "$out" == '' ]] || exit 1

	$cli insert db1 coll3 '{"Key1": "vK1", "Field1": 1}' \
		'{"Key1": "vK2", "Field1": 10}'

	cat <<'EOF' | $cli transact "db1" -
[
{"insert" : { "collection" : "coll3", "documents": [{"Key1": "vK20000", "Field1": 20022}]}},
{"replace" : { "collection" : "coll3", "documents": [{"Key1": "vK30000", "Field1": 30033}]}},
{"update" : { "collection" : "coll3", "filter" : { "Key1": "vK2" }, "fields" : { "$set" : { "Field1" : 10000111 }}}},
{"delete" : { "collection" : "coll3", "filter" : { "Key1": "vK1" }}},
{"read"   : { "collection" : "coll3", "filter" : {}, "fields" : {}}}
]
EOF

	out=$($cli read db1 coll3)
exp_out='{"Key1": "vK2", "Field1": 10000111}
{"Key1": "vK20000", "Field1": 20022}
{"Key1": "vK30000", "Field1": 30033}'
	diff -w -u <(echo "$exp_out") <(echo "$out")

	$cli insert db1 coll4 '{"Key1": "vK1", "Field1": 1}' \
		'{"Key1": "vK2", "Field1": 10}'

	cat <<'EOF' | $cli transact "db1" -
{"operation": "insert", "collection" : "coll4", "documents": [{"Key1": "vK200", "Field1": 20}]}
{"operation": "replace", "collection" : "coll4", "documents": [{"Key1": "vK300", "Field1": 30}]}
{"operation": "update", "collection" : "coll4", "filter" : { "Key1": "vK1" }, "fields" : { "$set" : { "Field1" : 10000 }}}
{"operation": "delete", "collection" : "coll4", "filter" : { "Key1": "vK2" }}
{"operation": "read", "collection" : "coll4"}
EOF

	out=$($cli read db1 coll4)
	exp_out='{"Key1": "vK1", "Field1": 10000}
{"Key1": "vK200", "Field1": 20}
{"Key1": "vK300", "Field1": 30}'
	diff -w -u <(echo "$exp_out") <(echo "$out")

	db_negative_tests
	db_errors_tests
	db_generate_schema_test

	$cli drop collection db1 coll1 coll2 coll3 coll4 coll5 coll6 coll7
	$cli drop database db1
}

db_negative_tests() {
	#broken json
	echo '{"Key1": "vK10", "Fiel' | $cli insert db1 coll1 - && exit 1
	$cli insert db1 coll1 '{"Key1": "vK10", "Fiel' && exit 1
	#broken array
	echo '[{"Key1": "vK10", "Field1": 10}' | $cli insert db1 coll1 - && exit 1
	$cli insert db1 coll1 '[{"Key1": "vK10", "Field1": 10}' && exit 1

	#not enough arguments
	$cli read "db1" && exit 1
	$cli update "db1" "coll1" '{"Key1": "vK1"}' && exit 1
	$cli replace db1 && exit 1
	$cli insert db1 && exit 1
	$cli delete db1 coll1 && exit 1
	$cli create collection && exit 1
	$cli create database && exit 1
	$cli drop collection db1 && exit 1
	$cli drop database && exit 1
	$cli list collections && exit 1

	true
}

error() {
	exp_out=$1
	shift
	out=$("$@" 2>&1 || true)
	diff -u <(echo "$exp_out") <(echo "$out")
}

# BUG: Unify HTTP and GRPC responses
# shellcheck disable=SC2086
db_errors_tests() {
	error "database doesn't exist 'db2'" $cli drop database db2 ||
	error "404 Not Found" $cli drop database db2

	error "database doesn't exist 'db2'" $cli drop collection db2 coll1 ||
	error "404 Not Found" $cli drop collection db2 coll1

	error "database doesn't exist 'db2'" $cli create collection db2 \
		'{ "title" : "coll1", "properties": { "Key1": { "type": "string" }, "Field1": { "type": "integer" }, "Field2": { "type": "integer" } }, "primary_key": ["Key1"] }' ||
	error "404 Not Found" $cli create collection db2 \
		'{ "title" : "coll1", "properties": { "Key1": { "type": "string" }, "Field1": { "type": "integer" }, "Field2": { "type": "integer" } }, "primary_key": ["Key1"] }'

	error "database doesn't exist 'db2'" $cli list collections db2 ||
	error "404 Not Found" $cli list collections db2

	error "database doesn't exist 'db2'" $cli insert db2 coll1 '{}' ||
	error "404 Not Found" $cli insert db2 coll1 '{}'

	error "database doesn't exist 'db2'" $cli read db2 coll1 '{}' ||
	error "404 Not Found" $cli read db2 coll1 '{}'

	error "database doesn't exist 'db2'" $cli update db2 coll1 '{}' '{}' ||
	error "404 Not Found" $cli update db2 coll1 '{}' '{}'

	error "database doesn't exist 'db2'" $cli delete db2 coll1 '{}' ||
	error "404 Not Found" $cli delete db2 coll1 '{}'

	$cli create database db2
	error "collection doesn't exist 'coll1'" $cli insert db2 coll1 '{}' ||
	error "404 Not Found" $cli insert db2 coll1 '{}'

	error "collection doesn't exist 'coll1'" $cli read db2 coll1 '{}' ||
	error "404 Not Found" $cli read db2 coll1 '{}'

	error "collection doesn't exist 'coll1'" $cli update db2 coll1 '{}' '{}' ||
	error "404 Not Found" $cli update db2 coll1 '{}' '{}'

	error "collection doesn't exist 'coll1'" $cli delete db2 coll1 '{}' ||
	error "404 Not Found" $cli delete db2 coll1 '{}'

	error "schema name is missing" $cli create collection db1 \
		'{ "properties": { "Key1": { "type": "string" }, "Field1": { "type": "integer" }, "Field2": { "type": "integer" } }, "primary_key": ["Key1"] }'

	$cli drop database db2
}

db_generate_schema_test() {
  error "sampledb created with the collections" $cli generate sample-schema --create
  $cli drop database sampledb
}

unset TIGRIS_PROTOCOL
unset TIGRIS_URL
db_tests

export TIGRIS_PROTOCOL=grpc
$cli config show | grep "protocol: grpc"
db_tests

export TIGRIS_PROTOCOL=http
$cli config show | grep "protocol: http"
db_tests

export TIGRIS_URL=localhost:8081
export TIGRIS_PROTOCOL=grpc
$cli config show | grep "protocol: grpc"
$cli config show | grep "url: localhost:8081"
db_tests

export TIGRIS_URL=localhost:8081
export TIGRIS_PROTOCOL=http
$cli config show | grep "protocol: http"
$cli config show | grep "url: localhost:8081"
db_tests

export TIGRIS_PROTOCOL=grpc
export TIGRIS_URL=http://localhost:8081
$cli config show | grep "protocol: grpc"
$cli config show | grep "url: http://localhost:8081"
db_tests

export TIGRIS_PROTOCOL=http
export TIGRIS_URL=grpc://localhost:8081
$cli config show | grep "protocol: http"
$cli config show | grep "url: grpc://localhost:8081"
db_tests

$cli local down

