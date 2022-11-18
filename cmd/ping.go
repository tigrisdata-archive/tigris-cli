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

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/util"
)

var pingTimeout time.Duration

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Checks connection to Tigris",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()

		err := client.InitLow()
		if err == nil {
			_, err = client.D.Health(ctx)
		}
		_ = util.Error(err, "ping")

		end := time.Now().Add(pingTimeout)
		sleep := 32 * time.Millisecond

		for err != nil && pingTimeout > 0 && time.Now().Add(sleep).Before(end) {
			_ = util.Error(err, "ping sleep %v", sleep)
			time.Sleep(sleep)

			if client.D != nil {
				_ = util.Error(client.D.Close(), "client close")
			}

			client.D = nil

			if err = client.InitLow(); err == nil {
				_, err = client.D.Health(ctx)
			}

			sleep *= 2

			if rem := time.Until(end); sleep > rem && rem > 0 {
				sleep = rem
				_ = util.Error(err, "ping sleep1 %v", sleep)
			}
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "FAILED\n")
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stderr, "OK\n")
		}
	},
}

func init() {
	pingCmd.Flags().DurationVarP(&pingTimeout, "timeout", "t", 0, "wait for ping to succeed for the specified timeout")

	dbCmd.AddCommand(pingCmd)
}
