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
	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/util"
)

var rootCmd = &cobra.Command{
	Use:   "tigris",
	Short: "tigris is a command line interface of Tigris data platform",
}

var dbCmd = rootCmd

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&util.Quiet, "quiet", "q", false,
		"Suppress informational messages")
}

var errUnableToReadProject = fmt.Errorf("please specify project name")

func getProjectName() string {
	// first user supplied flag
	// second env variable
	// third config file
	if project == "" {
		project = config.DefaultConfig.Project
		if project == "" {
			util.Fatal(errUnableToReadProject, "unable to read project")
		}
	}

	return project
}

var project string

func addProjectFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&project, "project", "p", "", "Specifies project: --project=foo")
}
