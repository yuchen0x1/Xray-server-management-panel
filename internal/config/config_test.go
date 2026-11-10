package config

import (
	"os"
	"strings"
	"testing"
)

func TestGetPanelVersion(t *testing.T) {
	orig := buildCommit
	t.Cleanup(func() { buildCommit = orig })

	buildCommit = ""
	if got := GetPanelVersion(); got != GetBaseVersion() {
		t.Fatalf("stable build: GetPanelVersion = %q, want %q", got, GetBaseVersion())
	}

	buildCommit = "1d1128cf"
	if got := GetPanelVersion(); got != "dev+1d1128cf" {
		t.Fatalf("dev build: GetPanelVersion = %q, want %q", got, "dev+1d1128cf")
	}

	buildCommit = "1d1128cf945c4615efa05cf41ba7fa766e2ee428"
	if got := GetPanelVersion(); got != "dev+1d1128cf" {
		t.Fatalf("dev build (full sha): GetPanelVersion = %q, want %q", got, "dev+1d1128cf")
	}
}

// TestGetBaseVersionIsUtf8 guards the embedded version file, not a specific
// release number: a truncated embed or a mangled encoding must fail here, while
// bumping the version must not require editing this test.
func TestGetBaseVersionIsUtf8(t *testing.T) {
	got := GetBaseVersion()
	if got == "" {
		t.Fatal("embedded version file is empty")
	}
	if strings.TrimSpace(got) != got {
		t.Fatalf("base version = %q carries stray whitespace", got)
	}
	for _, r := range got {
		if r != '.' && (r < '0' || r > '9') {
			t.Fatalf("base version = %q contains %q; want digits and dots only", got, r)
		}
	}
	if !strings.Contains(got, ".") {
		t.Fatalf("base version = %q, want a dotted release number", got)
	}
}

func TestGetPortOverride(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		set        bool
		wantPort   int
		configured bool
		wantErr    bool
	}{
		{name: "unset"},
		{name: "empty", value: "", set: true},
		{name: "whitespace", value: "   ", set: true},
		{name: "minimum", value: "1", set: true, wantPort: 1, configured: true},
		{name: "default panel port", value: "2053", set: true, wantPort: 2053, configured: true},
		{name: "surrounding whitespace", value: " 8080 ", set: true, wantPort: 8080, configured: true},
		{name: "maximum", value: "65535", set: true, wantPort: 65535, configured: true},
		{name: "zero", value: "0", set: true, configured: true, wantErr: true},
		{name: "above maximum", value: "65536", set: true, configured: true, wantErr: true},
		{name: "negative", value: "-1", set: true, configured: true, wantErr: true},
		{name: "non-numeric", value: "abc", set: true, configured: true, wantErr: true},
		{name: "decimal", value: "8080.0", set: true, configured: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				t.Setenv("XUI_PORT", tt.value)
			} else {
				original, existed := os.LookupEnv("XUI_PORT")
				if err := os.Unsetenv("XUI_PORT"); err != nil {
					t.Fatalf("unset XUI_PORT: %v", err)
				}
				t.Cleanup(func() {
					if existed {
						_ = os.Setenv("XUI_PORT", original)
					} else {
						_ = os.Unsetenv("XUI_PORT")
					}
				})
			}

			port, configured, err := GetPortOverride()
			if port != tt.wantPort {
				t.Errorf("port = %d, want %d", port, tt.wantPort)
			}
			if configured != tt.configured {
				t.Errorf("configured = %t, want %t", configured, tt.configured)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %t", err, tt.wantErr)
			}
		})
	}
}
