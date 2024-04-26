package sql

import (
	"database/sql"

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
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Options  string `yaml:"options,omitempty"`
}

func NewPostgresBackend(cfg *PostgresConfig) (*SQLBackend, error) {
	connStr := createConnectionString(cfg.Username, cfg.Password, cfg.Address, cfg.Database, cfg.Options)

	db, err := sql.Open("pgx", "postgres://"+connStr)
	if err != nil {
		return nil, err
	}

	s := &SQLBackend{
		db: db,
	}

	_, err = s.db.Exec(stmtCreateTable)
	if err != nil {
		return nil, err
	}

	s.reserve, err = s.db.Prepare(postgresReserve)
	if err != nil {
		return nil, err
	}

	s.getLocks, err = s.db.Prepare(postgresGetLocks)
	if err != nil {
		return nil, err
	}

	s.release, err = s.db.Prepare(postgresRelease)
	if err != nil {
		return nil, err
	}

	s.hasLock, err = s.db.Prepare(postgresHasLock)
	if err != nil {
		return nil, err
	}
	return s, nil
}
