// Copyright 2022 Tigris Data, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/util"
	"github.com/tigrisdata/tigris-client-go/driver"
)

const (
	Read                     = "read"
	Insert                   = "insert"
	Update                   = "update"
	Delete                   = "delete"
	Replace                  = "replace" // alias for insert_or_replace
	InsertOrReplace          = "insert_or_replace"
	CreateOrUpdateCollection = "create_or_update_collection"
	DropCollection           = "drop_collection"
	ListCollections          = "list_collections"
)

type Op struct {
	Operation  string
	Collection string
	Documents  []json.RawMessage
	Filter     json.RawMessage
	Fields     json.RawMessage
	Schema     json.RawMessage
}

type TxOp struct {
	Op                       `json:",inline"`
	Insert                   *Op
	Replace                  *Op
	InsertOrReplace          *Op `json:"insert_or_replace"`
	Delete                   *Op
	Update                   *Op
	Read                     *Op
	CreateOrUpdateCollection *Op `json:"create_or_update_collection"`
	DropCollection           *Op
	ListCollections          *Op
}

func execTxOp(ctx context.Context, tx driver.Tx, tp string, op *Op) {
	if op == nil {
		return
	}
	var err error
	switch tp {
	case Insert:
		ptr := unsafe.Pointer(&op.Documents)
		_, err = tx.Insert(ctx, op.Collection, *(*[]driver.Document)(ptr))
	case Update:
		_, err = tx.Update(ctx, op.Collection, driver.Filter(op.Filter), driver.Update(op.Fields))
	case Delete:
		_, err = tx.Delete(ctx, op.Collection, driver.Filter(op.Filter))
	case Replace, InsertOrReplace:
		ptr := unsafe.Pointer(&op.Documents)
		_, err = tx.Replace(ctx, op.Collection, *(*[]driver.Document)(ptr), &driver.ReplaceOptions{})
	case CreateOrUpdateCollection:
		err = tx.CreateOrUpdateCollection(ctx, op.Collection, driver.Schema(op.Schema))
	case DropCollection:
		err = tx.DropCollection(ctx, op.Collection)
	case ListCollections:
		colls, err := tx.ListCollections(ctx)
		if err != nil {
			util.Error(err, "transact operation failed")
		}
		for _, c := range colls {
			util.Stdout("%s\n", c)
		}
		return
	case Read:
		filter := json.RawMessage(`{}`)
		fields := json.RawMessage(`{}`)
		if len(op.Filter) > 0 {
			filter = op.Filter
		}
		if len(op.Fields) > 0 {
			fields = op.Fields
		}
		it, err := tx.Read(ctx, op.Collection, driver.Filter(filter), driver.Projection(fields))
		if err != nil {
			util.Error(err, "transact operation failed")
		}
		var d driver.Document
		for it.Next(&d) {
			util.Stdout("%s\n", string(d))
		}
		return
	default:
		util.Error(fmt.Errorf("unknown operation type: %s", op.Operation), "")
	}
	if err != nil {
		util.Error(err, "transact operation failed")
	}
}

var transactCmd = &cobra.Command{
	Use:     "transact {db} {operation}...|-",
	Aliases: []string{"tx"},
	Short:   "Executes a set of operations in a transaction",
	Long: `Executes a set of operations in a transaction.
All the read, write and schema operations are supported.`,
	Example: fmt.Sprintf(`
  # Perform a transaction that inserts and updates in three collections
  %[1]s tigris transact testdb \
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
`, rootCmd.Root().Name()),
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		db := args[0]
		client.Transact(cmd.Context(), db, func(ctx context.Context, tx driver.Tx) {
			iterateInput(ctx, cmd, 1, args, func(ctx context.Context, args []string, ops []json.RawMessage) {
				for _, iop := range ops {
					var op TxOp
					if err := json.Unmarshal(iop, &op); err != nil {
						util.Error(err, "begin transaction failed")
					}

					if op.Operation != "" {
						execTxOp(ctx, tx, op.Operation, &op.Op)
					}

					execTxOp(ctx, tx, InsertOrReplace, op.InsertOrReplace)
					execTxOp(ctx, tx, Replace, op.Replace)
					execTxOp(ctx, tx, Insert, op.Insert)
					execTxOp(ctx, tx, Read, op.Read)
					execTxOp(ctx, tx, Update, op.Update)
					execTxOp(ctx, tx, Delete, op.Delete)
					execTxOp(ctx, tx, CreateOrUpdateCollection, op.CreateOrUpdateCollection)
					execTxOp(ctx, tx, DropCollection, op.DropCollection)
					execTxOp(ctx, tx, ListCollections, op.ListCollections)
				}
			})
		})
	},
}

func init() {
	dbCmd.AddCommand(transactCmd)
}
