package sql

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type SQLiteConfig struct {
	File string `yaml:"file"`
}

func NewSQLiteBackend(cfg *SQLiteConfig) (*SQLBackend, error) {
	db, err := sql.Open("sqlite", cfg.File)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	db.SetMaxOpenConns(1)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to \"ping\" sqlite database: %w", err)
	}

	s := &SQLBackend{
		databaseType: "sqlite",
		db:           db,
	}

	err = s.init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	return s, nil
}
