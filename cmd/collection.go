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

var ErrSchemaNameMissing = fmt.Errorf("schema name is missing")

func createCollection(ctx context.Context, tx driver.Tx, raw driver.Schema) error {
	type Schema struct {
		Name string `json:"title"`
	}

	var schema Schema

	if err := json.Unmarshal(raw, &schema); err != nil {
		util.Fatal(err, "error parsing collection schema")
	}

	if schema.Name == "" {
		util.Fatal(ErrSchemaNameMissing, "create collection")
	}

	err := tx.CreateOrUpdateCollection(ctx, schema.Name, raw)

	return util.Error(err, "create collection: %v", schema.Name)
}

// DescribeCollectionResponse adapter to convert Schema field to json.RawMessage.
type DescribeCollectionResponse struct {
	Collection string                  `json:"collection,omitempty"`
	Metadata   *api.CollectionMetadata `json:"metadata,omitempty"`
	Schema     json.RawMessage         `json:"schema,omitempty"`
}

var describeCollectionCmd = &cobra.Command{
	Use:   "collection {collection}",
	Short: "Describes collection",
	Long:  "Returns collection schema including metadata",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.Get().UseDatabase(getProjectName()).DescribeCollection(ctx, args[0],
				&driver.DescribeCollectionOptions{SchemaFormat: format})
			if err != nil {
				return util.Error(err, "describe collection")
			}

			tr := DescribeCollectionResponse{
				Collection: resp.Collection,
				// Metadata:   resp.Metadata,
				Schema: resp.Schema,
			}

			err = util.PrettyJSON(tr)
			util.Fatal(err, "describe collection marshal")

			return nil
		})
	},
}

var listCollectionsCmd = &cobra.Command{
	Use:   "collections",
	Short: "Lists project collections",
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.Get().UseDatabase(getProjectName()).ListCollections(ctx)
			if err != nil {
				return util.Error(err, "list collections")
			}

			for _, v := range resp {
				util.Stdoutf("%s\n", v)
			}

			return nil
		})
	},
}

var createCollectionCmd = &cobra.Command{
	Use:     "collection {schema}...|-",
	Aliases: []string{"collections"},
	Short:   "Creates collection(s)",
	Long:    "Creates collections with provided schema.",
	Example: fmt.Sprintf(`
  # Pass the schema as a string
  %[1]s create collection --project=testdb '{
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
  %[1]s create collection --project=testdb </home/alice/users.json

  # Create collection with schema passed through stdin
  cat /home/alice/users.json | %[1]s create collection testdb -
  %[1]s describe collection --project=testdb users | jq .schema | %[1]s create collection testdb -
`, rootCmd.Root().Name()),
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			return client.Transact(ctx, getProjectName(), func(ctx context.Context, tx driver.Tx) error {
				return iterateInput(ctx, cmd, 0, args, func(ctx context.Context, args []string, docs []json.RawMessage) error {
					for _, v := range docs {
						if err := createCollection(ctx, tx, driver.Schema(v)); err != nil {
							return util.Error(err, "create collection %v", string(v))
						}
					}

					return nil
				})
			})
		})
	},
}

var dropCollectionCmd = &cobra.Command{
	Use:   "collection",
	Short: "Drops collection",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			return client.Transact(ctx, getProjectName(), func(ctx context.Context, tx driver.Tx) error {
				return iterateInput(ctx, cmd, 0, args, func(ctx context.Context, args []string, docs []json.RawMessage) error {
					for _, v := range docs {
						if err := tx.DropCollection(ctx, string(v)); err != nil {
							return util.Error(err, "drop collection")
						}
						util.Infof("dropped collection: %s", string(v))
					}

					return nil
				})
			})
		})
	},
}

var alterCollectionCmd = &cobra.Command{
	Use:   "collection {schema}",
	Short: "Updates collection schema",
	Long:  "Updates collection schema.",
	Example: fmt.Sprintf(`
  # Pass the schema as a string
  %[1]s alter collection --project=testdb '{
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
		withLogin(cmd.Context(), func(ctx context.Context) error {
			return client.Transact(ctx, getProjectName(), func(ctx context.Context, tx driver.Tx) error {
				return iterateInput(ctx, cmd, 1, args, func(ctx context.Context, args []string, docs []json.RawMessage) error {
					for _, v := range docs {
						if err := createCollection(ctx, tx, driver.Schema(v)); err != nil {
							return err
						}
					}

					return nil
				})
			})
		})
	},
}

func init() {
	addProjectFlag(dropCollectionCmd)
	addProjectFlag(createCollectionCmd)
	addProjectFlag(listCollectionsCmd)
	addProjectFlag(alterCollectionCmd)
	addProjectFlag(describeCollectionCmd)

	describeCollectionCmd.Flags().StringVarP(&format, "format", "f", "",
		"output schema in the requested format: go, typescript, java")

	dropCmd.AddCommand(dropCollectionCmd)
	createCmd.AddCommand(createCollectionCmd)
	listCmd.AddCommand(listCollectionsCmd)
	alterCmd.AddCommand(alterCollectionCmd)
	describeCmd.AddCommand(describeCollectionCmd)
}
