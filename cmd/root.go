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
	"os"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/cmd/search"
	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/util"
)

var rootCmd = &cobra.Command{
	Use:   "tigris",
	Short: "tigris is a command line interface of Tigris data platform",
}

var dbCmd = &cobra.Command{
	Use:     "db",
	Short:   "Database related commands",
	Aliases: []string{"database"},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1) //nolint:revive
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&util.Quiet, "quiet", "q", false,
		"Suppress informational messages")

	rootCmd.AddCommand(search.RootCmd)
	rootCmd.AddCommand(dbCmd)
}

func addProjectFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&config.DefaultConfig.Project,
		"project", "p", "", "Specifies project: --project=my_proj1")

	if cmd != branchCmd {
		cmd.PersistentFlags().StringVar(&config.DefaultConfig.Branch,
			"branch", "", "Specifies branch: --branch=my_br1")
	}
}
