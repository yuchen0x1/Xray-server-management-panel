package panel

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/util/crypto"
)

func initApiTokenDB(t *testing.T) *ApiTokenService {
	t.Helper()
	if err := database.InitDB(filepath.Join(t.TempDir(), "x-ui.db")); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() { _ = database.CloseDB() })
	return &ApiTokenService{}
}

// seedApiTokenRow inserts a row directly so a test can build states the
// service would refuse to create (unknown scope, past expiry). Enabled is
// written with an explicit update rather than the initial insert: the column
// carries gorm:"default:true", so GORM omits a zero-valued false on create.
func seedApiTokenRow(t *testing.T, row *model.ApiToken) {
	t.Helper()
	if row.Scope == "" {
		row.Scope = model.ApiScopeAdmin
	}
	wantEnabled := row.Enabled
	row.Enabled = true
	if err := database.GetDB().Create(row).Error; err != nil {
		t.Fatalf("seed api token: %v", err)
	}
	if !wantEnabled {
		if err := database.GetDB().Model(&model.ApiToken{}).Where("id = ?", row.Id).Update("enabled", false).Error; err != nil {
			t.Fatalf("seed api token disabled state: %v", err)
		}
	}
}

func TestMatchTokenFindsEnabledTokenByHash(t *testing.T) {
	svc := initApiTokenDB(t)
	created, err := svc.Create("ci", "", 0)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	row, ok := svc.MatchToken(created.Token)
	if !ok {
		t.Fatal("MatchToken rejected a token it just issued")
	}
	if row.Name != "ci" {
		t.Fatalf("matched row name = %q, want %q", row.Name, "ci")
	}
	if row.Scope != model.ApiScopeAdmin {
		t.Fatalf("matched row scope = %q, want %q (creation defaults to admin)", row.Scope, model.ApiScopeAdmin)
	}
	if got := svc.Match(created.Token); !got {
		t.Fatal("Match disagrees with MatchToken on a valid token")
	}
}

func TestMatchTokenRejectsUnknownAndDisabledRows(t *testing.T) {
	svc := initApiTokenDB(t)

	const disabledPlain = "disabled-token-value"
	seedApiTokenRow(t, &model.ApiToken{
		Name:    "off",
		Token:   crypto.HashTokenSHA256(disabledPlain),
		Enabled: false,
	})

	tests := []struct {
		name      string
		presented string
	}{
		{"empty value", ""},
		{"never issued", "some-other-token-value"},
		{"disabled row", disabledPlain},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := svc.MatchToken(tt.presented); ok {
				t.Fatalf("MatchToken(%q) = true, want false", tt.presented)
			}
			if svc.Match(tt.presented) {
				t.Fatalf("Match(%q) = true, want false", tt.presented)
			}
		})
	}
}

