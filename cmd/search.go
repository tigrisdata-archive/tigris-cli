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

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/login"
	"github.com/tigrisdata/tigris-cli/util"
	"github.com/tigrisdata/tigris-client-go/driver"
	"github.com/tigrisdata/tigris-client-go/search"
)

var (
	query         string
	searchFields  []string
	filter        string
	facet         string
	sort          []string
	includeFields []string
	excludeFields []string
	page          int32
	pageSize      int32
)

var dbSearchCmd = &cobra.Command{
	Use:   "search {collection}",
	Short: "Searches a collection for documents matching the query",
	Long:  "Executes a search query against collection and returns the search results.",
	//nolint:golint,lll
	Example: fmt.Sprintf(`
# Default search without any parameters will return all documents
%[1]s %[2]s

# Search for a text "Alice" in collection
%[1]s %[2]s -q "Alice"

# Search for a text "Alice" either in "firstName" or "lastName" fields
%[1]s %[2]s -q "Alice" -f "firstName,lastName"

# Filter for users with age > 23
%[1]s %[2]s -q "Alice" -f "firstName,lastName" --filter '{"age": {"$gt": 23}}'

# Aggregate results by current city and get top 10 cities
%[1]s %[2]s -q "Alice" -f "firstName,lastName" --filter '{"age": {"$gt": 23}}' --facet '{"currentCity": {"size": 10}}'

# Sort the results by age in increasing order
%[1]s %[2]s -q "Alice" -f "firstName,lastName" --filter '{"age": {"$gt": 23}}' --facet '{"currentCity": {"size": 10}}' --sort '{"age": "$asc"}'

# Exclude sensitive information from results
%[1]s %[2]s -q "Alice" -f "firstName,lastName" --filter '{"age": {"$gt": 23}}' --facet '{"currentCity": {"size": 10}}' --sort '{"age": "$asc"}' -x "phoneNumber,address"

# Paginate the results, with 15 per page
%[1]s %[2]s -q "Alice" -f "firstName,lastName" --filter '{"age": {"$gt": 23}}' --facet '{"currentCity": {"size": 10}}' --sort '{"age": "$asc"}' -x "phoneNumber,address" -p 1 -c 15

# Find users with last name exactly matching "Wong"
%[1]s %[2]s --filter '{"lastName": "Wong"}'
`, rootCmd.Root().Name(), "search --project=testdb users"),
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		login.Ensure(cmd.Context(), func(ctx context.Context) error {
			var sortArr driver.SortOrder
			if len(sort) > 0 {
				for _, v := range sort {
					sortArr = append(sortArr, json.RawMessage(v))
				}
			}
			request := &driver.SearchRequest{
				Q:             query,
				SearchFields:  searchFields,
				Filter:        driver.Filter(filter),
				Facet:         driver.Facet(facet),
				Sort:          sortArr,
				IncludeFields: includeFields,
				ExcludeFields: excludeFields,
				Page:          page,
				PageSize:      pageSize,
			}

			it, err := client.GetDB().Search(ctx, args[0], request)
			if err != nil {
				return util.Error(err, "search failed")
			}

			var resp driver.SearchResponse
			for it.Next(&resp) {
				r := &search.Result[interface{}]{}
				err := r.From(resp)
				util.Fatal(err, "search result conversion")

				resultJSON, err := json.MarshalIndent(r, "", " ")
				util.Fatal(err, "result marshalling")

				util.Stdoutf("%s\n", resultJSON)
			}

			return util.Error(it.Err(), "search result iteration")
		})
	},
}

func init() {
	dbSearchCmd.Flags().SortFlags = false

	dbSearchCmd.Flags().StringVarP(&query, "query", "q", "", "query string for searching across text fields")
	dbSearchCmd.Flags().StringSliceVarP(&searchFields, "searchFields", "f", []string{},
		"comma separated value of fields to project search query against")
	dbSearchCmd.Flags().StringVar(&filter, "filter", "{}", "further refine the search results using filters")
	dbSearchCmd.Flags().StringVar(&facet, "facet", "{}", "retrieve aggregate ")
	dbSearchCmd.Flags().StringSliceVar(&sort, "sort", nil, "order to sort the results")
	dbSearchCmd.Flags().StringSliceVarP(&includeFields, "includeFields", "i", []string{},
		"comma separated value of document fields to include in results")
	dbSearchCmd.Flags().StringSliceVarP(&excludeFields, "excludeFields", "x", []string{},
		"comma separated value of document fields to exclude in results")
	dbSearchCmd.Flags().Int32VarP(&page, "page", "g", 1, "page of results to retrieve")
	dbSearchCmd.Flags().Int32VarP(&pageSize, "pageSize", "c", 20, "count of results to be returned per page")

	addProjectFlag(dbSearchCmd)
	dbCmd.AddCommand(dbSearchCmd)
}
