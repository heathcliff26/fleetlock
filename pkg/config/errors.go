package config

type ErrUnknownLogLevel struct {
	level string
}

func NewErrUnknownLogLevel(level string) error {
	return &ErrUnknownLogLevel{
		level: level,
	}
}

func (e *ErrUnknownLogLevel) Error() string {
	return "Unknown log level " + e.level
}
