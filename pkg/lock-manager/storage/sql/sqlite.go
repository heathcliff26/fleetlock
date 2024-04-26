package sql

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type SQLiteConfig struct {
	File string `yaml:"file"`
}

func NewSQLiteBackend(cfg *SQLiteConfig) (*SQLBackend, error) {
	db, err := sql.Open("sqlite", cfg.File)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	s := &SQLBackend{
		db: db,
	}

	err = s.init()
	if err != nil {
		return nil, err
	}
	return s, nil
}
