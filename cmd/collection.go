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
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/util"
	api "github.com/tigrisdata/tigris-client-go/api/server/v1"
	"github.com/tigrisdata/tigris-client-go/driver"
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
	Short: "Describes collection",
	Long:  "Returns collection schema including metadata",
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
	Short: "Lists database collections",
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
	Short: "Creates collection(s)",
	Long: fmt.Sprintf(`Creates collections with provided schema.

Examples:

  # Pass the schema as a string
  %[1]s create collection testdb '{
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

  # Create collection with schema from a file
  # $ cat /home/alice/users.json
  # {
  #  "title": "users",
  #  "description": "Collection of documents with details of users",
  #  "properties": {
  #    "id": {
  #      "description": "A unique identifier for the user",
  #      "type": "integer"
  #    },
  #    "name": {
  #      "description": "Name of the user",
  #      "type": "string",
  #      "maxLength": 100
  #    }
  #  },
  #  "primary_key": [
  #    "id"
  #  ]
  # }
  %[1]s create collection testdb </home/alice/users.json

  # Create collection with schema passed through stdin
  cat /home/alice/users.json | %[1]s create collection testdb -
  %[1]s describe collection sampledb users | jq .schema | %[1]s create collection testdb -
`, rootCmd.Root().Name()),
	Args: cobra.MinimumNArgs(1),
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
	Short: "Drops collection",
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
	Use:   "collection {db} {schema}",
	Short: "Updates collection schema",
	Long: fmt.Sprintf(`Updates collection schema.

Examples:

  # Pass the schema as a string
  %[1]s alter collection testdb '{
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
`, rootCmd.Root().Name()),
	Args: cobra.MinimumNArgs(1),
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
