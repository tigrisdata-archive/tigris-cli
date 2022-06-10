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
	tclient "github.com/tigrisdata/tigris-cli/client"
	"github.com/tigrisdata/tigris-cli/config"
	"github.com/tigrisdata/tigris-cli/util"
)

const (
	ImagePath       = "tigrisdata/tigris"
	FDBImagePath    = "tigrisdata/foundationdb:7.1.7"
	SearchImagePath = "typesense/typesense:0.23.0"

	volumeName  = "fdbdata"
	networkName = "tigris_cli_network"

	SearchContainerName = "tigris-local-search"
	FDBContainerName    = "tigris-local-db"
	ContainerName       = "tigris-local-server"
)

var ImageTag = "latest"

var timeout = 10 * time.Second

func ensureVolume(cli *client.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	volumes, err := cli.VolumeList(ctx, filters.NewArgs())
	if err != nil {
		util.Error(err, "error listing volumes")
	}

	for _, v := range volumes.Volumes {
		if v.Name == volumeName {
			return
		}
	}

	_, err = cli.VolumeCreate(ctx, volumetypes.VolumeCreateBody{
		Driver: "local",
		Name:   volumeName,
	})
	if err != nil {
		util.Error(err, "error creating docker volume")
	}
}

func ensureNetwork(cli *client.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{Filters: filters.NewArgs()})
	if err != nil {
		util.Error(err, "docker network list failed")
	}

	for _, v := range networks {
		if v.Name == networkName {
			return
		}
	}

	_, err = cli.NetworkCreate(ctx, networkName, types.NetworkCreate{})
	if err != nil {
		util.Error(err, "docker network create failed")
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

func startContainer(cli *client.Client, cname string, image string, volumeMount string, port string, env []string) string {
	ctx := context.Background()

	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		log.Fatal().Err(err).Str("image", image).Msg("error pulling docker image")
	}
	defer func() { _ = reader.Close() }()

	if util.IsTTY(os.Stdout) {
		if err := util.DockerShowProgress(reader); err != nil {
			util.Error(err, "error pulling docker image")
		}
	} else {
		_, _ = io.Copy(os.Stdout, reader)
	}

	var m []mount.Mount

	if volumeMount != "" {
		m = append(m, mount.Mount{
			Type:   mount.TypeVolume,
			Source: volumeName,
			Target: volumeMount,
		})
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

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Hostname: cname,
			Image:    image,
			Tty:      false,
			Env:      env,
		},
		&container.HostConfig{
			Mounts:       m,
			PortBindings: pm,
			DNS:          []string{"127.0.0.11"},
			//DNS: []string{"172.17.0.1"},
			NetworkMode: networkName,
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := cli.ContainerExecCreate(ctx, container, types.ExecConfig{
		Cmd: cmd,
	})
	if err != nil {
		util.Error(err, "error executing command in docker container")
	}

	execID := response.ID
	if execID == "" {
		log.Fatal().Msg("error executing command in docker container")
	}

	resp, err := cli.ContainerExecAttach(ctx, execID, types.ExecStartCheck{})
	if err != nil {
		util.Error(err, "error executing command in docker container")
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
		util.Error(err, "tigris initialization failed")
	}
}

var serverUpCmd = &cobra.Command{
	Use:   "up [port] [version]",
	Short: "Starts an instance of Tigris for local development",
	Run: func(cmd *cobra.Command, args []string) {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			util.Error(err, "error creating docker client")
		}

		ensureVolume(cli)
		ensureNetwork(cli)

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

		stopContainer(cli, SearchContainerName)
		stopContainer(cli, FDBContainerName)
		stopContainer(cli, ContainerName)

		_ = startContainer(cli, SearchContainerName, SearchImagePath, "", "", []string{"TYPESENSE_API_KEY=ts_dev_key", "TYPESENSE_DATA_DIR=/tmp"})
		_ = startContainer(cli, FDBContainerName, FDBImagePath, "/var/lib/foundationdb", "", nil)
		_ = startContainer(cli, ContainerName, ImagePath+":"+ImageTag, "/etc/foundationdb", rport,
			[]string{
				"TIGRIS_SERVER_SEARCH_AUTH_KEY=ts_dev_key",
				fmt.Sprintf("TIGRIS_SERVER_SEARCH_HOST=%s", SearchContainerName),
			})

		execDockerCommand(cli, ContainerName, []string{"fdbcli", "--exec", "configure new single memory"})

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
		stopContainer(cli, FDBContainerName)
		stopContainer(cli, SearchContainerName)

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
