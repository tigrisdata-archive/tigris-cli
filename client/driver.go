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

package client

import (
	"context"
	"crypto/tls"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/util"
	cconfig "github.com/tigrisdata/tigris-client-go/config"
	"github.com/tigrisdata/tigris-client-go/driver"
)

var (
	// D is single instance of client.
	D driver.Driver

	// M is single instance of management service client.
	M   driver.Management
	cfg *cconfig.Driver

	// O is single instance of observability service client.
	O driver.Observability
)

func initConfig(inCfg *config.Config) {
	cfg = &cconfig.Driver{
		URL:          inCfg.URL,
		ClientID:     inCfg.ClientID,
		ClientSecret: inCfg.ClientSecret,
		Token:        inCfg.Token,
		Protocol:     inCfg.Protocol,
		Branch:       inCfg.Branch,
		SkipLocalTLS: inCfg.SkipLocalTLS,
	}

	if !cfg.SkipLocalTLS && (inCfg.UseTLS || (cfg.URL == "" && cfg.Protocol == "") ||
		strings.HasSuffix(cfg.URL, config.Domain)) {
		cfg.TLS = &tls.Config{MinVersion: tls.VersionTLS12}
	}
}

func Init(inCfg *config.Config) error {
	initConfig(inCfg)

	if D != nil {
		log.Err(D.Close()).Msg("driver close")
	}

	if O != nil {
		log.Err(O.Close()).Msg("observability close")
	}

	if M != nil {
		log.Err(M.Close()).Msg("management close")
	}

	D = nil
	M = nil
	O = nil

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

// Get returns an instance of client.
func Get() driver.Driver {
	initConfig(&config.DefaultConfig)

	err := InitLow()
	util.Fatal(err, "tigris client initialization low")

	return D
}

func GetDB() driver.Database {
	return Get().UseDatabase(config.GetProjectName())
}

func GetSearch() driver.SearchClient {
	return Get().UseSearch(config.GetProjectName())
}

// ManagementGet returns an instance of authentication API client.
func ManagementGet() driver.Management {
	initConfig(&config.DefaultConfig)

	if M == nil {
		ctx, cancel := util.GetContext(context.Background())
		defer cancel()

		drv, err := driver.NewManagement(ctx, cfg)
		util.Fatal(err, "tigris management client initialization")

		M = drv
	}

	return M
}

func ObservabilityGet() driver.Observability {
	initConfig(&config.DefaultConfig)

	if O == nil {
		ctx, cancel := util.GetContext(context.Background())
		defer cancel()

		drv, err := driver.NewObservability(ctx, cfg)
		util.Fatal(err, "tigris observability client initialization")

		O = drv
	}

	return O
}

func Transact(bctx context.Context, db string, fn func(ctx context.Context, tx driver.Tx) error) error {
	ctx, cancel := util.GetContext(bctx)
	defer cancel()

	tx, err := Get().UseDatabase(db).BeginTx(ctx)
	if err != nil {
		return util.Error(err, "begin transaction")
	}

	defer func() { _ = tx.Rollback(ctx) }()

	if err = fn(ctx, tx); err != nil {
		return err
	}

	err = tx.Commit(ctx)

	return util.Error(err, "commit transaction")
}
