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
	"encoding/json"
	"unsafe"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigrisdb-cli/client"
	"github.com/tigrisdata/tigrisdb-client-go/driver"
)

var replaceCmd = &cobra.Command{
	Use:     "replace {db} {collection} {document}...|{-}",
	Aliases: []string{"insert_or_replace"},
	Short:   "replace document",
	Long: `replace or insert one or multiple documents
		from command line or standard input`,
	Args: cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		iterateInput(cmd.Context(), 2, args, func(ctx context.Context, args []string, docs []json.RawMessage) {
			ptr := unsafe.Pointer(&docs)
			_, err := client.Get().Replace(ctx, args[0], args[1], *(*[]driver.Document)(ptr))
			if err != nil {
				log.Fatal().Err(err).Msg("replace documents failed")
			}
		})
	},
}

func init() {
	replaceCmd.Flags().Int32VarP(&BatchSize, "batch_size", "b", BatchSize, "set batch size")
	dbCmd.AddCommand(replaceCmd)
}
