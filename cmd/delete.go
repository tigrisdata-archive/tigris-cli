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

var deleteCmd = &cobra.Command{
	Use:   "delete {db} {collection} {filter}",
	Short: "Deletes document(s)",
	Long: fmt.Sprintf(`Deletes documents according to the provided filter.

Examples:

  # Delete a user where the value of the id field is 2
  %[1]s delete testdb users '{"id": 2}'

  # Delete users where the value of id field is 1 or 3
  %[1]s delete testdb users '{"$or": [{"id": 1}, {"id": 3}]}'
`, rootCmd.Root().Name()),
	Args: cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		_, err := client.Get().Delete(ctx, args[0], args[1], driver.Filter(args[2]))
		if err != nil {
			util.Error(err, "delete documents failed")
		}
	},
}

func init() {
	dbCmd.AddCommand(deleteCmd)
}
