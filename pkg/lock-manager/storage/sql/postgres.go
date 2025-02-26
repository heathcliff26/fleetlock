package sql

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	postgresReserve = `INSERT INTO locks (group_name, id, created)
		SELECT $1,$2,$3
		WHERE NOT EXISTS (
			SELECT 1 FROM locks WHERE group_name=$4 AND id=$5
		);`

	postgresGetLocks = `SELECT COUNT(*) FROM (
			SELECT id FROM locks WHERE group_name=$1
		) AS TMP;`

	postgresRelease = "DELETE FROM locks WHERE group_name=$1 AND id=$2;"

	postgresHasLock = "SELECT 1 FROM locks WHERE group_name=$1 AND id=$2;"
)

type PostgresConfig struct {
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Options  string `json:"options,omitempty"`
}

func NewPostgresBackend(cfg PostgresConfig) (*SQLBackend, error) {
	connStr := createConnectionString(cfg.Username, cfg.Password, cfg.Address, cfg.Database, cfg.Options)

	db, err := sql.Open("pgx", "postgres://"+connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping postgres database: %w", err)
	}

	s := &SQLBackend{
		databaseType: "postgres",
		db:           db,
	}

	err = s.init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	return s, nil
}
