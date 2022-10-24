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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/util"
	api "github.com/tigrisdata/tigris-client-go/api/server/v1"
)

var (
	ErrUnsupportedFormat = fmt.Errorf("unsupported format. supported formats are: JSON, TypeScripts, Go, Java")

	plural = pluralize.NewClient()
)

type tmplVars struct {
	URL         string
	Collections []Collection
	HasTime     bool
	HasUUID     bool
	DBName      string
	DBNameCamel string
}

type JSONToLangType interface {
	GetHeaderTemplate() string
	GetFooterTemplate() string
	HasTime(string) bool
	HasUUID(string) bool
}

type Collection struct {
	Name      string
	NameDecap string
	JSON      string
}

type Field struct {
	Type      string
	TypeDecap string

	Name       string
	NameDecap  string
	NameSnake  string
	NameJSON   string
	NamePlural string

	IsArray  bool
	IsObject bool

	AutoGenerate    bool
	PrimaryKeyIdx   int
	ArrayDimensions int

	Description string
}

type Object struct {
	Name      string
	NameDecap string
	NameSnake string
	NameJSON  string

	NamePlural string

	Description string

	Nested bool

	Fields []Field
}

func getGenerator(lang string) JSONToLangType {
	var genType JSONToLangType

	switch l := strings.ToLower(lang); l {
	case "go", "golang":
		genType = &JSONToGo{}
	case "ts", "typescript":
		genType = &JSONToTypeScript{}
	case "java":
		genType = &JSONToJava{}
	default:
		util.Fatal(ErrUnsupportedFormat, "")
	}

	return genType
}

func writeCollections(vars *tmplVars, w *bufio.Writer, collections []*api.CollectionDescription,
	lang string, genType JSONToLangType,
) {
	names := make([]Collection, 0, len(collections))

	for _, v := range collections {
		schemas := make(map[string]string)

		err := json.Unmarshal(v.Schema, &schemas)
		util.Fatal(err, "unmarshal schema")

		if s, ok := schemas[lang]; !ok {
			util.Fatal(ErrUnsupportedFormat, "schema not found for %v", lang)
		} else {
			if ht := genType.HasTime(s); ht {
				vars.HasTime = ht
			}

			if hu := genType.HasUUID(s); hu {
				vars.HasUUID = hu
			}

			_, err = w.Write([]byte(s))
			util.Fatal(err, "write to writer")
		}

		name := strcase.ToCamel(plural.Singular(v.Collection))
		nameDecap := strings.ToLower(name[0:1]) + name[1:]

		names = append(names, Collection{JSON: v.Collection, Name: name, NameDecap: nameDecap})
	}

	util.Fatal(w.Flush(), "flushing schema buffer")

	vars.Collections = names
}

func scaffoldFromDB(db string, collections []*api.CollectionDescription, lang string) []byte {
	genType := getGenerator(lang)

	bufObj := bytes.Buffer{}
	w := bufio.NewWriter(&bufObj)

	vars := tmplVars{
		URL:         config.DefaultConfig.URL,
		DBName:      db,
		DBNameCamel: strcase.ToCamel(db),
	}

	writeCollections(&vars, w, collections, lang, genType)

	buf := bytes.Buffer{}
	w = bufio.NewWriter(&buf)

	util.ExecTemplate(w, genType.GetHeaderTemplate(), vars)

	_, err := w.Write(bufObj.Bytes())
	util.Fatal(err, "writing objects buffer to writer")

	util.ExecTemplate(w, genType.GetFooterTemplate(), vars)

	util.Fatal(w.Flush(), "error flushing schema buffer")

	return buf.Bytes()
}

func ScaffoldFromDB(db string, collections []*api.CollectionDescription, lang string) error {
	buf := scaffoldFromDB(db, collections, lang)

	_, err := os.Stdout.Write(buf)
	util.Fatal(err, "error writing schema to stdout")

	return nil
}
