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

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/schema"
	"github.com/tigrisdata/tigris-cli/util"
	api "github.com/tigrisdata/tigris-client-go/api/server/v1"
	errcode "github.com/tigrisdata/tigris-client-go/code"
	"github.com/tigrisdata/tigris-client-go/driver"
	cschema "github.com/tigrisdata/tigris-client-go/schema"
)

var (
	AutoCreate     bool
	InferenceDepth int32
	PrimaryKey     []string
	AutoGenerate   []string

	CleanUpNULLs = true

	sch cschema.Schema // Accumulate inferred schema across batches
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

func cleanupNULLValuesLow(m map[string]any) {
	for k, v := range m {
		switch val := v.(type) {
		case map[string]any:
			cleanupNULLValuesLow(val)
		case []any:
			if len(val) == 0 {
				log.Debug().Str("name", k).Msg("removed empty array")
				delete(m, k)
			}
		case nil:
			log.Debug().Str("name", k).Msg("removed empty value")
			delete(m, k)
		}
	}
}

// cleanupNULLValues cleans up NULL values and empty arrays from the document.
func cleanupNULLValues(doc json.RawMessage) json.RawMessage {
	var m map[string]any

	err := jsoniter.Unmarshal([]byte(doc), &m)
	util.Fatal(err, "unmarshal doc for cleanup")

	cleanupNULLValuesLow(m)

	doc, err = jsoniter.Marshal(&m)
	util.Fatal(err, "marshal doc after cleanup")

	return doc
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
  %[1]s import --project=testdb users --create-collection --primary-key=id \
  '[
    {"id": 20, "name": "Jania McGrory"},
    {"id": 21, "name": "Bunny Instone"}
  ]'
`, rootCmd.Root().Name()),
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withLogin(cmd.Context(), func(ctx context.Context) error {
			resp, err := client.Get().UseDatabase(getProjectName()).DescribeCollection(ctx, args[0])
			if err == nil {
				err := json.Unmarshal(resp.Schema, &sch)
				util.Fatal(err, "unmarshal collection schema")
			}

			return iterateInput(cmd.Context(), cmd, 1, args,
				func(ctx context.Context, args []string, docs []json.RawMessage) error {
					ptr := unsafe.Pointer(&docs)

					_, err := client.Get().UseDatabase(getProjectName()).Insert(ctx, args[0], *(*[]driver.Document)(ptr))
					if err == nil {
						return nil // successfully inserted batch
					}

					//FIXME: errors.As(err, &ep) doesn't work
					//nolint:golint,errorlint
					ep, ok := err.(*driver.Error)
					if !ok || (ep.Code != api.Code_NOT_FOUND && ep.Code != errcode.InvalidArgument) ||
						ep.Code == api.Code_NOT_FOUND && !AutoCreate {
						return util.Error(err, "import documents (initial)")
					}

					if err := evolveSchema(ctx, getProjectName(), args[0], docs); err != nil {
						return err
					}

					// retry after schema update
					_, err = client.Get().UseDatabase(getProjectName()).Insert(ctx, args[0], *(*[]driver.Document)(ptr))
					if err == nil {
						return nil
					}

					if CleanUpNULLs {
						for k := range docs {
							docs[k] = cleanupNULLValues(docs[k])
						}
					}

					_, err = client.Get().UseDatabase(getProjectName()).Insert(ctx, args[0], *(*[]driver.Document)(ptr))

					log.Debug().Interface("docs", docs).Msg("import")

					return util.Error(err, "import documents (after schema update")
				})
		})
	},
}

func init() {
	importCmd.Flags().Int32VarP(&BatchSize, "batch-size", "b", BatchSize, "set batch size")
	importCmd.Flags().BoolVarP(&AutoCreate, "create-collection", "c", false,
		"Automatically create collection if it doesn't exist")
	importCmd.Flags().Int32VarP(&InferenceDepth, "inference-depth", "d", 0,
		"Number of records in the beginning of the stream to detect field types. It's equal to batch size if not set")
	importCmd.Flags().StringSliceVar(&PrimaryKey, "primary-key", []string{},
		"Comma separated list of field names which constitutes collection's primary key (only top level keys supported)")
	importCmd.Flags().StringSliceVarP(&AutoGenerate, "autogenerate", "a", []string{},
		"Comma separated list of autogenerated fields (only top level keys supported)")
	importCmd.Flags().BoolVar(&CleanUpNULLs, "cleanup-null-values", true,
		"Remove NULL values and empty arrays from the documents before importing")
	addProjectFlag(importCmd)
	dbCmd.AddCommand(importCmd)
}
