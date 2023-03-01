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

package login

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/pkg/browser"
	"github.com/rs/zerolog/log"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/templates"
	"github.com/tigrisdata/tigris-cli/util"
	ecode "github.com/tigrisdata/tigris-client-go/code"
	"github.com/tigrisdata/tigris-client-go/driver"
	"golang.org/x/net/context/ctxhttp"
	"golang.org/x/oauth2"
)

type instance struct {
	clientID string
	authHost string
	audience string
}

var (
	ErrStateMismatched  = fmt.Errorf("state is not matched")
	ErrInstanceNotFound = fmt.Errorf("instance not found")
)

var (
	callbackHost = "localhost:8585"
	callbackURL  = "http://" + callbackHost + "/callback"

	instances = map[string]instance{
		"api.dev.tigrisdata.cloud": {
			clientID: "zXKltgV3JhGwUqOCUWNmtU7aX5TytKGx",
			authHost: "https://tigrisdata-dev.us.auth0.com/",
			audience: "https://tigris-api-dev",
		},
		"api.perf.tigrisdata.cloud": {
			clientID: "zXKltgV3JhGwUqOCUWNmtU7aX5TytKGx",
			authHost: "https://tigrisdata-dev.us.auth0.com/",
			audience: "https://tigris-api-perf",
		},
		"api.preview.tigrisdata.cloud": {
			clientID: "GS8PrHA1aYblUR73yitqomc40ZYZ81jF",
			authHost: "https://tigrisdata.us.auth0.com/",
			audience: "https://tigris-api-preview",
		},
	}

	code  string
	token *oauth2.Token
)

type tmplVars struct {
	Title string
	Error string
}

func authorize(auth *Authenticator, state string, audience string) error {
	authURL := auth.AuthCodeURL(state, oauth2.SetAuthURLParam("audience", audience))

	log.Debug().Str("url", authURL).Msg("Open login link in the browser")

	util.Stderrf("Opening login page in the browser. Please continue login flow there.\n")

	if err := browser.OpenURL(authURL); err != nil {
		return util.Error(err, "Error opening login page")
	}

	return nil
}

func getToken(auth *Authenticator, code string) *oauth2.Token {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token, err := auth.Exchange(ctx, code)
	util.Fatal(err, "retrieving token")

	return token
}

func callback(wg *sync.WaitGroup, server *http.Server, auth *Authenticator, instanceURL string, state string) error {
	mux := http.NewServeMux()
	server.Handler = mux

	var retError error

	log.Debug().Msg("starting callback server")

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()

		if r.URL.Query().Get("state") != state {
			retError = ErrStateMismatched
			util.ExecTemplate(w, templates.LoginError, &tmplVars{Title: instanceURL, Error: retError.Error()})
			log.Debug().Str("want", state).Str("got", r.URL.Query().Get("state")).Msg("state is not matched")

			return
		}

		code = r.URL.Query().Get("code")

		log.Debug().Str("code", code).Msg("Auth code retrieved")

		token = getToken(auth, code)

		log.Debug().Str("accessToken", token.AccessToken).Str("refreshToken", token.RefreshToken).
			Msg("Access token retrieved")

		config.DefaultConfig.Token = token.AccessToken
		config.DefaultConfig.URL = instanceURL

		if err := config.Save(config.DefaultName, config.DefaultConfig); err != nil {
			retError = err
			log.Err(err).Msg("Error saving config")
			util.ExecTemplate(w, templates.LoginError, &tmplVars{Title: instanceURL, Error: err.Error()})

			return
		}

		util.ExecTemplate(w, templates.LoginSuccessful, &tmplVars{Title: instanceURL})

		util.Stderrf("Successfully logged in\n")
	})

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
		log.Debug().Msg("Callback server up-check received")
	})

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		_ = util.Error(err, "callback server failed")

		if retError == nil {
			retError = err
		}
	}

	log.Debug().Msg("callback server finished")

	return retError
}

func genRandomState() (string, error) {
	stateBin := make([]byte, 32)

	if n, err := rand.Read(stateBin); err != nil || n != 32 {
		return "", util.Error(fmt.Errorf("failed to generate random state %w", err), "")
	}

	return base64.StdEncoding.EncodeToString(stateBin), nil
}

type Authenticator struct {
	*oidc.Provider
	oauth2.Config
}

