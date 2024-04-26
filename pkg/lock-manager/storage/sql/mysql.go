package sql

import (
	"database/sql"

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
		return nil, err
	}

	s := &SQLBackend{
		db: db,
	}

	err = s.init()
	if err != nil {
		return nil, err
	}
	return s, nil
}
