package utils

import "fmt"

type ErrorGetNamespace struct {
	path string
	err  error
}

func NewErrorGetNamespace(path string, err error) error {
	return &ErrorGetNamespace{
		path: path,
		err:  err,
	}
}

func (e *ErrorGetNamespace) Error() string {
	return fmt.Sprintf("Could not retrieve namespace from \"%s\": %v", e.path, e.err)
}
