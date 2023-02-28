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
	"github.com/tigrisdata/tigris-cli/login"
	"github.com/tigrisdata/tigris-cli/util"
)

var deleteProjectCmd = &cobra.Command{
	Use:   "delete-project {name}",
	Short: "Deletes project",
	Long:  "Deletes project and all resources inside project.",
	Args:  cobra.ExactArgs(1),
	Example: fmt.Sprintf(`
  # Delete project named 'test-project'
  %[1]s delete-project test-project'

  # Delete project named 'test-project' (without user prompt)
  %[1]s delete-project test-project' --force
#
`, rootCmd.Root().Name()),
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			var userInput string
			if !forceDelete {
				util.Stdoutf("Are you sure you want to delete the project? (y/n)")
				_, err := fmt.Scanln(&userInput)
				util.Fatal(err, "delete-project")
			}
			if forceDelete || userInput == "y" || userInput == "Y" {
				_, err := client.Get().DeleteProject(ctx, args[0])

				return util.Error(err, "delete-project")
			}

			return nil
		})
	},
}

var forceDelete bool

func init() {
	deleteProjectCmd.PersistentFlags().BoolVarP(&forceDelete, "force", "f", false,
		"Skips user prompt and deletes the project")
	dbCmd.AddCommand(deleteProjectCmd)
}
