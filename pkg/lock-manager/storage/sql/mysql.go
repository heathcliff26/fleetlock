package sql

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLConfig struct {
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Options  string `yaml:"options,omitempty"`
}

func NewMySQLBackend(cfg *MySQLConfig) (*SQLBackend, error) {
	connStr := createConnectionString(cfg.Username, cfg.Password, cfg.Address, cfg.Database, cfg.Options)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping mysql database: %w", err)
	}

	s := &SQLBackend{
		databaseType: "mysql",
		db:           db,
	}

	err = s.init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	return s, nil
}