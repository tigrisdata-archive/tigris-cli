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
	"github.com/tigrisdata/tigris-cli/config"
	login "github.com/tigrisdata/tigris-cli/login"
	"github.com/tigrisdata/tigris-cli/util"
	"github.com/tigrisdata/tigris-client-go/driver"
)

var (
	rotate bool

	ErrWrongArgs   = fmt.Errorf("please provide name and description to update or use --rotate to rotate the secret")
	ErrAppNotFound = fmt.Errorf("app key not found")
)

var createAppKeyCmd = &cobra.Command{
	Use:   "app_key {name} {description}",
	Short: "Create app_key credentials",
	Long: `Creates new app_key credentials.
The output contains client_id and client_secret,
which can be used to authenticate using our official client SDKs.
Set the client_id and client_secret in the configuration of the corresponding SDK
Check the docs for more information: https://docs.tigrisdata.com/overview/authentication
`,
	Example: `
  tigris create app_key service1 "main api service"

  Output:

  {
    "id": "<client id here>",
    "name": "service1",
    "description": "main api service",
    "secret": "<client secret here",
    "created_at": 1663802082000,
    "created_by": "github|3436058"
  }

  tigris create app_key service2

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
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			description := ""
			if len(args) > 1 {
				description = args[1]
			}
			app, err := client.Get().CreateAppKey(ctx, config.GetProjectName(), args[0], description)
			if err != nil {
				return util.Error(err, "create app_key failed")
			}

			err = util.PrettyJSON(app)
			util.Fatal(err, "create app_key")

			return nil
		})
	},
}

var dropAppKeyCmd = &cobra.Command{
	Use:   "app_key {id}",
	Short: "Drop app_key credentials",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			if err := client.Get().DeleteAppKey(ctx, config.GetProjectName(), args[0]); err != nil {
				return util.Error(err, "drop app_key failed")
			}

			util.Stdoutf("successfully dropped app_key credentials\n")

			return nil
		})
	},
}

var alterAppKeyCmd = &cobra.Command{
	Use:   "app_key {id} {name} {description}",
	Short: "Alter app_key credentials",
	Example: `
tigris alter app_key <client id> new_name1 "new descr1"

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
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			// no name/descr and no explicit --rotate
			if len(args) < 2 && !rotate {
				util.Fatal(ErrWrongArgs, "alter app_key failed")
			}

			if len(args) >= 1 {
				desc := ""
				if len(args) > 2 {
					desc = args[2]
				}
				_, err := client.Get().UpdateAppKey(ctx, config.GetProjectName(), args[0], args[1], desc)
				if err != nil {
					return util.Error(err, "alter app_key failed")
				}
			}

			// rotate only when explicitly requested
			if rotate {
				sec, err := client.Get().RotateAppKeySecret(ctx, config.GetProjectName(), args[0])
				if err != nil {
					return util.Error(err, "alter app_key failed")
				}

				err = util.PrettyJSON(sec)
				util.Fatal(err, "alter app_key failed")
			}

			return nil
		})
	},
}

func getAppKey(ctx context.Context, filter string) (*driver.AppKey, error) {
	resp, err := client.Get().ListAppKeys(ctx, config.GetProjectName())
	if err != nil {
		return nil, util.Error(err, "list app_key failed")
	}

	for _, v := range resp {
		if v.Name == filter {
			return v, nil
		}
	}

	return nil, ErrAppNotFound
}

var listAppKeysCmd = &cobra.Command{
	Use:   "app_keys [name]",
	Short: "Lists app keys",
	Long:  "Lists available app keys. Optional parameter allows to return only the app key with the given name.",
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			if len(args) > 0 {
				app, err := getAppKey(ctx, args[0])
				if err != nil {
					return err
				}

				err = util.PrettyJSON(app)
				util.Fatal(err, "list app_keys")
			} else {
				resp, err := client.Get().ListAppKeys(ctx, config.GetProjectName())
				if err != nil {
					return util.Error(err, "list app_keys failed")
				}

				err = util.PrettyJSON(resp)
				util.Fatal(err, "list app_keys")
			}

			return nil
		})
	},
}

var listNamespacesCmd = &cobra.Command{
	Use:   "namespaces",
	Short: "Lists namespaces",
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
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
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			if err := client.ManagementGet().CreateNamespace(ctx, args[0]); err != nil {
				return util.Error(err, "create namespace failed")
			}

			util.Stdoutf("namespace successfully created\n")

			return nil
		})
	},
}

func init() {
	alterAppKeyCmd.Flags().BoolVarP(&rotate, "rotate", "r", false, "Rotate app key secret")

	addProjectFlag(dropAppKeyCmd)
	addProjectFlag(createAppKeyCmd)
	addProjectFlag(listAppKeysCmd)
	addProjectFlag(alterAppKeyCmd)

	dropCmd.AddCommand(dropAppKeyCmd)
	createCmd.AddCommand(createAppKeyCmd)
	listCmd.AddCommand(listAppKeysCmd)
	alterCmd.AddCommand(alterAppKeyCmd)

	listCmd.AddCommand(listNamespacesCmd)
	createCmd.AddCommand(createNamespaceCmd)
}
