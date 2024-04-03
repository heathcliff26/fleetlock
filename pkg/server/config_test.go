package server

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaults(t *testing.T) {
	t.Run("NewDefaultServerConfig", func(t *testing.T) {
		cfg := NewDefaultServerConfig()

		assert.Equal(t, &ServerConfig{}, cfg)
	})

	tMatrix := []struct {
		Name   string
		Config *ServerConfig
		Result *ServerConfig
	}{
		{
			Name: "SSLEnabled",
			Config: &ServerConfig{
				Listen: "",
				SSL: SSLConfig{
					Enabled: true,
				},
			},
			Result: &ServerConfig{
				Listen: ":" + strconv.Itoa(DEFAULT_SERVER_PORT_SSL),
				SSL: SSLConfig{
					Enabled: true,
				},
			},
		},
		{
			Name: "SSLDisabled",
			Config: &ServerConfig{
				Listen: "",
				SSL: SSLConfig{
					Enabled: false,
				},
			},
			Result: &ServerConfig{
				Listen: ":" + strconv.Itoa(DEFAULT_SERVER_PORT),
				SSL: SSLConfig{
					Enabled: false,
				},
			},
		},
		{
			Name: "CustomListen",
			Config: &ServerConfig{
				Listen: ":1234",
			},
			Result: &ServerConfig{
				Listen: ":1234",
			},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			tCase.Config.Defaults()
			assert.Equal(t, tCase.Result, tCase.Config)
		})
	}
}

func TestValidate(t *testing.T) {
	tMatrix := []struct {
		Name   string
		Config *ServerConfig
		Result error
	}{
		{
			Name: "SSLDisabled",
			Config: &ServerConfig{
				Listen: ":1234",
			},
			Result: nil,
		},
		{
			Name: "SSLValid",
			Config: &ServerConfig{
				SSL: SSLConfig{
					Enabled: true,
					Cert:    "foo.crt",
					Key:     "bar.key",
				},
			},
			Result: nil,
		},
		{
			Name: "SSLMissingKey",
			Config: &ServerConfig{
				SSL: SSLConfig{
					Enabled: true,
					Cert:    "foo.crt",
				},
			},
			Result: ErrorIncompleteSSlConfig{},
		},
		{
			Name: "SSLMissingCert",
			Config: &ServerConfig{
				SSL: SSLConfig{
					Enabled: true,
					Key:     "bar.crt",
				},
			},
			Result: ErrorIncompleteSSlConfig{},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			err := tCase.Config.Validate()

			assert.Equal(t, tCase.Result, err)
		})
	}
}
