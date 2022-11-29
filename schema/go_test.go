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
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	api "github.com/tigrisdata/tigris-client-go/api/server/v1"
)

//nolint:funlen
func TestGoSchemaGeneratorHeaderFooter(t *testing.T) {
	cases := []struct {
		name string
		in   []*api.CollectionDescription
		exp  string
	}{
		{
			"simple",
			[]*api.CollectionDescription{
				{Collection: "products", Schema: []byte(`
type Product struct {
	Name string ` + "`" + `json:"name"` + "`" + `
}
`)},
				{Collection: "user_names", Schema: []byte(`
type UserName struct {
	Name string ` + "`" + `json:"name"` + "`" + `
}
`)},
			},
			`package main

import (
	"context"
	"fmt"
	"time"

	"github.com/tigrisdata/tigris-client-go/config"
	"github.com/tigrisdata/tigris-client-go/tigris"
)

type Product struct {
	Name string ` + "`" + `json:"name"` + "`" + `
}

type UserName struct {
	Name string ` + "`" + `json:"name"` + "`" + `
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := tigris.NewClient(ctx, &config.Client{Driver: config.Driver{URL: ""}})

	if err != nil {
		panic(err)
	}
	defer client.Close()

	db, err := client.OpenDatabase(ctx, "test_db",
		&Product{},
		&UserName{},
	)
	if err != nil {
		panic(err)
	}

	collProduct := tigris.GetCollection[Product](db)

	itProduct, err := collProduct.ReadAll(ctx)
	if err != nil {
		panic(err)
	}

	var docProduct Product
	for i := 0; itProduct.Next(&docProduct) && i < 3; i++ {
		fmt.Printf("%+v\n", docProduct)
	}

	itProduct.Close()

	//collProduct.Insert(context.TODO(), &Product{/* Insert fields here */})

	collUserName := tigris.GetCollection[UserName](db)

	itUserName, err := collUserName.ReadAll(ctx)
	if err != nil {
		panic(err)
	}

	var docUserName UserName
	for i := 0; itUserName.Next(&docUserName) && i < 3; i++ {
		fmt.Printf("%+v\n", docUserName)
	}

	itUserName.Close()

	//collUserName.Insert(context.TODO(), &UserName{/* Insert fields here */})
}

// Check full API reference here: https://docs.tigrisdata.com/golang/

// Compile and run:
// * Put this output to main.go
// * go mod init
// * go mod tidy
// * go build -o test_db .
// * ./test_db
`,
		},
		{
			"has_time_uuid",
			[]*api.CollectionDescription{{Collection: "products", Schema: []byte(`
type Product struct {
	ArrInts []int64 ` + "`" + `json:"arrInts"` + "`" + `
	Bool bool ` + "`" + `json:"bool"` + "`" + `
	Byte1 []byte ` + "`" + `json:"byte1"` + "`" + `
	Id int32 ` + "`" + `json:"id"` + "`" + `
	Int64 int64 ` + "`" + `json:"int64"` + "`" + `
	// Int64WithDesc field description
	Int64WithDesc int64 ` + "`" + `json:"int64WithDesc"` + "`" + `
	Name string ` + "`" + `json:"name"` + "`" + `
	Price float64 ` + "`" + `json:"price"` + "`" + `
	Time1 time.Time ` + "`" + `json:"time1"` + "`" + `
	UUID1 uuid.UUID ` + "`" + `json:"uUID1"` + "`" + `
}
`)}},
			`package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tigrisdata/tigris-client-go/config"
	"github.com/tigrisdata/tigris-client-go/tigris"
)

type Product struct {
	ArrInts []int64 ` + "`" + `json:"arrInts"` + "`" + `
	Bool bool ` + "`" + `json:"bool"` + "`" + `
	Byte1 []byte ` + "`" + `json:"byte1"` + "`" + `
	Id int32 ` + "`" + `json:"id"` + "`" + `
	Int64 int64 ` + "`" + `json:"int64"` + "`" + `
	// Int64WithDesc field description
	Int64WithDesc int64 ` + "`" + `json:"int64WithDesc"` + "`" + `
	Name string ` + "`" + `json:"name"` + "`" + `
	Price float64 ` + "`" + `json:"price"` + "`" + `
	Time1 time.Time ` + "`" + `json:"time1"` + "`" + `
	UUID1 uuid.UUID ` + "`" + `json:"uUID1"` + "`" + `
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := tigris.NewClient(ctx, &config.Client{Driver: config.Driver{URL: ""}})

	if err != nil {
		panic(err)
	}
	defer client.Close()

	db, err := client.OpenDatabase(ctx, "test_db",
		&Product{},
	)
	if err != nil {
		panic(err)
	}

	collProduct := tigris.GetCollection[Product](db)

	itProduct, err := collProduct.ReadAll(ctx)
	if err != nil {
		panic(err)
	}

	var docProduct Product
	for i := 0; itProduct.Next(&docProduct) && i < 3; i++ {
		fmt.Printf("%+v\n", docProduct)
	}

	itProduct.Close()

	//collProduct.Insert(context.TODO(), &Product{/* Insert fields here */})
}

// Check full API reference here: https://docs.tigrisdata.com/golang/

// Compile and run:
// * Put this output to main.go
// * go mod init
// * go mod tidy
// * go build -o test_db .
// * ./test_db
`,
		},
	}

	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			for k, c := range v.in {
				m := make(map[string]interface{})
				m["go"] = string(c.Schema)
				b, err := json.Marshal(m)
				require.NoError(t, err)

				v.in[k].Schema = b
			}

			buf := scaffoldFromDB("test_db", v.in, "go")
			assert.Equal(t, v.exp, strings.ReplaceAll(string(buf), "\r\n", "\n"))
		})
	}
}
