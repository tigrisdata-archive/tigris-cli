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
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/util"
)

var pingTimeout time.Duration

func pingLow(cmdCtx context.Context, timeout time.Duration, initSleep time.Duration, linear bool) error {
	ctx, cancel := util.GetContext(cmdCtx)

	err := client.InitLow()
	if err == nil {
		_, err = client.D.Health(ctx)
	}

	_ = util.Error(err, "ping")

	cancel()

	end := time.Now().Add(timeout)
	sleep := initSleep

	for err != nil && timeout > 0 && time.Now().Add(sleep).Before(end) {
		_ = util.Error(err, "ping sleep %v", sleep)
		time.Sleep(sleep)

		if client.D != nil {
			_ = util.Error(client.D.Close(), "client close")
		}

		client.D = nil

		ctx, cancel = util.GetContext(cmdCtx)

		if err = client.InitLow(); err == nil {
			_, err = client.D.Health(ctx)
		}

		cancel()

		if !linear {
			sleep *= 2
		}

		if rem := time.Until(end); sleep > rem && rem > 0 {
			sleep = rem
			_ = util.Error(err, "ping sleep1 %v", sleep)
		}
	}

	return err
}

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Checks connection to Tigris",
	Run: func(cmd *cobra.Command, args []string) {
		if err := pingLow(cmd.Context(), pingTimeout, 32*time.Millisecond, false); err != nil {
			fmt.Fprintf(os.Stderr, "FAILED\n")
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "OK\n")
	},
}

func init() {
	pingCmd.Flags().DurationVarP(&pingTimeout, "timeout", "t", 0, "wait for ping to succeed for the specified timeout")

	rootCmd.AddCommand(pingCmd)
}
