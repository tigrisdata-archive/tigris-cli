# Tigris Command Line Interface

[![Build Status](https://github.com/tigrisdata/tigrisdb/workflows/go-lint/badge.svg)]()
[![Build Status](https://github.com/tigrisdata/tigrisdb/workflows/go-test/badge.svg)]()

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

## Linux

```shell
curl -sSL https://tigris.dev/cli-linux | sudo tar -xz -C /usr/local/bin
```

# Usage

```shell
$ tigris
tigris is a command line interface of Tigris data platform

Usage:
  tigris [command]

Available Commands:
  alter       Alters collection
  backup      Dumps documents and schemas to JSON files on a quiesced database
  completion  Generates completion script for shell
  config      Configuration commands
  create      Creates database/project, collection, namespace or application
  delete      delete project, collection or application
  describe    Describes database or collection
  dev         Starts and stops local development Tigris server
  docs        Generates CLI documentation in Markdown format
  generate    Generating helper assets such as sample schema
  help        Help about any command
  import      Import documents into collection
  insert      Inserts document(s)
  list        Lists databases/projects, collections or namespaces
  login       Authenticate on the Tigris instance
  logout      Logout from Tigris instance
  ping        Checks connection to Tigris
  publish     Publish message(s)
  quota       Quota related commands
  read        Reads and outputs documents
  replace     Inserts or replaces document(s)
  scaffold    Scaffold a project for specified language
  search      Searches a collection for documents matching the query
  server      Tigris server related commands
  subscribe   Subscribes to published messages
  transact    Executes a set of operations in a transaction
  update      Updates document(s)
  version     Shows tigris cli version

Flags:
  -h, --help   help for tigris

Use "tigris [command] --help" for more information about a command.
```

# Examples

```shell
# Start up Tigris locally on port 8081
tigris local up

# Create sample schema
tigris generate sample-schema --create

# Create database and collection
tigris create project my-first-project
tigris create collection my-first-project \
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
tigris insert_or_replace my-first-project users \
'[
    {"id": 1, "name": "Jania McGrory", "balance": 6045.7},
    {"id": 2, "name": "Bunny Instone", "balance": 2948.87}
]'

# Read data
tigris read my-first-project users '{"id": 1}'

# Perform a transaction
tigris transact my-first-project \
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

# License
This software is licensed under the [Apache 2.0](LICENSE).
