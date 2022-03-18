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
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigrisdb-cli/client"
	"github.com/tigrisdata/tigrisdb-cli/util"
	"github.com/tigrisdata/tigrisdb-client-go/driver"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete documents",
	Long:  `delete documents according to provided filter`,
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		_, err := client.Get().Delete(ctx, args[0], args[1], driver.Filter(args[2]), &driver.DeleteOptions{})
		if err != nil {
			log.Fatal().Err(err).Msg("delete documents failed")
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
