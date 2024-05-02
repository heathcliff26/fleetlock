package sql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/types"
	_ "modernc.org/sqlite"
)

const (
	stmtCreateTable = `CREATE TABLE IF NOT EXISTS locks (
	group_name VARCHAR(100) NOT NULL,
	id VARCHAR(100) NOT NULL,
	created TIMESTAMP NOT NULL,
	PRIMARY KEY (group_name,id)
	);`

	stmtReserve = `INSERT INTO locks (group_name, id, created)
		SELECT ?,?,?
		WHERE NOT EXISTS (
			SELECT 1 FROM locks WHERE group_name=? AND id=?
		);`

	stmtGetLocks = `SELECT COUNT(*) FROM (
			SELECT id FROM locks WHERE group_name=?
		) AS TMP;`

	stmtRelease = "DELETE FROM locks WHERE group_name=? AND id=?;"

	stmtHasLock = "SELECT 1 FROM locks WHERE group_name=? AND id=?;"
)

type SQLBackend struct {
	databaseType string

	db *sql.DB

	reserve  *sql.Stmt
	getLocks *sql.Stmt
	release  *sql.Stmt
	hasLock  *sql.Stmt
}

func (s *SQLBackend) init() error {
	var reserve, get, release, has string
	switch s.databaseType {
	case "postgres":
		reserve = postgresReserve
		get = postgresGetLocks
		release = postgresRelease
		has = postgresHasLock
	default:
		reserve = stmtReserve
		get = stmtGetLocks
		release = stmtRelease
		has = stmtHasLock
	}

	_, err := s.db.Exec(stmtCreateTable)
	if err != nil {
		return fmt.Errorf("failed to create lock table: %w", err)
	}

	s.reserve, err = s.db.Prepare(reserve)
	if err != nil {
		return fmt.Errorf("failed to prepare reserve statement: %w", err)
	}

	s.getLocks, err = s.db.Prepare(get)
	if err != nil {
		return fmt.Errorf("failed to prepare getLocks statement: %w", err)
	}

	s.release, err = s.db.Prepare(release)
	if err != nil {
		return fmt.Errorf("failed to prepare release statement: %w", err)
	}

	s.hasLock, err = s.db.Prepare(has)
	if err != nil {
		return fmt.Errorf("failed to prepare hasLock statement: %w", err)
	}

	return nil
}

// Reserve a lock for the given group.
// Returns true if the lock is successfully reserved, even if the lock is already held by the specific id
func (s *SQLBackend) Reserve(group string, id string) error {
	_, err := s.reserve.Exec(group, id, time.Now(), group, id)
	if err != nil {
		return fmt.Errorf("failed to reserve lock: %w", err)
	}

	return nil
}

// Returns the current number of locks for the given group
func (s *SQLBackend) GetLocks(group string) (int, error) {
	rows, err := s.getLocks.Query(group)
	if err != nil {
		return 0, fmt.Errorf("failed to run getLocks query: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		err = rows.Err()
		if err != nil {
			err = fmt.Errorf("failed to read rows from getLocks result: %w", err)
		}
		return 0, err
	}

	var count int
	err = rows.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to read next row from getLocks result: %w", err)
	}
	return count, nil
}

// Release the lock currently held by the id.
// Does not fail when no lock is held.
func (s *SQLBackend) Release(group string, id string) error {
	_, err := s.release.Exec(group, id)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	return nil
}

// Return all locks older than x
func (s *SQLBackend) GetStaleLocks(ts time.Duration) ([]types.Lock, error) {
	panic("TODO")
}

// Check if a given id already has a lock for this group
func (s *SQLBackend) HasLock(group, id string) (bool, error) {
	rows, err := s.hasLock.Query(group, id)
	if err != nil {
		return false, fmt.Errorf("failed to run hasLocks query: %w", err)
	}
	defer rows.Close()

	res := rows.Next()
	err = rows.Err()
	if err != nil {
		err = fmt.Errorf("failed to read rows from hasLocks result: %w", err)
	}
	return res, err
}

// Calls all necessary finalization if necessary
func (s *SQLBackend) Close() error {
	return s.db.Close()
}
