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
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/tigrisdata/tigris-cli/util"
)

var docsCmd = &cobra.Command{
	Use:   "docs {output directory}",
	Short: "Generates CLI documentation in Markdown format",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := doc.GenMarkdownTreeCustom(rootCmd, args[0], func(name string) string {
			name = path.Base(name)
			name = strings.TrimSuffix(name, path.Ext(name))
			name = strings.TrimPrefix(name, "tigris_")
			name = strings.ReplaceAll(name, "_", "-")
			title := strings.ReplaceAll(name, "-", " ")
			title = strings.Title(title) //nolint:staticcheck

			return fmt.Sprintf(`---
id: %s
title: %s
slug: /sdkstools/cli/%s
---
`, name, title, name)
		}, func(link string) string {
			return link
		})
		util.Fatal(err, "generating Markdown documentation")
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
}
