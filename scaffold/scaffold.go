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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gertd/go-pluralize"
	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/iancoleman/strcase"
	"github.com/rs/zerolog/log"
	"github.com/tigrisdata/tigris-cli/util"
	api "github.com/tigrisdata/tigris-client-go/api/server/v1"
	"github.com/tigrisdata/tigris-client-go/schema"
)

var (
	ErrUnsupportedFormat    = fmt.Errorf("unsupported language. supported are: TypeScript, Go, Java")
	ErrTemplatesInvalidPath = fmt.Errorf("only local templates path substitution is allowed")

	templatesRepoURL = "https://github.com/tigrisdata/tigris-templates"
	examplesRepoURL  = "https://github.com/tigrisdata/tigris-examples-%s"

	plural = pluralize.NewClient()

	ErrUnknownFramewrok = fmt.Errorf("unknown framework")
	ErrDirAlreadyExists = fmt.Errorf("output directory already exists")
)

type Config struct {
	TemplatesPath   string
	OutputDirectory string
	PackageName     string
	ProjectName     string
	Example         string
	Framework       string
	Components      []string
	Collections     []*api.CollectionDescription
	Language        string
	URL             string
	ClientID        string
	ClientSecret    string
}

type TmplVars struct {
	URL                string
	Collections        []Collection
	Collection         Collection
	ProjectName        string
	ProjectNameCamel   string
	PackageName        string
	ClientID           string
	ClientSecret       string
	DatabaseBranchName string
}

type JSONToLangType interface {
	HasTime(string) bool
	HasUUID(string) bool
}

type Collection struct {
	Name       string // UserName
	NameDecap  string // userName
	JSON       string // user_names
	Schema     string
	JSONSchema string

	JSONSingular string // user_name

	NamePluralDecap string // userNames
	NamePlural      string // UserNames

	PrimaryKey []string

	HasTime bool
	HasUUID bool
}

func ensureLocalTemplates(base string, lang string, envVar string, repoURL string) string {
	templatePath := ensureTemplatesDir(base, lang)

	if os.Getenv(envVar) != "" {
		// Do not allow remote path substitution
		_, err := url.ParseRequestURI(os.Getenv(envVar))
		if err == nil && !os.IsPathSeparator(repoURL[0]) {
			util.Fatal(ErrTemplatesInvalidPath, "get examples path from env")
		}

		repoURL = os.Getenv(envVar)
		repoURL, err = filepath.Abs(repoURL)
		util.Fatal(err, "translate relative examples path to absolute: %v", repoURL)
	}

	// FIXME: come up with more reliable way of detecting URL vs local file path
	if _, err := url.ParseRequestURI(repoURL); err == nil && !os.IsPathSeparator(repoURL[0]) {
		CloneGitRepo(context.Background(), repoURL, templatePath)
	} else {
		templatePath = repoURL
		log.Debug().Str("path", templatePath).Msg("using local templates")
	}

	return templatePath
}

func EnsureLocalExamples(lang string) string {
	log.Debug().Msg("ensureLocalExamples")
	return ensureLocalTemplates("examples", lang, "TIGRIS_EXAMPLES_PATH", fmt.Sprintf(examplesRepoURL, lang))
}

func EnsureLocalTemplates() string {
	log.Debug().Msg("ensureLocalTemplates")
	return ensureLocalTemplates("templates", "", "TIGRIS_TEMPLATES_PATH", templatesRepoURL)
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

func decodeSchemas(inSchema []byte, lang string) (string, *schema.Schema, string) {
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

	return s, &js, schemas["json"]
}

func writeCollection(_ *TmplVars, w *bufio.Writer, collection *api.CollectionDescription,
	lang string, genType JSONToLangType,
) *Collection {
	s, js, jss := decodeSchemas(collection.Schema, lang)

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

		Schema:     s,
		JSONSchema: jss,

		PrimaryKey: js.PrimaryKey,

		HasUUID: genType.HasUUID(s),
		HasTime: genType.HasTime(s),
	}
}

