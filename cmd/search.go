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
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/util"
	"github.com/tigrisdata/tigris-client-go/driver"
	"github.com/tigrisdata/tigris-client-go/search"
)

var query string
var searchFields []string
var filter string
var facet string
var sort string
var includeFields []string
var excludeFields []string
var page int32
var pageSize int32

var searchCmd = &cobra.Command{
	Use:   "search {db} {collection}",
	Short: "Searches a collection for documents matching the query",
	Long:  "Executes a search query against collection and returns the search results.",
	Example: fmt.Sprintf(`
# Default search without any parameters will return all documents
%[1]s %[2]s

# Search for a text "Alice" in collection
%[1]s %[2]s -q "Alice"

# Search for a text "Alice" either in "firstName" or "lastName" fields
%[1]s %[2]s -q "Alice" -f "firstName,lastName"

# Filter for users with age > 23
%[1]s %[2]s -q "Alice" -f "firstName,lastName" --filter "{\"age\": {\"$gt\": 23}}"

# Aggregate results by current city and get top 10 cities
%[1]s %[2]s -q "Alice" -f "firstName,lastName" --filter "{\"age\": {\"$gt\": 23}}" --facet "{\"currentCity\": {\"size\": 10}}"

# Exclude sensitive information from results
%[1]s %[2]s -q "Alice" -f "firstName,lastName" --filter "{\"age\": {\"$gt\": 23}}" --facet "{\"currentCity\": {\"size\": 10}}" -x "phoneNumber,address"

# Paginate the results, with 15 per page
%[1]s %[2]s -q "Alice" -f "firstName,lastName" --filter "{\"age\": {\"$gt\": 23}}" --facet "{\"currentCity\": {\"size\": 10}}" -x "phoneNumber,address" -p 1 -c 15

# Find users with last name exactly matching "Wong"
%[1]s %[2]s --filter "{\"lastName\": \"Wong\"}"
`, rootCmd.Root().Name(), "search testdb users"),
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()

		request := &driver.SearchRequest{
			Q:             query,
			SearchFields:  searchFields,
			Filter:        driver.Filter(filter),
			Facet:         driver.Facet(facet),
			IncludeFields: includeFields,
			ExcludeFields: excludeFields,
			Page:          page,
			PageSize:      pageSize,
		}

		it, err := client.Get().UseDatabase(args[0]).Search(ctx, args[1], request)
		if err != nil {
			util.Error(err, "search failed")
		}
		var resp driver.SearchResponse
		for it.Next(&resp) {
			r := &search.Result[interface{}]{}
			err := r.From(resp)
			if err != nil {
				util.Error(err, "search result conversion failed")
			}
			resultJson, _ := json.MarshalIndent(r, "", " ")
			util.Stdout("%s\n", resultJson)
		}
		if err := it.Err(); err != nil {
			util.Error(err, "search result iteration failed")
		}
	},
}

func init() {
	searchCmd.Flags().StringVarP(&query, "query", "q", "", "query string for searching across text fields")
	searchCmd.Flags().StringSliceVarP(&searchFields, "searchFields", "f", []string{}, "comma separated value of fields to project search query against")
	searchCmd.Flags().StringVar(&filter, "filter", "{}", "further refine the search results using filters")
	searchCmd.Flags().StringVar(&facet, "facet", "{}", "retrieve aggregate ")
	searchCmd.Flags().StringVar(&sort, "sort", "{}", "order to sort the results")
	searchCmd.Flags().StringSliceVarP(&includeFields, "includeFields", "i", []string{}, "comma separated value of document fields to include in results")
	searchCmd.Flags().StringSliceVarP(&excludeFields, "excludeFields", "x", []string{}, "comma separated value of document fields to exclude in results")
	searchCmd.Flags().Int32VarP(&page, "page", "p", 1, "page of results to retrieve")
	searchCmd.Flags().Int32VarP(&pageSize, "pageSize", "c", 20, "count of results to be returned per page")
	searchCmd.Flags().SortFlags = false
	dbCmd.AddCommand(searchCmd)
}
