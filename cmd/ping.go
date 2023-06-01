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
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/util"
)

var pingTimeout time.Duration

func pingCall(ctx context.Context, waitAuth bool) error {
	var err error

	if waitAuth {
		_, err = client.D.ListProjects(ctx)
	} else {
		_, err = client.D.Health(ctx)
	}

	return err
}

func localURL(url string) bool {
	return strings.HasPrefix(url, "localhost:") ||
		strings.HasPrefix(url, "127.0.0.1:") ||
		strings.HasPrefix(url, "http://localhost:") ||
		strings.HasPrefix(url, "http://127.0.0.1:") ||
		strings.HasPrefix(url, "[::1]") ||
		strings.HasPrefix(url, "http://[::1]:")
}

func initPingProgressBar(init bool) *progressbar.ProgressBar {
	if !init {
		return nil
	}

	return progressbar.NewOptions64(
		-1,
		progressbar.OptionSetDescription("Waiting for OK response"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(10),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionOnCompletion(func() {
			_, _ = fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(int(rand.Int63()%76)), //nolint:gosec
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionFullWidth(),
	)
}

func pingLow(cmdCtx context.Context, timeout time.Duration, sleep time.Duration, linear bool, waitAuth bool,
	pgBar bool,
) error {
	ctx, cancel := util.GetContext(cmdCtx)

	err := client.InitLow()
	if err == nil {
		err = pingCall(ctx, waitAuth)
	}

	_ = util.Error(err, "ping")

	cancel()

	pb := initPingProgressBar(pgBar)

	end := time.Now().Add(timeout)

	for err != nil && timeout > 0 && time.Now().Add(sleep).Before(end) {
		_ = util.Error(err, "ping sleep %v", sleep)
		time.Sleep(sleep)

		if client.D != nil {
			_ = util.Error(client.D.Close(), "client close")
		}

		client.D = nil

		ctx, cancel = util.GetContext(cmdCtx)

		if err = client.InitLow(); err == nil {
			if err = pingCall(ctx, waitAuth); err == nil {
				break
			}
		}

		cancel()

		if !linear {
			sleep *= 2
		}

		if rem := time.Until(end); sleep > rem && rem > 0 {
			sleep = rem
			_ = util.Error(err, "ping sleep1 %v", sleep)
		}

		if pb != nil {
			_ = pb.Add(int(sleep / time.Millisecond))
		}
	}

	return err
}

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Checks connection to Tigris",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		_ = client.Init(&config.DefaultConfig)

		waitForAuth := localURL(config.DefaultConfig.URL) && (config.DefaultConfig.Token != "" ||
			config.DefaultConfig.ClientSecret != "")

		if err = pingLow(cmd.Context(), pingTimeout, 32*time.Millisecond, localURL(config.DefaultConfig.URL),
			waitForAuth, util.IsTTY(os.Stdout) && !util.Quiet); err == nil {
			_, _ = fmt.Fprintf(os.Stderr, "OK\n")
			return
		}

		_, _ = fmt.Fprintf(os.Stderr, "FAILED\n")
		os.Exit(1) //nolint:revive
	},
}

func init() {
	pingCmd.Flags().DurationVarP(&pingTimeout, "timeout", "t", 0, "wait for ping to succeed for the specified timeout")

	rootCmd.AddCommand(pingCmd)
}
