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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/util"
	"github.com/tigrisdata/tigris-client-go/driver"
)

var limit int64
var skip int64

var readCmd = &cobra.Command{
	Use:   "read {db} {collection} {filter} {fields}",
	Short: "Reads and outputs documents",
	Long: `Reads documents according to provided filter and fields. 
If filter is not provided or an empty json document {} is passed as a filter, all documents in the collection are returned.
If fields are not provided or an empty json document {} is passed as fields, all the fields of the documents are selected.`,
	Example: fmt.Sprintf(`
  # Read a user document where id is 20
  # The output would be 
  #  {"id": 20, "name": "Jania McGrory"}
  %[1]s read testdb users '{"id": 20}'

  # Read user documents where id is 2 or 4
  # The output would be
  #  {"id": 2, "name": "Alice Wong"}
  #  {"id": 4, "name": "Jigar Joshi"}
  %[1]s read testdb users '{"$or": [{"id": 2}, {"id": 4}]}'

  # Read all documents in the user collection
  # The output would be
  #  {"id": 2, "name": "Alice Wong"}
  #  {"id": 4, "name": "Jigar Joshi"}
  #  {"id": 20, "name": "Jania McGrory"}
  #  {"id": 21, "name": "Bunny Instone"}
  %[1]s read testdb users
`, rootCmd.Root().Name()),
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		filter := `{}`
		fields := `{}`
		if len(args) > 2 {
			filter = args[2]
		}
		if len(args) > 3 {
			fields = args[3]
		}
		it, err := client.Get().UseDatabase(args[0]).Read(ctx, args[1],
			driver.Filter(filter),
			driver.Projection(fields),
			&driver.ReadOptions{Limit: limit, Skip: skip},
		)
		if err != nil {
			util.Error(err, "read documents failed")
		}
		var doc driver.Document
		for it.Next(&doc) {
			// Document came through GRPC may have \n at the end already
			if doc[len(doc)-1] == 0x0A {
				util.Stdout("%s", string(doc))
			} else {
				util.Stdout("%s\n", string(doc))
			}
		}
		if err := it.Err(); err != nil {
			util.Error(err, "iterate documents failed")
		}
	},
}

func init() {
	readCmd.Flags().Int64VarP(&limit, "limit", "l", 0, "limit number of returned results")
	readCmd.Flags().Int64VarP(&skip, "skip", "s", 0, "skip this many results in the beginning of the result set")
	dbCmd.AddCommand(readCmd)
}
