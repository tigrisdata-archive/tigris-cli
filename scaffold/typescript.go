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

package scaffold

import (
	"embed"
	"strings"

	"github.com/pkg/errors"
	"github.com/tigrisdata/tigris-cli/templates"
	"github.com/tigrisdata/tigris-cli/util"
)

type JSONToTypeScript struct{}

func (*JSONToTypeScript) HasTime(schema string) bool {
	return strings.Contains(schema, "DATE_TIME")
}

func (*JSONToTypeScript) HasUUID(schema string) bool {
	return strings.Contains(schema, "UUID")
}

func (*JSONToTypeScript) GetFS(dir string) embed.FS {
	switch dir {
	case "base":
		return templates.ScaffoldTypeScriptBase
	case "express", "":
		return templates.ScaffoldTypeScriptExpress
	default:
		util.Fatal(errors.Wrapf(ErrUnknownScaffoldType, "type: %s", dir), "get fs")
	}

	return embed.FS{} // this should never happen
}
