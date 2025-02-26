package k8s

type Config struct {
	Kubeconfig          string `json:"kubeconfig,omitempty"`
	DrainTimeoutSeconds int32  `json:"drainTimeoutSeconds,omitempty"`
	DrainRetries        int    `json:"drainRetries,omitempty"`
}

func NewDefaultConfig() Config {
	return Config{
		DrainTimeoutSeconds: 300,
	}
}
