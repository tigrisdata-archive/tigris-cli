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
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/util"
	"github.com/tigrisdata/tigris-client-go/driver"
)

var (
	dbFilter         []string
	collectionFilter []string
	destDir          string
	backupTimeout    int
)

func listDatabases(ctx context.Context) ([]string, error) {
	databases := []string{}

	resp, err := client.Get().ListDatabases(ctx)
	if err != nil {
		return nil, util.Error(err, "list databases")
	}

	for _, v := range resp {
		if len(dbFilter) == 0 || util.Contains(dbFilter, v) {
			databases = append(databases, v)
		}
	}

	return databases, nil
}

func listCollections(ctx context.Context, dbName string) ([]string, error) {
	collections := []string{}

	resp, err := client.Get().UseDatabase(dbName).ListCollections(ctx)
	if err != nil {
		return nil, util.Error(err, "list collections")
	}

	for _, v := range resp {
		if len(collectionFilter) == 0 || util.Contains(collectionFilter, v) {
			collections = append(collections, v)
		}
	}

	return collections, nil
}

func makeDir(dirname string) {
	if err := os.Mkdir(dirname, 0o700); err != nil {
		util.Fatal(err, "error creating directory %s", dirname)
	}
}

var backupCmd = &cobra.Command{
	Use:   "backup [filters]",
	Short: "Dumps documents and schemas to JSON files",
	Long: `Dumps documents and schemas to JSON files
	in the current working directory.

	If a database filter is provided it only dumps the schemas of the databases specified.
	Likewise, collection filters will limit the output to matching collection names.`,
	Args: cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(_ context.Context) error {
			util.Stdoutf(" [o] using timeout %d\n", backupTimeout)

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(backupTimeout)*time.Second)
			defer cancel()

			databases, err := listDatabases(ctx)
			if err != nil {
				return util.Error(err, "list databases")
			}

			for _, db := range databases {
				collections, err := listCollections(ctx, db)
				if err != nil {
					return util.Error(err, "error listing collections")
				}

				// Dump schema
				resp, err := client.Get().DescribeDatabase(ctx, db)
				if err != nil {
					return util.Error(err, "describe collection failed")
				}

				makeDir(destDir)
				filename := fmt.Sprintf("%s/%s.schema", destDir, db)
				for _, v := range resp.Collections {
					err := os.WriteFile(filename, []byte(string(v.Schema)), 0o600)
					util.Fatal(err, "error writing schema file")
				}
				util.Stdoutf(" [.] %s\n", filename)

				for _, collection := range collections {
					filter, fields := `{}`, `{}`
					it, err := client.Get().UseDatabase(db).Read(ctx, collection,
						driver.Filter(filter),
						driver.Projection(fields),
						nil,
					)
					if err != nil {
						return util.Error(err, "read failed")
					}
					defer it.Close()

					// Dump data
					filename = fmt.Sprintf("%s/%s.%s.json", destDir, db, collection)
					var doc driver.Document
					for it.Next(&doc) {
						err := os.WriteFile(filename, []byte(string(doc)+"\n"), 0o600)
						util.Fatal(err, "error writing file %s", filename)
					}

					util.Stdoutf(" [*] %s\n", filename)
				}
			}

			return nil
		})
	},
}

func init() {
	backupCmd.Flags().StringVarP(&destDir, "directory", "d", "./",
		"destination directory for backups")
	backupCmd.Flags().StringSliceVarP(&dbFilter, "databases", "D", []string{},
		"limit data dump to specified databases")
	backupCmd.Flags().StringSliceVarP(&collectionFilter, "collections", "C", []string{},
		"limit data dump to specified collections")
	backupCmd.Flags().IntVarP(&backupTimeout, "timeout", "t", 3600,
		"timeout specification in seconds")
	rootCmd.AddCommand(backupCmd)
}
