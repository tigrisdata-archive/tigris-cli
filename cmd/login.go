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

package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/login"
	"github.com/tigrisdata/tigris-cli/util"
)

var loginCmd = &cobra.Command{
	Use:   "login {url}",
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
	Example: `
# Login to the hosted platform
tigris login api.preview.tigrisdata.cloud

# Point all subsequent commands to locally running instance
tigris login dev
`,
	Run: func(cmd *cobra.Command, args []string) {
		var host string

		if len(args) > 0 {
			host = args[0]
		}

		if err := login.CmdLow(cmd.Context(), host); err != nil {
			os.Exit(1)
		}
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from Tigris instance",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.DefaultConfig
		cfg.Token = ""
		cfg.ClientSecret = ""
		cfg.ClientID = ""
		cfg.URL = ""

		err := config.Save(config.DefaultName, cfg)
		util.Fatal(err, "saving config")

		util.Stderrf("Successfully logged out\n")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
}
