package service

import "testing"

func TestImportDBRejectsConcurrentRestore(t *testing.T) {
	databaseImportMu.Lock()
	defer databaseImportMu.Unlock()

	for i := range 2 {
		s := &ServerService{}
		err := s.ImportDB(nil)
		want := "A database restore is already in progress; wait for it to finish before retrying\n"
		if err == nil || err.Error() != want {
			t.Fatalf("service %d: ImportDB() = %v, want %q", i, err, want)
		}
	}
}
