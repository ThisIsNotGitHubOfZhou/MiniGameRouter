package api

import (
	"github.com/ThisIsNotGitHubOfZhou/MiniGameRouter/sdk/service"
)

type MiniClient struct {
	name       string
	id         string
	host       string
	port       string
	protocol   string
	metadata   string
	weight     int
	timeout    int
	healthport string
}

var _ service.RegisterService = (*MiniClient)(nil)

var _ service.DiscoverService = (*MiniClient)(nil)

var _ service.HealthCheckService = (*MiniClient)(nil)

func NewMiniClient(name, host, port, protocol, metadata string, weight, timeout int) *MiniClient {
	return &MiniClient{
		name:     name,
		host:     host,
		port:     port,
		protocol: protocol,
		metadata: metadata,
		weight:   weight,
		timeout:  timeout,
	}
}

func (c *MiniClient) Name() string {
	return c.name
}

func (c *MiniClient) ID() string {
	return c.id
}

func (c *MiniClient) Host() string {
	return c.host
}

func (c *MiniClient) Port() string {
	return c.port
}

func (c *MiniClient) Protocol() string {
	return c.protocol
}

func (c *MiniClient) Metadata() string {
	return c.metadata
}

func (c *MiniClient) Weight() int {
	return c.weight
}

func (c *MiniClient) Timeout() int {
	return c.timeout
}

func (c *MiniClient) HealthPort() string {
	return c.healthport
}
