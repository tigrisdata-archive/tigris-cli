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

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/schema"
	"github.com/tigrisdata/tigris-cli/util"
	"github.com/tigrisdata/tigris-client-go/driver"
)

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Scaffold a project for specified language",
}

var goCmd = &cobra.Command{
	Use:     "go",
	Aliases: []string{"golang"},
	Short:   "Scaffold a new Go project from Tigris project",
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.Get().DescribeDatabase(ctx, getProjectName(),
				&driver.DescribeDatabaseOptions{SchemaFormat: "go"})
			if err != nil {
				return util.Error(err, "describe collection failed")
			}

			err = schema.ScaffoldFromDB(getProjectName(), resp.Collections, "go")
			util.Fatal(err, "scaffold from database")

			return nil
		})
	},
}

var typeScriptCmd = &cobra.Command{
	Use:     "typescript",
	Aliases: []string{"ts"},
	Short:   "Scaffold a new TypeScript project from Tigris project",
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.Get().DescribeDatabase(ctx, getProjectName(),
				&driver.DescribeDatabaseOptions{SchemaFormat: "typescript"})
			if err != nil {
				return util.Error(err, "describe collection failed")
			}

			err = schema.ScaffoldFromDB(getProjectName(), resp.Collections, "typescript")
			util.Fatal(err, "scaffold from database")

			return nil
		})
	},
}

var javaCmd = &cobra.Command{
	Use:   "java",
	Short: "Scaffold a new Java project from Tigris project",
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.Get().DescribeDatabase(ctx, getProjectName(),
				&driver.DescribeDatabaseOptions{SchemaFormat: "java"})
			if err != nil {
				return util.Error(err, "describe collection failed")
			}

			err = schema.ScaffoldFromDB(getProjectName(), resp.Collections, "java")
			util.Fatal(err, "scaffold from database")

			return nil
		})
	},
}

func init() {
	addProjectFlag(scaffoldCmd)
	scaffoldCmd.AddCommand(goCmd)
	scaffoldCmd.AddCommand(typeScriptCmd)
	scaffoldCmd.AddCommand(javaCmd)

	rootCmd.AddCommand(scaffoldCmd)
}
