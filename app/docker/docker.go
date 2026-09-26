// Package docker provides Docker client.
package docker

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"net/http"
	"path/filepath"

	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/tlsconfig"
)

// Client defines Docker client.
type Client struct {
	*docker.Client
}

// ContainerJSONList returns the list of the container information.
func (c *Client) ContainerJSONList(ctx context.Context) ([]*container.InspectResponse, error) {
	containers, err := c.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	res := make([]*container.InspectResponse, 0, len(containers))

	for _, cont := range containers {
		ci, _, err := c.ContainerInspectWithRaw(ctx, cont.ID, true)
		if err != nil {
			return nil, err
		}
		res = append(res, &ci)
	}
	return res, nil
}

// NewClient creates a new Docker client.
func NewClient(ctx context.Context, host, certPath, version string, verify bool) (*Client, error) {
	cli, err := docker.NewClientWithOpts(func(c *docker.Client) error {
		return setOpts(c, host, certPath, version, verify)
	})

	if err != nil {
		return nil, err
	}

	cli.NegotiateAPIVersion(ctx)
	return &Client{cli}, nil
}

func setOpts(c *docker.Client, host, certPath, version string, verify bool) error {
	if certPath != "" {
		options := tlsconfig.Options{
			CAFile:             filepath.Join(certPath, "ca.pem"),
			CertFile:           filepath.Join(certPath, "cert.pem"),
			KeyFile:            filepath.Join(certPath, "key.pem"),
			InsecureSkipVerify: !verify,
		}
		tlsc, err := tlsconfig.Client(options)
		if err != nil {
			return err
		}

		httpClient := &http.Client{
			Transport:     &http.Transport{TLSClientConfig: tlsc},
			CheckRedirect: docker.CheckRedirect,
		}

		if err := docker.WithHTTPClient(httpClient)(c); err != nil {
			return err
		}
	}

	if host != "" {
		if err := docker.WithHost(host)(c); err != nil {
			return err
		}
	}

	if version != "" {
		if err := docker.WithVersion(version)(c); err != nil {
			return err
		}
	}
	return nil
}