func substCollectionFn(fn string, c *Collection) string {
	name := strings.ReplaceAll(fn, "_collection_name_", c.JSONSingular)
	name = strings.ReplaceAll(name, "_collection_names_", c.JSON)
	name = strings.ReplaceAll(name, "_Collection_name_", c.Name)
	name = strings.ReplaceAll(name, "_Collection_names_", c.NamePlural)

	return name
}

func walkDir(ffs fs.FS, rootPath string, outDir string, vars *TmplVars) error {
	return fs.WalkDir(ffs, ".", func(path string, d fs.DirEntry, err error) error {
		util.Fatal(err, "walk template directory. rootPath '%v'", rootPath)

		log.Debug().Str("path", path).Msg("walk-dir")

		if !d.IsDir() {
			tmpl, err := fs.ReadFile(ffs, path)
			// b, err := ffs.ReadFile(path) // embed.FS
			util.Fatal(err, "read template file: %s", path)

			path = strings.TrimSuffix(path, ".gotmpl")
			path = strings.TrimSuffix(path, ".gohtml")

			dir := filepath.Dir(path)
			dir = strings.TrimPrefix(dir, rootPath)
			fn := filepath.Base(path)

			dir = strings.ReplaceAll(dir, "_java_pkg_", strings.ReplaceAll(vars.PackageName, ".", string(filepath.Separator)))
			dir = strings.ReplaceAll(dir, "_db_name_", vars.ProjectName)

			dir = filepath.Join(outDir, dir)
			fn = strings.ReplaceAll(fn, "_db_name_", vars.ProjectNameCamel)

			if strings.Contains(strings.ToLower(dir+fn), "_collection_name_") ||
				strings.Contains(strings.ToLower(dir+fn), "_collection_names_") {
				log.Debug().Interface("colls", vars.Collections).Msg("per coll")

				for _, c := range vars.Collections {
					vars.Collection = c

					ofn := substCollectionFn(fn, &vars.Collection)
					odir := substCollectionFn(dir, &vars.Collection)
					oname := filepath.Join(odir, ofn)

					err = os.MkdirAll(odir, 0o755)
					util.Fatal(err, "MkdirAll: %s", odir)

					util.ExecFileTemplate(oname, string(tmpl), vars)
				}
			} else {
				err = os.MkdirAll(dir, 0o755)
				util.Fatal(err, "MkdirAll: %s", dir)
				util.ExecFileTemplate(filepath.Join(dir, fn), string(tmpl), vars)
			}
		}

		return nil
	})
}

func project(cfg *Config) {
	genType := getGenerator(cfg.Language)

	if cfg.PackageName == "" {
		cfg.PackageName = cfg.ProjectName
	}

	vars := TmplVars{
		URL:              cfg.URL,
		ProjectName:      cfg.ProjectName,
		ProjectNameCamel: strcase.ToCamel(cfg.ProjectName),
		PackageName:      cfg.PackageName,
		ClientID:         cfg.ClientID,
		ClientSecret:     cfg.ClientSecret,
	}

	colls := make([]Collection, 0, len(cfg.Collections))

	for _, v := range cfg.Collections {
		c := writeCollection(&vars, nil, v, cfg.Language, genType)
		colls = append(colls, *c)
	}

	vars.Collections = colls

	if cfg.Example != "" {
		rootPath := filepath.Join(cfg.TemplatesPath, cfg.Example)

		ffs := os.DirFS(rootPath)

		err := walkDir(ffs, rootPath, cfg.OutputDirectory, &vars)
		util.Fatal(err, "processing examples")

		return
	}

	l := cfg.Language // overwrite ts -> typescript in path
	if cfg.Language == "ts" {
		l = "typescript"
	}

	rootPath := filepath.Join(cfg.TemplatesPath, "source", l, cfg.Framework)

	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		util.Infof("Available frameworks for language '%s:", l)

		list := util.ListDir(filepath.Join(cfg.TemplatesPath, "source", l))

		for k := range list {
			if k != "base" {
				util.Infof("\t* %s", k)
			}
		}

		util.Infof("")

		util.Fatal(fmt.Errorf("%w: %s", ErrUnknownFramewrok, cfg.Framework), "frameworks")
	}

	err := execComponents(rootPath, cfg.OutputDirectory, cfg.Components, &vars)
	util.Fatal(err, "processed components")
}

