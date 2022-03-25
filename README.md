# TigrisDB command line interface utility

[![Build Status](https://github.com/tigrisdata/tigrisdb/workflows/go-lint/badge.svg)]()

# Install

```sh
go install github.com/tigrisdata/tigrisdb-cli@latest
```

# Example

```sh
tigrisdb-cli drop collection db1 coll1
tigrisdb-cli drop database db1

tigrisdb-cli create database db1
tigrisdb-cli create collection db1 coll1 \
	'{ "properties": { "Key1": { "type": "string" }, "Field1": { "type": "int" } }, "primary_key": ["Key1"] }'

tigrisdb-cli list databases
tigrisdb-cli list collections db1

tigrisdb-cli insert "db1" "coll1" '{"Key1": "vK1", "Field1": 1}' \
	'{"Key1": "vK2", "Field1": 10}'

tigrisdb-cli read "db1" "coll1" \
	'{"$or" : [ {"Key1": "vK1"}, {"Key1": "vK2"} ]}'
#Output:
#{"Key1": "vK1", "Field1": 1}
#{"Key1": "vK2", "Field1": 10}

tigrisdb-cli update "db1" "coll1" \
	'{"Key1": "vK1"}' '{"$set" : {"Field1": 1000}}'

tigrisdb-cli read "db1" "coll1" '{}'
#Output:
#{"Key1": "vK1", "Field1": 1000}
#{"Key1": "vK2", "Field1": 10}

tigrisdb-cli delete "db1" "coll1" '{"Key1": "vK1"}'

tigrisdb-cli read "db1" "coll1" {}'
#Output:
#{"Key1": "vK2", "Field1": 10}

tigrisdb-cli drop collection db1 coll1
tigrisdb-cli drop database db1
```

# License
This software is licensed under the [Apache 2.0](LICENSE).
