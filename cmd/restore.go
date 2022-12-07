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
	"os"
	"regexp"
	"strings"
	"time"
	"unsafe"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/util"
	"github.com/tigrisdata/tigris-client-go/driver"
)

var (
	srcDir         string
	dbPrefix       string
	dbPostfix      string
	dbSeparator    string
	restoreTimeout int
)

func amendDatabaseName(dbname string) string {
	var tags []string

	if dbPrefix != "" {
		tags = append(tags, dbPrefix)
	}

	tags = append(tags, dbname)

	if dbPostfix != "" {
		tags = append(tags, dbPostfix)
	}

	return strings.Join(tags, dbSeparator)
}

func findDatabases() ([]string, error) {
	databases := make([]string, 0)

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, util.Error(err, "error reading directory")
	}

	// See if db names need to be filtered
	dbPattern := `.*`
	if len(projectFilter) > 0 {
		dbPattern = strings.Join(projectFilter, "|")
	}

	// Obtain database names from schema files
	schemaPattern := fmt.Sprintf(
		`^(?P<name>(%s))[.]%s$`,
		dbPattern,
		schemaFileExtension)
	re := regexp.MustCompile(schemaPattern)

	for _, v := range entries {
		if v.IsDir() {
			continue
		}

		// Extract database name from filename
		match := re.FindStringSubmatch(v.Name())
		if len(match) == 0 {
			continue
		}

		databases = append(databases, match[1])
	}

	return databases, nil
}

func findCollections(db string) ([]string, error) {
	collections := make([]string, 0)

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, util.Error(err, "error reading directory")
	}

	// See if collections need to be filtered
	collPattern := `.*`
	if len(collectionFilter) > 0 {
		collPattern = strings.Join(collectionFilter, "|")
	}

	// Obtain schema names from json files
	backupPattern := fmt.Sprintf(
		`^%s[.](?P<collection>(%s))[.]%s$`,
		db,
		collPattern,
		backupFileExtension)
	re := regexp.MustCompile(backupPattern)

	for _, v := range entries {
		if v.IsDir() {
			continue
		}

		// Extract database name from filename
		match := re.FindStringSubmatch(v.Name())
		if len(match) == 0 {
			continue
		}

		collections = append(collections, match[1])
	}

	return collections, nil
}

func createDatabase(ctx context.Context, db string) error {
	err := client.Get().CreateDatabase(ctx, db)
	return util.Error(err, "create database")
}

func restoreDatabase(ctx context.Context, db, path string) error {
	docs := make([]json.RawMessage, 0)

	f, err := os.Open(path)
	if err != nil {
		return util.Error(err, "failed to read schema")
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	for dec.More() {
		var v json.RawMessage
		err := dec.Decode(&v)
		util.Fatal(err, "reading documents from stream of documents")

		docs = append(docs, v)
	}

	restoreDB := amendDatabaseName(db)
	if err := createDatabase(ctx, restoreDB); err != nil {
		return util.Error(err, "create database failed")
	}

	return client.Transact(ctx, restoreDB, func(ctx context.Context, tx driver.Tx) error {
		for _, v := range docs {
			if err := createCollection(ctx, tx, driver.Schema(v)); err != nil {
				return util.Error(err, "failed to create collection")
			}
		}

		return nil
	})
}

func restoreCollection(ctx context.Context, db, collection, path string) error {
	restoreDB := amendDatabaseName(db)

	f, err := os.Open(path)
	if err != nil {
		return util.Error(err, "failed to read collection")
	}
	defer f.Close()

	dec := json.NewDecoder(f)

	for {
		docs := make([]json.RawMessage, 0)
		i := int32(0)

		for ; i < BatchSize && dec.More(); i++ {
			var v json.RawMessage

			err := dec.Decode(&v)
			util.Fatal(err, "reading documents from stream of documents")

			docs = append(docs, v)
		}

		if i > 0 {
			ptr := unsafe.Pointer(&docs)

			_, err := client.Get().UseDatabase(restoreDB).Insert(ctx, collection, *(*[]driver.Document)(ptr))
			if err != nil {
				return util.Error(err, "insert document")
			}
		} else {
			break
		}
	}

	return nil
}

var restoreCmd = &cobra.Command{
	Use:   "restore [filters]",
	Short: "restores documents and schemas from JSON files",
	Long: `restores documents and schemas from JSON files in a directory specified in the argument.

	If a database filter is provided it only restores the schemas of the databases specified.
	Likewise, collection filters will limit the input to matching collection names.

	SYNOPSIS

	Restore all the data and schema with the same name as they were captured:
	$ tigris restore -d /path/to/backup

	Restore a single project called mydb:
	$ tigris restore -d /path/to/backup -P mydb

	Restore mydb project under the name mydb_restored:
	$ tigris restore -d /path/to/backup -P mydb -e restored

	Restore mydb as restored.mydb:
	$ tigris restore -d /path/to/backup -P mydb -b restored -s '.'

	`,
	Args: cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(_ context.Context) error {
			util.Stdoutf(" [i] using timeout %d\n", restoreTimeout)
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(restoreTimeout)*time.Second)
			defer cancel()

			databases, err := findDatabases()
			if err != nil {
				return util.Error(err, "failed to find databases")
			}

			for _, db := range databases {
				util.Stdoutf(" [*] %s\n", db)
				file := fmt.Sprintf("%s/%s.%s", srcDir, db, schemaFileExtension)
				if err = restoreDatabase(ctx, db, file); err != nil {
					return util.Error(err, "failed to restore database")
				}

				collections, err := findCollections(db)
				if err != nil {
					return util.Error(err, "failed to find collections")
				}

				for _, collection := range collections {
					util.Stdoutf(" [.] %s => %s\n", db, collection)
					file := fmt.Sprintf("%s/%s.%s.%s", srcDir, db, collection, backupFileExtension)
					if err = restoreCollection(ctx, db, collection, file); err != nil {
						return util.Error(err, "failed to restore collection")
					}
				}
			}

			return nil
		})
	},
}

func init() {
	restoreCmd.Flags().StringVarP(&srcDir, "directory", "d", "./tigris-backup",
		"input file directory")
	restoreCmd.Flags().StringVarP(&dbPrefix, "prefix", "b", "",
		"prefixes restored database names")
	restoreCmd.Flags().StringVarP(&dbPostfix, "postfix", "e", "",
		"postfixes restored database names")
	restoreCmd.Flags().StringVarP(&dbSeparator, "separator", "s", "_",
		"separator to use for pre and postfixes")
	restoreCmd.Flags().StringSliceVarP(&projectFilter, "projects", "P", []string{},
		"limit data restore to specified projects")
	restoreCmd.Flags().StringSliceVarP(&collectionFilter, "collections", "C", []string{},
		"limit data restore to specified collections")
	restoreCmd.Flags().IntVarP(&restoreTimeout, "timeout", "t", 3600,
		"timeout specification in seconds")
	rootCmd.AddCommand(restoreCmd)
}
