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

var ImageTag = "latest"

func stopContainer(client *client.Client, cname string) {
	ctx := context.Background()

	log.Debug().Msg("stopping local instance")

	if err := client.ContainerStop(ctx, cname, nil); err != nil {
		if !errdefs.IsNotFound(err) {
			log.Fatal().Err(err).Str("name", cname).Msg("error stopping container")
		}
	}

	opts := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	if err := client.ContainerRemove(ctx, cname, opts); err != nil {
		if !errdefs.IsNotFound(err) {
			log.Fatal().Err(err).Str("name", cname).Msg("error stopping container")
		}
	}

	log.Debug().Msg("local instance stopped")
}

func startContainer(cli *client.Client, cname string, image string, port string, env []string) string {
	ctx := context.Background()

	log.Debug().Msg("starting local instance")

	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		log.Fatal().Err(err).Str("image", image).Msg("error pulling docker image")
	}
	defer func() { _ = reader.Close() }()

	log.Debug().Msg("local Docker image pulled")

	if util.IsTTY(os.Stdout) {
		if err := util.DockerShowProgress(reader); err != nil {
			util.Error(err, "error pulling docker image")
		}
	} else {
		_, _ = io.Copy(os.Stdout, reader)
	}

	pm := nat.PortMap{}

	if port != "" {
		p, err := nat.ParsePortSpec(port)
		if err != nil {
			log.Fatal().Err(err).Str("image", image).Msg("error parsing port")
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
		log.Fatal().Err(err).Str("image", image).Msg("error creating container docker image")
	}

	log.Debug().Msg("starting container")

	if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Fatal().Err(err).Str("image", image).Msg("error starting docker image")
	}

	log.Debug().Msg("local instance started successfully")

	return resp.ID
}

func waitServerUp() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Debug().Msg("waiting local instance to start")

	inited := false
	var err error

	if err = tclient.Init(config.DefaultConfig); err != nil {
		util.Error(err, "client init failed")
	}

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

		if err == context.DeadlineExceeded {
			break
		}

		time.Sleep(10 * time.Millisecond)

		select {
		case <-ctx.Done():
			err = fmt.Errorf("timeout waiting server to start")
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
	Use:   "up [port] [version]",
	Short: "Starts an instance of Tigris for local development",
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

		waitServerUp()

		fmt.Printf("Tigris is running at localhost:%s\n", port)
		if port != "8081" {
			fmt.Printf("run 'export TIGRIS_URL=localhost:%s' for tigris cli to automatically connect\n", port)
		}
	},
}

var serverDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stops local Tigris instance",
	Run: func(cmd *cobra.Command, args []string) {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			util.Error(err, "error creating docker client")
		}

		stopContainer(cli, ContainerName)

		fmt.Printf("Tigris stopped\n")
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
	Use:   "local",
	Short: "Starts and stops local Tigris server",
}

func init() {
	serverLogsCmd.Flags().BoolP("follow", "f", false, "follow logs output")
	localCmd.AddCommand(serverLogsCmd)
	localCmd.AddCommand(serverUpCmd)
	localCmd.AddCommand(serverDownCmd)
	dbCmd.AddCommand(localCmd)
}
