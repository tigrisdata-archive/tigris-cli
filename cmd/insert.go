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

// insertCmd represents the insert command
var insertCmd = &cobra.Command{
	Use:   "insert",
	Short: "insert documents",
	Long:  `insert one or multiple documents`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		docs := make([]driver.Document, 0, len(args))
		for _, v := range args[2:] {
			docs = append(docs, driver.Document(v))
		}
		_, err := client.Get().Insert(ctx, args[0], args[1], docs, &driver.InsertOptions{})
		if err != nil {
			log.Fatal().Err(err).Msg("insert documents failed")
		}
	},
}

func init() {
	rootCmd.AddCommand(insertCmd)
}
