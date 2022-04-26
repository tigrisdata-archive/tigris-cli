# Tigris command line interface

[![Build Status](https://github.com/tigrisdata/tigrisdb/workflows/go-lint/badge.svg)]()
[![Build Status](https://github.com/tigrisdata/tigrisdb/workflows/go-test/badge.svg)]()

# Documentation
* [Quickstart](https://docs.tigrisdata.com/quickstart/with-cli)

# Installation

## macOS

```shell
curl -sSL https://tigris.dev/cli-macos | sudo tar -xz -C /usr/local/bin
```

## Linux

```shell
curl -sSL https://tigris.dev/cli-linux | sudo tar -xz -C /usr/local/bin
```

# CLI Usage

```shell
$ tigris
tigris is a command line interface of TigrisDB database

Usage:
  tigris [command]

Available Commands:
  alter       alter collection
  completion  Generate the autocompletion script for the specified shell
  create      creates database or collection
  delete      delete documents
  describe    describe database or collection
  drop        drop database or collection
  generate    Generating helper assets such as sample schema
  help        Help about any command
  insert      insert document
  list        list databases of the project or list database collections
  local       Starting and stopping local TigrisDB server
  ping        Check connection to the TigrisDB
  read        read documents
  replace     replace document
  transact    run a set of operations in a transaction
  update      update documents
  version     show tigrisdb-cli version

Flags:
  -h, --help   help for tigris

Use "tigris [command] --help" for more information about a command.
```

For more information on usage visit the
[Tigris Docs page](https://docs.tigrisdata.com/quickstart/with-cli).

# Examples

```shell
# Start Tigris up locally on localhost:8081
tigris local up

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

# Insert two documents into the collection coll1
tigris insert db1 coll1 '{"Key1": "vK1", "Field1": 1}' \
	'{"Key1": "vK2", "Field1": 10}'

# Read documents with keys equal to vK1 and vK2
# Output:
# {"Key1": "vK1", "Field1": 1}
# {"Key1": "vK2", "Field1": 10}
tigris read db1 coll1 \
	'{"$or" : [ {"Key1": "vK1"}, {"Key1": "vK2"} ]}'

# Update Field1 of the document with the key vK1
tigris update db1 coll1 \
	'{"Key1": "vK1"}' '{"$set" : {"Field1": 1000}}'

# Read all documents from the collection coll1
# Output:
# {"Key1": "vK1", "Field1": 1000}
# {"Key1": "vK2", "Field1": 10}
tigris read db1 coll1 '{}'

# Delete a document from the collection coll1 with the key vK1
tigris delete "db1" "coll1" '{"Key1": "vK1"}'

# Drop the collection coll1
tigris drop collection db1 coll1

# Drop the database db1
tigris drop database db1

# Stop Tigris running locally
tigris local down
```

# Community & Support

* [Slack Community](https://join.slack.com/t/tigrisdatacommunity/shared_invite/zt-16fn5ogio-OjxJlgttJIV0ZDywcBItJQ)
* [GitHub Issues](https://github.com/tigrisdata/tigrisdb-cli/issues)
* [GitHub Discussions](https://github.com/tigrisdata/tigrisdb/discussions)

# License
This software is licensed under the [Apache 2.0](LICENSE).
