package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPersistLocalBackupRetentionCountAndPermissions(t *testing.T) {
	dbDir := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", dbDir)
	t.Setenv("XUI_BACKUP_RETENTION_COUNT", "2")
	t.Setenv("XUI_BACKUP_RETENTION_DAYS", "")

	backupDir := filepath.Join(dbDir, "backups")
	if err := os.Mkdir(backupDir, 0o700); err != nil {
		t.Fatal(err)
	}
	oldest := filepath.Join(backupDir, localBackupPrefix+"oldest.db")
	newer := filepath.Join(backupDir, localBackupPrefix+"newer.db")
	unrelated := filepath.Join(backupDir, "manual.db")
	temporary := filepath.Join(backupDir, ".x-ui-backup-pending.tmp")
	for _, path := range []string{oldest, newer, unrelated, temporary} {
		if err := os.WriteFile(path, []byte(path), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now()
	if err := os.Chtimes(oldest, now.Add(-2*time.Hour), now.Add(-2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(newer, now.Add(-time.Hour), now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}

	service := &ServerService{}
	if err := service.persistLocalBackup([]byte("latest"), "host_2026-11-16_120000.db"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(oldest); !os.IsNotExist(err) {
		t.Fatalf("oldest backup still exists, stat error=%v", err)
	}
	if _, err := os.Stat(newer); err != nil {
		t.Fatalf("newer backup missing: %v", err)
	}
	for _, path := range []string{unrelated, temporary} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("unmanaged backup file %s missing: %v", path, err)
		}
	}
	latest := filepath.Join(backupDir, localBackupPrefix+"host_2026-11-16_120000.db")
	info, err := os.Stat(latest)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 && os.PathSeparator != '\\' {
		t.Fatalf("latest backup mode = %o, want 600", info.Mode().Perm())
	}
	data, err := os.ReadFile(latest)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "latest" {
		t.Fatalf("latest backup = %q, want latest", data)
	}
}

func TestPersistLocalBackupRetentionAgeAndPathSafety(t *testing.T) {
	dbDir := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", dbDir)
	t.Setenv("XUI_BACKUP_RETENTION_COUNT", "")
	t.Setenv("XUI_BACKUP_RETENTION_DAYS", "1")

	backupDir := filepath.Join(dbDir, "backups")
	if err := os.Mkdir(backupDir, 0o700); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(backupDir, localBackupPrefix+"stale.db")
	keep := filepath.Join(backupDir, localBackupPrefix+"keep.db")
	for _, path := range []string{stale, keep} {
		if err := os.WriteFile(path, []byte(path), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now()
	if err := os.Chtimes(stale, now.Add(-48*time.Hour), now.Add(-48*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(keep, now.Add(-time.Hour), now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}

	service := &ServerService{}
	if err := service.persistLocalBackup([]byte("latest"), "host_2026-11-16_120001.db"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale backup still exists, stat error=%v", err)
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatalf("fresh backup missing: %v", err)
	}

	t.Setenv("XUI_BACKUP_RETENTION_COUNT", "1")
	if err := service.persistLocalBackup([]byte("blocked"), filepath.Join("..", "outside.db")); err == nil {
		t.Fatal("path traversal filename was accepted")
	}
}

func TestBackupRetentionConfigRejectsNegativeValues(t *testing.T) {
	t.Setenv("XUI_BACKUP_RETENTION_COUNT", "-1")
	t.Setenv("XUI_BACKUP_RETENTION_DAYS", "")
	if _, _, err := backupRetentionConfig(); err == nil {
		t.Fatal("negative retention count was accepted")
	}
}

func TestPersistLocalBackupDisabledByDefault(t *testing.T) {
	dbDir := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", dbDir)
	t.Setenv("XUI_BACKUP_RETENTION_COUNT", "")
	t.Setenv("XUI_BACKUP_RETENTION_DAYS", "")

	service := &ServerService{}
	if err := service.persistLocalBackup([]byte("disabled"), "host_2026-11-16_120002.db"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dbDir, "backups")); !os.IsNotExist(err) {
		t.Fatalf("backup directory exists when retention is disabled, stat error=%v", err)
	}
}
