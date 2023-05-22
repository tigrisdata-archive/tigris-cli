#!/bin/bash

set -ex
PS4='${LINENO}: '

if [ -z "$cli" ]; then
	cli="$(pwd)/tigris"
fi

if [ -z "$TIGRIS_TEST_PORT" ]; then
	TIGRIS_TEST_PORT=8090
fi

cli="$SUDO $cli"

DATA_DIR=/tmp/tigris-cli-local-test

test_persistence_low() {
  rm -rf $DATA_DIR
  rm "$HOME/.tigris/tigris-cli.yaml"

  #shellcheck disable=SC2086
  $cli local up --data-dir=$DATA_DIR $1 $2
  [ -f $DATA_DIR/initialized ] || exit 1

  $cli create project persistence_test
  $cli list projects|grep persistence_test

  $cli local down

  #shellcheck disable=SC2086
  $cli local up $2
  $cli list projects | grep persistence_test

  [ -f $DATA_DIR/initialized ] || exit 1

  $cli local down

  #shellcheck disable=SC2086
  $cli local up --skip-auth $TIGRIS_TEST_PORT
  TIGRIS_TOKEN="" TIGRIS_URL=localhost:$TIGRIS_TEST_PORT $cli list projects || grep persistence_test

  $cli local down
}

test_persistence() {
  HOME=/tmp/ test_persistence_low
  HOME=/tmp/ TIGRIS_URL=localhost:$TIGRIS_TEST_PORT test_persistence_low --token-admin-auth "$TIGRIS_TEST_PORT"
}
