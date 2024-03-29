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
	"io"
	"net"
	"os"
	"strings"
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
	"github.com/tigrisdata/tigris-cli/login"
	"github.com/tigrisdata/tigris-cli/util"
)

const (
	ImagePath     = "tigrisdata/tigris-local"
	ContainerName = "tigris-local-server"
)

var (
	ImageTag = "latest"

	follow     bool
	loginParam bool

	waitUpTimeout    = 30 * time.Second
	pingSleepTimeout = 10 * time.Millisecond
)

func getClient(ctx context.Context) *client.Client {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	util.Fatal(err, "creating docker client")

	ctx, cancel := util.GetContext(ctx)
	defer cancel()

	p, err := cli.Ping(ctx)
	if err != nil {
		home, err := os.UserHomeDir()
		util.Fatal(err, "getting home dir")

		cli, err = client.NewClientWithOpts(client.FromEnv,
			client.WithHost("unix://"+home+"/.colima/default/docker.sock"), client.WithAPIVersionNegotiation())
		util.Fatal(err, "retrying creating docker client")

		p, err = cli.Ping(ctx)
		util.Fatal(err, "pinging docker daemon")
	}

	log.Debug().Interface("version", p).Msg("docker version")

	return cli
}

func stopContainer(cl *client.Client, cname string) {
	ctx := context.Background()

	log.Debug().Msg("stopping local instance")

	if err := cl.ContainerStop(ctx, cname, container.StopOptions{}); err != nil {
		if !errdefs.IsNotFound(err) {
			util.Fatal(err, "error stopping container: %s", cname)
		}
	}

	opts := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	if err := cl.ContainerRemove(ctx, cname, opts); err != nil {
		if !errdefs.IsNotFound(err) {
			util.Fatal(err, "error stopping container: %s", cname)
		}
	}

	log.Debug().Msg("local instance stopped")
}

func pullImage(cli *client.Client, image string, port string) {
	ctx := context.Background()

	log.Debug().Str("port", port).Msg("starting local instance")

	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		util.Fatal(err, "error pulling docker image: %s", image)
	}

	log.Debug().Msg("local Docker image pulled")

	if util.IsTTY(os.Stdout) {
		if err := util.DockerShowProgress(reader); err != nil {
			_ = reader.Close()

			util.Fatal(err, "error pulling docker image: %s", image)
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
			util.Fatal(err, "error parsing port: %s", port)
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
		util.Fatal(err, "error creating container docker image: %s", image)
	}

	log.Debug().Msg("starting container")

	if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		util.Fatal(err, "starting container id=%s", resp.ID)
	}

	log.Debug().Msg("local instance started successfully")

	return resp.ID
}

func waitServerUp(port string) {
	log.Debug().Msg("waiting local instance to start")

	cfg := config.DefaultConfig
	cfg.URL = fmt.Sprintf("localhost:%s", port)
	cfg.Token = ""
	cfg.ClientSecret = ""
	cfg.ClientID = ""

	err := tclient.Init(&cfg)
	util.Fatal(err, "init tigris client")

	if err = pingLow(context.Background(), waitUpTimeout, pingSleepTimeout, true, true,
		util.IsTTY(os.Stdout) && !util.Quiet); err != nil {
		util.Fatal(err, "tigris initialization failed")
	}

	log.Debug().Msg("wait finished successfully")
}

var serverUpCmd = &cobra.Command{
	Use:     "start [port] [version]",
	Aliases: []string{"up"},
	Short:   "Starts an instance of Tigris for local development",
	Run: func(cmd *cobra.Command, args []string) {
		cli := getClient(cmd.Context())

		configPort := "8081"
		u := strings.Split(config.DefaultConfig.URL, ":")
		if len(u) == 2 {
			configPort = u[1]
		}

		var port string
		if len(args) > 0 {
			port = args[0]
		} else { // only using DefaultConfig when port not specified in command line
			port = configPort
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

		ctx, cancel := util.GetContext(cmd.Context())
		defer cancel()

		if config.DefaultConfig.Project != "" {
			util.Infof("Creating project: %s", config.DefaultConfig.Project)
			_, err := tclient.Get().CreateProject(ctx, config.DefaultConfig.Project)
			util.Fatal(err, "creating project on start")

			if config.DefaultConfig.Branch != "" && config.DefaultConfig.Branch != DefaultBranch {
				util.Infof("Creating branch: %s", config.DefaultConfig.Branch)
				_, err := tclient.Get().UseDatabase(config.DefaultConfig.Project).CreateBranch(ctx, config.DefaultConfig.Branch)
				util.Fatal(err, "creating branch on start")
			}
		}

		if loginParam {
			login.LocalLogin(net.JoinHostPort("localhost", port), "")
		} else if port != configPort { // inform user when specified ports are different
			util.Stdoutf("run 'export TIGRIS_URL=localhost:%s' for tigris cli to connect to the local instance\n", port)
		}
	},
}

var serverDownCmd = &cobra.Command{
	Use:     "stop",
	Aliases: []string{"down"},
	Short:   "Stops local Tigris instance",
	Run: func(cmd *cobra.Command, args []string) {
		cli := getClient(cmd.Context())

		stopContainer(cli, ContainerName)

		util.Stdoutf("Tigris stopped\n")
	},
}

var serverLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Shows logs from local Tigris instance",
	Run: func(cmd *cobra.Command, args []string) {
		cli := getClient(cmd.Context())

		logs, err := cli.ContainerLogs(cmd.Context(), ContainerName, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     follow,
		})
		if err != nil {
			util.Fatal(err, "error reading container logs")
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
	serverLogsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow logs output")
	localCmd.AddCommand(serverLogsCmd)

	serverUpCmd.Flags().BoolVarP(&loginParam, "login", "l", false, "login to the local instance after starting it")
	serverUpCmd.Flags().StringVarP(&config.DefaultConfig.Project, "create-project", "p", config.DefaultConfig.Project,
		"create project after start")
	serverUpCmd.Flags().StringVarP(&config.DefaultConfig.Branch, "create-branch", "b", config.DefaultConfig.Branch,
		"create database branch after start")
	localCmd.AddCommand(serverUpCmd)

	localCmd.AddCommand(serverDownCmd)
	rootCmd.AddCommand(localCmd)
}
