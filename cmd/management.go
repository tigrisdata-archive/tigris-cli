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
	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/util"
)

var createApplicationCmd = &cobra.Command{
	Use:   "application {name} {description}",
	Short: "Create application credentials",
	Long: `Creates new application credentials.
The output contains client_id and client_secret,
which can be used to authenticate using our official client SDKs.
Set the client_id and client_secret in the configuration of the corresponding SDK
Check the docs for more information: https://docs.tigrisdata.com/overview/authentication
`,
	Example: `
  tigris create application service1 "main api service"

  Output:

  {
    "id": "<client id here>",
    "name": "service2",
    "description": "main api service",
    "secret": "<client secret here",
    "created_at": 1663802082000,
    "created_by": "github|3436058"
  }`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		app, err := client.ManagementGet().CreateApplication(ctx, args[0], args[1])
		if err != nil {
			util.Error(err, "create application failed")
		}

		if err := util.PrettyJSON(app); err != nil {
			util.Error(err, "create application failed")
		}
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
	Example: `
tigris alter application <client id> new_name1 "new descr1"

Output:
{
  "id": "<client id>",
  "name": "new_name1",
  "description": "new descr1",
  "secret": "<client secrete here",
  "created_at": 1663802082000,
  "created_by": "github|3436058"
}
`,
	Args: cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()

		_, err := client.ManagementGet().UpdateApplication(ctx, args[0], args[1], args[2])
		if err != nil {
			util.Error(err, "alter application failed")
		}

		sec, err := client.ManagementGet().RotateClientSecret(ctx, args[0])
		if err != nil {
			util.Error(err, "alter application failed")
		}

		if err := util.PrettyJSON(sec); err != nil {
			util.Error(err, "alter application failed")
		}
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
			if err := util.PrettyJSON(v); err != nil {
				util.Error(err, "list applications failed")
			}
		}
	},
}

func init() {
	dropCmd.AddCommand(dropApplicationCmd)
	createCmd.AddCommand(createApplicationCmd)
	listCmd.AddCommand(listApplicationsCmd)
	alterCmd.AddCommand(alterApplicationCmd)
}
