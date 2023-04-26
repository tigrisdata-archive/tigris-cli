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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/login"
	"github.com/tigrisdata/tigris-cli/scaffold"
	"github.com/tigrisdata/tigris-cli/templates"
	"github.com/tigrisdata/tigris-cli/util"
	api "github.com/tigrisdata/tigris-client-go/api/server/v1"
	"github.com/tigrisdata/tigris-client-go/driver"
)

var (
	schemaOnly    bool
	format        string
	createEnvFile bool
)

var listProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Lists projects",
	Long:  "This command will list all projects.",
	Run: func(cmd *cobra.Command, _ []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
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
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.Get().DescribeDatabase(ctx, config.GetProjectName(),
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

func writeEnvFile(ctx context.Context, proj string) {
	clientID, clientSecret, err := getAppKeyForTemplate(ctx, proj)
	util.Fatal(err, "get app key")

	buf := bytes.Buffer{}
	util.ExecTemplate(&buf, templates.DotEnv, &scaffold.TmplVars{
		ClientID:           clientID,
		ClientSecret:       clientSecret,
		URL:                config.DefaultConfig.URL,
		ProjectName:        proj,
		DatabaseBranchName: "main",
	})

	f, err := os.Create(".env")
	util.Fatal(err, "create .env file")

	_, err = f.Write(buf.Bytes())
	util.Fatal(err, "write .env file")

	err = f.Close()
	util.Fatal(err, "close .env file")

	util.Infof("Written .env file")
}

var createProjectCmd = &cobra.Command{
	Use:   "project {name}",
	Short: "Creates project",
	Args:  cobra.MinimumNArgs(1),
	Long:  "This command will create a project. Optionally allows to bootstrap database and application code",
	Example: fmt.Sprintf(`
	# Create Tigris project with no collections
	%[1]s %[2]s 

	# Create project and bootstrap collections from the schemaTemplate
	%[1]s %[2]s --schema-template todo

	# Create project and scaffold application code using TypeSript and Express framework'
	%[1]s %[2]s --framework=express

	# Both bootstrap collections and scaffold Express application
	%[1]s %[2]s --schema-template todo --framework=express

	# Create project and create application code from the examples repository
	%[1]s %[2]s --schema-template todo --framework=express
`, rootCmd.Root().Name(), "create project"),
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			_, err := client.Get().CreateProject(ctx, args[0])
			if err != nil {
				return util.Error(err, "create project")
			}

			config.DefaultConfig.Project = args[0]

			if fromExample != "" || schemaTemplate != "" || framework != "" {
				return scaffoldProject(ctx)
			} else if createEnvFile {
				writeEnvFile(ctx, args[0])
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

	createProjectCmd.Flags().BoolVar(&createEnvFile, "create-env-file", false,
		"Create .env file with Tigris connection credentials")

	addScaffoldProjectFlags(createProjectCmd)

	createCmd.AddCommand(createProjectCmd)
	listCmd.AddCommand(listProjectsCmd)
	describeCmd.AddCommand(describeDatabaseCmd)
}
