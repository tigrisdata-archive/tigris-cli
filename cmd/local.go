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
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	volumetypes "github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	tclient "github.com/tigrisdata/tigrisdb-cli/client"
	"github.com/tigrisdata/tigrisdb-cli/config"
)

const (
	ImagePath        = "tigrisdata/tigrisdb"
	FDBImagePath     = "foundationdb/foundationdb:6.3.23"
	volumeName       = "fdbdata"
	FDBContainerName = "tigrisdb-local-fdb"
	ContainerName    = "tigrisdb-local-server"
)

var ImageTag = "1.0.0-alpha.6"

func ensureVolume(cli *client.Client) {
	ctx := context.Background()

	volumes, err := cli.VolumeList(ctx, filters.NewArgs())
	if err != nil {
		log.Fatal().Err(err).Msg("error listing volumes")
	}

	found := false
	for _, v := range volumes.Volumes {
		if v.Name == volumeName {
			found = true
		}
	}

	if !found {
		_, err := cli.VolumeCreate(ctx, volumetypes.VolumeCreateBody{
			Driver: "local",
			Name:   volumeName,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("error creating docker volume")
		}
	}
}

func stopContainer(client *client.Client, cname string) {
	ctx := context.Background()

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
}

func startContainer(cli *client.Client, cname string, image string, volumeMount string, port string) string {
	ctx := context.Background()

	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		log.Fatal().Err(err).Str("image", image).Msg("error pulling docker image")
	}
	defer func() { _ = reader.Close() }()

	_, _ = io.Copy(os.Stdout, reader)

	m := mount.Mount{
		Type:   mount.TypeVolume,
		Source: volumeName,
		Target: volumeMount,
	}

	p, err := nat.ParsePortSpec(port)
	if err != nil {
		log.Fatal().Err(err).Str("image", image).Msg("error parsing port")
	}

	pm := nat.PortMap{}

	for _, v := range p {
		pm[v.Port] = []nat.PortBinding{v.Binding}
	}

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image: image,
			Tty:   false,
		},
		&container.HostConfig{
			Mounts:       []mount.Mount{m},
			PortBindings: pm,
		}, nil, nil, cname)
	if err != nil {
		log.Fatal().Err(err).Str("image", image).Msg("error creating container docker image")
	}

	if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Fatal().Err(err).Str("image", image).Msg("error starting docker image")
	}

	return resp.ID
}

func execDockerCommand(cli *client.Client, container string, cmd []string) {
	ctx := context.Background()

	response, err := cli.ContainerExecCreate(ctx, container, types.ExecConfig{
		Cmd: cmd,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("error executing command in docker container")
	}

	execID := response.ID
	if execID == "" {
		log.Fatal().Msg("error executing command in docker container")
	}

	resp, err := cli.ContainerExecAttach(ctx, execID, types.ExecStartCheck{})
	if err != nil {
		log.Fatal().Err(err).Msg("error executing command in docker container")
	}
	defer resp.Close()
}

func waitServerUp() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inited := false
	var err error
	for {
		if !inited {
			if err = tclient.Init(config.DefaultConfig); err == nil {
				inited = true
			}
		} else {
			_, err = tclient.Get().ListDatabases(ctx)
			if err == nil {
				break
			}
		}

		time.Sleep(10 * time.Millisecond)

		select {
		case <-ctx.Done():
			break
		default:
		}
	}
	if err != nil {
		log.Fatal().Err(err).Msg("tigrisdb initialization failed")
	}
}

var serverUpCmd = &cobra.Command{
	Use:   "up [port] [version]",
	Short: "Start an instance of TigrisDB for local development",
	Run: func(cmd *cobra.Command, args []string) {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			log.Fatal().Err(err).Msg("error creating docker client")
		}

		ensureVolume(cli)

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

		stopContainer(cli, FDBContainerName)
		stopContainer(cli, ContainerName)

		_ = startContainer(cli, FDBContainerName, FDBImagePath, "/var/fdb", "4500")
		_ = startContainer(cli, ContainerName, ImagePath+":"+ImageTag, "/etc/foundationdb", rport)

		execDockerCommand(cli, ContainerName, []string{"fdbcli", "--exec", "configure new single memory"})

		waitServerUp()

		fmt.Printf("TigrisDB is running at localhost:%s\n", port)
		if port != "8081" {
			fmt.Printf("run 'export TIGRISDB_URL=localhost:%s' for tigrisdb-cli to automatically connect\n", port)
		}
	},
}

var serverDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop local TigrisDB instance",
	Run: func(cmd *cobra.Command, args []string) {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			log.Fatal().Err(err).Msg("error creating docker client")
		}

		stopContainer(cli, FDBContainerName)
		stopContainer(cli, ContainerName)

		fmt.Printf("TigrisDB stopped\n")
	},
}

var serverLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show logs of local TigrisDB instance",
	Run: func(cmd *cobra.Command, args []string) {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			log.Fatal().Err(err).Msg("error creating docker client")
		}

		ctx := context.Background()

		follow, err := cmd.Flags().GetBool("follow")
		if err != nil {
			log.Fatal().Err(err).Msg("error getting 'follow' flag")
		}

		logs, err := cli.ContainerLogs(ctx, ContainerName, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     follow,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("error reading container logs")
		}

		_, _ = stdcopy.StdCopy(os.Stdout, os.Stderr, logs)
	},
}

var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Starting and stopping local TigrisDB server",
}

func init() {
	serverLogsCmd.Flags().BoolP("follow", "f", false, "follow logs output")
	localCmd.AddCommand(serverLogsCmd)
	localCmd.AddCommand(serverUpCmd)
	localCmd.AddCommand(serverDownCmd)
	dbCmd.AddCommand(localCmd)
}
