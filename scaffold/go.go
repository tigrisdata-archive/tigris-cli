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
	"fmt"
	"html/template"
	"os"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/templates"
	"github.com/tigrisdata/tigris-cli/util"
)

var ErrDirAlreadyExists = fmt.Errorf("directory already exists")

type TemplateVars struct {
	Company     string
	Project     string
	DB          string
	Collections []string
}

var GoCmd = &cobra.Command{
	Use:     "go {Company} {Project} {DB} {Collection}...",
	Aliases: []string{"golang"},
	Short:   "Scaffold a new Go project",
	Args:    cobra.MinimumNArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		tv := &TemplateVars{
			Company:     args[0],
			Project:     args[1],
			DB:          args[2],
			Collections: args[3:],
		}

		if _, err := os.Stat(tv.Project); !os.IsNotExist(err) {
			util.Fatal(fmt.Errorf("%w: %s", ErrDirAlreadyExists, tv.Project),
				"directory of being scaffolded project already exists")
		}

		err := os.Mkdir(tv.Project, 0o755)
		util.Fatal(err, "mkdir "+tv.Project)

		f, err := os.Create(tv.Project + "/main.go")
		util.Fatal(err, "create "+tv.Project+"/main.go")

		defer func() {
			err := f.Close()
			_ = util.Error(err, "close main.go")
		}()

		tmpl, err := template.New("main").Parse(templates.ScaffoldGo)
		util.Fatal(err, "template parse")

		err = tmpl.Execute(f, tv)
		util.Fatal(err, "template execute")

		err = os.WriteFile(fmt.Sprintf("%s/go.mod", tv.Project), []byte(fmt.Sprintf(`
module github.com/%s/%s

go 1.18
`, tv.Company, tv.Project)), 0o600)
		util.Fatal(err, "write go.mod")

		util.Stdoutf(`
Execute the following to compile and run scaffolded project:
	cd %s
	go mod tidy
	go build .
	tigris local up
	./%s
`, tv.Project, tv.Project)
	},
}
