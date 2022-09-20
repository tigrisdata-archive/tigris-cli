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
)

var createApplicationCmd = &cobra.Command{
	Use:   "application {name} {description}",
	Short: "Create application credentials",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		app, err := client.ManagementGet().CreateApplication(ctx, args[0], args[1])
		if err != nil {
			util.Error(err, "create application failed")
		}

		b, err := json.Marshal(app)
		if err != nil {
			util.Error(err, "create application failed")
		}

		util.Stdout("%s\n", string(b))
	},
}

var dropApplicationCmd = &cobra.Command{
	Use:   "application {id}",
	Short: "Drop application credentials",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		if err := client.ManagementGet().DeleteApplication(ctx, args[0]); err != nil {
			util.Error(err, "drop application failed")
		}

		util.Stdout("successfully dropped application credentials\n")
	},
}

var alterApplicationCmd = &cobra.Command{
	Use:   "application {id} {name} {description}",
	Short: "Alter application credentials",
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		app, err := client.ManagementGet().UpdateApplication(ctx, args[0], args[1], args[2])
		if err != nil {
			util.Error(err, "alter application failed")
		}

		b, err := json.Marshal(app)
		if err != nil {
			util.Error(err, "alter application failed")
		}

		util.Stdout("%s\n", string(b))
	},
}

var listApplicationsCmd = &cobra.Command{
	Use:   "applications",
	Short: "Lists applications",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		resp, err := client.ManagementGet().ListApplications(ctx)
		if err != nil {
			util.Error(err, "list collections failed")
		}
		for _, v := range resp {
			b, err := json.Marshal(v)
			if err != nil {
				util.Error(err, "list application failed")
			}
			util.Stdout("%s\n", string(b))
		}
	},
}

func init() {
	dropCmd.AddCommand(dropApplicationCmd)
	createCmd.AddCommand(createApplicationCmd)
	listCmd.AddCommand(listApplicationsCmd)
	alterCmd.AddCommand(alterApplicationCmd)
}
