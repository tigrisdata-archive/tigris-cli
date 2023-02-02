// Copyright 2022-2023 Tigris Data, Inc.
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
	Operation  string            `json:"operation"`
	Collection string            `json:"collection"`
	Documents  []json.RawMessage `json:"documents"`
	Filter     json.RawMessage   `json:"filter"`
	Fields     json.RawMessage   `json:"fields"`
	Schema     json.RawMessage   `json:"schema"`
}

type TxOp struct {
	Op                       `json:",inline"`
	Insert                   *Op `json:"insert"`
	Replace                  *Op `json:"replace"`
	InsertOrReplace          *Op `json:"insert_or_replace"`
	Delete                   *Op `json:"delete"`
	Update                   *Op `json:"update"`
	Read                     *Op `json:"read"`
	CreateOrUpdateCollection *Op `json:"create_or_update_collection"`
	DropCollection           *Op `json:"drop_collection"`
	ListCollections          *Op `json:"list_collections"`
}

var ErrUnknownOperationType = fmt.Errorf("unknown operation type")

func execTxOpRead(ctx context.Context, tx driver.Tx, op *Op) error {
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
		return util.Error(err, "transact operation read")
	}

	var d driver.Document
	for it.Next(&d) {
		util.Stdoutf("%s\n", string(d))
	}

	return util.Error(it.Err(), "tx read op")
}

func execTxOpListColls(ctx context.Context, tx driver.Tx) error {
	colls, err := tx.ListCollections(ctx)
	if err != nil {
		return util.Error(err, "transact operation failed")
	}

	for _, c := range colls {
		util.Stdoutf("%s\n", c)
	}

	return nil
}

func execTxOpLow(ctx context.Context, tx driver.Tx, tp string, op *Op) error {
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
		return execTxOpListColls(ctx, tx)
	case Read:
		return execTxOpRead(ctx, tx, op)
	default:
		util.Fatal(fmt.Errorf("%w: %s", ErrUnknownOperationType, op.Operation), "")
	}

	return err
}

func execTxOp(ctx context.Context, tx driver.Tx, tp string, op *Op) error {
	if op == nil {
		return nil
	}

	err := execTxOpLow(ctx, tx, tp, op)

	return util.Error(err, "transact operation failed")
}

var transactCmd = &cobra.Command{
	Use:     "transact {operation}...|-",
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
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			return client.Transact(ctx, getProjectName(), func(ctx context.Context, tx driver.Tx) error {
				return iterateInput(ctx, cmd, 1, args, func(ctx context.Context, args []string, ops []json.RawMessage) error {
					for _, iop := range ops {
						var op TxOp

						err := json.Unmarshal(iop, &op)
						util.Fatal(err, "begin transaction")

						if op.Operation != "" {
							if err = execTxOp(ctx, tx, op.Operation, &op.Op); err != nil {
								return util.Error(err, "execute tx "+op.Operation)
							}
						}

						if err = execTxOp(ctx, tx, InsertOrReplace, op.InsertOrReplace); err != nil {
							return util.Error(err, "execute tx InsertOrReplace")
						}

						if err = execTxOp(ctx, tx, Replace, op.Replace); err != nil {
							return util.Error(err, "execute tx Replace")
						}

						if err = execTxOp(ctx, tx, Insert, op.Insert); err != nil {
							return util.Error(err, "execute tx Insert")
						}

						if err = execTxOp(ctx, tx, Read, op.Read); err != nil {
							return util.Error(err, "execute tx Read")
						}

						if err = execTxOp(ctx, tx, Update, op.Update); err != nil {
							return util.Error(err, "execute tx Update")
						}

						if err = execTxOp(ctx, tx, Delete, op.Delete); err != nil {
							return util.Error(err, "execute tx Delete")
						}

						if err = execTxOp(ctx, tx, CreateOrUpdateCollection, op.CreateOrUpdateCollection); err != nil {
							return util.Error(err, "execute tx CreateOrUpdateCollection")
						}

						if err = execTxOp(ctx, tx, DropCollection, op.DropCollection); err != nil {
							return util.Error(err, "execute tx DropCollection")
						}

						if err = execTxOp(ctx, tx, ListCollections, op.ListCollections); err != nil {
							return util.Error(err, "execute tx ListCollections")
						}
					}

					return nil
				})
			})
		})
	},
}

func init() {
	addProjectFlag(transactCmd)
	dbCmd.AddCommand(transactCmd)
}
