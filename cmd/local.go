package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	volumetypes "github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const (
	ImagePath        = "tigrisdata/tigrisdb"
	FDBImagePath     = "foundationdb/foundationdb:6.3.23"
	volumeName       = "fdbdata"
	FDBContainerName = "tigrisdb-cli-fdb"
	ContainerName    = "tigrisdb-cli-server"
)

var ImageTag = "1.0.0-alpha.4"

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

var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Starting and stopping local TigrisDB server",
}

func init() {
	rootCmd.AddCommand(localCmd)
	localCmd.AddCommand(serverUpCmd)
	localCmd.AddCommand(serverDownCmd)
}
