package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		address  string
		database string
		options  string
		expected string
	}{
		{
			name:     "AllFields",
			username: "user",
			password: "pass",
			address:  "localhost:3306",
			database: "testdb",
			options:  "charset=utf8mb4",
			expected: "user:pass@localhost:3306/testdb?charset=utf8mb4",
		},
		{
			name:     "NoUsername",
			username: "",
			password: "pass",
			address:  "localhost:3306",
			database: "testdb",
			options:  "charset=utf8mb4",
			expected: "localhost:3306/testdb?charset=utf8mb4",
		},
		{
			name:     "NoPassword",
			username: "user",
			password: "",
			address:  "localhost:3306",
			database: "testdb",
			options:  "charset=utf8mb4",
			expected: "user@localhost:3306/testdb?charset=utf8mb4",
		},
		{
			name:     "NoOptions",
			username: "user",
			password: "pass",
			address:  "localhost:3306",
			database: "testdb",
			options:  "",
			expected: "user:pass@localhost:3306/testdb",
		},
		{
			name:     "MinimalConnection",
			username: "",
			password: "",
			address:  "localhost:3306",
			database: "testdb",
			options:  "",
			expected: "localhost:3306/testdb",
		},
		{
			name:     "EmptyDatabase",
			username: "user",
			password: "pass",
			address:  "localhost:3306",
			database: "",
			options:  "",
			expected: "user:pass@localhost:3306/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createConnectionString(tt.username, tt.password, tt.address, tt.database, tt.options)
			assert.Equal(t, tt.expected, result)
		})
	}
}