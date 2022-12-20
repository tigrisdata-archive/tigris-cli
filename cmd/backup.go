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
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	units "github.com/docker/go-units"
	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/util"
	"github.com/tigrisdata/tigris-client-go/driver"
)

var (
	projectFilter       []string
	collectionFilter    []string
	destDir             string
	backupTimeout       int
	verboseBackup       bool
	backupFileExtension = "backup"
	schemaFileExtension = "schema"
)

// listProjects returns the projects/databases available in Tigris as a string array
// but filters the output using the filters specified via the command line.
func listProjects(ctx context.Context) ([]string, error) {
	projects := make([]string, 0)

	resp, err := client.Get().ListProjects(ctx)
	if err != nil {
		return nil, util.Error(err, "list projects")
	}

	for _, v := range resp {
		if len(projectFilter) == 0 || util.Contains(projectFilter, v) {
			projects = append(projects, v)
		}
	}

	return projects, nil
}

// listCollections enumerates the collections available in Tigris as a string array
// but filters the output using the filters specified via the command line.
func listCollections(ctx context.Context, dbName string) ([]string, error) {
	collections := make([]string, 0)

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

// writeCollection downloads the data of a collection from Tigris and stores it
// in a file post-fixed with backupFileExtension. The function returns the bytes
// written and an error, if applicable.
func writeCollection(ctx context.Context, db, collection, file string) (int, error) {
	var (
		doc   driver.Document
		bytes int
	)

	it, err := client.Get().UseDatabase(db).Read(ctx, collection,
		driver.Filter(`{}`),
		driver.Projection(`{}`),
	)
	if err != nil {
		return bytes, util.Error(err, "read failed")
	}
	defer it.Close()

	f, err := os.Create(file)
	if err != nil {
		return bytes, util.Error(err, "failed writing file")
	}

	writer := bufio.NewWriter(f)
	for it.Next(&doc) {
		b, err := writer.WriteString(string(doc) + "\n")
		bytes += b

		util.Fatal(err, "error writing file %s", file)
	}

	if err = writer.Flush(); err != nil {
		return bytes, util.Error(err, "failed to flush file")
	}

	if err = f.Close(); err != nil {
		return bytes, util.Error(err, "failed to close file")
	}

	return bytes, nil
}

// writeSchema downloads the schema of all collections related to the database
// specified and stores in the file specified.
func writeSchema(ctx context.Context, db, file string) (int, error) {
	var bytes int

	resp, err := client.Get().DescribeDatabase(ctx, db)
	if err != nil {
		return bytes, util.Error(err, "describe collection failed")
	}

	f, err := os.Create(file)
	if err != nil {
		return bytes, util.Error(err, "failed writing file")
	}

	writer := bufio.NewWriter(f)
	for _, v := range resp.Collections {
		b, err := writer.WriteString(string(v.Schema) + "\n")
		bytes += b

		if err != nil {
			return bytes, util.Error(err, "error writing schema file")
		}
	}

	if err = writer.Flush(); err != nil {
		return bytes, util.Error(err, "failed to flush file")
	}

	if err = f.Close(); err != nil {
		return bytes, util.Error(err, "failed to close file")
	}

	return bytes, nil
}

func printStats(start time.Time, bytes float64) {
	if verboseBackup {
		stop := time.Since(start).Seconds()
		util.Stdoutf(" [>] %s (%s/s in %.1fs)\n", units.HumanSize(bytes), units.HumanSize(bytes/stop), stop)
	}
}

var backupCmd = &cobra.Command{
	Use:   "backup [filters]",
	Short: "Dumps documents and schemas to JSON files",
	Long: `Dumps documents and schemas to JSON files into the directory specified in the argument.

	If a project name filter is provided it only dumps the schemas of the projects specified.
	Likewise, collection filters will limit the output to matching collection names.`,
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(_ context.Context) error {
			util.Stdoutf(" [i] using timeout %d\n", backupTimeout)
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(backupTimeout)*time.Second)
			defer cancel()

			projects, err := listProjects(ctx)
			if err != nil {
				return util.Error(err, "failed to list projects")
			}

			if err := os.Mkdir(destDir, 0o700); err != nil {
				return util.Error(err, "failed to create backup dir")
			}

			for _, db := range projects {
				path := fmt.Sprintf("%s/%s.%s", destDir, db, schemaFileExtension)
				util.Stdoutf(" [.] %s\n", path)
				start := time.Now()
				bytes, err := writeSchema(ctx, db, path)
				if err != nil {
					return util.Error(err, "failed to write schema")
				}
				printStats(start, float64(bytes))

				collections, err := listCollections(ctx, db)
				if err != nil {
					return util.Error(err, "error listing collections")
				}
				for _, collection := range collections {
					start := time.Now()
					path := fmt.Sprintf("%s/%s.%s.%s", destDir, db, collection, backupFileExtension)
					util.Stdoutf(" [*] %s\n", path)
					bytes, err := writeCollection(ctx, db, collection, path)
					if err != nil {
						return util.Error(err, "failed to write collection")
					}
					printStats(start, float64(bytes))
				}
			}

			return nil
		})
	},
}

func init() {
	backupCmd.Flags().StringVarP(&destDir, "directory", "d", "./tigris-backup",
		"destination directory for backups")
	backupCmd.Flags().StringSliceVarP(&projectFilter, "projects", "P", []string{},
		"limit data dump to specified projects")
	backupCmd.Flags().StringSliceVarP(&collectionFilter, "collections", "C", []string{},
		"limit data dump to specified collections")
	backupCmd.Flags().IntVarP(&backupTimeout, "timeout", "t", 3600,
		"timeout specification in seconds")
	backupCmd.Flags().BoolVarP(&verboseBackup, "verbose", "v", false,
		"verbose output")
	rootCmd.AddCommand(backupCmd)
}
