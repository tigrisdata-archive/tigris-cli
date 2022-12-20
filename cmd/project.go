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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/scaffold"
	"github.com/tigrisdata/tigris-cli/util"
	api "github.com/tigrisdata/tigris-client-go/api/server/v1"
	"github.com/tigrisdata/tigris-client-go/driver"
)

var (
	schemaOnly bool
	format     string
)

var listProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Lists projects",
	Long:  "This command will list all projects.",
	Run: func(cmd *cobra.Command, _ []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.Get().ListProjects(ctx)
			if err != nil {
				return util.Error(err, "list projects")
			}

			for _, v := range resp {
				util.Stdoutf("%s\n", v)
			}

			return nil
		})
	},
}

// DescribeDatabaseResponse adapter to convert schema to json.RawMessage.
type DescribeDatabaseResponse struct {
	DB          string                        `json:"db,omitempty"`
	Metadata    *api.DatabaseMetadata         `json:"metadata,omitempty"`
	Collections []*DescribeCollectionResponse `json:"collections,omitempty"`
}

var describeDatabaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Describes database",
	Long:  "Returns schema and metadata for all the collections in the database",
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.Get().DescribeDatabase(ctx, getProjectName(),
				&driver.DescribeProjectOptions{SchemaFormat: format})
			if err != nil {
				return util.Error(err, "describe collection failed")
			}

			if schemaOnly {
				for _, v := range resp.Collections {
					util.Stdoutf("%s\n", string(v.Schema))
				}
			} else {
				tr := DescribeDatabaseResponse{
					Metadata: resp.Metadata,
				}

				for _, v := range resp.Collections {
					tr.Collections = append(tr.Collections, &DescribeCollectionResponse{
						Collection: v.Collection,
						// Metadata:   v.Metadata,
						Schema: v.Schema,
					})
				}

				b, err := json.Marshal(tr)
				util.Fatal(err, "describe database")

				util.Stdoutf("%s\n", string(b))
			}

			return nil
		})
	},
}

var createProjectCmd = &cobra.Command{
	Use:   "project {name}",
	Short: "Creates project",
	Args:  cobra.MinimumNArgs(1),
	Long:  "This command will create a project. Optionally allows to bootstrap database and application code",
	Example: fmt.Sprintf(`
	# Create Tigris project with no collections
	%[1]s %[2]s 

	# Create project and bootstrap collections from the template
	%[1]s %[2]s --template todo

	# Create project and scaffold application code using TypeSript and Express framework'
	%[1]s %[2]s --framework=express

	# Both bootstrap collections and scaffold Express application
	%[1]s %[2]s --template todo --framework=express
`, rootCmd.Root().Name(), "create project"),
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			if template != "" {
				if err := scaffold.Schema(ctx, args[0], template, true, false); err != nil {
					return util.Error(err, "bootstrapping schema: %s", template)
				}
			} else {
				_, err := client.Get().CreateProject(ctx, args[0])
				if err != nil {
					return util.Error(err, "create project")
				}
			}

			if framework != "" {
				return scaffoldProject(ctx, args[0])
			}

			return nil
		})
	},
}

func init() {
	describeDatabaseCmd.Flags().BoolVarP(&schemaOnly, "schema-only", "s", false,
		"dump only schema of all database collections")

	describeDatabaseCmd.Flags().StringVarP(&format, "format", "f", "",
		"output schema in the requested format: go, typescript, java")

	addScaffoldProjectFlags(createProjectCmd)

	createCmd.AddCommand(createProjectCmd)
	listCmd.AddCommand(listProjectsCmd)
	describeCmd.AddCommand(describeDatabaseCmd)
}
