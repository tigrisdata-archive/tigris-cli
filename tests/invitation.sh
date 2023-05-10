#!/bin/bash

set -ex

# There is no way to automate invitations testing,
# so as it requires TLS setup on the local instance.

TIGRIS_TOKEN=$(cat ../../docker/test-token-rsa.jwt)
TIGRIS_URL=localhost:8082

export TIGRIS_TOKEN
export TIGRIS_URL

cli=./tigris

$cli config show

$cli invitation create --email test1@tigrisdata.com --role editor_a
$cli invitation create --email test2@tigrisdata.com --role editor_b
$cli invitation create --email test3@tigrisdata.com --role editor_c
$cli invitation create --email test4@tigrisdata.com --role editor_d

$cli invitations list

$cli invitation delete --email test3@tigrisdata.com --status PENDING

$cli invitations list

#	verify - valid code
$cli invitation verify --email test4@tigrisdata.com --code code1

# verify - invalid code
$cli invitation verify --email test4@tigrisdata.com --code code1
