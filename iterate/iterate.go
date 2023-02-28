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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"unicode"

	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/util"
)

var (
	ErrNotAllDocsProcessed = fmt.Errorf("not all documents processed")

	BatchSize int32 = 100
)

func detectArray(r io.RuneScanner) bool {
	var c rune

	for {
		var err error

		c, _, err = r.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return false
			}

			util.Fatal(err, "reading input")

			return false
		}

		if !unicode.IsSpace(c) {
			break
		}
	}

	if err := r.UnreadRune(); err != nil {
		util.Fatal(err, "reading input")
	}

	return c == '['
}

func readArray(r []byte) []json.RawMessage {
	arr := make([]json.RawMessage, 0)
	if err := json.Unmarshal(r, &arr); err != nil {
		util.Fatal(err, "reading parsing array of documents")
	}

	return arr
}

func iterateStream(ctx context.Context, args []string, r io.Reader, fn func(ctx2 context.Context, args []string,
	docs []json.RawMessage) error,
) error {
	var bar *progressbar.ProgressBar

	if util.IsTTY(os.Stdout) {
		bar = progressbar.Default(-1)
	}

	dec := json.NewDecoder(r)

	for {
		docs := make([]json.RawMessage, 0, BatchSize)

		var i int32

		for ; i < BatchSize && dec.More(); i++ {
			var v json.RawMessage

			err := dec.Decode(&v)
			util.Fatal(err, "reading documents from stream of documents")

			docs = append(docs, v)
		}

		if i == 0 {
			break
		} else if err := varyBatch(ctx, args, docs, fn); err != nil {
			return err
		}

		if util.IsTTY(os.Stdout) {
			_ = bar.Add(int(i))
		}
	}

	return nil
}

// varyBatch dynamically reduces the batch on document-exceeded-limit error and retries.
func varyBatch(ctx context.Context, args []string, docs []json.RawMessage,
	process func(ctx2 context.Context, args []string, docs []json.RawMessage) error,
) error {
	first := 0
	last := len(docs)
	total := 0

	for first < len(docs) {
		if err := process(ctx, args, docs[first:last]); err != nil {
			if (err.Error() == "document exceeds limit" || err.Error() == "transaction exceeds limit") &&
				last-first > 1 {
				last = first + (last-first)/2 // exponentially reduce the batch size

				log.Debug().Msgf("reducing batch size. first=%d, last=%d, len=%d", first, last, len(docs))

				continue
			} else if last-first == 1 {
				log.Debug().RawJSON("doc", docs[first]).Msgf("failed to process")
			}

			return err
		}

		log.Debug().Msgf("succeeded batch. first=%d, last=%d, len=%d", first, last, len(docs))

		sz := last - first // retain and reuse the batch-size which succeeded
		total += sz

		first = last

		last = first + sz
		if last > len(docs) {
			last = len(docs)
		}
	}

	if total != len(docs) {
		util.InternalError(ErrNotAllDocsProcessed, "processed: %d, expected %d", total, len(docs))
	}

	return nil
}

func iterateArray(ctx context.Context, args []string, r io.Reader, fn func(ctx2 context.Context, args []string,
	docs []json.RawMessage) error,
) error {
	buf, err := io.ReadAll(r)
	util.Fatal(err, "error reading documents")

	allDocs := readArray(buf)

	var bar *progressbar.ProgressBar

	if util.IsTTY(os.Stdout) {
		bar = progressbar.Default(int64(len(allDocs)))
	}

	for j := 0; j < len(allDocs); {
		docs := make([]json.RawMessage, 0, BatchSize)

		var i int32
		for ; i < BatchSize && j < len(allDocs); i++ {
			docs = append(docs, allDocs[j])
			j++
		}

		if err = varyBatch(ctx, args, docs, fn); err != nil {
			return err
		}

		if util.IsTTY(os.Stdout) {
			_ = bar.Add(int(i))
		}
	}

	return nil
}

// Input reads repeated command parameters from standard input or args.
// Supports newline delimited stream of objects and arrays of objects.
func Input(ctx context.Context, cmd *cobra.Command, docsPosition int, args []string,
	fn func(ctx2 context.Context, args []string, docs []json.RawMessage) error,
) error {
	if len(args) > docsPosition && args[docsPosition] != "-" {
		docs := make([]json.RawMessage, 0, len(args))

		for _, v := range args[docsPosition:] {
			if detectArray(bufio.NewReader(bytes.NewReader([]byte(v)))) {
				docs = append(docs, readArray([]byte(v))...)
			} else {
				docs = append(docs, json.RawMessage(v))
			}
		}

		return fn(ctx, args, docs)
	} else if len(args) <= docsPosition && util.IsTTY(os.Stdin) {
		_, _ = fmt.Fprintf(os.Stderr, "not enougn arguments\n")
		_ = cmd.Usage()
		os.Exit(1)
	}

	// stdin not a TTY or "-" is specified
	r := bufio.NewReader(os.Stdin)
	if detectArray(r) {
		return iterateArray(ctx, args, r, fn)
	}

	return iterateStream(ctx, args, r, fn)
}
