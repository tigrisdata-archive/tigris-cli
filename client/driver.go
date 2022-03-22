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

package client

import (
	"context"

	"github.com/tigrisdata/tigrisdb-cli/config"
	"github.com/tigrisdata/tigrisdb-client-go/driver"
)

// D is single instance of client
var D driver.Driver

func Init(config config.Config) error {
	driver.DefaultProtocol = driver.HTTP
	drv, err := driver.NewDriver(context.Background(), config.URL, &driver.Config{Token: config.Token})
	if err != nil {
		return err
	}

	D = drv

	return nil
}

// Get returns an instance of instance of client
func Get() driver.Driver {
	return D
}
