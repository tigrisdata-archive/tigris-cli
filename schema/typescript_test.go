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

//nolint:golint,dupl
package schema

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tigrisdata/tigris-cli/config"
	api "github.com/tigrisdata/tigris-client-go/api/server/v1"
)

//nolint:funlen
func TestTypeScriptSchemaGeneratorHeaderFooter(t *testing.T) {
	cases := []struct {
		name string
		in   []*api.CollectionDescription
		exp  string
	}{
		{
			"simple",
			[]*api.CollectionDescription{
				{Collection: "products", Schema: []byte(`
export interface Product {
  name: string;
}

export const productSchema: TigrisSchema<Product> = {
  name: {
    type: TigrisDataTypes.STRING,
  },
};
`)},
				{Collection: "user_names", Schema: []byte(`
export interface UserName extends TigrisCollectionType {
  name: string;
}

export const userNameSchema: TigrisSchema<UserName> = {
  name: {
    type: TigrisDataTypes.STRING,
  },
};
`)},
			},
			`import {
  TigrisCollectionType,
  TigrisDataTypes,
  TigrisSchema,
} from "@tigrisdata/core/dist/types";

import {DB, Tigris, TigrisClientConfig} from "@tigrisdata/core";

export interface Product {
  name: string;
}

export const productSchema: TigrisSchema<Product> = {
  name: {
    type: TigrisDataTypes.STRING,
  },
};

export interface UserName extends TigrisCollectionType {
  name: string;
}

export const userNameSchema: TigrisSchema<UserName> = {
  name: {
    type: TigrisDataTypes.STRING,
  },
};

async function main() {
    var tigris = new Tigris({
      serverUrl: "localhost:8081",
    });
    var db = await tigris.createDatabaseIfNotExists("test_db");

    await Promise.all([
      db.createOrUpdateCollection<Product>("products", productSchema),
      db.createOrUpdateCollection<UserName>("user_names", userNameSchema),
    ])

    var collProduct = db.getCollection<Product>("products");

    const productcursor = collProduct.findMany();
    try {
      for await (const doc of productcursor) {
        console.log(doc)
      }
    } catch (error) {
        console.log(error)
    }

    //collProduct.insertOne(<Product>{})

    var collUserName = db.getCollection<UserName>("user_names");

    const userNamecursor = collUserName.findMany();
    try {
      for await (const doc of userNamecursor) {
        console.log(doc)
      }
    } catch (error) {
        console.log(error)
    }

    //collUserName.insertOne(<UserName>{})
}

main()
  .then(async () => {
    console.log("All done ...")
  })
  .catch(async (e) => {
    console.error(e)
    process.exit(1);
  })

// Check full API reference here: https://docs.tigrisdata.com/typescript/

// Build and run:
// * Put this content to index.ts
// * npm install @tigrisdata/core ts-node
// * npx ts-node index.ts
`,
		},
		{
			"has_time_uuid",
			[]*api.CollectionDescription{{Collection: "products", Schema: []byte(`
export interface Product {
  arrInts: string;
  bool: boolean;
  byte1: string;
  id: number;
  int64: string;
  // int64WithDesc field description
  int64WithDesc: string;
  name: string;
  price: number;
  time1: string;
  uUID1: string;
}

export const productSchema: TigrisSchema<Product> = {
  arrInts: {
    type: TigrisDataTypes.ARRAY,
    items: {
      type: TigrisDataTypes.INT64,
    },
  },
  bool: {
    type: TigrisDataTypes.BOOLEAN,
  },
  byte1: {
    type: TigrisDataTypes.BYTE_STRING,
  },
  id: {
    type: TigrisDataTypes.INT32,
  },
  int64: {
    type: TigrisDataTypes.INT64,
  },
  int64WithDesc: {
    type: TigrisDataTypes.INT64,
  },
  name: {
    type: TigrisDataTypes.STRING,
  },
  price: {
    type: TigrisDataTypes.NUMBER,
  },
  time1: {
    type: TigrisDataTypes.DATE_TIME,
  },
  uUID1: {
    type: TigrisDataTypes.UUID,
  },
};
`)}},
			`import {
  TigrisCollectionType,
  TigrisDataTypes,
  TigrisSchema,
} from "@tigrisdata/core/dist/types";

import {DB, Tigris, TigrisClientConfig} from "@tigrisdata/core";

export interface Product {
  arrInts: string;
  bool: boolean;
  byte1: string;
  id: number;
  int64: string;
  // int64WithDesc field description
  int64WithDesc: string;
  name: string;
  price: number;
  time1: string;
  uUID1: string;
}

export const productSchema: TigrisSchema<Product> = {
  arrInts: {
    type: TigrisDataTypes.ARRAY,
    items: {
      type: TigrisDataTypes.INT64,
    },
  },
  bool: {
    type: TigrisDataTypes.BOOLEAN,
  },
  byte1: {
    type: TigrisDataTypes.BYTE_STRING,
  },
  id: {
    type: TigrisDataTypes.INT32,
  },
  int64: {
    type: TigrisDataTypes.INT64,
  },
  int64WithDesc: {
    type: TigrisDataTypes.INT64,
  },
  name: {
    type: TigrisDataTypes.STRING,
  },
  price: {
    type: TigrisDataTypes.NUMBER,
  },
  time1: {
    type: TigrisDataTypes.DATE_TIME,
  },
  uUID1: {
    type: TigrisDataTypes.UUID,
  },
};

async function main() {
    var tigris = new Tigris({
      serverUrl: "localhost:8081",
    });
    var db = await tigris.createDatabaseIfNotExists("test_db");

    await Promise.all([
      db.createOrUpdateCollection<Product>("products", productSchema),
    ])

    var collProduct = db.getCollection<Product>("products");

    const productcursor = collProduct.findMany();
    try {
      for await (const doc of productcursor) {
        console.log(doc)
      }
    } catch (error) {
        console.log(error)
    }

    //collProduct.insertOne(<Product>{})
}

main()
  .then(async () => {
    console.log("All done ...")
  })
  .catch(async (e) => {
    console.error(e)
    process.exit(1);
  })

// Check full API reference here: https://docs.tigrisdata.com/typescript/

// Build and run:
// * Put this content to index.ts
// * npm install @tigrisdata/core ts-node
// * npx ts-node index.ts
`,
		},
	}

	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			for k, c := range v.in {
				m := make(map[string]interface{})
				m["ts"] = string(c.Schema)
				b, err := json.Marshal(m)
				require.NoError(t, err)

				v.in[k].Schema = b
			}

			config.DefaultConfig.URL = "localhost:8081"
			buf := scaffoldFromDB("test_db", v.in, "ts")
			assert.Equal(t, v.exp, strings.ReplaceAll(string(buf), "\r\n", "\n"))
		})
	}
}
