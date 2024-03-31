package sqlite

import (
	"database/sql"
	"time"

	"github.com/heathcliff26/fleetlock/pkg/lock-manager/types"
	_ "github.com/mattn/go-sqlite3"
)

const (
	stmtCreateTable = `CREATE TABLE IF NOT EXISTS locks (
	group_name TEXT NOT NULL,
	id TEXT NOT NULL,
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
		);`

	stmtRelease = "DELETE FROM locks WHERE group_name=? AND id=?;"

	stmtHasLock = "SELECT 1 FROM locks WHERE group_name=? AND id=?;"
)

type SQLBackend struct {
	db *sql.DB

	reserve  *sql.Stmt
	getLocks *sql.Stmt
	release  *sql.Stmt
	hasLock  *sql.Stmt
}

type SQLiteConfig struct {
	File string `yaml:"file"`
}

func NewSQLiteBackend(cfg *SQLiteConfig) (*SQLBackend, error) {
	db, err := sql.Open("sqlite3", cfg.File)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(stmtCreateTable)
	if err != nil {

		return nil, err
	}

	reserve, err := db.Prepare(stmtReserve)
	if err != nil {
		return nil, err
	}

	getLocks, err := db.Prepare(stmtGetLocks)
	if err != nil {
		return nil, err
	}

	release, err := db.Prepare(stmtRelease)
	if err != nil {
		return nil, err
	}

	hasLock, err := db.Prepare(stmtHasLock)
	if err != nil {
		return nil, err
	}

	return &SQLBackend{
		db:       db,
		reserve:  reserve,
		getLocks: getLocks,
		release:  release,
		hasLock:  hasLock,
	}, nil
}

// Reserve a lock for the given group.
// Returns true if the lock is successfully reserved, even if the lock is already held by the specific id
func (s *SQLBackend) Reserve(group string, id string) error {
	_, err := s.reserve.Exec(group, id, time.Now(), group, id)
	if err != nil {
		return err
	}

	return nil
}

// Returns the current number of locks for the given group
func (s *SQLBackend) GetLocks(group string) (int, error) {
	rows, err := s.getLocks.Query(group)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, rows.Err()
	}

	var count int
	err = rows.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Release the lock currently held by the id.
// Does not fail when no lock is held.
func (s *SQLBackend) Release(group string, id string) error {
	_, err := s.release.Exec(group, id)
	if err != nil {
		return err
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
		return false, err
	}
	defer rows.Close()

	return rows.Next(), rows.Err()
}

// Calls all necessary finalization if necessary
func (s *SQLBackend) Close() error {
	return s.db.Close()
}
