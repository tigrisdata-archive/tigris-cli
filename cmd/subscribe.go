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
	"github.com/tigrisdata/tigris-client-go/driver"
)

var subscribeLimit int32

var subscribeCmd = &cobra.Command{
	Use:     "subscribe {db} {collection} {filter}",
	Short:   "Subscribes to published messages",
	Long:    "Streams messages in real-time until cancelled.",
	Example: fmt.Sprintf("%[1]s subscribe testdb", rootCmd.Root().Name()),
	Args:    cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			it, err := client.Get().UseDatabase(args[0]).Subscribe(ctx, args[1], driver.Filter(args[2]))
			if err != nil {
				return util.Error(err, "subscribe messages failed")
			}

			var doc driver.Document
			for i := int32(0); (subscribeLimit == 0 || i < subscribeLimit) && it.Next(&doc); i++ {
				util.Stdoutf("%s\n", string(doc))
			}

			return util.Error(it.Err(), "iterate messages failed")
		})
	},
}

func init() {
	subscribeCmd.Flags().Int32VarP(&subscribeLimit, "limit", "l", 0, "limit number of results returned")

	dbCmd.AddCommand(subscribeCmd)
}
