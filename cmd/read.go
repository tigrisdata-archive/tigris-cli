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
	"github.com/tigrisdata/tigris-client-go/driver"
)

var readCmd = &cobra.Command{
	Use:   "read {db} {collection} [filter [fields]]",
	Short: "read documents",
	Long: `read documents according to provided filter and fields
if filter is not provided or has special {} value, read returns all documents in the collection
if fields is not provided or has special {} value, read returns all the fields of the document
`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		filter := `{}`
		fields := `{}`
		if len(args) > 2 {
			filter = args[2]
		}
		if len(args) > 3 {
			fields = args[3]
		}
		it, err := client.Get().Read(ctx, args[0], args[1], driver.Filter(filter), driver.Projection(fields))
		if err != nil {
			util.Error(err, "read documents failed")
		}
		var doc driver.Document
		for it.Next(&doc) {
			util.Stdout("%s\n", string(doc))
		}
		if err := it.Err(); err != nil {
			util.Error(err, "iterate documents failed")
		}
	},
}

func init() {
	dbCmd.AddCommand(readCmd)
}
