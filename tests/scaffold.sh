#!/bin/bash

set -ex

if [ -z "$cli" ]; then
	cli="./tigris"
fi

APP_PORT=8080

# first parameter is path
# second parameter is document to write
request() {
  curl -X POST "localhost:$APP_PORT/$1" -H 'Content-Type: application/json' -d "$2"
  echo
}

test_crud_routes() {
  sleep 3 # give some time server to start

  request users '{"id":1, "name":"John","balance":100}' #Id=1
  request users '{"id":2, "name":"Jane","balance":200}' #Id=2

  request products '{"id":1, "name":"Avocado","price":10,"quantity":5}' #Id=1
  request products '{"id":2, "name":"Gold","price":3000,"quantity":1}' #Id=2

  #low balance
  request orders '{"id":1, "user_id":1,"productItems":[{"Id":2,"Quantity":1}]}' || true
  # low stock
  request orders '{"id":2, "user_id":1,"productItems":[{"Id":1,"Quantity":10}]}' || true

  request orders '{"id":3, "user_id":1,"productItems":[{"Id":1,"Quantity":5}]}' #Id=1

  curl localhost:$APP_PORT/users/1
  echo
  curl localhost:$APP_PORT/products/1
  echo
  curl localhost:$APP_PORT/orders/1
  echo

  # search
  request users/search '{"q":"john"}'
  request products/search '{"q":"avocado","searchFields": ["name"]}'
}

db=eshop

clean() {
  $cli delete-project -f --project $db || true
  rm -rf /tmp/cli-test/
}

scaffold() {
  $cli create project --project $db \
    --template=ecommerce \
    --framework="$2" \
    --language "$1" \
    --package-name="github.com/tigrisdata/$db" \
    --output-directory=/tmp/cli-test
}

test_base_go() {
  clean

  scaffold go base

  tree /tmp/cli-test/$db
  cd /tmp/cli-test/$db

  task run

  cd -

  clean
}

test_base_typescript() {
  clean

  scaffold typescript base

  tree /tmp/cli-test/$db
  cd /tmp/cli-test/$db

  npm ci
  npm start

  cd -

  clean
}

test_gin_go() {
  clean

  scaffold go gin

  tree /tmp/cli-test/$db
  cd /tmp/cli-test/$db

  task "$1"

  cd -

  test_crud_routes

  clean
}

test_express_typescript() {
  clean

  scaffold typescript express

  tree /tmp/cli-test/$db
  cd /tmp/cli-test/$db

  npm ci
  npm run "$1"

  cd -

  test_crud_routes

  clean
}

test_scaffold() {
  #FIXME: Reenabled after next CLI release
  #The tests use released version of CLI
  #and unreleased commit adds Colima support on MacOS
  #and fixes server health check on startup
  return

  test_gin_go run:docker
  test_base_go

  export PORT=8080

  test_express_typescript start:docker
  test_base_typescript
}

