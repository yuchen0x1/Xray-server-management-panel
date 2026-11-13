package database

import (
	"context"
	"database/sql"
	"os"
)

func SnapshotSQLiteToBytes(srcPath string) ([]byte, error) {
	if _, err := os.Stat(srcPath); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", srcPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	file, err := os.CreateTemp("", "x-ui-snapshot-*.db")
	if err != nil {
		return nil, err
	}
	path := file.Name()
	defer os.Remove(path)
	if err := file.Close(); err != nil {
		return nil, err
	}
	if _, err := db.ExecContext(context.Background(), "VACUUM INTO ?", path); err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}
