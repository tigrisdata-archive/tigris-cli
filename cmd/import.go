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
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/iterate"
	"github.com/tigrisdata/tigris-cli/login"
	"github.com/tigrisdata/tigris-cli/schema"
	"github.com/tigrisdata/tigris-cli/util"
	api "github.com/tigrisdata/tigris-client-go/api/server/v1"
	errcode "github.com/tigrisdata/tigris-client-go/code"
	"github.com/tigrisdata/tigris-client-go/driver"
	cschema "github.com/tigrisdata/tigris-client-go/schema"
)

var (
	Append         bool
	NoCreate       bool
	InferenceDepth int32
	PrimaryKey     []string
	AutoGenerate   []string

	CleanUpNULLs = true

	CSVDelimiter        string
	CSVComment          string
	CSVTrimLeadingSpace bool

	CSVNoHeader bool

	sch cschema.Schema // Accumulate inferred schema across batches

	ErrCollectionShouldExist = fmt.Errorf("collection should exist to import CSV with no field names")
	ErrNoAppend              = fmt.Errorf(
		"collection exists. use --append if you need to add documents to existing collection")
)

func evolveSchema(ctx context.Context, db string, coll string, docs []json.RawMessage) error {
	// Allow to reduce inference depth in the case of huge batches
	id := len(docs)
	if InferenceDepth > 0 {
		id = int(InferenceDepth)
	}

	err := schema.Infer(&sch, coll, docs, PrimaryKey, AutoGenerate, id)
	util.Fatal(err, "infer schema")

	b, err := json.Marshal(sch)
	util.Fatal(err, "marshal schema: %s", string(b))

	err = client.Get().UseDatabase(db).CreateOrUpdateCollection(ctx, coll, b)

	return util.Error(err, "create or update collection")
}

func insertWithInference(ctx context.Context, coll string, docs []json.RawMessage) error {
	ptr := unsafe.Pointer(&docs)

	_, err := client.GetDB().Insert(ctx, coll, *(*[]driver.Document)(ptr))
	if err == nil {
		return nil // successfully inserted batch
	}

	//FIXME: errors.As(err, &ep) doesn't work
	//nolint:golint,errorlint
	ep, ok := err.(*driver.Error)
	if !ok || (ep.Code != api.Code_NOT_FOUND && ep.Code != errcode.InvalidArgument) ||
		ep.Code == api.Code_NOT_FOUND && NoCreate {
		return util.Error(err, "import documents (initial)")
	}

	if err := evolveSchema(ctx, config.GetProjectName(), coll, docs); err != nil {
		return err
	}

	// retry after schema update
	_, err = client.GetDB().Insert(ctx, coll, *(*[]driver.Document)(ptr))
	if err == nil {
		return nil
	}

	if CleanUpNULLs {
		for k := range docs {
			docs[k] = util.CleanupNULLValues(docs[k])
		}
	}

	_, err = client.GetDB().Insert(ctx, coll, *(*[]driver.Document)(ptr))

	log.Debug().Interface("docs", docs).Msg("import")

	return util.Error(err, "import documents (after schema update")
}

var importCmd = &cobra.Command{
	Use:   "import {collection} {document}...|-",
	Short: "Import documents into collection",
	Long: `Imports documents into the collection.
Input is a stream or array of JSON documents to import.

Automatically:
  * Detect the schema of the documents
  * Create collection with inferred schema
  * Evolve the schema as soon as it's backward compatible
`,
	Example: fmt.Sprintf(`
  %[1]s import --project=testdb users --primary-key=id \
  '[
    {"id": 20, "name": "Jania McGrory"},
    {"id": 21, "name": "Bunny Instone"}
  ]'
`, rootCmd.Root().Name()),
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.GetDB().DescribeCollection(ctx, args[0])
			if err == nil {
				if !Append {
					util.Fatal(ErrNoAppend, "describe collection")
				}
				err := json.Unmarshal(resp.Schema, &sch)
				util.Fatal(err, "unmarshal collection schema")
			} else if CSVNoHeader {
				util.Fatal(ErrCollectionShouldExist, "describe collection")
			}

			err = iterate.CSVConfigure(CSVDelimiter, CSVComment, CSVTrimLeadingSpace, CSVNoHeader)
			util.Fatal(err, "csv configure")

			return iterate.Input(cmd.Context(), cmd, 1, args,
				func(ctx context.Context, args []string, docs []json.RawMessage) error {
					return insertWithInference(ctx, args[0], docs)
				})
		})
	},
}

func init() {
	importCmd.Flags().Int32VarP(&iterate.BatchSize, "batch-size", "b", iterate.BatchSize, "set batch size")
	importCmd.Flags().BoolVarP(&Append, "append", "a", false,
		"Force append to existing collection")
	importCmd.Flags().BoolVar(&NoCreate, "no-create-collection", false,
		"Do not create collection automatically if it doesn't exist")
	importCmd.Flags().Int32VarP(&InferenceDepth, "inference-depth", "d", 0,
		"Number of records in the beginning of the stream to detect field types. It's equal to batch size if not set")
	importCmd.Flags().StringSliceVar(&PrimaryKey, "primary-key", []string{},
		"Comma separated list of field names which constitutes collection's primary key (only top level keys supported)")
	importCmd.Flags().StringSliceVar(&AutoGenerate, "autogenerate", []string{},
		"Comma separated list of autogenerated fields (only top level keys supported)")
	importCmd.Flags().BoolVar(&CleanUpNULLs, "cleanup-null-values", true,
		"Remove NULL values and empty arrays from the documents before importing")

	importCmd.Flags().StringVar(&CSVDelimiter, "csv-delimiter", "",
		"CSV delimiter")
	importCmd.Flags().BoolVar(&CSVTrimLeadingSpace, "csv-trim-leading-space", true,
		"Trim leading space in the fields")
	importCmd.Flags().StringVar(&CSVComment, "csv-comment", "",
		"CSV comment")

	importCmd.Flags().BoolVar(&schema.DetectByteArrays, "detect-byte-arrays", false,
		"Try detect byte arrays fields")
	importCmd.Flags().BoolVar(&schema.DetectUUIDs, "detect-uuids", true,
		"Try detect UUID fields")
	importCmd.Flags().BoolVar(&schema.DetectTimes, "detect-times", true,
		"Try detect date time fields")
	importCmd.Flags().BoolVar(&schema.DetectIntegers, "detect-integers", true,
		"Try detect integer fields")

	addProjectFlag(importCmd)
	rootCmd.AddCommand(importCmd)
}
