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
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/util"
	"github.com/tigrisdata/tigris-client-go/driver"
	"github.com/tigrisdata/tigris-client-go/schema"
)

var ErrUnsupportedSchemaTemplate = fmt.Errorf("unknown bootstrap schema template")

func Schema(ctx context.Context, db string, templatesPath string, name string, create bool, stdout bool) error {
	rootPath := filepath.Join(templatesPath, "schema", name)
	ffs := os.DirFS(rootPath)

	log.Debug().Str("rootPath", rootPath).Msg("schema bootstrap")

	if _, err := fs.ReadDir(ffs, "."); err != nil {
		list := util.ListDir(filepath.Join(templatesPath, "schema"))
		util.Infof("Available schema templates:")

		for k := range list {
			util.Infof("\t* %s", k)
		}

		util.Infof("")

		util.Fatal(ErrUnsupportedSchemaTemplate, "unknown schema '%s': %s", name, err.Error())
	}

	err := client.Transact(ctx, db, func(ctx context.Context, tx driver.Tx) error {
		return fs.WalkDir(ffs, ".", func(path string, d fs.DirEntry, err error) error {
			util.Fatal(err, "walk embedded template directory")

			log.Debug().Str("path", path).Msg("walk-dir")

			if !d.IsDir() && !strings.HasPrefix(d.Name(), ".") {
				b, err := fs.ReadFile(ffs, path)
				util.Fatal(err, "read template file")

				var sch schema.Schema

				err = json.Unmarshal(b, &sch)
				util.Fatal(err, "unmarshal schema")

				if stdout {
					util.Stdoutf("%s\n", string(b))
				} else if !create {
					err := os.WriteFile(fmt.Sprintf("%v.json", sch.Name), b, 0o600)
					util.Fatal(err, "writing schema file")
				} else if err = client.Get().UseDatabase(db).CreateOrUpdateCollection(ctx, sch.Name, b); err != nil {
					return err
				}
			}

			return nil
		})
	})

	return err
}
