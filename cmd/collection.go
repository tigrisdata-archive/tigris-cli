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

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigrisdb-cli/client"
	"github.com/tigrisdata/tigrisdb-cli/util"
	api "github.com/tigrisdata/tigrisdb-client-go/api/server/v1"
	"github.com/tigrisdata/tigrisdb-client-go/driver"
)

func createCollection(ctx context.Context, tx driver.Tx, raw driver.Schema) {
	type Schema struct {
		Name string `json:"title"`
	}
	var schema Schema
	if err := json.Unmarshal(raw, &schema); err != nil {
		util.Error(err, "error parsing collection schema")
	}
	if schema.Name == "" {
		util.Error(fmt.Errorf("schema name is missing"), "create collection failed")
	}
	err := tx.CreateOrUpdateCollection(ctx, schema.Name, raw)
	if err != nil {
		util.Error(err, "create collection failed")
	}
}

type DescribeCollectionResponse struct {
	Collection string                  `json:"collection,omitempty"`
	Metadata   *api.CollectionMetadata `json:"metadata,omitempty"`
	Schema     json.RawMessage         `json:"schema,omitempty"`
}

var describeCollectionCmd = &cobra.Command{
	Use:   "collection {db} {collection}",
	Short: "describe collection",
	Long:  "describe collection returns collection metadata, including schema",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		resp, err := client.Get().DescribeCollection(ctx, args[0], args[1])
		if err != nil {
			util.Error(err, "describe collection failed")
		}

		tr := DescribeCollectionResponse{
			Collection: resp.Collection,
			Metadata:   resp.Metadata,
			Schema:     resp.Schema,
		}

		b, err := json.Marshal(tr)
		if err != nil {
			util.Error(err, "describe collection failed")
		}

		util.Stdout("%s\n", string(b))
	},
}

var listCollectionsCmd = &cobra.Command{
	Use:   "collections {db}",
	Short: "list database collections",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		resp, err := client.Get().ListCollections(ctx, args[0])
		if err != nil {
			util.Error(err, "list collections failed")
		}
		for _, v := range resp {
			util.Stdout("%s\n", v)
		}
	},
}

var createCollectionCmd = &cobra.Command{
	Use:   "collection {db} {schema}...|-",
	Short: "create collection(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client.Transact(cmd.Context(), args[0], func(ctx context.Context, tx driver.Tx) {
			iterateInput(ctx, cmd, 1, args, func(ctx context.Context, args []string, docs []json.RawMessage) {
				for _, v := range docs {
					createCollection(ctx, tx, driver.Schema(v))
				}
			})
		})
	},
}

var dropCollectionCmd = &cobra.Command{
	Use:   "collection {db}",
	Short: "drop collection",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client.Transact(cmd.Context(), args[0], func(ctx context.Context, tx driver.Tx) {
			iterateInput(ctx, cmd, 1, args, func(ctx context.Context, args []string, docs []json.RawMessage) {
				for _, v := range docs {
					err := tx.DropCollection(ctx, string(v))
					if err != nil {
						util.Error(err, "drop collection failed")
					}
				}
			})
		})
	},
}

var alterCollectionCmd = &cobra.Command{
	Use:   "collection {db} {collection} {schema}",
	Short: "update collection schema",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client.Transact(cmd.Context(), args[0], func(ctx context.Context, tx driver.Tx) {
			iterateInput(ctx, cmd, 1, args, func(ctx context.Context, args []string, docs []json.RawMessage) {
				for _, v := range docs {
					createCollection(ctx, tx, driver.Schema(v))
				}
			})
		})
	},
}

func init() {
	dropCmd.AddCommand(dropCollectionCmd)
	createCmd.AddCommand(createCollectionCmd)
	listCmd.AddCommand(listCollectionsCmd)
	alterCmd.AddCommand(alterCollectionCmd)
	describeCmd.AddCommand(describeCollectionCmd)
}
