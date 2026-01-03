package server

import (
	"strconv"
)

const (
	DEFAULT_SERVER_PORT     = 8080
	DEFAULT_SERVER_PORT_SSL = 8443
)

type ServerConfig struct {
	Listen string    `yaml:"listen"`
	SSL    SSLConfig `yaml:"ssl,omitempty"`
}

type SSLConfig struct {
	Enabled bool   `yaml:"enabled,omitempty"`
	Cert    string `yaml:"cert,omitempty"`
	Key     string `yaml:"key,omitempty"`
}

// Create a default server config with
func NewDefaultServerConfig() *ServerConfig {
	return &ServerConfig{}
}

// Check if there are empty values that need to be replaced by default values
func (cfg *ServerConfig) Defaults() {
	if cfg.Listen == "" {
		if cfg.SSL.Enabled {
			cfg.Listen = ":" + strconv.Itoa(DEFAULT_SERVER_PORT_SSL)
		} else {
			cfg.Listen = ":" + strconv.Itoa(DEFAULT_SERVER_PORT)
		}
	}
}

// Validate Server config and set default listen addr if needed
func (cfg *ServerConfig) Validate() error {
	if cfg.SSL.Enabled {
		if cfg.SSL.Cert == "" || cfg.SSL.Key == "" {
			return ErrorIncompleteSSlConfig{}
		}
	}
	return nil
}
