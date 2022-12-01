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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/util"
)

var (
	rotate bool

	ErrWrongArgs = fmt.Errorf("please provide name and description to update or use --rotate to rotate the secret")
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
    "name": "service1",
    "description": "main api service",
    "secret": "<client secret here",
    "created_at": 1663802082000,
    "created_by": "github|3436058"
  }

  tigris create application service2

  Output:

  {
    "id": "<client id here>",
    "name": "service2",
    "secret": "<client secret here",
    "created_at": 1663802082001,
    "created_by": "github|3436058"
  }
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			description := ""
			if len(args) > 1 {
				description = args[1]
			}
			app, err := client.ManagementGet().CreateApplication(ctx, args[0], description)
			if err != nil {
				return util.Error(err, "create application failed")
			}

			err = util.PrettyJSON(app)
			util.Fatal(err, "create application")

			return nil
		})
	},
}

var dropApplicationCmd = &cobra.Command{
	Use:   "application {id}",
	Short: "Drop application credentials",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			if err := client.ManagementGet().DeleteApplication(ctx, args[0]); err != nil {
				return util.Error(err, "drop application failed")
			}

			util.Stdoutf("successfully dropped application credentials\n")

			return nil
		})
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
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			// no name/descr and no explicit --rotate
			if len(args) < 2 && !rotate {
				util.Fatal(ErrWrongArgs, "alter application failed")
			}

			if len(args) >= 1 {
				desc := ""
				if len(args) > 2 {
					desc = args[2]
				}
				_, err := client.ManagementGet().UpdateApplication(ctx, args[0], args[1], desc)
				if err != nil {
					return util.Error(err, "alter application failed")
				}
			}

			// rotate only when explicitly requested
			if rotate {
				sec, err := client.ManagementGet().RotateClientSecret(ctx, args[0])
				if err != nil {
					return util.Error(err, "alter application failed")
				}

				err = util.PrettyJSON(sec)
				util.Fatal(err, "alter application failed")
			}

			return nil
		})
	},
}

var listApplicationsCmd = &cobra.Command{
	Use:   "applications [name]",
	Short: "Lists applications",
	Long:  "Lists available applications. Optional parameter allows to return only the application with the given name.",
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.ManagementGet().ListApplications(ctx)
			if err != nil {
				return util.Error(err, "list applications failed")
			}

			if len(args) > 0 {
				for _, v := range resp {
					if v.Name == args[0] {
						err := util.PrettyJSON(v)
						util.Fatal(err, "list applications filtered")
					}
				}
			} else {
				err := util.PrettyJSON(resp)
				util.Fatal(err, "list applications")
			}

			return nil
		})
	},
}

var listNamespacesCmd = &cobra.Command{
	Use:   "namespaces",
	Short: "Lists namespaces",
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.ManagementGet().ListNamespaces(ctx)
			if err != nil {
				return util.Error(err, "list namespaces")
			}

			err = util.PrettyJSON(resp)
			util.Fatal(err, "list namespaces failed")

			return nil
		})
	},
}

var createNamespaceCmd = &cobra.Command{
	Use:   "namespace {name}",
	Short: "Create namespace",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			if err := client.ManagementGet().CreateNamespace(ctx, args[0]); err != nil {
				return util.Error(err, "create namespace failed")
			}

			util.Stdoutf("namespace successfully created\n")

			return nil
		})
	},
}

func init() {
	alterApplicationCmd.Flags().BoolVarP(&rotate, "rotate", "r", false, "Rotate application secret")

	addProjectFlag(dropApplicationCmd)
	addProjectFlag(createApplicationCmd)
	addProjectFlag(listApplicationsCmd)
	addProjectFlag(alterApplicationCmd)

	dropCmd.AddCommand(dropApplicationCmd)
	createCmd.AddCommand(createApplicationCmd)
	listCmd.AddCommand(listApplicationsCmd)
	alterCmd.AddCommand(alterApplicationCmd)

	listCmd.AddCommand(listNamespacesCmd)
	createCmd.AddCommand(createNamespaceCmd)
}
