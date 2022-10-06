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
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	tclient "github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/util"
)

const (
	ImagePath     = "tigrisdata/tigris-local"
	ContainerName = "tigris-local-server"
)

var (
	ImageTag = "latest"

	ErrServerStartTimeout = fmt.Errorf("timeout waiting server to start")
)

func stopContainer(client *client.Client, cname string) {
	ctx := context.Background()

	log.Debug().Msg("stopping local instance")

	if err := client.ContainerStop(ctx, cname, nil); err != nil {
		if !errdefs.IsNotFound(err) {
			util.Error(err, "error stopping container: %s", cname)
		}
	}

	opts := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	if err := client.ContainerRemove(ctx, cname, opts); err != nil {
		if !errdefs.IsNotFound(err) {
			util.Error(err, "error stopping container: %s", cname)
		}
	}

	log.Debug().Msg("local instance stopped")
}

func pullImage(cli *client.Client, image string, port string) {
	ctx := context.Background()

	log.Debug().Str("port", port).Msg("starting local instance")

	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		util.Error(err, "error pulling docker image: %s", image)
	}

	log.Debug().Msg("local Docker image pulled")

	if util.IsTTY(os.Stdout) {
		if err := util.DockerShowProgress(reader); err != nil {
			_ = reader.Close()

			util.Error(err, "error pulling docker image: %s", image)
		}
	} else {
		_, _ = io.Copy(os.Stdout, reader)
	}

	_ = reader.Close()
}

func startContainer(cli *client.Client, cname string, image string, port string, env []string) string {
	ctx := context.Background()

	log.Debug().Msg("starting local instance")

	pullImage(cli, image, port)

	pm := nat.PortMap{}

	if port != "" {
		p, err := nat.ParsePortSpec(port)
		if err != nil {
			util.Error(err, "error parsing port: %s", port)
		}

		for _, v := range p {
			pm[v.Port] = []nat.PortBinding{v.Binding}
		}
	}

	log.Debug().Msg("creating container")

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Hostname: cname,
			Image:    image,
			Tty:      false,
			Env:      env,
		},
		&container.HostConfig{
			PortBindings: pm,
		}, nil, nil, cname)
	if err != nil {
		util.Error(err, "error creating container docker image: %s", image)
	}

	log.Debug().Msg("starting container")

	if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		util.Error(err, "starting container id=%s", resp.ID)
	}

	log.Debug().Msg("local instance started successfully")

	return resp.ID
}

func waitServerUp(port string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Debug().Msg("waiting local instance to start")

	inited := false

	cfg := config.DefaultConfig
	cfg.URL = fmt.Sprintf("localhost:%s", port)
	cfg.Token = ""
	cfg.ClientSecret = ""
	cfg.ClientID = ""

	if err := tclient.Init(&cfg); err != nil {
		util.Error(err, "client init failed")
	}

	var err error

L:
	for {
		if !inited {
			if err = tclient.InitLow(); err == nil {
				inited = true
			}
		} else {
			_, err = tclient.Get().ListDatabases(ctx)
			if err == nil {
				break
			}
		}

		if errors.Is(err, context.DeadlineExceeded) {
			break
		}

		time.Sleep(10 * time.Millisecond)

		select {
		case <-ctx.Done():
			err = ErrServerStartTimeout

			break L
		default:
		}
	}

	if err != nil {
		util.Error(err, "tigris initialization failed")
	}

	log.Debug().Msg("wait finished successfully")
}

var serverUpCmd = &cobra.Command{
	Use:     "start [port] [version]",
	Aliases: []string{"up"},
	Short:   "Starts an instance of Tigris for local development",
	Run: func(cmd *cobra.Command, args []string) {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			util.Error(err, "error creating docker client")
		}

		port := "8081"
		if len(args) > 0 {
			port = args[0]
		}
		rport := port + ":8081"

		if len(args) > 1 {
			t := args[1]
			if t[0] == 'v' {
				t = t[1:]
			}
			ImageTag = t
		}

		stopContainer(cli, ContainerName)
		_ = startContainer(cli, ContainerName, ImagePath+":"+ImageTag, rport, nil)

		waitServerUp(port)

		util.Stdoutf("Tigris is running at localhost:%s\n", port)
		if port != "8081" {
			util.Stdoutf("run 'export TIGRIS_URL=localhost:%s' for tigris cli to automatically connect\n", port)
		}
	},
}

var serverDownCmd = &cobra.Command{
	Use:     "stop",
	Aliases: []string{"down"},
	Short:   "Stops local Tigris instance",
	Run: func(cmd *cobra.Command, args []string) {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			util.Error(err, "error creating docker client")
		}

		stopContainer(cli, ContainerName)

		util.Stdoutf("Tigris stopped\n")
	},
}

var serverLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Shows logs from local Tigris instance",
	Run: func(cmd *cobra.Command, args []string) {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			util.Error(err, "error creating docker client")
		}

		ctx := context.Background()

		follow, err := cmd.Flags().GetBool("follow")
		if err != nil {
			util.Error(err, "error reading 'follow' option")
		}

		logs, err := cli.ContainerLogs(ctx, ContainerName, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     follow,
		})
		if err != nil {
			util.Error(err, "error reading container logs")
		}

		_, _ = stdcopy.StdCopy(os.Stdout, os.Stderr, logs)
	},
}

var localCmd = &cobra.Command{
	Use:     "dev",
	Aliases: []string{"local"},
	Short:   "Starts and stops local development Tigris server",
}

func init() {
	serverLogsCmd.Flags().BoolP("follow", "f", false, "follow logs output")
	localCmd.AddCommand(serverLogsCmd)
	localCmd.AddCommand(serverUpCmd)
	localCmd.AddCommand(serverDownCmd)
	dbCmd.AddCommand(localCmd)
}