// A row whose scope was hand-edited to something this build never issued must
// not authenticate: falling back to admin would turn a typo into a privilege
// escalation.
func TestMatchTokenRejectsUnknownScope(t *testing.T) {
	svc := initApiTokenDB(t)

	const plaintext = "unknown-scope-token"
	if err := database.GetDB().Create(&model.ApiToken{
		Name:    "typo",
		Token:   crypto.HashTokenSHA256(plaintext),
		Enabled: true,
		Scope:   "admni",
	}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	if _, ok := svc.MatchToken(plaintext); ok {
		t.Fatal("MatchToken accepted a token with an unrecognized scope")
	}
}

func TestMatchTokenRejectsExpiredToken(t *testing.T) {
	svc := initApiTokenDB(t)

	const plaintext = "expired-token-value"
	seedApiTokenRow(t, &model.ApiToken{
		Name:      "stale",
		Token:     crypto.HashTokenSHA256(plaintext),
		Enabled:   true,
		ExpiresAt: nowMilli() - 1,
	})

	if _, ok := svc.MatchToken(plaintext); ok {
		t.Fatal("MatchToken accepted an expired token")
	}
}

func TestMatchTokenAcceptsFutureExpiry(t *testing.T) {
	svc := initApiTokenDB(t)
	created, err := svc.Create("short-lived", model.ApiScopeMonitor, nowMilli()+60_000)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Scope != model.ApiScopeMonitor {
		t.Fatalf("created scope = %q, want %q", created.Scope, model.ApiScopeMonitor)
	}

	row, ok := svc.MatchToken(created.Token)
	if !ok {
		t.Fatal("MatchToken rejected an unexpired monitor token")
	}
	if row.Scope != model.ApiScopeMonitor {
		t.Fatalf("matched scope = %q, want %q", row.Scope, model.ApiScopeMonitor)
	}
}

func TestNormalizeScope(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: "", want: model.ApiScopeAdmin},
		{in: "  Admin ", want: model.ApiScopeAdmin},
		{in: "monitor", want: model.ApiScopeMonitor},
		{in: "NODE-SYNC", want: model.ApiScopeNodeSync},
		{in: "node-admin", want: model.ApiScopeNodeAdmin},
		{in: "root", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := NormalizeScope(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("NormalizeScope(%q) = %q, want an error", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeScope(%q) returned %v", tt.in, err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeScope(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCreateRejectsPastExpiry(t *testing.T) {
	svc := initApiTokenDB(t)
	if _, err := svc.Create("past", "", nowMilli()-1); err == nil {
		t.Fatal("Create accepted an expiry in the past")
	}
	if _, err := svc.Create("negative", "", -1); err == nil {
		t.Fatal("Create accepted a negative expiry")
	}
}

func TestRecreateByNameKeepsOneTokenAndPreservesScope(t *testing.T) {
	svc := initApiTokenDB(t)
	first, err := svc.Create("grafana", model.ApiScopeMonitor, 0)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	rotated, err := svc.RecreateByName("grafana", "")
	if err != nil {
		t.Fatalf("RecreateByName: %v", err)
	}
	if rotated.Scope != model.ApiScopeMonitor {
		t.Fatalf("rotated scope = %q, want the stored %q", rotated.Scope, model.ApiScopeMonitor)
	}

	if _, ok := svc.MatchToken(first.Token); ok {
		t.Fatal("the previous token still authenticates after rotation")
	}
	if _, ok := svc.MatchToken(rotated.Token); !ok {
		t.Fatal("the rotated token does not authenticate")
	}

	rows, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("List returned %d rows after two rotations of one name, want 1", len(rows))
	}
}

// An expired token must not be revived by rotation, otherwise the CLI could be
// used to extend a credential the operator deliberately time-boxed.
func TestRecreateByNameRefusesExpiredToken(t *testing.T) {
	svc := initApiTokenDB(t)
	seedApiTokenRow(t, &model.ApiToken{
		Name:      "time-boxed",
		Token:     crypto.HashTokenSHA256("old-value"),
		Enabled:   true,
		ExpiresAt: nowMilli() - 1,
	})

	_, err := svc.RecreateByName("time-boxed", "")
	if err == nil {
		t.Fatal("RecreateByName revived an expired token")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Fatalf("error = %v, want it to mention expiry", err)
	}
}

func TestRecreateByNameRefusesUnknownStoredScope(t *testing.T) {
	svc := initApiTokenDB(t)
	if err := database.GetDB().Create(&model.ApiToken{
		Name:    "hand-edited",
		Token:   crypto.HashTokenSHA256("value"),
		Enabled: true,
		Scope:   "superuser",
	}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, err := svc.RecreateByName("hand-edited", "")
	if err == nil {
		t.Fatal("RecreateByName rotated a token with an unrecognized stored scope")
	}
}

func TestRecreateByNameRejectsOverlongName(t *testing.T) {
	svc := initApiTokenDB(t)
	if _, err := svc.RecreateByName(strings.Repeat("n", 65), ""); err == nil {
		t.Fatal("RecreateByName accepted a name longer than the column allows")
	}
}

func TestExpectedScopeGuardsBlockWrongRow(t *testing.T) {
	svc := initApiTokenDB(t)
	created, err := svc.Create("monitoring", model.ApiScopeMonitor, 0)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.SetEnabledExpectedScope(created.Id, model.ApiScopeAdmin, false); err == nil {
		t.Fatal("SetEnabledExpectedScope disabled a row whose scope did not match")
	}
	if err := svc.DeleteExpectedScope(created.Id, model.ApiScopeAdmin); err == nil {
		t.Fatal("DeleteExpectedScope deleted a row whose scope did not match")
	}
	if err := svc.DeleteExpectedScope(created.Id, ""); err == nil {
		t.Fatal("DeleteExpectedScope accepted an empty expected scope")
	}

	if err := svc.SetEnabledExpectedScope(created.Id, model.ApiScopeMonitor, false); err != nil {
		t.Fatalf("SetEnabledExpectedScope with the right scope: %v", err)
	}
	if _, ok := svc.MatchToken(created.Token); ok {
		t.Fatal("a disabled token still authenticates")
	}
	if err := svc.DeleteExpectedScope(created.Id, model.ApiScopeMonitor); err != nil {
		t.Fatalf("DeleteExpectedScope with the right scope: %v", err)
	}
	if _, err := svc.List(); err != nil && !errors.Is(err, nil) {
		t.Fatalf("List after delete: %v", err)
	}
}
