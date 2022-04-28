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

var insertCmd = &cobra.Command{
	Use:   "insert {db} {collection} {document}...|-",
	Short: "Inserts document(s)",
	Long: fmt.Sprintf(`Inserts one or more documents into a collection.

Examples:

  # Insert a single document into the users collection
  %[1]s insert testdb users '{"id": 1, "name": "Alice Alan"}

  # Insert multiple documents into the users collection
  %[1]s insert testdb users \
  '[
    {"id": 20, "name": "Jania McGrory"},
    {"id": 21, "name": "Bunny Instone"}
  ]'

  # Pass documents to insert via stdin
  # $ cat /home/alice/user_records.json
  # [
  #  {"id": 20, "name": "Jania McGrory"},
  #  {"id": 21, "name": "Bunny Instone"}
  # ]
  cat /home/alice/user_records.json | %[1]s insert testdb users -
`, rootCmd.Root().Name()),
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		iterateInput(cmd.Context(), cmd, 2, args, func(ctx context.Context, args []string, docs []json.RawMessage) {
			ptr := unsafe.Pointer(&docs)
			_, err := client.Get().Insert(ctx, args[0], args[1], *(*[]driver.Document)(ptr))
			if err != nil {
				util.Error(err, "insert documents failed")
			}
		})
	},
}

func init() {
	insertCmd.Flags().Int32VarP(&BatchSize, "batch_size", "b", BatchSize, "set batch size")
	dbCmd.AddCommand(insertCmd)
}
