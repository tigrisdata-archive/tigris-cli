# Tigris Command Line Interface

[![Build Status](https://github.com/tigrisdata/tigrisdb/workflows/go-lint/badge.svg)]()
[![Build Status](https://github.com/tigrisdata/tigrisdb/workflows/go-test/badge.svg)]()

[![NPM](https://img.shields.io/npm/v/@tigrisdata/tigris-cli)]()
[![tigris](https://snapcraft.io/tigris/badge.svg)](https://snapcraft.io/tigris)

Create databases and collections, read and write data, perform transactions,
stream events and setup Tigris locally, all from the command line.

# Documentation

- [Quickstart](https://docs.tigrisdata.com/quickstart/with-cli)
- [Working Locally](https://docs.tigrisdata.com/cli/working-locally)
- [Command Reference](https://docs.tigrisdata.com/cli)

# Installation

The tigris CLI tool can be installed as follows:

## macOS

```shell
curl -sSL https://tigris.dev/cli-macos | sudo tar -xz -C /usr/local/bin
```

or

```shell
brew install tigrisdata/tigris/tigris-cli
```

## Linux

```shell
curl -sSL https://tigris.dev/cli-linux | sudo tar -xz -C /usr/local/bin
```

or

```shell
sudo snap install tigris
```

## Cross-platform using NPM

```shell
sudo npm i -g @tigrisdata/tigris-cli
```

# Usage

```shell
tigris is a command line interface of Tigris data platform

Usage:
  tigris [command]

Available Commands:
  alter          Alters collection
  backup         Dumps documents and schemas to JSON files
  branch         Working with Tigris branches
  completion     Generates completion script for shell
  config         Configuration commands
  create         Creates project, collection, namespace or app_key
  db             Database related commands
  delete         Deletes document(s)
  delete-project Deletes project
  describe       Describes database or collection
  dev            Starts and stops local development Tigris server
  docs           Generates CLI documentation in Markdown format
  drop           Drops collection or application
  generate       Generating helper assets such as sample schema
  help           Help about any command
  import         Import documents into collection
  insert         Inserts document(s)
  list           Lists projects, collections or namespaces
  login          Authenticate on the Tigris instance
  logout         Logout from Tigris instance
  ping           Checks connection to Tigris
  quota          Quota related commands
  read           Reads and outputs documents
  replace        Inserts or replaces document(s)
  restore        restores documents and schemas from JSON files
  scaffold       Scaffold new application for project
  search         Search related commands
  server         Tigris server related commands
  transact       Executes a set of operations in a transaction
  update         Updates document(s)
  version        Shows tigris cli version

Flags:
  -h, --help    help for tigris
  -q, --quiet   Suppress informational messages

Use "tigris [command] --help" for more information about a command.
```

# Examples

```shell
# Start up Tigris locally on port 8081
tigris local up

# Create sample schema
tigris generate sample-schema --create

# Create database and collection
tigris create project --project=test
tigris create collection --project=test \
'{
    "title": "users",
    "description": "Collection of documents with details of users",
    "properties": {
      "id": {
        "description": "A unique identifier for the user",
        "type": "integer"
      },
      "name": {
        "description": "Name of the user",
        "type": "string",
        "maxLength": 100
      }
    },
    "primary_key": [
      "id"
    ]
}'

# Insert some data
tigris insert_or_replace --project=test users \
'[
    {"id": 1, "name": "Jania McGrory", "balance": 6045.7},
    {"id": 2, "name": "Bunny Instone", "balance": 2948.87}
]'

# Read data
tigris read --project=test users '{"id": 1}'

# Perform a transaction
tigris transact --project=test \
'[
  {
    "insert": {
      "collection": "orders",
      "documents": [{
          "id": 1, "user_id": 1, "order_total": 53.89, "products": [{"id": 1, "quantity": 1}]
        }]
    }
  },
  {
    "update": {
      "collection": "users", "fields": {"$set": {"balance": 5991.81}}, "filter": {"id": 1}
    }
  },
  {
    "update": {
      "collection": "products", "fields": {"$set": {"quantity": 6357}}, "filter": {"id": 1}
    }
  }
]'

# Shutdown the local Tigris
tigris local down
```

# Development

Run `make setup` it installs required test dependencies:
- [golangci-lint](https://github.com/golangci/golangci-lint)
- [shellcheck](https://github.com/koalaman/shellcheck)

Use `make test-fast` and `make lint` for testing the changes.

# License

This software is licensed under the [Apache 2.0](LICENSE).
