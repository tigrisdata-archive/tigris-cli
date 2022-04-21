package scaffold

import (
	"fmt"
	"html/template"
	"os"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/util"
)

type TemplateVars struct {
	Company     string
	Project     string
	Db          string
	Collections []string
}

var TemplateBody = `
package main

import (
	"context"
	"fmt"

	"github.com/tigrisdata/tigris-client-go/tigris"
	"github.com/tigrisdata/tigris-client-go/config"
	"github.com/tigrisdata/tigris-client-go/filter"
	"github.com/tigrisdata/tigris-client-go/update"
)
{{range .Collections}}
// {{.}} is a collection of documents
type {{.}} struct {
	StrField  string ` + "`" + `json:"str_field" tigris:"primary_key"` + "`" + `
	IntField  int64
	BoolField bool ` + "`" + `json:",omitempty"` + "`" + `
}
{{end}}
{{range .Collections}}
func handle{{.}}(ctx context.Context, db *tigris.Database) error {
	coll := tigris.GetCollection[{{.}}](db)

	log("Executing operations on {{.}}\n")

	d1 := &{{.}}{StrField: "value1", IntField: 111, BoolField: true}
	d2 := &{{.}}{StrField: "value2", IntField: 222, BoolField: false}

	log("Inserting documents into {{.}}:\n\t%+v\n\t%+v\n", d1, d2)

	if _, err := coll.Insert(ctx, d1, d2); err != nil {
		return err
	}

	d3 := &{{.}}{StrField: "value2", IntField: 333}

	log("Replacing document into {{.}}:\n\t%+v\n", d3)

	if _, err := coll.InsertOrReplace(ctx, d3); err != nil {
		return err
	}

	log("Reading one document from {{.}} where StrField=%v\n", "value1")

	doc1, err := coll.ReadOne(ctx, filter.Eq("str_field", "value1"))
	if err != nil {
		return err
	}

	log("\t%+v\n", doc1)

	log("Updating documents int {{.}} where StrField=%v Or StrField=%v, Setting IntField=123\n", "value1", "value3")

	if _, err = coll.Update(ctx,
		filter.Or(
			filter.Eq("str_field", "value1"),
			filter.Eq("str_field", "value3"),
		),
		update.Set("IntField", 123),
	); err != nil {
		return err
	}

	log("Reading all documents from {{.}}\n")

	it, err := coll.ReadAll(ctx)
	if err != nil {
		return err
	}

	var doc2 {{.}}
	for it.Next(&doc2) {
		fmt.Printf("\t%+v\n", doc2)
	}

	if it.Err() != nil {
		return it.Err()
	}

	log("Deleting documents from {{.}} where StrField=%v\n", "value2")

	_, err = coll.Delete(ctx, filter.Eq("str_field", "value2"))
	if err != nil {
		return err
	}

	log("Execute operations in a transaction\n")

	// Execute operations in a transaction
	err = db.Tx(ctx, func(ctx context.Context, tx *tigris.Tx) error {
		log("Get transactional collection object for {{.}}\n")

		txColl := tigris.GetTxCollection[{{.}}](tx)
		if err != nil {
			return err
		}

		d4 := &{{.}}{StrField: "value2", IntField: 111, BoolField: true}
		d5 := &{{.}}{StrField: "value3", IntField: 222}

		log("Inserting documents into {{.}}:\n\t%+v\n\t%+v\n", d4, d5)

		if _, err = txColl.Insert(ctx, d4, d5); err != nil {
			return err
		}

		log("Updating documents int {{.}} where StrField=%v\n", "value2")

		if _, err = coll.Update(ctx, filter.Eq("str_field", "value2"),
			update.Set("IntField", 678)); err != nil {
			return err
		}

		if _, err = coll.Delete(ctx, filter.Eq("str_field", "value1")); err != nil {
			return err
		}

		log("Reading documents from {{.}} where StrField=%v\n", "value1")

		it, err = txColl.ReadAll(ctx)
		if err != nil {
			return err
		}

		var doc3 {{.}}
		for it.Next(&doc3) {
			fmt.Printf("\t%+v\n", doc3)
		}

		return it.Err()
	})
	if err != nil {
		return err
	}
	log("Committed transaction to {{.}}\n")

	log("Dropping {{.}}\n")

	return coll.Drop(ctx)
}
{{end}}
func main() {
	ctx := context.TODO()

	log("Create or open existing database, update schema of existing collections\n")

    cfg := &config.Database{Driver: config.Driver{URL: "localhost:8081"}}
	db, err := tigris.OpenDatabase(ctx, cfg, "{{.Db}}", {{range .Collections}}
		&{{.}}{},
{{end}})
	if err != nil {
		panic(err)
	}
{{range .Collections}}
	if err := handle{{.}}(context.TODO(), db); err != nil {
		panic(err)
	}
{{end}}
	log("SUCCESS\n")
}

func log(format string, args ...interface{}) {
	_, _ = fmt.Printf(format, args...)
}
`

var GoCmd = &cobra.Command{
	Use:     "go {Company} {Project} {Db} {Collection}...",
	Aliases: []string{"golang"},
	Short:   "Scaffold a new Go project",
	Args:    cobra.MinimumNArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		tv := &TemplateVars{
			Company:     args[0],
			Project:     args[1],
			Db:          args[2],
			Collections: args[3:],
		}

		if _, err := os.Stat(tv.Project); !os.IsNotExist(err) {
			util.Error(fmt.Errorf("directory %v already exists", tv.Project), "")
		}

		err := os.Mkdir(tv.Project, 0755)
		if err != nil {
			util.Error(err, "")
		}

		f, err := os.Create(tv.Project + "/main.go")
		if err != nil {
			util.Error(err, "")
		}
		defer func() {
			err := f.Close()
			util.Error(err, "failed to close main.go")
		}()

		tmpl, err := template.New("main").Parse(TemplateBody)
		if err != nil {
			util.Error(err, "")
		}
		if err := tmpl.Execute(f, tv); err != nil {
			util.Error(err, "")
		}

		if err := os.WriteFile(fmt.Sprintf("%s/go.mod", tv.Project), []byte(fmt.Sprintf(`
module github.com/%s/%s

go 1.18
`, tv.Company, tv.Project)), 0644); err != nil {
			util.Error(err, "")
		}

		util.Stdout(`
Execute the following to compile and run scaffolded project:
	cd %s
	go mod tidy
	go build .
	tigris local up
	./%s
`, tv.Project, tv.Project)
	},
}
