package service

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/iotest"
)

func TestReplaceGeofile(t *testing.T) {
	for _, tt := range []struct {
		name string
		body func() io.Reader
		want string
		err  error
	}{
		{name: "complete download", body: func() io.Reader { return strings.NewReader("updated rules") }, want: "updated rules"},
		{name: "interrupted download", body: func() io.Reader {
			return io.MultiReader(strings.NewReader("partial rules"), iotest.ErrReader(io.ErrUnexpectedEOF))
		}, want: "original rules", err: io.ErrUnexpectedEOF},
		{name: "empty download", body: func() io.Reader { return strings.NewReader("") }, want: "original rules", err: errors.New("downloaded geofile is empty")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "geoip.dat")
			if err := os.WriteFile(path, []byte("original rules"), 0o640); err != nil {
				t.Fatal(err)
			}
			err := replaceGeofile(path, tt.body())
			if tt.err == nil && err != nil {
				t.Fatalf("replaceGeofile: %v", err)
			}
			if tt.err != nil && (err == nil || err.Error() != tt.err.Error()) {
				t.Fatalf("replaceGeofile error = %v, want %v", err, tt.err)
			}
			content, err := os.ReadFile(path)
			if err != nil || string(content) != tt.want {
				t.Fatalf("geofile content = %q, error = %v; want %q", content, err, tt.want)
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 1 || entries[0].Name() != "geoip.dat" {
				t.Fatalf("download left temporary files: entries = %v, error = %v", entries, err)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if runtime.GOOS != "windows" && info.Mode().Perm() != 0o640 {
				t.Fatalf("permissions = %o, want 640", info.Mode().Perm())
			}
		})
	}
}

func TestReplaceGeofilePreservesOriginalOnTruncatedHTTPResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "100")
		_, _ = io.WriteString(w, "partial rules")
	}))
	defer server.Close()
	resp, err := server.Client().Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	dir := t.TempDir()
	path := filepath.Join(dir, "geosite.dat")
	if err := os.WriteFile(path, []byte("original rules"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := replaceGeofile(path, resp.Body); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("replaceGeofile error = %v, want unexpected EOF", err)
	}
	content, err := os.ReadFile(path)
	if err != nil || string(content) != "original rules" {
		t.Fatalf("geofile content = %q, error = %v; want original rules", content, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 || entries[0].Name() != "geosite.dat" {
		t.Fatalf("download left temporary files: entries = %v, error = %v", entries, err)
	}
}