func Ensure(cctx context.Context, fn func(ctx context.Context) error) {
	ctx, cancel := util.GetContext(cctx)

	err := fn(ctx)

	cancel()

	if err == nil {
		return
	}

	var ep *driver.Error
	if !errors.As(err, &ep) || ep.Code != ecode.Unauthenticated ||
		os.Getenv(driver.EnvClientID) != "" || os.Getenv(driver.EnvClientSecret) != "" ||
		config.DefaultConfig.ClientID != "" || config.DefaultConfig.ClientSecret != "" || !util.IsTTY(os.Stdin) {
		util.PrintError(err)
		os.Exit(1)
	}

	lctx, lcancel := util.GetContext(cctx)

	if err = CmdLow(lctx, GetHost("")); err != nil {
		lcancel()
		os.Exit(1)
	}

	lcancel()

	_ = client.Init(&config.DefaultConfig)

	ctx1, cancel1 := util.GetContext(cctx)

	err = fn(ctx1)

	cancel1()

	if err != nil {
		util.PrintError(err)
		os.Exit(1)
	}
}

func GetHost(host string) string {
	if host == "" {
		host = config.DefaultURL
		if os.Getenv(driver.EnvURL) != "" {
			host = os.Getenv(driver.EnvURL)
		}

		if config.DefaultConfig.URL != "" {
			host = config.DefaultConfig.URL
		}
	}

	return host
}

func waitCallbackServerUp() {
	c := http.Client{Timeout: 5 * time.Second}

	log.Debug().Msg("Waiting for callback server start")

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		r, err := ctxhttp.Get(ctx, &c, "http://"+callbackHost+"/ping")

		cancel()

		if err == nil {
			_ = util.Error(r.Body.Close(), "closing ping body")

			if r.StatusCode == http.StatusOK {
				log.Debug().Msg("Callback server has started")

				break
			}
		}
	}
}

func localLogin(host string) bool {
	if host == "local" || host == "dev" || strings.HasPrefix(host, "localhost") {
		// handle the cases without a port
		if host == "local" || host == "dev" || host == "localhost" {
			host = "localhost:8081"
		}

		LocalLogin(host)

		return true
	}

	return false
}

//nolint:golint,funlen
func CmdLow(_ context.Context, host string) error {
	var err error

	host = GetHost(host)

	if localLogin(host) {
		return nil
	}

	state, err := genRandomState()
	if err != nil {
		return err
	}

	inst, ok := instances[host]
	if !ok {
		return util.Error(fmt.Errorf("%w: %s", ErrInstanceNotFound, host), "Instance config not found")
	}

	p, err := oidc.NewProvider(context.Background(), inst.authHost)
	if err != nil {
		return util.Error(err, "authorize error")
	}

	auth := &Authenticator{
		Provider: p,
		Config: oauth2.Config{
			ClientID:    inst.clientID,
			RedirectURL: callbackURL,
			Endpoint:    p.Endpoint(),
		},
	}

	var callbackErr, authorizeErr error

	server := &http.Server{Addr: callbackHost, ReadHeaderTimeout: util.GetTimeout()}

	var (
		wg1 sync.WaitGroup
		wg  sync.WaitGroup
	)

	wg1.Add(1)
	wg.Add(1)

	go func() {
		defer wg.Done()

		callbackErr = callback(&wg1, server, auth, host, state)
	}()

	waitCallbackServerUp()

	authorizeErr = authorize(auth, state, inst.audience)

	log.Debug().Msg("Waiting for login flow to finish in the browser")

	wg1.Wait()

	log.Debug().Msg("Shutting down callback server")

	ctx1, cancel := util.GetContext(context.Background())
	defer cancel()

	if err = server.Shutdown(ctx1); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return util.Error(err, "shutdown callback server failed")
	}

	log.Debug().Msg("Callback handler finished")

	wg.Wait()

	if authorizeErr != nil {
		return util.Error(authorizeErr, "authorize error")
	}

	return util.Error(callbackErr, "callback error")
}

func LocalLogin(host string) {
	config.DefaultConfig.ClientSecret = ""
	config.DefaultConfig.ClientID = ""
	config.DefaultConfig.Token = ""
	config.DefaultConfig.URL = host

	err := config.Save(config.DefaultName, config.DefaultConfig)
	util.Fatal(err, "saving config for localhost")

	util.Stderrf("Successfully logged in to %s\n", host)
}
