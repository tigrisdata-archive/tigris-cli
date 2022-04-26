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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/util"
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Check connection to Tigris",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()
		_, err := client.Get().ListDatabases(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAILED\n")
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stderr, "OK\n")
		}
	},
}

func init() {
	dbCmd.AddCommand(pingCmd)
}
