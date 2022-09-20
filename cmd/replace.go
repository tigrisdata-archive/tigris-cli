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

var replaceCmd = &cobra.Command{
	Use:     "replace {db} {collection} {document}...|-",
	Aliases: []string{"insert_or_replace"},
	Short:   "Inserts or replaces document(s)",
	Long:    "Inserts new documents or replaces existing documents.",
	Example: fmt.Sprintf(`
  # Insert new documents
  %[1]s replace testdb users '{"id": 1, "name": "John Wong"}'

  # Replace existing document
  # Existing document with the following data will get replaced
  #  {"id": 20, "name": "Jania McGrory"}
  %[1]s replace testdb users '{"id": 20, "name": "Alice Alan"}'

  # Insert or replace multiple documents
  # Existing document with the following data will get replaced
  #  {"id": 20, "name": "Alice Alan"}
  #  {"id": 21, "name": "Bunny Instone"}
  # While the new document {"id": 19, "name": "New User"} will get inserted
  %[1]s replace testdb users \
  '[
    {"id": 19, "name": "New User"},
    {"id": 20, "name": "Replaced Alice Alan"},
    {"id": 21, "name": "Replaced Bunny Instone"}
  ]'
`, rootCmd.Root().Name()),
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		iterateInput(cmd.Context(), cmd, 2, args, func(ctx context.Context, args []string, docs []json.RawMessage) {
			ptr := unsafe.Pointer(&docs)
			_, err := client.Get().UseDatabase(args[0]).Replace(ctx, args[1], *(*[]driver.Document)(ptr))
			if err != nil {
				util.Error(err, "replace documents failed")
			}
		})
	},
}

func init() {
	replaceCmd.Flags().Int32VarP(&BatchSize, "batch_size", "b", BatchSize, "set batch size")
	dbCmd.AddCommand(replaceCmd)
}
