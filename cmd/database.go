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
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/util"
	api "github.com/tigrisdata/tigrisdb-client-go/api/server/v1"
)

var listDatabasesCmd = &cobra.Command{
	Use:   "databases",
	Short: "list databases",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		resp, err := client.Get().ListDatabases(ctx)
		if err != nil {
			util.Error(err, "list databases failed")
		}
		for _, v := range resp {
			util.Stdout("%s\n", v)
		}
	},
}

type DescribeDatabaseResponse struct {
	Db          string                        `json:"db,omitempty"`
	Metadata    *api.DatabaseMetadata         `json:"metadata,omitempty"`
	Collections []*DescribeCollectionResponse `json:"collections,omitempty"`
}

var describeDatabaseCmd = &cobra.Command{
	Use:   "database {db}",
	Short: "describe database",
	Long:  "describe database returns metadata for all the collections in the database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		resp, err := client.Get().DescribeDatabase(ctx, args[0])
		if err != nil {
			util.Error(err, "describe collection failed")
		}

		schemaOnly, err := cmd.Flags().GetBool("schema-only")
		if err != nil {
			util.Error(err, "error reading the 'schema-only' option")
		}

		if schemaOnly {
			for _, v := range resp.Collections {
				util.Stdout("%s\n", string(v.Schema))
			}
		} else {
			tr := DescribeDatabaseResponse{
				Db:       resp.Db,
				Metadata: resp.Metadata,
			}

			for _, v := range resp.Collections {
				tr.Collections = append(tr.Collections, &DescribeCollectionResponse{
					Collection: v.Collection,
					Metadata:   v.Metadata,
					Schema:     v.Schema,
				})
			}

			b, err := json.Marshal(tr)
			if err != nil {
				util.Error(err, "describe database failed")
			}

			util.Stdout("%s\n", string(b))
		}
	},
}

var createDatabaseCmd = &cobra.Command{
	Use:   "database {db}",
	Short: "create database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		err := client.Get().CreateDatabase(ctx, args[0])
		if err != nil {
			util.Error(err, "create database failed")
		}
	},
}

var dropDatabaseCmd = &cobra.Command{
	Use:   "database {db}",
	Short: "drop database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		err := client.Get().DropDatabase(ctx, args[0])
		if err != nil {
			util.Error(err, "drop database failed")
		}
	},
}

func init() {
	describeDatabaseCmd.Flags().BoolP("schema-only", "s", false, "dump only schema of all database collections")
	dropCmd.AddCommand(dropDatabaseCmd)
	createCmd.AddCommand(createDatabaseCmd)
	listCmd.AddCommand(listDatabasesCmd)
	describeCmd.AddCommand(describeDatabaseCmd)
}
