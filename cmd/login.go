package cmd

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/pkg/browser"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/templates"
	"github.com/tigrisdata/tigris-cli/util"
	"golang.org/x/oauth2"
)

type instance struct {
	clientID string
	authHost string
	audience string
}

var (
	callbackHost = "localhost:8585"
	callbackURL  = "http://" + callbackHost + "/callback"

	instances = map[string]instance{
		"api.dev.tigrisdata.cloud": {
			clientID: "zXKltgV3JhGwUqOCUWNmtU7aX5TytKGx",
			authHost: "https://tigrisdata-dev.us.auth0.com/",
			audience: "https://tigris-api-dev",
		},
		"api.preview.tigrisdata.cloud": {
			clientID: "GS8PrHA1aYblUR73yitqomc40ZYZ81jF",
			authHost: "https://tigrisdata.us.auth0.com/",
			audience: "https://tigris-api-preview",
		},
	}

	code  string
	token *oauth2.Token

	tmplSuccess *template.Template
	tmplError   *template.Template
)

type tmplVars struct {
	Title string
	Error string
}

func authorize(auth *Authenticator, state string, audience string) error {
	authURL := auth.AuthCodeURL(state, oauth2.SetAuthURLParam("audience", audience))

	log.Debug().Str("url", authURL).Msg("Open login link in the browser")

	util.Stdout("Opening login page in the browser. Please continue login flow there.\n")

	if err := browser.OpenURL(authURL); err != nil {
		util.Error(err, "Error opening login page")
		return err
	}

	return nil
}

func getToken(auth *Authenticator, code string) *oauth2.Token {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token, err := auth.Exchange(ctx, code)
	if err != nil {
		util.Error(err, "error retrieving token")
	}

	return token
}

func execTemplate(w io.Writer, tmpl *template.Template, vars *tmplVars) {
	if err := tmpl.Execute(w, vars); err != nil {
		log.Err(err).Msg("execute template failed")
	}
}

func callback(wg *sync.WaitGroup, server *http.Server, auth *Authenticator, instanceURL string, state string) error {
	mux := http.NewServeMux()
	server.Handler = mux

	var retError error

	log.Debug().Msg("starting callback server")

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()

		if r.URL.Query().Get("state") != state {
			retError = fmt.Errorf("state is not matched")
			execTemplate(w, tmplError, &tmplVars{Title: instanceURL, Error: retError.Error()})
			log.Debug().Str("want", state).Str("got", r.URL.Query().Get("state")).Msg("state is not matched")
			return
		}

		code = r.URL.Query().Get("code")

		log.Debug().Str("code", code).Msg("Auth code retrieved")

		token = getToken(auth, code)

		log.Debug().Str("accessToken", token.AccessToken).Str("refreshToken", token.RefreshToken).Msg("Access token retrieved")

		config.DefaultConfig.Token = token.AccessToken
		config.DefaultConfig.URL = instanceURL

		if err := config.Save(config.DefaultName, config.DefaultConfig); err != nil {
			retError = err
			log.Err(err).Msg("Error saving config")
			execTemplate(w, tmplError, &tmplVars{Title: instanceURL, Error: err.Error()})
			return
		}

		execTemplate(w, tmplSuccess, &tmplVars{Title: instanceURL})

		util.Stdout("Successfully logged in\n")
	})

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
		log.Debug().Msg("Callback server up-check received")
	})

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		util.Error(err, "callback server failed")
		if retError == nil {
			retError = err
		}
	}

	log.Debug().Msg("callback server finished")
	return retError
}

func genRandomState() string {
	stateBin := make([]byte, 32)
	n, err := rand.Read(stateBin)
	if err != nil || n != 32 {
		util.Error(fmt.Errorf("failed to generate random state %v", err.Error()), "")
	}

	return base64.StdEncoding.EncodeToString(stateBin)
}

type Authenticator struct {
	*oidc.Provider
	oauth2.Config
}

var loginCmd = &cobra.Command{
	Use:   "login [url]",
	Short: "Authenticate on the Tigris instance",
	Long: `Performs authentication flow on the Tigris instance
* Run "tigris login [url]",
* It opens a login page in the browser
* Enter organization name in the prompt. Click "Continue" button.
* On the new page click "Continue with Google" or "Continue with Github",
  depending on which OIDC provider you prefer.
* You will be asked for you Google or Github password,
  if are not already signed in to the account
* You'll see "Successfully authenticated" on success
* You can now return to the terminal and start using the CLI`,
	Args:    cobra.MinimumNArgs(1),
	Example: `tigris login api.preview.tigrisdata.cloud`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		tmplSuccess, err = template.New("login_success").Parse(templates.LoginSuccessful)
		if err != nil {
			util.Error(err, "error parsing login success template")
		}

		tmplError, err = template.New("login_error").Parse(templates.LoginError)
		if err != nil {
			util.Error(err, "error parsing login error template")
		}

		state := genRandomState()

		host := args[0]
		inst, ok := instances[host]
		if !ok {
			util.Error(fmt.Errorf("instance not found: %s", host), "Instance config not found")
		}

		p, err := oidc.NewProvider(context.Background(), inst.authHost)
		if err != nil {
			util.Error(err, "authorize error")
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

		server := &http.Server{Addr: callbackHost}

		var wg1 sync.WaitGroup
		wg1.Add(1)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			callbackErr = callback(&wg1, server, auth, host, state)
		}()

		c := http.Client{Timeout: 5 * time.Second}

		log.Debug().Err(err).Msg("Waiting for callback server start")
		for {
			r, err := c.Get("http://" + callbackHost + "/ping")
			if err == nil && r.StatusCode == http.StatusOK {
				log.Debug().Msg("Callback server has started")
				break
			}
		}

		authorizeErr = authorize(auth, state, inst.audience)

		log.Debug().Msg("Waiting for login flow to finish in the browser")

		wg1.Wait()

		log.Debug().Msg("Shutting down callback server")

		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()

		err = server.Shutdown(ctx)
		if err != nil && err != http.ErrServerClosed {
			util.Error(err, "shutdown callback server failed")
		}

		log.Debug().Msg("Callback handler finished")

		wg.Wait()

		if authorizeErr != nil {
			util.Error(authorizeErr, "authorize error")
		}

		if callbackErr != nil {
			util.Error(callbackErr, "callback error")
		}
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from Tigris instance",
	Run: func(cmd *cobra.Command, args []string) {
		config.DefaultConfig.Token = ""
		config.DefaultConfig.ApplicationSecret = ""
		config.DefaultConfig.ApplicationID = ""
		config.DefaultConfig.URL = ""
		if err := config.Save(config.DefaultName, config.DefaultConfig); err != nil {
			util.Error(err, "Failure saving config")
		}

		util.Stdout("Successfully logged out\n")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
}
