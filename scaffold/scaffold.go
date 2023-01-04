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

package scaffold

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gertd/go-pluralize"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/iancoleman/strcase"
	"github.com/rs/zerolog/log"
	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/util"
	api "github.com/tigrisdata/tigris-client-go/api/server/v1"
	"github.com/tigrisdata/tigris-client-go/schema"
)

var (
	ErrUnsupportedFormat = fmt.Errorf("unsupported format. supported formats are: JSON, TypeScript, Go, Java")

	plural = pluralize.NewClient()
)

var templateDeps = map[string][]string{
	"gin":     {"base"},
	"express": {"base"},
}

type tmplVars struct {
	URL          string
	Collections  []Collection
	Collection   Collection
	DBName       string
	DBNameCamel  string
	PackageName  string
	ClientID     string
	ClientSecret string
}

type JSONToLangType interface {
	HasTime(string) bool
	HasUUID(string) bool
	GetFS(dir string) embed.FS
}

type Collection struct {
	Name      string // UserName
	NameDecap string // userName
	JSON      string // user_names
	Schema    string

	JSONSingular string // user_name

	NamePluralDecap string // userNames
	NamePlural      string // UserNames

	PrimaryKey []string

	HasTime bool
	HasUUID bool
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

func getSchemas(inSchema []byte, lang string) (string, *schema.Schema) {
	schemas := make(map[string]string)

	err := json.Unmarshal(inSchema, &schemas)
	util.Fatal(err, "unmarshal schema")

	s, ok := schemas[lang]
	if !ok {
		util.Fatal(ErrUnsupportedFormat, "schema not found for %v", lang)
	}

	if schemas["json"] == "" {
		util.Fatal(ErrUnsupportedFormat, "json schema not found")
	}

	var js schema.Schema

	err = json.Unmarshal([]byte(schemas["json"]), &js)
	util.Fatal(err, "unmarshalling json schema")

	return s, &js
}

func writeCollection(_ *tmplVars, w *bufio.Writer, collection *api.CollectionDescription,
	lang string, genType JSONToLangType,
) *Collection {
	s, js := getSchemas(collection.Schema, lang)

	if js.CollectionType == "messages" {
		return nil
	}

	if w != nil {
		_, err := w.Write([]byte(s))
		util.Fatal(err, "write to writer")
	}

	name := strcase.ToCamel(plural.Singular(collection.Collection))
	namePlural := plural.Plural(name)

	return &Collection{
		JSON:         collection.Collection,
		JSONSingular: plural.Singular(collection.Collection),

		Name:      name,
		NameDecap: strings.ToLower(name[0:1]) + name[1:],

		NamePlural:      namePlural,
		NamePluralDecap: strings.ToLower(namePlural[0:1]) + namePlural[1:],

		Schema: s,

		PrimaryKey: js.PrimaryKey,

		HasUUID: genType.HasUUID(s),
		HasTime: genType.HasTime(s),
	}
}

func ExecFileTemplate(fn string, tmpl string, vars any) {
	buf := bytes.Buffer{}
	w := bufio.NewWriter(&buf)

	util.ExecTemplate(w, tmpl, vars)

	err := w.Flush()
	util.Fatal(err, "flushing template writer")

	err = os.WriteFile(fn, buf.Bytes(), 0o600)
	util.Fatal(err, "write file")
}

func walkDir(ffs embed.FS, rootPath string, pkgName string, outDir string, vars *tmplVars,
	collections []*api.CollectionDescription, lang string, genType JSONToLangType,
) error {
	return fs.WalkDir(ffs, rootPath, func(path string, d fs.DirEntry, err error) error {
		util.Fatal(err, "walk embedded template directory")

		log.Debug().Str("path", path).Msg("walk-dir")

		if !d.IsDir() {
			b, err := ffs.ReadFile(path)
			util.Fatal(err, "read template file")

			path = strings.TrimSuffix(path, ".gotmpl")
			path = strings.TrimSuffix(path, ".gohtml")

			dir := filepath.Dir(path)
			dir = strings.TrimPrefix(dir, rootPath)
			fn := filepath.Base(path)

			dir = strings.ReplaceAll(dir, "_java_pkg_", strings.ReplaceAll(pkgName, ".", string(filepath.Separator)))

			err = os.MkdirAll(filepath.Join(outDir, dir), 0o755)
			util.Fatal(err, "MkdirAll: %s", dir)

			fn = strings.ReplaceAll(fn, "_db_name_", vars.DBNameCamel)

			if strings.Contains(strings.ToLower(fn), "_collection_name_") ||
				strings.Contains(strings.ToLower(fn), "_collection_names_") {
				for _, v := range collections {
					c := writeCollection(vars, nil, v, lang, genType)
					if c == nil {
						continue
					}

					vars.Collection = *c

					name := strings.ReplaceAll(fn, "_collection_name_", c.JSONSingular)
					name = strings.ReplaceAll(name, "_collection_names_", c.JSON)
					name = strings.ReplaceAll(name, "_Collection_name_", c.Name)
					name = strings.ReplaceAll(name, "_Collection_names_", c.NamePlural)
					ExecFileTemplate(filepath.Join(outDir, dir, name), string(b), vars)
				}
			} else {
				ExecFileTemplate(filepath.Join(outDir, dir, fn), string(b), vars)
			}
		}

		return nil
	})
}

func project(outDir string, pkgName string, db string, flavor string, collections []*api.CollectionDescription,
	lang string,
	clientID string,
	clientSecret string,
) {
	genType := getGenerator(lang)

	if pkgName == "" {
		pkgName = db
	}

	vars := tmplVars{
		URL:          config.DefaultConfig.URL,
		DBName:       db,
		DBNameCamel:  strcase.ToCamel(db),
		PackageName:  pkgName,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	ffs := genType.GetFS(flavor)

	colls := make([]Collection, 0, len(collections))

	for _, v := range collections {
		c := writeCollection(&vars, nil, v, lang, genType)
		if c != nil {
			colls = append(colls, *c)
		}
	}

	vars.Collections = colls

	rootPath := filepath.Join("scaffold", lang, flavor)

	err := walkDir(ffs, rootPath, pkgName, outDir, &vars, collections, lang, genType)
	util.Fatal(err, "walk templates dir %s", flavor)
}

func Project(outDir string, pkgName string, dbName string, flavor string, colls []*api.CollectionDescription,
	lang string,
	clientID string,
	clientSecret string,
) {
	outDir = filepath.Join(outDir, dbName)

	for _, v := range templateDeps[flavor] {
		project(outDir, pkgName, dbName, v, colls, lang, clientID, clientSecret)
	}

	project(outDir, pkgName, dbName, flavor, colls, lang, clientID, clientSecret)

	InitGit(outDir)
}

func InitGit(outDir string) {
	log.Debug().Msgf("Initializing git repository %s", outDir)

	repo, err := git.PlainInit(outDir, false)
	util.Fatal(err, "init git repo: %s", outDir)

	w, err := repo.Worktree()
	util.Fatal(err, "get git worktree")

	// FIXME: w.Add() should respect .gitignore
	// https://github.com/go-git/go-git/issues/597
	w.Excludes = append(w.Excludes, gitignore.ParsePattern(".env*", nil))

	err = w.AddWithOptions(&git.AddOptions{All: true})
	util.Fatal(err, "git add")

	_, err = w.Commit("Initial", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "TigrisData, Inc.",
			Email: "firsov@tigrisdata.com",
			When:  time.Now(),
		},
	})
	util.Fatal(err, "git commit initial")
}
