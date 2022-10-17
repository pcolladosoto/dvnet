package dvnet

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type containerInfo struct {
	ID  string
	PID int
}

var dockerCli *client.Client

func runContainer(img, name string) (string, int, error) {
	ctx := context.Background()
	resp, err := dockerCli.ContainerCreate(ctx, &container.Config{
		Image:    img,
		Hostname: name,
	},
		&container.HostConfig{
			NetworkMode: "none",
			Sysctls: map[string]string{
				"net.ipv4.ip_forward":                "1",
				"net.ipv6.conf.all.disable_ipv6":     "0",
				"net.bridge.bridge-nf-call-iptables": "0",
			},
			CapAdd: []string{"SYS_ADMIN", "NET_ADMIN"},
			DNS:    []string{"1.1.1.1", "8.8.8.8"},
		},
		nil,
		nil,
		name,
	)

	if err != nil {
		// log.error("couldn't create container %s: %v\n", name, err)
		return "", 0, err
	}

	if err := dockerCli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", 0, err
	}

	containerInfo, err := dockerCli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return "", 0, err
	}

	return resp.ID, containerInfo.State.Pid, nil
}

func removeContainer(id string) error {
	ctx := context.Background()
	dockerCli.ContainerStop(ctx, id, nil)
	return dockerCli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{})
}
