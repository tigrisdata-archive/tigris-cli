# TigrisDB command line interface utility

[![Build Status](https://github.com/tigrisdata/tigrisdb/workflows/go-lint/badge.svg)]()
[![Build Status](https://github.com/tigrisdata/tigrisdb/workflows/go-test/badge.svg)]()

# Install

## macOS

```sh
curl -sSL https://tigris.dev/cli-macos | sudo tar -xz -C /usr/local/bin
```

## Linux

```sh
curl -sSL https://tigris.dev/cli-linux | sudo tar -xz -C /usr/local/bin
```

# Example

```sh
tigris local up # brings local TigrisDB up on localhost:8081

tigris create database db1
tigris create collection db1 '{
  "name" : "coll1",
  "properties": {
      "Key1":   { "type": "string" },
      "Field1": { "type": "integer" }
  },
  "primary_key": ["Key1"]
}'

tigris list databases
tigris list collections db1

# Insert two documents into coll1
tigris insert db1 coll1 '{"Key1": "vK1", "Field1": 1}' \
	'{"Key1": "vK2", "Field1": 10}'

# Read documents with keys equal to vK1 and vK2
tigris read db1 coll1 \
	'{"$or" : [ {"Key1": "vK1"}, {"Key1": "vK2"} ]}'
#Output:
#{"Key1": "vK1", "Field1": 1}
#{"Key1": "vK2", "Field1": 10}

# Update Field1 of the document with the key vK1
tigris update db1 coll1 \
	'{"Key1": "vK1"}' '{"$set" : {"Field1": 1000}}'

# Read all documents from the coll1
tigris read db1 coll1 '{}'
#Output:
#{"Key1": "vK1", "Field1": 1000}
#{"Key1": "vK2", "Field1": 10}

tigris delete "db1" "coll1" '{"Key1": "vK1"}'

tigris read "db1" "coll1" '{}'
#Output:
#{"Key1": "vK2", "Field1": 10}

tigris drop collection db1 coll1
tigris drop database db1

tigris local down
```

# License
This software is licensed under the [Apache 2.0](LICENSE).
