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

package schema

import (
	"strings"

	"github.com/tigrisdata/tigris-cli/templates"
)

type JSONToJava struct{}

func (*JSONToJava) GetHeaderTemplate() string {
	return templates.SchemaJavaHeader
}

func (*JSONToJava) GetFooterTemplate() string {
	return templates.SchemaJavaFooter
}

func (*JSONToJava) HasTime(schema string) bool {
	return strings.Contains(schema, "private Date")
}

func (*JSONToJava) HasUUID(schema string) bool {
	return strings.Contains(schema, "private UUID")
}
