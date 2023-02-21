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
	"os"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/util"
)

var checkout bool

var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "Working with Tigris branches",
}

var createBranchCmd = &cobra.Command{
	Use:   "create {branch_name}",
	Short: "Creates Tigris branch",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			_, err := client.Get().UseDatabase(getProjectName()).CreateBranch(ctx, args[0])
			if err != nil {
				return util.Error(err, "create branch")
			}

			util.Infof("Branch %s created successfully", args[0])

			if checkout {
				checkoutBranch(args[0])
			}

			return nil
		})
	},
}

var deleteBranchCmd = &cobra.Command{
	Use:   "delete {branch_name}",
	Short: "Deletes Tigris branch",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			_, err := client.Get().UseDatabase(getProjectName()).DeleteBranch(ctx, args[0])
			if err != nil {
				return util.Error(err, "delete branch")
			}

			util.Infof("Branch %s deleted successfully", args[0])

			return nil
		})
	},
}

var listBranchesCmd = &cobra.Command{
	Use:   "list",
	Short: "List Tigris branches",
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.Get().DescribeDatabase(ctx, getProjectName())
			if err != nil {
				return util.Error(err, "list branches")
			}

			for _, v := range resp.Branches {
				util.Infof("%s", v)
			}

			return nil
		})
	},
}

var showBranchCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current Tigris branch",
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			branch := config.DefaultConfig.Branch

			if branch == "" {
				branch = "main"
			}

			util.Stdoutf("%s", branch)

			if util.IsTTY(os.Stdout) {
				util.Stdoutf("\n")
			}

			return nil
		})
	},
}

func checkoutBranch(name string) {
	config.DefaultConfig.Branch = name

	err := config.Save(config.DefaultName, config.DefaultConfig)
	util.Fatal(err, "saving branch config")

	util.Infof("Branch %s successfully checked-out", name)
}

var checkoutBranchCmd = &cobra.Command{
	Use:   "checkout {branch_name}",
	Short: "Checkout Tigris branch",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		checkoutBranch(args[0])
	},
}

func init() {
	rootCmd.AddCommand(branchCmd)

	addProjectFlag(branchCmd)

	createBranchCmd.Flags().BoolVarP(&checkout, "checkout", "c", false, "activate created branch")

	branchCmd.AddCommand(createBranchCmd)
	branchCmd.AddCommand(deleteBranchCmd)
	branchCmd.AddCommand(checkoutBranchCmd)
	branchCmd.AddCommand(listBranchesCmd)
	branchCmd.AddCommand(showBranchCmd)
}
