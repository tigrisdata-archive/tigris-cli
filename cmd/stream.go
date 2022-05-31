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
	"encoding/json"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/util"
	"github.com/tigrisdata/tigris-client-go/driver"
)

var streamCmd = &cobra.Command{
	Use:     "stream {db} [collection]",
	Short:   "Streams and outputs events",
	Long:    "Streams events in real-time until cancelled.",
	Example: fmt.Sprintf("%[1]s stream testdb", rootCmd.Root().Name()),
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		var collection string
		if len(args) > 1 {
			collection = args[1]
		}

		it, err := client.Get().UseDatabase(args[0]).Stream(ctx, collection)
		if err != nil {
			util.Error(err, "stream events failed")
		}
		var event driver.Event
		for it.Next(&event) {
			out := struct {
				TxId       []byte          `json:"tx_id"`
				Collection string          `json:"collection"`
				Op         string          `json:"op"`
				Key        []byte          `json:"key,omitempty"`
				LKey       []byte          `json:"lkey,omitempty"`
				RKey       []byte          `json:"rkey,omitempty"`
				Data       json.RawMessage `json:"data,omitempty"`
				Last       bool            `json:"last"`
			}{
				TxId:       event.TxId,
				Collection: event.Collection,
				Op:         event.Op,
				Key:        event.Key,
				LKey:       event.Lkey,
				RKey:       event.Rkey,
				Data:       event.Data,
				Last:       event.Last,
			}

			json, err := jsoniter.MarshalIndent(out, "", "  ")
			if err != nil {
				util.Error(err, "stream events failed")
			}
			util.Stdout("%s\n", string(json))
		}
		if err := it.Err(); err != nil {
			util.Error(err, "iterate events failed")
		}
	},
}

func init() {
	dbCmd.AddCommand(streamCmd)
}
