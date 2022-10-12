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
)

var quotaLimitsCmd = &cobra.Command{
	Use:   "limits",
	Short: "Show quota limits for the namespace user logged in to",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()

		l, err := client.ObservabilityGet().QuotaLimits(ctx)
		if err != nil {
			util.Error(err, "quota limits failed")
		}

		if err := util.PrettyJSON(l); err != nil {
			util.Error(err, "quota limits failed")
		}
	},
}

var quotaUsageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Show current quota usage for the namespace user logged in to",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()

		u, err := client.ObservabilityGet().QuotaUsage(ctx)
		if err != nil {
			util.Error(err, "quota usage failed")
		}

		if err := util.PrettyJSON(u); err != nil {
			util.Error(err, "quota usage failed")
		}
	},
}

var quotaCmd = &cobra.Command{
	Use:   "quota",
	Short: "Quota related commands",
}

func init() {
	quotaCmd.AddCommand(quotaLimitsCmd)
	quotaCmd.AddCommand(quotaUsageCmd)
	rootCmd.AddCommand(quotaCmd)
}
