package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env"
)

type (
	// The Config type contains the configuration values used by the node.
	Config struct {
		GossipHosts []string `env:"GOSSIP_HOSTS"`
		Port        string   `env:"PORT"`
		GossipPort  int      `env:"GOSSIP_PORT"`
	}
)

// Load attempts to load the configuration of the node from the environment.
// It also sanitizes the GossipHosts field to ensure we don't contain this
// node's address.
func Load() (*Config, error) {
	var out Config

	if err := env.Parse(&out); err != nil {
		return nil, err
	}

	host, err := os.Hostname()

	if err != nil {
		return nil, err
	}

	var addrs []string
	for _, addr := range out.GossipHosts {
		if strings.Contains(addr, host) {
			continue
		}

		addrs = append(addrs, fmt.Sprintf("%s:%v", addr, out.GossipPort))
	}

	out.GossipHosts = addrs

	return &out, nil
}
