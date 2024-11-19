package k8s

type Config struct {
	Kubeconfig          string `yaml:"kubeconfig,omitempty"`
	DrainTimeoutSeconds int32  `yaml:"drainTimeoutSeconds,omitempty"`
}

func NewDefaultConfig() Config {
	return Config{
		DrainTimeoutSeconds: 300,
	}
}
