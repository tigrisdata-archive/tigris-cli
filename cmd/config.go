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
	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/util"
	"gopkg.in/yaml.v2"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration commands",
}

var showConfigCmd = &cobra.Command{
	Use:   "show",
	Short: "Returns effective CLI configuration",
	Run: func(cmd *cobra.Command, args []string) {
		info, err := yaml.Marshal(config.DefaultConfig)
		util.Fatal(err, "marshal config")

		if string(info) == "{}\n" {
			return
		}

		util.Stdoutf("%s\n", string(info))
	},
}

func init() {
	configCmd.AddCommand(showConfigCmd)
	rootCmd.AddCommand(configCmd)
}
