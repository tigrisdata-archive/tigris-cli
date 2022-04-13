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
)

var listDatabasesCmd = &cobra.Command{
	Use:   "databases",
	Short: "list databases",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		resp, err := client.Get().ListDatabases(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("list databases failed")
		}
		for _, v := range resp {
			util.Stdout("%s\n", v)
		}
	},
}

var createDatabaseCmd = &cobra.Command{
	Use:   "database {db}",
	Short: "create database",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		err := client.Get().CreateDatabase(ctx, args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("create database failed")
		}
	},
}

var dropDatabaseCmd = &cobra.Command{
	Use:   "database {db}",
	Short: "drop database",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		err := client.Get().DropDatabase(ctx, args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("drop database failed")
		}
	},
}

func init() {
	dropCmd.AddCommand(dropDatabaseCmd)
	createCmd.AddCommand(createDatabaseCmd)
	listCmd.AddCommand(listDatabasesCmd)
}
