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

var listCollectionsCmd = &cobra.Command{
	Use:   "collections",
	Short: "list collections",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		resp, err := client.Get().ListCollections(ctx, args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("list collections failed")
		}
		for _, v := range resp {
			util.Stdout(v)
		}
	},
}

var createCollectionCmd = &cobra.Command{
	Use:   "collection",
	Short: "create collection",
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		err := client.Get().CreateCollection(ctx, args[0], args[1], driver.Schema(args[2]), &driver.CollectionOptions{})
		if err != nil {
			log.Fatal().Err(err).Msg("create collection failed")
		}
	},
}

var dropCollectionCmd = &cobra.Command{
	Use:   "collection",
	Short: "drop collection",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		err := client.Get().DropCollection(ctx, args[0], args[1], &driver.CollectionOptions{})
		if err != nil {
			log.Fatal().Err(err).Msg("drop collection failed")
		}
	},
}

var alterCollectionCmd = &cobra.Command{
	Use:   "collection",
	Short: "alter collection",
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		err := client.Get().AlterCollection(ctx, args[0], args[1], driver.Schema(args[2]), &driver.CollectionOptions{})
		if err != nil {
			log.Fatal().Err(err).Msg("alter collection failed")
		}
	},
}

func init() {
	dropCmd.AddCommand(dropCollectionCmd)
	createCmd.AddCommand(createCollectionCmd)
	listCmd.AddCommand(listCollectionsCmd)
	alterCmd.AddCommand(alterCollectionCmd)
}