func execComponents(rootPath string, outDir string, components []string, vars *TmplVars) error {
	list := util.ListDir(rootPath)

	keys := make([]string, 0, len(list))
	for k := range list {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, v := range keys {
		l := strings.SplitN(v, "_", 2)

		if len(l) > 1 {
			l[0] = l[1]
		}

		// If no explicitly specified component, include all by default.
		include := len(components) == 0

		// not_default_ prefix allows to exclude component from default processing
		if strings.HasPrefix(l[0], "not_default_") {
			l[0] = strings.ReplaceAll(l[0], "not_default_", "")
			include = false
		}

		for _, c := range components {
			if c == l[0] {
				include = true
			}
		}

		log.Debug().Str("component", l[0]).Bool("included", include).Msg("processing component")

		if include {
			componentPath := filepath.Join(rootPath, v)
			ffs := os.DirFS(componentPath)

			if err := walkDir(ffs, componentPath, outDir, vars); err != nil {
				return fmt.Errorf("%w: walk templates dir %s", err, componentPath)
			}
		}
	}

	return nil
}

func Project(cfg *Config) {
	cfg.OutputDirectory = filepath.Join(cfg.OutputDirectory, cfg.ProjectName)

	log.Debug().Str("outDir", cfg.OutputDirectory).Msg("check output directory exists")

	if _, err := os.Stat(cfg.OutputDirectory); err == nil {
		util.Fatal(fmt.Errorf("%w: %s", ErrDirAlreadyExists, cfg.OutputDirectory), "output directory already exists")
	}

	project(cfg)

	InitGit(cfg.OutputDirectory)
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

func ensureTemplatesDir(base string, lang string) string {
	c, err := os.UserCacheDir()
	util.Fatal(err, "get cache directory")

	path := filepath.Join(c, "tigris", base)
	if lang != "" {
		path = filepath.Join(c, "tigris", fmt.Sprintf("%s-%s", base, lang))
	}

	err = os.MkdirAll(path, 0o755)
	util.Fatal(err, "mkdir cache directory: %s", path)

	return path
}

func cloneGitRepo(ctx context.Context, url string, path string) error {
	remoteBranch := "beta"

	log.Debug().Str("url", url).Str("path", path).Msg("Cloning git repo")

	_, err := git.PlainCloneContext(ctx, path, false, &git.CloneOptions{
		URL:           url,
		ReferenceName: plumbing.NewBranchReferenceName(remoteBranch),
		Depth:         1,
	})

	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		_ = util.Error(err, "updating existing repo url %v, dir %v", url, path)

		var repo *git.Repository

		if repo, err = git.PlainOpen(path); err != nil {
			return util.Error(err, "OpenRepo url %v, dir %v", url, path)
		}

		// Fetched updated remote branch
		err = repo.FetchContext(ctx, &git.FetchOptions{
			Force: true, Depth: 1,
			RefSpecs: []gitconfig.RefSpec{
				gitconfig.RefSpec("+refs/heads/" + remoteBranch + ":refs/remotes/origin/" + remoteBranch),
			},
		})
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			log.Debug().Msg("templates repository is up-to-date")

			err = nil
		}

		// Checking out just fetched remote branch
		tr, err := repo.Worktree()
		if err != nil {
			return util.Error(err, "getting git worktree")
		}

		err = tr.Checkout(&git.CheckoutOptions{Branch: plumbing.NewRemoteReferenceName("origin", remoteBranch)})
		if err != nil {
			return util.Error(err, "checking out fetched branch")
		}
	}

	return util.Error(err, "CloneGitRepo url %v, dir %v", url, path)
}

func CloneGitRepo(ctx context.Context, url string, path string) {
	if err := cloneGitRepo(ctx, url, path); err == nil {
		return
	}

	err := os.RemoveAll(path)
	_ = util.Error(err, "remove existing repo: %s", path)

	err = cloneGitRepo(ctx, url, path)
	util.Fatal(err, "retrying git repo clone")
}
