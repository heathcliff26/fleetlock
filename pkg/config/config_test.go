package config

import (
	"errors"
	"log/slog"
	"reflect"
	"testing"

	"github.com/heathcliff26/fleetlock/pkg/k8s"
	lockmanager "github.com/heathcliff26/fleetlock/pkg/lock-manager"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/kubernetes"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/sql"
	"github.com/heathcliff26/fleetlock/pkg/lock-manager/storage/valkey"
	"github.com/heathcliff26/fleetlock/pkg/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidConfigs(t *testing.T) {
	c1 := &Config{
		LogLevel:         "debug",
		KubernetesConfig: k8s.NewDefaultConfig(),
		Server: &server.ServerConfig{
			Listen: "127.0.0.1:8443",
			SSL: server.SSLConfig{
				Enabled: true,
				Cert:    "foo.crt",
				Key:     "foo.key",
			},
		},
		Storage: lockmanager.StorageConfig{
			Type: "sqlite",
			SQLite: sql.SQLiteConfig{
				File: "foo.db",
			},
		},
		Groups: lockmanager.Groups{
			"default": lockmanager.GroupConfig{
				Slots: 5,
			},
			"foo": lockmanager.GroupConfig{
				Slots: 3,
			},
			"bar": lockmanager.GroupConfig{
				Slots: 10,
			},
		},
	}

	c2 := DefaultConfig()
	c2.Defaults()
	c2.LogLevel = "error"
	c2.Server.Listen = "[::1]:1234"

	defaultConfig := DefaultConfig()
	defaultConfig.Defaults()

	cfgKubernetes := DefaultConfig()
	cfgKubernetes.Defaults()
	cfgKubernetes.Storage = lockmanager.StorageConfig{
		Type: "kubernetes",
		Kubernetes: kubernetes.KubernetesConfig{
			Namespace:  "test",
			Kubeconfig: "some-path",
		},
	}
	cfgKubernetes.KubernetesConfig.Kubeconfig = "some-path"
	cfgKubernetes.KubernetesConfig.DrainTimeoutSeconds = 60
	cfgKubernetes.KubernetesConfig.DrainRetries = 3

	tMatrix := []struct {
		Name, Path string
		Result     *Config
	}{
		{
			Name:   "ValidConfig1",
			Path:   "testdata/valid-config-1.yaml",
			Result: c1,
		},
		{
			Name:   "ValidConfig2",
			Path:   "testdata/valid-config-2.yaml",
			Result: c2,
		},
		{
			Name:   "EmptyGroups",
			Path:   "testdata/empty-groups.yaml",
			Result: defaultConfig,
		},
		{
			Name:   "NoConfig",
			Path:   "",
			Result: defaultConfig,
		},
		{
			Name:   "ValidConfigKubernetes",
			Path:   "testdata/valid-config-kubernetes.yaml",
			Result: cfgKubernetes,
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			c, err := LoadConfig(tCase.Path, false)

			require.Nil(t, err, "Should load config")
			assert.Equal(t, tCase.Result, c, "Config should match expected result")
		})
	}
}

func TestSetLogLevel(t *testing.T) {
	tMatrix := []struct {
		Name  string
		Level slog.Level
		Error error
	}{
		{"debug", slog.LevelDebug, nil},
		{"info", slog.LevelInfo, nil},
		{"warn", slog.LevelWarn, nil},
		{"error", slog.LevelError, nil},
		{"DEBUG", slog.LevelDebug, nil},
		{"INFO", slog.LevelInfo, nil},
		{"WARN", slog.LevelWarn, nil},
		{"ERROR", slog.LevelError, nil},
		{"Unknown", 0, &ErrUnknownLogLevel{"Unknown"}},
	}
	t.Cleanup(func() {
		err := setLogLevel(DEFAULT_LOG_LEVEL)
		if err != nil {
			t.Fatalf("Failed to cleanup after test: %v", err)
		}
	})

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			err := setLogLevel(tCase.Name)

			require.Equal(t, tCase.Error, err, "Error should match expected result")
			if err == nil {
				assert.Equal(t, tCase.Level, logLevel.Level())
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	c := DefaultConfig()

	assert.Equal(t, DEFAULT_LOG_LEVEL, c.LogLevel)
}

func TestInvalidConfigs(t *testing.T) {
	tMatrix := []struct {
		Name, Path string
		Result     string
	}{
		{
			Name:   "InvalidPath",
			Path:   "not-a-valid-file.yaml",
			Result: "*fs.PathError",
		},
		{
			Name:   "NotYaml",
			Path:   "testdata/invalid-1.yaml",
			Result: "*fmt.wrapError",
		},
		{
			Name:   "InvalidLogLevel",
			Path:   "testdata/invalid-2.yaml",
			Result: "*config.ErrUnknownLogLevel",
		},
		{
			Name:   "InvalidServerConfig",
			Path:   "testdata/invalid-3.yaml",
			Result: "server.ErrorIncompleteSSlConfig",
		},
		{
			Name:   "InvalidGroups",
			Path:   "testdata/invalid-4.yaml",
			Result: "errors.ErrorGroupSlotsOutOfRange",
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			cfg, err := LoadConfig(tCase.Path, false)

			assert := assert.New(t)

			assert.Nil(cfg)
			assert.Equal(tCase.Result, reflect.TypeOf(err).String())
		})
	}
}

func TestEnvSubstitution(t *testing.T) {
	c := DefaultConfig()
	c.Storage.Type = "valkey"
	c.Storage.Valkey = valkey.ValkeyConfig{
		Addrs:    []string{"localhost:4321"},
		Username: "valkey",
		Password: "testpass",
		DB:       5,
	}
	c.Defaults()

	t.Setenv("FLEETLOCK_VALKEY_USERNAME", "valkey")
	t.Setenv("FLEETLOCK_VALKEY_PASSWORD", "testpass")

	cfg, err := LoadConfig("testdata/env-substitution.yaml", true)

	assert := assert.New(t)

	assert.Nil(err)
	assert.Equal(c, cfg)
}

func TestErrUnknownLogLevel(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		expectedMsg string
	}{
		{
			name:        "InvalidDebugLevel",
			level:       "invalid",
			expectedMsg: "Unknown log level invalid",
		},
		{
			name:        "EmptyLevel",
			level:       "",
			expectedMsg: "Unknown log level ",
		},
		{
			name:        "NumericLevel",
			level:       "123",
			expectedMsg: "Unknown log level 123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewErrUnknownLogLevel(tt.level)
			assert.Error(t, err)
			assert.Equal(t, tt.expectedMsg, err.Error())
			
			// Test type assertion
			var unknownLogLevelErr *ErrUnknownLogLevel
			assert.True(t, errors.As(err, &unknownLogLevelErr))
		})
	}
}
