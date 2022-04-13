# Examples

# Backup and restore database schema

```sh
tigris db list collections db1 --dump-schema >db1_schemas.json #FIXME: implement
# Duplicate schema from db1 to db2
tigris db create collection db2 <db1_schemas.json
# Restore from URL
curl https://example.com/schemas.json | tigris db create collection db2 - 
```

# Backup and restore all documents in the collection

```sh
tigris db read db1 coll1 >db1_coll1_documents.json
tigris db insert db1 coll1 <db1_coll1_documents.json
```  