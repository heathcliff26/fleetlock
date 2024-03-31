package server

type ErrorIncompleteSSlConfig struct{}

func (e ErrorIncompleteSSlConfig) Error() string {
	return "SSL is enabled but either key or certificate is missing"
}
