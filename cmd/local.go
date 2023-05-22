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
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
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
	ImagePath = "tigrisdata/tigris-local"

	ContainerName = "tigris-local-server"
)

var (
	ImageTag = "latest"

	follow     bool
	loginParam bool

	waitUpTimeout    = 30 * time.Second
	pingSleepTimeout = 20 * time.Millisecond

	skipAuth       bool
	tokenAdminAuth bool = util.IsWindows()

	defaultDataDir = "/var/lib/tigris/"
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

func startContainer(cli *client.Client, cname string, image string, port string, dataDir string, env []string) string {
	ctx := context.Background()

	log.Debug().Msg("starting local instance")

	if strings.Contains(image, "/") {
		pullImage(cli, image, port)
	}

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

	var mounts []mount.Mount

	if dataDir != "" {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: dataDir,
			Target: "/var/lib/tigris",
		})

		log.Debug().Str("host_dir", dataDir).Str("docker_dir", "/var/lib/tigris").Msg("adding bind mount")
	}

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Hostname: cname,
			Image:    image,
			Tty:      false,
			Env:      env,
		},
		&container.HostConfig{
			PortBindings: pm,
			Mounts:       mounts,
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

func waitServerUp(url string, waitAuth bool) {
	log.Debug().Str("url", url).Bool("waitAuth", waitAuth).Msg("waiting local instance to start")

	cfg := config.DefaultConfig
	cfg.URL = url

	if !waitAuth {
		cfg.Token = ""
		cfg.ClientSecret = ""
		cfg.ClientID = ""
	}

	err := tclient.Init(&cfg)
	util.Fatal(err, "init tigris client")

	if err = pingLow(context.Background(), waitUpTimeout, pingSleepTimeout, true, waitAuth,
		util.IsTTY(os.Stdout) && !util.Quiet); err != nil {
		util.Fatal(err, "tigris initialization failed")
	}

	log.Debug().Msg("wait finished successfully")
}

func configureEnv(initialized bool) []string {
	var env []string

	if !initialized {
		env = append(env, "TIGRIS_LOCAL_PERSISTENCE=1")

		cfg := &config.DefaultConfig

		if cfg.DataDir == "" {
			cfg.DataDir = defaultDataDir
		}

		err := os.MkdirAll(cfg.DataDir, 0o700)

		util.Fatal(err, "ensuring data directory")
	}

	if skipAuth {
		env = append(env, "TIGRIS_SKIP_LOCAL_AUTH=1")
	} else if !initialized {
		env = append(env, "TIGRIS_BOOTSTRAP_LOCAL_AUTH=1")
		if tokenAdminAuth {
			env = append(env, "TIGRIS_LOCAL_GENERATE_ADMIN_TOKEN=1")
		}
	}

	return env
}

func configureLocal(local bool) ([]string, bool) {
	var (
		env         []string
		initialized bool
	)

	dataDir := config.DefaultConfig.DataDir

	if dataDir != "" {
		_, err := os.Stat(dataDir + "/initialized")
		initialized = err == nil

		if initialized {
			log.Debug().Msg("data directory not empty, skip bootstrap")
		}
	}

	if local {
		env = configureEnv(initialized)
	} else {
		if initialized {
			util.Stdoutf("Warning: Local Tigris instance is initialized in provided data directory: %s"+
				", but start of ephemeral instance requested.\n Did you mean `tigris local up ...` instead?\n", dataDir)
			util.Stdoutf("Remove --data-dir parameter to avoid this warning when starting ephemeral instance.\n")
		}

		config.DefaultConfig.DataDir = ""
	}

	return env, initialized
}

func getPortAndTag(args []string) string {
	port := "8081"
	if len(args) > 0 {
		port = args[0]
	}

	if len(args) > 1 {
		t := args[1]
		if t[0] == 'v' {
			t = t[1:]
		}

		ImageTag = t
	}

	return port
}

func createProjectAndBranch(ctx context.Context) {
	cfg := config.DefaultConfig

	if cfg.Project != "" {
		util.Infof("Creating project: %s", cfg.Project)
		_, err := tclient.Get().CreateProject(ctx, cfg.Project)
		util.Fatal(err, "creating project on start")

		if cfg.Branch != "" && cfg.Branch != DefaultBranch {
			util.Infof("Creating branch: %s", cfg.Branch)
			_, err := tclient.Get().UseDatabase(cfg.Project).CreateBranch(ctx, cfg.Branch)
			util.Fatal(err, "creating branch on start")
		}
	}
}

func setupToken() {
	cfg := &config.DefaultConfig

	tokenFile := cfg.DataDir + "/user_admin_token.txt"

	for {
		_, err := os.Stat(tokenFile)
		if err == nil {
			break
		}
	}

	token, err := os.ReadFile(tokenFile)
	util.Fatal(err, "reading token file")

	cfg.Token = strings.Trim(string(token), "\n\r\t ")
	cfg.SkipLocalTLS = true
	cfg.Protocol = "http"
}

func localUp(cmd *cobra.Command, args []string, local bool) {
	cli := getClient(cmd.Context())

	env, initialized := configureLocal(local)

	dataDir := config.DefaultConfig.DataDir

	port := getPortAndTag(args)

	stopContainer(cli, ContainerName)
	_ = startContainer(cli, ContainerName, ImagePath+":"+ImageTag, port+":8081", dataDir, env)

	cfg := &config.DefaultConfig

	if local && !tokenAdminAuth && cfg.Token == "" {
		cfg.URL = dataDir + "/server/unix.sock"
	} else {
		cfg.URL = net.JoinHostPort("localhost", port)
	}

	waitServerUp(cfg.URL, false)

	if local && !skipAuth {
		if tokenAdminAuth {
			setupToken()
		}

		waitServerUp(cfg.URL, true)
	}

	util.Stdoutf("\nTigris is listening at %s\n", cfg.URL)

	if dataDir != "" {
		util.Stdoutf("Data directory: %s\n", dataDir)
	}

	ctx, cancel := util.GetContext(cmd.Context())
	defer cancel()

	createProjectAndBranch(ctx)

	if loginParam || (local && !initialized && !skipAuth) {
		login.LocalLogin(cfg.URL, cfg.Token)
	} else if cfg.URL != "localhost:8081" {
		util.Stdoutf("run 'export TIGRIS_URL=%s' for tigris cli to connect to the local instance\n", cfg.URL)
	}
}

var localUpCmd = &cobra.Command{
	Use:     "start [port] [version]",
	Aliases: []string{"up"},
	Short:   "Starts an instance of Tigris",
	Long:    "Start local instance of Tigris. The data is persisted between runs. Authentication is enabled by default",
	Run: func(cmd *cobra.Command, args []string) {
		localUp(cmd, args, true)
	},
}

var devUpCmd = &cobra.Command{
	Use:     "start [port] [version]",
	Aliases: []string{"up"},
	Short:   "Starts an instance of Tigris for local development",
	Long:    "Start and instance of Tigris for local development. The data is not persisted between runs",
	Run: func(cmd *cobra.Command, args []string) {
		localUp(cmd, args, false)
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
	Use:   "local",
	Short: "Starts and stops local persistent and authenticated Tigris server",
}

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Starts and stops local development Tigris server",
}

func init() {
	serverLogsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow logs output")
	localCmd.AddCommand(serverLogsCmd)

	devUpCmd.Flags().BoolVarP(&loginParam, "login", "l", false, "login to the local instance after starting it")

	devUpCmd.Flags().StringVarP(&config.DefaultConfig.Project, "create-project", "p", config.DefaultConfig.Project,
		"create project after start")
	devUpCmd.Flags().StringVarP(&config.DefaultConfig.Branch, "create-branch", "b", config.DefaultConfig.Branch,
		"create database branch after start")

	localUpCmd.Flags().StringVarP(&config.DefaultConfig.DataDir, "data-dir", "d", "",
		"Directory for data persistence")
	localUpCmd.Flags().BoolVar(&skipAuth, "skip-auth", false,
		"Start unauthenticated local instance")
	localUpCmd.Flags().BoolVar(&tokenAdminAuth, "token-admin-auth", tokenAdminAuth,
		"Use token instead of unix socket peer for instance administrator authentication")

	localCmd.AddCommand(localUpCmd)
	localCmd.AddCommand(serverDownCmd)

	devCmd.AddCommand(devUpCmd)
	devCmd.AddCommand(serverDownCmd)

	rootCmd.AddCommand(localCmd)
	rootCmd.AddCommand(devCmd)
}
