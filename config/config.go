package config

import (
	"github.com/caarlos0/env"
)

type (
	Config struct {
		GossipAddrs []string `env:"GOSSIP_ADDRS"`
		Port        string   `env:"PORT"`
		GossipPort  int      `env:"GOSSIP_PORT"`
	}
)

func Load() (*Config, error) {
	var out Config

	if err := env.Parse(&out); err != nil {
		return nil, err
	}

	return &out, nil
}
