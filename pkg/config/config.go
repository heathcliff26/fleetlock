package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/heathcliff26/fleetlock/pkg/k8s"
	lockmanager "github.com/heathcliff26/fleetlock/pkg/lock-manager"
	"github.com/heathcliff26/fleetlock/pkg/server"
	"go.yaml.in/yaml/v3"
)

const (
	DEFAULT_LOG_LEVEL = "info"
)

var logLevel = &slog.LevelVar{}

// Initialize the logger
func init() {
	logLevel = &slog.LevelVar{}
	opts := slog.HandlerOptions{
		Level: logLevel,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	slog.SetDefault(logger)
}

type Config struct {
	LogLevel         string                    `yaml:"logLevel,omitempty"`
	KubernetesConfig k8s.Config                `yaml:"kubernetes,omitempty"`
	Server           *server.ServerConfig      `yaml:"server,omitempty"`
	Storage          lockmanager.StorageConfig `yaml:"storage,omitempty"`
	Groups           lockmanager.Groups        `yaml:"groups,omitempty"`
}

// Parse a given string and set the resulting log level
func setLogLevel(level string) error {
	switch strings.ToLower(level) {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "info":
		logLevel.Set(slog.LevelInfo)
	case "warn":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	default:
		return NewErrUnknownLogLevel(level)
	}
	return nil
}

func DefaultConfig() *Config {
	return &Config{
		LogLevel:         DEFAULT_LOG_LEVEL,
		KubernetesConfig: k8s.NewDefaultConfig(),
		Server:           server.NewDefaultServerConfig(),
		Storage:          lockmanager.NewDefaultStorageConfig(),
	}
}

// Loads the config from the given path.
// When path is empty, returns default config.
// Returns error when the given config is invalid.
func LoadConfig(path string, env bool) (*Config, error) {
	c := DefaultConfig()

	if path == "" {
		c.Defaults()
		return c, nil
	}

	// #nosec G304 -- Local users can decide on their file path themselves.
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if env {
		f = []byte(os.ExpandEnv(string(f)))
	}

	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return nil, err
	}

	c.Defaults()

	err = c.Validate()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) Defaults() {
	c.Server.Defaults()
	if c.Groups == nil {
		c.Groups = lockmanager.NewDefaultGroups()
	}

	c.Storage.Kubernetes.Kubeconfig = c.KubernetesConfig.Kubeconfig
}

func (c *Config) Validate() error {
	err := setLogLevel(c.LogLevel)
	if err != nil {
		return err
	}

	err = c.Server.Validate()
	if err != nil {
		return err
	}

	err = c.Groups.Validate()
	if err != nil {
		return err
	}

	return nil
}
