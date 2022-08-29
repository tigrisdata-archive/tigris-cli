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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/util"
)

var BatchSize int32 = 100

func detectArray(r *bufio.Reader) bool {
	var err error
	var c rune

	for {
		c, _, err = r.ReadRune()
		if err != nil {
			if err == io.EOF {
				return false
			}
			util.Error(err, "error reading input")
			return false
		}
		if !unicode.IsSpace(c) {
			break
		}
	}

	err = r.UnreadRune()
	if err != nil {
		util.Error(err, "error reading input")
	}

	return c == '['
}

func iterateArray(r []byte) []json.RawMessage {
	arr := make([]json.RawMessage, 0)
	err := json.Unmarshal(r, &arr)
	if err != nil {
		util.Error(err, "reading parsing array of documents")
	}
	return arr
}

func iterateStream(ctx context.Context, args []string, r io.Reader, fn func(ctx2 context.Context, args []string, docs []json.RawMessage)) {
	dec := json.NewDecoder(r)
	for {
		docs := make([]json.RawMessage, 0, BatchSize)
		var i int32
		for ; i < BatchSize && dec.More(); i++ {
			var v json.RawMessage
			err := dec.Decode(&v)
			if err != nil {
				util.Error(err, "reading documents from stream of documents")
			}
			docs = append(docs, v)
		}
		if i > 0 {
			fn(ctx, args, docs)
		} else {
			break
		}
	}
}

// iterateInput reads repeated command parameters from standard input or args
// Support newline delimited stream of objects and arrays of objects
func iterateInput(ctx context.Context, cmd *cobra.Command, docsPosition int, args []string, fn func(ctx2 context.Context, args []string, docs []json.RawMessage)) {
	if len(args) > docsPosition && args[docsPosition] != "-" {
		docs := make([]json.RawMessage, 0, len(args))
		for _, v := range args[docsPosition:] {
			if detectArray(bufio.NewReader(bytes.NewReader([]byte(v)))) {
				docs = append(docs, iterateArray([]byte(v))...)
			} else {
				docs = append(docs, json.RawMessage(v))
			}
		}
		fn(ctx, args, docs)
	} else if len(args) <= docsPosition && util.IsTTY(os.Stdin) {
		_, _ = fmt.Fprintf(os.Stderr, "not enougn arguments\n")
		_ = cmd.Usage()
		os.Exit(1)
	} else {
		r := bufio.NewReader(os.Stdin)
		if detectArray(r) {
			buf, err := io.ReadAll(r)
			if err != nil {
				util.Error(err, "error reading documents")
			}
			docs := iterateArray(buf)
			fn(ctx, args, docs)
		} else {
			iterateStream(ctx, args, r, fn)
		}
	}
}
