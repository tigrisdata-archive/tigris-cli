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

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/util"
	api "github.com/tigrisdata/tigris-client-go/api/server/v1"
)

var schemaOnly bool

var listDatabasesCmd = &cobra.Command{
	Use:   "databases",
	Short: "Lists databases",
	Run: func(cmd *cobra.Command, _ []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.Get().ListDatabases(ctx)
			if err != nil {
				return util.Error(err, "list databases")
			}

			for _, v := range resp {
				util.Stdoutf("%s\n", v)
			}

			return nil
		})
	},
}

// DescribeDatabaseResponse adapter to convert schema to json.RawMessage.
type DescribeDatabaseResponse struct {
	DB          string                        `json:"db,omitempty"`
	Metadata    *api.DatabaseMetadata         `json:"metadata,omitempty"`
	Collections []*DescribeCollectionResponse `json:"collections,omitempty"`
}

var describeDatabaseCmd = &cobra.Command{
	Use:   "database {db}",
	Short: "Describes database",
	Long:  "Returns schema and metadata for all the collections in the database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.Get().DescribeDatabase(ctx, args[0])
			if err != nil {
				return util.Error(err, "describe collection failed")
			}

			if schemaOnly {
				for _, v := range resp.Collections {
					util.Stdoutf("%s\n", string(v.Schema))
				}
			} else {
				tr := DescribeDatabaseResponse{
					DB: resp.Db,
					// Metadata: resp.Metadata,
				}

				for _, v := range resp.Collections {
					tr.Collections = append(tr.Collections, &DescribeCollectionResponse{
						Collection: v.Collection,
						// Metadata:   v.Metadata,
						Schema: v.Schema,
					})
				}

				b, err := json.Marshal(tr)
				util.Fatal(err, "describe database")

				util.Stdoutf("%s\n", string(b))
			}

			return nil
		})
	},
}

var createDatabaseCmd = &cobra.Command{
	Use:   "database {db}",
	Short: "Creates database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			err := client.Get().CreateDatabase(ctx, args[0])
			return util.Error(err, "create database")
		})
	},
}

var dropDatabaseCmd = &cobra.Command{
	Use:   "database {db}",
	Short: "Drops database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			err := client.Get().DropDatabase(ctx, args[0])
			return util.Error(err, "drop database")
		})
	},
}

func init() {
	describeDatabaseCmd.Flags().BoolVarP(&schemaOnly, "schema-only", "s", false,
		"dump only schema of all database collections")

	dropCmd.AddCommand(dropDatabaseCmd)
	createCmd.AddCommand(createDatabaseCmd)
	listCmd.AddCommand(listDatabasesCmd)
	describeCmd.AddCommand(describeDatabaseCmd)
}
