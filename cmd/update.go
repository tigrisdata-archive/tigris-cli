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

package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	login "github.com/tigrisdata/tigris-cli/login"
	"github.com/tigrisdata/tigris-cli/util"
	"github.com/tigrisdata/tigris-client-go/driver"
)

var updateCmd = &cobra.Command{
	Use:   "update {collection} {filter} {fields}",
	Short: "Updates document(s)",
	Long:  "Updates the field values in documents according to provided filter.",
	Example: fmt.Sprintf(`
  # Update the field "name" of user where the value of the id field is 2
  %[1]s update --project=myproj users '{"id": 19}' '{"$set": {"name": "Updated New User"}}'
`, rootCmd.Root().Name()),
	Args: cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			_, err := client.GetDB().
				Update(ctx, args[0], driver.Filter(args[1]), driver.Update(args[2]))

			return util.Error(err, "update documents failed")
		})
	},
}

func init() {
	addProjectFlag(updateCmd)
	rootCmd.AddCommand(updateCmd)
}
