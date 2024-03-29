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

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/config"
	login "github.com/tigrisdata/tigris-cli/login"
	"github.com/tigrisdata/tigris-cli/util"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows tigris cli version",
	Run: func(cmd *cobra.Command, args []string) {
		util.Stdoutf("tigris version %s\n", util.Version)
	},
}

var serverVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Returns server's version",
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.Get().Info(ctx)
			if err != nil {
				return util.Error(err, "get server info")
			}

			util.Stdoutf("tigris server version at %s is %s\n", config.DefaultConfig.URL, resp.ServerVersion)

			return nil
		})
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
