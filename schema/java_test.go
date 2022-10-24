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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tigrisdata/tigris-cli/config"
	api "github.com/tigrisdata/tigris-client-go/api/server/v1"
)

//nolint:funlen,maintidx
func TestJavaSchemaGeneratorHeaderFooter(t *testing.T) {
	cases := []struct {
		name string
		in   []*api.CollectionDescription
		exp  string
	}{
		{
			"simple",
			[]*api.CollectionDescription{
				{Collection: "products", Schema: []byte(`
class Product {
    private String name;

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public Product() {};

    public Product(
        String name
    ) {
        this.name = name;
    };

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }

        Product other = (Product) o;
        return
            name == other.name;
    }

    @Override
    public int hashCode() {
        return Objects.hash(
            name
        );
    }
}
`)},
				{Collection: "user_names", Schema: []byte(`
@TigrisCollection(value = "user_names")
class UserName implements TigrisDocumentCollectionType {
    private String name;

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public UserName() {};

    public UserName(
        String name
    ) {
        this.name = name;
    };

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }

        UserName other = (UserName) o;
        return
            name == other.name;
    }

    @Override
    public int hashCode() {
        return Objects.hash(
            name
        );
    }
}
`)},
			},
			`package com.tigrisdata.client.schema;

import com.tigrisdata.db.client.TigrisDatabase;
import com.tigrisdata.db.client.TigrisCollection;
import com.tigrisdata.db.client.config.TigrisConfiguration;
import com.tigrisdata.db.annotation.TigrisField;
import com.tigrisdata.db.annotation.TigrisPrimaryKey;
import com.tigrisdata.db.type.TigrisDocumentCollectionType;
import com.tigrisdata.db.client.StandardTigrisClient;
import com.tigrisdata.db.client.TigrisClient;
import com.tigrisdata.db.client.error.TigrisException;
import java.util.Objects;
import java.util.Arrays;

class Product {
    private String name;

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public Product() {};

    public Product(
        String name
    ) {
        this.name = name;
    };

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }

        Product other = (Product) o;
        return
            name == other.name;
    }

    @Override
    public int hashCode() {
        return Objects.hash(
            name
        );
    }
}

@TigrisCollection(value = "user_names")
class UserName implements TigrisDocumentCollectionType {
    private String name;

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public UserName() {};

    public UserName(
        String name
    ) {
        this.name = name;
    };

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }

        UserName other = (UserName) o;
        return
            name == other.name;
    }

    @Override
    public int hashCode() {
        return Objects.hash(
            name
        );
    }
}

public class TestDbApp {
    public static void main(String[] args) throws TigrisException {
        TigrisConfiguration configuration = TigrisConfiguration.newBuilder("localhost:8081")
            .withAuthConfig(new TigrisConfiguration.AuthConfig("paste client_id here", "paste client_secret here"))
            .build();

        TigrisClient client = StandardTigrisClient.getInstance(configuration);
        TigrisDatabase db = client.createDatabaseIfNotExists("test_db");

        db.createOrUpdateCollections(
            Product.class,
            UserName.class
        );

        TigrisCollection<Product> collProduct = db.getCollection(Product.class);
        collProduct.insert(new Product( /* Document fields here */));
        TigrisCollection<UserName> collUserName = db.getCollection(UserName.class);
        collUserName.insert(new UserName( /* Document fields here */));
    }
}

// Check full API reference here: https://docs.tigrisdata.com/java/

// Build and run:
// * mkdir -p src/main/java/com/tigrisdata/client/schema/
// * Put this output to src/main/java/com/tigrisdata/client/schema/TestDbApp.java
// * Put the following in the pom.xml
// * mvn clean compile

/*
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.tigrisdata</groupId>
    <artifactId>test_db</artifactId>
    <version>1.0.0</version>

    <dependencies>
        <dependency>
            <groupId>com.tigrisdata</groupId>
            <artifactId>tigris-client</artifactId>
            <version>1.0.0-beta.4</version>
        </dependency>
    </dependencies>
    <properties>
        <maven.compiler.target>1.8</maven.compiler.target>
        <maven.compiler.source>1.8</maven.compiler.source>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
        <project.reporting.outputEncoding>UTF-8</project.reporting.outputEncoding>
    </properties>
</project>
*/
`,
		},
		{
			"has_time_uuid",
			[]*api.CollectionDescription{{Collection: "products", Schema: []byte(`
class Product {
    private long[] arrInts;
    private boolean bool;
    private byte[] byte1;
    private int id;
    private long int64;
    @TigrisField(description = "field description")
    private long int64WithDesc;
    private String name;
    private double price;
    private Date time1;
    private UUID uUID1;

    public long[] getArrInts() {
        return arrInts;
    }

    public void setArrInts(long[] arrInts) {
        this.arrInts = arrInts;
    }

    public boolean isBool() {
        return bool;
    }

    public void setBool(boolean bool) {
        this.bool = bool;
    }

    public byte[] getByte1() {
        return byte1;
    }

    public void setByte1(byte[] byte1) {
        this.byte1 = byte1;
    }

    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }

    public long getInt64() {
        return int64;
    }

    public void setInt64(long int64) {
        this.int64 = int64;
    }

    public long getInt64WithDesc() {
        return int64WithDesc;
    }

    public void setInt64WithDesc(long int64WithDesc) {
        this.int64WithDesc = int64WithDesc;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public double getPrice() {
        return price;
    }

    public void setPrice(double price) {
        this.price = price;
    }

    public Date getTime1() {
        return time1;
    }

    public void setTime1(Date time1) {
        this.time1 = time1;
    }

    public UUID getUUID1() {
        return uUID1;
    }

    public void setUUID1(UUID uUID1) {
        this.uUID1 = uUID1;
    }

    public Product() {};

    public Product(
        long[] arrInts,
        boolean bool,
        byte[] byte1,
        int id,
        long int64,
        long int64WithDesc,
        String name,
        double price,
        Date time1,
        UUID uUID1
    ) {
        this.arrInts = arrInts;
        this.bool = bool;
        this.byte1 = byte1;
        this.id = id;
        this.int64 = int64;
        this.int64WithDesc = int64WithDesc;
        this.name = name;
        this.price = price;
        this.time1 = time1;
        this.uUID1 = uUID1;
    };

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }

        Product other = (Product) o;
        return
            Arrays.equals(arrInts, other.arrInts) &&
            bool == other.bool &&
            byte1 == other.byte1 &&
            id == other.id &&
            int64 == other.int64 &&
            int64WithDesc == other.int64WithDesc &&
            name == other.name &&
            price == other.price &&
            time1 == other.time1 &&
            uUID1 == other.uUID1;
    }

    @Override
    public int hashCode() {
        return Objects.hash(
            arrInts,
            bool,
            byte1,
            id,
            int64,
            int64WithDesc,
            name,
            price,
            time1,
            uUID1
        );
    }
}
`)}},
			`package com.tigrisdata.client.schema;

import com.tigrisdata.db.client.TigrisDatabase;
import com.tigrisdata.db.client.TigrisCollection;
import com.tigrisdata.db.client.config.TigrisConfiguration;
import com.tigrisdata.db.annotation.TigrisField;
import com.tigrisdata.db.annotation.TigrisPrimaryKey;
import com.tigrisdata.db.type.TigrisDocumentCollectionType;
import com.tigrisdata.db.client.StandardTigrisClient;
import com.tigrisdata.db.client.TigrisClient;
import com.tigrisdata.db.client.error.TigrisException;
import java.util.Objects;
import java.util.Arrays;
import java.util.UUID;
import java.util.Date;

class Product {
    private long[] arrInts;
    private boolean bool;
    private byte[] byte1;
    private int id;
    private long int64;
    @TigrisField(description = "field description")
    private long int64WithDesc;
    private String name;
    private double price;
    private Date time1;
    private UUID uUID1;

    public long[] getArrInts() {
        return arrInts;
    }

    public void setArrInts(long[] arrInts) {
        this.arrInts = arrInts;
    }

    public boolean isBool() {
        return bool;
    }

    public void setBool(boolean bool) {
        this.bool = bool;
    }

    public byte[] getByte1() {
        return byte1;
    }

    public void setByte1(byte[] byte1) {
        this.byte1 = byte1;
    }

    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }

    public long getInt64() {
        return int64;
    }

    public void setInt64(long int64) {
        this.int64 = int64;
    }

    public long getInt64WithDesc() {
        return int64WithDesc;
    }

    public void setInt64WithDesc(long int64WithDesc) {
        this.int64WithDesc = int64WithDesc;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public double getPrice() {
        return price;
    }

    public void setPrice(double price) {
        this.price = price;
    }

    public Date getTime1() {
        return time1;
    }

    public void setTime1(Date time1) {
        this.time1 = time1;
    }

    public UUID getUUID1() {
        return uUID1;
    }

    public void setUUID1(UUID uUID1) {
        this.uUID1 = uUID1;
    }

    public Product() {};

    public Product(
        long[] arrInts,
        boolean bool,
        byte[] byte1,
        int id,
        long int64,
        long int64WithDesc,
        String name,
        double price,
        Date time1,
        UUID uUID1
    ) {
        this.arrInts = arrInts;
        this.bool = bool;
        this.byte1 = byte1;
        this.id = id;
        this.int64 = int64;
        this.int64WithDesc = int64WithDesc;
        this.name = name;
        this.price = price;
        this.time1 = time1;
        this.uUID1 = uUID1;
    };

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }

        Product other = (Product) o;
        return
            Arrays.equals(arrInts, other.arrInts) &&
            bool == other.bool &&
            byte1 == other.byte1 &&
            id == other.id &&
            int64 == other.int64 &&
            int64WithDesc == other.int64WithDesc &&
            name == other.name &&
            price == other.price &&
            time1 == other.time1 &&
            uUID1 == other.uUID1;
    }

    @Override
    public int hashCode() {
        return Objects.hash(
            arrInts,
            bool,
            byte1,
            id,
            int64,
            int64WithDesc,
            name,
            price,
            time1,
            uUID1
        );
    }
}

public class TestDbApp {
    public static void main(String[] args) throws TigrisException {
        TigrisConfiguration configuration = TigrisConfiguration.newBuilder("localhost:8081")
            .withAuthConfig(new TigrisConfiguration.AuthConfig("paste client_id here", "paste client_secret here"))
            .build();

        TigrisClient client = StandardTigrisClient.getInstance(configuration);
        TigrisDatabase db = client.createDatabaseIfNotExists("test_db");

        db.createOrUpdateCollections(
            Product.class
        );

        TigrisCollection<Product> collProduct = db.getCollection(Product.class);
        collProduct.insert(new Product( /* Document fields here */));
    }
}

// Check full API reference here: https://docs.tigrisdata.com/java/

// Build and run:
// * mkdir -p src/main/java/com/tigrisdata/client/schema/
// * Put this output to src/main/java/com/tigrisdata/client/schema/TestDbApp.java
// * Put the following in the pom.xml
// * mvn clean compile

/*
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.tigrisdata</groupId>
    <artifactId>test_db</artifactId>
    <version>1.0.0</version>

    <dependencies>
        <dependency>
            <groupId>com.tigrisdata</groupId>
            <artifactId>tigris-client</artifactId>
            <version>1.0.0-beta.4</version>
        </dependency>
    </dependencies>
    <properties>
        <maven.compiler.target>1.8</maven.compiler.target>
        <maven.compiler.source>1.8</maven.compiler.source>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
        <project.reporting.outputEncoding>UTF-8</project.reporting.outputEncoding>
    </properties>
</project>
*/
`,
		},
	}

	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			for k, c := range v.in {
				m := make(map[string]interface{})
				m["java"] = string(c.Schema)
				b, err := json.Marshal(m)
				require.NoError(t, err)

				v.in[k].Schema = b
			}

			config.DefaultConfig.URL = "localhost:8081"
			buf := scaffoldFromDB("test_db", v.in, "java")
			assert.Equal(t, v.exp, string(buf))
		})
	}
}
