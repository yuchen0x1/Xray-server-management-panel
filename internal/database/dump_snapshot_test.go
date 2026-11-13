package database

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func openDumpTestDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestDumpPreservesSQLiteStorageTypes(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "source.db")
	src := openDumpTestDB(t, srcPath)
	if _, err := src.Exec(`CREATE TABLE "quoted""table" (id INTEGER PRIMARY KEY AUTOINCREMENT, value);`); err != nil {
		t.Fatal(err)
	}
	values := []any{[]byte("plain utf8"), []byte{}, []byte{0, 255, 128}, "before\x00after", "o'brien", nil}
	for _, value := range values {
		if _, err := src.Exec(`INSERT INTO "quoted""table" (value) VALUES (?)`, value); err != nil {
			t.Fatal(err)
		}
	}
	dumpPath := filepath.Join(dir, "backup.dump")
	if err := DumpSQLite(srcPath, dumpPath); err != nil {
		t.Fatal(err)
	}
	dstPath := filepath.Join(dir, "restored.db")
	if err := RestoreSQLite(dumpPath, dstPath); err != nil {
		t.Fatal(err)
	}
	dst := openDumpTestDB(t, dstPath)
	for i, value := range values {
		var got any
		var storage string
		if err := dst.QueryRow(`SELECT value, typeof(value) FROM "quoted""table" WHERE id = ?`, i+1).Scan(&got, &storage); err != nil {
			t.Fatal(err)
		}
		switch want := value.(type) {
		case []byte:
			actual, ok := got.([]byte)
			if !ok || storage != "blob" || !bytes.Equal(actual, want) {
				t.Errorf("row %d: got %v (%s), want blob %v", i, got, storage, want)
			}
		case string:
			if got != want || storage != "text" {
				t.Errorf("row %d: got %q (%s), want text %q", i, got, storage, want)
			}
		case nil:
			if got != nil || storage != "null" {
				t.Errorf("row %d: got %v (%s), want null", i, got, storage)
			}
		}
	}
}

func TestDumpSnapshotExcludesConcurrentWrites(t *testing.T) {
	dir := t.TempDir()
	src := openDumpTestDB(t, filepath.Join(dir, "source.db"))
	if _, err := src.Exec(`PRAGMA journal_mode=WAL; CREATE TABLE parent (value INTEGER); CREATE TABLE child (value INTEGER); INSERT INTO parent VALUES (1); INSERT INTO child VALUES (1);`); err != nil {
		t.Fatal(err)
	}
	tx, err := src.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	var initial int
	if err := tx.QueryRow(`SELECT value FROM parent`).Scan(&initial); err != nil || initial != 1 {
		t.Fatalf("snapshot initial value = %d, error = %v", initial, err)
	}
	if _, err := src.Exec(`BEGIN; UPDATE parent SET value=2; UPDATE child SET value=2; COMMIT;`); err != nil {
		t.Fatal(err)
	}
	data, err := dumpSQLiteSnapshot(tx)
	if err != nil {
		t.Fatal(err)
	}
	dumpPath := filepath.Join(dir, "snapshot.dump")
	if err := os.WriteFile(dumpPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	dstPath := filepath.Join(dir, "restored.db")
	if err := RestoreSQLite(dumpPath, dstPath); err != nil {
		t.Fatal(err)
	}
	dst := openDumpTestDB(t, dstPath)
	for _, table := range []string{"parent", "child"} {
		var got int
		if err := dst.QueryRow(`SELECT value FROM ` + table).Scan(&got); err != nil || got != 1 {
			t.Errorf("snapshot %s = %d, error = %v; want 1", table, got, err)
		}
		if err := src.QueryRow(`SELECT value FROM ` + table).Scan(&got); err != nil || got != 2 {
			t.Errorf("live %s = %d, error = %v; want 2", table, got, err)
		}
	}
}

func TestSQLiteSnapshotIncludesWALData(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "source.db")
	src := openDumpTestDB(t, srcPath)
	if _, err := src.Exec(`PRAGMA journal_mode=WAL; PRAGMA wal_autocheckpoint=0; CREATE TABLE probe (id INTEGER PRIMARY KEY, value TEXT); INSERT INTO probe VALUES (1, 'committed in WAL');`); err != nil {
		t.Fatal(err)
	}
	wal, err := os.Stat(srcPath + "-wal")
	if err != nil || wal.Size() == 0 {
		t.Fatalf("expected populated WAL: %v", err)
	}
	data, err := SnapshotSQLiteToBytes(srcPath)
	if err != nil {
		t.Fatal(err)
	}
	dstPath := filepath.Join(dir, "snapshot.db")
	if err := os.WriteFile(dstPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	dst := openDumpTestDB(t, dstPath)
	var value, integrity string
	if err := dst.QueryRow(`PRAGMA integrity_check`).Scan(&integrity); err != nil || integrity != "ok" {
		t.Fatalf("snapshot integrity = %q, error = %v", integrity, err)
	}
	if err := dst.QueryRow(`SELECT value FROM probe WHERE id=1`).Scan(&value); err != nil || value != "committed in WAL" {
		t.Fatalf("snapshot value = %q, error = %v", value, err)
	}
	if _, err := src.Exec(`INSERT INTO probe VALUES (2, 'after snapshot')`); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := dst.QueryRow(`SELECT count(*) FROM probe`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("snapshot rows = %d, error = %v; want 1", count, err)
	}
}
