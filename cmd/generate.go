// Copyright 2022-2023 Tigris Data, Inc.
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
	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/login"
	"github.com/tigrisdata/tigris-cli/scaffold"
	"github.com/tigrisdata/tigris-cli/util"
)

const sampleDBName = "sampledb"

var (
	create bool
	stdout bool
)

var sampleSchemaCmd = &cobra.Command{
	Use:   "sample-schema",
	Short: "Generates sample schema",
	Long:  "Generates sample schema consisting of three collections: products, users, orders.",
	Example: fmt.Sprintf(`
  # Generate sample schema files in current directory orders.json, products.json and users.json
  %[1]s generate sample-schema

  # Create the database sampledb and sample collections
  %[1]s generate sample-schema --create

  # Generate sample schema and output it to stdout 
  %[1]s generate sample-schema --stdout
`, rootCmd.Root().Name()),
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			templatesPath := scaffold.EnsureLocalTemplates()

			if create {
				if _, err := client.Get().CreateProject(ctx, config.GetProjectName()); err != nil {
					return util.Error(err, "create database")
				}
			}

			err := scaffold.Schema(cmd.Context(), sampleDBName, templatesPath, "ecommerce", create, stdout)

			return util.Error(err, "generating sample schema")
		})
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generating helper assets such as sample schema",
}

func init() {
	sampleSchemaCmd.Flags().BoolVarP(&create, "create", "c", false, "create the sample database and collections")
	sampleSchemaCmd.Flags().BoolVarP(&stdout, "stdout", "s", false, "dump sample schemas to stdout")
	addProjectFlag(sampleSchemaCmd)

	generateCmd.AddCommand(sampleSchemaCmd)
	rootCmd.AddCommand(generateCmd)
}
