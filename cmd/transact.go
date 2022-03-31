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
	"unsafe"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigrisdb-cli/client"
	"github.com/tigrisdata/tigrisdb-cli/util"
	"github.com/tigrisdata/tigrisdb-client-go/driver"
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

func execTxOp(ctx context.Context, db string, tp string, op *Op) {
	if op == nil {
		return
	}
	var err error
	switch tp {
	case Insert:
		ptr := unsafe.Pointer(&op.Documents)
		_, err = client.Get().Insert(ctx, db, op.Collection, *(*[]driver.Document)(ptr))
	case Update:
		_, err = client.Get().Update(ctx, db, op.Collection, driver.Filter(op.Filter), driver.Fields(op.Fields))
	case Delete:
		_, err = client.Get().Delete(ctx, db, op.Collection, driver.Filter(op.Filter))
	case Replace, InsertOrReplace:
		ptr := unsafe.Pointer(&op.Documents)
		_, err = client.Get().Replace(ctx, db, op.Collection, *(*[]driver.Document)(ptr))
	case CreateOrUpdateCollection:
		err = client.Get().CreateOrUpdateCollection(ctx, db, op.Collection, driver.Schema(op.Schema))
	case DropCollection:
		err = client.Get().DropCollection(ctx, db, op.Collection)
	case ListCollections:
		colls, err := client.Get().ListCollections(ctx, db)
		if err != nil {
			log.Fatal().Err(err).Str("type", op.Operation).Msgf("transact operation failed")
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
		it, err := client.Get().Read(ctx, db, op.Collection, driver.Filter(filter), driver.Fields(fields))
		if err != nil {
			log.Fatal().Err(err).Str("op", op.Operation).Msgf("transact operation failed")
		}
		var d driver.Document
		for it.Next(&d) {
			util.Stdout("%s\n", string(d))
		}
		return
	default:
		log.Fatal().Err(err).Str("op", op.Operation).Msgf("unknown operation type")
	}
	if err != nil {
		log.Fatal().Err(err).Str("op", op.Operation).Msgf("transact operation failed")
	}
}

var transactCmd = &cobra.Command{
	Use:     "transact {db} begin|commit|rollback|{operation}...|-",
	Aliases: []string{"tx"},
	Short:   "run a set of operations in a transaction",
	Long: `Run a set of operations in a transaction
Operations can be provided in the command line or from standard input`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		db := args[0]
		client.Transact(cmd.Context(), db, func(ctx context.Context, tx driver.Tx) {
			iterateInput(cmd.Context(), 1, args, func(ctx context.Context, args []string, ops []json.RawMessage) {
				for _, iop := range ops {
					var op TxOp
					if err := json.Unmarshal(iop, &op); err != nil {
						log.Fatal().Err(err).Msg("begin transaction failed")
					}

					if op.Operation != "" {
						execTxOp(ctx, db, op.Operation, &op.Op)
					}

					execTxOp(ctx, db, InsertOrReplace, op.InsertOrReplace)
					execTxOp(ctx, db, Replace, op.Replace)
					execTxOp(ctx, db, Insert, op.Insert)
					execTxOp(ctx, db, Read, op.Read)
					execTxOp(ctx, db, Update, op.Update)
					execTxOp(ctx, db, Delete, op.Delete)
					execTxOp(ctx, db, CreateOrUpdateCollection, op.CreateOrUpdateCollection)
					execTxOp(ctx, db, DropCollection, op.DropCollection)
					execTxOp(ctx, db, ListCollections, op.ListCollections)
				}
			})
		})
	},
}

func init() {
	dbCmd.AddCommand(transactCmd)
}
