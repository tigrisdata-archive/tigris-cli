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
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/util"
	cconfig "github.com/tigrisdata/tigris-client-go/config"
	"github.com/tigrisdata/tigris-client-go/driver"
)

// D is single instance of client
var D driver.Driver

// A is single instance of auth service client
var A driver.Auth
var cfg *cconfig.Driver

func Init(config config.Config) error {
	proto := strings.ToLower(strings.Trim(config.Protocol, " "))
	if proto == "grpc" {
		driver.DefaultProtocol = driver.GRPC
	} else if proto == "https" || proto == "http" {
		driver.DefaultProtocol = driver.HTTP
	} else if proto != "" {
		return fmt.Errorf("unknown protocol set by TIGRIS_PROTOCOL: %s. allowed: grpc, http, https", proto)
	}

	// URL prefix has precedence over environment variable
	url := config.URL
	if strings.HasPrefix(config.URL, "http://") {
		driver.DefaultProtocol = driver.HTTP
	} else if strings.HasPrefix(config.URL, "https://") {
		driver.DefaultProtocol = driver.HTTP
	} else if strings.HasPrefix(config.URL, "grpc://") {
		driver.DefaultProtocol = driver.GRPC
		url = strings.TrimPrefix(config.URL, "grpc://")
	}

	//Client would use HTTPS if scheme is not explicitly specified
	//Avoid this for localhost connections
	if !strings.Contains(url, "://") && driver.DefaultProtocol == driver.HTTP &&
		(strings.HasPrefix(url, "localhost") || strings.HasPrefix(url, "127.0.0.1")) {
		url = "http://" + url
	}

	cfg = &cconfig.Driver{
		URL:               url,
		ApplicationId:     config.ApplicationID,
		ApplicationSecret: config.ApplicationSecret,
		Token:             config.Token,
	}

	if config.UseTLS || cfg.ApplicationSecret != "" || cfg.ApplicationId != "" || cfg.Token != "" {
		cfg.TLS = &tls.Config{}
	}

	_ = os.Unsetenv("TIGRIS_PROTOCOL")

	return nil
}

func InitLow() error {
	if D == nil {
		ctx, cancel := util.GetContext(context.Background())
		defer cancel()

		drv, err := driver.NewDriver(ctx, cfg)
		if err != nil {
			return err
		}
		D = drv
	}
	return nil
}

// Get returns an instance of client
func Get() driver.Driver {
	if err := InitLow(); err != nil {
		util.Error(err, "tigris client initialization failed")
	}
	return D
}

// AuthGet returns an instance of authentication API client
func AuthGet() driver.Auth {
	if A == nil {
		ctx, cancel := util.GetContext(context.Background())
		defer cancel()

		drv, err := driver.NewAuth(ctx, cfg)
		if err != nil {
			util.Error(err, "tigris client initialization failed")
		}
		A = drv
	}
	return A
}

func Transact(bctx context.Context, db string, fn func(ctx context.Context, tx driver.Tx)) {
	ctx, cancel := util.GetContext(bctx)
	defer cancel()

	tx, err := Get().BeginTx(ctx, db)
	if err != nil {
		util.Error(err, "begin transaction failed")
	}
	defer func() { _ = tx.Rollback(ctx) }()

	fn(ctx, tx)

	if err := tx.Commit(ctx); err != nil {
		util.Error(err, "commit transaction failed")
	}
}
