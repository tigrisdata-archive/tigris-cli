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

package search

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/iterate"
	"github.com/tigrisdata/tigris-cli/login"
	"github.com/tigrisdata/tigris-cli/util"
	"github.com/tigrisdata/tigris-client-go/driver"
)

var ErrSchemaNameMissing = fmt.Errorf("schema name is missing")

func createIndex(ctx context.Context, raw driver.Schema) error {
	type Schema struct {
		Name string `json:"title"`
	}

	var schema Schema

	if err := json.Unmarshal(raw, &schema); err != nil {
		util.Fatal(err, "error parsing index schema")
	}

	if schema.Name == "" {
		util.Fatal(ErrSchemaNameMissing, "create index")
	}

	err := client.GetSearch().CreateOrUpdateIndex(ctx, schema.Name, raw)

	return util.Error(err, "create index: %v", schema.Name)
}

// DescribeIndexResponse adapter to convert Schema field to json.RawMessage.
type DescribeIndexResponse struct {
	Index  string          `json:"index,omitempty"`
	Schema json.RawMessage `json:"schema,omitempty"`
}

var describeIndexCmd = &cobra.Command{
	Use:   "describe {index}",
	Short: "Describes index",
	Long:  "Returns index schema including metadata",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.GetSearch().GetIndex(ctx, args[0])
			if err != nil {
				return util.Error(err, "describe index")
			}

			tr := DescribeIndexResponse{
				Index: args[0],
				// Metadata:   resp.Metadata,
				Schema: resp.Schema,
			}

			err = util.PrettyJSON(tr)
			util.Fatal(err, "describe index marshal")

			return nil
		})
	},
}

var listIndexesCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists project indexes",
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.GetSearch().ListIndexes(ctx, nil)
			if err != nil {
				return util.Error(err, "list indexes")
			}

			for _, v := range resp {
				util.Stdoutf("%s\n", v)
			}

			return nil
		})
	},
}

var createIndexCmd = &cobra.Command{
	Use:     "create {schema}...|-",
	Aliases: []string{"indexes"},
	Short:   "Creates index(s)",
	Long:    "Creates indexes with provided schema.",
	Example: fmt.Sprintf(`
  # Pass the schema as a string
  %[1]s create index --project=myproj '{
	"title": "users",
	"description": "Index of documents with details of users",
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
	}
  }'

  # Create index with schema from a file
  # $ cat /home/alice/users.json
  # {
  #  "title": "users",
  #  "description": "Index of documents with details of users",
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
  #  }
  # }
  %[1]s create index --project=myproj </home/alice/users.json

  # Create index with schema passed through stdin
  cat /home/alice/users.json | %[1]s create index myproj -
  %[1]s describe index --project=myproj users | jq .schema | %[1]s create index myproj -
`, "tigris search"),
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			return iterate.Input(ctx, cmd, 0, args, func(ctx context.Context, args []string, docs []json.RawMessage) error {
				for _, v := range docs {
					if err := createIndex(ctx, driver.Schema(v)); err != nil {
						return util.Error(err, "create index %v", string(v))
					}
				}

				return nil
			})
		})
	},
}

var deleteIndexCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete index",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			return iterate.Input(ctx, cmd, 0, args, func(ctx context.Context, args []string, docs []json.RawMessage) error {
				for _, v := range docs {
					if err := client.GetSearch().DeleteIndex(ctx, string(v)); err != nil {
						return util.Error(err, "delete index")
					}
					util.Infof("deleted index: %s", string(v))
				}

				return nil
			})
		})
	},
}

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Search index related commands",
}

func init() {
	addProjectFlag(deleteIndexCmd)
	addProjectFlag(createIndexCmd)
	addProjectFlag(listIndexesCmd)
	addProjectFlag(describeIndexCmd)

	indexCmd.AddCommand(deleteIndexCmd)
	indexCmd.AddCommand(createIndexCmd)
	indexCmd.AddCommand(listIndexesCmd)
	indexCmd.AddCommand(describeIndexCmd)

	RootCmd.AddCommand(indexCmd)
}
