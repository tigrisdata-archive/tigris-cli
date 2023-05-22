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

package iterate

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/tigrisdata/tigris-cli/util"
)

var (
	CSVDelimiter        rune
	CSVTrimLeadingSpace bool
	CSVComment          rune
	CSVNoHeader         bool

	ErrDelimiterTooLong = fmt.Errorf("delimiter should be one character")
	ErrCommentTooLong   = fmt.Errorf("comment should be one character")
)

func CSVConfigure(delimiter string, comment string, trimLeadingSpace bool, noHeader bool) error {
	if delimiter != "" {
		if len(delimiter) > 1 {
			return ErrDelimiterTooLong
		}

		CSVDelimiter = rune(delimiter[0])
	}

	if comment != "" {
		if len(comment) > 1 {
			return ErrCommentTooLong
		}

		CSVComment = rune(comment[0])
	}

	CSVTrimLeadingSpace = trimLeadingSpace
	CSVNoHeader = noHeader

	return nil
}

func findKey(d map[string]any, names [][]string, key int) map[string]any {
	for i := 0; i < len(names[key])-1; i++ {
		if d[names[key][i]] == nil {
			d[names[key][i]] = make(map[string]any)
		}

		d, _ = d[names[key][i]].(map[string]any)
	}

	return d
}

func readCSVBatch(reader *csv.Reader, names [][]string, batchSize int) []json.RawMessage {
	var docs []json.RawMessage

	for i := 0; i < batchSize; i++ {
		row, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return docs
		}

		util.Fatal(err, "read csv row")

		fields := make(map[string]any)

		for k, v := range row {
			d := findKey(fields, names, k)

			var f float64

			f, err = strconv.ParseFloat(v, 64)

			switch {
			case err == nil:
				d[names[k][len(names[k])-1]] = f
			case strings.TrimSpace(v) == "null":
				d[names[k][len(names[k])-1]] = nil
			case strings.TrimSpace(v) == "true":
				d[names[k][len(names[k])-1]] = true
			case strings.TrimSpace(v) == "false":
				d[names[k][len(names[k])-1]] = false
			default:
				d[names[k][len(names[k])-1]] = v
			}
		}

		b, err := json.Marshal(fields)
		util.Fatal(err, "marshal")

		docs = append(docs, b)
	}

	return docs
}

func iterateCSVStream(ctx context.Context, args []string, r io.Reader, fn func(ctx2 context.Context, args []string,
	docs []json.RawMessage) error,
) error {
	csvReader := csv.NewReader(r)

	if CSVComment != rune(0) {
		csvReader.Comment = CSVComment
		csvReader.Comment = ';'
	}

	if CSVDelimiter != rune(0) {
		csvReader.Comma = CSVDelimiter
	}

	csvReader.TrimLeadingSpace = CSVTrimLeadingSpace

	headers, err := csvReader.Read()
	util.Fatal(err, "read CSV headers: %+v", headers)

	numFields := len(headers)
	names := make([][]string, numFields)

	for k, v := range headers {
		names[k] = strings.Split(v, ".")
	}

	var bar *progressbar.ProgressBar

	for {
		docs := readCSVBatch(csvReader, names, int(BatchSize))

		if util.IsTTY(os.Stdout) {
			bar = progressbar.Default(-1)
		}

		if len(docs) == 0 {
			break
		} else if err := varyBatch(ctx, args, docs, fn); err != nil {
			return err
		}

		if util.IsTTY(os.Stdout) {
			_ = bar.Add(len(docs))
		}
	}

	return nil
}
